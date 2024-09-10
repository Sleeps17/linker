package handlers

import (
	"context"
	"errors"
	"github.com/Sleeps17/linker/internal/models"
	"github.com/Sleeps17/linker/internal/storage"
	"github.com/Sleeps17/linker/pkg/random"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
)

type LinkService interface {
	PostLink(ctx context.Context, username, topic, link, alias string) (err error)
	PickLink(ctx context.Context, username, topic, alias string) (link string, err error)
	DeleteLink(ctx context.Context, username, topic, alias string) (err error)
	ListLinks(ctx context.Context, username, topic string) (links []string, aliases []string, err error)
}

type LinkHandler struct {
	linkService LinkService
}

func NewLinkHandler(log *slog.Logger, linkService LinkService) *LinkHandler {
	return &LinkHandler{
		linkService: linkService,
	}
}

func (h *LinkHandler) Register(router *gin.Engine) {
	router.POST("/links", h.postLink)
	router.GET("/links", h.getLink)
	router.DELETE("/links", h.deleteLink)
	router.GET("/links/list", h.listLinks)
}

func (h *LinkHandler) postLink(c *gin.Context) {
	var req models.PostLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, models.ApiError{
			Message: "Неверный формат запроса",
			Error:   err.Error(),
		})
		return
	}

	if req.Alias == "" {
		req.Alias = random.Alias()
	}

	if err := h.linkService.PostLink(c, req.Username, req.Topic, req.Link, req.Alias); err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			c.AbortWithStatusJSON(http.StatusNotFound, models.ApiError{
				Message: "Пользователь не найден",
				Error:   err.Error(),
			})
			return
		}

		if errors.Is(err, storage.ErrTopicNotFound) {
			c.AbortWithStatusJSON(http.StatusNotFound, models.ApiError{
				Message: "Топик не найден",
				Error:   err.Error(),
			})
			return
		}

		if errors.Is(err, storage.ErrAliasAlreadyExists) {
			c.AbortWithStatusJSON(http.StatusConflict, models.ApiError{
				Message: "Алиас уже существует",
				Error:   err.Error(),
			})
			return
		}

		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ApiError{
			Message: "Не удалось сохранить ссылку",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.PostLinkResponse{Alias: req.Alias})
}

func (h *LinkHandler) getLink(c *gin.Context) {
	var req models.PickLinkRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, models.ApiError{
			Message: "Неверный формат запроса",
			Error:   err.Error(),
		})
		return
	}

	link, err := h.linkService.PickLink(c, req.Username, req.Topic, req.Alias)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			c.AbortWithStatusJSON(http.StatusNotFound, models.ApiError{
				Message: "Пользователь не найден",
				Error:   err.Error(),
			})
			return
		}

		if errors.Is(err, storage.ErrTopicNotFound) {
			c.AbortWithStatusJSON(http.StatusNotFound, models.ApiError{
				Message: "Топик не найден",
				Error:   err.Error(),
			})
			return
		}

		if errors.Is(err, storage.ErrAliasNotFound) {
			c.AbortWithStatusJSON(http.StatusNotFound, models.ApiError{
				Message: "Адиас не найден",
				Error:   err.Error(),
			})
			return
		}

		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ApiError{
			Message: "Не удалось получить ссылку",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.PickLinkResponse{Link: link})
}

func (h *LinkHandler) deleteLink(c *gin.Context) {
	var req models.DeleteLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, models.ApiError{
			Message: "Неверный формат запроса",
			Error:   err.Error(),
		})
		return
	}

	if err := h.linkService.DeleteLink(c, req.Username, req.Topic, req.Alias); err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			c.AbortWithStatusJSON(http.StatusNotFound, models.ApiError{
				Message: "Пользователь не найден",
				Error:   err.Error(),
			})
			return
		}

		if errors.Is(err, storage.ErrTopicNotFound) {
			c.AbortWithStatusJSON(http.StatusNotFound, models.ApiError{
				Message: "Топик не найден",
				Error:   err.Error(),
			})
			return
		}

		if errors.Is(err, storage.ErrAliasNotFound) {
			c.AbortWithStatusJSON(http.StatusNotFound, models.ApiError{
				Message: "Алиас не найден",
				Error:   err.Error(),
			})
			return
		}

		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ApiError{
			Message: "Не удалось удалить ссылку",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.DeleteLinkResponse{Alias: req.Alias})
}

func (h *LinkHandler) listLinks(c *gin.Context) {
	var req models.ListLinksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, models.ApiError{
			Message: "Неверный формат запроса",
			Error:   err.Error(),
		})
		return
	}

	links, aliases, err := h.linkService.ListLinks(c, req.Username, req.Topic)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			c.AbortWithStatusJSON(http.StatusNotFound, models.ApiError{
				Message: "Пользователь не найден",
				Error:   err.Error(),
			})
			return
		}

		if errors.Is(err, storage.ErrTopicNotFound) {
			c.AbortWithStatusJSON(http.StatusNotFound, models.ApiError{
				Message: "Топик не найден",
				Error:   err.Error(),
			})
			return
		}

		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ApiError{
			Message: "Не удалось получить список ссылок",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.ListLinksResponse{Links: links, Aliases: aliases})
}
