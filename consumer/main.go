package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"consumer/internal/config"
	"consumer/internal/container"
	ordersFeature "consumer/internal/features/orders"
	"consumer/internal/seed"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, err := container.New(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to initialize container: %v", err)
	}
	defer func(ctx context.Context) {
		_ = c.Close(ctx)
	}(ctx)

	if err := seed.SeedDatabase(ctx, c.DB); err != nil {
		log.Fatalf("failed to seed database: %v", err)
	}

	router := gin.Default()

	orderService := c.Orders
	orderController := ordersFeature.NewController(orderService)
	orderController.RegisterRoutes(router)

	// Start Kafka consumer for orders
	stopKafka := ordersFeature.StartKafkaConsumer(ctx, cfg.KafkaBroker, "orders", "consumer-orders-group", orderService)
	c.ShutdownFns = append(c.ShutdownFns, stopKafka)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
