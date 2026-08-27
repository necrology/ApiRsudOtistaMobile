package repository

import (
	"crypto/subtle"
	"database/sql"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const maxOTPVerificationAttempts = 5

func createUserOTPChallenge(db *sql.DB, table string, userID int64, otpHash string) error {
	switch table {
	case "otp_user_mobile", "otp_password_reset_mobile", "otp_account_deletion_mobile":
	default:
		return fmt.Errorf("unsupported otp table %q", table)
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err = tx.Exec(fmt.Sprintf(`
		UPDATE %s
		SET is_used = TRUE, used_at = COALESCE(used_at, NOW())
		WHERE user_id = ? AND is_used = FALSE
	`, table), userID); err != nil {
		return err
	}

	if _, err = tx.Exec(fmt.Sprintf(`
		INSERT INTO %s(user_id, otp_code, otp_hash, expired_at)
		VALUES(?, '', ?, DATE_ADD(NOW(), INTERVAL 5 MINUTE))
	`, table), userID, otpHash); err != nil {
		return err
	}

	return tx.Commit()
}

// verifyStoredOTP supports the bcrypt hashes written by the hardened API and
// a short-lived plaintext fallback for OTP rows created by the previous
// version. Legacy rows expire after five minutes and are never created by the
// new code.
func verifyStoredOTP(stored string, provided string) bool {
	if stored == "" || provided == "" {
		return false
	}

	if strings.HasPrefix(stored, "$2a$") ||
		strings.HasPrefix(stored, "$2b$") ||
		strings.HasPrefix(stored, "$2y$") {
		return bcrypt.CompareHashAndPassword([]byte(stored), []byte(provided)) == nil
	}

	if len(stored) != len(provided) {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(stored), []byte(provided)) == 1
}
