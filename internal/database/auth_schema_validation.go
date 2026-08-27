package database

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
)

var requiredAuthColumns = map[string][]string{
	"user_mobile": {
		"id", "username", "email", "no_rm", "patient_id", "phone", "full_name",
		"password", "email_verified", "verification_token", "verified_at",
		"medical_record_verified_at", "is_deleted", "deleted_at", "created_at", "updated_at",
	},
	"otp_user_mobile": {
		"id", "user_id", "otp_code", "otp_hash", "attempt_count", "last_attempt_at",
		"locked_at", "used_at", "expired_at", "is_used", "created_at",
	},
	"otp_verif_email_mobile": {
		"id", "username", "email", "no_rm", "phone", "full_name", "otp_code",
		"otp_hash", "attempt_count", "last_attempt_at", "locked_at", "used_at",
		"expired_at", "is_used", "created_at",
	},
	"otp_password_reset_mobile": {
		"id", "user_id", "otp_code", "otp_hash", "attempt_count", "last_attempt_at",
		"locked_at", "used_at", "expired_at", "is_used", "created_at",
	},
	"otp_account_deletion_mobile": {
		"id", "user_id", "otp_code", "otp_hash", "attempt_count", "last_attempt_at",
		"locked_at", "used_at", "expired_at", "is_used", "created_at",
	},
	"otp_medical_record_claim_mobile": {
		"id", "user_id", "patient_id", "no_rm", "patient_name", "otp_code",
		"otp_hash", "attempt_count", "last_attempt_at", "locked_at", "used_at",
		"expired_at", "is_used", "created_at",
	},
	"session_user_mobile": {
		"id", "family_id", "parent_session_id", "user_id", "access_token_hash",
		"refresh_token_hash", "access_expires_at", "refresh_expires_at", "rotated_at",
		"revoked_at", "revoke_reason", "replaced_by_session_id", "created_at", "updated_at",
	},
	"auth_ticket_mobile": {
		"id", "verification_id", "email", "purpose", "ticket_hash", "expires_at",
		"used_at", "revoked_at", "created_at",
	},
}

var requiredAuthUniqueIndexes = map[string]map[string][]string{
	"user_mobile": {
		"uk_user_email":      {"email"},
		"uk_user_patient_id": {"patient_id"},
	},
	"session_user_mobile": {
		"uk_session_user_mobile_access_hash":  {"access_token_hash"},
		"uk_session_user_mobile_refresh_hash": {"refresh_token_hash"},
	},
	"auth_ticket_mobile": {
		"uk_auth_ticket_mobile_hash": {"ticket_hash"},
	},
}

// ValidateAuthSchema hanya membaca information_schema. API memanggilnya saat
// startup agar health tidak 200 pada deployment dengan migrasi auth parsial.
func ValidateAuthSchema(db *sql.DB, schema string) error {
	if db == nil || strings.TrimSpace(schema) == "" {
		return fmt.Errorf("database and schema are required")
	}
	schema = strings.TrimSpace(schema)

	tables := make(map[string]string)
	rows, err := db.Query(`
		SELECT TABLE_NAME, COALESCE(ENGINE, '')
		FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = ?
		AND TABLE_NAME IN (
			'user_mobile', 'otp_user_mobile', 'otp_verif_email_mobile',
			'otp_password_reset_mobile', 'otp_account_deletion_mobile',
			'otp_medical_record_claim_mobile',
			'session_user_mobile', 'auth_ticket_mobile'
		)
	`, schema)
	if err != nil {
		return err
	}
	for rows.Next() {
		var table string
		var engine string
		if err = rows.Scan(&table, &engine); err != nil {
			rows.Close()
			return err
		}
		tables[table] = engine
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return err
	}
	if err = rows.Close(); err != nil {
		return err
	}

	columns := make(map[string]map[string]bool)
	rows, err = db.Query(`
		SELECT TABLE_NAME, COLUMN_NAME
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = ?
		AND TABLE_NAME IN (
			'user_mobile', 'otp_user_mobile', 'otp_verif_email_mobile',
			'otp_password_reset_mobile', 'otp_account_deletion_mobile',
			'otp_medical_record_claim_mobile',
			'session_user_mobile', 'auth_ticket_mobile'
		)
	`, schema)
	if err != nil {
		return err
	}
	for rows.Next() {
		var table string
		var column string
		if err = rows.Scan(&table, &column); err != nil {
			rows.Close()
			return err
		}
		if columns[table] == nil {
			columns[table] = make(map[string]bool)
		}
		columns[table][column] = true
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return err
	}
	if err = rows.Close(); err != nil {
		return err
	}

	uniqueIndexes := make(map[string]map[string][]string)
	rows, err = db.Query(`
		SELECT TABLE_NAME, INDEX_NAME, SEQ_IN_INDEX, COLUMN_NAME
		FROM information_schema.STATISTICS
		WHERE TABLE_SCHEMA = ?
		AND NON_UNIQUE = 0
		AND TABLE_NAME IN ('user_mobile', 'session_user_mobile', 'auth_ticket_mobile')
		ORDER BY TABLE_NAME, INDEX_NAME, SEQ_IN_INDEX
	`, schema)
	if err != nil {
		return err
	}
	for rows.Next() {
		var table string
		var index string
		var sequence int
		var column string
		if err = rows.Scan(&table, &index, &sequence, &column); err != nil {
			rows.Close()
			return err
		}
		if uniqueIndexes[table] == nil {
			uniqueIndexes[table] = make(map[string][]string)
		}
		columnsInIndex := uniqueIndexes[table][index]
		if sequence != len(columnsInIndex)+1 {
			columnsInIndex = append(columnsInIndex, "<invalid-sequence>")
		}
		uniqueIndexes[table][index] = append(columnsInIndex, column)
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return err
	}
	if err = rows.Close(); err != nil {
		return err
	}

	return validateAuthSchemaSnapshot(tables, columns, uniqueIndexes)
}

func validateAuthSchemaSnapshot(
	tables map[string]string,
	columns map[string]map[string]bool,
	uniqueIndexes map[string]map[string][]string,
) error {
	problems := make([]string, 0)
	for table, requiredColumns := range requiredAuthColumns {
		engine, exists := tables[table]
		if !exists {
			problems = append(problems, "missing table "+table)
			continue
		}
		if !strings.EqualFold(engine, "InnoDB") {
			problems = append(problems, table+" engine must be InnoDB")
		}
		for _, column := range requiredColumns {
			if !columns[table][column] {
				problems = append(problems, "missing column "+table+"."+column)
			}
		}
	}
	for table, indexes := range requiredAuthUniqueIndexes {
		for index, expectedColumns := range indexes {
			actualColumns, exists := uniqueIndexes[table][index]
			if !exists {
				problems = append(problems, "missing unique index "+table+"."+index)
				continue
			}
			if !equalIdentifierLists(actualColumns, expectedColumns) {
				problems = append(problems, fmt.Sprintf(
					"invalid unique index %s.%s columns: got (%s), want (%s)",
					table,
					index,
					strings.Join(actualColumns, ", "),
					strings.Join(expectedColumns, ", "),
				))
			}
		}
	}

	if len(problems) > 0 {
		sort.Strings(problems)
		return fmt.Errorf("auth schema validation failed: %s", strings.Join(problems, "; "))
	}
	return nil
}

func equalIdentifierLists(actual []string, expected []string) bool {
	if len(actual) != len(expected) {
		return false
	}
	for index := range expected {
		if !strings.EqualFold(actual[index], expected[index]) {
			return false
		}
	}
	return true
}
