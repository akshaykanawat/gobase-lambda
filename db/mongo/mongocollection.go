package mongo

import (
	"context"
	"gobase-lambda/log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Collection struct {
	DatabaseName   string
	CollectionName string
	Collection     *mongo.Collection
	ctx            context.Context
	log            *log.Log
	hashFieldMap   map[string]interface{}
}

func NewDefaultCollection(ctx context.Context, mongoURI, databaseName, collectionName string) (*Collection, error) {
	logger := log.GetDefaultLogger()
	client, err := GetClient(ctx, mongoURI, logger)
	if err != nil {
		return nil, err
	}
	return NewCollection(ctx, client.Client, databaseName, collectionName, logger), nil
}

func NewCollection(ctx context.Context, client *mongo.Client, databaseName, collectionName string, log *log.Log) *Collection {
	collectionOptions := options.Collection()
	return &Collection{DatabaseName: databaseName, CollectionName: collectionName, Collection: client.Database(databaseName).Collection(collectionName, collectionOptions), ctx: ctx, log: log}
}

func (m *Collection) SetHashList(hasList []string) {
	m.hashFieldMap = make(map[string]interface{}, len(hasList))
	for _, val := range hasList {
		m.hashFieldMap[val] = nil
	}
}
