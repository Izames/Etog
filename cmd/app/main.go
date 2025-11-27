package main

import (
	"Etog/internal/config"
	"Etog/internal/http-server/handlers"
	"Etog/internal/lib/slog/handler"
	"Etog/storage/psql"
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	var migrationPath string
	flag.StringVar(&migrationPath, "migrations-path", "", "Path to migration folder")

	flag.Parse()
	conf := config.MustLoad()
	log := NewLogger(conf.Env)
	log.Info("sss")
	database := psql.New(fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		conf.Db.User,
		conf.Db.Pass,
		conf.Db.Host,
		conf.Db.Port,
		conf.Db.Dbname,
		conf.Db.Sslmode))
	mockHandler := handlers.NewMockEventHandler(log, database)
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: false, // временно отключите
		MaxAge:           12 * time.Hour,
	}))

	rMockEv := router.Group("/event")
	rMockEv.POST("/add", mockHandler.CreateMockEvent)
	rMockEv.GET("/get/:id", mockHandler.GetMockEvent)
	rMockEv.GET("/get", mockHandler.GetMockEvents)
	rMockEv.PATCH("/update/:id", mockHandler.UpdateMockEvent)
	rMockEv.DELETE("/delete/:id", mockHandler.DeleteMockEvent)

	server := &http.Server{
		Addr:           fmt.Sprintf(":%s", conf.Port),
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Error("error server closing:", err)
	} else {
		log.Info("Server gracefully shutdown")
	}

	//TODO: Сделать документацию для конфига
	//TODO: сделать обращения к базе данных
	//TODO: создать запросы для апи
	//TODO: создать бизнес-логику для выполнения АПИ
	//TODO: связать все
	//TODO: изучить тестирование и пройтись им по всему проекту
}

func NewLogger(env string) *slog.Logger {
	level := slog.LevelDebug
	switch env {
	case "dev":
		level = slog.LevelDebug
	case "prod":
		level = slog.LevelInfo
	}
	return slog.New(handler.NewHandler(level))
}
