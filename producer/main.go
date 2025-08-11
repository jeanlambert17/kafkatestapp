package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"producer/internal/config"
	"producer/internal/container"
	analytics "producer/internal/features/analytics"
	orders "producer/internal/features/orders"
	rests "producer/internal/features/restaurants"
)

func main() {
	cfg := config.Load()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c, err := container.New(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to init container: %v", err)
	}
	defer c.Close(context.Background())

	r := gin.Default()

	ordersController := orders.NewController(c.Orders)
	ordersController.RegisterRoutes(r)
	restaurantsController := rests.NewController(c.Restaurants)
	restaurantsController.RegisterRoutes(r)
	analyticsController := analytics.NewController(c.Analytics)
	analyticsController.RegisterRoutes(r)

	srv := &http.Server{Addr: ":" + cfg.Port, Handler: r}
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
