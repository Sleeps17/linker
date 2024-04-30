package mongodb

import (
	"context"
	"errors"
	"fmt"
	"github.com/Sleeps17/linker/internal/storage"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	emptyAlias = ""
)

var (
	emptyLinks []string
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

func (s *Storage) Post(ctx context.Context, username, link, alias string) (err error) {
	filter := bson.D{{Key: "username", Value: username}, {Key: "link", Value: link}}

	var result Record
	if err := s.records.FindOne(ctx, filter).Decode(&result); err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			return storage.ErrAliasAlreadyExists
		}
	}

	data := Record{
		Username: username,
		Link:     link,
		Alias:    alias,
	}
	_, err = s.records.InsertOne(ctx, data)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) Pick(ctx context.Context, username, alias string) (link string, err error) {
	filter := bson.D{{Key: "username", Value: username}, {Key: "alias", Value: alias}}

	var result Record
	if err := s.records.FindOne(ctx, filter).Decode(&result); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return emptyAlias, storage.ErrRecordNotFound
		}
		return emptyAlias, err
	}

	return result.Link, nil
}

func (s *Storage) List(ctx context.Context, username string) (links []string, err error) {
	filter := bson.D{{Key: "username", Value: username}}

	cursor, err := s.records.Find(ctx, filter)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return emptyLinks, nil
		}
		return emptyLinks, err
	}
	defer func() { _ = cursor.Close(ctx) }()

	for cursor.Next(ctx) {
		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
		link := result["alias"].(string)
		links = append(links, link)
	}

	if err := cursor.Err(); err != nil {
		return emptyLinks, err
	}

	return links, nil
}

func (s *Storage) Delete(ctx context.Context, username string, alias string) error {
	filter := bson.D{{Key: "username", Value: username}, {Key: "alias", Value: alias}}

	res, err := s.records.DeleteOne(ctx, filter)

	if err != nil {
		return err
	}

	if res.DeletedCount == 0 {
		return storage.ErrRecordNotFound
	}

	return nil
}
