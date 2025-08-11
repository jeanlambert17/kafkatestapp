package container

import (
	"context"

	"consumer/internal/config"
	dbconn "consumer/internal/db"
	items "consumer/internal/features/items"
	orders "consumer/internal/features/orders"

	"go.mongodb.org/mongo-driver/mongo"
)

type Container struct {
	Config      config.Config
	MongoClient *mongo.Client
	DB          *mongo.Database

	Orders *orders.Service
	Items  *items.Service

	// shutdown functions
	ShutdownFns []func()
}

func New(ctx context.Context, cfg config.Config) (*Container, error) {
	client, database, err := dbconn.Connect(ctx, cfg.MongoURI, cfg.MongoDatabase)
	if err != nil {
		return nil, err
	}

	container := &Container{
		Config:      cfg,
		MongoClient: client,
		DB:          database,
	}

	container.Items = items.NewService(database)

	// Initialize feature services
	container.Orders = orders.NewService(database, *container.Items)

	return container, nil
}

func (c *Container) Close(ctx context.Context) error {
	for i := len(c.ShutdownFns) - 1; i >= 0; i-- {
		if c.ShutdownFns[i] != nil {
			c.ShutdownFns[i]()
		}
	}
	if c.MongoClient != nil {
		return c.MongoClient.Disconnect(ctx)
	}
	return nil
}
