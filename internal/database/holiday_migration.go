package database

import (
	"database/sql"
	"fmt"
)

func EnsureHolidaySchema(db *sql.DB, schema string) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS tanggal_libur_rs (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			tanggal DATE NOT NULL,
			nama_libur VARCHAR(180) NOT NULL,
			jenis ENUM('holiday','leave','manual') NOT NULL DEFAULT 'holiday',
			sumber VARCHAR(80) NOT NULL DEFAULT 'tanggalmerah.upset.dev',
			aktif TINYINT NOT NULL DEFAULT 1,
			raw_json JSON NULL,
			synced_at TIMESTAMP NULL DEFAULT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_tanggal_libur_rs (tanggal, jenis, nama_libur),
			KEY idx_tanggal_libur_rs_tanggal (tanggal, aktif),
			KEY idx_tanggal_libur_rs_jenis (jenis, aktif)
		)`,
	}

	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			return err
		}
	}

	indexes := []struct {
		table string
		index string
		ddl   string
	}{
		{"tanggal_libur_rs", "idx_tanggal_libur_rs_tanggal", "ALTER TABLE `tanggal_libur_rs` ADD KEY `idx_tanggal_libur_rs_tanggal` (`tanggal`, `aktif`)"},
		{"tanggal_libur_rs", "idx_tanggal_libur_rs_jenis", "ALTER TABLE `tanggal_libur_rs` ADD KEY `idx_tanggal_libur_rs_jenis` (`jenis`, `aktif`)"},
	}

	for _, index := range indexes {
		if err := ensureIndex(db, schema, index.table, index.index, index.ddl); err != nil {
			return fmt.Errorf("ensure index %s.%s: %w", index.table, index.index, err)
		}
	}

	return nil
}
