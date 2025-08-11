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
)

type Service struct {
	db         *mongo.Database
	collection *mongo.Collection
	items      items.Service
}

type ListOrdersResponse struct {
	Results []models.Order `json:"results"`
	Count   int64          `json:"count"`
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
	return s.finalizeAndInsert(ctx.Request.Context(), order)
}

// CreateOrderFromEvent allows creating an order from a background consumer using a std context and explicit restaurant id
func (s *Service) CreateOrderFromEvent(ctx context.Context, restaurantID primitive.ObjectID, itemIDs []primitive.ObjectID) (primitive.ObjectID, error) {
	order := &models.Order{
		RestaurantID: restaurantID,
		Items:        itemIDs,
	}
	return s.finalizeAndInsert(ctx, order)
}

// finalizeAndInsert consolidates common steps to complete and persist an order
func (s *Service) finalizeAndInsert(ctx context.Context, order *models.Order) (primitive.ObjectID, error) {
	if order.ID.IsZero() {
		order.ID = primitive.NewObjectID()
	}
	if order.CreationDate.IsZero() {
		order.CreationDate = time.Now().UTC()
	}

	fetchedItems, itemsErr := s.items.ListItems(ctx, order.Items)
	if itemsErr != nil {
		return primitive.NilObjectID, itemsErr
	}
	var totalCost float64
	var totalPrice float64
	for _, it := range fetchedItems {
		totalCost += it.Cost
		totalPrice += it.Price
	}
	order.TotalCost = totalCost
	order.TotalPrice = totalPrice

	if _, err := s.collection.InsertOne(ctx, order); err != nil {
		return primitive.NilObjectID, err
	}
	return order.ID, nil
}

func (s *Service) GetOrderByID(ctx context.Context, id primitive.ObjectID) (*models.Order, error) {
	var order models.Order
	err := s.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&order)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

func (s *Service) ListOrders(ctx *gin.Context) (ListOrdersResponse, error) {
	restaurantIDHex := ctx.Request.Header.Get("x-org")
	if restaurantIDHex == "" {
		return ListOrdersResponse{}, errors.New("missing or invalid x-org header")
	}
	restaurantID, restaurantIDErr := primitive.ObjectIDFromHex(restaurantIDHex)
	if restaurantIDErr != nil {
		return ListOrdersResponse{}, errors.New("invalid x-org header format")
	}

	filter := bson.D{{Key: "restaurantId", Value: restaurantID}}
	reqCtx := ctx.Request.Context()
	count, err := s.collection.CountDocuments(reqCtx, filter)
	if err != nil {
		return ListOrdersResponse{}, err
	}

	cursor, err := s.collection.Find(reqCtx, filter)
	if err != nil {
		return ListOrdersResponse{}, err
	}
	defer cursor.Close(reqCtx)

	var orders []models.Order
	for cursor.Next(reqCtx) {
		var order models.Order
		if err := cursor.Decode(&order); err != nil {
			return ListOrdersResponse{}, err
		}
		orders = append(orders, order)
	}
	if err := cursor.Err(); err != nil {
		return ListOrdersResponse{}, err
	}

	return ListOrdersResponse{Results: orders, Count: count}, nil
}

func (s *Service) UpdateOrder(ctx context.Context, id primitive.ObjectID, update bson.M) error {
	_, err := s.collection.UpdateByID(ctx, id, bson.M{"$set": update})
	return err
}

func (s *Service) DeleteOrder(ctx context.Context, id primitive.ObjectID) error {
	_, err := s.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
