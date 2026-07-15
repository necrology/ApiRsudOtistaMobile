package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"apirusdotistamobile/internal/model"
	"apirusdotistamobile/internal/repository"
)

const (
	DefaultAccessTokenTTL  = 15 * time.Minute
	DefaultRefreshTokenTTL = 30 * 24 * time.Hour
	opaqueTokenBytes       = 32
	maximumOpaqueTokenSize = 512
)

var (
	ErrInvalidSessionUser   = errors.New("invalid mobile session user")
	ErrInvalidAccessToken   = errors.New("invalid or expired access token")
	ErrInvalidRefreshToken  = errors.New("invalid refresh token")
	ErrRefreshTokenExpired  = errors.New("refresh token expired")
	ErrRefreshTokenReused   = errors.New("refresh token reuse detected; session family revoked")
	ErrRefreshTooSoon       = errors.New("refresh token rotated too soon")
	ErrInvalidSessionConfig = errors.New("invalid mobile session configuration")
)

// SessionStore is deliberately hash-only: implementations must never receive
// or return a raw access or refresh token.
type SessionStore interface {
	CreateSession(context.Context, model.SessionUserMobile) (*model.SessionUserMobile, error)
	FindPrincipalByAccessTokenHash(context.Context, []byte, time.Time) (*model.SessionPrincipal, error)
	RotateRefreshToken(context.Context, []byte, model.SessionUserMobile, time.Time) (*model.SessionUserMobile, error)
	RevokeFamilyByAccessTokenHash(context.Context, []byte, string, time.Time) error
	RevokeFamily(context.Context, string, string, time.Time) error
	RevokeUser(context.Context, int64, string, time.Time) error
}

