# Deploy API RSUD Otista Mobile dengan Docker

Dokumen ini menjelaskan cara menjalankan API `ApiRsudOtistaMobile` memakai Docker Compose, parameter `.env` yang harus diisi, dan cara verifikasi setelah deploy.

Jika butuh versi yang lebih sederhana untuk operator/non-programmer, baca `docs/deploy-api-step-by-step.md`.

## Ringkasan Stack

| Komponen | Versi / Paket |
| --- | --- |
| Bahasa | Go 1.26.2 |
| HTTP framework | Fiber v2.52.13 |
| Database driver | go-sql-driver/mysql v1.10.0 |
| Env loader | godotenv v1.5.1 |
| Auth crypto | golang.org/x/crypto v0.51.0 |
| Email | gomail v2 via SMTP |
| Docker build image | golang:1.26-alpine |
| Docker runtime image | alpine:3.22 |
| Database | MySQL/MariaDB compatible, database SIMRS sudah tersedia |

## File Penting

| File | Fungsi |
| --- | --- |
| `Dockerfile` | Build binary Go statis dan runtime Alpine |
| `docker-compose.yml` | Menjalankan service `api` |
| `.env.docker.example` | Template environment untuk Docker |
| `.env` | Environment real server, jangan commit ke Git |
| `historyQuery/history.sql` | Catatan ALTER/index/query pendukung |
| `cmd/api/main.go` | Entry point aplikasi |

## Parameter Environment

Isi file `.env` di root source API. Nilai di bawah adalah contoh format, bukan secret produksi.

```env
DB_HOST=192.168.1.x
DB_PORT=3306
DB_USER=otista_app
DB_PASSWORD=<secret-kuat-dari-secret-manager>
DB_NAME=apirusdotistamobile

SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_EMAIL=nama-akun@gmail.com
SMTP_PASSWORD=<app-password-smtp>

HOLIDAY_API_BASE_URL=https://tanggalmerah.upset.dev
```

Keterangan parameter:

| Parameter | Wajib | Keterangan |
| --- | --- | --- |
| `DB_HOST` | Ya | Host database yang bisa dijangkau container API. Jika database di server yang sama, gunakan IP server atau hostname yang dikenali container. |
| `DB_PORT` | Ya | Port MySQL/MariaDB, biasanya `3306`. |
| `DB_USER` | Ya | User runtime least-privilege untuk API, bukan `root`. Minimal punya akses baca/tulis tabel yang dipakai aplikasi serta membaca metadata skema sendiri. |
| `DB_PASSWORD` | Ya | Password user database. |
| `DB_NAME` | Ya | Nama database SIMRS/API. |
| `SMTP_HOST` | Ya untuk OTP/email | Host SMTP, contoh Gmail: `smtp.gmail.com`. |
| `SMTP_PORT` | Ya untuk OTP/email | Port SMTP, Gmail TLS biasanya `587`. |
| `SMTP_EMAIL` | Ya untuk OTP/email | Email pengirim OTP/verifikasi. |
| `SMTP_PASSWORD` | Ya untuk OTP/email | App password SMTP. Untuk Gmail gunakan Google App Password, bukan password login utama. |
| `HOLIDAY_API_BASE_URL` | Opsional | Base URL Tanggal Merah API untuk kalender libur booking, default `https://tanggalmerah.upset.dev`. |
| `APP_NAME` | Opsional | Nama aplikasi, default `ApiRsudOtistaMobile`. |
| `APP_ENV` | Opsional | `production` untuk server. |
| `APP_PORT` | Opsional | Port internal aplikasi. Compose saat ini expose `127.0.0.1:8080->8080`. |
| `TZ` | Opsional | Timezone container, default di compose: `Asia/Jakarta`. |

## Cara Deploy

1. Masuk ke folder source API di server.

```bash
cd /home/servermaul/otista-api
```

2. Buat atau update file `.env`.

```bash
cp .env.docker.example .env
nano .env
```

3. Backup database, lalu terapkan bagian `2026-07-15` di `historyQuery/history.sql` memakai akun migrasi/DDL. Alternatif terkontrol adalah `go run ./cmd/migrate`. Ikuti preflight dan validasi lengkap pada `docs/mobile-auth-security-deployment.md`.

4. Build dan jalankan container.

```bash
docker compose up -d --build api
```

5. Cek status container.

```bash
docker compose ps api
```

Status yang diharapkan:

```text
otista-api   Up ... (healthy)   127.0.0.1:8080->8080/tcp
```

6. Cek log bila ada error.

```bash
docker compose logs -f --tail=100 api
```

## Verifikasi Setelah Deploy

Health check:

