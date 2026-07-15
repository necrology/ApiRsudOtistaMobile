package repository

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"apirusdotistamobile/internal/model"
)

var ErrInvalidAuthTicket = errors.New("invalid or expired auth ticket")

func (r *AuthRepository) CreateAuthTicket(
	ctx context.Context,
	verificationID int64,
	email string,
	purpose string,
	ticketHash [sha256.Size]byte,
	expiresAt time.Time,
) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err = tx.ExecContext(ctx, `
		UPDATE auth_ticket_mobile
		SET revoked_at = NOW()
		WHERE email = ?
		AND purpose = ?
		AND used_at IS NULL
		AND revoked_at IS NULL
	`, email, purpose); err != nil {
		return err
	}

	if _, err = tx.ExecContext(ctx, `
		INSERT INTO auth_ticket_mobile(
			verification_id,
			email,
			purpose,
			ticket_hash,
			expires_at
		) VALUES(?, ?, ?, ?, ?)
	`, verificationID, email, purpose, ticketHash[:], expiresAt); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *AuthRepository) CompleteRegistration(
	ctx context.Context,
	purpose string,
	ticketHash [sha256.Size]byte,
	passwordHash string,
	now time.Time,
) (*model.User, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var ticketID int64
	var user model.User
	err = tx.QueryRowContext(ctx, `
		SELECT
			t.id,
			v.username,
			v.email,
			COALESCE(v.phone, ''),
			COALESCE(v.full_name, '')
		FROM auth_ticket_mobile t
		INNER JOIN otp_verif_email_mobile v ON v.id = t.verification_id
		WHERE t.ticket_hash = ?
		AND t.purpose = ?
		AND v.email = t.email
		AND t.used_at IS NULL
		AND t.revoked_at IS NULL
		AND t.expires_at > ?
		AND v.is_used = TRUE
		LIMIT 1
		FOR UPDATE
	`, ticketHash[:], purpose, now).Scan(
		&ticketID,
		&user.Username,
		&user.Email,
		&user.Phone,
		&user.FullName,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidAuthTicket
		}
		return nil, err
	}

	result, err := tx.ExecContext(ctx, `
		INSERT INTO user_mobile(
			username,
			email,
			phone,
			full_name,
			password,
			email_verified,
			verified_at
		) VALUES(?, ?, ?, ?, ?, TRUE, NOW())
	`, user.Username, user.Email, nullableString(user.Phone), nullableString(user.FullName), passwordHash)
	if err != nil {
		return nil, fmt.Errorf("create registered user: %w", err)
	}

	user.ID, err = result.LastInsertId()
	if err != nil {
		return nil, err
	}
	user.Password = passwordHash
	user.EmailVerified = true

	if _, err = tx.ExecContext(ctx, `
		UPDATE auth_ticket_mobile
		SET used_at = NOW()
		WHERE id = ? AND used_at IS NULL
	`, ticketID); err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &user, nil
}
