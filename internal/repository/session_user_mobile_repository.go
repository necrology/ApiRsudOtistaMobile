package repository

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"apirusdotistamobile/internal/model"
)

var (
	ErrSessionNotFound      = errors.New("mobile session not found")
	ErrSessionExpired       = errors.New("mobile session expired")
	ErrSessionRevoked       = errors.New("mobile session revoked")
	ErrRefreshTokenReused   = errors.New("mobile refresh token reused")
	ErrRefreshTooSoon       = errors.New("mobile refresh token rotated too soon")
	ErrInvalidSessionRecord = errors.New("invalid mobile session record")
)

const MinimumRefreshRotationInterval = 2 * time.Second

// SessionUserMobileRepository persists opaque session token hashes in
// session_user_mobile. It deliberately has no method that accepts or returns a
// raw token.
type SessionUserMobileRepository struct {
	DB *sql.DB
}

func NewSessionUserMobileRepository(db *sql.DB) *SessionUserMobileRepository {
	return &SessionUserMobileRepository{DB: db}
}

// CreateSession starts a new independent session family, normally after a
// successful password and OTP verification.
func (r *SessionUserMobileRepository) CreateSession(
	ctx context.Context,
	session model.SessionUserMobile,
) (*model.SessionUserMobile, error) {
	if err := validateSessionRecord(session, true); err != nil {
		return nil, err
	}

	result, err := r.DB.ExecContext(ctx, `
		INSERT INTO session_user_mobile (
			family_id,
			parent_session_id,
			user_id,
			access_token_hash,
			refresh_token_hash,
			access_expires_at,
			refresh_expires_at,
			created_at,
			updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		session.FamilyID,
		nullablePositiveInt64(session.ParentSessionID),
		session.UserID,
		session.AccessTokenHash,
		session.RefreshTokenHash,
		session.AccessExpiresAt,
		session.RefreshExpiresAt,
		session.CreatedAt,
		session.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	session.ID, err = result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &session, nil
}

// FindPrincipalByAccessTokenHash resolves the authenticated identity while
// enforcing access-token expiry, revocation, and account deletion in one query.
func (r *SessionUserMobileRepository) FindPrincipalByAccessTokenHash(
	ctx context.Context,
	accessTokenHash []byte,
	now time.Time,
) (*model.SessionPrincipal, error) {
	if len(accessTokenHash) != sha256.Size {
		return nil, ErrSessionNotFound
	}

	var principal model.SessionPrincipal
	err := r.DB.QueryRowContext(ctx, `
		SELECT
			s.id,
			s.family_id,
			s.user_id,
			COALESCE(u.patient_id, 0),
			COALESCE(u.email, ''),
			COALESCE(u.no_rm, ''),
			s.access_expires_at
		FROM session_user_mobile s
		INNER JOIN user_mobile u ON u.id = s.user_id
		WHERE s.access_token_hash = ?
		AND s.revoked_at IS NULL
		AND s.rotated_at IS NULL
		AND s.access_expires_at > ?
		AND COALESCE(u.is_deleted, FALSE) = FALSE
		LIMIT 1
	`, accessTokenHash, now).Scan(
		&principal.SessionID,
		&principal.FamilyID,
		&principal.UserID,
		&principal.PatientID,
		&principal.Email,
		&principal.NoRM,
		&principal.AccessExpiresAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, err
	}

	return &principal, nil
}

// RotateRefreshToken atomically consumes a refresh token and creates its
// successor. Reusing an already-rotated token revokes every generation in the
// family, including the newest one.
func (r *SessionUserMobileRepository) RotateRefreshToken(
	ctx context.Context,
	refreshTokenHash []byte,
	replacement model.SessionUserMobile,
	now time.Time,
) (*model.SessionUserMobile, error) {
	if len(refreshTokenHash) != sha256.Size {
		return nil, ErrSessionNotFound
	}

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	type currentSession struct {
		ID               int64
		FamilyID         string
		UserID           int64
		RefreshExpiresAt time.Time
		RotatedAt        sql.NullTime
		RevokedAt        sql.NullTime
		CreatedAt        time.Time
	}

	var current currentSession
	err = tx.QueryRowContext(ctx, `
		SELECT
			id,
			family_id,
			user_id,
			refresh_expires_at,
			rotated_at,
			revoked_at,
			created_at
		FROM session_user_mobile
		WHERE refresh_token_hash = ?
		LIMIT 1
		FOR UPDATE
	`, refreshTokenHash).Scan(
		&current.ID,
		&current.FamilyID,
		&current.UserID,
		&current.RefreshExpiresAt,
		&current.RotatedAt,
		&current.RevokedAt,
		&current.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, err
	}

	if current.RotatedAt.Valid {
		if err = revokeFamilyTx(ctx, tx, current.FamilyID, now, "refresh_token_reuse"); err != nil {
			return nil, err
		}
		if err = tx.Commit(); err != nil {
			return nil, err
		}
		return nil, ErrRefreshTokenReused
	}

	if current.RevokedAt.Valid {
		return nil, ErrSessionRevoked
	}
	if now.Before(current.CreatedAt.Add(MinimumRefreshRotationInterval)) {
		return nil, ErrRefreshTooSoon
	}

	if !current.RefreshExpiresAt.After(now) {
		if err = revokeFamilyTx(ctx, tx, current.FamilyID, now, "refresh_token_expired"); err != nil {
			return nil, err
		}
		if err = tx.Commit(); err != nil {
			return nil, err
		}
		return nil, ErrSessionExpired
	}

	// Pertahankan batas absolut keluarga sesi dari generasi pertama. Rotasi
	// tidak boleh menggeser expiry 30 hari terus-menerus. Jika keluarga hampir
	// kedaluwarsa, access token baru ikut dibatasi sampai waktu yang sama.
	inheritAbsoluteFamilyExpiry(&replacement, current.RefreshExpiresAt)
	replacement.FamilyID = current.FamilyID
	replacement.ParentSessionID = current.ID
	replacement.UserID = current.UserID
	replacement.CreatedAt = now
	replacement.UpdatedAt = now
	if err = validateSessionRecord(replacement, true); err != nil {
		return nil, err
	}

	result, err := tx.ExecContext(ctx, `
		INSERT INTO session_user_mobile (
			family_id,
			parent_session_id,
			user_id,
			access_token_hash,
			refresh_token_hash,
			access_expires_at,
			refresh_expires_at,
			created_at,
			updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		replacement.FamilyID,
		replacement.ParentSessionID,
		replacement.UserID,
		replacement.AccessTokenHash,
		replacement.RefreshTokenHash,
		replacement.AccessExpiresAt,
		replacement.RefreshExpiresAt,
		replacement.CreatedAt,
		replacement.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	replacement.ID, err = result.LastInsertId()
	if err != nil {
		return nil, err
	}

	updateResult, err := tx.ExecContext(ctx, `
		UPDATE session_user_mobile
		SET
			rotated_at = ?,
			revoked_at = ?,
			revoke_reason = ?,
			replaced_by_session_id = ?,
			updated_at = ?
		WHERE id = ?
		AND rotated_at IS NULL
		AND revoked_at IS NULL
	`, now, now, "refresh_rotated", replacement.ID, now, current.ID)
	if err != nil {
		return nil, err
	}
	updatedRows, err := updateResult.RowsAffected()
	if err != nil {
		return nil, err
	}
	if updatedRows != 1 {
		return nil, ErrSessionRevoked
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &replacement, nil
}

// RevokeFamilyByAccessTokenHash implements logout without requiring a live
// access token. It intentionally returns success for an unknown token so the
// logout endpoint cannot be used to enumerate sessions.
func (r *SessionUserMobileRepository) RevokeFamilyByAccessTokenHash(
	ctx context.Context,
	accessTokenHash []byte,
	reason string,
	now time.Time,
) error {
	if len(accessTokenHash) != sha256.Size {
		return nil
	}

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var familyID string
	err = tx.QueryRowContext(ctx, `
		SELECT family_id
		FROM session_user_mobile
		WHERE access_token_hash = ?
		LIMIT 1
	`, accessTokenHash).Scan(&familyID)
	if errors.Is(err, sql.ErrNoRows) {
		return tx.Commit()
	}
	if err != nil {
		return err
	}

	if err = revokeFamilyTx(ctx, tx, familyID, now, reason); err != nil {
		return err
	}

	return tx.Commit()
}

// RevokeFamily revokes all token generations for one device/login family.
func (r *SessionUserMobileRepository) RevokeFamily(
	ctx context.Context,
	familyID string,
	reason string,
	now time.Time,
) error {
	if strings.TrimSpace(familyID) == "" {
		return ErrSessionNotFound
	}

	_, err := r.DB.ExecContext(ctx, `
		UPDATE session_user_mobile
		SET
			revoke_reason = CASE
				WHEN revoked_at IS NULL THEN ?
				ELSE revoke_reason
			END,
			revoked_at = COALESCE(revoked_at, ?),
			updated_at = ?
		WHERE family_id = ?
	`, normalizeRevokeReason(reason), now, now, familyID)
	return err
}

// RevokeUser revokes every session for an account, for example after a
// password reset, account disable, or suspected compromise.
func (r *SessionUserMobileRepository) RevokeUser(
	ctx context.Context,
	userID int64,
	reason string,
	now time.Time,
) error {
	if userID <= 0 {
		return ErrSessionNotFound
	}

	_, err := r.DB.ExecContext(ctx, `
		UPDATE session_user_mobile
		SET
			revoke_reason = CASE
				WHEN revoked_at IS NULL THEN ?
				ELSE revoke_reason
			END,
			revoked_at = COALESCE(revoked_at, ?),
			updated_at = ?
		WHERE user_id = ?
	`, normalizeRevokeReason(reason), now, now, userID)
	return err
}

func revokeFamilyTx(
	ctx context.Context,
	tx *sql.Tx,
	familyID string,
	now time.Time,
	reason string,
) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE session_user_mobile
		SET
			revoke_reason = CASE
				WHEN revoked_at IS NULL THEN ?
				ELSE revoke_reason
			END,
			revoked_at = COALESCE(revoked_at, ?),
			updated_at = ?
		WHERE family_id = ?
	`, normalizeRevokeReason(reason), now, now, familyID)
	return err
}

func validateSessionRecord(session model.SessionUserMobile, requireFamily bool) error {
	if session.UserID <= 0 {
		return fmt.Errorf("%w: user_id must be positive", ErrInvalidSessionRecord)
	}
	if requireFamily && strings.TrimSpace(session.FamilyID) == "" {
		return fmt.Errorf("%w: family_id is required", ErrInvalidSessionRecord)
	}
	if len(session.AccessTokenHash) != sha256.Size || len(session.RefreshTokenHash) != sha256.Size {
		return fmt.Errorf("%w: token hashes must be SHA-256", ErrInvalidSessionRecord)
	}
	if session.AccessExpiresAt.IsZero() || session.RefreshExpiresAt.IsZero() {
		return fmt.Errorf("%w: token expiry is required", ErrInvalidSessionRecord)
	}
	if session.RefreshExpiresAt.Before(session.AccessExpiresAt) {
		return fmt.Errorf("%w: refresh expiry must not precede access expiry", ErrInvalidSessionRecord)
	}
	return nil
}

func inheritAbsoluteFamilyExpiry(session *model.SessionUserMobile, familyExpiresAt time.Time) {
	session.RefreshExpiresAt = familyExpiresAt
	if session.AccessExpiresAt.After(familyExpiresAt) {
		session.AccessExpiresAt = familyExpiresAt
	}
}

func normalizeRevokeReason(reason string) string {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = "revoked"
	}

	const maxCharacters = 100
	if utf8.RuneCountInString(reason) <= maxCharacters {
		return reason
	}

	runes := []rune(reason)
	return string(runes[:maxCharacters])
}

func nullablePositiveInt64(value int64) any {
	if value <= 0 {
		return nil
	}
	return value
}
