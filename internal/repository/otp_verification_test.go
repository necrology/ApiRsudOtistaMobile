package repository

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestVerifyStoredOTPHashedAndLegacyCompatibility(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		stored   string
		provided string
		want     bool
	}{
		{name: "bcrypt valid", stored: string(hash), provided: "123456", want: true},
		{name: "bcrypt invalid", stored: string(hash), provided: "654321", want: false},
		{name: "legacy valid during transition", stored: "123456", provided: "123456", want: true},
		{name: "legacy invalid", stored: "123456", provided: "123457", want: false},
		{name: "blank rejected", stored: "", provided: "", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := verifyStoredOTP(test.stored, test.provided); got != test.want {
				t.Fatalf("verifyStoredOTP() = %v, want %v", got, test.want)
			}
		})
	}
}
