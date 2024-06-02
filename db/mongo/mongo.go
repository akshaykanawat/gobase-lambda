package mongo

import (
	"context"
	"time"

	"gobase-lambda/log"
	"gobase-lambda/utils"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Mongo struct {
	UrlHash string
	Client  *mongo.Client
	ctx     context.Context
	log     *log.Log
}

var ErrNoDocuments = mongo.ErrNoDocuments

var mongoCache clientCache = make(clientCache)

func GetDefaultClient(ctx context.Context, mongoURI string) (*Mongo, error) {
	return GetClient(ctx, mongoURI, log.GetDefaultLogger())
}

func NewMongoClient(ctx context.Context, mongoURI string, logger *log.Log) (*mongo.Client, error) {
	connectionOptions := options.Client()
	connectionOptions.ApplyURI(mongoURI)
	connectionOptions.SetConnectTimeout(time.Minute)
	connectionOptions.SetMaxConnIdleTime(time.Minute * 12)
	client, err := mongo.Connect(ctx, connectionOptions)
	if err != nil {
		logger.Error("Error creating mongo connection", err)
		return nil, err
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		logger.Error("Error pinging mongo server", err)
		return nil, err
	}
	return client, nil
}

func GetClient(ctx context.Context, mongoURI string, logger *log.Log) (*Mongo, error) {
	uriHash := utils.GetHash(mongoURI)
	cachedClient, ok := mongoCache[uriHash]
	var client *mongo.Client
	var err error
	if ok && time.Now().Before(cachedClient.expireTime) {
		logger.Debug("Mongo client taken from cache", nil)
		client = cachedClient.client
	} else {
		client, err = NewMongoClient(ctx, mongoURI, logger)
		if err != nil {
			return nil, err
		}
		mongoCache.cache(uriHash, client)
	}
	return &Mongo{Client: client, ctx: ctx, log: logger, UrlHash: uriHash}, nil
}
