package services

import (
	"Etog/internal/domain/entity"
	"Etog/internal/domain/entity/DTO/event_dto"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"mime/multipart"

	"github.com/pgvector/pgvector-go"
)

type EventService struct {
	log     *slog.Logger
	Storage Storage
	S3      S3
}

type Storage interface {
	AddEvent(ctx context.Context, event *entity.Event, contact *entity.Contact) error
}

type S3 interface {
	Delete(url string) error
	Upload(fileHeader *multipart.FileHeader, pathS3 string) (string, error)
}

func NewEventService(log *slog.Logger, storage Storage, s3 S3) *EventService {
	if log == nil {
		log = slog.Default()
	}
	return &EventService{log: log, Storage: storage, S3: s3}
}

func (es *EventService) CreateEvent(ctx context.Context, event *event_dto.EventAdd, files []*multipart.FileHeader, creatorId int) (int, error) {
	const op = "service.CreateEvent"
	log := es.log.With(slog.String("op", op))
	var urls []string
	for _, file := range files {
		url, err := es.S3.Upload(file, "event-media")
		if err != nil {
			log.Error(err.Error())
			for _, url := range urls {
				err = es.S3.Delete(url)
				if err != nil {
					log.Error("Failed to delete url: " + url)
				}
			}
			return 500, errors.New("server error")
		}
		urls = append(urls, url)
	}

	jsondata, err := json.Marshal(urls)
	if err != nil {
		log.Error(err.Error())
		return 500, errors.New("server error")
	}

	contact := entity.Contact{
		Phone:    event.Phone,
		Mail:     event.Mail,
		Telegram: event.Telegram,
		Deleted:  false,
	}

	cords := pgvector.NewVector([]float32{
		event.Latitude,
		event.Longitude,
	})
	eventDb := entity.Event{
		Organizer:   creatorId,
		Name:        event.Name,
		Price:       event.Price,
		Address:     event.Address,
		City:        event.City,
		Cords:       cords,
		Type:        event.Type,
		Date:        event.Date,
		Description: event.Description,
		Passed:      event.Passed,
		MaxPeople:   event.MaxPeople,
		Media:       jsondata,
		Deleted:     false,
	}
	err = es.Storage.AddEvent(ctx, &eventDb, &contact)
	if err != nil {
		for _, url := range urls {
			if deleteErr := es.S3.Delete(url); deleteErr != nil {
				log.Error(
					"failed to delete S3 file",
					slog.String("url", url),
					slog.String("error", deleteErr.Error()),
				)
			}
		}

		return 500, errors.New("server error")
	}
	return 200, nil
}
