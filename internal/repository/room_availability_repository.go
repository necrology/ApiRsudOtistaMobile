package repository

import (
	"context"
	"database/sql"
	"strings"
)

type HospitalRepository struct {
	DB *sql.DB
}

type RoomAvailability struct {
	ID           string `json:"id"`
	GeneralCode  string `json:"general_code"`
	Name         string `json:"nama"`
	ClassNames   string `json:"kelas"`
	RoomNames    string `json:"ruangan"`
	RoomCount    int64  `json:"jumlah_kamar"`
	TotalBeds    int64  `json:"jumlah_bed"`
	FilledBeds   int64  `json:"terisi"`
	EmptyBeds    int64  `json:"kosong"`
	ActiveStays  int64  `json:"rawat_inap_aktif"`
	Renovation   int64  `json:"renovasi"`
	LastUpdated  string `json:"updated_at"`
	Availability string `json:"keterangan"`
}

// Polyclinic adalah proyeksi publik yang eksplisit dari tabel polis. Jangan
// mengganti endpoint ini dengan SELECT * atau parameter kolom dari klien.
type Polyclinic struct {
	ID          int64  `json:"id"`
	Name        string `json:"nama"`
	Group       string `json:"kelompok"`
	Type        string `json:"politype"`
	BPJSCode    string `json:"bpjs"`
	RoomCode    string `json:"kode_ruangan"`
	OpensAt     string `json:"buka"`
	ClosesAt    string `json:"tutup"`
	Practice    string `json:"praktik"`
	Quota       int64  `json:"kuota"`
	OnlineQuota int64  `json:"kuota_online"`
	Filled      int64  `json:"terisi"`
}

func NewHospitalRepository(db *sql.DB) *HospitalRepository {
	return &HospitalRepository{DB: db}
}

func (r *HospitalRepository) Polyclinics(
	ctx context.Context,
	search string,
	page int,
	limit int,
) ([]Polyclinic, Pagination, error) {
	search = strings.TrimSpace(search)
	if page < 1 {
		page = 1
	}
	limit = normalizeMobileLimit(limit)
	like := "%" + search + "%"

	where := `
		WHERE ? = ''
		OR COALESCE(p.nama, '') LIKE ?
		OR COALESCE(p.kelompok, '') LIKE ?
		OR COALESCE(p.politype, '') LIKE ?
		OR COALESCE(p.bpjs, '') LIKE ?
		OR COALESCE(p.kode_ruangan, '') LIKE ?
	`
	args := []any{search, like, like, like, like, like}

	var total int64
	if err := r.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM polis p "+where, args...).Scan(&total); err != nil {
		return nil, Pagination{}, err
	}

	offset := (page - 1) * limit
	query := `
		SELECT
			p.id,
			COALESCE(p.nama, ''),
			COALESCE(p.kelompok, ''),
			COALESCE(p.politype, ''),
			COALESCE(p.bpjs, ''),
			COALESCE(p.kode_ruangan, ''),
			COALESCE(DATE_FORMAT(p.buka, '%H:%i'), ''),
			COALESCE(DATE_FORMAT(p.tutup, '%H:%i'), ''),
			COALESCE(p.praktik, ''),
			COALESCE(p.kuota, 0),
			COALESCE(p.kuota_online, 0),
			COALESCE(p.terisi, 0)
		FROM polis p
	` + where + `
		ORDER BY COALESCE(p.nama, '') ASC, p.id ASC
		LIMIT ? OFFSET ?
	`
	listArgs := append(append([]any{}, args...), limit, offset)
	rows, err := r.DB.QueryContext(ctx, query, listArgs...)
	if err != nil {
		return nil, Pagination{}, err
	}
	defer rows.Close()

	items := make([]Polyclinic, 0)
	for rows.Next() {
		var item Polyclinic
		if err = rows.Scan(
			&item.ID,
			&item.Name,
			&item.Group,
			&item.Type,
			&item.BPJSCode,
			&item.RoomCode,
			&item.OpensAt,
			&item.ClosesAt,
			&item.Practice,
			&item.Quota,
			&item.OnlineQuota,
			&item.Filled,
		); err != nil {
			return nil, Pagination{}, err
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, Pagination{}, err
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))
	return items, Pagination{
		Page:       page,
		Limit:      limit,
		HasNext:    int64(offset+len(items)) < total,
		TotalRows:  &total,
		TotalPages: &totalPages,
	}, nil
}

