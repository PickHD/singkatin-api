package infrastructure

import (
	"context"
	"fmt"
	"singkatin-api/auth/internal/config"
	"singkatin-api/auth/pkg/logger"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoConnectionProvider struct {
	db *mongo.Database
}

func NewMongoConnection(ctx context.Context, cfg *config.Config) *MongoConnectionProvider {
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%d", cfg.Database.Host, cfg.Database.Port)))
	if err != nil {
		logger.Errorf("failed connect mongoDB, error: %v", err)
	}
	db := mongoClient.Database(cfg.Database.Name)
	return &MongoConnectionProvider{db: db}
}

func (m *MongoConnectionProvider) GetCollection(name string) *mongo.Collection {
	return m.db.Collection(name)
}

func (m *MongoConnectionProvider) Client() *mongo.Client {
	return m.db.Client()
}

func (m *MongoConnectionProvider) GetDatabase() *mongo.Database {
	return m.db
}
