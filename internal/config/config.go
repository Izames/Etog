package config

import (
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Port     string   `yaml:"port" env-default:"8080"`
	Db       Database `yaml:"db_conf"`
	Env      string   `yaml:"env"`
	JWTKey   string   `yaml:"jwt_key"`
	MailData MailData `yaml:"mail_data"`
	RedisDb  RedisDb  `yaml:"redis_db"`
}

type Database struct {
	Host    string `yaml:"host" env-required:"true"`
	User    string `yaml:"user" env-required:"true"`
	Pass    string `yaml:"password" env-required:"true"`
	Dbname  string `yaml:"dbname" env-required:"true"`
	Port    string `yaml:"port" env-default:"5432"`
	Sslmode string `yaml:"sslmode" env-required:"true"`
}

type MailData struct {
	From     string `yaml:"from" env-required:"true"`
	Host     string `yaml:"host" env-required:"true"`
	Port     string `yaml:"port" env-required:"true"`
	Password string `yaml:"password" env-required:"true"`
}

type RedisDb struct {
	Host     string `yaml:"host" env-required:"true"`
	Port     string `yaml:"port" env-default:"6379"`
	Password string `yaml:"password" env-required:"true"`
	User     string `yaml:"user" env-required:"true"`
	Index    string `yaml:"index" env-required:"true"`
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
