# Deployment Hardening Auth Mobile

Dokumen ini adalah runbook operator untuk merilis hardening autentikasi API RSUD Otista Mobile. Perubahan yang dijelaskan berada di source Go dan skema MySQL. Pengerjaan source **tidak** mengubah Cloudflare, reverse proxy, firewall, DNS, Docker host, sertifikat TLS, atau konfigurasi server produksi.

Target API produksi:

```text
https://api-mobile.rsudotista.my.id/api/v1
```

## Ringkasan perubahan

- Endpoint pasien dan booking milik pasien memakai identitas dari sesi server, bukan `email`, `identifier`, atau `no_rm` dari query/body.
- Access token dan refresh token bersifat opaque (acak), sedangkan database hanya menyimpan hash SHA-256 token.
- Access token berlaku 15 menit dan keluarga refresh token memiliki batas absolut 30 hari sejak login.
- Refresh token dirotasi setiap dipakai. Pemakaian ulang refresh token lama mencabut seluruh keluarga sesi.
- Refresh token yang dipanggil kembali kurang dari dua detik setelah diterbitkan mendapat `429` dan `Retry-After: 2` untuk membatasi churn sesi.
- OTP baru disimpan sebagai hash bcrypt, dibatasi lima percobaan, serta memiliki waktu penggunaan dan waktu penguncian.
- Registrasi dua tahap memakai `registration_ticket` sekali pakai. Email saja tidak lagi cukup untuk memanggil `set-password`.
- Pengiriman email OTP memakai dua worker dan antrean tetap 64 job; request baru ditolak `503` bila SMTP tidak dikonfigurasi atau antrean penuh. Saat shutdown normal, API menutup antrean dan mencoba menghabiskan job selama maksimal delapan detik.
- Operasi auth mahal dibatasi maksimal empat request serentak per proses untuk menahan burst bcrypt; overflow mendapat `503` dan `Retry-After: 1`.
- Generic table explorer dan endpoint daftar seluruh booking tidak boleh tersedia pada API publik.

## Perubahan skema

DDL resmi berada pada bagian `2026-07-15: Session bearer, tiket registrasi, dan hardening OTP mobile` di `historyQuery/history.sql`.

Tabel baru:

- `session_user_mobile`: hash access/refresh token, masa berlaku, relasi rotasi, dan status pencabutan sesi.
- `auth_ticket_mobile`: hash tiket registrasi sekali pakai, tujuan tiket, masa berlaku, dan status pemakaian/pencabutan.

Kolom berikut ditambahkan pada masing-masing tabel `otp_user_mobile`, `otp_verif_email_mobile`, `otp_password_reset_mobile`, dan `otp_medical_record_claim_mobile`:

- `otp_hash VARCHAR(255) NULL`
- `attempt_count SMALLINT UNSIGNED NOT NULL DEFAULT 0`
- `last_attempt_at DATETIME(6) NULL`
- `locked_at DATETIME(6) NULL`
- `used_at DATETIME(6) NULL`

Tabel `otp_medical_record_claim_mobile` juga memerlukan `patient_name VARCHAR(150) NULL` untuk snapshot nama pasien pada proses klaim.

Kolom legacy `otp_code` sengaja belum dihapus untuk memudahkan rollback singkat. Binary baru menulis nilai kosong ke `otp_code` dan hash bcrypt ke `otp_hash`. Jangan mengubah nama tabel baru: seluruh tabel khusus fitur ini memakai akhiran `_mobile`.

Binary tidak me-rename tabel generik seperti `user`, `otp_user`, atau tabel inti SIMRS. Jika ada instalasi legacy yang memakai nama tersebut, migrasi datanya harus dilakukan operator setelah memverifikasi struktur dan kepemilikan tabel; jangan memakai `RENAME TABLE` otomatis pada startup.

## Prasyarat dan preflight

1. Jadwalkan maintenance pada jam traffic rendah. `ALTER TABLE` dapat menunggu metadata lock.
2. Pastikan artifact binary/image dibangun dari commit yang sama dengan `historyQuery/history.sql` yang akan diterapkan.
3. Simpan checksum/tag image dan binary versi lama untuk rollback.
4. Pastikan kapasitas disk database dan lokasi backup mencukupi.
5. Jangan mencatat OTP, access token, refresh token, `registration_ticket`, password, atau isi header `Authorization` ke log deployment.

