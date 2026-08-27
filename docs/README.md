# ApiRsudOtistaMobile

API RSUD Otista Mobile menggunakan Go Fiber dan MySQL.

## Arsitektur

- `cmd/api`: entry point aplikasi Go.
- `internal/config`: konfigurasi environment.
- `internal/database`: koneksi MySQL dan migrasi auth mobile.
- `internal/http/handlers`: handler endpoint.
- `internal/http/routes`: registrasi route.
- `internal/repository`: query data MySQL.
- `historyQuery/history.sql`: catatan alter dan DDL tambahan yang harus diikuti saat perubahan skema.
- [`docs/mobile-auth-security-deployment.md`](mobile-auth-security-deployment.md): runbook migrasi dan deployment hardening autentikasi/session mobile.
- `docs/api-docker-deploy.md`: panduan deploy API dengan Docker, versi stack, env, dan endpoint verifikasi.
- `docs/deploy-api-step-by-step.md`: panduan deploy API ke server dengan bahasa langkah demi langkah untuk operator.
- [`docs/privacy-account-deletion.md`](privacy-account-deletion.md): kontrak kebijakan privasi, penghapusan akun, migrasi, dan smoke test.

## Database Produksi

Aplikasi membaca database SIMRS yang ditentukan melalui environment. Host, port,
nama database, dan kredensial produksi harus diberikan operator; dokumentasi ini
tidak mengasumsikan topologi server tertentu. Gunakan akun aplikasi
least-privilege, bukan akun `root`.

Tabel inti yang dipakai fitur mobile:

- `pasiens`: master pasien.
- `user_mobile`: akun mobile dan relasi ke pasien lewat `patient_id`.
- `polis`: master poli dan jam layanan.
- `pegawais`: master dokter/pegawai.
- `registrasis_dummy`: staging pendaftaran/antrian dari integrasi luar dan mobile.
- `registrasis`: pusat transaksi pendaftaran/kunjungan.
- `antrian_poli`: nomor antrian poli dan status panggil.
- `resume_pasiens`, `hasillabs`, `order_lab`, `order_radiologi`, `radiologi_ekspertises`: data riwayat dan hasil pemeriksaan.

Relasi booking umum mobile:

- `user_mobile.patient_id -> pasiens.id`
- API membuat booking awal ke `registrasis_dummy`.
- `registrasis_dummy.no_rm -> pasiens.no_rm`
- `registrasis_dummy.kode_poli -> polis.bpjs` atau `polis.id`
- `registrasis_dummy.dokter_id -> pegawais.id`
- Pembeda umum mobile: `flag='mobile_umum'`, `jenisrequest=NULL`, `jenisdaftar='android'`
- Proses pendaftaran berikutnya dapat mengisi `registrasis_dummy.registrasi_id -> registrasis.id`.

## Endpoint Baru

Booking umum:

```text
POST /api/v1/mobile/booking/general
GET  /api/v1/mobile/booking/calendar
```

Request body contoh:

```json
{
  "poli_id": 27,
  "tanggal": "2026-06-08",
  "bayar": "2",
  "jenis_pasien": "umum",
  "dokter_id": "42",
  "queue_group": "HDLO",
  "is_jkn": false
}
```

Pada binary yang sudah di-hardening, `identifier`, `email`, dan `no_rm` pada payload booking lama tidak dipakai sebagai sumber identitas. Endpoint membuat booking untuk pasien yang terhubung pada Bearer session.

Respons booking:

- `registration_id`
- `dummy_id`
- `queue_id`
- `registration_code`
- `queue_number`
- `queue_code`
- `queue_group`
- `poli_id`
- `poli_name`
- `queue_date`
- `service_mode`
- `source`

Booking umum masuk ke `registrasis_dummy` terlebih dahulu, bukan langsung ke `registrasis` dan `antrian_poli`. BPJS tidak dibuat sebagai booking internal; aplikasi mobile diarahkan ke Mobile JKN via launcher Android dan fallback Play Store.

Detail endpoint, response lengkap, dan query pendaftaran ada di `docs/mobile-booking-general-api.md`.

## Menjalankan Lokal

1. Pastikan database SIMRS sudah tersedia dan bisa diakses dari `.env`.
2. Salin `.env.example` menjadi `.env` lalu sesuaikan koneksi MySQL, SMTP, dan port.
3. Jalankan catatan ALTER/index yang relevan dari `historyQuery/history.sql` bila belum pernah diterapkan. Untuk database lokal, alternatifnya jalankan `go run ./cmd/migrate`.
4. Jalankan API:

```powershell
go run ./cmd/api
```

Endpoint yang tersedia:

