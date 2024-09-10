package handlers

import (
	"context"
	"errors"
	"github.com/Sleeps17/linker/internal/models"
	"github.com/Sleeps17/linker/internal/storage"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
)

type TopicService interface {
	PostTopic(ctx context.Context, username, topic string) (topicID uint32, err error)
	DeleteTopic(ctx context.Context, username, topic string) (topicID uint32, err error)
	ListTopics(ctx context.Context, username string) (topics []string, err error)
}

type TopicHandler struct {
	topicService TopicService
}

func NewTopicHandler(log *slog.Logger, topicService TopicService) *TopicHandler {
	return &TopicHandler{
		topicService: topicService,
	}
}

func (h *TopicHandler) Register(router *gin.Engine) {
	router.POST("/topics", h.postTopic)
	router.DELETE("/topics", h.deleteTopic)
	router.GET("/topics", h.listTopics)
}

func (h *TopicHandler) postTopic(c *gin.Context) {
	var req models.PostTopicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, models.ApiError{
			Message: "Неверный формат запроса",
			Error:   err.Error(),
		})
		return
	}

	id, err := h.topicService.PostTopic(c, req.Username, req.Topic)
	if err != nil {
		if errors.Is(err, storage.ErrTopicAlreadyExists) {
			c.AbortWithStatusJSON(http.StatusConflict, models.ApiError{
				Message: "Топик с таким названием уже существует",
				Error:   err.Error(),
			})
			return
		}

		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ApiError{
			Message: "Не удалось создать топик",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.PostTopicResponse{TopicID: id})
}

func (h *TopicHandler) deleteTopic(c *gin.Context) {
	var req models.DeleteTopicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, models.ApiError{
			Message: "Неверный формат запроса",
			Error:   err.Error(),
		})
		return
	}

	id, err := h.topicService.DeleteTopic(c, req.Username, req.Topic)
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
		}

		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ApiError{
			Message: "Не удалось удалить топик",
			Error:   err.Error(),
		})
	}

	c.JSON(http.StatusOK, models.DeleteTopicResponse{TopicID: id})
}

func (h *TopicHandler) listTopics(c *gin.Context) {
	var req models.ListTopicsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, models.ApiError{
			Message: "Неверный формат запроса",
			Error:   err.Error(),
		})
		return
	}

	topics, err := h.topicService.ListTopics(c, req.Username)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			c.AbortWithStatusJSON(http.StatusNotFound, models.ApiError{
				Message: "Пользователь не найден",
				Error:   err.Error(),
			})
			return
		}

		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ApiError{
			Message: "Не удалось получить список топиков",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.ListTopicsResponse{Topics: topics})
}
