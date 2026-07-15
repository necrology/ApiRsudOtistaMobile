package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type MobilePatientRepository struct {
	DB *sql.DB
}

type MobilePatientProfile struct {
	ID          int64  `json:"id"`
	NoRM        string `json:"no_rm"`
	FullName    string `json:"nama"`
	Gender      string `json:"kelamin"`
	BirthPlace  string `json:"tmplahir"`
	BirthDate   string `json:"tgllahir"`
	Address     string `json:"alamat"`
	PhoneNumber string `json:"nohp"`
	BloodType   string `json:"golda"`
}

type MobilePatientVisit struct {
	ID             int64  `json:"id"`
	Registration   string `json:"reg_id"`
	QueueNumber    string `json:"nomorantrian"`
	VisitDate      string `json:"tanggal_kunjungan"`
	CreatedAt      string `json:"created_at"`
	FinishedAt     string `json:"tgl_pulang"`
	PatientType    string `json:"jenis_pasien"`
	PaymentType    string `json:"bayar"`
	ServiceType    string `json:"tipe_rawat"`
	Status         string `json:"status"`
	Polyclinic     string `json:"poli"`
	Doctor         string `json:"dokter"`
	QueueGroup     string `json:"kelompok_antrian"`
	ClassGroup     string `json:"kelas_rawat,omitempty"`
	RoomName       string `json:"kamar,omitempty"`
	BedName        string `json:"bed,omitempty"`
	InpatientSince string `json:"tgl_masuk_rawat_inap,omitempty"`
}

type MobilePatientMedicalSummary struct {
	ID             int64  `json:"id"`
	RegistrationID int64  `json:"registrasi_id"`
	Registration   string `json:"reg_id"`
	VisitDate      string `json:"tanggal_kunjungan"`
	Polyclinic     string `json:"poli"`
	Doctor         string `json:"dokter"`
	Diagnosis      string `json:"diagnosa"`
	Action         string `json:"tindakan"`
	Note           string `json:"keterangan"`
}

type MobilePatientMedicalResumeDocument struct {
	PatientID           int64
	PatientName         string
	NoRM                string
	Gender              string
	BirthPlace          string
	BirthDate           string
	Address             string
	PhoneNumber         string
	RegistrationID      int64
	Registration        string
	ServiceType         string
	Polyclinic          string
	Doctor              string
	VisitDate           string
	DischargeDate       string
	ControlDate         string
	ControlPolyclinic   string
	Anamnesis           string
	PhysicalExam        string
	NursingCare         string
	InpatientIndication string
	DiseaseHistory      string
	SupportingExam      string
	MainDiagnosis       string
	MainICD             string
	AdditionalDiagnosis string
	AdditionalICD       string
	Action              string
	ActionICD           string
	Medication          string
	Therapy             string
	Note                string
	DischargeMethod     string
	DischargeCondition  string
	FollowUpInstruction string
	IsInpatient         bool
}

type MobilePatientLabResult struct {
	ID                int64                          `json:"id"`
	NoLab             string                         `json:"no_lab"`
	Registration      string                         `json:"reg_id"`
	VisitDate         string                         `json:"tanggal_kunjungan"`
	Polyclinic        string                         `json:"poli"`
	Doctor            string                         `json:"dokter"`
	Responsible       string                         `json:"penanggungjawab"`
	ExamDate          string                         `json:"tgl_pemeriksaan"`
	ReceivedDate      string                         `json:"tgl_bahanditerima"`
	ResultDate        string                         `json:"tgl_hasilselesai"`
	PrintDate         string                         `json:"tgl_cetak"`
	StartTime         string                         `json:"jam"`
	FinishTime        string                         `json:"jamkeluar"`
	Sample            string                         `json:"sample"`
	Message           string                         `json:"pesan"`
	Impression        string                         `json:"kesan"`
	Suggestion        string                         `json:"saran"`
	OrderExaminations string                         `json:"pemeriksaan_order"`
	OrderType         string                         `json:"jenis_order"`
	LabType           string                         `json:"tipe_lab"`
	Diagnosis         string                         `json:"diagnosa"`
	Details           []MobilePatientLabResultDetail `json:"rincian"`
}

type MobilePatientLabResultDetail struct {
	ID            int64  `json:"id"`
	Source        string `json:"source,omitempty"`
	Code          string `json:"kode_jenis_tes,omitempty"`
	Section       string `json:"section"`
	Category      string `json:"kategori"`
	Examination   string `json:"pemeriksaan"`
	Reference     string `json:"rujukan"`
	ReferenceLow  string `json:"nilai_rujukan_bawah"`
	ReferenceHigh string `json:"nilai_rujukan_atas"`
	Unit          string `json:"satuan"`
	ResultText    string `json:"hasil_text"`
	ResultValue   string `json:"hasil"`
	Flag          string `json:"flag,omitempty"`
	Note          string `json:"catatan,omitempty"`
	Sequence      int64  `json:"sequence,omitempty"`
	DrawTime      string `json:"draw_time,omitempty"`
	ValidateTime  string `json:"validate_time,omitempty"`
}

type MobilePatientRadiologyResult struct {
	ID              int64  `json:"id"`
	Registration    string `json:"reg_id"`
	VisitDate       string `json:"tanggal_kunjungan"`
	Polyclinic      string `json:"poli"`
	Doctor          string `json:"dokter"`
	DocumentNumber  string `json:"no_dokument"`
	ExamDate        string `json:"tanggal_periksa"`
	ResultDate      string `json:"tanggal_ekspertise"`
	Examination     string `json:"pemeriksaan"`
	OrderType       string `json:"jenis_order"`
	Status          string `json:"status"`
	Source          string `json:"source"`
	ClinicalNote    string `json:"klinis"`
	ResultSummary   string `json:"resume"`
	Expertise       string `json:"ekspertise"`
	OrderCreatedAt  string `json:"order_created_at"`
	ResultCreatedAt string `json:"result_created_at"`
}

type MobilePatientPrescription struct {
	ID           int64                             `json:"id"`
	NoResep      string                            `json:"no_resep"`
	Registration string                            `json:"reg_id"`
	VisitDate    string                            `json:"tanggal_kunjungan"`
	Polyclinic   string                            `json:"poli"`
	Doctor       string                            `json:"dokter"`
	CreatedAt    string                            `json:"created_at"`
	Status       string                            `json:"status"`
	Note         string                            `json:"catatan"`
	Details      []MobilePatientPrescriptionDetail `json:"rincian"`
}

