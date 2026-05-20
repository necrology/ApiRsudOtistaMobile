package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	SMTP     SMTPConfig
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

type SMTPConfig struct {
	Host     string
	Port     string
	Email    string
	Password string
}

func Load() Config {
	_ = godotenv.Load()

	return Config{
		App: AppConfig{
			Name: env("APP_NAME", "ApiRsudOtistaMobile"),
			Env:  env("APP_ENV", "development"),
			Port: env("APP_PORT", "8081"),
		},
		Database: DatabaseConfig{
			Host:     env("DB_HOST", "127.0.0.1"),
			Port:     env("DB_PORT", "3307"),
			User:     env("DB_USER", "root"),
			Password: env("DB_PASSWORD", "qwerty@123"),
			Name:     env("DB_NAME", "rsud_otista"),
		},
		SMTP: SMTPConfig{
			Host:     env("SMTP_HOST", ""),
			Port:     env("SMTP_PORT", ""),
			Email:    env("SMTP_EMAIL", ""),
			Password: env("SMTP_PASSWORD", ""),
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
