package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
}

type AppConfig struct {
	Name string
	Env  string
	Port string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

func Load() Config {
	_ = godotenv.Load()

	return Config{
		App: AppConfig{
			Name: env("APP_NAME", "ApiRsudOtistaMobile"),
			Env:  env("APP_ENV", "development"),
			Port: env("APP_PORT", "8080"),
		},
		Database: DatabaseConfig{
			Host:     env("DB_HOST", "127.0.0.1"),
			Port:     env("DB_PORT", "3306"),
			User:     env("DB_USER", "root"),
			Password: env("DB_PASSWORD", ""),
			Name:     env("DB_NAME", "apirusdotistamobile"),
		},
	}
}

func (c DatabaseConfig) DSN() string {
	auth := c.User
	if c.Password != "" {
		auth = fmt.Sprintf("%s:%s", c.User, c.Password)
	}

	return fmt.Sprintf(
		"%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		auth,
		c.Host,
		c.Port,
		c.Name,
	)
}

func env(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
