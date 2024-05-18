package linker

import (
	"context"
	"errors"
	linkerV2 "github.com/Sleeps17/linker-protos/gen/go/linker"
	urlShortener "github.com/Sleeps17/linker/internal/clients/url-shortener"
	"github.com/Sleeps17/linker/internal/storage"
	"github.com/Sleeps17/linker/pkg/random"
	"github.com/go-playground/validator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log/slog"
	"strings"
)

const (
	minimalUsernameLength = 8
	emptyAlias            = ""
	emptyTopic            = ""
)

type TopicService interface {
	PostTopic(ctx context.Context, username, topic string) (topicID uint32, err error)
	DeleteTopic(ctx context.Context, username, topic string) (topicID uint32, err error)
	ListTopics(ctx context.Context, username string) (topics []string, err error)
}

type LinkService interface {
	PostLink(ctx context.Context, username, topic, link, alias string) (err error)
	PickLink(ctx context.Context, username, topic, alias string) (link string, err error)
	DeleteLink(ctx context.Context, username, topic, alias string) (err error)
	ListLinks(ctx context.Context, username, topic string) (links []string, aliases []string, err error)
}

type serverAPI struct {
	linkerV2.UnimplementedLinkerServer
	log           *slog.Logger
	topicService  TopicService
	linkerService LinkService
	urlShortener  urlShortener.UrlShortener
}

func Register(
	s *grpc.Server,
	log *slog.Logger,
	linkerService LinkService,
	topicService TopicService,
	urlShortener urlShortener.UrlShortener,
) {
	linkerV2.RegisterLinkerServer(
		s, &serverAPI{
			log:           log,
			linkerService: linkerService,
			topicService:  topicService,
			urlShortener:  urlShortener,
		},
	)
}

func (s *serverAPI) PostTopic(ctx context.Context, req *linkerV2.PostTopicRequest) (*linkerV2.PostTopicResponse, error) {
	username := req.GetUsername()
	topic := req.GetTopic()

	s.log.Info("try to handle post topic request", slog.String("username", username), slog.String("topic", topic))

	if len(username) < minimalUsernameLength {
		s.log.Info("request with invalid username")
		return nil, status.Error(codes.InvalidArgument, MsgInvalidUsername)
	}

	if topic == emptyTopic {
		s.log.Info("request with empty topic")
		return nil, status.Error(codes.InvalidArgument, MsgEmptyTopic)
	}

	topicId, err := s.topicService.PostTopic(ctx, username, topic)
	if err != nil {

		if errors.Is(err, storage.ErrTopicAlreadyExists) {
			s.log.Info("topic already exists", slog.String("user", username), slog.String("topic", topic))
			return nil, status.Error(codes.InvalidArgument, MsgTopicAlreadyExists)
		}

		s.log.Error("failed to handle post topic request", slog.String("err", err.Error()))
		return nil, status.Error(codes.Internal, MsgInternalError)
	}

	s.log.Info("post topic request handled successfully", slog.Any("topic_id", topicId))
	return &linkerV2.PostTopicResponse{TopicId: topicId}, nil
}

func (s *serverAPI) DeleteTopic(ctx context.Context, req *linkerV2.DeleteTopicRequest) (*linkerV2.DeleteTopicResponse, error) {
	username := req.GetUsername()
	topic := req.GetTopic()

	s.log.Info("try to handle delete topic request", slog.String("username", username), slog.String("topic", topic))

	if len(username) < minimalUsernameLength {
		s.log.Info("request with invalid username")
		return nil, status.Error(codes.InvalidArgument, MsgInvalidUsername)
	}

	if topic == emptyTopic {
		s.log.Info("request with empty topic")
		return nil, status.Error(codes.InvalidArgument, MsgEmptyTopic)
	}

	topicId, err := s.topicService.DeleteTopic(ctx, username, topic)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			s.log.Info("user not found", slog.String("user", username), slog.String("topic", topic))
			return nil, status.Error(codes.InvalidArgument, MsgUserNotFound)
		}

		if errors.Is(err, storage.ErrTopicNotFound) {
			s.log.Info("topic not found", slog.String("user", username), slog.String("topic", topic))
			return nil, status.Error(codes.InvalidArgument, MsgTopicNotFound)
		}

		s.log.Error("failed to handle delete topic request", slog.String("err", err.Error()))
		return nil, status.Error(codes.Internal, MsgInternalError)
	}

	s.log.Info("delete topic request handled successfully", slog.Any("topic_id", topicId))
	return &linkerV2.DeleteTopicResponse{TopicId: topicId}, nil
}