Periksa versi dan database aktif:

```sql
SELECT VERSION() AS database_version, DATABASE() AS active_database;
```

Periksa keberadaan tabel terkait:

```sql
SELECT TABLE_NAME
FROM information_schema.TABLES
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME IN (
    'user_mobile',
    'otp_user_mobile',
    'otp_verif_email_mobile',
    'otp_password_reset_mobile',
    'otp_medical_record_claim_mobile',
    'session_user_mobile',
    'auth_ticket_mobile'
  )
ORDER BY TABLE_NAME;
```

Periksa kemungkinan tabel auth legacy sebelum membuat tabel `_mobile`:

```sql
SELECT TABLE_NAME, TABLE_ROWS
FROM information_schema.TABLES
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME IN (
    'user', 'otp_user', 'otp_verif_email', 'otp_password_reset',
    'user_mobile', 'otp_user_mobile', 'otp_verif_email_mobile',
    'otp_password_reset_mobile'
  )
ORDER BY TABLE_NAME;

SELECT TABLE_NAME, COLUMN_NAME, COLUMN_TYPE
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME IN ('user', 'otp_user', 'otp_verif_email', 'otp_password_reset')
ORDER BY TABLE_NAME, ORDINAL_POSITION;
```

STOP dan minta DBA memverifikasi kepemilikan/fingerprint bila tabel `_mobile` kosong atau belum ada tetapi tabel legacy berisi data. Jangan menganggap tabel bernama `user` sebagai akun mobile karena nama tersebut dapat dimiliki SIMRS inti. Setelah terkonfirmasi, lakukan `INSERT ... SELECT` dengan mapping kolom eksplisit, backup, rekonsiliasi jumlah row, dan uji login di staging. Jangan menggunakan `RENAME TABLE` otomatis.

Periksa storage engine:

```sql
SELECT TABLE_NAME, ENGINE
FROM information_schema.TABLES
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME IN (
    'user_mobile',
    'otp_user_mobile',
    'otp_verif_email_mobile',
    'otp_password_reset_mobile',
    'otp_medical_record_claim_mobile',
    'session_user_mobile',
    'auth_ticket_mobile'
  )
ORDER BY TABLE_NAME;
```

Hentikan deployment bila salah satu tabel tersebut bukan `InnoDB`. Bagian 2026-07-15 di `history.sql` menyediakan `ALTER TABLE ... ENGINE=InnoDB` yang dijalankan hanya untuk tabel terkait. Konversi engine dapat mengunci tabel, sehingga wajib dilakukan pada maintenance window setelah backup.

Periksa kolom OTP sebelum menjalankan `ALTER`:

```sql
SELECT TABLE_NAME, COLUMN_NAME, COLUMN_TYPE, IS_NULLABLE, COLUMN_DEFAULT
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME IN (
    'otp_user_mobile',
    'otp_verif_email_mobile',
    'otp_password_reset_mobile',
    'otp_medical_record_claim_mobile'
  )
  AND COLUMN_NAME IN (
    'otp_hash',
    'attempt_count',
    'last_attempt_at',
    'locked_at',
    'used_at',
    'patient_name'
  )
ORDER BY TABLE_NAME, ORDINAL_POSITION;
```

Jika salah satu tabel sudah memiliki sebagian kolom baru, jalankan hanya pernyataan `ADD COLUMN` yang belum ada. Jangan menjalankan satu blok multi-column secara utuh pada tabel yang sudah termigrasi sebagian.

Periksa duplikasi sebelum menambah unique index. Deployment harus dihentikan dan datanya direkonsiliasi bila query berikut menghasilkan baris:

```sql
SELECT LOWER(TRIM(email)) AS normalized_email, COUNT(*) AS total
FROM user_mobile
GROUP BY LOWER(TRIM(email))
HAVING COUNT(*) > 1;

SELECT patient_id, COUNT(*) AS total
FROM user_mobile
WHERE patient_id IS NOT NULL
GROUP BY patient_id
HAVING COUNT(*) > 1;
```

Periksa nama dan sifat unique index setelah migrasi:

