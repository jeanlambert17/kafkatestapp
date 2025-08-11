package items

import (
	"consumer/internal/models"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Service struct {
	db         *mongo.Database
	collection *mongo.Collection
}

func NewService(database *mongo.Database) *Service {
	return &Service{
		db:         database,
		collection: database.Collection("items"),
	}
}

func (s *Service) ListItems(ctx context.Context, ids []primitive.ObjectID) ([]models.Item, error) {
	var items []models.Item
	cursor, err := s.collection.Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var item models.Item
		if err := cursor.Decode(&item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return items, nil
}
