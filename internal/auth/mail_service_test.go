package auth

import "testing"

func TestNormalizeSMTPPasswordRemovesGmailDisplaySpaces(t *testing.T) {
	got := normalizeSMTPPassword("smtp.gmail.com", "abcd efgh ijkl mnop")
	if got != "abcdefghijklmnop" {
		t.Fatalf("normalized password = %q", got)
	}
}

func TestNormalizeSMTPPasswordPreservesNonGmailInternalSpaces(t *testing.T) {
	got := normalizeSMTPPassword("smtp.example.test", "  secret phrase  ")
	if got != "secret phrase" {
		t.Fatalf("normalized password = %q", got)
	}
}
