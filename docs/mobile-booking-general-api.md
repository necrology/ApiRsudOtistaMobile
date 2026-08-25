# API Booking Antrian Umum Mobile

Dokumen ini menjelaskan endpoint booking antrian umum dari aplikasi Android/Flutter dan alur data ke database SIMRS.

## Ringkasan Alur

Booking umum mobile tidak langsung membuat data ke `registrasis` dan `antrian_poli`.

Alur yang benar:

1. Aplikasi mobile memanggil `POST /api/v1/mobile/booking/general`.
2. API memvalidasi akun `user_mobile` sudah terhubung ke `pasiens` lewat `patient_id`.
3. API memvalidasi tanggal booking terhadap hari libur nasional, cuti bersama, dan hari Minggu.
4. API membuat record ke `registrasis_dummy`.
5. Record dummy diberi pembeda:
   - `flag = 'mobile_umum'`
   - `jenisrequest = NULL`
   - `jenisdaftar = 'android'`
   - `jenis_registrasi = 'antrian'`
   - `status = 'pending'`
5. Bagian pendaftaran/SIMRS mengambil data dari `registrasis_dummy`, lalu proses internal berikutnya membuat `registrasis` dan `antrian_poli`.

Flag ini sengaja berbeda dari Mobile JKN. Query BPJS lama yang memakai `jenisrequest IN (1,2)` atau `request LIKE '%mobile_jkn%'` tidak akan mengambil booking umum mobile.

Format nomor mengikuti data Android umum existing:

- `kodebooking` dibiarkan `NULL`
- `nomorantrian` berisi `DDMMYYYY + kode_poli + angkaantrian`, contoh `21062026INT18`
- `angkaantrian` berisi angka murni, contoh `18`

## Endpoint

Base path:

```text
/api/v1
```

### 1. Opsi Poli, Jadwal, Kuota, dan Dokter

```http
GET /api/v1/mobile/booking/options/{poli_id}
```

Contoh:

```bash
curl "http://127.0.0.1:8080/api/v1/mobile/booking/options/27"
```

Response:

```json
{
  "success": true,
  "data": {
    "poli": {
      "id": 27,
      "nama": "Klinik Penyakit Dalam",
      "kelompok": "PENYAKIT DALAM",
      "politype": "RAWAT JALAN",
      "bpjs": "INT",
      "kode_ruangan": "RJ01",
      "buka": "08:00",
      "tutup": "12:00",
      "praktik": "Senin-Sabtu",
      "kuota": 40,
      "kuota_online": 10,
      "terisi": 12,
      "queue_group": "PENYAKIT DALAM",
      "queue_group_hint": "INT",
      "dokter_list": [
        {
          "id": 42,
          "nama": "dr. Contoh",
          "kode_antrian": "D042",
          "kode_bpjs": "12345",
          "general_code": "D042"
        }
      ]
    }
  }
}
```

Data yang dipakai aplikasi:

| Field | Keterangan |
| --- | --- |
| `poli.id` | Dikirim kembali sebagai `poli_id` saat booking |
| `poli.kuota`, `kuota_online`, `terisi` | Ditampilkan untuk informasi kuota |
| `poli.buka`, `tutup`, `praktik` | Ditampilkan untuk jadwal |
| `dokter_list[].id` | Dikirim sebagai `dokter_id` |
| `queue_group` | Boleh dikirim sebagai `queue_group`; jika kosong API memakai kelompok poli |

### 2. Kalender Booking dan Hari Libur

```http
GET /api/v1/mobile/booking/calendar?year=2026&month=6&poli_id=27
```

Endpoint ini dipakai aplikasi untuk menandai tanggal yang tutup pada kalender pilihan tanggal registrasi.

Sumber data utama hari libur nasional dan cuti bersama adalah Tanggal Merah API / upset.dev. API menyimpan hasil sinkronisasi ke tabel lokal `tanggal_libur_rs` agar tetap bisa memakai cache saat API publik sedang lambat atau tidak dapat dijangkau.

