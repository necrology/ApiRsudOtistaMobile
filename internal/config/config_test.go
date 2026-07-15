package config

import "testing"

func TestProductionConfigRejectsRootAndMissingSMTP(t *testing.T) {
	cfg := Config{
		App: AppConfig{Env: "production"},
		Database: DatabaseConfig{
			Host: "db", Port: "3306", User: "root", Password: "a-real-secret", Name: "simrs",
		},
	}
	if err := cfg.ValidateRuntime(); err == nil {
		t.Fatal("production root database user was accepted")
	}

	cfg.Database.User = "otista_app"
	if err := cfg.ValidateRuntime(); err == nil {
		t.Fatal("production runtime without SMTP was accepted")
	}
}

func TestProductionMigrationConfigMayUsePrivilegedAccount(t *testing.T) {
	cfg := Config{
		App: AppConfig{Env: "production"},
		Database: DatabaseConfig{
			Host: "db", Port: "3306", User: "root", Password: "a-real-secret", Name: "simrs",
		},
	}
	if err := cfg.ValidateDatabase(); err != nil {
		t.Fatalf("valid migration database config rejected: %v", err)
	}
}

func TestProductionConfigAcceptsLeastPrivilegeRuntime(t *testing.T) {
	cfg := Config{
		App: AppConfig{Env: "production"},
		Database: DatabaseConfig{
			Host: "db", Port: "3306", User: "otista_app", Password: "a-real-secret", Name: "simrs",
		},
		SMTP: SMTPConfig{
			Host: "smtp.example.test", Port: "587", Email: "sender@example.test", Password: "smtp-real-secret",
		},
	}
	if err := cfg.ValidateRuntime(); err != nil {
		t.Fatalf("valid production config rejected: %v", err)
	}
}

func TestProductionConfigRejectsDocumentedPlaceholderSecrets(t *testing.T) {
	for _, secret := range []string{
		"change-this-app-password",
		"password_database",
		"your-google-app-password",
		"google_app_password",
		"<secret-dari-secret-manager>",
		"<app-password-smtp>",
	} {
		if !isPlaceholderSecret(secret) {
			t.Fatalf("placeholder secret %q was accepted", secret)
		}
	}
}
