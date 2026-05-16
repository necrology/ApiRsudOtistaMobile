SET NAMES utf8mb4;
SET GLOBAL local_infile = 1;

CREATE DATABASE IF NOT EXISTS `apirusdotistamobile`
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_unicode_ci;

USE `apirusdotistamobile`;

SET FOREIGN_KEY_CHECKS = 0;

DROP TABLE IF EXISTS `pasiens`;
DROP TABLE IF EXISTS `pegawais`;
DROP TABLE IF EXISTS `polis`;
DROP TABLE IF EXISTS `kamars`;

CREATE TABLE `pegawais` (
  `id` INT UNSIGNED NOT NULL,
  `nama` VARCHAR(100) NULL,
  `nik` VARCHAR(32) NULL,
  `kuota_poli` VARCHAR(10) NULL,
  `kategori_pegawai` VARCHAR(10) NULL,
  `kelompok_pegawai` VARCHAR(10) NULL,
  `subkelompok_pegawai` VARCHAR(10) NULL,
  `poli_id` VARCHAR(10) NULL,
  `poli_type` VARCHAR(10) NULL,
  `kode_bpjs` VARCHAR(20) NULL,
  `kode_antrian` VARCHAR(20) NULL,
  `kode_dokter_inhealth` VARCHAR(20) NULL,
  `general_code` VARCHAR(20) NULL,
  `tgllahir` VARCHAR(20) NULL,
  `tmplahir` VARCHAR(50) NULL,
  `kelamin` VARCHAR(10) NULL,
  `agama` VARCHAR(30) NULL,
  `alamat` TEXT NULL,
  `sip` VARCHAR(80) NULL,
  `str` VARCHAR(80) NULL,
  `kompetensi` VARCHAR(30) NULL,
  `no_reg` VARCHAR(30) NULL,
  `tupoksi` VARCHAR(30) NULL,
  `sip_awal` VARCHAR(20) NULL,
  `sip_akhir` VARCHAR(20) NULL,
  `user_id` VARCHAR(20) NULL,
  `id_dokterss` VARCHAR(30) NULL,
  `tanda_tangan` VARCHAR(150) NULL,
  `foto_profile` VARCHAR(150) NULL,
  `nip` VARCHAR(40) NULL,
  `golongan` VARCHAR(20) NULL,
  `golongan_tmt` VARCHAR(20) NULL,
  `jabatan` VARCHAR(120) NULL,
  `jabatan_tmt` VARCHAR(20) NULL,
  `status_tte` VARCHAR(10) NULL,
  `smf` VARCHAR(10) NULL,
  `created_at` VARCHAR(30) NULL,
  `updated_at` VARCHAR(30) NULL,
  `is_dokter` VARCHAR(10) NULL,
  `deleted_at` VARCHAR(30) NULL,
  `user_deleted` VARCHAR(10) NULL,
  PRIMARY KEY (`id`),
  KEY `idx_pegawais_nama` (`nama`),
  KEY `idx_pegawais_nik` (`nik`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `polis` (
  `id` INT UNSIGNED NOT NULL,
  `nama` VARCHAR(80) NULL,
  `audio` VARCHAR(50) NULL,
  `kelompok` VARCHAR(20) NULL,
  `lantai` VARCHAR(10) NULL,
  `kelas` VARCHAR(10) NULL,
  `kode_ruangan` VARCHAR(20) NULL,
  `general_code` VARCHAR(20) NULL,
  `urutan` VARCHAR(10) NULL,
  `politype` VARCHAR(10) NULL,
  `flag` VARCHAR(10) NULL,
  `bpjs` VARCHAR(20) NULL,
  `inhealth` VARCHAR(20) NULL,
  `instalasi_id` VARCHAR(20) NULL,
  `kamar_id` VARCHAR(20) NULL,
  `dokter_id` TEXT NULL,
  `perawat_id` TEXT NULL,
  `kuota` VARCHAR(10) NULL,
  `kuota_online` VARCHAR(10) NULL,
  `terisi` VARCHAR(10) NULL,
  `loket` VARCHAR(10) NULL,
  `kode_loket` VARCHAR(10) NULL,
  `posisi_loket` VARCHAR(20) NULL,
  `buka` VARCHAR(20) NULL,
  `tutup` VARCHAR(20) NULL,
  `praktik` VARCHAR(10) NULL,
  `sunday` VARCHAR(10) NULL,
  `monday` VARCHAR(10) NULL,
  `tuesday` VARCHAR(10) NULL,
  `wednesday` VARCHAR(10) NULL,
  `thursday` VARCHAR(10) NULL,
  `friday` VARCHAR(10) NULL,
  `saturday` VARCHAR(10) NULL,
  `jkn_sunday` VARCHAR(10) NULL,
  `jkn_monday` VARCHAR(10) NULL,
  `jkn_tuesday` VARCHAR(10) NULL,
  `jkn_wednesday` VARCHAR(10) NULL,
  `jkn_thursday` VARCHAR(10) NULL,
  `jkn_friday` VARCHAR(10) NULL,
  `jkn_saturday` VARCHAR(10) NULL,
  `layar_lcd` VARCHAR(10) NULL,
  `created_at` VARCHAR(30) NULL,
  `updated_at` VARCHAR(30) NULL,
  `deleted_at` VARCHAR(30) NULL,
  `id_location_ss` VARCHAR(60) NULL,
  `description` VARCHAR(100) NULL,
  `satusehat_room` VARCHAR(10) NULL,
  PRIMARY KEY (`id`),
  KEY `idx_polis_nama` (`nama`),
  KEY `idx_polis_bpjs` (`bpjs`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `kamars` (
  `id` INT UNSIGNED NOT NULL,
  `nama` VARCHAR(80) NULL,
  `kode` VARCHAR(30) NULL,
  `id_location_ss` VARCHAR(60) NULL,
  `kelas_id` VARCHAR(10) NULL,
  `kelompokkelas_id` VARCHAR(10) NULL,
  `conf_rl31_id` VARCHAR(10) NULL,
  `created_at` VARCHAR(30) NULL,
  `updated_at` VARCHAR(30) NULL,
  `deleted_at` VARCHAR(30) NULL,
  `hidden` VARCHAR(10) NULL,
  PRIMARY KEY (`id`),
  KEY `idx_kamars_nama` (`nama`),
  KEY `idx_kamars_kode` (`kode`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `pasiens` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `no_rm` VARCHAR(20) NOT NULL,
  `nik` VARCHAR(16) NOT NULL,
  `nama` VARCHAR(120) NOT NULL,
  `jenis_kelamin` ENUM('L','P') NOT NULL,
  `tempat_lahir` VARCHAR(80) NULL,
  `tanggal_lahir` DATE NULL,
  `alamat` TEXT NULL,
  `no_hp` VARCHAR(20) NULL,
  `golongan_darah` ENUM('A','B','AB','O') NULL,
  `agama` VARCHAR(30) NULL,
  `status_pernikahan` ENUM('Belum Kawin','Kawin','Cerai') NULL,
  `pekerjaan` VARCHAR(80) NULL,
  `penanggung_jawab` VARCHAR(120) NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_pasiens_no_rm` (`no_rm`),
  UNIQUE KEY `uk_pasiens_nik` (`nik`),
  KEY `idx_pasiens_nama` (`nama`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

LOAD DATA LOCAL INFILE 'C:/Users/Maulana PC/Downloads/pegawais.csv'
INTO TABLE `pegawais`
CHARACTER SET utf8mb4
FIELDS TERMINATED BY ',' OPTIONALLY ENCLOSED BY '"'
LINES TERMINATED BY '\r\n'
IGNORE 1 LINES
(`id`, `nama`, `nik`, `kuota_poli`, `kategori_pegawai`, `kelompok_pegawai`, `subkelompok_pegawai`,
 `poli_id`, `poli_type`, `kode_bpjs`, `kode_antrian`, `kode_dokter_inhealth`, `general_code`,
 `tgllahir`, `tmplahir`, `kelamin`, `agama`, `alamat`, `sip`, `str`, `kompetensi`, `no_reg`,
 `tupoksi`, `sip_awal`, `sip_akhir`, `user_id`, `id_dokterss`, `tanda_tangan`, `foto_profile`,
 `nip`, `golongan`, `golongan_tmt`, `jabatan`, `jabatan_tmt`, `status_tte`, `smf`, `created_at`,
 `updated_at`, `is_dokter`, `deleted_at`, `user_deleted`);

LOAD DATA LOCAL INFILE 'C:/Users/Maulana PC/Downloads/polis.csv'
INTO TABLE `polis`
CHARACTER SET utf8mb4
FIELDS TERMINATED BY ',' OPTIONALLY ENCLOSED BY '"'
LINES TERMINATED BY '\r\n'
IGNORE 1 LINES
(`id`, `nama`, `audio`, `kelompok`, `lantai`, `kelas`, `kode_ruangan`, `general_code`, `urutan`,
 `politype`, `flag`, `bpjs`, `inhealth`, `instalasi_id`, `kamar_id`, `dokter_id`, `perawat_id`,
 `kuota`, `kuota_online`, `terisi`, `loket`, `kode_loket`, `posisi_loket`, `buka`, `tutup`,
 `praktik`, `sunday`, `monday`, `tuesday`, `wednesday`, `thursday`, `friday`, `saturday`,
 `jkn_sunday`, `jkn_monday`, `jkn_tuesday`, `jkn_wednesday`, `jkn_thursday`, `jkn_friday`,
 `jkn_saturday`, `layar_lcd`, `created_at`, `updated_at`, `deleted_at`, `id_location_ss`,
 `description`, `satusehat_room`);

LOAD DATA LOCAL INFILE 'C:/Users/Maulana PC/Downloads/kamars.csv'
INTO TABLE `kamars`
CHARACTER SET utf8mb4
FIELDS TERMINATED BY ',' OPTIONALLY ENCLOSED BY '"'
LINES TERMINATED BY '\r\n'
IGNORE 1 LINES
(`id`, `nama`, `kode`, `id_location_ss`, `kelas_id`, `kelompokkelas_id`, `conf_rl31_id`,
 `created_at`, `updated_at`, `deleted_at`, `hidden`);

INSERT INTO `pasiens` (
  `no_rm`, `nik`, `nama`, `jenis_kelamin`, `tempat_lahir`, `tanggal_lahir`, `alamat`, `no_hp`,
  `golongan_darah`, `agama`, `status_pernikahan`, `pekerjaan`, `penanggung_jawab`, `created_at`, `updated_at`
)
SELECT
  CONCAT('RM', LPAD(seq.n, 6, '0')) AS `no_rm`,
  CONCAT('3273', LPAD(seq.n, 12, '0')) AS `nik`,
  CONCAT(
    ELT(MOD(seq.n, 20) + 1, 'Ahmad', 'Siti', 'Dewi', 'Rizky', 'Budi', 'Asep', 'Nur', 'Rina', 'Dian', 'Agus',
                              'Fitri', 'Taufik', 'Sri', 'Yusuf', 'Lina', 'Dedi', 'Maya', 'Hendra', 'Nia', 'Fajar'),
    ' ',
    ELT(MOD(seq.n * 3, 20) + 1, 'Pratama', 'Salsabila', 'Saputra', 'Permata', 'Wijaya', 'Rahmawati', 'Kurniawan',
                                  'Lestari', 'Maulana', 'Anggraini', 'Hidayat', 'Putri', 'Santoso', 'Utami',
                                  'Firmansyah', 'Maharani', 'Nugraha', 'Oktaviani', 'Setiawan', 'Wulandari'),
    ' ',
    LPAD(seq.n, 3, '0')
  ) AS `nama`,
  IF(MOD(seq.n, 2) = 0, 'P', 'L') AS `jenis_kelamin`,
  ELT(MOD(seq.n, 10) + 1, 'Bandung', 'Garut', 'Tasikmalaya', 'Cimahi', 'Sumedang',
                            'Cianjur', 'Bogor', 'Subang', 'Purwakarta', 'Sukabumi') AS `tempat_lahir`,
  DATE_SUB(DATE_SUB(CURDATE(), INTERVAL (18 + MOD(seq.n, 55)) YEAR), INTERVAL MOD(seq.n * 7, 365) DAY) AS `tanggal_lahir`,
  CONCAT('Jl. Sehat Dummy No. ', seq.n, ', RT ', LPAD(MOD(seq.n, 9) + 1, 2, '0'),
         '/RW ', LPAD(MOD(seq.n, 7) + 1, 2, '0'), ', Jawa Barat') AS `alamat`,
  CONCAT('08', LPAD(seq.n, 10, '0')) AS `no_hp`,
  ELT(MOD(seq.n, 4) + 1, 'A', 'B', 'AB', 'O') AS `golongan_darah`,
  ELT(MOD(seq.n, 5) + 1, 'Islam', 'Kristen', 'Katolik', 'Hindu', 'Buddha') AS `agama`,
  ELT(MOD(seq.n, 3) + 1, 'Belum Kawin', 'Kawin', 'Cerai') AS `status_pernikahan`,
  ELT(MOD(seq.n, 8) + 1, 'Pelajar', 'Mahasiswa', 'Karyawan Swasta', 'Wiraswasta',
                            'PNS', 'Guru', 'Ibu Rumah Tangga', 'Petani') AS `pekerjaan`,
  CONCAT(
    ELT(MOD(seq.n * 5, 12) + 1, 'Aminah', 'Jajang', 'Ratna', 'Suryadi', 'Yuli', 'Cecep',
                                'Murni', 'Dadang', 'Rohmah', 'Wawan', 'Neneng', 'Usep'),
    ' ',
    LPAD(seq.n, 3, '0')
  ) AS `penanggung_jawab`,
  NOW() AS `created_at`,
  NOW() AS `updated_at`
FROM (
  SELECT ones.n + tens.n * 10 + 1 AS n
  FROM (
    SELECT 0 AS n UNION ALL SELECT 1 UNION ALL SELECT 2 UNION ALL SELECT 3 UNION ALL SELECT 4
    UNION ALL SELECT 5 UNION ALL SELECT 6 UNION ALL SELECT 7 UNION ALL SELECT 8 UNION ALL SELECT 9
  ) AS ones
  CROSS JOIN (
    SELECT 0 AS n UNION ALL SELECT 1 UNION ALL SELECT 2 UNION ALL SELECT 3 UNION ALL SELECT 4
    UNION ALL SELECT 5 UNION ALL SELECT 6 UNION ALL SELECT 7 UNION ALL SELECT 8 UNION ALL SELECT 9
  ) AS tens
) AS seq
ORDER BY seq.n;

SET FOREIGN_KEY_CHECKS = 1;