func (s *serverAPI) ListTopics(ctx context.Context, req *linkerV2.ListTopicsRequest) (*linkerV2.ListTopicsResponse, error) {
	username := req.GetUsername()

	if len(username) < minimalUsernameLength {
		s.log.Info("request with invalid username")
		return nil, status.Error(codes.InvalidArgument, MsgInvalidUsername)
	}

	topics, err := s.topicService.ListTopics(ctx, username)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			s.log.Info("user not found", slog.String("user", username))
			return nil, status.Error(codes.InvalidArgument, MsgUserNotFound)
		}

		s.log.Error("failed to handle list topics request", slog.String("err", err.Error()))
		return nil, status.Error(codes.Internal, MsgInternalError)
	}

	s.log.Info("list topics request handled successfully")
	return &linkerV2.ListTopicsResponse{Topics: topics}, nil
}

func (s *serverAPI) PostLink(ctx context.Context, req *linkerV2.PostLinkRequest) (*linkerV2.PostLinkResponse, error) {
	username := req.GetUsername()
	topic := req.GetTopic()
	link := req.GetLink()
	alias := req.GetAlias()

	s.log.Info("try to handle post request", slog.String("username", username), slog.String("alias", alias))

	if len(username) < minimalUsernameLength {
		s.log.Info("request with invalid username")
		return nil, status.Error(codes.InvalidArgument, MsgInvalidUsername)
	}

	if topic == emptyTopic {
		s.log.Info("request with empty topic")
		return nil, status.Error(codes.InvalidArgument, MsgEmptyTopic)
	}

	if alias == emptyAlias {
		s.log.Info("request with empty alias, need to generate")
		alias = random.Alias()
	}

	if err := validator.New().Var(link, "required,url"); err != nil {
		s.log.Info("request with invalid link", slog.String("link", link))
		return nil, status.Error(codes.InvalidArgument, MsgInvalidLink)
	}

	s.log.Info("try to short link", slog.String("link", link))
	newLink, err := s.urlShortener.SaveURL(ctx, link, alias)
	if err != nil {
		s.log.Info("failed to short link", slog.String("err", err.Error()))
	} else {
		link = newLink
		s.log.Info("short link generated", slog.String("link", link))
	}

	if err := s.linkerService.PostLink(ctx, username, topic, link, alias); err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			s.log.Info("user not found", slog.String("user", username))
			return nil, status.Error(codes.InvalidArgument, MsgUserNotFound)
		}

		if errors.Is(err, storage.ErrTopicNotFound) {
			s.log.Info("topic not found", slog.String("user", username), slog.String("topic", topic))
			return nil, status.Error(codes.InvalidArgument, MsgTopicNotFound)
		}

		if errors.Is(err, storage.ErrAliasAlreadyExists) {
			s.log.Info("alias already exists", slog.String("alias", alias))
			return nil, status.Error(codes.InvalidArgument, MsgAliasAlreadyExists)
		}

		s.log.Error("failed to handle post request", slog.String("err", err.Error()))
		return nil, status.Error(codes.Internal, MsgInternalError)
	}

	s.log.Info("post request handled successfully", slog.String("alias", alias))
	return &linkerV2.PostLinkResponse{Alias: alias}, nil
}

func (s *serverAPI) PickLink(ctx context.Context, req *linkerV2.PickLinkRequest) (*linkerV2.PickLinkResponse, error) {
	username := req.GetUsername()
	topic := req.GetTopic()
	alias := req.GetAlias()

	s.log.Info("try to handle pick request", slog.String("username", username), slog.String("alias", alias))

	if len(username) < minimalUsernameLength {
		s.log.Info("request with invalid username")
		return nil, status.Error(codes.InvalidArgument, MsgInvalidUsername)
	}

	if topic == emptyTopic {
		s.log.Info("request with empty topic")
		return nil, status.Error(codes.InvalidArgument, MsgEmptyTopic)
	}

	if alias == emptyAlias {
		s.log.Info("request with empty alias")
		return nil, status.Error(codes.InvalidArgument, MsgEmptyAlias)
	}

	link, err := s.linkerService.PickLink(ctx, username, topic, alias)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			s.log.Info("user not found", slog.String("user", username))
			return nil, status.Error(codes.InvalidArgument, MsgUserNotFound)
		}

		if errors.Is(err, storage.ErrTopicNotFound) {
			s.log.Info("topic not found", slog.String("user", username), slog.String("topic", topic))
			return nil, status.Error(codes.InvalidArgument, MsgTopicNotFound)
		}

		if errors.Is(err, storage.ErrAliasNotFound) {
			s.log.Info("alias not found", slog.String("alias", alias))
			return nil, status.Error(codes.InvalidArgument, MsgAliasNotFound)
		}

		s.log.Error("failed to handle pick request", slog.String("err", err.Error()))
		return nil, status.Error(codes.Internal, MsgInternalError)
	}

	s.log.Info("pick request handled successfully", slog.String("alias", alias))
	return &linkerV2.PickLinkResponse{Link: link}, nil
}

