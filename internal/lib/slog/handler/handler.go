package handler

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/fatih/color"
)

type PrettyHandler struct {
	slog.Handler
	l *log.Logger
}

func NewHandler(level slog.Level) *PrettyHandler {
	return &PrettyHandler{
		Handler: slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		}),
		l: log.New(os.Stdout, "", 0),
	}
}

func (h *PrettyHandler) Handle(ctx context.Context, r slog.Record) error {
	level := r.Level.String()
	switch r.Level {
	case slog.LevelDebug:
		level = color.MagentaString(level)
	case slog.LevelInfo:
		level = color.BlueString(level)
	case slog.LevelWarn:
		level = color.YellowString(level)
	case slog.LevelError:
		level = color.RedString(level)
	}

	time := r.Time.Format("Jan _2 15:04:05")

	fmt.Printf("[%s] %s %s", time, level, r.Message)

	return nil

}
