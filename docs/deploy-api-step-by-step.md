# Panduan Deploy API Go ke Server

Panduan ini dibuat untuk deploy API Go `ApiRsudOtistaMobile` ke server Debian yang memakai Docker.

Bahasa dibuat sederhana supaya bisa diikuti oleh operator/non-programmer.

## Tujuan

Setelah mengikuti panduan ini, API akan berjalan di server dan bisa dicek lewat:

```text
https://otista.maulana-gandawijaya.my.id/api/v1/health
```

Jika hasilnya `status: ok`, berarti API hidup dan berhasil melakukan ping database. Jika database gagal, endpoint mengembalikan `503`.

## Hal Penting Sebelum Mulai

1. Database utama rumah sakit sudah ada.
2. Jangan membuat ulang database.
3. Jangan menjalankan script yang berisi `DROP TABLE`.
4. Jangan import schema dummy/local ke database production.
5. Perubahan database hanya dari `historyQuery/history.sql`, dan dijalankan selektif sesuai kebutuhan.
6. Jangan menulis password asli di chat, README, atau file dokumentasi.

## Yang Dibutuhkan

Di komputer lokal:

1. Source API Go ada di:

```text
C:\Source Golang\ApiRsudOtistaMobile
```

2. PuTTY/PSCP sudah terinstall:

```text
C:\Program Files\PuTTY\plink.exe
C:\Program Files\PuTTY\pscp.exe
```

Di server:

1. Server bisa diakses lewat SSH.
2. Docker dan Docker Compose sudah terinstall.
3. Folder API ada di:

```text
/home/servermaul/otista-api
```

## Gambaran Singkat Proses

Alurnya seperti ini:

1. Buat paket source API dari komputer lokal.
2. Upload paket itu ke server.
3. Extract paket di server.
4. Pastikan file `.env` server benar.
5. Jalankan Docker Compose untuk build dan restart API.
6. Cek health API.

## Langkah 1: Buka PowerShell di Komputer Lokal

Buka PowerShell, lalu masuk ke folder source API:

```powershell
cd "C:\Source Golang\ApiRsudOtistaMobile"
```

Pastikan foldernya benar:

```powershell
dir
```

Harus terlihat file seperti:

```text
Dockerfile
docker-compose.yml
go.mod
cmd
internal
historyQuery
```

## Langkah 2: Test API di Lokal Sebelum Deploy

Jalankan test Go:

```powershell
go test ./...
```

Jika berhasil, biasanya keluar seperti:

```text
?    apirusdotistamobile/cmd/api    [no test files]
...
```

Jika ada error, jangan deploy dulu. Perbaiki errornya.

## Langkah 3: Buat Paket Source API

Jalankan perintah ini dari PowerShell:

```powershell
$stamp = Get-Date -Format 'yyyyMMdd-HHmmss'
$archiveDir = '.deploy'
New-Item -ItemType Directory -Force -Path $archiveDir | Out-Null
$archive = Join-Path $archiveDir "otista-api-update-$stamp.tar.gz"
tar -czf $archive --exclude=.git --exclude=.env --exclude=api.exe cmd internal historyQuery database scripts docs go.mod go.sum Dockerfile docker-compose.yml .dockerignore
Write-Output $archive
```

Catat nama file hasilnya, contoh:

```text
.deploy\otista-api-update-20260611-103000.tar.gz
```

Catatan:

- `.env` tidak ikut dikirim agar password server tidak tertimpa.
- `.git` tidak ikut dikirim agar paket lebih kecil.
- `api.exe` lokal Windows tidak ikut dikirim.

## Langkah 4: Upload Paket ke Server

Ganti nama file archive sesuai hasil Langkah 3.

```powershell
$archive = "C:\Source Golang\ApiRsudOtistaMobile\.deploy\NAMA_FILE_ARCHIVE.tar.gz"
```

Upload ke server:

