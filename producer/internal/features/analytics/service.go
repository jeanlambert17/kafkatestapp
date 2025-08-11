package analytics

import (
	"errors"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Service struct {
	db        *mongo.Database
	orders    *mongo.Collection
	items     *mongo.Collection
	dailyAggs *mongo.Collection
}

func NewService(database *mongo.Database) *Service {
	return &Service{
		db:        database,
		orders:    database.Collection("orders"),
		items:     database.Collection("items"),
		dailyAggs: database.Collection("daily_aggregates"),
	}
}

type DailyAggregate struct {
	Day         time.Time `bson:"day" json:"day"`
	TotalOrders int64     `bson:"totalOrders" json:"totalOrders"`
	Revenue     float64   `bson:"revenue" json:"revenue"`
}

func (s *Service) DailyAggregates(ctx *gin.Context, params url.Values) ([]DailyAggregate, error) {
	restaurantIDStr := params.Get("restaurantId")
	if restaurantIDStr == "" {
		return nil, errors.New("restaurantId is required")
	}
	restaurantID, err := primitive.ObjectIDFromHex(restaurantIDStr)
	if err != nil {
		return nil, errors.New("invalid restaurantId")
	}
	fromInclusive, toExclusive, err := parseFromTo(params)
	if err != nil {
		return nil, err
	}
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "restaurantId", Value: restaurantID},
			{Key: "day", Value: bson.D{{Key: "$gte", Value: fromInclusive}, {Key: "$lt", Value: toExclusive}}},
		}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "day", Value: 1}}}},
	}
	cursor, err := s.dailyAggs.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var out []DailyAggregate
	for cursor.Next(ctx) {
		var agg DailyAggregate
		if err := cursor.Decode(&agg); err != nil {
			return nil, err
		}
		out = append(out, agg)
	}
	return out, cursor.Err()
}

type PopularItem struct {
	ItemID   primitive.ObjectID `bson:"itemId" json:"itemId"`
	Name     string             `bson:"name" json:"name"`
	Quantity int64              `bson:"quantity" json:"quantity"`
	Revenue  float64            `bson:"revenue" json:"revenue"`
}

func (s *Service) MostPopularItems(ctx *gin.Context, params url.Values) ([]PopularItem, error) {
	fromInclusive, toExclusive, err := parseFromTo(params)
	if err != nil {
		return nil, err
	}
	pipeline := mongo.Pipeline{
		// filter by date range
		bson.D{{Key: "$match", Value: bson.D{{Key: "creationDate", Value: bson.D{{Key: "$gte", Value: fromInclusive}, {Key: "$lt", Value: toExclusive}}}}}},
		// explode order items
		bson.D{{Key: "$unwind", Value: "$items"}},
		// group by itemId and sum quantities from order items
		bson.D{{Key: "$group", Value: bson.D{{Key: "_id", Value: "$items.itemId"}, {Key: "quantity", Value: bson.D{{Key: "$sum", Value: "$items.quantity"}}}}}},
		// attach item details for price/name
		bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "items"}, {Key: "localField", Value: "_id"}, {Key: "foreignField", Value: "_id"}, {Key: "as", Value: "item"}}}},
		bson.D{{Key: "$unwind", Value: "$item"}},
		// compute revenue = price * quantity
		bson.D{{Key: "$addFields", Value: bson.D{{Key: "revenue", Value: bson.D{{Key: "$multiply", Value: bson.A{"$item.price", "$quantity"}}}}}}},
		// order by qty, then revenue
		bson.D{{Key: "$sort", Value: bson.D{{Key: "quantity", Value: -1}, {Key: "revenue", Value: -1}}}},
		// shape response
		bson.D{{Key: "$project", Value: bson.D{{Key: "_id", Value: 0}, {Key: "itemId", Value: "$_id"}, {Key: "name", Value: "$item.name"}, {Key: "quantity", Value: 1}, {Key: "revenue", Value: 1}}}},
	}
	cursor, err := s.orders.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var out []PopularItem
	for cursor.Next(ctx) {
		var p PopularItem
		if err := cursor.Decode(&p); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, cursor.Err()
}

// parseFromTo parses required from/to in MM/DD/YYYY and converts to [from 00:00, toNext 00:00)
func parseFromTo(params url.Values) (time.Time, time.Time, error) {
	fromStr := params.Get("from")
	toStr := params.Get("to")
	if fromStr == "" || toStr == "" {
		return time.Time{}, time.Time{}, errors.New("from and to are required (MM/DD/YYYY)")
	}
	fromDate, err := time.Parse("01/02/2006", fromStr)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("from must be MM/DD/YYYY")
	}
	toDate, err := time.Parse("01/02/2006", toStr)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("to must be MM/DD/YYYY")
	}
	fromStart := time.Date(fromDate.Year(), fromDate.Month(), fromDate.Day(), 0, 0, 0, 0, time.UTC)
	toNext := time.Date(toDate.Year(), toDate.Month(), toDate.Day(), 0, 0, 0, 0, time.UTC).Add(24 * time.Hour)
	return fromStart, toNext, nil
}
