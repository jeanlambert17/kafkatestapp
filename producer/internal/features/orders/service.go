package orders

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"producer/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Service struct {
	broker     string
	topic      string
	db         *mongo.Database
	collection *mongo.Collection
	redis      *redis.Client
}

// NewService preserves old signature (no DB) for flexibility
func NewService(database *mongo.Database, broker string, redisClient *redis.Client) *Service {
	return &Service{
		broker:     broker,
		topic:      "orders",
		db:         database,
		collection: database.Collection("orders"),
		redis:      redisClient,
	}
}

type CreateOrderEventItem struct {
	ID       string `json:"id"`
	Quantity int    `json:"quantity"`
}

type CreateOrderEvent struct {
	RestaurantID string                 `json:"restaurantId"`
	Items        []CreateOrderEventItem `json:"items"`
}

func (s *Service) PublishOrder(ctx context.Context, req CreateOrderEvent) error {
	w := &kafka.Writer{
		Addr:                   kafka.TCP(s.broker),
		Topic:                  s.topic,
		AllowAutoTopicCreation: true,
		BatchTimeout:           100 * time.Millisecond,
	}
	defer w.Close()
	payload, err := json.Marshal(req)
	if err != nil {
		return err
	}
	if err := w.WriteMessages(ctx, kafka.Message{Value: payload}); err != nil {
		return err
	}
	if s.redis != nil && req.RestaurantID != "" {
		_ = s.redis.Del(ctx, recentCacheKey(req.RestaurantID)).Err()
	}
	return nil
}

type ListOrdersResponse struct {
	From    string         `json:"from"`
	Count   int64          `json:"count"`
	Results []models.Order `json:"results"`
}

func (s *Service) ListOrders(ctx *gin.Context, restaurantID primitive.ObjectID) (ListOrdersResponse, error) {
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
		var o models.Order
		if err := cursor.Decode(&o); err != nil {
			return ListOrdersResponse{}, err
		}
		orders = append(orders, o)
	}
	if err := cursor.Err(); err != nil {
		return ListOrdersResponse{}, err
	}
	return ListOrdersResponse{Count: count, Results: orders}, nil
}

const recentWindow = 15 * time.Minute
const recentCacheTTL = 5 * time.Minute

func recentCacheKey(restaurantID string) string {
	return fmt.Sprintf("recent_orders:%s", restaurantID)
}

func (s *Service) RecentOrders(ctx *gin.Context, org string) (ListOrdersResponse, error) {
	if s.redis != nil {
		if cached, err := s.redis.Get(ctx, recentCacheKey(org)).Bytes(); err == nil && len(cached) > 0 {
			var data ListOrdersResponse
			if unmarshalErr := json.Unmarshal(cached, &data); unmarshalErr == nil {
				data.From = "redis"
				return data, unmarshalErr
			}
		}
	}
	rid, err := primitive.ObjectIDFromHex(org)
	if err != nil {
		return ListOrdersResponse{}, errors.New("invalid x-org header format")
	}
	// This should be today's orders, maybe "pending" orders
	since := time.Now().Add(-recentWindow)
	filter := bson.D{
		{Key: "restaurantId", Value: rid},
		{Key: "creationDate", Value: bson.M{"$gte": since}},
	}
	reqCtx := ctx.Request.Context()
	cursor, err := s.collection.Find(reqCtx, filter)
	if err != nil {
		return ListOrdersResponse{}, err
	}
	defer cursor.Close(reqCtx)
	var orders []models.Order
	for cursor.Next(reqCtx) {
		var o models.Order
		if err := cursor.Decode(&o); err != nil {
			return ListOrdersResponse{}, err
		}
		orders = append(orders, o)
	}
	if err := cursor.Err(); err != nil {
		return ListOrdersResponse{}, err
	}
	data := ListOrdersResponse{Results: orders, Count: int64(len(orders)), From: "database"}
	if b, err := json.Marshal(data); err == nil {
		_ = s.redis.Set(ctx, recentCacheKey(org), b, recentCacheTTL).Err()
	}
	return data, nil
}
