package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"os"
)

type Config struct {
	Port string   `yaml:"port" env-default:"8080"`
	Db   Database `yaml:"db_conf"`
}

type Database struct {
	Host    string `yaml:"host" env-required:"true"`
	User    string `yaml:"user" env-required:"true"`
	Pass    string `yaml:"password" env-required:"true"`
	Dbname  string `yaml:"dbname" env-required:"true"`
	Port    string `yaml:"port" env-default:"5432"`
	Sslmode string `yaml:"sslmode" env-required:"true"`
}

func MustLoad() Config {
	path := "./config/dev.yaml"
	if _, err := os.Stat(path); err != nil {
		panic(err)
	}
	var config Config
	if err := cleanenv.ReadConfig(path, &config); err != nil {
		panic(err)
	}
	return config
}
