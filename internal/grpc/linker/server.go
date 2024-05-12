package linker

import (
	"context"
	"errors"
	linkerV1 "github.com/Sleeps17/linker-protos/gen/go/linker"
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
)

type Service interface {
	Post(ctx context.Context, username, link, alias string) (err error)
	Pick(ctx context.Context, username, alias string) (link string, err error)
	List(ctx context.Context, username string) (links []string, aliases []string, err error)
	Delete(ctx context.Context, username, alias string) error
}

type serverAPI struct {
	linkerV1.UnimplementedLinkerServer
	log           *slog.Logger
	urlShortener  urlShortener.UrlShortener
	linkerService Service
}

func Register(s *grpc.Server, linkerService Service, log *slog.Logger, urlShortener urlShortener.UrlShortener) {
	linkerV1.RegisterLinkerServer(s, &serverAPI{linkerService: linkerService, log: log, urlShortener: urlShortener})
}

func (s *serverAPI) Post(ctx context.Context, req *linkerV1.PostRequest) (*linkerV1.PostResponse, error) {
	username := req.GetUsername()
	link := req.GetLink()
	alias := req.GetAlias()

	s.log.Info("try to handle post request", slog.String("username", username), slog.String("alias", alias))

	if len(username) < minimalUsernameLength {
		s.log.Info("request with invalid username")
		return nil, status.Error(codes.InvalidArgument, MsgInvalidUsername)
	}

	if alias == emptyAlias {
		s.log.Info("request with empty alias, need to generate")
		alias = random.Alias()
	}

	if err := validator.New().Var(link, "required,url"); err != nil {
		s.log.Info("request with invalid link", slog.String("link", link))
		return nil, status.Error(codes.InvalidArgument, MsgInvalidLink)
	}

	newLink, err := s.urlShortener.SaveURL(ctx, link, alias)
	if err != nil {
		s.log.Info("filed to short link", slog.String("err", err.Error()))
	} else {
		link = newLink
		s.log.Info("short link generated", slog.String("link", link))
	}

	if err := s.linkerService.Post(ctx, username, link, alias); err != nil {
		if errors.Is(err, storage.ErrAliasAlreadyExists) {
			s.log.Info("alias already exists", slog.String("alias", alias))
			return nil, status.Error(codes.InvalidArgument, MsgAliasAlreadyExists)
		}

		s.log.Error("filed to handle post request", slog.String("err", err.Error()))
		return nil, status.Error(codes.Internal, MsgInternalError)
	}

	s.log.Info("post request handled successfully", slog.String("alias", alias))
	return &linkerV1.PostResponse{Alias: alias}, nil
}

func (s *serverAPI) Pick(ctx context.Context, req *linkerV1.PickRequest) (*linkerV1.PickResponse, error) {
	username := req.GetUsername()
	alias := req.GetAlias()

	s.log.Info("try to handle pick request", slog.String("username", username), slog.String("alias", alias))

	if len(username) < minimalUsernameLength {
		s.log.Info("request with invalid username")
		return nil, status.Error(codes.InvalidArgument, MsgInvalidUsername)
	}

	if alias == emptyAlias {
		s.log.Info("request with empty alias")
		return nil, status.Error(codes.InvalidArgument, MsgEmptyAlias)
	}

	link, err := s.linkerService.Pick(ctx, username, alias)
	if err != nil {
		if errors.Is(err, storage.ErrRecordNotFound) {
			s.log.Info("record not found", slog.String("alias", alias))
			return nil, status.Error(codes.InvalidArgument, MsgRecordNotFound)
		} else if errors.Is(err, storage.ErrUserNotFound) {
			s.log.Info("user not found", slog.String("username", username))
			return nil, status.Error(codes.InvalidArgument, MsgUserNotFound)
		} else if errors.Is(err, storage.ErrAliasNotFound) {
			s.log.Info("alias not found", slog.String("alias", alias))
			return nil, status.Error(codes.InvalidArgument, MsgAliasNotFound)
		}

		s.log.Error("failed to handle pick request", slog.String("err", err.Error()))
		return nil, status.Error(codes.Internal, MsgInternalError)
	}

	s.log.Info("pick request handled successfully", slog.String("alias", alias))
	return &linkerV1.PickResponse{Link: link}, nil
}

func (s *serverAPI) List(ctx context.Context, req *linkerV1.ListRequest) (*linkerV1.ListResponse, error) {
	username := req.GetUsername()

	s.log.Info("try to handle list request", slog.String("username", username))

	if len(username) < minimalUsernameLength {
		s.log.Info("request with invalid username")
		return nil, status.Error(codes.InvalidArgument, MsgInvalidUsername)
	}

	links, aliases, err := s.linkerService.List(ctx, username)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			s.log.Info("user not found", slog.String("username", username))
			return nil, status.Error(codes.InvalidArgument, MsgUserNotFound)
		}
		s.log.Error("filed to handle list request", slog.String("err", err.Error()))
		return nil, status.Error(codes.Internal, MsgInternalError)
	}

	s.log.Info("list request handled successfully")
	return &linkerV1.ListResponse{Links: links, Aliases: aliases}, nil
}

func (s *serverAPI) Delete(ctx context.Context, req *linkerV1.DeleteRequest) (*linkerV1.DeleteResponse, error) {
	username := req.GetUsername()
	alias := req.GetAlias()

	s.log.Info("try to handle delete request", slog.String("username", username), slog.String("alias", alias))

	if len(username) < minimalUsernameLength {
		s.log.Info("request with invalid username")
		return nil, status.Error(codes.InvalidArgument, MsgInvalidUsername)
	}

	if alias == emptyAlias {
		s.log.Info("request with empty alias")
		return nil, status.Error(codes.InvalidArgument, MsgEmptyAlias)
	}

	if err := s.linkerService.Delete(ctx, username, alias); err != nil {
		if errors.Is(err, storage.ErrRecordNotFound) {
			s.log.Info("record not found", slog.String("alias", alias))
			return nil, status.Error(codes.InvalidArgument, MsgRecordNotFound)
		} else if errors.Is(err, storage.ErrAliasNotFound) {
			s.log.Info("alias not found", slog.String("alias", alias))
			return nil, status.Error(codes.InvalidArgument, MsgAliasNotFound)
		} else if errors.Is(err, storage.ErrUserNotFound) {
			return nil, status.Error(codes.InvalidArgument, MsgUserNotFound)
		}

		s.log.Error("filed to handle delete request", slog.String("err", err.Error()))
		return nil, status.Error(codes.Internal, MsgInternalError)
	}

	err := s.urlShortener.DeleteURL(ctx, getAlias(alias))
	if err != nil {
		s.log.Error("failed to delete url", slog.String("err", err.Error()))
	}

	s.log.Info("delete request handled successfully", slog.String("alias", alias))
	return &linkerV1.DeleteResponse{Alias: alias}, nil
}

func getAlias(link string) string {
	parts := strings.Split(link, "/")
	lastPart := parts[len(parts)-1]
	return lastPart
}
