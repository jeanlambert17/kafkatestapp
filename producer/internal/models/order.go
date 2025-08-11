package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Order struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	RestaurantID primitive.ObjectID `bson:"restaurantId" json:"restaurantId"`
	TotalPrice   float64            `bson:"totalPrice" json:"totalPrice"`
	TotalCost    float64            `bson:"totalCost" json:"totalCost"`
	CreationDate time.Time          `bson:"creationDate" json:"creationDate"`
	Items        []OrderItem        `bson:"items" json:"items"`
}

type OrderItem struct {
	ItemID   primitive.ObjectID `bson:"itemId" json:"id"`
	Quantity int                `bson:"quantity" json:"quantity"`
}
