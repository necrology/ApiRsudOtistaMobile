-- History perubahan skema produksi API RSUD Otista Mobile.
--
-- Catatan penting:
-- 1. Database utama SIMRS sudah ada. Jangan membuat ulang database.
-- 2. Jangan menjalankan script import schema penuh di production.
-- 3. Jalankan query di file ini secara selektif sesuai kebutuhan, bukan asal run semua.
-- 4. Pastikan index/kolom belum ada sebelum menjalankan ALTER agar tidak error duplicate.
-- 5. Nama tabel auth mobile yang benar adalah `user_mobile`, bukan `user`.

-- ============================================================
-- 2026-05-29: Skema auth mobile sesuai database saat ini
-- ============================================================

CREATE TABLE IF NOT EXISTS `user_mobile` (
    `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
    `username` VARCHAR(100) NOT NULL,
    `email` VARCHAR(100) NOT NULL,
    `no_rm` VARCHAR(20) NULL,
    `patient_id` BIGINT UNSIGNED NULL,
    `phone` VARCHAR(30) NULL,
    `full_name` VARCHAR(150) NULL,
    `password` VARCHAR(255) NOT NULL,
    `email_verified` BOOLEAN DEFAULT FALSE,
    `verification_token` VARCHAR(255) NULL,
    `verified_at` TIMESTAMP NULL DEFAULT NULL,
    `medical_record_verified_at` TIMESTAMP NULL DEFAULT NULL,
    `is_deleted` BOOLEAN DEFAULT FALSE,
    `deleted_at` TIMESTAMP NULL DEFAULT NULL,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY `uk_user_email` (`email`),
    KEY `idx_user_mobile_no_rm` (`no_rm`),
    UNIQUE KEY `uk_user_patient_id` (`patient_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `otp_user_mobile` (
    `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
    `user_id` BIGINT NOT NULL,
    `otp_code` VARCHAR(10) NOT NULL,
    `expired_at` TIMESTAMP NOT NULL,
    `is_used` BOOLEAN DEFAULT FALSE,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT `fk_otp_user_mobile_user`
        FOREIGN KEY (`user_id`) REFERENCES `user_mobile` (`id`)
        ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `otp_verif_email_mobile` (
    `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
    `username` VARCHAR(100) NOT NULL,
    `email` VARCHAR(100) NOT NULL,
    `no_rm` VARCHAR(20) NULL,
    `phone` VARCHAR(30) NULL,
    `full_name` VARCHAR(150) NULL,
    `otp_code` VARCHAR(10) NOT NULL,
    `expired_at` TIMESTAMP NOT NULL,
    `is_used` BOOLEAN DEFAULT FALSE,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `otp_password_reset_mobile` (
    `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
    `user_id` BIGINT NOT NULL,
    `otp_code` VARCHAR(10) NOT NULL,
    `expired_at` TIMESTAMP NOT NULL,
    `is_used` BOOLEAN DEFAULT FALSE,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT `fk_otp_password_reset_mobile_user`
        FOREIGN KEY (`user_id`) REFERENCES `user_mobile` (`id`)
        ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `otp_medical_record_claim_mobile` (
    `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
    `user_id` BIGINT NOT NULL,
    `patient_id` BIGINT UNSIGNED NOT NULL,
    `no_rm` VARCHAR(20) NOT NULL,
    `patient_name` VARCHAR(150) NULL,
    `otp_code` VARCHAR(10) NOT NULL,
    `expired_at` TIMESTAMP NOT NULL,
    `is_used` BOOLEAN DEFAULT FALSE,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT `fk_otp_medical_record_claim_mobile_user`
        FOREIGN KEY (`user_id`) REFERENCES `user_mobile` (`id`)
        ON DELETE CASCADE,
    KEY `idx_otp_medical_record_claim_user` (`user_id`),
    KEY `idx_otp_medical_record_claim_no_rm` (`no_rm`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Kolom tambahan untuk database yang pernah dibuat dengan skema auth lama.
-- Jalankan hanya jika kolom belum tersedia.
ALTER TABLE `user_mobile`
    ADD COLUMN `no_rm` VARCHAR(20) NULL AFTER `email`;

ALTER TABLE `user_mobile`
    ADD COLUMN `patient_id` BIGINT UNSIGNED NULL AFTER `no_rm`;

ALTER TABLE `user_mobile`
    ADD COLUMN `phone` VARCHAR(30) NULL AFTER `patient_id`;

ALTER TABLE `user_mobile`
    ADD COLUMN `full_name` VARCHAR(150) NULL AFTER `phone`;

ALTER TABLE `user_mobile`
    ADD COLUMN `email_verified` BOOLEAN DEFAULT FALSE AFTER `password`;

ALTER TABLE `user_mobile`
    ADD COLUMN `verification_token` VARCHAR(255) NULL AFTER `email_verified`;

ALTER TABLE `user_mobile`
    ADD COLUMN `verified_at` TIMESTAMP NULL DEFAULT NULL AFTER `verification_token`;

ALTER TABLE `user_mobile`
    ADD COLUMN `medical_record_verified_at` TIMESTAMP NULL DEFAULT NULL AFTER `verified_at`;

ALTER TABLE `user_mobile`
    ADD COLUMN `is_deleted` BOOLEAN DEFAULT FALSE AFTER `medical_record_verified_at`;

ALTER TABLE `user_mobile`
    ADD COLUMN `deleted_at` TIMESTAMP NULL DEFAULT NULL AFTER `is_deleted`;

ALTER TABLE `user_mobile`
    ADD COLUMN `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP AFTER `deleted_at`;

ALTER TABLE `user_mobile`
    ADD COLUMN `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP AFTER `created_at`;

ALTER TABLE `otp_verif_email_mobile`
    ADD COLUMN `no_rm` VARCHAR(20) NULL AFTER `email`;

ALTER TABLE `otp_verif_email_mobile`
    ADD COLUMN `phone` VARCHAR(30) NULL AFTER `no_rm`;

ALTER TABLE `otp_verif_email_mobile`
    ADD COLUMN `full_name` VARCHAR(150) NULL AFTER `phone`;

-- Index auth mobile.
-- Jalankan hanya jika index belum tersedia.
ALTER TABLE `user_mobile`
    ADD KEY `idx_user_mobile_no_rm` (`no_rm`);

ALTER TABLE `user_mobile`
    ADD UNIQUE KEY `uk_user_patient_id` (`patient_id`);

ALTER TABLE `otp_user_mobile`
    ADD KEY `idx_otp_user_mobile_user_used_expired` (`user_id`, `is_used`, `expired_at`);

ALTER TABLE `otp_verif_email_mobile`
    ADD KEY `idx_otp_verif_email_mobile_email_used_expired` (`email`, `is_used`, `expired_at`);

ALTER TABLE `otp_password_reset_mobile`
    ADD KEY `idx_otp_password_reset_mobile_user_used_expired` (`user_id`, `is_used`, `expired_at`);

ALTER TABLE `otp_medical_record_claim_mobile`
    ADD KEY `idx_otp_medical_record_claim_mobile_user_used_expired` (`user_id`, `is_used`, `expired_at`);

-- Index lama `uk_user_no_rm` tidak dipakai lagi karena relasi pasien
-- memakai `user_mobile.patient_id -> pasiens.id`.
-- Jalankan hanya jika index tersebut masih ada.
ALTER TABLE `user_mobile`
    DROP INDEX `uk_user_no_rm`;

-- ============================================================
-- 2026-06-08: Booking umum dan antrian poli mobile
-- ============================================================
--
-- Skema existing yang dipakai:
-- - `user_mobile.patient_id` -> `pasiens.id`
-- - `registrasis_dummy.no_rm` -> `pasiens.no_rm`
-- - `registrasis_dummy.kode_poli` -> `polis.bpjs` atau `polis.id`
-- - `registrasis_dummy.dokter_id` -> `pegawais.id`
-- - `registrasis_dummy.registrasi_id` -> `registrasis.id` setelah diproses pendaftaran
--
-- API tidak membuat tabel antrian baru untuk fitur ini dan tidak langsung
-- membuat `registrasis` / `antrian_poli`.
-- Booking umum mobile masuk ke `registrasis_dummy` lebih dulu, seperti alur BPJS,
-- tetapi dibedakan dengan flag khusus:
-- - `flag` = 'mobile_umum'
-- - `jenisrequest` = NULL, mengikuti pola `android` umum existing
-- - `jenisdaftar` = 'android'
-- - `jenis_registrasi` = 'antrian'
-- - `status` default = 'pending'
--
-- Format nomor mengikuti pendaftaran umum Android existing:
-- - `registrasis_dummy.kodebooking` = NULL
-- - `registrasis_dummy.nomorantrian` = DDMMYYYY + kode_poli + angkaantrian
-- - `registrasis_dummy.angkaantrian` = angka murni
--
-- Endpoint terkait:
-- - POST /api/v1/mobile/booking/general
-- - GET  /api/v1/mobile/booking/general/mine
-- - GET  /api/v1/mobile/booking/options/{poli_id}
-- POST dan GET /general/mine wajib memakai Bearer session pasien.
-- GET /general lintas pasien tidak didaftarkan pada API publik.

-- Index pendukung list booking pendaftaran dan validasi 1 pasien 1 booking per hari.
-- Jalankan hanya jika index belum tersedia.
ALTER TABLE `registrasis_dummy`
    ADD INDEX `idx_registrasis_dummy_mobile_umum_patient_daily` (
        `flag`,
        `no_rm`,
        `tglperiksa`,
        `status`
    );

ALTER TABLE `registrasis_dummy`
    ADD INDEX `idx_registrasis_dummy_mobile_umum_poli_daily` (
        `flag`,
        `kode_poli`,
        `tglperiksa`,
        `angkaantrian`,
        `status`
    );

ALTER TABLE `registrasis_dummy`
    ADD INDEX `idx_registrasis_dummy_mobile_umum_registrasi` (
        `flag`,
        `registrasi_id`
    );

-- Index legacy untuk data booking umum mobile lama yang sudah pernah masuk
-- langsung ke `registrasis` dan `antrian_poli` sebelum alur dummy dipakai.
-- Jalankan hanya jika masih perlu membaca data legacy dan index belum tersedia.
ALTER TABLE `registrasis`
    ADD INDEX `idx_registrasis_mobile_booking` (
        `input_from`,
        `pasien_id`,
        `poli_id`,
        `tgl_order`,
        `deleted_at`
    );

ALTER TABLE `antrian_poli`
    ADD INDEX `idx_antrian_poli_mobile_booking` (
        `poli_id`,
        `tanggal`,
        `status`,
        `panggil`,
        `nomor`
    );

-- ============================================================
-- 2026-06-10: History antrian, resep, lab, dan radiologi mobile
-- ============================================================

-- Index tambahan untuk data legacy booking umum mobile yang pernah langsung
-- masuk ke `registrasis`. Alur aktif saat ini memakai `registrasis_dummy`
-- dengan `idx_registrasis_dummy_mobile_umum_patient_daily` di section 2026-06-08.
-- Index ini juga membantu halaman history nomor antrian pasien untuk data lama.
-- Jalankan hanya jika index belum tersedia.
ALTER TABLE `registrasis`
    ADD INDEX `idx_registrasis_mobile_patient_daily` (
        `pasien_id`,
        `tgl_order`,
        `input_from`,
        `jkn`,
        `deleted_at`
    );

-- Index pendukung resep/obat pasien dari `penjualans` dan `penjualandetails`.
-- Jalankan hanya jika index belum tersedia.
ALTER TABLE `penjualans`
    ADD INDEX `idx_penjualans_mobile_patient` (
        `registrasi_id`,
        `deleted_at`,
        `created_at`
    );

ALTER TABLE `penjualandetails`
    ADD INDEX `idx_penjualandetails_mobile_prescription` (
        `penjualan_id`,
        `deleted_at`
    );

-- Index pendukung hasil lab pasien.
-- Jalankan hanya jika index belum tersedia.
ALTER TABLE `hasillabs`
    ADD INDEX `idx_hasillabs_mobile_registration` (
        `registrasi_id`,
        `deleted_at`,
        `tgl_hasilselesai`,
        `tgl_pemeriksaan`
    );

-- Index pendukung radiologi pasien.
-- Jalankan hanya jika index belum tersedia.
ALTER TABLE `order_radiologi`
    ADD INDEX `idx_order_radiologi_mobile_registration` (
        `registrasi_id`,
        `created_at`
    );

ALTER TABLE `hasilradiologis`
    ADD INDEX `idx_hasilradiologis_mobile_registration` (
        `registrasi_id`,
        `created_at`
    );

ALTER TABLE `radiologi_ekspertises`
    ADD INDEX `idx_radiologi_ekspertises_mobile_patient` (
        `registrasi_id`,
        `pasien_id`,
        `tanggal_eksp`
    );

-- ============================================================
-- 2026-06-21: Kalender libur booking umum mobile
-- ============================================================
--
-- Data utama hari libur nasional dan cuti bersama disinkronkan dari
-- Tanggal Merah API / upset.dev oleh API backend, lalu disimpan sebagai cache
-- lokal di `tanggal_libur_rs`.
--
-- Aturan booking umum mobile:
-- - Hari libur nasional (`jenis='holiday'`) tidak bisa dipakai booking.
-- - Cuti bersama (`jenis='leave'`) tidak bisa dipakai booking.
-- - Hari Minggu tidak bisa dipakai booking.
-- - Data manual dapat ditambahkan dengan `jenis='manual'` dan `aktif=1`.
--
-- Endpoint terkait:
-- - GET /api/v1/mobile/booking/calendar?year=2026&month=6&poli_id=1
-- - POST /api/v1/mobile/booking/general

CREATE TABLE IF NOT EXISTS `tanggal_libur_rs` (
    `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
    `tanggal` DATE NOT NULL,
    `nama_libur` VARCHAR(180) NOT NULL,
    `jenis` ENUM('holiday', 'leave', 'manual') NOT NULL DEFAULT 'holiday',
    `sumber` VARCHAR(80) NOT NULL DEFAULT 'tanggalmerah.upset.dev',
    `aktif` TINYINT NOT NULL DEFAULT 1,
    `raw_json` JSON NULL,
    `synced_at` TIMESTAMP NULL DEFAULT NULL,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY `uk_tanggal_libur_rs` (`tanggal`, `jenis`, `nama_libur`),
    KEY `idx_tanggal_libur_rs_tanggal` (`tanggal`, `aktif`),
    KEY `idx_tanggal_libur_rs_jenis` (`jenis`, `aktif`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- 2026-06-22: Jadwal gabungan booking dan hasil lab LIS mobile
-- ============================================================
--
-- Jadwal booking mobile sekarang menggabungkan jadwal poliklinik dari `polis`
-- dengan jadwal dokter dari `jadwaldokters`.
--
-- Hasil laboratorium mobile memprioritaskan JSON LIS dari `lica_results`,
-- lalu tetap memakai metadata pasien dari `hasillabs` -> `pasiens`.
--
-- Jalankan hanya jika index belum tersedia.

ALTER TABLE `jadwaldokters`
    ADD INDEX `idx_jadwaldokters_mobile_poli_dokter` (
        `poli`,
        `dokter`
    );

ALTER TABLE `lica_results`
    ADD INDEX `idx_lica_results_mobile_no_lab_tgl` (
        `no_lab`,
        `tgl_pemeriksaan`,
        `id`
    );

ALTER TABLE `hasillabs`
    ADD INDEX `idx_hasillabs_mobile_patient_lis` (
        `pasien_id`,
        `deleted_at`,
        `no_lab`,
        `tgl_hasilselesai`,
        `tgl_pemeriksaan`
    );

-- ============================================================
-- 2026-07-15: Session bearer, tiket registrasi, dan hardening OTP mobile
-- ============================================================
--
-- WAJIB DIBACA OPERATOR:
-- 1. Backup tabel `user_mobile` dan seluruh tabel `otp_*_mobile` lebih dulu.
-- 2. CREATE TABLE aman dijalankan ulang karena memakai IF NOT EXISTS.
-- 3. ALTER TABLE di bawah dijalankan hanya untuk kolom yang belum tersedia.
--    Binary API hanya memvalidasi information_schema dan tidak menjalankan DDL.
--    Gunakan `go run ./cmd/migrate` atau terapkan statement ini secara terkontrol.
-- 4. Jalankan pada jam traffic rendah karena ALTER dapat menunggu metadata lock.
-- 5. Jangan menyimpan atau mencatat token/OTP mentah di log deployment.

CREATE TABLE IF NOT EXISTS `session_user_mobile` (
    `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
    `family_id` VARCHAR(64) NOT NULL,
    `parent_session_id` BIGINT NULL,
    `user_id` BIGINT NOT NULL,
    `access_token_hash` BINARY(32) NOT NULL,
    `refresh_token_hash` BINARY(32) NOT NULL,
    `access_expires_at` DATETIME(6) NOT NULL,
    `refresh_expires_at` DATETIME(6) NOT NULL,
    `rotated_at` DATETIME(6) NULL DEFAULT NULL,
    `revoked_at` DATETIME(6) NULL DEFAULT NULL,
    `revoke_reason` VARCHAR(100) NULL,
    `replaced_by_session_id` BIGINT NULL,
    `created_at` DATETIME(6) NOT NULL,
    `updated_at` DATETIME(6) NOT NULL,
    UNIQUE KEY `uk_session_user_mobile_access_hash` (`access_token_hash`),
    UNIQUE KEY `uk_session_user_mobile_refresh_hash` (`refresh_token_hash`),
    KEY `idx_session_user_mobile_user_active`
        (`user_id`, `revoked_at`, `refresh_expires_at`),
    KEY `idx_session_user_mobile_family_active`
        (`family_id`, `revoked_at`),
    KEY `idx_session_user_mobile_parent` (`parent_session_id`),
    KEY `idx_session_user_mobile_replacement` (`replaced_by_session_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `auth_ticket_mobile` (
    `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
    `verification_id` BIGINT NOT NULL,
    `email` VARCHAR(100) NOT NULL,
    `purpose` VARCHAR(40) NOT NULL,
    `ticket_hash` BINARY(32) NOT NULL,
    `expires_at` DATETIME(6) NOT NULL,
    `used_at` DATETIME(6) NULL DEFAULT NULL,
    `revoked_at` DATETIME(6) NULL DEFAULT NULL,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY `uk_auth_ticket_mobile_hash` (`ticket_hash`),
    KEY `idx_auth_ticket_mobile_email_active`
        (`email`, `purpose`, `used_at`, `expires_at`),
    KEY `idx_auth_ticket_mobile_verification` (`verification_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Transaksi, SELECT ... FOR UPDATE, rotasi refresh, dan attempt counter OTP
-- memerlukan InnoDB. Jalankan statement berikut hanya untuk tabel yang hasil
-- preflight ENGINE-nya bukan InnoDB.
ALTER TABLE `user_mobile` ENGINE=InnoDB;
ALTER TABLE `otp_user_mobile` ENGINE=InnoDB;
ALTER TABLE `otp_verif_email_mobile` ENGINE=InnoDB;
ALTER TABLE `otp_password_reset_mobile` ENGINE=InnoDB;
ALTER TABLE `otp_medical_record_claim_mobile` ENGINE=InnoDB;
ALTER TABLE `session_user_mobile` ENGINE=InnoDB;
ALTER TABLE `auth_ticket_mobile` ENGINE=InnoDB;

-- Indeks wajib. CREATE TABLE di atas sudah membentuk indeks untuk instalasi baru.
-- Statement berikut hanya diperlukan untuk tabel lama/parsial yang nama indeksnya
-- belum tersedia. Periksa information_schema.STATISTICS sebelum menjalankannya.

ALTER TABLE `user_mobile`
    ADD UNIQUE INDEX `uk_user_email` (`email`);

ALTER TABLE `user_mobile`
    ADD UNIQUE INDEX `uk_user_patient_id` (`patient_id`);

ALTER TABLE `session_user_mobile`
    ADD UNIQUE INDEX `uk_session_user_mobile_access_hash` (`access_token_hash`);

ALTER TABLE `session_user_mobile`
    ADD UNIQUE INDEX `uk_session_user_mobile_refresh_hash` (`refresh_token_hash`);

ALTER TABLE `session_user_mobile`
    ADD INDEX `idx_session_user_mobile_user_active`
        (`user_id`, `revoked_at`, `refresh_expires_at`);

ALTER TABLE `session_user_mobile`
    ADD INDEX `idx_session_user_mobile_family_active`
        (`family_id`, `revoked_at`);

ALTER TABLE `session_user_mobile`
    ADD INDEX `idx_session_user_mobile_parent` (`parent_session_id`);

ALTER TABLE `session_user_mobile`
    ADD INDEX `idx_session_user_mobile_replacement` (`replaced_by_session_id`);

ALTER TABLE `auth_ticket_mobile`
    ADD UNIQUE INDEX `uk_auth_ticket_mobile_hash` (`ticket_hash`);

ALTER TABLE `auth_ticket_mobile`
    ADD INDEX `idx_auth_ticket_mobile_email_active`
        (`email`, `purpose`, `used_at`, `expires_at`);

ALTER TABLE `auth_ticket_mobile`
    ADD INDEX `idx_auth_ticket_mobile_verification` (`verification_id`);

-- Kolom OTP dan snapshot klaim baru. Setiap ADD COLUMN dipisah agar database yang sudah termigrasi
-- sebagian dapat menjalankan hanya statement untuk kolom yang belum ada.
-- `otp_code` sengaja dipertahankan untuk rollback singkat; binary baru menulis
-- string kosong ke `otp_code` dan bcrypt ke `otp_hash`.

ALTER TABLE `otp_medical_record_claim_mobile`
    ADD COLUMN `patient_name` VARCHAR(150) NULL AFTER `no_rm`;

ALTER TABLE `otp_user_mobile`
    ADD COLUMN `otp_hash` VARCHAR(255) NULL AFTER `otp_code`;

ALTER TABLE `otp_user_mobile`
    ADD COLUMN `attempt_count` SMALLINT UNSIGNED NOT NULL DEFAULT 0 AFTER `otp_hash`;

ALTER TABLE `otp_user_mobile`
    ADD COLUMN `last_attempt_at` DATETIME(6) NULL DEFAULT NULL AFTER `attempt_count`;

ALTER TABLE `otp_user_mobile`
    ADD COLUMN `locked_at` DATETIME(6) NULL DEFAULT NULL AFTER `last_attempt_at`;

ALTER TABLE `otp_user_mobile`
    ADD COLUMN `used_at` DATETIME(6) NULL DEFAULT NULL AFTER `locked_at`;

ALTER TABLE `otp_verif_email_mobile`
    ADD COLUMN `otp_hash` VARCHAR(255) NULL AFTER `otp_code`;

ALTER TABLE `otp_verif_email_mobile`
    ADD COLUMN `attempt_count` SMALLINT UNSIGNED NOT NULL DEFAULT 0 AFTER `otp_hash`;

ALTER TABLE `otp_verif_email_mobile`
    ADD COLUMN `last_attempt_at` DATETIME(6) NULL DEFAULT NULL AFTER `attempt_count`;

ALTER TABLE `otp_verif_email_mobile`
    ADD COLUMN `locked_at` DATETIME(6) NULL DEFAULT NULL AFTER `last_attempt_at`;

ALTER TABLE `otp_verif_email_mobile`
    ADD COLUMN `used_at` DATETIME(6) NULL DEFAULT NULL AFTER `locked_at`;

ALTER TABLE `otp_password_reset_mobile`
    ADD COLUMN `otp_hash` VARCHAR(255) NULL AFTER `otp_code`;

ALTER TABLE `otp_password_reset_mobile`
    ADD COLUMN `attempt_count` SMALLINT UNSIGNED NOT NULL DEFAULT 0 AFTER `otp_hash`;

ALTER TABLE `otp_password_reset_mobile`
    ADD COLUMN `last_attempt_at` DATETIME(6) NULL DEFAULT NULL AFTER `attempt_count`;

ALTER TABLE `otp_password_reset_mobile`
    ADD COLUMN `locked_at` DATETIME(6) NULL DEFAULT NULL AFTER `last_attempt_at`;

ALTER TABLE `otp_password_reset_mobile`
    ADD COLUMN `used_at` DATETIME(6) NULL DEFAULT NULL AFTER `locked_at`;

ALTER TABLE `otp_medical_record_claim_mobile`
    ADD COLUMN `otp_hash` VARCHAR(255) NULL AFTER `otp_code`;

ALTER TABLE `otp_medical_record_claim_mobile`
    ADD COLUMN `attempt_count` SMALLINT UNSIGNED NOT NULL DEFAULT 0 AFTER `otp_hash`;

ALTER TABLE `otp_medical_record_claim_mobile`
    ADD COLUMN `last_attempt_at` DATETIME(6) NULL DEFAULT NULL AFTER `attempt_count`;

ALTER TABLE `otp_medical_record_claim_mobile`
    ADD COLUMN `locked_at` DATETIME(6) NULL DEFAULT NULL AFTER `last_attempt_at`;

ALTER TABLE `otp_medical_record_claim_mobile`
    ADD COLUMN `used_at` DATETIME(6) NULL DEFAULT NULL AFTER `locked_at`;

-- ============================================================
-- 2026-08-27: OTP penghapusan akun SIPANTES OTISTA
-- ============================================================
-- Backup tabel auth mobile sebelum migrasi. Tabel ini hanya menyimpan OTP
-- ter-hash untuk konfirmasi penghapusan akun dan tidak menyimpan rekam medis.

CREATE TABLE IF NOT EXISTS `otp_account_deletion_mobile` (
    `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
    `user_id` BIGINT NOT NULL,
    `otp_code` VARCHAR(10) NOT NULL,
    `otp_hash` VARCHAR(255) NULL,
    `attempt_count` SMALLINT UNSIGNED NOT NULL DEFAULT 0,
    `last_attempt_at` DATETIME(6) NULL DEFAULT NULL,
    `locked_at` DATETIME(6) NULL DEFAULT NULL,
    `used_at` DATETIME(6) NULL DEFAULT NULL,
    `expired_at` TIMESTAMP NOT NULL,
    `is_used` BOOLEAN DEFAULT FALSE,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT `fk_otp_account_deletion_user_mobile`
        FOREIGN KEY (`user_id`) REFERENCES `user_mobile` (`id`) ON DELETE CASCADE,
    KEY `idx_otp_account_deletion_user_used_expired`
        (`user_id`, `is_used`, `expired_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
