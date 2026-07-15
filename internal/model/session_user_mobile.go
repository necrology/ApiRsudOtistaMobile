package model

import "time"

// SessionUserMobile represents one generation in an opaque-token session
// family. Only SHA-256 token hashes are persisted in this record; raw tokens
// exist only in SessionTokenPair and are returned to the client once.
type SessionUserMobile struct {
	ID                  int64
	FamilyID            string
	ParentSessionID     int64
	UserID              int64
	AccessTokenHash     []byte
	RefreshTokenHash    []byte
	AccessExpiresAt     time.Time
	RefreshExpiresAt    time.Time
	RotatedAt           *time.Time
	RevokedAt           *time.Time
	RevokeReason        string
	ReplacedBySessionID int64
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// SessionTokenPair is the one-time response issued to an authenticated mobile
// client. Callers must never log or persist these raw values on the API side.
type SessionTokenPair struct {
	TokenType        string    `json:"token_type"`
	AccessToken      string    `json:"access_token"`
	AccessExpiresAt  time.Time `json:"access_expires_at"`
	AccessExpiresIn  int64     `json:"access_expires_in"`
	RefreshToken     string    `json:"refresh_token"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
	RefreshExpiresIn int64     `json:"refresh_expires_in"`
}

// SessionPrincipal is the trusted identity derived from an access token. The
// patient attributes come from user_mobile, never from request query values.
type SessionPrincipal struct {
	SessionID       int64
	FamilyID        string
	UserID          int64
	PatientID       int64
	Email           string
	NoRM            string
	AccessExpiresAt time.Time
}

// HasLinkedPatient reports whether the authenticated account is linked to a
// patient record and therefore may access patient-scoped endpoints.
func (p SessionPrincipal) HasLinkedPatient() bool {
	return p.UserID > 0 && p.PatientID > 0
}
