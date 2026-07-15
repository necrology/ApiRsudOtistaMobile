package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"apirusdotistamobile/internal/config"
)

func TestOTPEmailDispatcherIsBounded(t *testing.T) {
	cfg := config.SMTPConfig{
		Host: "smtp.example.test", Port: "587", Email: "sender@example.test", Password: "secret",
	}
	dispatcher := newOTPEmailDispatcher(cfg, 0, 1)

	if err := dispatcher.Enqueue("one@example.test", "123456", "subject", "purpose"); err != nil {
		t.Fatal(err)
	}
	if err := dispatcher.Enqueue("two@example.test", "123456", "subject", "purpose"); !errors.Is(err, ErrMailQueueFull) {
		t.Fatalf("second enqueue error = %v, want ErrMailQueueFull", err)
	}
}

func TestOTPEmailDispatcherDrainsAndRejectsAfterClose(t *testing.T) {
	cfg := config.SMTPConfig{
		Host: "smtp.example.test", Port: "587", Email: "sender@example.test", Password: "secret",
	}
	delivered := make(chan string, 1)
	dispatcher := newOTPEmailDispatcherWithSender(
		cfg,
		1,
		1,
		func(_ config.SMTPConfig, to string, _ string, _ string, _ string) error {
			delivered <- to
			return nil
		},
	)

	if err := dispatcher.Enqueue("one@example.test", "123456", "subject", "purpose"); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := dispatcher.Close(ctx); err != nil {
		t.Fatalf("close failed: %v", err)
	}
	if got := <-delivered; got != "one@example.test" {
		t.Fatalf("delivered to %q", got)
	}
	if err := dispatcher.Enqueue("two@example.test", "123456", "subject", "purpose"); !errors.Is(err, ErrMailUnavailable) {
		t.Fatalf("enqueue after close error = %v, want ErrMailUnavailable", err)
	}
}

func TestOTPEmailDispatcherRejectsIncompleteConfig(t *testing.T) {
	dispatcher := newOTPEmailDispatcher(config.SMTPConfig{}, 0, 1)
	if err := dispatcher.Enqueue("one@example.test", "123456", "subject", "purpose"); !errors.Is(err, ErrMailUnavailable) {
		t.Fatalf("enqueue error = %v, want ErrMailUnavailable", err)
	}
}
