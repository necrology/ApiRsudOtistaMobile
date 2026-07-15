package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"apirusdotistamobile/internal/model"
)

type AuthRepository struct {
	DB *sql.DB
}

var (
	ErrNoRMNotLinked = errors.New("no rm not linked")
	ErrNoRMAmbiguous = errors.New("no rm ambiguous")
)

func NewAuthRepository(db *sql.DB) *AuthRepository {
	return &AuthRepository{
		DB: db,
	}
}

func (r *AuthRepository) CreateVerifEmail(
	user model.User,
	otpHash string,
) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err = tx.Exec(`
		UPDATE otp_verif_email_mobile
		SET is_used = TRUE, used_at = COALESCE(used_at, NOW())
		WHERE email = ? AND is_used = FALSE
	`, user.Email); err != nil {
		return err
	}

	query := `
	INSERT INTO otp_verif_email_mobile(
		username,
		email,
		phone,
		full_name,
		otp_code,
		otp_hash,
		expired_at
	)
	VALUES(
		?,
		?,
		?,
		?,
		'',
		?,
		DATE_ADD(NOW(), INTERVAL 5 MINUTE)
	)
	`

	_, err = tx.Exec(
		query,
		user.Username,
		user.Email,
		nullableString(user.Phone),
		nullableString(user.FullName),
		otpHash,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *AuthRepository) VerifyOTPNewUser(
	email string,
	otp string,
) (int64, bool, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return 0, false, err
	}
	defer tx.Rollback()

	query := `
	SELECT id, COALESCE(otp_hash, ''), otp_code, attempt_count
	FROM otp_verif_email_mobile
	WHERE email = ?
	AND is_used = FALSE
	AND expired_at > NOW()
	AND locked_at IS NULL
	ORDER BY id DESC
	LIMIT 1
	FOR UPDATE
	`

	var otpID int64
	var otpHash string
	var legacyOTP string
	var attemptCount int
	err = tx.QueryRow(query, email).Scan(&otpID, &otpHash, &legacyOTP, &attemptCount)
	if err != nil {
		return 0, false, err
	}

	storedOTP := otpHash
	if storedOTP == "" {
		storedOTP = legacyOTP
	}
	if !verifyStoredOTP(storedOTP, strings.TrimSpace(otp)) {
		_, err = tx.Exec(`
			UPDATE otp_verif_email_mobile
			SET attempt_count = attempt_count + 1,
				last_attempt_at = NOW(),
				locked_at = CASE
					WHEN attempt_count + 1 >= ? THEN NOW()
					ELSE locked_at
				END
			WHERE id = ?
		`, maxOTPVerificationAttempts, otpID)
		if err != nil {
			return 0, false, err
		}
		if err = tx.Commit(); err != nil {
			return 0, false, err
		}
		return 0, false, nil
	}

	updateQuery := `
	UPDATE otp_verif_email_mobile
	SET is_used = TRUE, used_at = NOW(), last_attempt_at = NOW()
	WHERE id = ?
	`

	_, err = tx.Exec(updateQuery, otpID)
	if err != nil {
		return 0, false, err
	}

	if err = tx.Commit(); err != nil {
		return 0, false, err
	}

	return otpID, true, nil
}

func (r *AuthRepository) Create(
	user model.User,
) error {

	query := `
	INSERT INTO user_mobile(
		username,
		email,
		phone,
		full_name,
		password,
		email_verified
	)
	VALUES(
		?,
		?,
		?,
		?,
		?,
		true
	)
	`

	_, err := r.DB.Exec(
		query,
		user.Username,
		user.Email,
		nullableString(user.Phone),
		nullableString(user.FullName),
		user.Password,
	)

	return err
}

func (r *AuthRepository) FindByIdentifier(identifier string) (*model.User, error) {
	query := `
	SELECT id, username, email, COALESCE(no_rm, ''), COALESCE(patient_id, 0), COALESCE(phone, ''), COALESCE(full_name, ''), password, COALESCE(email_verified, false)
	FROM user_mobile
	WHERE email = ?
	AND COALESCE(is_deleted, false) = false
	LIMIT 1
	`

	var user model.User

	err := r.DB.QueryRow(query, identifier).
		Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.NoRM,
			&user.PatientID,
			&user.Phone,
			&user.FullName,
			&user.Password,
			&user.EmailVerified,
		)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *AuthRepository) FindLinkedUserByNoRM(noRM string) (*model.User, error) {
	query := `
	SELECT id, username, email, COALESCE(no_rm, ''), COALESCE(patient_id, 0), COALESCE(phone, ''), COALESCE(full_name, ''), password, COALESCE(email_verified, false)
	FROM user_mobile
	WHERE no_rm = ?
	AND COALESCE(patient_id, 0) <> 0
	AND COALESCE(is_deleted, false) = false
	ORDER BY id ASC
	LIMIT 2
	`

	rows, err := r.DB.Query(query, noRM)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]model.User, 0, 2)
	for rows.Next() {
		var user model.User
		if err = rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.NoRM,
			&user.PatientID,
			&user.Phone,
			&user.FullName,
			&user.Password,
			&user.EmailVerified,
		); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, ErrNoRMNotLinked
	}
	if len(users) > 1 {
		return nil, ErrNoRMAmbiguous
	}

	return &users[0], nil
}

