package mongodb

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Storage struct {
	records Records
}

type Records struct {
	*mongo.Collection
}

type Record struct {
	Username string `bson:"username"`
	Link     string `bson:"link"`
	Alias    string `bson:"alias"`
}

func MustNew(ctx context.Context, connString string, dbName string, collectionName string) *Storage {

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connString))
	if err != nil {
		panic(fmt.Sprintf("Failed connect to mongo: %v", err))
	}

	if err := client.Ping(ctx, nil); err != nil {
		panic(fmt.Sprintf("Failed ping mongo: %v", err))
	}

	records := Records{
		Collection: client.Database(dbName).Collection(collectionName),
	}

	return &Storage{
		records: records,
	}
}

func (s *Storage) Close(ctx context.Context) error {
	return s.records.Database().Client().Disconnect(ctx)
}

func (s *Storage) Post(ctx context.Context, username, link, alias string) (_alias string, err error) {
	panic("IMPLEMENT ME")
}

func (s *Storage) Pick(ctx context.Context, username, alias string) (link string, err error) {
	panic("IMPLEMENT ME")
}

func (s *Storage) List(ctx context.Context, username string) (links []string, err error) {
	panic("IMPLEMENT ME")
}