```sql
SELECT TABLE_NAME, INDEX_NAME, NON_UNIQUE, SEQ_IN_INDEX, COLUMN_NAME
FROM information_schema.STATISTICS
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME IN ('user_mobile', 'session_user_mobile', 'auth_ticket_mobile')
  AND INDEX_NAME IN (
    'uk_user_email',
    'uk_user_patient_id',
    'uk_session_user_mobile_access_hash',
    'uk_session_user_mobile_refresh_hash',
    'uk_auth_ticket_mobile_hash'
  )
ORDER BY TABLE_NAME, INDEX_NAME, SEQ_IN_INDEX;
```

Kelima indeks harus ada, `NON_UNIQUE` harus bernilai `0`, dan kolomnya berturut-turut harus `email`, `patient_id`, `access_token_hash`, `refresh_token_hash`, serta `ticket_hash` sesuai nama indeks. Migrator sengaja tidak menjatuhkan indeks bernama benar tetapi berisi kolom salah; hentikan deployment, review data, lalu perbaiki indeks secara eksplisit bila validator menolaknya.

## Backup

Ambil backup konsisten sebelum DDL. Gunakan akun backup dan parameter koneksi yang dikelola operator; jangan menaruh password pada command history.

```bash
mysqldump --single-transaction --routines --triggers --events \
  -h <db-host> -P <db-port> -u <backup-user> -p \
  <db-name> \
  user_mobile \
  otp_user_mobile \
  otp_verif_email_mobile \
  otp_password_reset_mobile \
  otp_medical_record_claim_mobile \
  > rsud-mobile-auth-before-<YYYYMMDD-HHMM>.sql
```

Verifikasi file backup tidak kosong, simpan checksum, dan lakukan uji restore di lingkungan nonproduksi bila prosedur restore belum pernah dibuktikan.

## Urutan deployment

1. Hentikan rollout jika backup atau preflight gagal.
2. Terapkan bagian hardening tanggal 2026-07-15 dari `historyQuery/history.sql` pada database target.
3. Ulangi query `information_schema` di atas dan pastikan dua tabel baru serta seluruh kolom OTP tersedia.
4. Jalankan test/build dari source yang akan dirilis:

   ```bash
   go test ./...
   go vet ./...
   go build -o bin/otista-api ./cmd/api
   ```

5. Deploy binary atau image baru, lalu restart service sesuai prosedur lingkungan tersebut.
6. Periksa startup log tanpa menampilkan data rahasia.
7. Jalankan smoke test pada bagian berikut sebelum membuka traffic penuh.

Atur grace period `SIGTERM` service/container sekurang-kurangnya 10 detik agar request aktif dan antrean email sempat ditutup. Antrean email berada di memori, sehingga crash proses tetap dapat menghilangkan job; monitoring SMTP dan mekanisme permintaan OTP ulang tetap diperlukan.

Binary API tidak menjalankan DDL saat startup. Terapkan `history.sql` lebih dahulu dengan akun migrasi yang berwenang, lalu jalankan API menggunakan akun runtime least-privilege. Migrator eksplisit `go run ./cmd/migrate` dapat dipakai setelah backup dan review DBA; command itu boleh memakai akun DDL khusus, tetapi jangan memakai akun tersebut untuk runtime dan jangan menjadikannya auto-migration setiap restart.

Pada `APP_ENV=production`, startup API bersifat fail-fast:

- konfigurasi database dan SMTP harus lengkap, port valid, dan secret bukan placeholder;
- akun database runtime `root` ditolak;
- tabel/kolom auth, storage engine `InnoDB`, dan lima unique index di atas divalidasi secara read-only melalui `information_schema`;
- bila validasi gagal, proses API berhenti sebelum membuka port. Jalankan migrasi dan perbaiki konfigurasi; jangan menghapus validasi untuk memaksa service hidup.

## Kontrak sesi

Access token dikirim hanya melalui header berikut:

```http
Authorization: Bearer <access_token>
```

Jangan kirim token di query string. `email`, `identifier`, `no_rm`, dan `patient_id` yang mungkin masih ada pada payload/query klien lama tidak menjadi sumber identitas untuk endpoint terproteksi. Handler pasien membentuk scope hanya dari `user_id + patient_id` pada sesi/database; email dan nomor rekam medis tetap ada sebagai atribut bisnis, bukan bukti otorisasi dari klien.

Respons sukses `POST /auth/verify-otp` dan `POST /auth/set-password` berisi identitas pengguna beserta:

