package config

import (
	"errors"
	"fmt"
	"net/mail"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	SMTP     SMTPConfig
	Holiday  HolidayConfig
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

type HolidayConfig struct {
	BaseURL string
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
			User:     env("DB_USER", "otista_app"),
			Password: env("DB_PASSWORD", ""),
			Name:     env("DB_NAME", "rsud_otista"),
		},
		SMTP: SMTPConfig{
			Host:     env("SMTP_HOST", ""),
			Port:     env("SMTP_PORT", ""),
			Email:    env("SMTP_EMAIL", ""),
			Password: env("SMTP_PASSWORD", ""),
		},
		Holiday: HolidayConfig{
			BaseURL: env("HOLIDAY_API_BASE_URL", "https://tanggalmerah.upset.dev"),
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

func (c Config) ValidateDatabase() error {
	if !strings.EqualFold(strings.TrimSpace(c.App.Env), "production") {
		return nil
	}

	if strings.TrimSpace(c.Database.Host) == "" ||
		strings.TrimSpace(c.Database.Port) == "" ||
		strings.TrimSpace(c.Database.User) == "" ||
		strings.TrimSpace(c.Database.Password) == "" ||
		strings.TrimSpace(c.Database.Name) == "" {
		return errors.New("production database configuration is incomplete")
	}
	if isPlaceholderSecret(c.Database.Password) {
		return errors.New("production database password is still a placeholder")
	}
	if _, err := strconv.Atoi(strings.TrimSpace(c.Database.Port)); err != nil {
		return errors.New("production database port is invalid")
	}
	return nil
}

func (c Config) ValidateRuntime() error {
	if err := c.ValidateDatabase(); err != nil {
		return err
	}
	if !strings.EqualFold(strings.TrimSpace(c.App.Env), "production") {
		return nil
	}
	if strings.EqualFold(strings.TrimSpace(c.Database.User), "root") {
		return errors.New("production runtime database user must not be root")
	}

	if strings.TrimSpace(c.SMTP.Host) == "" ||
		strings.TrimSpace(c.SMTP.Port) == "" ||
		strings.TrimSpace(c.SMTP.Email) == "" ||
		strings.TrimSpace(c.SMTP.Password) == "" {
		return errors.New("production SMTP configuration is incomplete")
	}
	port, err := strconv.Atoi(strings.TrimSpace(c.SMTP.Port))
	if err != nil || port < 1 || port > 65535 {
		return errors.New("production SMTP port is invalid")
	}
	address, err := mail.ParseAddress(strings.TrimSpace(c.SMTP.Email))
	if err != nil || address.Name != "" || address.Address != strings.TrimSpace(c.SMTP.Email) {
		return errors.New("production SMTP email is invalid")
	}
	if isPlaceholderSecret(c.SMTP.Password) {
		return errors.New("production SMTP password is still a placeholder")
	}
	return nil
}

func isPlaceholderSecret(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return value == "" || strings.Contains(value, "change-this") ||
		strings.Contains(value, "password_database") ||
		strings.Contains(value, "google-app-password") ||
		strings.Contains(value, "google_app_password") ||
		strings.HasPrefix(value, "your-") ||
		(strings.HasPrefix(value, "<") && strings.HasSuffix(value, ">")) ||
		value == "password"
}

func env(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
