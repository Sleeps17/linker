package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"time"
)

const (
	configPathEnv = "CONFIG_PATH"
)

type Config struct {
	Env                string                   `yaml:"env" env-default:"local"`
	Server             ServerConfig             `yaml:"server"`
	DataBase           PostgresDBConfig         `yaml:"data_base"`
	UrlShortenerClient UrlShortenerClientConfig `yaml:"url_shortener_client"`
}

type ServerConfig struct {
	Port    uint          `yaml:"port" env-default:"4404"`
	Timeout time.Duration `yaml:"timeout" env-default:"5s"`
}

type MongoDBConfig struct {
	Timeout    time.Duration `yaml:"timeout" env-default:"5s"`
	ConnString string        `yaml:"conn_string" env-required:"true"`
	DbName     string        `yaml:"db_name" env-default:"linker"`
	Collection string        `yaml:"collection" env-default:"links"`
}

type PostgresDBConfig struct {
	Timeout  time.Duration `yaml:"timeout" env-default:"10s"`
	Host     string        `yaml:"host" env-default:"db"`
	Port     string        `yaml:"port" env-default:"5432"`
	Name     string        `yaml:"name" env-required:"true"`
	Username string        `yaml:"username" env-required:"true"`
	Password string        `yaml:"password" env-required:"true"`
}

type UrlShortenerClientConfig struct {
	Host     string `yaml:"host" env-default:"localhost"`
	Port     string `yaml:"port" env-default:"8080"`
	Username string `yaml:"username" env-default:"pasha"`
	Password string `yaml:"password" env-default:"1234"`
}

func MustLoad() *Config {
	configPath := os.Getenv(configPathEnv)

	if configPath == "" {
		panic("CONFIG_PATH is not set")
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic(fmt.Sprintf("Failed parse config: %v", err))
	}

	return &cfg
}

func MustLoadByPath(configPath string) *Config {
	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic(fmt.Sprintf("Failed parse config: %v", err))
	}

	return &cfg
}
