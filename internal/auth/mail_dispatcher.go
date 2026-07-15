package auth

import (
	"context"
	"errors"
	"log"
	"net/mail"
	"strconv"
	"strings"
	"sync"

	"apirusdotistamobile/internal/config"
)

const (
	defaultMailWorkers   = 2
	defaultMailQueueSize = 64
)

var (
	ErrMailQueueFull   = errors.New("otp email queue is full")
	ErrMailUnavailable = errors.New("otp email service is unavailable")
)

type otpEmailJob struct {
	to      string
	otp     string
	subject string
	purpose string
}

type otpEmailSender func(config.SMTPConfig, string, string, string, string) error

// OTPEmailDispatcher membatasi jumlah goroutine SMTP dan memori antrean.
// Job menyimpan OTP mentah hanya selama menunggu worker dan tidak pernah log.
type OTPEmailDispatcher struct {
	config config.SMTPConfig
	queue  chan otpEmailJob
	send   otpEmailSender

	mu      sync.RWMutex
	closed  bool
	workers sync.WaitGroup
	done    chan struct{}
}

func newOTPEmailDispatcher(cfg config.SMTPConfig, workers int, queueSize int) *OTPEmailDispatcher {
	return newOTPEmailDispatcherWithSender(cfg, workers, queueSize, SendOTPEmailWithPurpose)
}

func newOTPEmailDispatcherWithSender(
	cfg config.SMTPConfig,
	workers int,
	queueSize int,
	sender otpEmailSender,
) *OTPEmailDispatcher {
	if workers < 0 {
		workers = 0
	}
	if queueSize < 1 {
		queueSize = 1
	}

	dispatcher := &OTPEmailDispatcher{
		config: cfg,
		queue:  make(chan otpEmailJob, queueSize),
		send:   sender,
		done:   make(chan struct{}),
	}
	dispatcher.workers.Add(workers)
	for worker := 0; worker < workers; worker++ {
		go dispatcher.work()
	}
	go func() {
		dispatcher.workers.Wait()
		close(dispatcher.done)
	}()
	return dispatcher
}

func newProductionOTPEmailDispatcher(cfg config.SMTPConfig) *OTPEmailDispatcher {
	return newOTPEmailDispatcher(cfg, defaultMailWorkers, defaultMailQueueSize)
}

func (d *OTPEmailDispatcher) Enqueue(to string, otp string, subject string, purpose string) error {
	if d == nil || !validSMTPConfig(d.config) {
		return ErrMailUnavailable
	}
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.closed {
		return ErrMailUnavailable
	}

	job := otpEmailJob{to: to, otp: otp, subject: subject, purpose: purpose}
	select {
	case d.queue <- job:
		return nil
	default:
		return ErrMailQueueFull
	}
}

func (d *OTPEmailDispatcher) work() {
	defer d.workers.Done()
	for job := range d.queue {
		if err := d.send(
			d.config,
			job.to,
			job.otp,
			job.subject,
			job.purpose,
		); err != nil {
			log.Printf("send otp email failed error_type=%T", err)
		}
	}
}

// Close menghentikan penerimaan job baru, menutup antrean, lalu menunggu job
// yang sudah diakui Enqueue diproses. Context menjaga shutdown tetap terbatas.
func (d *OTPEmailDispatcher) Close(ctx context.Context) error {
	if d == nil {
		return nil
	}

	d.mu.Lock()
	if !d.closed {
		d.closed = true
		close(d.queue)
	}
	d.mu.Unlock()

	select {
	case <-d.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func validSMTPConfig(cfg config.SMTPConfig) bool {
	if strings.TrimSpace(cfg.Host) == "" ||
		strings.TrimSpace(cfg.Port) == "" ||
		strings.TrimSpace(cfg.Email) == "" ||
		strings.TrimSpace(cfg.Password) == "" {
		return false
	}

	port, err := strconv.Atoi(strings.TrimSpace(cfg.Port))
	if err != nil || port < 1 || port > 65535 {
		return false
	}
	address, err := mail.ParseAddress(strings.TrimSpace(cfg.Email))
	return err == nil && address.Name == "" && address.Address == strings.TrimSpace(cfg.Email)
}
