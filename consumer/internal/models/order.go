package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Order struct {
	ID           primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	RestaurantID primitive.ObjectID   `bson:"restaurantId" json:"restaurantId"`
	TotalPrice   float64              `bson:"totalPrice" json:"totalPrice"`
	TotalCost    float64              `bson:"totalCost" json:"totalCost"`
	CreationDate time.Time            `bson:"creationDate" json:"creationDate"`
	Items        []primitive.ObjectID `bson:"items" json:"items"`
}