type SessionServiceConfig struct {
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

func DefaultSessionConfig() SessionServiceConfig {
	return SessionServiceConfig{
		AccessTokenTTL:  DefaultAccessTokenTTL,
		RefreshTokenTTL: DefaultRefreshTokenTTL,
	}
}

// SessionService issues and validates database-backed opaque tokens. The
// default constructor is suitable for production and always uses crypto/rand.
type SessionService struct {
	store         SessionStore
	accessTTL     time.Duration
	refreshTTL    time.Duration
	now           func() time.Time
	generateToken func() (string, []byte, error)
}

func NewSessionService(store SessionStore) *SessionService {
	service, err := NewSessionServiceWithConfig(store, DefaultSessionConfig())
	if err != nil {
		panic(err)
	}
	return service
}

func NewSessionServiceWithConfig(
	store SessionStore,
	config SessionServiceConfig,
) (*SessionService, error) {
	if isNilSessionStore(store) {
		return nil, fmt.Errorf("%w: store is required", ErrInvalidSessionConfig)
	}
	if config.AccessTokenTTL <= 0 {
		return nil, fmt.Errorf("%w: access token TTL must be positive", ErrInvalidSessionConfig)
	}
	if config.RefreshTokenTTL <= config.AccessTokenTTL {
		return nil, fmt.Errorf("%w: refresh token TTL must exceed access token TTL", ErrInvalidSessionConfig)
	}

	return &SessionService{
		store:         store,
		accessTTL:     config.AccessTokenTTL,
		refreshTTL:    config.RefreshTokenTTL,
		now:           time.Now,
		generateToken: generateOpaqueToken,
	}, nil
}

// Issue creates a new session family after the caller has completed all login
// factors. The raw token pair is returned once; only its hashes reach storage.
func (s *SessionService) Issue(
	ctx context.Context,
	userID int64,
) (*model.SessionTokenPair, error) {
	if userID <= 0 {
		return nil, ErrInvalidSessionUser
	}

	familyID, _, err := s.generateToken()
	if err != nil {
		return nil, err
	}
	accessToken, accessHash, err := s.generateToken()
	if err != nil {
		return nil, err
	}
	refreshToken, refreshHash, err := s.generateToken()
	if err != nil {
		return nil, err
	}

	now := s.now()
	session := model.SessionUserMobile{
		FamilyID:         familyID,
		UserID:           userID,
		AccessTokenHash:  accessHash,
		RefreshTokenHash: refreshHash,
		AccessExpiresAt:  now.Add(s.accessTTL),
		RefreshExpiresAt: now.Add(s.refreshTTL),
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	created, err := s.store.CreateSession(ctx, session)
	if err != nil {
		return nil, err
	}
	if created == nil {
		return nil, fmt.Errorf("%w: store returned an empty session", ErrInvalidSessionConfig)
	}

	return s.tokenPair(accessToken, refreshToken, created.AccessExpiresAt, created.RefreshExpiresAt, now), nil
}

// Authenticate resolves all trusted patient identity attributes from the
// database session and user record, not from user-controlled query parameters.
func (s *SessionService) Authenticate(
	ctx context.Context,
	accessToken string,
) (*model.SessionPrincipal, error) {
	accessHash, err := hashPresentedToken(accessToken)
	if err != nil {
		return nil, ErrInvalidAccessToken
	}

	principal, err := s.store.FindPrincipalByAccessTokenHash(ctx, accessHash, s.now())
	if errors.Is(err, repository.ErrSessionNotFound) ||
		errors.Is(err, repository.ErrSessionExpired) ||
		errors.Is(err, repository.ErrSessionRevoked) {
		return nil, ErrInvalidAccessToken
	}
	if err != nil {
		return nil, err
	}

	return principal, nil
}

// Refresh rotates both tokens. The repository performs the consume-and-create
// operation atomically and revokes the family if an old refresh token is used.
func (s *SessionService) Refresh(
	ctx context.Context,
	refreshToken string,
) (*model.SessionTokenPair, error) {
	refreshHash, err := hashPresentedToken(refreshToken)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	accessToken, accessHash, err := s.generateToken()
	if err != nil {
		return nil, err
	}
	newRefreshToken, newRefreshHash, err := s.generateToken()
	if err != nil {
		return nil, err
	}

	now := s.now()
	replacement := model.SessionUserMobile{
		AccessTokenHash:  accessHash,
		RefreshTokenHash: newRefreshHash,
		AccessExpiresAt:  now.Add(s.accessTTL),
		RefreshExpiresAt: now.Add(s.refreshTTL),
	}

	rotated, err := s.store.RotateRefreshToken(ctx, refreshHash, replacement, now)
	switch {
	case errors.Is(err, repository.ErrRefreshTokenReused):
		return nil, ErrRefreshTokenReused
	case errors.Is(err, repository.ErrRefreshTooSoon):
		return nil, ErrRefreshTooSoon
	case errors.Is(err, repository.ErrSessionExpired):
		return nil, ErrRefreshTokenExpired
	case errors.Is(err, repository.ErrSessionNotFound), errors.Is(err, repository.ErrSessionRevoked):
		return nil, ErrInvalidRefreshToken
	case err != nil:
		return nil, err
	}
	if rotated == nil {
		return nil, fmt.Errorf("%w: store returned an empty rotated session", ErrInvalidSessionConfig)
	}

	return s.tokenPair(accessToken, newRefreshToken, rotated.AccessExpiresAt, rotated.RefreshExpiresAt, now), nil
}

// Logout revokes the complete family represented by an access token. The
// operation is intentionally idempotent for missing or already-revoked rows.
func (s *SessionService) Logout(ctx context.Context, accessToken string) error {
	accessHash, err := hashPresentedToken(accessToken)
	if err != nil {
		return ErrInvalidAccessToken
	}
	return s.store.RevokeFamilyByAccessTokenHash(ctx, accessHash, "logout", s.now())
}

func (s *SessionService) RevokeFamily(
	ctx context.Context,
	familyID string,
	reason string,
) error {
	if strings.TrimSpace(familyID) == "" {
		return ErrInvalidAccessToken
	}
	return s.store.RevokeFamily(ctx, familyID, reason, s.now())
}

func (s *SessionService) RevokeUser(
	ctx context.Context,
	userID int64,
	reason string,
) error {
	if userID <= 0 {
		return ErrInvalidSessionUser
	}
	return s.store.RevokeUser(ctx, userID, reason, s.now())
}

func (s *SessionService) tokenPair(
	accessToken string,
	refreshToken string,
	accessExpiresAt time.Time,
	refreshExpiresAt time.Time,
	now time.Time,
) *model.SessionTokenPair {
	return &model.SessionTokenPair{
		TokenType:        "Bearer",
		AccessToken:      accessToken,
		AccessExpiresAt:  accessExpiresAt,
		AccessExpiresIn:  secondsUntil(now, accessExpiresAt),
		RefreshToken:     refreshToken,
		RefreshExpiresAt: refreshExpiresAt,
		RefreshExpiresIn: secondsUntil(now, refreshExpiresAt),
	}
}

func secondsUntil(now time.Time, expiry time.Time) int64 {
	remaining := expiry.Sub(now)
	if remaining <= 0 {
		return 0
	}

	// Bulatkan ke atas supaya waktu proses sub-detik tidak mengurangi TTL
	// kontrak satu detik pada respons yang baru diterbitkan.
	seconds := int64(remaining / time.Second)
	if remaining%time.Second != 0 {
		seconds++
	}
	return seconds
}

func generateOpaqueToken() (string, []byte, error) {
	randomBytes := make([]byte, opaqueTokenBytes)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", nil, fmt.Errorf("generate opaque token: %w", err)
	}

	rawToken := base64.RawURLEncoding.EncodeToString(randomBytes)
	hash := sha256.Sum256([]byte(rawToken))
	return rawToken, hash[:], nil
}

func hashPresentedToken(rawToken string) ([]byte, error) {
	if rawToken == "" || len(rawToken) > maximumOpaqueTokenSize || strings.TrimSpace(rawToken) != rawToken {
		return nil, errors.New("invalid opaque token")
	}

	decoded, err := base64.RawURLEncoding.DecodeString(rawToken)
	if err != nil || len(decoded) != opaqueTokenBytes {
		return nil, errors.New("invalid opaque token")
	}

	hash := sha256.Sum256([]byte(rawToken))
	return hash[:], nil
}

func isNilSessionStore(store SessionStore) bool {
	if store == nil {
		return true
	}

	value := reflect.ValueOf(store)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