Jika Tanggal Merah API tidak memiliki data untuk tahun tertentu, kalender tetap ditampilkan tanpa penanda libur/cuti untuk tanggal tersebut. Hari Minggu tetap ditandai tutup oleh sistem.

Query parameter:

| Parameter | Wajib | Keterangan |
| --- | --- | --- |
| `year` | Opsional | Tahun kalender. Default tahun berjalan |
| `month` | Opsional | Bulan kalender 1-12. Default bulan berjalan |
| `poli_id` | Opsional | Diterima untuk kompatibilitas aplikasi, tetapi hari Minggu tetap ditutup |

Response:

```json
{
  "success": true,
  "data": {
    "year": 2026,
    "month": 6,
    "days": [
      {
        "date": "2026-06-01",
        "day_name": "Senin",
        "is_sunday": false,
        "has_sunday_schedule": false,
        "is_holiday": false,
        "is_leave": false,
        "is_off_day": false,
        "is_open": true,
        "reason": "",
        "holiday_names": []
      },
      {
        "date": "2026-06-07",
        "day_name": "Minggu",
        "is_sunday": true,
        "has_sunday_schedule": false,
        "is_holiday": false,
        "is_leave": false,
        "is_off_day": true,
        "is_open": false,
        "reason": "Hari Minggu",
        "holiday_names": []
      }
    ]
  }
}
```

Makna field penting:

| Field | Keterangan |
| --- | --- |
| `is_holiday` | Hari libur nasional dari Tanggal Merah API atau cache lokal |
| `is_leave` | Cuti bersama dari Tanggal Merah API atau cache lokal |
| `is_sunday` | Tanggal jatuh pada hari Minggu |
| `has_sunday_schedule` | Field kompatibilitas; saat ini selalu `false` karena verifikasi jadwal dokter/poli tidak dipakai |
| `is_open` | `true` jika tanggal boleh dipakai booking |
| `reason` | Alasan tanggal tertutup, misalnya nama libur/cuti atau `Hari Minggu` |

### 3. Buat Booking Antrian Umum

```http
POST /api/v1/mobile/booking/general
Content-Type: application/json
Authorization: Bearer REPLACE_ME # gitleaks:allow
```

Body:

```json
{
  "poli_id": 27,
  "tanggal": "2026-06-21",
  "bayar": "2",
  "jenis_pasien": "umum",
  "dokter_id": "42",
  "queue_group": "PENYAKIT DALAM",
  "is_jkn": false
}
```

Keterangan body:

| Field | Wajib | Keterangan |
| --- | --- | --- |
| `poli_id` | Ya | ID poli dari endpoint opsi |
| `tanggal` | Ya | Format `YYYY-MM-DD` |
| `bayar` | Opsional | Kode cara bayar. Default API: `2` |
| `jenis_pasien` | Opsional | Endpoint ini tetap dipaksa menjadi `umum` |
| `dokter_id` | Opsional | ID dokter/pegawai dari `dokter_list` |
| `queue_group` | Opsional | Kelompok antrian. Jika kosong API pakai kelompok/nama poli |
| `is_jkn` | Opsional | Diabaikan untuk endpoint ini; booking tetap `umum` |

Response sukses:

```json
{
  "success": true,
  "data": {
    "message": "booking berhasil dibuat",
    "data": {
      "registration_id": 0,
      "dummy_id": 335894,
      "queue_id": 0,
      "registration_code": "21062026INT18",
      "queue_number": "18",
      "queue_code": "U-018",
      "queue_group": "PENYAKIT DALAM",
      "poli_id": 27,
      "poli_name": "Klinik Penyakit Dalam",
      "queue_date": "2026-06-21",
      "service_mode": "umum",
      "source": "mobile_umum",
      "existing": false
    }
  }
}
```

Jika pasien sudah punya booking umum mobile pada tanggal yang sama, API mengembalikan data booking yang sudah ada dengan `existing = true`. Ini dipakai agar pendaftaran tidak bisa dua kali dalam sehari.

Response error umum:

