package restaurants

import (
	"producer/internal/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Service struct {
	db             *mongo.Database
	restaurantsCol *mongo.Collection
}

func NewService(database *mongo.Database) *Service {
	return &Service{db: database, restaurantsCol: database.Collection("restaurants")}
}

type RestaurantWithItems struct {
	models.Restaurant `bson:",inline"`
	Items             []models.Item `bson:"items" json:"items"`
}

func (s *Service) ListRestaurants(ctx *gin.Context) ([]RestaurantWithItems, error) {
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "items"},
			{Key: "localField", Value: "_id"},
			{Key: "foreignField", Value: "restaurantId"},
			{Key: "as", Value: "items"},
		}}},
	}
	cursor, err := s.restaurantsCol.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var out []RestaurantWithItems
	for cursor.Next(ctx) {
		var r RestaurantWithItems
		if err := cursor.Decode(&r); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, cursor.Err()
}
