package orders

import (
	"context"
	"errors"
	"time"

	"consumer/internal/features/items"
	"consumer/internal/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Service struct {
	db         *mongo.Database
	collection *mongo.Collection
	items      items.Service
}

func NewService(database *mongo.Database, items items.Service) *Service {
	return &Service{
		db:         database,
		collection: database.Collection("orders"),
		items:      items,
	}
}

func (s *Service) CreateOrder(ctx *gin.Context, order *models.Order) (primitive.ObjectID, error) {
	restaurantIDHex := ctx.Request.Header.Get("x-org")
	if restaurantIDHex == "" {
		return primitive.NilObjectID, errors.New("missing or invalid x-org header")
	}
	restaurantID, restaurantIDErr := primitive.ObjectIDFromHex(restaurantIDHex)
	if restaurantIDErr != nil {
		return primitive.NilObjectID, errors.New("invalid x-org header format")
	}

	order.RestaurantID = restaurantID
	return s.createOrder(ctx.Request.Context(), order)
}

// CreateOrderFromEvent allows creating an order from a background consumer using a std context and explicit restaurant id
func (s *Service) CreateOrderFromEvent(ctx context.Context, restaurantID primitive.ObjectID, items []models.OrderItem) (primitive.ObjectID, error) {
	order := &models.Order{
		RestaurantID: restaurantID,
		Items:        items,
	}
	return s.createOrder(ctx, order)
}

// createOrder consolidates common steps to complete and persist an order
func (s *Service) createOrder(ctx context.Context, order *models.Order) (primitive.ObjectID, error) {
	if order.ID.IsZero() {
		order.ID = primitive.NewObjectID()
	}
	if order.CreationDate.IsZero() {
		order.CreationDate = time.Now().UTC()
	}

	// Build list of item IDs from order items
	itemIDs := make([]primitive.ObjectID, 0, len(order.Items))
	for _, it := range order.Items {
		itemIDs = append(itemIDs, it.ItemID)
	}
	fetchedItems, itemsErr := s.items.ListItems(ctx, itemIDs)
	if itemsErr != nil {
		return primitive.NilObjectID, itemsErr
	}
	var totalCost float64
	var totalPrice float64
	// Map itemID -> item for price/cost lookup
	for _, it := range order.Items {
		for _, ref := range fetchedItems {
			if ref.ID == it.ItemID {
				totalCost += ref.Cost * float64(it.Quantity)
				totalPrice += ref.Price * float64(it.Quantity)
				break
			}
		}
	}
	order.TotalCost = totalCost
	order.TotalPrice = totalPrice

	if _, err := s.collection.InsertOne(ctx, order); err != nil {
		return primitive.NilObjectID, err
	}

	// Update materialized daily aggregate (one document per day and restaurant)
	dayUTC := time.Date(order.CreationDate.UTC().Year(), order.CreationDate.UTC().Month(), order.CreationDate.UTC().Day(), 0, 0, 0, 0, time.UTC)
	filter := bson.M{
		"restaurantId": order.RestaurantID,
		"day":          dayUTC,
	}
	update := bson.M{
		"$setOnInsert": bson.M{
			"restaurantId": order.RestaurantID,
			"day":          dayUTC,
		},
		"$inc": bson.M{
			"totalOrders": 1,
			"revenue":     order.TotalPrice,
		},
	}
	_, _ = s.db.Collection("daily_aggregates").UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))

	return order.ID, nil
}
