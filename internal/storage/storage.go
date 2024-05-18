package storage

import (
	"context"
	"errors"
)

type Storage interface {
	PostTopic(ctx context.Context, username, topic string) (topicID uint32, err error)
	DeleteTopic(ctx context.Context, username, topic string) (topicID uint32, err error)
	ListTopics(ctx context.Context, username string) (topics []string, err error)

	PostLink(ctx context.Context, username, topic, link, alias string) (err error)
	PickLink(ctx context.Context, username, topic, alias string) (link string, err error)
	DeleteLink(ctx context.Context, username, topic, alias string) (err error)
	ListLinks(ctx context.Context, username, topic string) (links []string, aliases []string, err error)

	Close(ctx context.Context) error
}

var (
	ErrTopicAlreadyExists = errors.New("topic already exists")
	ErrTopicNotFound      = errors.New("topic not found")

	ErrUserNotFound = errors.New("username not found")

	ErrAliasNotFound      = errors.New("alias not found")
	ErrAliasAlreadyExists = errors.New("alias already exists")

	ErrRecordNotFound = errors.New("alias not found")
)