```powershell
& "C:\Program Files\PuTTY\pscp.exe" -batch $archive "servermaul@192.168.1.17:/home/servermaul/otista-api/.deploy/"
```

Jika diminta password, masukkan password server.

Jika sukses, akan terlihat proses upload sampai `100%`.

## Langkah 5: Masuk ke Server

Dari PowerShell:

```powershell
& "C:\Program Files\PuTTY\plink.exe" -ssh servermaul@192.168.1.17
```

Masukkan password server jika diminta.

Setelah berhasil, masuk ke folder API:

```bash
cd /home/servermaul/otista-api
```

## Langkah 6: Extract Paket Source

Lihat isi folder deploy:

```bash
ls -lah .deploy
```

Extract archive yang baru diupload:

```bash
tar -xzf .deploy/NAMA_FILE_ARCHIVE.tar.gz
```

Contoh:

```bash
tar -xzf .deploy/otista-api-update-20260611-103000.tar.gz
```

## Langkah 7: Cek File Environment `.env`

Pastikan file `.env` ada:

```bash
ls -lah .env
```

Lihat isi tanpa membagikan password ke orang lain:

```bash
nano .env
```

Minimal isi `.env` harus seperti ini:

```env
APP_NAME=ApiRsudOtistaMobile
APP_ENV=production
APP_PORT=8080

DB_HOST=127.0.0.1
DB_PORT=33061
DB_USER=otista_app
DB_PASSWORD=<secret-kuat-dari-secret-manager>
DB_NAME=nama_database

SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_EMAIL=email_pengirim@gmail.com
SMTP_PASSWORD=<app-password-smtp>

HOLIDAY_API_BASE_URL=https://tanggalmerah.upset.dev
```

Keterangan:

| Parameter | Isi dengan |
| --- | --- |
| `APP_ENV` | `production` untuk server |
| `APP_PORT` | `8080` |
| `DB_HOST` | alamat host database |
| `DB_PORT` | port database |
| `DB_USER` | user database |
| `DB_PASSWORD` | password database |
| `DB_NAME` | nama database utama |
| `SMTP_HOST` | `smtp.gmail.com` jika memakai Gmail |
| `SMTP_PORT` | `587` |
| `SMTP_EMAIL` | email pengirim OTP |
| `SMTP_PASSWORD` | Google App Password |
| `HOLIDAY_API_BASE_URL` | `https://tanggalmerah.upset.dev`, dipakai untuk sinkron libur nasional dan cuti bersama |

Gunakan akun database runtime least-privilege dan jangan memakai `root`. Pada mode production, API sengaja gagal startup jika password masih placeholder, konfigurasi DB/SMTP belum lengkap, atau skema auth belum siap.

Simpan file di nano:

```text
CTRL + O
ENTER
CTRL + X
```

## Langkah 7A: Backup dan Terapkan Migrasi Auth

Sebelum build/restart, backup database lalu ikuti runbook `docs/mobile-auth-security-deployment.md`. Terapkan bagian `2026-07-15` dari `historyQuery/history.sql` memakai akun migrasi/DDL, kemudian periksa tabel, kolom, engine `InnoDB`, dan unique index. Jangan memakai akun DDL tersebut sebagai `DB_USER` runtime.

API tidak menjalankan `ALTER TABLE` otomatis. Jika langkah ini dilewati, binary baru akan berhenti sebelum membuka port agar deployment parsial tidak melayani traffic.

## Langkah 8: Build dan Restart API dengan Docker

Karena Docker biasanya butuh akses root, jalankan:

```bash
su
```

Masukkan password root/server jika diminta.

Pastikan masih di folder API:

```bash
cd /home/servermaul/otista-api
```

Build dan restart API:

```bash
docker compose up -d --build api
```

Tunggu sampai proses selesai.

Jika sukses, akan terlihat container `otista-api` dibuat ulang dan started.

## Langkah 9: Cek Status Container

Jalankan:

```bash
docker compose ps api
```

Hasil yang bagus kira-kira:

