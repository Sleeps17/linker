package storage

import (
	"context"
	"errors"
)

type Storage interface {
	Post(ctx context.Context, username, link, alias string) (err error)
	Pick(ctx context.Context, username, alias string) (link string, err error)
	List(ctx context.Context, username string) (links []string, err error)
	Delete(ctx context.Context, username, alias string) error
	Close(ctx context.Context) error
}

var (
	ErrRecordNotFound     = errors.New("alias not found")
	ErrAliasAlreadyExists = errors.New("alias already exists")
)
