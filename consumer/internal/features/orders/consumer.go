package orders

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrdersEvent struct {
	RestaurantID string   `json:"restaurantId"`
	Items        []string `json:"items"`
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

			itemIDs := make([]primitive.ObjectID, 0, len(evt.Items))
			for _, idStr := range evt.Items {
				oid, err := primitive.ObjectIDFromHex(idStr)
				if err != nil {
					log.Printf("invalid item id: %s", idStr)
					itemIDs = nil
					break
				}
				itemIDs = append(itemIDs, oid)
			}

			if len(itemIDs) == 0 {
				continue
			}

			if _, err := svc.CreateOrderFromEvent(runCtx, restaurantID, itemIDs); err != nil {
				log.Printf("failed to create order from event: %v", err)
				continue
			}
		}
	}()

	return func() { cancel() }
}