func (r *AuthRepository) FindUserByEmail(email string) (*model.User, error) {
	query := `
	SELECT id, username, email, COALESCE(no_rm, ''), COALESCE(patient_id, 0), COALESCE(phone, ''), COALESCE(full_name, ''), password, COALESCE(email_verified, false)
	FROM user_mobile
	WHERE email = ?
	AND COALESCE(is_deleted, false) = false
	LIMIT 1
	`

	var user model.User

	err := r.DB.QueryRow(query, email).
		Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.NoRM,
			&user.PatientID,
			&user.Phone,
			&user.FullName,
			&user.Password,
			&user.EmailVerified,
		)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *AuthRepository) FindUserByID(userID int64) (*model.User, error) {
	query := `
	SELECT id, username, email, COALESCE(no_rm, ''), COALESCE(patient_id, 0), COALESCE(phone, ''), COALESCE(full_name, ''), password, COALESCE(email_verified, false)
	FROM user_mobile
	WHERE id = ?
	AND COALESCE(is_deleted, false) = false
	LIMIT 1
	`

	var user model.User
	if err := r.DB.QueryRow(query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.NoRM,
		&user.PatientID,
		&user.Phone,
		&user.FullName,
		&user.Password,
		&user.EmailVerified,
	); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *AuthRepository) FindPatientByMedicalRecordClaim(
	noRM string,
	nik string,
	birthDate string,
) (*model.PatientMedicalRecord, error) {
	birthDateColumn, err := r.firstExistingColumn("pasiens", []string{"tanggal_lahir", "tgllahir"})
	if err != nil {
		return nil, err
	}

	phoneColumn, err := r.firstExistingColumn("pasiens", []string{"no_hp", "nohp"})
	phoneExpression := "''"
	if err == nil {
		phoneExpression = fmt.Sprintf("COALESCE(%s, '')", quoteIdentifier(phoneColumn))
	} else if err != sql.ErrNoRows {
		return nil, err
	}

	birthDateExpression := quoteIdentifier(birthDateColumn)
	birthDateValues := birthDateVariants(birthDate)

	query := fmt.Sprintf(`
	SELECT id, no_rm, COALESCE(nama, ''), %s
	FROM pasiens
	WHERE no_rm = ?
	AND nik = ?
	AND (
		DATE(%s) = ?
		OR TRIM(%s) IN (?, ?, ?, ?, ?)
	)
	ORDER BY id ASC
	LIMIT 2
	`, phoneExpression, birthDateExpression, birthDateExpression)

	rows, err := r.DB.Query(
		query,
		noRM,
		nik,
		birthDate,
		birthDateValues[0],
		birthDateValues[1],
		birthDateValues[2],
		birthDateValues[3],
		birthDateValues[4],
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	patients := make([]model.PatientMedicalRecord, 0, 2)
	for rows.Next() {
		var patient model.PatientMedicalRecord
		if err = rows.Scan(
			&patient.ID,
			&patient.NoRM,
			&patient.FullName,
			&patient.Phone,
		); err != nil {
			return nil, err
		}
		patients = append(patients, patient)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	if len(patients) != 1 {
		return nil, sql.ErrNoRows
	}

	return &patients[0], nil
}

func (r *AuthRepository) IsPatientLinkedToAnotherUser(
	patientID int64,
	userID int64,
) (bool, error) {
	query := `
	SELECT COUNT(*)
	FROM user_mobile
	WHERE patient_id = ?
	AND id <> ?
	AND COALESCE(is_deleted, false) = false
	`

	var count int
	if err := r.DB.QueryRow(query, patientID, userID).Scan(&count); err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *AuthRepository) CreateMedicalRecordClaimOTP(
	userID int64,
	patient model.PatientMedicalRecord,
	otpHash string,
) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err = tx.Exec(`
		UPDATE otp_medical_record_claim_mobile
		SET is_used = TRUE, used_at = COALESCE(used_at, NOW())
		WHERE user_id = ? AND is_used = FALSE
	`, userID); err != nil {
		return err
	}

	query := `
	INSERT INTO otp_medical_record_claim_mobile(
		user_id,
		patient_id,
		no_rm,
		patient_name,
		otp_code,
		otp_hash,
		expired_at
	)
	VALUES(
		?,
		?,
		?,
		?,
		'',
		?,
		DATE_ADD(NOW(), INTERVAL 5 MINUTE)
	)
	`

	_, err = tx.Exec(
		query,
		userID,
		patient.ID,
		patient.NoRM,
		patient.FullName,
		otpHash,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *AuthRepository) ConfirmMedicalRecordClaim(
	authenticatedUserID int64,
	otp string,
) (*model.User, bool, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return nil, false, err
	}
	defer tx.Rollback()

	query := `
	SELECT c.id, c.user_id, c.patient_id, c.no_rm,
		COALESCE(c.patient_name, ''), COALESCE(c.otp_hash, ''),
		c.otp_code, c.attempt_count
	FROM otp_medical_record_claim_mobile c
	INNER JOIN user_mobile u ON u.id = c.user_id
	WHERE u.id = ?
	AND c.is_used = FALSE
	AND c.expired_at > NOW()
	AND c.locked_at IS NULL
	AND COALESCE(u.is_deleted, false) = false
	ORDER BY c.id DESC
	LIMIT 1
	FOR UPDATE
	`

	var claimID int64
	var userID int64
	var patientID int64
	var noRM string
	var patientName string
	var otpHash string
	var legacyOTP string
	var attemptCount int
	if err = tx.QueryRow(query, authenticatedUserID).Scan(
		&claimID,
		&userID,
		&patientID,
		&noRM,
		&patientName,
		&otpHash,
		&legacyOTP,
		&attemptCount,
	); err != nil {
		return nil, false, err
	}

	storedOTP := otpHash
	if storedOTP == "" {
		storedOTP = legacyOTP
	}
	if !verifyStoredOTP(storedOTP, strings.TrimSpace(otp)) {
		_, err = tx.Exec(`
			UPDATE otp_medical_record_claim_mobile
			SET attempt_count = attempt_count + 1,
				last_attempt_at = NOW(),
				locked_at = CASE
					WHEN attempt_count + 1 >= ? THEN NOW()
					ELSE locked_at
				END
			WHERE id = ?
		`, maxOTPVerificationAttempts, claimID)
		if err != nil {
			return nil, false, err
		}
		if err = tx.Commit(); err != nil {
			return nil, false, err
		}
		return nil, false, nil
	}

	var linkedCount int
	if err = tx.QueryRow(`
		SELECT COUNT(*)
		FROM user_mobile
		WHERE patient_id = ?
		AND id <> ?
		AND COALESCE(is_deleted, false) = false
	`, patientID, userID).Scan(&linkedCount); err != nil {
		return nil, false, err
	}

	if linkedCount > 0 {
		return nil, false, nil
	}

	if _, err = tx.Exec(`
		UPDATE user_mobile
		SET
			no_rm = ?,
			patient_id = ?,
			full_name = COALESCE(NULLIF(full_name, ''), ?),
			medical_record_verified_at = NOW()
		WHERE id = ?
	`, noRM, patientID, patientName, userID); err != nil {
		return nil, false, err
	}

	if _, err = tx.Exec(`
		UPDATE otp_medical_record_claim_mobile
		SET is_used = TRUE, used_at = NOW(), last_attempt_at = NOW()
		WHERE id = ?
	`, claimID); err != nil {
		return nil, false, err
	}

	var user model.User
	if err = tx.QueryRow(`
		SELECT id, username, email, COALESCE(no_rm, ''), COALESCE(patient_id, 0), COALESCE(phone, ''), COALESCE(full_name, ''), password, COALESCE(email_verified, false)
		FROM user_mobile
		WHERE id = ?
	`, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.NoRM,
		&user.PatientID,
		&user.Phone,
		&user.FullName,
		&user.Password,
		&user.EmailVerified,
	); err != nil {
		return nil, false, err
	}

	if err = tx.Commit(); err != nil {
		return nil, false, err
	}

	return &user, true, nil
}

func (r *AuthRepository) CreateOTPLogin(
	userID int64,
	otpHash string,
) error {
	return createUserOTPChallenge(r.DB, "otp_user_mobile", userID, otpHash)
}

func (r *AuthRepository) CreatePasswordResetOTP(
	userID int64,
	otpHash string,
) error {
	return createUserOTPChallenge(r.DB, "otp_password_reset_mobile", userID, otpHash)
}

func (r *AuthRepository) ResetPassword(
	userID int64,
	otp string,
	password string,
) (bool, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	query := `
	SELECT id, COALESCE(otp_hash, ''), otp_code, attempt_count
	FROM otp_password_reset_mobile
	WHERE user_id = ?
	AND is_used = FALSE
	AND expired_at > NOW()
	AND locked_at IS NULL
	ORDER BY id DESC
	LIMIT 1
	FOR UPDATE
	`

	var otpID int64
	var otpHash string
	var legacyOTP string
	var attemptCount int
	if err = tx.QueryRow(query, userID).Scan(&otpID, &otpHash, &legacyOTP, &attemptCount); err != nil {
		return false, err
	}

	storedOTP := otpHash
	if storedOTP == "" {
		storedOTP = legacyOTP
	}
	if !verifyStoredOTP(storedOTP, strings.TrimSpace(otp)) {
		_, err = tx.Exec(`
			UPDATE otp_password_reset_mobile
			SET attempt_count = attempt_count + 1,
				last_attempt_at = NOW(),
				locked_at = CASE
					WHEN attempt_count + 1 >= ? THEN NOW()
					ELSE locked_at
				END
			WHERE id = ?
		`, maxOTPVerificationAttempts, otpID)
		if err != nil {
			return false, err
		}
		if err = tx.Commit(); err != nil {
			return false, err
		}
		return false, nil
	}

	if _, err = tx.Exec(`UPDATE user_mobile SET password = ? WHERE id = ?`, password, userID); err != nil {
		return false, err
	}

	if _, err = tx.Exec(`
		UPDATE otp_password_reset_mobile
		SET is_used = TRUE, used_at = NOW(), last_attempt_at = NOW()
		WHERE id = ?
	`, otpID); err != nil {
		return false, err
	}

	// Password dan pencabutan seluruh sesi harus berhasil dalam transaksi yang
	// sama agar refresh token lama tidak tetap aktif setelah password berubah.
	if _, err = tx.Exec(`
		UPDATE session_user_mobile
		SET revoke_reason = CASE
				WHEN revoked_at IS NULL THEN 'password_reset'
				ELSE revoke_reason
			END,
			revoked_at = COALESCE(revoked_at, NOW(6)),
			updated_at = NOW(6)
		WHERE user_id = ?
	`, userID); err != nil {
		return false, err
	}

	if err = tx.Commit(); err != nil {
		return false, err
	}

	return true, nil
}

func (r *AuthRepository) VerifyOTPLogin(
	userID int64,
	otp string,
) (bool, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	query := `
	SELECT id, COALESCE(otp_hash, ''), otp_code, attempt_count
	FROM otp_user_mobile
	WHERE user_id = ?
	AND is_used = FALSE
	AND expired_at > NOW()
	AND locked_at IS NULL
	ORDER BY id DESC
	LIMIT 1
	FOR UPDATE
	`

	var otpID int64
	var otpHash string
	var legacyOTP string
	var attemptCount int
	err = tx.QueryRow(query, userID).Scan(&otpID, &otpHash, &legacyOTP, &attemptCount)
	if err != nil {
		return false, err
	}

	storedOTP := otpHash
	if storedOTP == "" {
		storedOTP = legacyOTP
	}
	if !verifyStoredOTP(storedOTP, strings.TrimSpace(otp)) {
		_, err = tx.Exec(`
			UPDATE otp_user_mobile
			SET attempt_count = attempt_count + 1,
				last_attempt_at = NOW(),
				locked_at = CASE
					WHEN attempt_count + 1 >= ? THEN NOW()
					ELSE locked_at
				END
			WHERE id = ?
		`, maxOTPVerificationAttempts, otpID)
		if err != nil {
			return false, err
		}
		if err = tx.Commit(); err != nil {
			return false, err
		}
		return false, nil
	}

	updateQuery := `
	UPDATE otp_user_mobile
	SET is_used = TRUE, used_at = NOW(), last_attempt_at = NOW()
	WHERE id = ?
	`

	_, err = tx.Exec(updateQuery, otpID)
	if err != nil {
		return false, err
	}

	if err = tx.Commit(); err != nil {
		return false, err
	}

	return true, nil
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}

	return value
}

func (r *AuthRepository) firstExistingColumn(table string, columns []string) (string, error) {
	for _, column := range columns {
		var count int
		if err := r.DB.QueryRow(`
			SELECT COUNT(*)
			FROM information_schema.COLUMNS
			WHERE TABLE_SCHEMA = DATABASE()
			AND TABLE_NAME = ?
			AND COLUMN_NAME = ?
		`, table, column).Scan(&count); err != nil {
			return "", err
		}
		if count > 0 {
			return column, nil
		}
	}

	return "", sql.ErrNoRows
}

func birthDateVariants(value string) []string {
	parsed, err := time.Parse("2006-01-02", strings.TrimSpace(value))
	if err != nil {
		return []string{value, value, value, value, value}
	}

	return []string{
		parsed.Format("2006-01-02"),
		parsed.Format("02-01-2006"),
		parsed.Format("02/01/2006"),
		parsed.Format("2006/01/02"),
		parsed.Format("20060102"),
	}
}
