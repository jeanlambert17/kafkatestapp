package container

import (
	"context"

	"producer/internal/config"
	dbconn "producer/internal/db"
	analytics "producer/internal/features/analytics"
	"producer/internal/features/orders"
	rests "producer/internal/features/restaurants"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

type Container struct {
	Config   config.Config
	Redis    *redis.Client
	DBClient *mongo.Client
	DB       *mongo.Database

	Orders      *orders.Service
	Restaurants *rests.Service
	Analytics   *analytics.Service
}

func New(ctx context.Context, cfg config.Config) (*Container, error) {
	c := &Container{Config: cfg}

	// Redis client
	c.Redis = redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	if err := c.Redis.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	// MongoDB
	client, database, err := dbconn.Connect(ctx, cfg.MongoURI, cfg.MongoDatabase)
	if err != nil {
		return nil, err
	}
	c.DBClient, c.DB = client, database

	// Services
	c.Orders = orders.NewService(c.DB, cfg.KafkaBroker, c.Redis)
	c.Restaurants = rests.NewService(c.DB)
	c.Analytics = analytics.NewService(c.DB)

	return c, nil
}

func (c *Container) Close(_ context.Context) error {
	if c.Redis != nil {
		_ = c.Redis.Close()
	}
	if c.DBClient != nil {
		_ = c.DBClient.Disconnect(context.Background())
	}
	return nil
}
