package database

import (
	"database/sql"
	"fmt"
	"strings"
)

func EnsureAuthSchema(db *sql.DB, schema string) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS user_mobile (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			username VARCHAR(100) NOT NULL,
			email VARCHAR(100) NOT NULL,
			no_rm VARCHAR(20) NULL,
			patient_id BIGINT UNSIGNED NULL,
			phone VARCHAR(30) NULL,
			full_name VARCHAR(150) NULL,
			password VARCHAR(255) NOT NULL,
			email_verified BOOLEAN DEFAULT FALSE,
			verification_token VARCHAR(255) NULL,
			verified_at TIMESTAMP NULL DEFAULT NULL,
			medical_record_verified_at TIMESTAMP NULL DEFAULT NULL,
			is_deleted BOOLEAN DEFAULT FALSE,
			deleted_at TIMESTAMP NULL DEFAULT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_user_email (email),
			KEY idx_user_mobile_no_rm (no_rm),
			UNIQUE KEY uk_user_patient_id (patient_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS otp_user_mobile (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			user_id BIGINT NOT NULL,
			otp_code VARCHAR(10) NOT NULL,
			otp_hash VARCHAR(255) NULL,
			attempt_count SMALLINT UNSIGNED NOT NULL DEFAULT 0,
			last_attempt_at DATETIME(6) NULL DEFAULT NULL,
			locked_at DATETIME(6) NULL DEFAULT NULL,
			used_at DATETIME(6) NULL DEFAULT NULL,
			expired_at TIMESTAMP NOT NULL,
			is_used BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES user_mobile(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS otp_verif_email_mobile (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			username VARCHAR(100) NOT NULL,
			email VARCHAR(100) NOT NULL,
			no_rm VARCHAR(20) NULL,
			phone VARCHAR(30) NULL,
			full_name VARCHAR(150) NULL,
			otp_code VARCHAR(10) NOT NULL,
			otp_hash VARCHAR(255) NULL,
			attempt_count SMALLINT UNSIGNED NOT NULL DEFAULT 0,
			last_attempt_at DATETIME(6) NULL DEFAULT NULL,
			locked_at DATETIME(6) NULL DEFAULT NULL,
			used_at DATETIME(6) NULL DEFAULT NULL,
			expired_at TIMESTAMP NOT NULL,
			is_used BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS otp_password_reset_mobile (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			user_id BIGINT NOT NULL,
			otp_code VARCHAR(10) NOT NULL,
			otp_hash VARCHAR(255) NULL,
			attempt_count SMALLINT UNSIGNED NOT NULL DEFAULT 0,
			last_attempt_at DATETIME(6) NULL DEFAULT NULL,
			locked_at DATETIME(6) NULL DEFAULT NULL,
			used_at DATETIME(6) NULL DEFAULT NULL,
			expired_at TIMESTAMP NOT NULL,
			is_used BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES user_mobile(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS otp_account_deletion_mobile (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			user_id BIGINT NOT NULL,
			otp_code VARCHAR(10) NOT NULL,
			otp_hash VARCHAR(255) NULL,
			attempt_count SMALLINT UNSIGNED NOT NULL DEFAULT 0,
			last_attempt_at DATETIME(6) NULL DEFAULT NULL,
			locked_at DATETIME(6) NULL DEFAULT NULL,
			used_at DATETIME(6) NULL DEFAULT NULL,
			expired_at TIMESTAMP NOT NULL,
			is_used BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES user_mobile(id) ON DELETE CASCADE,
			KEY idx_otp_account_deletion_user_used_expired (user_id, is_used, expired_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS otp_medical_record_claim_mobile (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			user_id BIGINT NOT NULL,
			patient_id BIGINT UNSIGNED NOT NULL,
			no_rm VARCHAR(20) NOT NULL,
			patient_name VARCHAR(150) NULL,
			otp_code VARCHAR(10) NOT NULL,
			otp_hash VARCHAR(255) NULL,
			attempt_count SMALLINT UNSIGNED NOT NULL DEFAULT 0,
			last_attempt_at DATETIME(6) NULL DEFAULT NULL,
			locked_at DATETIME(6) NULL DEFAULT NULL,
			used_at DATETIME(6) NULL DEFAULT NULL,
			expired_at TIMESTAMP NOT NULL,
			is_used BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES user_mobile(id) ON DELETE CASCADE,
			KEY idx_otp_medical_record_claim_user (user_id),
			KEY idx_otp_medical_record_claim_no_rm (no_rm)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS session_user_mobile (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			family_id VARCHAR(64) NOT NULL,
			parent_session_id BIGINT NULL,
			user_id BIGINT NOT NULL,
			access_token_hash BINARY(32) NOT NULL,
			refresh_token_hash BINARY(32) NOT NULL,
			access_expires_at DATETIME(6) NOT NULL,
			refresh_expires_at DATETIME(6) NOT NULL,
			rotated_at DATETIME(6) NULL DEFAULT NULL,
			revoked_at DATETIME(6) NULL DEFAULT NULL,
			revoke_reason VARCHAR(100) NULL,
			replaced_by_session_id BIGINT NULL,
			created_at DATETIME(6) NOT NULL,
			updated_at DATETIME(6) NOT NULL,
			UNIQUE KEY uk_session_user_mobile_access_hash (access_token_hash),
			UNIQUE KEY uk_session_user_mobile_refresh_hash (refresh_token_hash),
			KEY idx_session_user_mobile_user_active (user_id, revoked_at, refresh_expires_at),
			KEY idx_session_user_mobile_family_active (family_id, revoked_at),
			KEY idx_session_user_mobile_parent (parent_session_id),
			KEY idx_session_user_mobile_replacement (replaced_by_session_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS auth_ticket_mobile (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			verification_id BIGINT NOT NULL,
			email VARCHAR(100) NOT NULL,
			purpose VARCHAR(40) NOT NULL,
			ticket_hash BINARY(32) NOT NULL,
			expires_at DATETIME(6) NOT NULL,
			used_at DATETIME(6) NULL DEFAULT NULL,
			revoked_at DATETIME(6) NULL DEFAULT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY uk_auth_ticket_mobile_hash (ticket_hash),
			KEY idx_auth_ticket_mobile_email_active (email, purpose, used_at, expires_at),
			KEY idx_auth_ticket_mobile_verification (verification_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
	}

	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			return err
		}
	}

	for _, table := range []string{
		"user_mobile",
		"otp_user_mobile",
		"otp_verif_email_mobile",
		"otp_password_reset_mobile",
		"otp_account_deletion_mobile",
		"otp_medical_record_claim_mobile",
		"session_user_mobile",
		"auth_ticket_mobile",
	} {
		if err := ensureInnoDB(db, schema, table); err != nil {
			return fmt.Errorf("ensure InnoDB engine for %s: %w", table, err)
		}
	}

	columns := []struct {
		table  string
		column string
		ddl    string
	}{
		{"user_mobile", "no_rm", "ALTER TABLE `user_mobile` ADD COLUMN `no_rm` VARCHAR(20) NULL AFTER `email`"},
		{"user_mobile", "patient_id", "ALTER TABLE `user_mobile` ADD COLUMN `patient_id` BIGINT UNSIGNED NULL AFTER `no_rm`"},
		{"user_mobile", "phone", "ALTER TABLE `user_mobile` ADD COLUMN `phone` VARCHAR(30) NULL AFTER `patient_id`"},
		{"user_mobile", "full_name", "ALTER TABLE `user_mobile` ADD COLUMN `full_name` VARCHAR(150) NULL AFTER `phone`"},
		{"user_mobile", "email_verified", "ALTER TABLE `user_mobile` ADD COLUMN `email_verified` BOOLEAN DEFAULT FALSE AFTER `password`"},
		{"user_mobile", "verification_token", "ALTER TABLE `user_mobile` ADD COLUMN `verification_token` VARCHAR(255) NULL AFTER `email_verified`"},
		{"user_mobile", "verified_at", "ALTER TABLE `user_mobile` ADD COLUMN `verified_at` TIMESTAMP NULL DEFAULT NULL AFTER `verification_token`"},
		{"user_mobile", "medical_record_verified_at", "ALTER TABLE `user_mobile` ADD COLUMN `medical_record_verified_at` TIMESTAMP NULL DEFAULT NULL AFTER `verified_at`"},
		{"user_mobile", "is_deleted", "ALTER TABLE `user_mobile` ADD COLUMN `is_deleted` BOOLEAN DEFAULT FALSE AFTER `medical_record_verified_at`"},
		{"user_mobile", "deleted_at", "ALTER TABLE `user_mobile` ADD COLUMN `deleted_at` TIMESTAMP NULL DEFAULT NULL AFTER `is_deleted`"},
		{"user_mobile", "created_at", "ALTER TABLE `user_mobile` ADD COLUMN `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP AFTER `deleted_at`"},
		{"user_mobile", "updated_at", "ALTER TABLE `user_mobile` ADD COLUMN `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP AFTER `created_at`"},
		{"otp_verif_email_mobile", "no_rm", "ALTER TABLE `otp_verif_email_mobile` ADD COLUMN `no_rm` VARCHAR(20) NULL AFTER `email`"},
		{"otp_verif_email_mobile", "phone", "ALTER TABLE `otp_verif_email_mobile` ADD COLUMN `phone` VARCHAR(30) NULL AFTER `no_rm`"},
		{"otp_verif_email_mobile", "full_name", "ALTER TABLE `otp_verif_email_mobile` ADD COLUMN `full_name` VARCHAR(150) NULL AFTER `phone`"},
		{"otp_user_mobile", "otp_hash", "ALTER TABLE `otp_user_mobile` ADD COLUMN `otp_hash` VARCHAR(255) NULL AFTER `otp_code`"},
		{"otp_user_mobile", "attempt_count", "ALTER TABLE `otp_user_mobile` ADD COLUMN `attempt_count` SMALLINT UNSIGNED NOT NULL DEFAULT 0 AFTER `otp_hash`"},
		{"otp_user_mobile", "last_attempt_at", "ALTER TABLE `otp_user_mobile` ADD COLUMN `last_attempt_at` DATETIME(6) NULL DEFAULT NULL AFTER `attempt_count`"},
		{"otp_user_mobile", "locked_at", "ALTER TABLE `otp_user_mobile` ADD COLUMN `locked_at` DATETIME(6) NULL DEFAULT NULL AFTER `last_attempt_at`"},
		{"otp_user_mobile", "used_at", "ALTER TABLE `otp_user_mobile` ADD COLUMN `used_at` DATETIME(6) NULL DEFAULT NULL AFTER `locked_at`"},
		{"otp_verif_email_mobile", "otp_hash", "ALTER TABLE `otp_verif_email_mobile` ADD COLUMN `otp_hash` VARCHAR(255) NULL AFTER `otp_code`"},
		{"otp_verif_email_mobile", "attempt_count", "ALTER TABLE `otp_verif_email_mobile` ADD COLUMN `attempt_count` SMALLINT UNSIGNED NOT NULL DEFAULT 0 AFTER `otp_hash`"},
		{"otp_verif_email_mobile", "last_attempt_at", "ALTER TABLE `otp_verif_email_mobile` ADD COLUMN `last_attempt_at` DATETIME(6) NULL DEFAULT NULL AFTER `attempt_count`"},
		{"otp_verif_email_mobile", "locked_at", "ALTER TABLE `otp_verif_email_mobile` ADD COLUMN `locked_at` DATETIME(6) NULL DEFAULT NULL AFTER `last_attempt_at`"},
		{"otp_verif_email_mobile", "used_at", "ALTER TABLE `otp_verif_email_mobile` ADD COLUMN `used_at` DATETIME(6) NULL DEFAULT NULL AFTER `locked_at`"},
		{"otp_password_reset_mobile", "otp_hash", "ALTER TABLE `otp_password_reset_mobile` ADD COLUMN `otp_hash` VARCHAR(255) NULL AFTER `otp_code`"},
		{"otp_password_reset_mobile", "attempt_count", "ALTER TABLE `otp_password_reset_mobile` ADD COLUMN `attempt_count` SMALLINT UNSIGNED NOT NULL DEFAULT 0 AFTER `otp_hash`"},
		{"otp_password_reset_mobile", "last_attempt_at", "ALTER TABLE `otp_password_reset_mobile` ADD COLUMN `last_attempt_at` DATETIME(6) NULL DEFAULT NULL AFTER `attempt_count`"},
		{"otp_password_reset_mobile", "locked_at", "ALTER TABLE `otp_password_reset_mobile` ADD COLUMN `locked_at` DATETIME(6) NULL DEFAULT NULL AFTER `last_attempt_at`"},
		{"otp_password_reset_mobile", "used_at", "ALTER TABLE `otp_password_reset_mobile` ADD COLUMN `used_at` DATETIME(6) NULL DEFAULT NULL AFTER `locked_at`"},
		{"otp_account_deletion_mobile", "otp_hash", "ALTER TABLE `otp_account_deletion_mobile` ADD COLUMN `otp_hash` VARCHAR(255) NULL AFTER `otp_code`"},
		{"otp_account_deletion_mobile", "attempt_count", "ALTER TABLE `otp_account_deletion_mobile` ADD COLUMN `attempt_count` SMALLINT UNSIGNED NOT NULL DEFAULT 0 AFTER `otp_hash`"},
		{"otp_account_deletion_mobile", "last_attempt_at", "ALTER TABLE `otp_account_deletion_mobile` ADD COLUMN `last_attempt_at` DATETIME(6) NULL DEFAULT NULL AFTER `attempt_count`"},
		{"otp_account_deletion_mobile", "locked_at", "ALTER TABLE `otp_account_deletion_mobile` ADD COLUMN `locked_at` DATETIME(6) NULL DEFAULT NULL AFTER `last_attempt_at`"},
		{"otp_account_deletion_mobile", "used_at", "ALTER TABLE `otp_account_deletion_mobile` ADD COLUMN `used_at` DATETIME(6) NULL DEFAULT NULL AFTER `locked_at`"},
		{"otp_medical_record_claim_mobile", "patient_name", "ALTER TABLE `otp_medical_record_claim_mobile` ADD COLUMN `patient_name` VARCHAR(150) NULL AFTER `no_rm`"},
		{"otp_medical_record_claim_mobile", "otp_hash", "ALTER TABLE `otp_medical_record_claim_mobile` ADD COLUMN `otp_hash` VARCHAR(255) NULL AFTER `otp_code`"},
		{"otp_medical_record_claim_mobile", "attempt_count", "ALTER TABLE `otp_medical_record_claim_mobile` ADD COLUMN `attempt_count` SMALLINT UNSIGNED NOT NULL DEFAULT 0 AFTER `otp_hash`"},
		{"otp_medical_record_claim_mobile", "last_attempt_at", "ALTER TABLE `otp_medical_record_claim_mobile` ADD COLUMN `last_attempt_at` DATETIME(6) NULL DEFAULT NULL AFTER `attempt_count`"},
		{"otp_medical_record_claim_mobile", "locked_at", "ALTER TABLE `otp_medical_record_claim_mobile` ADD COLUMN `locked_at` DATETIME(6) NULL DEFAULT NULL AFTER `last_attempt_at`"},
		{"otp_medical_record_claim_mobile", "used_at", "ALTER TABLE `otp_medical_record_claim_mobile` ADD COLUMN `used_at` DATETIME(6) NULL DEFAULT NULL AFTER `locked_at`"},
	}

	for _, column := range columns {
		if err := ensureColumn(db, schema, column.table, column.column, column.ddl); err != nil {
			return fmt.Errorf("ensure column %s.%s: %w", column.table, column.column, err)
		}
	}

	if err := dropIndexIfExists(db, schema, "user_mobile", "uk_user_no_rm"); err != nil {
		return fmt.Errorf("drop deprecated index user_mobile.uk_user_no_rm: %w", err)
	}

	indexes := []struct {
		table string
		index string
		ddl   string
	}{
		{"user_mobile", "uk_user_email", "ALTER TABLE `user_mobile` ADD UNIQUE KEY `uk_user_email` (`email`)"},
		{"user_mobile", "idx_user_mobile_no_rm", "ALTER TABLE `user_mobile` ADD KEY `idx_user_mobile_no_rm` (`no_rm`)"},
		{"user_mobile", "uk_user_patient_id", "ALTER TABLE `user_mobile` ADD UNIQUE KEY `uk_user_patient_id` (`patient_id`)"},
		{"otp_user_mobile", "idx_otp_user_mobile_user_used_expired", "ALTER TABLE `otp_user_mobile` ADD KEY `idx_otp_user_mobile_user_used_expired` (`user_id`, `is_used`, `expired_at`)"},
		{"otp_verif_email_mobile", "idx_otp_verif_email_mobile_email_used_expired", "ALTER TABLE `otp_verif_email_mobile` ADD KEY `idx_otp_verif_email_mobile_email_used_expired` (`email`, `is_used`, `expired_at`)"},
		{"otp_password_reset_mobile", "idx_otp_password_reset_mobile_user_used_expired", "ALTER TABLE `otp_password_reset_mobile` ADD KEY `idx_otp_password_reset_mobile_user_used_expired` (`user_id`, `is_used`, `expired_at`)"},
		{"otp_account_deletion_mobile", "idx_otp_account_deletion_user_used_expired", "ALTER TABLE `otp_account_deletion_mobile` ADD KEY `idx_otp_account_deletion_user_used_expired` (`user_id`, `is_used`, `expired_at`)"},
		{"otp_medical_record_claim_mobile", "idx_otp_medical_record_claim_mobile_user_used_expired", "ALTER TABLE `otp_medical_record_claim_mobile` ADD KEY `idx_otp_medical_record_claim_mobile_user_used_expired` (`user_id`, `is_used`, `expired_at`)"},
		{"session_user_mobile", "uk_session_user_mobile_access_hash", "ALTER TABLE `session_user_mobile` ADD UNIQUE KEY `uk_session_user_mobile_access_hash` (`access_token_hash`)"},
		{"session_user_mobile", "uk_session_user_mobile_refresh_hash", "ALTER TABLE `session_user_mobile` ADD UNIQUE KEY `uk_session_user_mobile_refresh_hash` (`refresh_token_hash`)"},
		{"session_user_mobile", "idx_session_user_mobile_user_active", "ALTER TABLE `session_user_mobile` ADD KEY `idx_session_user_mobile_user_active` (`user_id`, `revoked_at`, `refresh_expires_at`)"},
		{"session_user_mobile", "idx_session_user_mobile_family_active", "ALTER TABLE `session_user_mobile` ADD KEY `idx_session_user_mobile_family_active` (`family_id`, `revoked_at`)"},
		{"session_user_mobile", "idx_session_user_mobile_parent", "ALTER TABLE `session_user_mobile` ADD KEY `idx_session_user_mobile_parent` (`parent_session_id`)"},
		{"session_user_mobile", "idx_session_user_mobile_replacement", "ALTER TABLE `session_user_mobile` ADD KEY `idx_session_user_mobile_replacement` (`replaced_by_session_id`)"},
		{"auth_ticket_mobile", "uk_auth_ticket_mobile_hash", "ALTER TABLE `auth_ticket_mobile` ADD UNIQUE KEY `uk_auth_ticket_mobile_hash` (`ticket_hash`)"},
		{"auth_ticket_mobile", "idx_auth_ticket_mobile_email_active", "ALTER TABLE `auth_ticket_mobile` ADD KEY `idx_auth_ticket_mobile_email_active` (`email`, `purpose`, `used_at`, `expires_at`)"},
		{"auth_ticket_mobile", "idx_auth_ticket_mobile_verification", "ALTER TABLE `auth_ticket_mobile` ADD KEY `idx_auth_ticket_mobile_verification` (`verification_id`)"},
	}

	for _, index := range indexes {
		if err := ensureIndex(db, schema, index.table, index.index, index.ddl); err != nil {
			return fmt.Errorf("ensure index %s.%s: %w", index.table, index.index, err)
		}
	}

	return nil
}

func ensureColumn(db *sql.DB, schema string, table string, column string, ddl string) error {
	var count int
	if err := db.QueryRow(`
		SELECT COUNT(*)
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = ?
		AND TABLE_NAME = ?
		AND COLUMN_NAME = ?
	`, schema, table, column).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	_, err := db.Exec(ddl)
	return err
}

func ensureInnoDB(db *sql.DB, schema string, table string) error {
	var engine string
	if err := db.QueryRow(`
		SELECT COALESCE(ENGINE, '')
		FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = ?
		AND TABLE_NAME = ?
		LIMIT 1
	`, schema, table).Scan(&engine); err != nil {
		return err
	}
	if strings.EqualFold(engine, "InnoDB") {
		return nil
	}

	_, err := db.Exec(fmt.Sprintf("ALTER TABLE `%s` ENGINE=InnoDB", table))
	return err
}

func ensureIndex(db *sql.DB, schema string, table string, index string, ddl string) error {
	var count int
	if err := db.QueryRow(`
		SELECT COUNT(*)
		FROM information_schema.STATISTICS
		WHERE TABLE_SCHEMA = ?
		AND TABLE_NAME = ?
		AND INDEX_NAME = ?
	`, schema, table, index).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	_, err := db.Exec(ddl)
	return err
}

func dropIndexIfExists(db *sql.DB, schema string, table string, index string) error {
	var count int
	if err := db.QueryRow(`
		SELECT COUNT(*)
		FROM information_schema.STATISTICS
		WHERE TABLE_SCHEMA = ?
		AND TABLE_NAME = ?
		AND INDEX_NAME = ?
	`, schema, table, index).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		return nil
	}

	_, err := db.Exec(fmt.Sprintf("ALTER TABLE `%s` DROP INDEX `%s`", table, index))
	return err
}
