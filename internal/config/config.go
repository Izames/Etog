package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"os"
)

type Config struct {
	Port string `yaml:"port" env-default:"8080"`
}

func MustLoad() Config {
	path := "./config/config.yaml"
	if err, _ := os.Stat(path); err != nil {
		panic(err)
	}
	var config Config
	if err := cleanenv.ReadConfig(path, &config); err != nil {
		panic(err)
	}
	return config
}