type MobilePatientPrescriptionDetail struct {
	ID          int64  `json:"id"`
	DrugName    string `json:"nama_obat"`
	DrugCode    string `json:"kode_obat"`
	Unit        string `json:"satuan"`
	Quantity    int64  `json:"jumlah"`
	HowToUse    string `json:"cara_minum"`
	Dose        string `json:"takaran"`
	InfoPrimary string `json:"informasi1"`
	InfoExtra   string `json:"informasi2"`
	Note        string `json:"catatan"`
	Label       string `json:"etiket"`
	Compound    string `json:"obat_racikan"`
	Chronic     string `json:"is_kronis"`
}

func NewMobilePatientRepository(db *sql.DB) *MobilePatientRepository {
	return &MobilePatientRepository{DB: db}
}

func (r *MobilePatientRepository) Profile(ctx context.Context, userID int64, patientID int64) (*MobilePatientProfile, error) {
	if userID <= 0 || patientID <= 0 {
		return nil, sql.ErrNoRows
	}
	query := `
	SELECT
		p.id,
		COALESCE(p.no_rm, ''),
		COALESCE(p.nama, ''),
		COALESCE(p.kelamin, ''),
		COALESCE(p.tmplahir, ''),
		COALESCE(p.tgllahir, ''),
		COALESCE(p.alamat, ''),
		COALESCE(p.nohp, COALESCE(p.notlp, '')),
		COALESCE(p.golda, '')
	FROM user_mobile u
	INNER JOIN pasiens p ON p.id = u.patient_id
	WHERE u.id = ?
	AND u.patient_id = ?
	AND COALESCE(u.is_deleted, false) = false
	LIMIT 1
	`

	var profile MobilePatientProfile
	if err := r.DB.QueryRowContext(ctx, query, userID, patientID).Scan(
		&profile.ID,
		&profile.NoRM,
		&profile.FullName,
		&profile.Gender,
		&profile.BirthPlace,
		&profile.BirthDate,
		&profile.Address,
		&profile.PhoneNumber,
		&profile.BloodType,
	); err != nil {
		return nil, err
	}

	profile.Gender = normalizeGender(profile.Gender)

	return &profile, nil
}

