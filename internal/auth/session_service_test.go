package auth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"apirusdotistamobile/internal/model"
	"apirusdotistamobile/internal/repository"
)

type fakeSessionStore struct {
	created       model.SessionUserMobile
	principal     *model.SessionPrincipal
	findHash      []byte
	rotateInput   []byte
	replacement   model.SessionUserMobile
	rotateErr     error
	logoutHash    []byte
	revokedFamily string
	revokedUser   int64
}

func (f *fakeSessionStore) CreateSession(_ context.Context, session model.SessionUserMobile) (*model.SessionUserMobile, error) {
	f.created = session
	session.ID = 91
	return &session, nil
}

func (f *fakeSessionStore) FindPrincipalByAccessTokenHash(_ context.Context, hash []byte, _ time.Time) (*model.SessionPrincipal, error) {
	f.findHash = append([]byte(nil), hash...)
	if f.principal == nil {
		return nil, repository.ErrSessionNotFound
	}
	return f.principal, nil
}

func (f *fakeSessionStore) RotateRefreshToken(_ context.Context, hash []byte, replacement model.SessionUserMobile, _ time.Time) (*model.SessionUserMobile, error) {
	f.rotateInput = append([]byte(nil), hash...)
	f.replacement = replacement
	if f.rotateErr != nil {
		return nil, f.rotateErr
	}
	replacement.ID = 92
	replacement.UserID = 7
	replacement.FamilyID = "family"
	return &replacement, nil
}

func (f *fakeSessionStore) RevokeFamilyByAccessTokenHash(_ context.Context, hash []byte, _ string, _ time.Time) error {
	f.logoutHash = append([]byte(nil), hash...)
	return nil
}

func (f *fakeSessionStore) RevokeFamily(_ context.Context, familyID string, _ string, _ time.Time) error {
	f.revokedFamily = familyID
	return nil
}

func (f *fakeSessionStore) RevokeUser(_ context.Context, userID int64, _ string, _ time.Time) error {
	f.revokedUser = userID
	return nil
}

func TestGenerateOpaqueTokenUses256BitsAndSHA256(t *testing.T) {
	rawToken, hash, err := generateOpaqueToken()
	if err != nil {
		t.Fatal(err)
	}

	decoded, err := base64.RawURLEncoding.DecodeString(rawToken)
	if err != nil {
		t.Fatalf("token is not base64url: %v", err)
	}
	if len(decoded) != 32 {
		t.Fatalf("token entropy bytes = %d, want 32", len(decoded))
	}

	wantHash := sha256.Sum256([]byte(rawToken))
	if string(hash) != string(wantHash[:]) {
		t.Fatal("stored token hash is not SHA-256 of the presented token")
	}
	if string(hash) == rawToken {
		t.Fatal("raw token was stored instead of a hash")
	}
}

func TestNewSessionServiceRejectsTypedNilStore(t *testing.T) {
	var store *fakeSessionStore
	_, err := NewSessionServiceWithConfig(store, DefaultSessionConfig())
	if !errors.Is(err, ErrInvalidSessionConfig) {
		t.Fatalf("constructor error = %v, want ErrInvalidSessionConfig", err)
	}
}

func TestIssuePassesOnlyHashesToStore(t *testing.T) {
	store := &fakeSessionStore{}
	service := NewSessionService(store)
	fixedNow := time.Date(2026, 7, 15, 8, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return fixedNow }

	tokens := deterministicTokens(0x11, 0x22, 0x33)
	service.generateToken = func() (string, []byte, error) {
		token := tokens[0]
		tokens = tokens[1:]
		hash := sha256.Sum256([]byte(token))
		return token, hash[:], nil
	}

	pair, err := service.Issue(context.Background(), 7)
	if err != nil {
		t.Fatal(err)
	}

	if pair.AccessToken != deterministicToken(0x22) || pair.RefreshToken != deterministicToken(0x33) {
		t.Fatal("issued token pair does not contain the generated raw tokens")
	}
	if store.created.FamilyID != deterministicToken(0x11) {
		t.Fatal("session family was not generated independently")
	}
	if string(store.created.AccessTokenHash) == pair.AccessToken || string(store.created.RefreshTokenHash) == pair.RefreshToken {
		t.Fatal("raw tokens reached persistent session fields")
	}
	if !store.created.AccessExpiresAt.Equal(fixedNow.Add(DefaultAccessTokenTTL)) {
		t.Fatalf("unexpected access expiry: %s", store.created.AccessExpiresAt)
	}
	if !store.created.RefreshExpiresAt.Equal(fixedNow.Add(DefaultRefreshTokenTTL)) {
		t.Fatalf("unexpected refresh expiry: %s", store.created.RefreshExpiresAt)
	}
}

