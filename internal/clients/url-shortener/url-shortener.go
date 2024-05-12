package urlShortener

import "context"

type UrlShortener interface {
	SaveURL(ctx context.Context, url string, alias string) (_alias string, err error)
	DeleteURL(ctx context.Context, alias string) (err error)
}
