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
	emptyLink   = ""
	zeroTopicId = 0
	zeroUserId  = 0
)

var (
	emptyTopics  = []string{}
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

func (s *Storage) PostTopic(ctx context.Context, username, topic string) (uint32, error) {
	const op = "postgresql.PostTopic"

	var userId uint32
	userId, err := s.findUser(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			userId, err = s.insertUser(ctx, username)

			if err != nil {
				return zeroTopicId, fmt.Errorf("%s: %w", op, err)
			}
		} else {
			return zeroTopicId, fmt.Errorf("%s: %w", op, err)
		}
	}

	var topicId uint32
	err = s.db.QueryRowContext(ctx, insertTopicQuery, userId, topic).Scan(&topicId)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == "unique_violation" {
			return 0, storage.ErrTopicAlreadyExists
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return topicId, nil
}

func (s *Storage) DeleteTopic(ctx context.Context, username, topic string) (uint32, error) {
	const op = "postgresql.DeleteTopic"

	userId, err := s.findUser(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return zeroTopicId, storage.ErrUserNotFound
		}

		return zeroTopicId, fmt.Errorf("%s: %w", op, err)
	}

	topicId, err := s.findTopic(ctx, userId, topic)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return zeroTopicId, storage.ErrTopicNotFound
		}

		return zeroTopicId, fmt.Errorf("%s: %w", op, err)
	}

	_, err = s.db.ExecContext(ctx, deleteLinksByTopicQuery, userId, topicId)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return zeroTopicId, fmt.Errorf("%s: %w", op, err)
		}
	}

	if err := s.db.QueryRowContext(ctx, deleteTopicQuery, userId, topic).Scan(&topicId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return zeroTopicId, storage.ErrTopicNotFound
		}

		return zeroTopicId, fmt.Errorf("%s: %w", op, err)
	}

	return topicId, nil
}

func (s *Storage) ListTopics(ctx context.Context, username string) ([]string, error) {
	const op = "postgresql.ListTopics"

	userId, err := s.findUser(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return emptyTopics, storage.ErrUserNotFound
		}

		return emptyTopics, fmt.Errorf("%s: %w", op, err)
	}

	cursor, err := s.db.QueryContext(ctx, listTopicsQuery, userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return emptyTopics, nil
		}
		return emptyTopics, fmt.Errorf("%s: %w", op, err)
	}

	topics := make([]string, 0)

	var topic string
	for cursor.Next() {
		if err := cursor.Scan(&topic); err != nil {
			return emptyTopics, fmt.Errorf("%s: %w", op, err)
		}

		topics = append(topics, topic)
	}

	return topics, nil
}

func (s *Storage) PostLink(ctx context.Context, username, topic, link, alias string) error {
	const op = "postgresql.PostLink"

	userId, err := s.findUser(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storage.ErrUserNotFound
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	topicId, err := s.findTopic(ctx, userId, topic)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storage.ErrTopicNotFound
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = s.db.ExecContext(ctx, insertLinkQuery, userId, topicId, link, alias)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == "unique_violation" {
			return storage.ErrAliasAlreadyExists
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) PickLink(ctx context.Context, username, topic, alias string) (string, error) {
	const op = "postgresql.PickLink"

	userId, err := s.findUser(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return emptyLink, storage.ErrUserNotFound
		}

		return emptyLink, fmt.Errorf("%s: %w", op, err)
	}

	topicId, err := s.findTopic(ctx, userId, topic)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return emptyLink, storage.ErrTopicNotFound
		}

		return emptyLink, fmt.Errorf("%s: %w", op, err)
	}

	var link string
	err = s.db.QueryRowContext(ctx, selectLinkQuery, userId, topicId, alias).Scan(&link)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return emptyLink, storage.ErrAliasNotFound
		}

		return emptyLink, fmt.Errorf("%s: %w", op, err)
	}

	return link, nil
}

func (s *Storage) ListLinks(ctx context.Context, username, topic string) ([]string, []string, error) {
	const op = "postgresql.ListLinks"

	userId, err := s.findUser(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return emptyLinks, emptyAliases, storage.ErrUserNotFound
		}

		return emptyLinks, emptyAliases, fmt.Errorf("%s: %w", op, err)
	}

	topicId, err := s.findTopic(ctx, userId, topic)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return emptyLinks, emptyAliases, storage.ErrTopicNotFound
		}

		return emptyLinks, emptyAliases, fmt.Errorf("%s: %w", op, err)
	}

	cursor, err := s.db.QueryContext(ctx, listLinksQuery, userId, topicId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return emptyLinks, emptyAliases, nil
		}
		return emptyLinks, emptyAliases, fmt.Errorf("failed to select data: %w", err)
	}

	links := make([]string, 0)
	aliases := make([]string, 0)

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

func (s *Storage) DeleteLink(ctx context.Context, username, topic, alias string) error {
	const op = "postgresql.DeleteLink"

	userId, err := s.findUser(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storage.ErrUserNotFound
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	topicId, err := s.findTopic(ctx, userId, topic)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storage.ErrTopicNotFound
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	res, err := s.db.ExecContext(ctx, deleteLinkQuery, userId, topicId, alias)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storage.ErrAliasNotFound
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	affectedRowsCount, _ := res.RowsAffected()
	if affectedRowsCount == 0 {
		return storage.ErrAliasNotFound
	}

	return nil
}

func (s *Storage) init(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, createUsersTableQuery)
	if err != nil {
		return fmt.Errorf("failed to create USERS table: %w", err)
	}

	_, err = s.db.ExecContext(ctx, createTopicsTableQuery)
	if err != nil {
		return fmt.Errorf("failed to create TOPICS table: %w", err)
	}

	_, err = s.db.ExecContext(ctx, createLinksTableQuery)
	if err != nil {
		return fmt.Errorf("failed to create LINKS table: %w", err)
	}

	return nil
}

func (s *Storage) insertUser(ctx context.Context, username string) (uint32, error) {
	const op = "postgresql.AddUser"

	var userId uint32
	err := s.db.QueryRowContext(ctx, insertUserQuery, username).Scan(&userId)
	if err != nil {
		return zeroUserId, fmt.Errorf("%s: %w", op, err)
	}

	return userId, nil
}

func (s *Storage) findUser(ctx context.Context, username string) (uint32, error) {
	const op = "postgresql.FindUser"

	var userId uint32
	if err := s.db.QueryRowContext(ctx, selectUserQuery, username).Scan(&userId); err != nil {
		return zeroUserId, fmt.Errorf("%s: %w", op, err)
	}

	return userId, nil
}

func (s *Storage) findTopic(ctx context.Context, userId uint32, topic string) (uint32, error) {
	const op = "postgresql.FindTopic"

	var topicId uint32
	if err := s.db.QueryRowContext(ctx, selectTopicQuery, userId, topic).Scan(&topicId); err != nil {
		return zeroTopicId, fmt.Errorf("%s: %w", op, err)
	}

	return topicId, nil
}
