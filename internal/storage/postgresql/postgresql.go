package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/Sleeps17/linker/internal/storage"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

const (
	emptyLink = ""
)

var (
	emptyLinks   = []string{}
	emptyAliases = []string{}
)

type Storage struct {
	db *sql.DB
}

func MustNew(ctx context.Context, connString string) storage.Storage {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		panic("failed open db: " + err.Error())
	}

	if err := db.PingContext(ctx); err != nil {
		panic("failed ping db: " + err.Error())
	}

	s := &Storage{db: db}

	if err := s.init(context.TODO()); err != nil {
		panic("failed to init database: " + err.Error())
	}

	return s
}

func (s *Storage) Close(ctx context.Context) error {
	return s.db.Close()
}

func (s *Storage) Post(ctx context.Context, username, link, alias string) error {
	findUserQuery := `SELECT id FROM users WHERE username = $1;`

	var userID int
	err := s.db.QueryRowContext(ctx, findUserQuery, username).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			insertUserQuery := `INSERT INTO users (username) VALUES ($1);`
			_, err := s.db.ExecContext(ctx, insertUserQuery, username)
			if err != nil {
				return fmt.Errorf("failed to insert user: %w", err)
			}
		} else {
			return fmt.Errorf("failed to find user: %w", err)
		}
	}

	// TODO: end this function
	insertionQuery := `INSERT INTO links (user_id, link, alias) VALUES ($1, $2, $3);`
	_, err = s.db.ExecContext(ctx, insertionQuery, userID, link, alias)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == "unique_violation" {
			return storage.ErrAliasAlreadyExists
		}

		return fmt.Errorf("failed to insert link: %w", err)
	}

	return nil
}

func (s *Storage) Pick(ctx context.Context, username, alias string) (string, error) {
	queryFindUser := `SELECT id FROM users WHERE username = $1;`

	var userID int
	err := s.db.QueryRowContext(ctx, queryFindUser, username).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return emptyLink, storage.ErrUserNotFound
		}

		return emptyLink, fmt.Errorf("failed to find user: %w", err)
	}

	querySelectLink := `SELECT link FROM links WHERE user_id = $1 AND alias = $2;`

	var link string
	err = s.db.QueryRowContext(ctx, querySelectLink, userID, alias).Scan(&link)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return emptyLink, storage.ErrAliasNotFound
		}

		return emptyLink, fmt.Errorf("failed to find link: %w", err)
	}

	return link, nil
}

func (s *Storage) List(ctx context.Context, username string) ([]string, []string, error) {
	queryFindUser := `SELECT id FROM users WHERE username = $1;`

	var userID int
	err := s.db.QueryRowContext(ctx, queryFindUser, username).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return emptyLinks, emptyAliases, storage.ErrUserNotFound
		}

		return emptyLinks, emptyAliases, fmt.Errorf("failed to find user: %w", err)
	}

	querySelectData := `SELECT link, alias FROM links WHERE user_id = $1;`

	cursor, err := s.db.QueryContext(ctx, querySelectData, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return emptyLinks, emptyAliases, nil
		}
		return emptyLinks, emptyAliases, fmt.Errorf("failed to select data: %w", err)
	}

	links := make([]string, 0, 4)
	aliases := make([]string, 0, 4)

	var link, alias string
	for cursor.Next() {
		if err := cursor.Scan(&link, &alias); err != nil {
			return emptyLinks, emptyAliases, fmt.Errorf("failed to scan data: %w", err)
		}

		links = append(links, link)
		aliases = append(aliases, alias)
	}

	return links, aliases, nil
}

func (s *Storage) Delete(ctx context.Context, username string, alias string) error {
	queryFindUser := `SELECT id FROM users WHERE username = $1;`

	var userID int
	err := s.db.QueryRowContext(ctx, queryFindUser, username).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storage.ErrUserNotFound
		}

		return fmt.Errorf("failed to find user: %w", err)
	}

	queryDeleteRecord := `DELETE FROM links WHERE user_id = $1 AND alias = $2;`
	res, err := s.db.ExecContext(ctx, queryDeleteRecord, userID, alias)
	if err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}

	affectedRowsCount, _ := res.RowsAffected()
	if affectedRowsCount == 0 {
		return storage.ErrAliasNotFound
	}

	return nil
}

func (s *Storage) init(ctx context.Context) error {
	query := `CREATE TABLE IF NOT EXISTS "users" ("id" SERIAL PRIMARY KEY, "username" TEXT UNIQUE NOT NULL);`

	_, err := s.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create USERS table %w", err)
	}

	query = `CREATE TABLE IF NOT EXISTS "links" (
    "id" SERIAL PRIMARY KEY,
    "user_id" INT NOT NULL,
    "link" TEXT NOT NULL,
    "alias" TEXT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id),
    UNIQUE (user_id, alias)
);`

	_, err = s.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create LINKS table %w", err)
	}

	return nil
}