func (s *serverAPI) ListLinks(ctx context.Context, req *linkerV2.ListLinksRequest) (*linkerV2.ListLinksResponse, error) {
	username := req.GetUsername()
	topic := req.GetTopic()

	s.log.Info("try to handle list request", slog.String("username", username))

	if len(username) < minimalUsernameLength {
		s.log.Info("request with invalid username")
		return nil, status.Error(codes.InvalidArgument, MsgInvalidUsername)
	}

	if topic == emptyTopic {
		s.log.Info("request with empty topic")
		return nil, status.Error(codes.InvalidArgument, MsgEmptyTopic)
	}

	links, aliases, err := s.linkerService.ListLinks(ctx, username, topic)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			s.log.Info("user not found", slog.String("user", username))
			return nil, status.Error(codes.InvalidArgument, MsgUserNotFound)
		}

		if errors.Is(err, storage.ErrTopicNotFound) {
			s.log.Info("topic not found", slog.String("user", username), slog.String("topic", topic))
			return nil, status.Error(codes.InvalidArgument, MsgTopicNotFound)
		}

		s.log.Error("filed to handle list request", slog.String("err", err.Error()))
		return nil, status.Error(codes.Internal, MsgInternalError)
	}

	s.log.Info("list request handled successfully")
	return &linkerV2.ListLinksResponse{Links: links, Aliases: aliases}, nil
}

func (s *serverAPI) DeleteLink(ctx context.Context, req *linkerV2.DeleteLinkRequest) (*linkerV2.DeleteLinkResponse, error) {
	username := req.GetUsername()
	topic := req.GetTopic()
	alias := req.GetAlias()

	s.log.Info("try to handle delete request", slog.String("username", username), slog.String("alias", alias))

	if len(username) < minimalUsernameLength {
		s.log.Info("request with invalid username")
		return nil, status.Error(codes.InvalidArgument, MsgInvalidUsername)
	}

	if topic == emptyTopic {
		s.log.Info("request with empty topic")
		return nil, status.Error(codes.InvalidArgument, MsgEmptyTopic)
	}

	if alias == emptyAlias {
		s.log.Info("request with empty alias")
		return nil, status.Error(codes.InvalidArgument, MsgEmptyAlias)
	}

	if err := s.linkerService.DeleteLink(ctx, username, topic, alias); err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			s.log.Info("user not found", slog.String("user", username))
			return nil, status.Error(codes.InvalidArgument, MsgUserNotFound)
		}

		if errors.Is(err, storage.ErrTopicNotFound) {
			s.log.Info("topic not found", slog.String("user", username), slog.String("topic", topic))
			return nil, status.Error(codes.InvalidArgument, MsgTopicNotFound)
		}

		if errors.Is(err, storage.ErrAliasNotFound) {
			s.log.Info("alias not found", slog.String("alias", alias))
			return nil, status.Error(codes.InvalidArgument, MsgAliasNotFound)
		}

		s.log.Error("filed to handle delete request", slog.String("err", err.Error()))
		return nil, status.Error(codes.Internal, MsgInternalError)
	}

	s.log.Info("try to delete link from shortener", slog.String("alias", alias))
	err := s.urlShortener.DeleteURL(ctx, getAlias(alias))
	if err != nil {
		s.log.Info("failed to delete url", slog.String("err", err.Error()))
	}

	s.log.Info("delete request handled successfully", slog.String("alias", alias))
	return &linkerV2.DeleteLinkResponse{Alias: alias}, nil
}

func getAlias(link string) string {
	parts := strings.Split(link, "/")
	lastPart := parts[len(parts)-1]
	return lastPart
}