func TestAuthenticateHashesPresentedToken(t *testing.T) {
	rawToken := deterministicToken(0x44)
	principal := &model.SessionPrincipal{SessionID: 5, UserID: 7, PatientID: 9, Email: "patient@example.test", NoRM: "00123"}
	store := &fakeSessionStore{principal: principal}
	service := NewSessionService(store)

	got, err := service.Authenticate(context.Background(), rawToken)
	if err != nil {
		t.Fatal(err)
	}
	if got != principal {
		t.Fatal("unexpected principal")
	}
	wantHash := sha256.Sum256([]byte(rawToken))
	if string(store.findHash) != string(wantHash[:]) {
		t.Fatal("store did not receive the access-token hash")
	}
}

func TestRefreshMapsReuseAndReliesOnAtomicStoreRevocation(t *testing.T) {
	store := &fakeSessionStore{rotateErr: repository.ErrRefreshTokenReused}
	service := NewSessionService(store)

	_, err := service.Refresh(context.Background(), deterministicToken(0x55))
	if !errors.Is(err, ErrRefreshTokenReused) {
		t.Fatalf("refresh error = %v, want ErrRefreshTokenReused", err)
	}
}

func TestRefreshMapsStableFamilyCooldown(t *testing.T) {
	store := &fakeSessionStore{rotateErr: repository.ErrRefreshTooSoon}
	service := NewSessionService(store)

	_, err := service.Refresh(context.Background(), deterministicToken(0x55))
	if !errors.Is(err, ErrRefreshTooSoon) {
		t.Fatalf("refresh error = %v, want ErrRefreshTooSoon", err)
	}
}

func TestRefreshCreatesFreshHashOnlyReplacement(t *testing.T) {
	store := &fakeSessionStore{}
	service := NewSessionService(store)
	fixedNow := time.Date(2026, 7, 15, 8, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return fixedNow }

	tokens := deterministicTokens(0x66, 0x77)
	service.generateToken = func() (string, []byte, error) {
		token := tokens[0]
		tokens = tokens[1:]
		hash := sha256.Sum256([]byte(token))
		return token, hash[:], nil
	}

	pair, err := service.Refresh(context.Background(), deterministicToken(0x55))
	if err != nil {
		t.Fatal(err)
	}
	if pair.AccessToken != deterministicToken(0x66) || pair.RefreshToken != deterministicToken(0x77) {
		t.Fatal("refresh did not return the newly generated pair")
	}
	if string(store.replacement.AccessTokenHash) == pair.AccessToken || string(store.replacement.RefreshTokenHash) == pair.RefreshToken {
		t.Fatal("raw replacement token reached the store")
	}
	wantOldHash := sha256.Sum256([]byte(deterministicToken(0x55)))
	if string(store.rotateInput) != string(wantOldHash[:]) {
		t.Fatal("old refresh token was not hashed before rotation")
	}
}

func deterministicTokens(bytes ...byte) []string {
	tokens := make([]string, 0, len(bytes))
	for _, value := range bytes {
		tokens = append(tokens, deterministicToken(value))
	}
	return tokens
}

func deterministicToken(value byte) string {
	data := make([]byte, opaqueTokenBytes)
	for index := range data {
		data[index] = value
	}
	return base64.RawURLEncoding.EncodeToString(data)
}
