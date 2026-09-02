package handlers

import (
	"Etog/internal/config"
	"Etog/internal/domain/entity/DTO/event_dto"
	"context"
	"log/slog"
	"mime/multipart"
	"strconv"

	"github.com/gin-gonic/gin"
)

type EventHandler struct {
	log     *slog.Logger
	service EventService
	config  config.Config
}

type EventService interface {
	CreateEvent(ctx context.Context, event *event_dto.EventAdd, files []*multipart.FileHeader, creatorId int) (int, error)
	GetEventById(ctx context.Context, eventId int) (*event_dto.EventResponse, int, error)
	GetListOfEvents(ctx context.Context, page int, limit int, userId int, evType string,
		city string, sort string, order string, subscribedOnly string) (*[]event_dto.EventResponse, int, error)
}

func NewEventHandler(log *slog.Logger, service EventService, conf config.Config) *EventHandler {
	log.Info("New Event Handler Run Successfully")
	return &EventHandler{
		log:     log,
		service: service,
		config:  conf,
	}
}

func (eh *EventHandler) CreateEvent(ctx *gin.Context) {
	var event event_dto.EventAdd
	if err := ctx.ShouldBind(&event); err != nil {
		ctx.JSON(400, gin.H{"error": "validation error"})
		return
	}
	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.JSON(400, gin.H{"error": "validation error"})
		return
	}
	files := form.File["media"]
	userId := ctx.GetInt("userId")
	code, err := eh.service.CreateEvent(ctx, &event, files, userId)
	if err != nil {
		ctx.JSON(code, gin.H{"error": err.Error()})
	}
	ctx.JSON(200, gin.H{})
}

func (eh *EventHandler) GetEvent(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(400, gin.H{"error": "validation error"})
		return
	}
	event, code, err := eh.service.GetEventById(ctx, id)
	if err != nil {
		ctx.JSON(code, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(200, event)
}

func (eh *EventHandler) GetListOfEvents(ctx *gin.Context) {
	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil {
		ctx.JSON(400, gin.H{"error": "validation error"})
		return
	}
	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	if err != nil {
		ctx.JSON(400, gin.H{"error": "validation error"})
		return
	}
	if limit > 50 || limit < 1 {
		ctx.JSON(400, gin.H{"error": "validation error"})
		return
	}
	userId, err := strconv.Atoi(ctx.DefaultQuery("userId", "0"))
	if err != nil {
		ctx.JSON(400, gin.H{"error": "validation error"})
		return
	}
	evType := ctx.DefaultQuery("type", "")
	city := ctx.DefaultQuery("city", "")
	sort := ctx.DefaultQuery("sort", "")
	order := ctx.DefaultQuery("order", "")
	subscribedOnly := ctx.DefaultQuery("subscribed_only", "false")
	events, code, err := eh.service.GetListOfEvents(ctx, page, limit, userId, evType, city, sort, order, subscribedOnly)
	if err != nil {
		ctx.JSON(code, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(200, events)
}