```text
NAME         IMAGE               SERVICE   STATUS
otista-api   otista-api:latest   api       Up ... (healthy)
```

Jika status masih `starting`, tunggu 10 sampai 30 detik lalu cek lagi.

Jika status `unhealthy` atau restart terus, lihat log:

```bash
docker compose logs -f --tail=100 api
```

## Langkah 10: Cek API dari Server

Masih di server, jalankan:

```bash
curl http://127.0.0.1:8080/api/v1/health
```

Hasil sukses:

```json
{
  "success": true,
  "data": {
    "status": "ok"
  }
}
```

Jika health menghasilkan `503`, periksa `.env`, konektivitas, dan hak akses database.

## Langkah 11: Cek API dari Browser

Buka browser:

```text
https://otista.maulana-gandawijaya.my.id/api/v1/health
```

Jika muncul `success: true`, deploy sudah berhasil.

## Langkah 12: Cek Endpoint Penting

Cek opsi booking poli:

```bash
curl http://127.0.0.1:8080/api/v1/mobile/booking/options/1
```

Cek kalender libur booking:

```bash
curl "http://127.0.0.1:8080/api/v1/mobile/booking/calendar?year=2026&month=6&poli_id=1"
```

Cek history nomor antrian pasien dengan access token akun uji:

```bash
curl -H "Authorization: Bearer REPLACE_ME" \ # gitleaks:allow
  "http://127.0.0.1:8080/api/v1/mobile/booking/general/mine?all_dates=1&limit=5"
```

`GET /api/v1/mobile/booking/general` sengaja dinonaktifkan karena menampilkan booking lintas pasien.

## Cara Rollback Jika Deploy Bermasalah

Jika setelah deploy API error, cara paling sederhana:

1. Upload kembali archive deploy sebelumnya.
2. Extract archive lama.
3. Jalankan ulang:

```bash
docker compose up -d --build api
```

4. Cek health:

```bash
curl http://127.0.0.1:8080/api/v1/health
```

## Yang Tidak Boleh Dilakukan

Jangan menjalankan perintah yang:

1. Menghapus tabel.
2. Membuat ulang database.
3. Import schema dummy/local.
4. Menimpa file `.env` server dengan `.env` lokal.
5. Membagikan password database atau SMTP di chat/dokumen.

Khusus database:

- Gunakan database utama yang sudah ada.
- Perubahan struktur hanya dari `historyQuery/history.sql`.
- Jalankan query ALTER/index secara selektif, bukan semuanya sembarangan.

## Masalah Umum

### API tidak bisa konek database

Cek `.env`:

```bash
nano .env
```

Pastikan:

- `DB_HOST` benar.
- `DB_PORT` benar.
- `DB_USER` benar.
- `DB_PASSWORD` benar.
- `DB_NAME` benar.

### Email OTP tidak masuk

Cek:

- `SMTP_HOST=smtp.gmail.com`
- `SMTP_PORT=587`
- `SMTP_EMAIL` benar
- `SMTP_PASSWORD` memakai Google App Password

### Container tidak healthy

Cek log:

```bash
docker compose logs -f --tail=100 api
```

### Domain tidak bisa dibuka tapi localhost server bisa

Jika ini berhasil:

```bash
curl http://127.0.0.1:8080/api/v1/health
```

Tapi domain gagal, berarti masalah kemungkinan di reverse proxy, tunnel, DNS, atau SSL, bukan di aplikasi Go.

## Checklist Selesai Deploy

Centang satu per satu:

```text
[ ] Source terbaru sudah diupload ke server
[ ] Archive sudah diextract
[ ] File .env sudah benar
[ ] docker compose up -d --build api sukses
[ ] docker compose ps api status healthy
[ ] /api/v1/health dari server sukses
[ ] /api/v1/health dari domain sukses
[ ] Endpoint booking/options bisa dibuka
[ ] Tidak ada script import/drop database yang dijalankan
```
