# Kebijakan Privasi dan Penghapusan Akun SIPANTES

Dokumen ini menjelaskan kontrak fitur, migrasi database, pengujian, dan urutan deployment yang aman.

## URL Publik

Setelah API terbaru aktif pada domain produksi:

```text
https://api-mobile.rsudotista.my.id/privacy-policy
https://api-mobile.rsudotista.my.id/account-deletion
```

Alias berbahasa Indonesia juga tersedia:

```text
https://api-mobile.rsudotista.my.id/kebijakan-privasi
https://api-mobile.rsudotista.my.id/hapus-akun
```

## Kontrak Penghapusan

Alur aplikasi dan halaman web menggunakan password lalu OTP email enam digit:

```text
# Dari aplikasi, wajib Bearer session
POST /api/v1/auth/account-deletion/request
POST /api/v1/auth/account-deletion/confirm

# Dari halaman web publik
POST /api/v1/auth/account-deletion/web/request
POST /api/v1/auth/account-deletion/web/confirm
```

Penghapusan berhasil melakukan tindakan berikut dalam satu transaksi:

1. Menghapus sesi login dan challenge OTP akun mobile.
2. Menghapus tiket autentikasi dan verifikasi email yang terkait.
3. Melepas hubungan `patient_id` dan `no_rm` dari akun mobile.
4. Menghapus nama, email, telepon, password, dan data identitas akun dengan anonimisasi record audit.
5. Menandai akun sebagai dihapus agar tidak dapat digunakan kembali.

Penghapusan akun tidak menghapus tabel rekam medis, hasil pemeriksaan, resep, registrasi, maupun antrean rumah sakit. Data tersebut dikelola sebagai dokumen pelayanan kesehatan terpisah.

## Migrasi Database

Binary baru memerlukan tabel berikut:

```text
otp_account_deletion_mobile
```

DDL tersedia pada bagian `2026-08-27` di `historyQuery/history.sql`. Lakukan backup sebelum migrasi. Terapkan migrasi dengan akun DDL khusus, bukan akun runtime API.

Alternatif untuk lingkungan terkontrol:

```powershell
go run ./cmd/migrate
```

Pastikan akun runtime memiliki hak `SELECT`, `INSERT`, `UPDATE`, dan `DELETE` pada tabel auth mobile yang digunakan fitur ini.

## Validasi Lokal

```powershell
go test ./...
go vet ./...
go build ./cmd/api
```

## Urutan Deployment

1. Identifikasi container atau service yang benar-benar menjadi upstream domain produksi.
2. Backup `user_mobile`, seluruh `otp_*_mobile`, `session_user_mobile`, dan `auth_ticket_mobile`.
3. Terapkan DDL `2026-08-27` menggunakan akun migrasi.
4. Jalankan validasi skema.
5. Deploy binary API terbaru tanpa menimpa `.env` produksi.
6. Pastikan health, halaman legal, dan asset halaman dapat dibuka.
7. Uji alur penghapusan hanya dengan akun sintetis khusus pengujian.

Smoke test setelah deployment:

```bash
curl -fsS https://api-mobile.rsudotista.my.id/api/v1/health
curl -fsSI https://api-mobile.rsudotista.my.id/privacy-policy
curl -fsSI https://api-mobile.rsudotista.my.id/account-deletion
curl -fsSI https://api-mobile.rsudotista.my.id/legal/assets/legal.css
curl -fsSI https://api-mobile.rsudotista.my.id/legal/assets/account-deletion.js
```

## Temuan Server 27 Agustus 2026

Pemeriksaan read-only ke `192.168.1.17` menunjukkan:

- Server dapat diakses melalui SSH.
- Domain publik `api-mobile.rsudotista.my.id` mengembalikan health `200`.
- `docker compose ps` pada `/home/servermaul/otista-api` hanya menampilkan service database.
- Tidak ada listener API pada `127.0.0.1:8080` saat pemeriksaan.
- Konfigurasi Cloudflare Tunnel dimiliki pengguna `root` dan tidak dapat dibaca oleh akun deployment.

Karena upstream domain publik belum dapat dipastikan dari akun tersebut, jangan menjalankan deployment produksi sampai operator dengan akses konfigurasi tunnel mengonfirmasi target service yang benar.