```json
{
  "success": false,
  "message": "no rm belum terhubung ke akun mobile"
}
```

Validasi penting:

- Scope akun/pasien memakai `user_id + patient_id` dari Bearer session. Email dan No. RM hanya atribut server/database, bukan otorisasi dari query/body request.
- `user_mobile.patient_id` wajib mengarah ke `pasiens.id`.
- No. RM pada sesi harus sama dengan data `pasiens` yang terhubung.
- Satu pasien hanya boleh membuat satu booking umum mobile per tanggal, baik data masih di `registrasis_dummy` maupun sudah masuk `registrasis`.
- Hari libur nasional dan cuti bersama tidak bisa dipakai booking.
- Hari Minggu tidak bisa dipakai booking.

### 4. History Nomor Antrian Pasien

```http
GET /api/v1/mobile/booking/general/mine
Authorization: Bearer REPLACE_ME # gitleaks:allow
```

Query parameter:

| Parameter | Wajib | Keterangan |
| --- | --- | --- |
| `tanggal` / `date` | Opsional | Filter tanggal `YYYY-MM-DD` |
| `all_dates` | Opsional | Isi `1` untuk semua tanggal |
| `status` | Opsional | `menunggu`, `dipanggil`, `selesai`, `batal`, `semua` |
| `limit` | Opsional | Default `50`, maksimal `200` |

Contoh:

```bash
curl -H "Authorization: Bearer REPLACE_ME" \ # gitleaks:allow
  "http://127.0.0.1:8080/api/v1/mobile/booking/general/mine?all_dates=1&limit=30"
```

Response:

```json
{
  "success": true,
  "data": {
    "items": [
      {
        "registration_id": 0,
        "dummy_id": 335894,
        "queue_id": 0,
        "registration_code": "21062026INT18",
        "queue_number": "21062026INT18",
        "queue_code": "U-018",
        "queue_group": "Klinik Penyakit Dalam",
        "queue_date": "2026-06-21",
        "created_at": "2026-06-21 09:10",
        "registration_status": "pending",
        "status_reg": "pending",
        "queue_status": "pending",
        "queue_status_label": "menunggu",
        "call_status": "",
        "already_called": false,
        "patient_id": 1234,
        "no_rm": "123456",
        "nama_pasien": "NAMA PASIEN",
        "poli_id": 27,
        "poli_name": "Klinik Penyakit Dalam",
        "dokter_id": 42,
        "nama_dokter": "dr. Contoh",
        "input_from": "mobile_umum",
        "source": "mobile_umum",
        "jenis_pasien": "umum",
        "bayar": "2",
        "keterangan": "Booking antrian umum mobile"
      }
    ],
    "status_counts": [
      {
        "status": "menunggu",
        "count": 1
      }
    ]
  }
}
```

### 5. List Booking untuk Pendaftaran (dinonaktifkan)

```http
GET /api/v1/mobile/booking/general
```

Route ini tidak diregistrasikan pada API mobile publik karena menampilkan data lintas pasien dan belum ada role staf. Dashboard pendaftaran harus memakai jalur internal yang memiliki autentikasi dan otorisasi staf; sampai mekanisme role tersebut tersedia, gunakan query operasional berikut hanya dari akses database yang berwenang.

## Query Pendaftaran dari Database

Bagian pendaftaran dapat mengambil antrean umum mobile dari `registrasis_dummy`.