```bash
curl http://127.0.0.1:8080/api/v1/health
```

Contoh response:

```json
{
  "success": true,
  "data": {
    "status": "ok"
  }
}
```

Jika sudah dipasang reverse proxy/domain:

```bash
curl https://domain-rsud.example/api/v1/health
```

## Endpoint Utama Mobile

| Endpoint | Fungsi |
| --- | --- |
| `GET /api/v1/health` | Cek API dan database |
| `POST /api/v1/auth/register` | Registrasi akun mobile |
| `POST /api/v1/auth/login` | Login akun mobile |
| `POST /api/v1/auth/medical-record/request` | Request koneksi No. RM; wajib Bearer |
| `POST /api/v1/auth/medical-record/confirm` | Konfirmasi OTP No. RM; wajib Bearer |
| `GET /api/v1/mobile/patient/profile` | Profil pasien dari identitas Bearer session |
| `GET /api/v1/mobile/patient/visits` | Riwayat kunjungan |
| `GET /api/v1/mobile/patient/laboratory-results` | Hasil lab |
| `GET /api/v1/mobile/patient/radiology-results` | Hasil radiologi |
| `GET /api/v1/mobile/patient/prescriptions` | Resep/obat |
| `GET /api/v1/mobile/booking/options/{poli_id}` | Opsi poli, kuota, jadwal, dokter |
| `GET /api/v1/mobile/booking/calendar` | Kalender booking; menandai libur nasional, cuti bersama, dan Minggu tutup |
| `POST /api/v1/mobile/booking/general` | Buat booking antrian umum ke `registrasis_dummy` |
| `GET /api/v1/mobile/booking/general/mine` | History nomor antrian pasien |

Dokumentasi detail body, response, status, dan query pendaftaran ada di `docs/mobile-booking-general-api.md`.

Header wajib endpoint pasien dan booking milik pasien:

```text
Authorization: Bearer <access_token>
```

Parameter history nomor antrian:

```text
all_dates=1
limit=30
```

Body pendaftaran umum:

```json
{
  "poli_id": 1,
  "tanggal": "2026-06-11",
  "bayar": "2",
  "jenis_pasien": "umum",
  "dokter_id": "31",
  "queue_group": "RDO",
  "is_jkn": false
}
```

Catatan:

- Identitas pasien selalu diambil dari Bearer session; `identifier`, `email`, dan `no_rm` dari klien tidak diterima sebagai otorisasi.
- `pasiens.id` yang dipakai untuk registrasi diambil dari `user_mobile.patient_id`.
- Data booking umum mobile masuk ke `registrasis_dummy` terlebih dahulu dengan `flag='mobile_umum'`, `jenisrequest=NULL`, dan `jenisdaftar='android'`.
- Format nomor mengikuti pendaftaran umum Android existing: `kodebooking=NULL`, `nomorantrian=DDMMYYYY+kode_poli+angkaantrian`, `angkaantrian` angka murni.
- Satu pasien hanya boleh membuat satu pendaftaran umum mobile per tanggal.
- BPJS/JKN tetap diarahkan melalui Mobile JKN, bukan endpoint pendaftaran umum ini.

## Catatan Database

Database utama rumah sakit sudah tersedia. Deploy API tidak membutuhkan import schema penuh.

Jangan menjalankan script import yang membuat ulang tabel atau berisi `DROP TABLE` di database produksi. File import schema dummy/local tidak dipakai untuk deployment production.

Perubahan skema atau index pendukung disimpan di:

```text
historyQuery/history.sql
```

Jalankan ALTER/index dari file tersebut secara terkontrol sebelum binary baru dibuka untuk traffic, terutama di database produksi yang besar. Untuk tabel besar, jalankan saat traffic rendah atau gunakan prosedur maintenance database rumah sakit. Binary API tidak menjalankan DDL dan akan menolak startup bila tabel, kolom, engine, atau unique index auth belum sesuai.

## Troubleshooting Singkat

| Gejala | Cek |
| --- | --- |
| Health menghasilkan `503` | Pastikan `DB_HOST`, `DB_PORT`, user/password, dan firewall database benar |
| OTP/email gagal terkirim | Pastikan `SMTP_HOST`, `SMTP_PORT`, `SMTP_EMAIL`, `SMTP_PASSWORD`; untuk Gmail gunakan App Password |
| Container restart terus | Jalankan `docker compose logs -f --tail=100 api` |
| Endpoint pasien 404 akun | Pastikan `user_mobile.email`, `user_mobile.no_rm`, dan `user_mobile.patient_id` sudah terhubung |
| Endpoint pasien menghasilkan `401` | Pastikan client mengirim access token aktif pada header Bearer |