func (r *HospitalRepository) RoomAvailabilities(ctx context.Context, search string, limit int) ([]RoomAvailability, error) {
	search = strings.TrimSpace(search)
	limit = normalizeMobileLimit(limit)
	if limit < 20 {
		limit = 20
	}

	query := `
	SELECT
		room_group.id,
		room_group.general_code,
		room_group.nama,
		room_group.kelas,
		room_group.ruangan,
		room_group.jumlah_kamar,
		room_group.jumlah_bed,
		room_group.terisi,
		GREATEST(room_group.jumlah_bed - room_group.terisi, 0) AS kosong,
		room_group.rawat_inap_aktif,
		0 AS renovasi,
		room_group.updated_at,
		CASE
			WHEN room_group.jumlah_bed <= 0 THEN 'Bed belum tersedia'
			WHEN GREATEST(room_group.jumlah_bed - room_group.terisi, 0) > 0 THEN 'Tersedia'
			ELSE 'Penuh'
		END AS keterangan
	FROM (
		SELECT
			MIN(COALESCE(NULLIF(TRIM(a.general_code), ''), CAST(a.id AS CHAR))) AS id,
			MIN(COALESCE(NULLIF(TRIM(a.general_code), ''), CAST(a.id AS CHAR))) AS general_code,
			COALESCE(
				NULLIF(GROUP_CONCAT(DISTINCT NULLIF(TRIM(a.kelompok), '') ORDER BY a.kelompok SEPARATOR ', '), ''),
				MIN(COALESCE(NULLIF(TRIM(a.general_code), ''), CAST(a.id AS CHAR)))
			) AS nama,
			COALESCE(NULLIF(GROUP_CONCAT(DISTINCT NULLIF(TRIM(a.kelompok), '') ORDER BY a.kelompok SEPARATOR ', '), ''), '') AS kelas,
			COALESCE(NULLIF(GROUP_CONCAT(DISTINCT NULLIF(TRIM(b.nama), '') ORDER BY b.nama SEPARATOR ', '), ''), '') AS ruangan,
			COUNT(DISTINCT b.id) AS jumlah_kamar,
			COUNT(DISTINCT c.id) AS jumlah_bed,
			COUNT(DISTINCT CASE WHEN active_reg.id IS NOT NULL THEN c.id END) AS terisi,
			COUNT(DISTINCT CASE WHEN active_reg.id IS NOT NULL THEN c.id END) AS rawat_inap_aktif,
			COALESCE(DATE_FORMAT(MAX(COALESCE(c.updated_at, b.updated_at, a.updated_at)), '%Y-%m-%d %H:%i'), '') AS updated_at
		FROM kelompok_kelas a
		LEFT JOIN kamars b
			ON b.kelompokkelas_id = a.id
			AND b.deleted_at IS NULL
			AND COALESCE(b.hidden, '') <> 'Y'
		LEFT JOIN beds c
			ON c.kamar_id = b.id
			AND c.deleted_at IS NULL
			AND COALESCE(c.hidden, '') <> 'Y'
		LEFT JOIN rawatinaps ri
			ON ri.bed_id = c.id
			AND (ri.deleted_at IS NULL OR ri.deleted_at = '')
			AND ri.tgl_keluar IS NULL
		LEFT JOIN registrasis active_reg
			ON active_reg.id = ri.registrasi_id
			AND (active_reg.pulang IS NULL OR active_reg.pulang = '')
		WHERE a.deleted_at IS NULL
		GROUP BY COALESCE(NULLIF(TRIM(a.general_code), ''), CAST(a.id AS CHAR))
	) room_group
	WHERE
		? = ''
		OR room_group.general_code LIKE ?
		OR room_group.nama LIKE ?
		OR room_group.ruangan LIKE ?
	ORDER BY room_group.general_code ASC, room_group.nama ASC
	LIMIT ?
	`

	like := "%" + search + "%"
	rows, err := r.DB.QueryContext(ctx, query, search, like, like, like, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]RoomAvailability, 0)
	for rows.Next() {
		var item RoomAvailability
		if err = rows.Scan(
			&item.ID,
			&item.GeneralCode,
			&item.Name,
			&item.ClassNames,
			&item.RoomNames,
			&item.RoomCount,
			&item.TotalBeds,
			&item.FilledBeds,
			&item.EmptyBeds,
			&item.ActiveStays,
			&item.Renovation,
			&item.LastUpdated,
			&item.Availability,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}
