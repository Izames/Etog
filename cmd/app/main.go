package main

import (
	"Etog/internal/config"
	"Etog/internal/http-server/handlers"
	"Etog/internal/http-server/services"
	jwt2 "Etog/internal/lib/jwt"
	slog2 "Etog/internal/lib/slog"
	"Etog/internal/worker"
	"Etog/storage/psql"
	"Etog/storage/redis"
	s4 "Etog/storage/s3"
	"context"
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
	conf := config.MustLoad()
	log := NewLogger(conf.Env)
	database := psql.New(fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		conf.Db.User,
		conf.Db.Pass,
		conf.Db.Host,
		conf.Db.Port,
		conf.Db.Dbname,
		conf.Db.Sslmode), log)
	redisDb := redis.NewRedisDb(conf.RedisDb, log)
	jwt := jwt2.NewJwtLib(conf.JWTKey, database)
	s3, err := s4.NewS3(&conf.S3)
	if err != nil {
		log.Error("Failed to create S3 object")
		os.Exit(1)
	}
	authService := services.NewAuthService(database, redisDb, log, &conf.MailData, jwt, s3)
	accountHandler := handlers.NewAccountHandler(log, authService)
	mockHandler := handlers.NewMockEventHandler(log, database)

	eventService := services.NewEventService(log, database, s3)
	eventHandler := handlers.NewEventHandler(log, eventService, conf)

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
	rAccount := router.Group("/account")
	rEvent := router.Group("/event")

	rMockEv.POST("/add", mockHandler.CreateMockEvent)
	rMockEv.GET("/get/:id", mockHandler.GetMockEvent)
	rMockEv.GET("/get", mockHandler.GetMockEvents)
	rMockEv.PATCH("/update/:id", mockHandler.UpdateMockEvent)
	rMockEv.DELETE("/delete/:id", mockHandler.DeleteMockEvent)

	rAccount.POST("/register", accountHandler.Registration)
	rAccount.POST("/auth", accountHandler.Authenticate)
	rAccount.POST("/sendCode", accountHandler.SendCode)
	rAccount.POST("/confirmCode", accountHandler.ConfirmCode)
	rAccount.GET("/getAccount/:query", accountHandler.GetAccountByLoginOrId)
	rAccount.POST("/changepassword", accountHandler.ChangePassword)
	rAccount.POST("/confirmchangepassword", accountHandler.ConfirmChangePassword)
	rAccount.POST("/changeemail", accountHandler.ChangeMail)
	rAccount.POST("confirmchangeemail", accountHandler.ConfirmMail)
	rAccount.Use(jwt.JWTAuth())
	rAccount.POST("/changeData", accountHandler.ChangeData)
	rAccount.DELETE("/delete", accountHandler.DeleteAccount)
	rAccount.POST("/follow/:int", accountHandler.Subscribe)
	rAccount.POST("/unfollow/:int", accountHandler.Unsubscribe)
	rAccount.POST("/requestverification", accountHandler.RequestOfficial)
	rAccount.DELETE("/deletesessions", accountHandler.DeleteSessions)

	rEvent.POST("/create", eventHandler.CreateEvent)

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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	cleaner := worker.NewCleaner(database.ReturnDb(), log, time.Hour*2)
	cleaner.Run(ctx)

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("error server closing:", err)
	} else {
		log.Info("Server gracefully shutdown")
	}
}

func NewLogger(env string) *slog.Logger {
	level := slog.LevelDebug
	switch env {
	case "dev":
		level = slog.LevelDebug
	case "prod":
		level = slog.LevelInfo
	}
	log := slog.New(slog2.NewHandler(level))
	log.Info("Logger initialized successfully\n")
	return log
}