```json
{
  "token_type": "Bearer",
  "access_token": "<opaque-access-token>",
  "access_expires_at": "<timestamp>",
  "access_expires_in": 900,
  "refresh_token": "<opaque-refresh-token>",
  "refresh_expires_at": "<timestamp>",
  "refresh_expires_in": 2592000
}
```

Nilai token mentah hanya dikembalikan pada respons tersebut dan tidak dapat diambil kembali dari database.

Refresh sesi:

```http
POST /api/v1/auth/refresh
Content-Type: application/json

{"refresh_token":"<refresh-token-terakhir>"}
```

Respons refresh mengganti **kedua** token. Rotasi tidak memperpanjang batas absolut 30 hari keluarga sesi. Refresh yang dilakukan kurang dari dua detik setelah token diterbitkan menghasilkan `429` dengan `Retry-After: 2`. Klien harus menyimpan pasangan baru secara atomik dan membuang pasangan lama. Flutter harus memakai mekanisme single-flight/mutex: bila beberapa request mendapat `401` bersamaan, hanya satu request yang boleh melakukan refresh; request lain menunggu hasilnya. Refresh paralel dengan token yang sama dapat dianggap reuse dan mencabut seluruh keluarga sesi.

Logout:

```http
POST /api/v1/auth/logout
Authorization: Bearer <access-token>
```

Logout dan deteksi reuse mencabut seluruh keluarga sesi. Reset password mencabut seluruh sesi milik pengguna.

## Kontrak registrasi baru

Alur yang benar:

1. `POST /api/v1/auth/register` mengirim OTP.
2. `POST /api/v1/auth/verify-otp-new-user` memverifikasi OTP dan mengembalikan `registration_ticket` dengan waktu kedaluwarsa.
3. `POST /api/v1/auth/set-password` mengirim password baru dan tiket tersebut. Identitas registrasi berasal dari tiket, bukan email request.

Contoh langkah kedua:

```json
{
  "email": "pasien@example.com",
  "otp": "123456"
}
```

Contoh respons:

```json
{
  "message": "Register otp verified",
  "data": {
    "registration_ticket": "<opaque-one-time-ticket>",
    "expires_at": "<timestamp>"
  }
}
```

Contoh langkah ketiga:

```json
{
  "password": "<password-baru>",
  "registration_ticket": "<ticket-dari-langkah-sebelumnya>"
}
```

Tiket berlaku 10 menit, hanya untuk email/tujuan yang terkait, dan hanya dapat digunakan sekali. Jangan mencoba ulang `set-password` dengan tiket yang sama; mulai kembali dari permintaan/verifikasi OTP bila hasil request tidak dapat dipastikan.

## Matriks route setelah rollout

Semua path di bawah relatif terhadap `/api/v1`.

### Publik

- `GET /health`
- `GET /mobile/hospital/room-availabilities`
- `GET /mobile/hospital/polyclinics`
- `GET /polis` — alias kompatibilitas dengan proyeksi kolom yang dibatasi
- `GET /mobile/booking/options/:poli_id`
- `GET /mobile/booking/calendar`
- `POST /auth/register`
- `POST /auth/login`
- `POST /auth/verify-otp`
- `POST /auth/verify-otp-new-user`
- `POST /auth/set-password`
- `POST /auth/forgot-password`
- `POST /auth/reset-password`
- `POST /auth/refresh`

Endpoint publik auth dilindungi dua lapis rate limit aplikasi: bucket per subject memakai hash identifier/email/token yang dinormalisasi, dan bucket global per proses menahan serangan yang terus mengganti identifier. Operasi bcrypt/OTP juga berbagi batas empat pekerjaan serentak per proses. Token atau PII mentah tidak disimpan sebagai key limiter. Batas OTP pada database berlaku lintas proses API, sedangkan limiter, concurrency cap, dan antrean email berlaku per instance dan bukan pengganti rate limit di edge/load balancer.

### Wajib Bearer

- `GET /auth/me`
- `POST /auth/logout`
- `POST /auth/medical-record/request`
- `POST /auth/medical-record/confirm`
- `GET /mobile/patient/profile`
- `GET /mobile/patient/visits`
- `GET /mobile/patient/medical-summaries`
- `GET /mobile/patient/medical-summaries/:registration_id/pdf`
- `GET /mobile/patient/laboratory-results`
- `GET /mobile/patient/radiology-results`
- `GET /mobile/patient/prescriptions`
- `GET /mobile/booking/general/mine`
- `POST /mobile/booking/general`

