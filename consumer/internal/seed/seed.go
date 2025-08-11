package seed

import (
	"context"
	"time"

	"consumer/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func SeedDatabase(ctx context.Context, db *mongo.Database) error {
	restaurantsCol := db.Collection("restaurants")
	itemsCol := db.Collection("items")

	// Seed restaurants if empty
	count, err := restaurantsCol.CountDocuments(ctx, bson.D{})

	if err != nil {
		return err
	}

	if count == 0 {
		restaurants := []interface{}{
			models.Restaurant{ID: primitive.NewObjectID(), Name: "Sunset Diner"},
			models.Restaurant{ID: primitive.NewObjectID(), Name: "Ocean Bistro"},
			models.Restaurant{ID: primitive.NewObjectID(), Name: "Mountain Grill"},
		}
		if _, err := restaurantsCol.InsertMany(ctx, restaurants); err != nil {
			return err
		}

		// Read back to get IDs
		cursor, err := restaurantsCol.Find(ctx, bson.D{})
		if err != nil {
			return err
		}
		defer cursor.Close(ctx)

		var seeded []models.Restaurant
		if err := cursor.All(ctx, &seeded); err != nil {
			return err
		}

		var items []interface{}
		for _, r := range seeded {
			switch r.Name {
			case "Sunset Diner":
				items = append(items,
					models.Item{Name: "Sunset Burger", RestaurantID: r.ID, Price: 12.50, Cost: 7.10},
					models.Item{Name: "Golden Fries", RestaurantID: r.ID, Price: 4.25, Cost: 1.20},
					models.Item{Name: "Dusk Salad", RestaurantID: r.ID, Price: 9.40, Cost: 3.50},
				)
			case "Ocean Bistro":
				items = append(items,
					models.Item{Name: "Seaside Salmon", RestaurantID: r.ID, Price: 18.75, Cost: 10.20},
					models.Item{Name: "Lemon Garlic Shrimp", RestaurantID: r.ID, Price: 16.40, Cost: 8.00},
					models.Item{Name: "Harbor Clam Chowder", RestaurantID: r.ID, Price: 11.60, Cost: 5.10},
				)
			case "Mountain Grill":
				items = append(items,
					models.Item{Name: "Trail Steak", RestaurantID: r.ID, Price: 21.90, Cost: 12.00},
					models.Item{Name: "Campfire Chili", RestaurantID: r.ID, Price: 9.50, Cost: 3.80},
					models.Item{Name: "Summit Skillet", RestaurantID: r.ID, Price: 14.30, Cost: 6.20},
				)
			default:
				items = append(items,
					models.Item{Name: "Chef Special", RestaurantID: r.ID, Price: 13.30, Cost: 6.00},
				)
			}
		}
		if len(items) > 0 {
			if _, err := itemsCol.InsertMany(ctx, items); err != nil {
				return err
			}
		}
	}

	// Ensure orders collection has index on creationDate
	ordersCol := db.Collection("orders")
	_, _ = ordersCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "creationDate", Value: 1}},
		Options: nil,
	})

	// Ensure restaurants name index
	_, _ = restaurantsCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "name", Value: 1}},
	})

	// Add a small delay to avoid racing immediately with freshly created indices in some Mongo setups
	time.Sleep(100 * time.Millisecond)
	return nil
}
