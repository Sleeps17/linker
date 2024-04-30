package linker

import (
	"context"
	"errors"
	linkerV1 "github.com/Sleeps17/linker-protos/gen/go/linker"
	"github.com/Sleeps17/linker/internal/storage"
	"github.com/Sleeps17/linker/pkg/random"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log/slog"
)

var (
	MsgRecordNotFound     = "link with this alias was not found"
	MsgAliasAlreadyExists = "link with such an alias already exists"
	MsgInternalError      = "Something went wrong"
)

type Service interface {
	Post(ctx context.Context, username, link, alias string) (err error)
	Pick(ctx context.Context, username, alias string) (link string, err error)
	List(ctx context.Context, username string) (links []string, err error)
	Delete(ctx context.Context, username, alias string) error
}

type serverAPI struct {
	linkerV1.UnimplementedLinkerServer
	log           *slog.Logger
	linkerService Service
}

func Register(s *grpc.Server, linkerService Service, log *slog.Logger) {
	linkerV1.RegisterLinkerServer(s, &serverAPI{linkerService: linkerService, log: log})
}

func (s *serverAPI) Post(ctx context.Context, req *linkerV1.PostRequest) (*linkerV1.PostResponse, error) {
	username := req.GetUsername()
	url := req.GetUrl()
	alias := req.GetAlias()

	s.log.Info("try to handle post request", slog.String("username", username), slog.String("alias", alias))

	// TODO: add validation of url

	if alias == "" {
		alias = random.Alias()
	}

	if err := s.linkerService.Post(ctx, username, url, alias); err != nil {
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

	link, err := s.linkerService.Pick(ctx, username, alias)
	if err != nil {
		if errors.Is(err, storage.ErrRecordNotFound) {
			s.log.Info("record not found", slog.String("alias", alias))
			return nil, status.Error(codes.InvalidArgument, MsgRecordNotFound)
		}

		s.log.Error("filed to handle pick request", slog.String("err", err.Error()))
		return nil, status.Error(codes.Internal, MsgInternalError)
	}

	s.log.Info("pick request handled successfully", slog.String("alias", alias))
	return &linkerV1.PickResponse{Url: link}, nil
}

func (s *serverAPI) List(ctx context.Context, req *linkerV1.ListRequest) (*linkerV1.ListResponse, error) {
	username := req.GetUsername()

	s.log.Info("try to handle list request", slog.String("username", username))

	links, err := s.linkerService.List(ctx, username)
	if err != nil {
		s.log.Error("filed to handle list request", slog.String("err", err.Error()))
		return nil, status.Error(codes.Internal, MsgInternalError)
	}

	s.log.Info("list request handled successfully")
	return &linkerV1.ListResponse{Links: links}, nil
}

func (s *serverAPI) Delete(ctx context.Context, req *linkerV1.DeleteRequest) (*linkerV1.DeleteResponse, error) {
	username := req.GetUsername()
	alias := req.GetAlias()

	s.log.Info("try to handle delete request", slog.String("username", username), slog.String("alias", alias))

	if err := s.linkerService.Delete(ctx, username, alias); err != nil {
		if errors.Is(err, storage.ErrRecordNotFound) {
			s.log.Info("record not found", slog.String("alias", alias))
			return nil, status.Error(codes.InvalidArgument, MsgRecordNotFound)
		}

		s.log.Error("filed to handle delete request", slog.String("err", err.Error()))
		return nil, status.Error(codes.Internal, MsgInternalError)
	}

	s.log.Info("delete request handled successfully", slog.String("alias", alias))
	return &linkerV1.DeleteResponse{Alias: alias}, nil
}
