package linker

import (
	"context"
	linkerV1 "github.com/Sleeps17/linker-protos/gen/go/linker"
	"google.golang.org/grpc"
)

type LinkerService interface {
	Post(ctx context.Context, username, link, alias string) (_alias string, err error)
	Pick(ctx context.Context, username, alias string) (link string, err error)
	List(ctx context.Context, username string) (links []string, err error)
}

type serverAPI struct {
	linkerV1.UnimplementedLinkerServer
	linkerService LinkerService
}

func Register(s *grpc.Server, linkerService LinkerService) {
	linkerV1.RegisterLinkerServer(s, &serverAPI{linkerService: linkerService})
}

func (s *serverAPI) Post(ctx context.Context, req *linkerV1.PostRequest) (*linkerV1.PostResponse, error) {
	panic("IMPLEMENT ME")
}

func (s *serverAPI) Pick(ctx context.Context, req *linkerV1.PickRequest) (*linkerV1.PickResponse, error) {
	panic("IMPLEMENT ME")
}

func (s *serverAPI) List(ctx context.Context, req *linkerV1.ListRequest) (*linkerV1.ListResponse, error) {
	panic("IMPLEMENT ME")
}
