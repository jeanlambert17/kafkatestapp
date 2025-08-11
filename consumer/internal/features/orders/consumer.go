package orders

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"consumer/internal/models"

	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrdersEvent struct {
	RestaurantID string      `json:"restaurantId"`
	Items        []eventItem `json:"items"`
}

type eventItem struct {
	ID       string `json:"id"`
	Quantity int    `json:"quantity"`
}

func StartKafkaConsumer(ctx context.Context, broker string, topic string, groupID string, svc *Service) func() {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{broker},
		Topic:   topic,
		// GroupID:        groupID,
		MinBytes:       1,    // 1B
		MaxBytes:       10e6, // 10MB
		CommitInterval: time.Second,
	})

	runCtx, cancel := context.WithCancel(ctx)
	go func() {
		defer r.Close()
		for {
			message, err := r.ReadMessage(runCtx)
			if err != nil {
				if runCtx.Err() != nil {
					return
				}
				log.Printf("kafka read error: %v", err)
				continue
			}

			var evt OrdersEvent
			if err := json.Unmarshal(message.Value, &evt); err != nil {
				log.Printf("kafka unmarshal error: %v", err)
				continue
			}
			restaurantID, err := primitive.ObjectIDFromHex(evt.RestaurantID)
			if err != nil {
				log.Printf("invalid restaurant id: %v", err)
				continue
			}

			orderItems := make([]models.OrderItem, 0, len(evt.Items))
			for _, it := range evt.Items {
				oid, err := primitive.ObjectIDFromHex(it.ID)
				if err != nil || it.Quantity <= 0 {
					log.Printf("invalid item payload")
					orderItems = nil
					break
				}
				orderItems = append(orderItems, models.OrderItem{ItemID: oid, Quantity: it.Quantity})
			}

			if len(orderItems) == 0 {
				continue
			}

			if _, err := svc.CreateOrderFromEvent(runCtx, restaurantID, orderItems); err != nil {
				log.Printf("failed to create order from event: %v", err)
				continue
			}
		}
	}()

	return func() {
		cancel()
	}
}
