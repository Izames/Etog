package handlers

import (
	"Etog/internal/domain/entity"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type MockEventHandler struct {
	log     *slog.Logger
	service MockEventService
}

type MockEventService interface {
	CreateMockEvent(mockEvent entity.MockEvent) error
	GetMockEvent(id int) (*entity.MockEvent, error)
	GetMockEvents() (*[]entity.MockEvent, error)
	UpdateMockEvent(id int, newMockEvent map[string]interface{}) (*entity.MockEvent, error)
	DeleteMockEvent(id int) error
}

func NewMockEventHandler(log *slog.Logger, service MockEventService) *MockEventHandler {
	return &MockEventHandler{
		log:     log,
		service: service,
	}
}

func (s *MockEventHandler) CreateMockEvent(ctx *gin.Context) {
	const op = "handlers.MockEvent.CreateMockEvent"
	var mockEvent entity.MockEvent
	s.log.With(slog.String("op", op))

	s.log.Debug("Create Mock Event Attempt")

	if err := ctx.ShouldBindJSON(&mockEvent); err != nil {
		s.log.Error("Failed to parse request body")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "при создании были указаны неверные данные"})
	}

	if err := s.service.CreateMockEvent(mockEvent); err != nil {
		s.log.Error("Failed in database to create Mock Event")
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "произошла серверная ошибка"})
	}

	s.log.Info("Create Mock Event Success")
	ctx.JSON(http.StatusCreated, gin.H{"message": "SUCCESSFULLY"})
}

func (s *MockEventHandler) GetMockEvent(ctx *gin.Context) {
	var mockEvent *entity.MockEvent
	const op = "handlers.MockEvent.GetMockEvent"
	s.log.With(slog.String("op", op))

	s.log.Debug("Get Mock Event Attempt")

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		s.log.Error("Failed to parse id: ", ctx.Param("id"))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "некорректный идентификатор мероприятия"})
	}

	mockEvent, err = s.service.GetMockEvent(id)
	if err != nil {
		s.log.Error("Failed in database to get Mock Event")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "не удалось найти мероприятие"})
	}

	s.log.Info("Get Mock Event Success")
	ctx.JSON(http.StatusOK, gin.H{"mock-event": mockEvent})
}

func (s *MockEventHandler) GetMockEvents(ctx *gin.Context) {
	var mockEvents *[]entity.MockEvent
	const op = "handlers.MockEvent.GetMockEvents"
	s.log.With(slog.String("op", op))

	s.log.Debug("Get Mock Event Attempt")

	mockEvents, err := s.service.GetMockEvents()
	if err != nil {
		s.log.Error("Failed in database to get Mock Events")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "не удалось найти никаких мероприятий"})
	}

	s.log.Info("Get Mock Event Success")
	ctx.JSON(http.StatusOK, gin.H{"mock-events": mockEvents})
}

func (s *MockEventHandler) UpdateMockEvent(ctx *gin.Context) {
	const op = "handlers.MockEvent.UpdateMockEvent"
	mockEvent := make(map[string]interface{})
	var newMockEvent *entity.MockEvent
	s.log.With(slog.String("op", op))

	s.log.Debug("Update Mock Event Attempt")

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		s.log.Error("Failed to parse id: ", ctx.Param("id"))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "неверный формат идентификатора мероприятия"})
	}

	if err := ctx.ShouldBindJSON(&mockEvent); err != nil {
		s.log.Error("Failed to parse request body")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "не удалось обновить мероприятие"})
	}

	newMockEvent, err = s.service.UpdateMockEvent(id, mockEvent)
	if err != nil {
		s.log.Error("Failed in database to update Mock Event")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "не удалось обновить мероприятие"})
	}
	s.log.Info("Update Mock Event Success")
	ctx.JSON(http.StatusOK, gin.H{"new-mock-event": newMockEvent})
}

func (s *MockEventHandler) DeleteMockEvent(ctx *gin.Context) {
	const op = "handlers.MockEvent.DeleteMockEvent"
	s.log.With(slog.String("op", op))

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		s.log.Error("Failed to parse id: ", ctx.Param("id"))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "неверный формат идентификатора мероприятия"})
	}

	if err := s.service.DeleteMockEvent(id); err != nil {
		s.log.Error("Failed in database to delete Mock Event")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "не удалось удалить мероприятие"})
	}
	s.log.Info("Delete Mock Event Success")

	ctx.JSON(http.StatusOK, gin.H{"message": "SUCCESSFULLY"})
}
