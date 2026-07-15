package repository

import (
	"testing"
	"time"

	"apirusdotistamobile/internal/model"
)

func TestInheritAbsoluteFamilyExpiryDoesNotExtendSession(t *testing.T) {
	now := time.Date(2026, 7, 15, 8, 0, 0, 0, time.UTC)
	familyExpiry := now.Add(5 * time.Minute)
	session := model.SessionUserMobile{
		AccessExpiresAt:  now.Add(15 * time.Minute),
		RefreshExpiresAt: now.Add(30 * 24 * time.Hour),
	}

	inheritAbsoluteFamilyExpiry(&session, familyExpiry)

	if !session.AccessExpiresAt.Equal(familyExpiry) {
		t.Fatalf("access expiry = %s, want family expiry %s", session.AccessExpiresAt, familyExpiry)
	}
	if !session.RefreshExpiresAt.Equal(familyExpiry) {
		t.Fatalf("refresh expiry = %s, want family expiry %s", session.RefreshExpiresAt, familyExpiry)
	}
	if err := validateSessionRecord(model.SessionUserMobile{
		FamilyID:         "family",
		UserID:           7,
		AccessTokenHash:  make([]byte, 32),
		RefreshTokenHash: make([]byte, 32),
		AccessExpiresAt:  session.AccessExpiresAt,
		RefreshExpiresAt: session.RefreshExpiresAt,
	}, true); err != nil {
		t.Fatalf("equal access/refresh expiry near family boundary must be valid: %v", err)
	}
}