Route pasien dan booking juga mensyaratkan akun sudah terhubung ke pasien. Akun valid yang belum terhubung mendapat `403`; token hilang/tidak valid/kedaluwarsa mendapat `401` dan `WWW-Authenticate: Bearer`.

Klaim rekam medis memakai `user_id` dari Bearer session. Body `/auth/medical-record/request` hanya memerlukan `password`, `no_rm`, `nik`, dan `birth_date`; body `/auth/medical-record/confirm` hanya memerlukan `otp`. Field email dari klien diabaikan dan tidak menjadi lookup akun.

### Dinonaktifkan dari API publik

- `/tables`
- `/tables/:table`
- `/tables/:table/search`
- `/:table`
- `/:table/search`
- `/:table/:id`
- `GET /mobile/booking/general` (daftar seluruh pasien)

Jangan membuat alias kompatibilitas untuk route tersebut. Endpoint generik memungkinkan enumerasi tabel/kolom dan data yang tidak semestinya berada pada boundary aplikasi mobile.

## Smoke test

Gunakan akun uji khusus, data pasien uji, dan client yang tidak merekam token. Simpan hanya status code serta hasil sanitasi.

1. Health:

   ```bash
   curl -i https://api-mobile.rsudotista.my.id/api/v1/health
   ```

   Harapan: `200` dan respons minimal. Bila database tidak siap, rilis harus menghasilkan status non-`2xx` sesuai implementasi health terbaru.

2. Endpoint pasien tanpa token:

   ```bash
   curl -i https://api-mobile.rsudotista.my.id/api/v1/mobile/patient/profile
   ```

   Harapan: `401`, bukan data pasien.

3. Endpoint generik yang sudah dipensiunkan:

   ```bash
   curl -i https://api-mobile.rsudotista.my.id/api/v1/tables
   curl -i https://api-mobile.rsudotista.my.id/api/v1/users
   ```

   Harapan: `404`.

4. Selesaikan login dan OTP menggunakan akun uji. Pastikan respons verifikasi menghasilkan pasangan access/refresh token, kemudian panggil `/auth/me` dan endpoint pasien dengan Bearer token.
5. Ubah `email` atau `no_rm` pada query/body menjadi milik pasien uji lain. Harapan: respons tetap mengikuti identitas sesi atau payload ditolak; data pasien lain tidak boleh keluar.
6. Refresh satu kali, lalu pastikan pasangan lama tidak digunakan lagi dan pasangan baru dapat mengakses `/auth/me`.
7. Pada lingkungan nonproduksi, gunakan kembali refresh token lama. Harapan: `401` dan seluruh keluarga sesi tersebut tidak dapat dipakai lagi.
8. Logout menggunakan token terbaru. Harapan: request berikutnya ke `/auth/me` mendapat `401`.
9. Uji OTP salah sampai lima kali menggunakan akun uji. Harapan: verifikasi terkunci; OTP mentah tidak muncul di database atau log.
10. Verifikasi `/polis`, kalender, pilihan dokter/poli, dan ketersediaan kamar tetap bisa dibaca tanpa token serta hanya menampilkan field yang memang diperlukan aplikasi.

Jika Cloudflare mengembalikan challenge atau halaman HTML sebelum request mencapai API, hasil tersebut tidak membuktikan health aplikasi. Operator yang memiliki akses edge perlu mengizinkan health check terkontrol atau menguji origin dari jaringan/server yang berwenang.

## Kesiapan aplikasi Flutter

Jangan mengaktifkan proteksi endpoint pasien/booking untuk traffic pengguna sampai build Flutter yang kompatibel siap. Klien harus:

- menyimpan access/refresh token di secure storage, bukan preference/log biasa;
- menambahkan header Bearer pada seluruh route terproteksi;
- melakukan refresh single-flight dan menyimpan pasangan token baru secara atomik;
- menghapus token saat logout, reset password, atau refresh gagal;
- tidak memakai email/no RM dari input pengguna sebagai otorisasi;
- menangani `401` terpisah dari `403` akun belum terhubung;
- menyelesaikan alur `registration_ticket` sebelum `set-password`.

