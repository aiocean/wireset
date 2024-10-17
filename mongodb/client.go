package mongodb

import (
	"context"
	"errors"
	"os"

	"github.com/google/wire"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type Config struct {
	MongoDBURI string
}

var MongoFromEnvWireSet = wire.NewSet(
	NewConfigFromEnv,
	NewMongoDbClient,
)

func NewConfigFromEnv() (*Config, error) {
	mongoDBURI := os.Getenv("MONGODB_URI")
	if mongoDBURI == "" {
		return nil, errors.New("MONGODB_CONNECTION_URI is not found")
	}

	return &Config{
		MongoDBURI: mongoDBURI,
	}, nil
}

func NewMongoDbClient(config *Config, logger *zap.Logger) (*mongo.Client, func(), error) {
	ctx := context.Background()
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(config.MongoDBURI).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		if err := client.Disconnect(ctx); err != nil {
			logger.Error("failed to disconnect from mongodb", zap.Error(err))
		}
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, cleanup, err
	}

	return client, cleanup, nil
}