func (r *MobilePatientRepository) Visits(ctx context.Context, userID int64, patientID int64, limit int) ([]MobilePatientVisit, error) {
	profile, err := r.Profile(ctx, userID, patientID)
	if err != nil {
		return nil, err
	}

	limit = normalizeMobileLimit(limit)

	query := `
	SELECT
		r.id,
		COALESCE(r.reg_id, ''),
		COALESCE(NULLIF(r.nomorantrian, ''), COALESCE(CAST(a.nomor AS CHAR), '')),
		COALESCE(DATE_FORMAT(ra.tgl_masuk, '%Y-%m-%d'), DATE_FORMAT(r.tgl_order, '%Y-%m-%d'), DATE_FORMAT(r.created_at, '%Y-%m-%d'), ''),
		COALESCE(DATE_FORMAT(r.created_at, '%Y-%m-%d %H:%i'), ''),
		COALESCE(DATE_FORMAT(ra.tgl_keluar, '%Y-%m-%d %H:%i'), DATE_FORMAT(r.tgl_pulang, '%Y-%m-%d %H:%i'), ''),
		COALESCE(r.jenis_pasien, ''),
		COALESCE(NULLIF(cb.carabayar, ''), NULLIF(r.bayar, ''), ''),
		COALESCE(NULLIF(r.tipe_rawat, ''), CASE WHEN ra.id IS NOT NULL THEN 'Rawat Inap' ELSE '' END),
		CASE
			WHEN ra.id IS NOT NULL AND ra.tgl_keluar IS NULL THEN 'Dirawat'
			ELSE COALESCE(r.status, '')
		END,
		COALESCE(NULLIF(po.nama, ''), kk.kelompok, ''),
		COALESCE(d.nama, ''),
		COALESCE(a.kelompok, ''),
		COALESCE(kk.kelompok, ''),
		COALESCE(k.nama, ''),
		COALESCE(b.nama, ''),
		COALESCE(DATE_FORMAT(ra.tgl_masuk, '%Y-%m-%d %H:%i'), '')
	FROM registrasis r
	LEFT JOIN polis po ON po.id = r.poli_id
	LEFT JOIN pegawais d ON d.id = CAST(r.dokter_id AS UNSIGNED)
	LEFT JOIN antrian_poli a ON a.id = r.antrian_poli_id
	LEFT JOIN carabayars cb ON cb.id = CAST(NULLIF(r.bayar, '') AS UNSIGNED)
	LEFT JOIN rawatinaps ra ON ra.id = (
		SELECT ra2.id
		FROM rawatinaps ra2
		WHERE ra2.registrasi_id = r.id
		AND (ra2.deleted_at IS NULL OR ra2.deleted_at = '')
		ORDER BY COALESCE(ra2.tgl_masuk, ra2.created_at) DESC, ra2.id DESC
		LIMIT 1
	)
	LEFT JOIN kelompok_kelas kk ON kk.id = ra.kelompokkelas_id
	LEFT JOIN kamars k ON k.id = ra.kamar_id
	LEFT JOIN beds b ON b.id = ra.bed_id
	WHERE r.pasien_id = ?
	ORDER BY COALESCE(DATE(ra.tgl_masuk), r.tgl_order, DATE(r.created_at)) DESC, r.id DESC
	LIMIT ?
	`

	rows, err := r.DB.QueryContext(ctx, query, profile.ID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	visits := make([]MobilePatientVisit, 0)
	for rows.Next() {
		var visit MobilePatientVisit
		if err = rows.Scan(
			&visit.ID,
			&visit.Registration,
			&visit.QueueNumber,
			&visit.VisitDate,
			&visit.CreatedAt,
			&visit.FinishedAt,
			&visit.PatientType,
			&visit.PaymentType,
			&visit.ServiceType,
			&visit.Status,
			&visit.Polyclinic,
			&visit.Doctor,
			&visit.QueueGroup,
			&visit.ClassGroup,
			&visit.RoomName,
			&visit.BedName,
			&visit.InpatientSince,
		); err != nil {
			return nil, err
		}
		visit.PatientType = normalizePatientType(visit.PatientType)
		visit.PaymentType = normalizePaymentType(visit.PaymentType)
		visits = append(visits, visit)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return visits, nil
}

func (r *MobilePatientRepository) MedicalSummaries(ctx context.Context, userID int64, patientID int64, limit int) ([]MobilePatientMedicalSummary, error) {
	profile, err := r.Profile(ctx, userID, patientID)
	if err != nil {
		return nil, err
	}

	limit = normalizeMobileLimit(limit)
	scanLimit := limit * 4
	if scanLimit < 20 {
		scanLimit = 20
	}
	if scanLimit > 100 {
		scanLimit = 100
	}

	query := `
	WITH recent_registrations AS (
		SELECT
			r.id,
			r.reg_id,
			r.tgl_order,
			r.created_at,
			r.poli_id,
			r.dokter_id,
			r.diagnosa_awal,
			r.diagnosa_inap,
			r.icd
		FROM registrasis r
		WHERE r.pasien_id = ?
		ORDER BY COALESCE(r.tgl_order, DATE(r.created_at)) DESC, r.id DESC
		LIMIT ?
	)
	SELECT
		COALESCE(MAX(rp.id), r.id),
		r.id,
		COALESCE(r.reg_id, ''),
		COALESCE(DATE_FORMAT(r.tgl_order, '%Y-%m-%d'), DATE_FORMAT(r.created_at, '%Y-%m-%d'), ''),
		COALESCE(po.nama, ''),
		COALESCE(d.nama, ''),
		COALESCE(
			NULLIF(GROUP_CONCAT(DISTINCT NULLIF(rp.diagnosa, '') SEPARATOR '\n'), ''),
			NULLIF(GROUP_CONCAT(DISTINCT NULLIF(CONCAT_WS(' - ', NULLIF(TRIM(j10.icd10), ''), NULLIF(TRIM(COALESCE(NULLIF(NULLIF(TRIM(j10.diagnosis), ''), TRIM(j10.icd10)), ij10.nama, '')), '')), '') SEPARATOR '\n'), ''),
			NULLIF(GROUP_CONCAT(DISTINCT NULLIF(CONCAT_WS(' - ', NULLIF(TRIM(p10.icd10), ''), NULLIF(TRIM(COALESCE(NULLIF(NULLIF(TRIM(p10.diagnosis), ''), TRIM(p10.icd10)), ip10.nama, '')), '')), '') SEPARATOR '\n'), ''),
			NULLIF(CONCAT_WS(' - ', NULLIF(TRIM(r.diagnosa_awal), ''), NULLIF(TRIM(COALESCE(ir_awal.nama, '')), '')), ''),
			NULLIF(CONCAT_WS(' - ', NULLIF(TRIM(r.diagnosa_inap), ''), NULLIF(TRIM(COALESCE(ir_inap.nama, '')), '')), ''),
			NULLIF(CONCAT_WS(' - ', NULLIF(TRIM(r.icd), ''), NULLIF(TRIM(COALESCE(ir_icd.nama, '')), '')), ''),
			''
		) AS diagnosis_text,
		COALESCE(
			NULLIF(GROUP_CONCAT(DISTINCT NULLIF(rp.tindakan, '') SEPARATOR '\n'), ''),
			NULLIF(GROUP_CONCAT(DISTINCT NULLIF(CONCAT_WS(' - ', NULLIF(TRIM(j9.icd9), ''), NULLIF(TRIM(COALESCE(NULLIF(NULLIF(TRIM(j9.diagnosis), ''), TRIM(j9.icd9)), ij9.nama, '')), '')), '') SEPARATOR '\n'), ''),
			NULLIF(GROUP_CONCAT(DISTINCT NULLIF(CONCAT_WS(' - ', NULLIF(TRIM(p9.icd9), ''), NULLIF(TRIM(COALESCE(NULLIF(NULLIF(TRIM(p9.diagnosis), ''), TRIM(p9.icd9)), ip9.nama, '')), '')), '') SEPARATOR '\n'), ''),
			''
		) AS action_text,
		COALESCE(NULLIF(GROUP_CONCAT(DISTINCT NULLIF(rp.keterangan, '') SEPARATOR '\n'), ''), '')
	FROM recent_registrations r
	LEFT JOIN resume_pasiens rp ON rp.registrasi_id = r.id
	LEFT JOIN jkn_icd10s j10 ON j10.registrasi_id = r.id
	LEFT JOIN jkn_icd9s j9 ON j9.registrasi_id = r.id
	LEFT JOIN perawatan_icd10s p10 ON p10.registrasi_id = r.id
	LEFT JOIN perawatan_icd9s p9 ON p9.registrasi_id = r.id
	LEFT JOIN icd10s ij10 ON ij10.nomor = j10.icd10
	LEFT JOIN icd9s ij9 ON ij9.nomor = j9.icd9
	LEFT JOIN icd10s ip10 ON ip10.nomor = p10.icd10
	LEFT JOIN icd9s ip9 ON ip9.nomor = p9.icd9
	LEFT JOIN icd10s ir_awal ON ir_awal.nomor = TRIM(r.diagnosa_awal)
	LEFT JOIN icd10s ir_inap ON ir_inap.nomor = TRIM(r.diagnosa_inap)
	LEFT JOIN icd10s ir_icd ON ir_icd.nomor = TRIM(r.icd)
	LEFT JOIN polis po ON po.id = r.poli_id
	LEFT JOIN pegawais d ON d.id = CAST(r.dokter_id AS UNSIGNED)
	GROUP BY
		r.id,
		r.reg_id,
		r.tgl_order,
		r.created_at,
		po.nama,
		d.nama,
		r.diagnosa_awal,
		r.diagnosa_inap,
		r.icd,
		ir_awal.nama,
		ir_inap.nama,
		ir_icd.nama
	HAVING diagnosis_text <> '' OR action_text <> ''
	ORDER BY COALESCE(r.tgl_order, DATE(r.created_at)) DESC, COALESCE(MAX(rp.id), r.id) DESC
	LIMIT ?
	`

	rows, err := r.DB.QueryContext(ctx, query, profile.ID, scanLimit, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	summaries := make([]MobilePatientMedicalSummary, 0)
	for rows.Next() {
		var summary MobilePatientMedicalSummary
		if err = rows.Scan(
			&summary.ID,
			&summary.RegistrationID,
			&summary.Registration,
			&summary.VisitDate,
			&summary.Polyclinic,
			&summary.Doctor,
			&summary.Diagnosis,
			&summary.Action,
			&summary.Note,
		); err != nil {
			return nil, err
		}
		summaries = append(summaries, summary)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return summaries, nil
}

func (r *MobilePatientRepository) MedicalResumeDocument(ctx context.Context, userID int64, patientID int64, registrationID int64) (*MobilePatientMedicalResumeDocument, error) {
	profile, err := r.Profile(ctx, userID, patientID)
	if err != nil {
		return nil, err
	}
	if registrationID <= 0 {
		return nil, sql.ErrNoRows
	}

	query := `
	SELECT
		p.id,
		COALESCE(p.nama, ''),
		COALESCE(p.no_rm, ''),
		COALESCE(p.kelamin, ''),
		COALESCE(p.tmplahir, ''),
		COALESCE(p.tgllahir, ''),
		COALESCE(p.alamat, ''),
		COALESCE(p.nohp, COALESCE(p.notlp, '')),
		r.id,
		COALESCE(r.reg_id, ''),
		COALESCE(r.tipe_rawat, ''),
		COALESCE(po.nama, ''),
		COALESCE(d.nama, ''),
		COALESCE(DATE_FORMAT(ra.tgl_masuk, '%d-%m-%Y %H:%i'), DATE_FORMAT(r.tgl_order, '%d-%m-%Y'), DATE_FORMAT(r.created_at, '%d-%m-%Y %H:%i'), ''),
		COALESCE(DATE_FORMAT(ra.tgl_keluar, '%d-%m-%Y %H:%i'), DATE_FORMAT(r.tgl_pulang, '%d-%m-%Y %H:%i'), ''),
		CASE WHEN ra.id IS NULL THEN 0 ELSE 1 END
	FROM registrasis r
	INNER JOIN pasiens p ON p.id = r.pasien_id
	LEFT JOIN polis po ON po.id = r.poli_id
	LEFT JOIN pegawais d ON d.id = CAST(r.dokter_id AS UNSIGNED)
	LEFT JOIN rawatinaps ra ON ra.id = (
		SELECT ra2.id
		FROM rawatinaps ra2
		WHERE ra2.registrasi_id = r.id
		AND (ra2.deleted_at IS NULL OR ra2.deleted_at = '')
		ORDER BY ra2.id DESC
		LIMIT 1
	)
	WHERE r.id = ?
	AND r.pasien_id = ?
	LIMIT 1
	`

	var doc MobilePatientMedicalResumeDocument
	if err = r.DB.QueryRowContext(ctx, query, registrationID, profile.ID).Scan(
		&doc.PatientID,
		&doc.PatientName,
		&doc.NoRM,
		&doc.Gender,
		&doc.BirthPlace,
		&doc.BirthDate,
		&doc.Address,
		&doc.PhoneNumber,
		&doc.RegistrationID,
		&doc.Registration,
		&doc.ServiceType,
		&doc.Polyclinic,
		&doc.Doctor,
		&doc.VisitDate,
		&doc.DischargeDate,
		&doc.IsInpatient,
	); err != nil {
		return nil, err
	}

	doc.Gender = normalizeGender(doc.Gender)
	doc.IsInpatient = doc.IsInpatient || strings.Contains(strings.ToLower(doc.ServiceType), "inap")
	doc.ControlPolyclinic = doc.Polyclinic

	if err = r.fillMedicalResumeContent(ctx, &doc); err != nil {
		return nil, err
	}

	return &doc, nil
}

func (r *MobilePatientRepository) fillMedicalResumeContent(ctx context.Context, doc *MobilePatientMedicalResumeDocument) error {
	content := map[string]any{}
	var rawContent sql.NullString
	if err := r.DB.QueryRowContext(ctx, `
		SELECT COALESCE(CAST(content AS CHAR), '')
		FROM emr_resume
		WHERE registrasi_id = ?
		AND (deleted_at IS NULL OR deleted_at = '')
		ORDER BY id DESC
		LIMIT 1
	`, doc.RegistrationID).Scan(&rawContent); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if rawContent.Valid && strings.TrimSpace(rawContent.String) != "" {
		_ = json.Unmarshal([]byte(rawContent.String), &content)
	}

	doc.Anamnesis = resumeJSONValue(content, "anamnesa")
	doc.PhysicalExam = resumeJSONValue(content, "pemeriksaan_fisik")
	doc.MainDiagnosis = resumeJSONValue(content, "diagnosa_utama")
	doc.MainICD = resumeJSONValue(content, "icdx_diagnosa_utama")
	doc.AdditionalDiagnosis = resumeJSONValue(content, "diagnosa_tambahan")
	doc.AdditionalICD = resumeJSONValue(content, "icdx_diagnosa_tambahan")
	doc.Action = resumeJSONValue(content, "tindakan")
	doc.ActionICD = resumeJSONValue(content, "icdix_tindakan")
	doc.Medication = resumeJSONValue(content, "pengobatan")
	doc.Therapy = doc.Medication
	doc.DischargeMethod = resumeJSONValue(content, "cara_pulang")
	doc.DischargeCondition = resumeJSONValue(content, "kondisi_pulang")
	doc.Note = resumeJSONValue(content, "catatan")

	var (
		resumeDiagnosis string
		resumeAction    string
		resumeNote      string
	)
	if err := r.DB.QueryRowContext(ctx, `
		SELECT
			COALESCE(NULLIF(GROUP_CONCAT(DISTINCT NULLIF(diagnosa, '') SEPARATOR '\n'), ''), ''),
			COALESCE(NULLIF(GROUP_CONCAT(DISTINCT NULLIF(tindakan, '') SEPARATOR '\n'), ''), ''),
			COALESCE(NULLIF(GROUP_CONCAT(DISTINCT NULLIF(keterangan, '') SEPARATOR '\n'), ''), '')
		FROM resume_pasiens
		WHERE registrasi_id = ?
	`, doc.RegistrationID).Scan(&resumeDiagnosis, &resumeAction, &resumeNote); err != nil {
		return err
	}

	doc.MainDiagnosis = coalescePatientString(doc.MainDiagnosis, resumeDiagnosis)
	doc.Action = coalescePatientString(doc.Action, resumeAction)
	doc.Note = coalescePatientString(doc.Note, resumeNote)

	var (
		keluhan    string
		riwayatNow string
		vitals     string
		blood      string
		temp       sql.NullFloat64
		pulse      sql.NullInt64
	)
	if err := r.DB.QueryRowContext(ctx, `
		SELECT
			COALESCE(keluhan_utama, ''),
			COALESCE(riwayat_penyakit_sekarang, ''),
			COALESCE(tanda_vital, ''),
			COALESCE(tekanan_darah, ''),
			suhu_tubuh,
			nadi
		FROM emr_riwayats
		WHERE registrasi_id = ?
		AND (deleted_at IS NULL OR deleted_at = '')
		ORDER BY id DESC
		LIMIT 1
	`, doc.RegistrationID).Scan(&keluhan, &riwayatNow, &vitals, &blood, &temp, &pulse); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	doc.Anamnesis = coalescePatientString(doc.Anamnesis, joinPatientLines(keluhan, riwayatNow))
	physicalFallback := vitals
	if blood != "" {
		physicalFallback = joinPatientLines(physicalFallback, "Tekanan Darah: "+blood)
	}
	if temp.Valid && temp.Float64 > 0 {
		physicalFallback = joinPatientLines(physicalFallback, fmt.Sprintf("Suhu: %.1f", temp.Float64))
	}
	if pulse.Valid && pulse.Int64 > 0 {
		physicalFallback = joinPatientLines(physicalFallback, fmt.Sprintf("Nadi: %d", pulse.Int64))
	}
	doc.PhysicalExam = coalescePatientString(doc.PhysicalExam, physicalFallback)

	var inpatientNote string
	var inpatientHistory sql.NullString
	if err := r.DB.QueryRowContext(ctx, `
		SELECT
			COALESCE(keterangan, ''),
			COALESCE(riwayat, '')
		FROM emr_inap_medical_records
		WHERE registrasi_id = ?
		AND (deleted_at IS NULL OR deleted_at = '')
		ORDER BY id DESC
		LIMIT 1
	`, doc.RegistrationID).Scan(&inpatientNote, &inpatientHistory); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	doc.InpatientIndication = coalescePatientString(doc.InpatientIndication, inpatientNote)
	if inpatientHistory.Valid {
		doc.DiseaseHistory = coalescePatientString(doc.DiseaseHistory, extractResumeJSONText(inpatientHistory.String, "riwayat_penyakit_sekarang"), inpatientHistory.String)
	}

	mainICD, additionalICD, err := r.icd10Texts(ctx, doc.RegistrationID)
	if err != nil {
		return err
	}
	doc.MainICD = coalescePatientString(doc.MainICD, mainICD)
	doc.AdditionalICD = coalescePatientString(doc.AdditionalICD, additionalICD)

	actionICD, err := r.icd9Text(ctx, doc.RegistrationID)
	if err != nil {
		return err
	}
	doc.ActionICD = coalescePatientString(doc.ActionICD, actionICD)

	doc.MainDiagnosis = coalescePatientString(doc.MainDiagnosis, "-")
	doc.AdditionalDiagnosis = coalescePatientString(doc.AdditionalDiagnosis, "-")
	doc.Action = coalescePatientString(doc.Action, "-")
	doc.Medication = coalescePatientString(doc.Medication, "-")
	doc.Therapy = coalescePatientString(doc.Therapy, doc.Medication)
	doc.Anamnesis = coalescePatientString(doc.Anamnesis, "-")
	doc.PhysicalExam = coalescePatientString(doc.PhysicalExam, "-")
	doc.SupportingExam = coalescePatientString(doc.SupportingExam, "-")
	doc.Note = coalescePatientString(doc.Note, "-")

	return nil
}

func (r *MobilePatientRepository) LaboratoryResults(ctx context.Context, userID int64, patientID int64, limit int) ([]MobilePatientLabResult, error) {
	profile, err := r.Profile(ctx, userID, patientID)
	if err != nil {
		return nil, err
	}

	limit = normalizeMobileLimit(limit)

	query := `
	SELECT
		hl.id,
		COALESCE(hl.no_lab, ''),
		COALESCE(r.reg_id, ''),
		COALESCE(DATE_FORMAT(r.tgl_order, '%Y-%m-%d'), DATE_FORMAT(lr.tgl_pemeriksaan, '%Y-%m-%d'), DATE_FORMAT(hl.created_at, '%Y-%m-%d'), ''),
		COALESCE(po.nama, ''),
		COALESCE(d.nama, rd.nama, ''),
		COALESCE(NULLIF(hl.penanggungjawab, ''), lr.pemeriksa, ''),
		COALESCE(DATE_FORMAT(lr.tgl_pemeriksaan, '%Y-%m-%d'), DATE_FORMAT(hl.tgl_pemeriksaan, '%Y-%m-%d'), ''),
		COALESCE(DATE_FORMAT(hl.tgl_bahanditerima, '%Y-%m-%d'), ''),
		COALESCE(DATE_FORMAT(hl.tgl_hasilselesai, '%Y-%m-%d'), DATE_FORMAT(lr.tgl_pemeriksaan, '%Y-%m-%d'), ''),
		COALESCE(DATE_FORMAT(hl.tgl_cetak, '%Y-%m-%d'), ''),
		COALESCE(NULLIF(hl.jam, ''), DATE_FORMAT(lr.tgl_pemeriksaan, '%H:%i'), ''),
		COALESCE(hl.jamkeluar, ''),
		COALESCE(hl.sample, ''),
		COALESCE(NULLIF(hl.pesan, ''), lr.keterangan, ''),
		COALESCE(hl.kesan, ''),
		COALESCE(hl.saran, ''),
		COALESCE(ol.pemeriksaan, ''),
		COALESCE(ol.jenis, ''),
		COALESCE(ol.tipe_lab, ''),
		COALESCE(ol.diagnosa, ''),
		COALESCE(lr.json, '')
	FROM hasillabs hl
	INNER JOIN pasiens p ON p.id = hl.pasien_id
	LEFT JOIN lica_results lr ON lr.id = (
		SELECT lr2.id
		FROM lica_results lr2
		WHERE lr2.no_lab = hl.no_lab
		ORDER BY lr2.tgl_pemeriksaan DESC, lr2.id DESC
		LIMIT 1
	)
	LEFT JOIN registrasis r ON r.id = hl.registrasi_id
	LEFT JOIN order_lab ol ON ol.id = hl.order_lab_id
	LEFT JOIN polis po ON po.id = r.poli_id
	LEFT JOIN pegawais d ON d.id = hl.dokter_id
	LEFT JOIN pegawais rd ON rd.id = CAST(r.dokter_id AS UNSIGNED)
	WHERE p.id = ?
	AND hl.deleted_at IS NULL
	ORDER BY (lr.id IS NULL) ASC, COALESCE(lr.tgl_pemeriksaan, hl.tgl_hasilselesai, hl.tgl_pemeriksaan, hl.created_at) DESC, hl.id DESC
	LIMIT ?
	`

	rows, err := r.DB.QueryContext(ctx, query, profile.ID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]MobilePatientLabResult, 0)
	resultIndexByID := make(map[int64]int)
	labIDs := make([]int64, 0)

	for rows.Next() {
		var result MobilePatientLabResult
		var licaJSON string
		if err = rows.Scan(
			&result.ID,
			&result.NoLab,
			&result.Registration,
			&result.VisitDate,
			&result.Polyclinic,
			&result.Doctor,
			&result.Responsible,
			&result.ExamDate,
			&result.ReceivedDate,
			&result.ResultDate,
			&result.PrintDate,
			&result.StartTime,
			&result.FinishTime,
			&result.Sample,
			&result.Message,
			&result.Impression,
			&result.Suggestion,
			&result.OrderExaminations,
			&result.OrderType,
			&result.LabType,
			&result.Diagnosis,
			&licaJSON,
		); err != nil {
			return nil, err
		}
		result.Details = licaResultDetailsFromJSON(licaJSON)
		if len(result.Details) == 0 {
			result.Details = []MobilePatientLabResultDetail{}
			resultIndexByID[result.ID] = len(results)
			labIDs = append(labIDs, result.ID)
		}
		results = append(results, result)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	if len(labIDs) == 0 {
		return results, nil
	}

	details, err := r.labResultDetails(ctx, labIDs)
	if err != nil {
		return nil, err
	}

	for labID, detailItems := range details {
		index, ok := resultIndexByID[labID]
		if !ok {
			continue
		}
		results[index].Details = detailItems
	}

	return results, nil
}

func (r *MobilePatientRepository) labResultDetails(ctx context.Context, labIDs []int64) (map[int64][]MobilePatientLabResultDetail, error) {
	placeholders := sqlPlaceholders(len(labIDs))
	args := make([]any, 0, len(labIDs))
	for _, labID := range labIDs {
		args = append(args, labID)
	}

	query := `
	SELECT
		rh.hasillab_id,
		rh.id,
		COALESCE(ls.nama, ''),
		COALESCE(lk.nama, ''),
		COALESCE(lb.nama, ''),
		COALESCE(lb.rujukan, ''),
		COALESCE(lb.nilairujukanbawah, ''),
		COALESCE(lb.nilairujukanatas, ''),
		COALESCE(lb.satuan, ''),
		COALESCE(rh.hasiltext, ''),
		COALESCE(rh.hasil, '')
	FROM rincian_hasillabs rh
	LEFT JOIN labsections ls ON ls.id = rh.labsection_id
	LEFT JOIN labkategoris lk ON lk.id = rh.labkategori_id
	LEFT JOIN laboratoria lb ON lb.id = rh.laboratoria_id
	WHERE rh.hasillab_id IN (` + placeholders + `)
	ORDER BY rh.hasillab_id DESC, ls.nama ASC, lk.nama ASC, lb.nama ASC, rh.id ASC
	`

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	details := make(map[int64][]MobilePatientLabResultDetail)
	for rows.Next() {
		var (
			labID  int64
			detail MobilePatientLabResultDetail
		)
		if err = rows.Scan(
			&labID,
			&detail.ID,
			&detail.Section,
			&detail.Category,
			&detail.Examination,
			&detail.Reference,
			&detail.ReferenceLow,
			&detail.ReferenceHigh,
			&detail.Unit,
			&detail.ResultText,
			&detail.ResultValue,
		); err != nil {
			return nil, err
		}
		detail.Source = "rincian_hasillabs"
		details[labID] = append(details[labID], detail)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return details, nil
}

func licaResultDetailsFromJSON(raw string) []MobilePatientLabResultDetail {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	var payload []map[string]any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil
	}

	details := make([]MobilePatientLabResultDetail, 0, len(payload))
	for index, item := range payload {
		examination := jsonMapString(item, "test_name")
		result := jsonMapString(item, "result")
		if examination == "" && result == "" {
			continue
		}

		sequence := jsonMapInt64(item, "sequence")
		id := sequence
		if id <= 0 {
			id = int64(index + 1)
		}

		details = append(details, MobilePatientLabResultDetail{
			ID:           id,
			Source:       "lica_results",
			Code:         jsonMapString(item, "kode_jenis_tes"),
			Section:      jsonMapString(item, "group_test"),
			Category:     jsonMapString(item, "sub_group"),
			Examination:  examination,
			Reference:    jsonMapString(item, "nilai_normal"),
			Unit:         jsonMapString(item, "unit"),
			ResultText:   result,
			ResultValue:  result,
			Flag:         jsonMapString(item, "flag"),
			Note:         jsonMapString(item, "notes"),
			Sequence:     sequence,
			DrawTime:     jsonMapString(item, "draw_time"),
			ValidateTime: jsonMapString(item, "validate_time"),
		})
	}

	return details
}

func jsonMapString(item map[string]any, key string) string {
	value, ok := item[key]
	if !ok || value == nil {
		return ""
	}

	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case float64:
		return strings.TrimSpace(strconv.FormatFloat(typed, 'f', -1, 64))
	case bool:
		if typed {
			return "true"
		}
		return "false"
	default:
		return strings.TrimSpace(fmt.Sprint(typed))
	}
}

func jsonMapInt64(item map[string]any, key string) int64 {
	value := jsonMapString(item, key)
	if value == "" {
		return 0
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0
	}
	return parsed
}

func (r *MobilePatientRepository) RadiologyResults(ctx context.Context, userID int64, patientID int64, limit int) ([]MobilePatientRadiologyResult, error) {
	profile, err := r.Profile(ctx, userID, patientID)
	if err != nil {
		return nil, err
	}

	limit = normalizeMobileLimit(limit)

	query := `
	SELECT
		COALESCE(MAX(re.id), MAX(hr.id), MAX(ord.id), r.id) AS id,
		COALESCE(r.reg_id, ''),
		COALESCE(DATE_FORMAT(r.tgl_order, '%Y-%m-%d'), DATE_FORMAT(r.created_at, '%Y-%m-%d'), ''),
		COALESCE(po.nama, ''),
		COALESCE(MAX(dokter_ahli.nama), MAX(dokter_pengirim.nama), MAX(dokter_reg.nama), MAX(hr.dokter), ''),
		COALESCE(MAX(re.no_dokument), ''),
		COALESCE(MAX(re.tglPeriksa), ''),
		COALESCE(DATE_FORMAT(MAX(re.tanggal_eksp), '%Y-%m-%d'), ''),
		COALESCE(GROUP_CONCAT(DISTINCT ord.pemeriksaan SEPARATOR ', '), GROUP_CONCAT(DISTINCT hr.pemeriksaan SEPARATOR ', '), ''),
		COALESCE(GROUP_CONCAT(DISTINCT ord.jenis SEPARATOR ', '), ''),
		COALESCE(MAX(ord.status), ''),
		COALESCE(MAX(ord.source), ''),
		COALESCE(GROUP_CONCAT(DISTINCT re.klinis SEPARATOR '\n'), ''),
		COALESCE(GROUP_CONCAT(DISTINCT dr.resum SEPARATOR '\n'), ''),
		COALESCE(GROUP_CONCAT(DISTINCT re.ekspertise SEPARATOR '\n'), ''),
		COALESCE(DATE_FORMAT(MAX(ord.created_at), '%Y-%m-%d %H:%i'), ''),
		COALESCE(DATE_FORMAT(MAX(hr.created_at), '%Y-%m-%d %H:%i'), '')
	FROM registrasis r
	LEFT JOIN polis po ON po.id = r.poli_id
	LEFT JOIN pegawais dokter_reg ON dokter_reg.id = CAST(r.dokter_id AS UNSIGNED)
	LEFT JOIN order_radiologi ord ON ord.registrasi_id = r.id
	LEFT JOIN hasilradiologis hr ON hr.registrasi_id = r.id
	LEFT JOIN detailradiologis dr ON dr.hasilradiologi_id = hr.id
	LEFT JOIN radiologi_ekspertises re ON re.registrasi_id = r.id
		AND (re.pasien_id = ? OR re.pasien_id IS NULL OR re.pasien_id = 0)
	LEFT JOIN pegawais dokter_ahli ON dokter_ahli.id = re.dokter_id
	LEFT JOIN pegawais dokter_pengirim ON dokter_pengirim.id = re.dokter_pengirim
	WHERE r.pasien_id = ?
	AND (ord.id IS NOT NULL OR hr.id IS NOT NULL OR re.id IS NOT NULL)
	GROUP BY r.id, r.reg_id, r.tgl_order, r.created_at, po.nama
	ORDER BY COALESCE(MAX(re.tanggal_eksp), MAX(ord.created_at), MAX(hr.created_at), r.created_at) DESC, r.id DESC
	LIMIT ?
	`

	rows, err := r.DB.QueryContext(ctx, query, profile.ID, profile.ID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]MobilePatientRadiologyResult, 0)
	for rows.Next() {
		var result MobilePatientRadiologyResult
		if err = rows.Scan(
			&result.ID,
			&result.Registration,
			&result.VisitDate,
			&result.Polyclinic,
			&result.Doctor,
			&result.DocumentNumber,
			&result.ExamDate,
			&result.ResultDate,
			&result.Examination,
			&result.OrderType,
			&result.Status,
			&result.Source,
			&result.ClinicalNote,
			&result.ResultSummary,
			&result.Expertise,
			&result.OrderCreatedAt,
			&result.ResultCreatedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (r *MobilePatientRepository) Prescriptions(ctx context.Context, userID int64, patientID int64, limit int) ([]MobilePatientPrescription, error) {
	profile, err := r.Profile(ctx, userID, patientID)
	if err != nil {
		return nil, err
	}

	limit = normalizeMobileLimit(limit)

	query := `
	SELECT
		pj.id,
		COALESCE(pj.no_resep, ''),
		COALESCE(r.reg_id, ''),
		COALESCE(DATE_FORMAT(r.tgl_order, '%Y-%m-%d'), DATE_FORMAT(r.created_at, '%Y-%m-%d'), ''),
		COALESCE(po.nama, ''),
		COALESCE(d.nama, rd.nama, ''),
		COALESCE(DATE_FORMAT(pj.created_at, '%Y-%m-%d %H:%i'), ''),
		COALESCE(pj.done_input, ''),
		COALESCE(pj.catatan, '')
	FROM penjualans pj
	INNER JOIN registrasis r ON r.id = pj.registrasi_id
	LEFT JOIN polis po ON po.id = r.poli_id
	LEFT JOIN pegawais d ON d.id = pj.dokter_id
	LEFT JOIN pegawais rd ON rd.id = CAST(r.dokter_id AS UNSIGNED)
	WHERE r.pasien_id = ?
	AND pj.deleted_at IS NULL
	ORDER BY pj.created_at DESC, pj.id DESC
	LIMIT ?
	`

	rows, err := r.DB.QueryContext(ctx, query, profile.ID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]MobilePatientPrescription, 0)
	resultIndexByID := make(map[int64]int)
	prescriptionIDs := make([]int64, 0)

	for rows.Next() {
		var prescription MobilePatientPrescription
		if err = rows.Scan(
			&prescription.ID,
			&prescription.NoResep,
			&prescription.Registration,
			&prescription.VisitDate,
			&prescription.Polyclinic,
			&prescription.Doctor,
			&prescription.CreatedAt,
			&prescription.Status,
			&prescription.Note,
		); err != nil {
			return nil, err
		}
		prescription.Note = normalizePatientVisibleNote(prescription.Note)
		prescription.Details = []MobilePatientPrescriptionDetail{}
		resultIndexByID[prescription.ID] = len(results)
		prescriptionIDs = append(prescriptionIDs, prescription.ID)
		results = append(results, prescription)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	if len(prescriptionIDs) == 0 {
		return results, nil
	}

	details, err := r.prescriptionDetails(ctx, prescriptionIDs)
	if err != nil {
		return nil, err
	}

	for prescriptionID, detailItems := range details {
		index, ok := resultIndexByID[prescriptionID]
		if !ok {
			continue
		}
		results[index].Details = detailItems
	}

	return results, nil
}

func (r *MobilePatientRepository) prescriptionDetails(ctx context.Context, prescriptionIDs []int64) (map[int64][]MobilePatientPrescriptionDetail, error) {
	placeholders := sqlPlaceholders(len(prescriptionIDs))
	args := make([]any, 0, len(prescriptionIDs))
	for _, prescriptionID := range prescriptionIDs {
		args = append(args, prescriptionID)
	}

	query := `
	SELECT
		pd.penjualan_id,
		pd.id,
		COALESCE(mo.nama, ''),
		COALESCE(mo.kode, ''),
		COALESCE(mo.satuan_obat, ''),
		COALESCE(pd.jumlah, 0),
		COALESCE(pd.cara_minum, ''),
		COALESCE(pd.takaran, ''),
		COALESCE(pd.informasi1, ''),
		COALESCE(pd.informasi2, ''),
		COALESCE(pd.catatan, ''),
		COALESCE(pd.etiket, ''),
		COALESCE(pd.obat_racikan, ''),
		COALESCE(pd.is_kronis, '')
	FROM penjualandetails pd
	LEFT JOIN masterobats mo ON mo.id = pd.masterobat_id
	WHERE pd.penjualan_id IN (` + placeholders + `)
	AND pd.deleted_at IS NULL
	ORDER BY pd.penjualan_id DESC, pd.id ASC
	`

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	details := make(map[int64][]MobilePatientPrescriptionDetail)
	for rows.Next() {
		var (
			prescriptionID int64
			detail         MobilePatientPrescriptionDetail
		)
		if err = rows.Scan(
			&prescriptionID,
			&detail.ID,
			&detail.DrugName,
			&detail.DrugCode,
			&detail.Unit,
			&detail.Quantity,
			&detail.HowToUse,
			&detail.Dose,
			&detail.InfoPrimary,
			&detail.InfoExtra,
			&detail.Note,
			&detail.Label,
			&detail.Compound,
			&detail.Chronic,
		); err != nil {
			return nil, err
		}
		detail.Note = normalizePatientVisibleNote(detail.Note)
		details[prescriptionID] = append(details[prescriptionID], detail)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return details, nil
}

func normalizeMobileLimit(limit int) int {
	if limit < 1 {
		return 20
	}
	if limit > 100 {
		return 100
	}
	return limit
}

func (r *MobilePatientRepository) icd10Texts(ctx context.Context, registrationID int64) (string, string, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT
			COALESCE(NULLIF(TRIM(px.icd10), ''), ''),
			COALESCE(NULLIF(TRIM(COALESCE(NULLIF(NULLIF(TRIM(px.diagnosis), ''), TRIM(px.icd10)), ic.nama, '')), ''), ''),
			COALESCE(px.kategori, ''),
			COALESCE(px.jenis, '')
		FROM (
			SELECT icd10, diagnosis, kategori, jenis
			FROM perawatan_icd10s
			WHERE registrasi_id = ?
			UNION ALL
			SELECT icd10, diagnosis, '' AS kategori, jenis
			FROM jkn_icd10s
			WHERE registrasi_id = ?
		) px
		LEFT JOIN icd10s ic ON ic.nomor = px.icd10
	`, registrationID, registrationID)
	if err != nil {
		return "", "", err
	}
	defer rows.Close()

	mainItems := make([]string, 0)
	additionalItems := make([]string, 0)
	seenMain := map[string]bool{}
	seenAdditional := map[string]bool{}
	for rows.Next() {
		var code, name, category, kind string
		if err = rows.Scan(&code, &name, &category, &kind); err != nil {
			return "", "", err
		}
		item := strings.TrimSpace(strings.TrimSpace(code + " " + name))
		if item == "" {
			continue
		}
		targetAdditional := strings.Contains(strings.ToLower(category), "tambahan") || strings.TrimSpace(kind) == "2"
		if targetAdditional {
			if !seenAdditional[item] {
				additionalItems = append(additionalItems, item)
				seenAdditional[item] = true
			}
			continue
		}
		if !seenMain[item] {
			mainItems = append(mainItems, item)
			seenMain[item] = true
		}
	}
	if err = rows.Err(); err != nil {
		return "", "", err
	}

	if len(mainItems) == 0 && len(additionalItems) > 0 {
		mainItems = append(mainItems, additionalItems[0])
		additionalItems = additionalItems[1:]
	}

	return strings.Join(mainItems, "\n"), strings.Join(additionalItems, "\n"), nil
}

func (r *MobilePatientRepository) icd9Text(ctx context.Context, registrationID int64) (string, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT
			COALESCE(NULLIF(TRIM(px.icd9), ''), ''),
			COALESCE(NULLIF(TRIM(COALESCE(NULLIF(NULLIF(TRIM(px.diagnosis), ''), TRIM(px.icd9)), ic.nama, '')), ''), '')
		FROM (
			SELECT icd9, diagnosis
			FROM perawatan_icd9s
			WHERE registrasi_id = ?
			UNION ALL
			SELECT icd9, diagnosis
			FROM jkn_icd9s
			WHERE registrasi_id = ?
		) px
		LEFT JOIN icd9s ic ON ic.nomor = px.icd9
	`, registrationID, registrationID)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	items := make([]string, 0)
	seen := map[string]bool{}
	for rows.Next() {
		var code, name string
		if err = rows.Scan(&code, &name); err != nil {
			return "", err
		}
		item := strings.TrimSpace(strings.TrimSpace(code + " " + name))
		if item == "" || seen[item] {
			continue
		}
		items = append(items, item)
		seen[item] = true
	}
	if err = rows.Err(); err != nil {
		return "", err
	}

	return strings.Join(items, "\n"), nil
}

func resumeJSONValue(content map[string]any, key string) string {
	value, ok := content[key]
	if !ok || value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return normalizePatientVisibleNote(typed)
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case bool:
		if typed {
			return "Ya"
		}
		return "Tidak"
	default:
		return normalizePatientVisibleNote(fmt.Sprint(typed))
	}
}

func extractResumeJSONText(raw string, key string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	payload := map[string]any{}
	if json.Unmarshal([]byte(raw), &payload) != nil {
		return ""
	}

	return resumeJSONValue(payload, key)
}

func coalescePatientString(values ...string) string {
	for _, value := range values {
		value = normalizePatientVisibleNote(value)
		if value != "" && value != "-" {
			return value
		}
	}
	return ""
}

func joinPatientLines(values ...string) string {
	lines := make([]string, 0, len(values))
	for _, value := range values {
		value = normalizePatientVisibleNote(value)
		if value != "" && value != "-" {
			lines = append(lines, value)
		}
	}
	return strings.Join(lines, "\n")
}

func sqlPlaceholders(count int) string {
	if count <= 0 {
		return ""
	}

	return strings.TrimSuffix(strings.Repeat("?,", count), ",")
}

func normalizePatientVisibleNote(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	if isJSONPayload(trimmed) {
		return ""
	}

	return trimmed
}

func isJSONPayload(value string) bool {
	if len(value) < 2 {
		return false
	}

	start := value[0]
	end := value[len(value)-1]
	if !((start == '{' && end == '}') || (start == '[' && end == ']')) {
		return false
	}

	var payload any
	return json.Unmarshal([]byte(value), &payload) == nil
}

func normalizeGender(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case "l", "1", "m", "male", "laki", "laki-laki", "laki laki":
		return "Laki-laki"
	case "p", "2", "f", "female", "perempuan":
		return "Perempuan"
	default:
		return strings.TrimSpace(value)
	}
}

func normalizePatientType(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case "umum", "u", "2":
		return "Umum"
	case "jkn", "bpjs", "1":
		return "JKN/BPJS"
	default:
		return strings.TrimSpace(value)
	}
}

func normalizePaymentType(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case "1", "bpjs", "jkn":
		return "JKN/BPJS"
	case "2", "umum":
		return "Umum"
	default:
		return strings.TrimSpace(value)
	}
}