Viewer PDF eksternal umumnya tidak dapat menambahkan header Bearer. Sebelum endpoint PDF diproteksi di produksi, Flutter harus mengunduh byte PDF melalui HTTP client terautentikasi lalu membuka file lokal, atau backend menyediakan URL sekali pakai yang sangat singkat. Jangan menaruh access token pada URL PDF.

## Rollback

Jika smoke test gagal:

1. Tutup traffic ke binary baru atau kembalikan service ke binary/image lama sesuai prosedur platform.
2. Jangan langsung menjatuhkan `session_user_mobile`, `auth_ticket_mobile`, atau kolom OTP baru. Skema tambahan bersifat kompatibel untuk rollback aplikasi dan menyimpan bukti audit yang berguna.
3. Setelah rollback, minta pengguna meminta OTP baru. Binary lama tidak dapat memverifikasi OTP yang dibuat binary baru karena binary baru menyimpan `otp_code` kosong dan nilai sebenarnya sebagai bcrypt di `otp_hash`.
4. Sesi opaque dari binary baru tidak dikenali oleh alur lama; pengguna mungkin perlu login ulang.
5. Jika rollback dipicu isu data, hentikan flow auth baru dan restore backup hanya melalui prosedur DBA yang tervalidasi. Jangan restore seluruh database secara buta karena dapat menimpa transaksi pasien yang terjadi setelah backup.
6. Catat waktu mulai/selesai, versi binary, DDL yang sudah diterapkan, gejala, dan keputusan rollback tanpa mencatat rahasia.

Menghapus tabel/kolom baru bukan bagian dari rollback cepat. Lakukan hanya setelah investigasi, masa retensi audit, backup baru, dan persetujuan pemilik data.

## Pemeliharaan dan cleanup

Jadwalkan cleanup menggunakan event scheduler/job operator setelah kebijakan retensi disetujui. Contoh berikut memakai retensi 90 hari untuk sesi dan 30 hari untuk tiket/OTP; sesuaikan dengan kebutuhan audit RS sebelum dijalankan.

Lakukan estimasi dahulu:

```sql
SELECT COUNT(*) AS deletable_sessions
FROM session_user_mobile
WHERE (refresh_expires_at < NOW() - INTERVAL 90 DAY)
   OR (revoked_at IS NOT NULL AND updated_at < NOW() - INTERVAL 90 DAY);

SELECT COUNT(*) AS deletable_tickets
FROM auth_ticket_mobile
WHERE created_at < NOW() - INTERVAL 30 DAY
  AND (expires_at < NOW() OR used_at IS NOT NULL OR revoked_at IS NOT NULL);
```

Hapus secara batch kecil agar tidak membuat lock panjang:

```sql
DELETE FROM session_user_mobile
WHERE (refresh_expires_at < NOW() - INTERVAL 90 DAY)
   OR (revoked_at IS NOT NULL AND updated_at < NOW() - INTERVAL 90 DAY)
LIMIT 1000;

DELETE FROM auth_ticket_mobile
WHERE created_at < NOW() - INTERVAL 30 DAY
  AND (expires_at < NOW() OR used_at IS NOT NULL OR revoked_at IS NOT NULL)
LIMIT 1000;
```

Untuk empat tabel OTP, hapus hanya row yang sudah kedaluwarsa dan berumur melewati retensi. Jalankan per tabel, per batch, dan pantau replica lag/lock:

```sql
DELETE FROM otp_user_mobile
WHERE expired_at < NOW() - INTERVAL 30 DAY
LIMIT 1000;
```

Terapkan pola yang sama pada `otp_verif_email_mobile`, `otp_password_reset_mobile`, dan `otp_medical_record_claim_mobile`. Jangan menjalankan cleanup pertama kali langsung di produksi tanpa uji pada salinan database.

## Batas tanggung jawab source Go

Hardening source ini mengurangi CWE-306 pada boundary aplikasi, tetapi keamanan domain publik tetap memerlukan pekerjaan operator server: TLS origin-to-edge, aturan Cloudflare/WAF, pembatasan origin, secret management, rotasi kredensial database/SMTP, log redaction, monitoring, backup, dan rate limiting terdistribusi. Item tersebut tidak diterapkan oleh perubahan source ini dan harus diverifikasi terpisah oleh pihak yang memiliki akses server.
