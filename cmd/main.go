package main

import (
	"Etog/cmd/migrator"
	"Etog/internal/config"
	"flag"
	"fmt"
	"log/slog"
	"os"
)

func main() {
	var migrationPath string
	flag.StringVar(&migrationPath, "migrations-path", "", "Path to migration folder")
	flag.Parse()
	conf := config.MustLoad()
	log := NewLogger()
	migr := migrator.NewMigrator("file://"+migrationPath, fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=%s",
		conf.Db.User, conf.Db.Pass, conf.Db.Host, conf.Db.Port, conf.Db.Dbname, conf.Db.Sslmode))
	migr.MigrateUp()
	log.Info("migrates up")
	//TODO: разработать базу данных при помощи миграций
	//TODO: сделать обращения к базе данных
	//TODO: создать запросы для апи
	//TODO: создать бизнес-логику для выполнения АПИ
	//TODO: связать все
	//TODO: изучить тестирование и пройтись им по всему проекту
}

func NewLogger() *slog.Logger {
	var log *slog.Logger
	slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	return log
}
