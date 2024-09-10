package linker

import (
	"context"
	"github.com/Sleeps17/linker/internal/config"
	"github.com/Sleeps17/linker/internal/http/linker/handlers"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Server struct {
	api    *http.Server
	router *gin.Engine
}

func NewServer(cfg *config.ServerConfig, handlers ...handlers.Handler) *Server {
	g := gin.Default()

	for _, handler := range handlers {
		handler.Register(g)
	}

	srv := &http.Server{
		Handler:      g,
		Addr:         cfg.Port,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
	}

	return &Server{
		api:    srv,
		router: g,
	}
}

func (s *Server) Run() error {
	return s.api.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.api.Shutdown(ctx)
}
