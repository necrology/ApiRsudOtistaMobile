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



--tambah table user login
CREATE TABLE user (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,

    username VARCHAR(100) NOT NULL,

    email VARCHAR(100) NOT NULL UNIQUE,

    password VARCHAR(255) NOT NULL,

    is_deleted BOOLEAN DEFAULT FALSE,

    deleted_at TIMESTAMP NULL DEFAULT NULL,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    ON UPDATE CURRENT_TIMESTAMP
);

--tambah table otp_user
CREATE TABLE otp_user (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,

    user_id BIGINT NOT NULL,

    otp_code VARCHAR(10) NOT NULL,

    expired_at TIMESTAMP NOT NULL,

    is_used BOOLEAN DEFAULT FALSE,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id)
    REFERENCES user(id)
    ON DELETE CASCADE
);