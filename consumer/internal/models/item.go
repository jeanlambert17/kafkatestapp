package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Item struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name         string             `bson:"name" json:"name"`
	RestaurantID primitive.ObjectID `bson:"restaurantId" json:"restaurantId"`
	Price        float64            `bson:"price" json:"price"`
	Cost         float64            `bson:"cost" json:"cost"`
	Quantity     int                `bson:"quantity" json:"quantity"`
}
