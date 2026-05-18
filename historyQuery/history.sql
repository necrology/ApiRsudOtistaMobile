-- File ini untuk menyimpan query perubahan pada tabel / buat tabel baru

-- Query table test
CREATE TABLE table_test (
    id INT AUTO_INCREMENT PRIMARY KEY,
    nama VARCHAR(50) NOT NULL,
    usia INT,
    tanggal_daftar DATE
);

-- Query alter table test
ALTER TABLE test ADD email VARCHAR(100) AFTER nama;

