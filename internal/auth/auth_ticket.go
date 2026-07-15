package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	registrationTicketPurpose = "registration_set_password"
	registrationTicketTTL     = 10 * time.Minute
	authTicketRandomBytes     = 32
)

type RegistrationTicket struct {
	Token     string
	ExpiresAt time.Time
}

func newRegistrationTicket() (*RegistrationTicket, [sha256.Size]byte, error) {
	random := make([]byte, authTicketRandomBytes)
	if _, err := rand.Read(random); err != nil {
		return nil, [sha256.Size]byte{}, fmt.Errorf("generate registration ticket: %w", err)
	}

	token := "rsm_reg_" + base64.RawURLEncoding.EncodeToString(random)
	hash := sha256.Sum256([]byte(token))
	return &RegistrationTicket{
		Token:     token,
		ExpiresAt: time.Now().Add(registrationTicketTTL),
	}, hash, nil
}

func hashPresentedAuthTicket(token string) ([sha256.Size]byte, error) {
	token = strings.TrimSpace(token)
	const prefix = "rsm_reg_"
	if !strings.HasPrefix(token, prefix) {
		return [sha256.Size]byte{}, errors.New("invalid registration ticket")
	}

	random, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(token, prefix))
	if err != nil || len(random) != authTicketRandomBytes {
		return [sha256.Size]byte{}, errors.New("invalid registration ticket")
	}

	return sha256.Sum256([]byte(token)), nil
}
