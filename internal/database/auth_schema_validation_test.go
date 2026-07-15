package database

import (
	"strings"
	"testing"
)

func completeAuthSchemaSnapshot() (
	map[string]string,
	map[string]map[string]bool,
	map[string]map[string][]string,
) {
	tables := make(map[string]string, len(requiredAuthColumns))
	columns := make(map[string]map[string]bool, len(requiredAuthColumns))
	for table, requiredColumns := range requiredAuthColumns {
		tables[table] = "InnoDB"
		columns[table] = make(map[string]bool, len(requiredColumns))
		for _, column := range requiredColumns {
			columns[table][column] = true
		}
	}

	indexes := make(map[string]map[string][]string, len(requiredAuthUniqueIndexes))
	for table, requiredIndexes := range requiredAuthUniqueIndexes {
		indexes[table] = make(map[string][]string, len(requiredIndexes))
		for index, requiredColumns := range requiredIndexes {
			indexes[table][index] = append([]string(nil), requiredColumns...)
		}
	}
	return tables, columns, indexes
}

func TestValidateAuthSchemaSnapshotAcceptsCompleteSchema(t *testing.T) {
	tables, columns, indexes := completeAuthSchemaSnapshot()
	if err := validateAuthSchemaSnapshot(tables, columns, indexes); err != nil {
		t.Fatalf("complete schema rejected: %v", err)
	}
}

func TestValidateAuthSchemaSnapshotReportsUnsafePartialSchema(t *testing.T) {
	tables, columns, indexes := completeAuthSchemaSnapshot()
	tables["otp_user_mobile"] = "MyISAM"
	delete(columns["session_user_mobile"], "refresh_token_hash")
	delete(indexes["user_mobile"], "uk_user_email")

	err := validateAuthSchemaSnapshot(tables, columns, indexes)
	if err == nil {
		t.Fatal("unsafe partial schema was accepted")
	}
	for _, expected := range []string{
		"otp_user_mobile engine must be InnoDB",
		"missing column session_user_mobile.refresh_token_hash",
		"missing unique index user_mobile.uk_user_email",
	} {
		if !strings.Contains(err.Error(), expected) {
			t.Fatalf("validation error %q does not contain %q", err, expected)
		}
	}
}

func TestValidateAuthSchemaSnapshotRejectsMisnamedIndexColumns(t *testing.T) {
	tables, columns, indexes := completeAuthSchemaSnapshot()
	indexes["user_mobile"]["uk_user_email"] = []string{"username"}

	err := validateAuthSchemaSnapshot(tables, columns, indexes)
	if err == nil || !strings.Contains(err.Error(), "invalid unique index user_mobile.uk_user_email columns") {
		t.Fatalf("wrong index columns were accepted: %v", err)
	}
}
