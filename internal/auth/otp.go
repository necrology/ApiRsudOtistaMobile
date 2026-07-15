package auth

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"golang.org/x/crypto/bcrypt"
)

const otpUpperBound = 1_000_000

// GenerateOTP returns a six digit cryptographically random OTP. Random-source
// failures are returned to the caller; a predictable fallback must never be
// used for an authentication challenge.
func GenerateOTP() (string, error) {
	value, err := rand.Int(rand.Reader, big.NewInt(otpUpperBound))
	if err != nil {
		return "", fmt.Errorf("generate otp: %w", err)
	}

	return fmt.Sprintf("%06d", value.Int64()), nil
}

func hashOTP(otp string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(otp), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash otp: %w", err)
	}

	return string(hash), nil
}
