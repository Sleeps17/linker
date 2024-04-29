package storage

import "context"

type Storage interface {
	Post(ctx context.Context, username, link, alias string) (_alias string, err error)
	Pick(ctx context.Context, username, alias string) (link string, err error)
	List(ctx context.Context, username string) (links []string, err error)
	Close(ctx context.Context) error
}
