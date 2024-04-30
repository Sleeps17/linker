package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
)

const (
	configPathEnv = "CONFIG_PATH"
)

type Config struct {
	Env      string         `yaml:"env" env-default:"local"`
	Server   ServerConfig   `yaml:"server"`
	DataBase DataBaseConfig `yaml:"data_base"`
}

type ServerConfig struct {
	Host string `yaml:"host" env-default:"127.0.0.1"`
	Port uint   `yaml:"port" env-default:"4404"`
}

type DataBaseConfig struct {
	ConnString string `yaml:"conn_string" env-required:"true"`
	DbName     string `yaml:"db_name" env-default:"linker"`
	Collection string `yaml:"collection" env-default:"links"`
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