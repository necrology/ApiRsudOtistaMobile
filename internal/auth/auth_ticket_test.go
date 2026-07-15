package auth

import (
	"strings"
	"testing"
)

func TestRegistrationTicketIsOpaqueAndStrictlyParsed(t *testing.T) {
	ticket, issuedHash, err := newRegistrationTicket()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(ticket.Token, "rsm_reg_") {
		t.Fatalf("unexpected ticket prefix: %q", ticket.Token)
	}

	presentedHash, err := hashPresentedAuthTicket(ticket.Token)
	if err != nil {
		t.Fatal(err)
	}
	if issuedHash != presentedHash {
		t.Fatal("presented ticket hash differs from issued hash")
	}

	for _, invalid := range []string{"", "rsm_reg_", "wrong_prefix", ticket.Token + "x"} {
		if _, err = hashPresentedAuthTicket(invalid); err == nil {
			t.Fatalf("invalid ticket %q was accepted", invalid)
		}
	}
}

func TestAuthInputValidationBoundaries(t *testing.T) {
	if !validOTPFormat("123456") || validOTPFormat("12345") || validOTPFormat("12345a") {
		t.Fatal("OTP format validation is not strict six-digit numeric")
	}
	if err := validatePassword("abc12345"); err != nil {
		t.Fatalf("valid password rejected: %v", err)
	}
	if err := validatePassword(" abc12345"); err == nil {
		t.Fatal("leading whitespace password accepted")
	}
	if err := validatePassword(strings.Repeat("a", 71) + "1b"); err == nil {
		t.Fatal("password above bcrypt 72-byte limit accepted")
	}
	if _, err := normalizeEmail("Patient <patient@example.test>"); err == nil {
		t.Fatal("display-name email accepted")
	}
}