```sql
SELECT
    rd.id AS dummy_id,
    rd.registrasi_id,
    rd.kodebooking,
    rd.nomorantrian,
    rd.angkaantrian,
    rd.no_rm,
    rd.nik,
    rd.nama,
    rd.no_hp,
    rd.tglperiksa,
    rd.kode_poli,
    COALESCE(po_exact.id, po_code.id) AS poli_id,
    COALESCE(po_exact.nama, po_code.nama) AS nama_poli,
    rd.kode_dokter,
    rd.dokter_id,
    pg.nama AS nama_dokter,
    rd.kode_cara_bayar,
    rd.jenisdaftar,
    rd.jenisrequest,
    rd.status,
    rd.flag,
    rd.jampraktek,
    rd.keterangan,
    rd.request,
    rd.created_at,
    rd.updated_at
FROM registrasis_dummy rd
LEFT JOIN polis po_exact
    ON JSON_VALID(rd.request)
   AND po_exact.id = CAST(JSON_UNQUOTE(JSON_EXTRACT(rd.request, '$.poli_id')) AS UNSIGNED)
LEFT JOIN polis po_code
    ON po_exact.id IS NULL
   AND po_code.id = (
        SELECT p2.id
        FROM polis p2
        WHERE rd.kode_poli <> ''
          AND (p2.bpjs = rd.kode_poli OR CAST(p2.id AS CHAR) = rd.kode_poli)
        ORDER BY p2.id ASC
        LIMIT 1
   )
LEFT JOIN pegawais pg
    ON pg.id = rd.dokter_id
WHERE rd.flag = 'mobile_umum'
  AND rd.jenisdaftar = 'android'
  AND rd.status IN ('pending', 'terdaftar', 'checkin', 'dilayani')
ORDER BY rd.tglperiksa ASC, rd.id ASC;
```

Query per tanggal:

```sql
SELECT
    rd.id AS dummy_id,
    rd.kodebooking,
    rd.nomorantrian,
    rd.angkaantrian,
    rd.no_rm,
    rd.nama,
    rd.tglperiksa,
    rd.kode_poli,
    COALESCE(po_exact.nama, po_code.nama) AS nama_poli,
    rd.dokter_id,
    pg.nama AS nama_dokter,
    rd.status,
    rd.request
FROM registrasis_dummy rd
LEFT JOIN polis po_exact
    ON JSON_VALID(rd.request)
   AND po_exact.id = CAST(JSON_UNQUOTE(JSON_EXTRACT(rd.request, '$.poli_id')) AS UNSIGNED)
LEFT JOIN polis po_code
    ON po_exact.id IS NULL
   AND po_code.id = (
        SELECT p2.id
        FROM polis p2
        WHERE rd.kode_poli <> ''
          AND (p2.bpjs = rd.kode_poli OR CAST(p2.id AS CHAR) = rd.kode_poli)
        ORDER BY p2.id ASC
        LIMIT 1
   )
LEFT JOIN pegawais pg
    ON pg.id = rd.dokter_id
WHERE rd.flag = 'mobile_umum'
  AND rd.tglperiksa = CURDATE()
ORDER BY rd.id ASC;
```

Field JSON `request` menyimpan salinan payload penting, antara lain:

```json
{
  "source": "mobile_umum",
  "flag": "mobile_umum",
  "jenisrequest": null,
  "jenisdaftar": "android",
  "pasien_id": 1234,
  "no_rm": "123456",
  "nik": "3200000000000000",
  "nama": "NAMA PASIEN",
  "tglperiksa": "2026-06-21",
  "poli_id": 27,
  "kode_poli": "INT",
  "nama_poli": "Klinik Penyakit Dalam",
  "dokter_id": "42",
  "kode_dokter": "D042",
  "nama_dokter": "dr. Contoh",
  "kodebooking": null,
  "nomorantrian": "21062026INT18",
  "angkaantrian": 18,
  "queue_code": "U-018",
  "jenis_pasien": "umum",
  "bayar": "2"
}
```

## Status

Mapping status untuk aplikasi:

| Status `registrasis_dummy.status` | Label API |
| --- | --- |
| `pending`, `terdaftar`, `checkin`, kosong | `menunggu` |
| `dilayani` | `dipanggil` |
| `selesai`, `selesai_dilayani` | `selesai` |
| `batal`, `dibatalkan` | `batal` |

## Catatan Deployment

Database produksi sudah ada. Jangan menjalankan import schema penuh.

Perubahan pendukung disimpan di `historyQuery/history.sql`. Jalankan hanya ALTER/index yang belum pernah diterapkan, dan lakukan pada jam traffic rendah bila tabel besar.
