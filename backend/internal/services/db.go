package services

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	// DefaultDatabaseName is used when no explicit database name is provided.
	DefaultDatabaseName = "hotelapp"
)

// ConnectMongo initialises a MongoDB client with a timeout to avoid hanging
// connections during startup.
func ConnectMongo(ctx context.Context, uri string) (*mongo.Client, error) {
	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.NewClient(clientOpts)
	if err != nil {
		return nil, err
	}

	c, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := client.Connect(c); err != nil {
		return nil, err
	}

	if err := client.Ping(c, nil); err != nil {
		_ = client.Disconnect(c)
		return nil, err
	}

	return client, nil
}