```text
# Publik
GET  /api/v1/health
GET  /api/v1/polis
GET  /api/v1/mobile/hospital/polyclinics
GET  /api/v1/mobile/hospital/room-availabilities
GET  /api/v1/mobile/booking/options/:poli_id
GET  /api/v1/mobile/booking/calendar
POST /api/v1/auth/register
POST /api/v1/auth/login
POST /api/v1/auth/verify-otp
POST /api/v1/auth/verify-otp-new-user
POST /api/v1/auth/set-password
POST /api/v1/auth/forgot-password
POST /api/v1/auth/reset-password
POST /api/v1/auth/refresh
POST /api/v1/auth/account-deletion/web/request
POST /api/v1/auth/account-deletion/web/confirm
GET  /privacy-policy
GET  /account-deletion

# Wajib Authorization: Bearer <access_token>
GET  /api/v1/auth/me
POST /api/v1/auth/logout
POST /api/v1/auth/medical-record/request
POST /api/v1/auth/medical-record/confirm
POST /api/v1/auth/account-deletion/request
POST /api/v1/auth/account-deletion/confirm
POST /api/v1/mobile/booking/general
GET  /api/v1/mobile/booking/general/mine
GET  /api/v1/mobile/patient/profile
GET  /api/v1/mobile/patient/visits
GET  /api/v1/mobile/patient/medical-summaries
GET  /api/v1/mobile/patient/medical-summaries/:registration_id/pdf
GET  /api/v1/mobile/patient/laboratory-results
GET  /api/v1/mobile/patient/radiology-results
GET  /api/v1/mobile/patient/prescriptions
```

Route tabel generik (`/tables`, `/:table`, search, dan detail) serta `GET /mobile/booking/general` dinonaktifkan dari API publik. Kontrak lengkap dan langkah rollout ada di [runbook hardening auth](mobile-auth-security-deployment.md).

## Deploy Produksi

Untuk binary dengan Bearer session, OTP bcrypt, `registration_ticket`, dan penghapusan akun, operator wajib mengikuti [runbook hardening auth](mobile-auth-security-deployment.md) serta [runbook penghapusan akun](privacy-account-deletion.md). Urutan amannya adalah backup, terapkan bagian DDL 2026-07-15 dan 2026-08-27 dari `historyQuery/history.sql`, deploy binary, lalu jalankan smoke test. API produksi menolak akun runtime `root`, konfigurasi DB/SMTP yang belum lengkap, dan skema auth parsial saat startup. Perubahan source ini tidak mengubah Cloudflare atau konfigurasi server.

### 1. Siapkan environment

Buat file `.env` pada server produksi dengan isi yang sesuai:

```env
APP_NAME=ApiRsudOtistaMobile
APP_ENV=production
APP_PORT=8080
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=otista_app
DB_PASSWORD=<secret-dari-secret-manager>
DB_NAME=rsud_otista
SMTP_HOST=...
SMTP_PORT=587
SMTP_EMAIL=...
SMTP_PASSWORD=...
HOLIDAY_API_BASE_URL=https://tanggalmerah.upset.dev
```

Jika MySQL berada di container internal, pastikan host/port yang dipakai bisa dijangkau dari proses API tanpa membuka port database ke internet.

### 2. Sinkronkan database

Database utama SIMRS sudah ada, jadi jangan menjalankan import schema penuh atau script yang berisi `DROP TABLE`.

Jalankan catatan ALTER/index dari `historyQuery/history.sql` secara selektif dan terkontrol untuk memastikan skema auth mobile, hasil pasien, dan booking umum sesuai dengan code.

### 3. Build dan run

```powershell
go test ./...
go vet ./...
go build -o bin/otista-api ./cmd/api
./bin/otista-api
```

Atau jalankan sebagai service/container bila server produksi memang memakai Docker.

### 4. Verifikasi

- `GET /api/v1/health`
- Pastikan `GET /api/v1/tables` menghasilkan `404`.
- Pastikan endpoint pasien tanpa Bearer token menghasilkan `401`.
- Selesaikan login/OTP dan verifikasi `GET /api/v1/auth/me` memakai access token.
- Verifikasi refresh rotation dan logout sesuai runbook hardening auth.
- `GET /api/v1/mobile/booking/calendar`
- `POST /api/v1/mobile/booking/general` memakai Bearer session akun pasien uji.
- `GET /api/v1/mobile/patient/visits` memakai Bearer session akun pasien uji.

### 5. Catatan produksi

- Booking umum mobile masuk ke `registrasis_dummy` dengan flag khusus `mobile_umum`.
- Hari libur nasional/cuti bersama disinkronkan dari Tanggal Merah API dan disimpan ke `tanggal_libur_rs`.
- BPJS tetap diarahkan ke Mobile JKN, bukan dibuat booking internal ganda.
- Setiap perubahan skema harus dicatat di `historyQuery/history.sql` supaya deployment prod tidak kehilangan jejak.

## Docker

Project ini dijalankan sebagai service Docker `api`. Database MySQL/MariaDB menggunakan database SIMRS yang sudah ada dan dikoneksikan lewat `.env`.

Lalu jalankan:

```powershell
copy .env.docker.example .env
docker compose up -d --build
```

Endpoint container API tersedia di host pada `127.0.0.1:8080`, sehingga bisa diarahkan dari Cloudflare Tunnel ke `http://localhost:8080`.
