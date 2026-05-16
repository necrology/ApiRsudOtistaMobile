# ApiRsudOtistaMobile

API RSUD Otista Mobile menggunakan Go Fiber dan MySQL.

## Struktur

```text
cmd/api                 Entry point aplikasi
internal/config         Konfigurasi environment
internal/database       Koneksi MySQL
internal/http/handlers  Handler HTTP
internal/http/response  Format response JSON
internal/http/routes    Registrasi route
internal/repository     Query data MySQL
internal/resources      Definisi resource API
database                Script SQL/import data
```

## Menjalankan

1. Pastikan MySQL XAMPP aktif.
2. Jalankan script database jika belum:

```powershell
& 'C:\xampp\mysql\bin\mysql.exe' --local-infile=1 -u root --execute="source database/mysql_schema_import.sql"
```

3. Salin `.env.example` menjadi `.env`, lalu sesuaikan jika perlu.
4. Jalankan API:

```powershell
go run ./cmd/api
```

Endpoint read-only:

```text
GET /api/v1/health
GET /api/v1/tables
GET /api/v1/tables/:table
GET /api/v1/tables/:table/search
GET /api/v1/:table
GET /api/v1/:table/search
GET /api/v1/:table/:id
```

Contoh:

```text
GET /api/v1/pasiens?q=ahmad&limit=10
GET /api/v1/pasiens?q=ahmad&columns=id,no_rm,nama&with_total=true
GET /api/v1/pegawais?search=dokter&search_columns=nama,jabatan
GET /api/v1/tables/pasiens
GET /api/v1/pasiens/1
```

Parameter pencarian:

```text
q atau search       Kata kunci pencarian.
page               Halaman data, default 1.
limit              Jumlah data per halaman, default 20, maksimal 100.
columns            Kolom yang ingin ditampilkan, dipisah koma.
search_columns     Kolom teks yang dipakai untuk mencari, dipisah koma.
with_total         Isi true jika butuh total_rows dan total_pages.
```

API hanya menyediakan endpoint `GET`. Nama tabel dan kolom divalidasi dari metadata database, query memakai parameter binding, dan pagination mengambil `limit + 1` data untuk mendeteksi halaman berikutnya tanpa selalu menjalankan `COUNT(*)`.

## Docker

Project ini sudah bisa dijalankan sebagai dua service Docker:

- `api`: aplikasi Go Fiber pada port internal `8080`.
- `mysql`: database MySQL dengan volume persistent `otista_mysql_data`.

Untuk menyiapkan dump database dari XAMPP lokal:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\export_mysql_dump.ps1
```

Lalu jalankan:

```powershell
copy .env.docker.example .env
docker compose up -d --build
```

Endpoint container API tersedia di host pada `127.0.0.1:8080`, sehingga bisa diarahkan dari Cloudflare Tunnel ke `http://localhost:8080`.
