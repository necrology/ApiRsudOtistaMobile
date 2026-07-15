package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"apirusdotistamobile/internal/model"
)

type BookingRepository struct {
	DB       *sql.DB
	Holidays *HolidayRepository
}

type BookingValidationError struct {
	Message string
}

func (e *BookingValidationError) Error() string {
	return e.Message
}

func bookingValidationError(message string) error {
	return &BookingValidationError{Message: message}
}

type BookingTarget struct {
	PatientID   int64
	NoRM        string
	Email       string
	PoliID      int64
	Tanggal     time.Time
	Bayar       string
	JenisPasien string
	DoctorID    string
	QueueGroup  string
	IsJkn       bool
}

type BookingListFilter struct {
	Status    string
	Date      *time.Time
	Limit     int
	patientID int64
}

const (
	generalDummyFlag             = "mobile_umum"
	generalDummyJenisDaftar      = "android"
	generalDummyJenisRegistrasi  = "antrian"
	generalDummyDefaultStatus    = "pending"
	generalDummyDefaultKunjungan = "1"
)

type bookingContext struct {
	poliID       int64
	poliName     string
	poliBPJS     string
	poliCode     string
	doctorID     string
	doctorName   string
	doctorCode   string
	queueGroup   string
	practiceTime string
	queueDate    time.Time
	serviceMode  string
}

type bookingPatient struct {
	ID         int64
	NoRM       string
	NIK        string
	Name       string
	Phone      string
	Gender     string
	BirthPlace string
	BirthDate  string
	Address    string
}

func NewBookingRepository(db *sql.DB, holidays *HolidayRepository) *BookingRepository {
	return &BookingRepository{DB: db, Holidays: holidays}
}

func (r *BookingRepository) BookingOptions(ctx context.Context, poliID int64) (*model.BookingOptionsResponse, error) {
	if poliID <= 0 {
		return nil, bookingValidationError("poli tidak valid")
	}

	var poli model.BookingPoliOption
	if err := r.DB.QueryRowContext(ctx, `
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
			COALESCE(p.terisi, 0),
			COALESCE(p.kelompok, ''),
			COALESCE(p.bpjs, '')
		FROM polis p
		WHERE p.id = ?
		LIMIT 1
	`, poliID).Scan(
		&poli.ID,
		&poli.Nama,
		&poli.Kelompok,
		&poli.Politype,
		&poli.BpjsCode,
		&poli.KodeRuangan,
		&poli.Buka,
		&poli.Tutup,
		&poli.Praktik,
		&poli.Kuota,
		&poli.KuotaOnline,
		&poli.Terisi,
		&poli.QueueGroup,
		&poli.QueueGroupHint,
	); err != nil {
		return nil, err
	}

	doctorIDs := splitDoctorIDs(poliID, r.DB)
	if len(doctorIDs) > 0 {
		placeholders := strings.TrimSuffix(strings.Repeat("?,", len(doctorIDs)), ",")
		args := make([]any, 0, len(doctorIDs))
		for _, id := range doctorIDs {
			args = append(args, id)
		}
		query := fmt.Sprintf(`
			SELECT id, COALESCE(nama, ''), COALESCE(kode_antrian, ''), COALESCE(kode_bpjs, ''), COALESCE(general_code, '')
			FROM pegawais
			WHERE id IN (%s)
			ORDER BY nama ASC
		`, placeholders)
		rows, err := r.DB.QueryContext(ctx, query, args...)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var doctor model.BookingDoctorOption
				if err = rows.Scan(&doctor.ID, &doctor.Nama, &doctor.KodeAntrian, &doctor.KodeBpjs, &doctor.GeneralCode); err != nil {
					return nil, err
				}
				poli.Doctors = append(poli.Doctors, doctor)
			}
			if err = rows.Err(); err != nil {
				return nil, err
			}
		}
	}

	if poli.QueueGroup == "" {
		poli.QueueGroup = poli.Kelompok
	}
	if poli.QueueGroupHint == "" {
		poli.QueueGroupHint = poli.Politype
	}

	schedules, schedulesByDoctor, err := r.loadUnifiedBookingSchedules(ctx, poli, poli.Doctors)
	if err != nil {
		return nil, err
	}
	poli.Schedules = schedules
	for i := range poli.Doctors {
		if doctorSchedules := schedulesByDoctor[poli.Doctors[i].ID]; len(doctorSchedules) > 0 {
			poli.Doctors[i].Schedules = doctorSchedules
		}
	}

	return &model.BookingOptionsResponse{Poli: poli}, nil
}

func splitDoctorIDs(poliID int64, db *sql.DB) []int64 {
	var raw sql.NullString
	if err := db.QueryRow(`SELECT COALESCE(dokter_id, '') FROM polis WHERE id = ? LIMIT 1`, poliID).Scan(&raw); err != nil || !raw.Valid {
		return nil
	}

	parts := strings.Split(raw.String, ",")
	ids := make([]int64, 0, len(parts))
	for _, part := range parts {
		id, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64)
		if err == nil && id > 0 {
			ids = append(ids, id)
		}
	}
	return ids
}

func (r *BookingRepository) CreateGeneralBooking(ctx context.Context, target BookingTarget) (*model.BookingQueueResponse, error) {
	ctxData, err := r.loadBookingContext(ctx, target)
	if err != nil {
		return nil, err
	}
	if r.Holidays != nil {
		if err = r.Holidays.ValidateBookingDate(ctx, ctxData.queueDate, ctxData.poliID); err != nil {
			return nil, err
		}
	}

	return r.createBookingTx(ctx, ctxData, target)
}

func (r *BookingRepository) loadBookingContext(ctx context.Context, target BookingTarget) (*bookingContext, error) {
	if target.PatientID <= 0 {
		return nil, bookingValidationError("pasien tidak valid")
	}
	if target.PoliID <= 0 {
		return nil, bookingValidationError("poli tidak valid")
	}

	var (
		poliName  string
		poliBPJS  string
		poliType  sql.NullString
		groupName sql.NullString
		doctorID  sql.NullString
		buka      string
		tutup     string
	)

	if err := r.DB.QueryRowContext(ctx, `
		SELECT
			COALESCE(nama, ''),
			COALESCE(bpjs, ''),
			COALESCE(politype, ''),
			COALESCE(kelompok, ''),
			COALESCE(dokter_id, ''),
			COALESCE(DATE_FORMAT(buka, '%H:%i'), ''),
			COALESCE(DATE_FORMAT(tutup, '%H:%i'), '')
		FROM polis
		WHERE id = ?
		LIMIT 1
	`, target.PoliID).Scan(&poliName, &poliBPJS, &poliType, &groupName, &doctorID, &buka, &tutup); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, bookingValidationError("poli tidak valid")
		}
		return nil, err
	}

	queueDate := target.Tanggal
	if queueDate.IsZero() {
		queueDate = time.Now()
	}
	queueDate = time.Date(queueDate.Year(), queueDate.Month(), queueDate.Day(), 0, 0, 0, 0, queueDate.Location())

	serviceMode := "umum"

	queueGroupValue := coalesceBookingString(target.QueueGroup, groupName.String, firstCSVValue(doctorID.String), poliType.String)
	if queueGroupValue == "" {
		queueGroupValue = poliName
	}

	selectedDoctorID := coalesceBookingString(target.DoctorID, firstCSVValue(doctorID.String))
	doctorName := ""
	doctorCode := selectedDoctorID
	if selectedDoctorID != "" {
		var (
			dbDoctorID   string
			dbDoctorName string
			kodeAntrian  string
			kodeBPJS     string
			kodeGeneral  string
		)
		err := r.DB.QueryRowContext(ctx, `
			SELECT
				COALESCE(CAST(id AS CHAR), ''),
				COALESCE(nama, ''),
				COALESCE(kode_antrian, ''),
				COALESCE(kode_bpjs, ''),
				COALESCE(general_code, '')
			FROM pegawais
			WHERE id = ?
			LIMIT 1
		`, selectedDoctorID).Scan(&dbDoctorID, &dbDoctorName, &kodeAntrian, &kodeBPJS, &kodeGeneral)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		selectedDoctorID = coalesceBookingString(dbDoctorID, selectedDoctorID)
		doctorName = dbDoctorName
		doctorCode = coalesceBookingString(kodeGeneral, kodeBPJS, kodeAntrian, selectedDoctorID)
	}

	poliCode := coalesceBookingString(poliBPJS, strconv.FormatInt(target.PoliID, 10))
	practiceTime := ""
	if buka != "" || tutup != "" {
		practiceTime = strings.Trim(strings.Join([]string{buka, tutup}, "-"), "-")
	}

	return &bookingContext{
		poliID:       target.PoliID,
		poliName:     poliName,
		poliBPJS:     poliBPJS,
		poliCode:     poliCode,
		doctorID:     selectedDoctorID,
		doctorName:   doctorName,
		doctorCode:   doctorCode,
		queueGroup:   queueGroupValue,
		practiceTime: practiceTime,
		queueDate:    queueDate,
		serviceMode:  serviceMode,
	}, nil
}

func (r *BookingRepository) createBookingTx(ctx context.Context, ctxData *bookingContext, target BookingTarget) (ret *model.BookingQueueResponse, err error) {
	conn, err := r.DB.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	tx, err := conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	patient, err := r.loadBookingPatientTx(ctx, tx, target.PatientID)
	if err != nil {
		return nil, err
	}
	if target.NoRM != "" && patient.NoRM != "" && target.NoRM != patient.NoRM {
		return nil, bookingValidationError("no rekam medis tidak sesuai dengan akun pasien")
	}

	if existing, existingErr := r.findExistingGeneralDummyBookingTx(ctx, tx, patient.NoRM, ctxData.queueDate); existingErr == nil && existing != nil {
		existing.Existing = true
		if err = tx.Commit(); err != nil {
			return nil, err
		}
		return existing, nil
	} else if existingErr != nil && !errors.Is(existingErr, sql.ErrNoRows) {
		return nil, existingErr
	}

	if existing, existingErr := r.findExistingGeneralBookingTx(ctx, tx, target.PatientID, ctxData.queueDate); existingErr == nil && existing != nil {
		existing.Existing = true
		if err = tx.Commit(); err != nil {
			return nil, err
		}
		return existing, nil
	} else if existingErr != nil && !errors.Is(existingErr, sql.ErrNoRows) {
		return nil, existingErr
	}

	lockName := fmt.Sprintf("mobile_booking_queue:%s:%s:%s", generalDummyFlag, ctxData.poliCode, ctxData.queueDate.Format("2006-01-02"))
	if err = acquireMySQLLock(ctx, tx, lockName); err != nil {
		return nil, err
	}
	defer func() {
		_, _ = conn.ExecContext(ctx, `SELECT RELEASE_LOCK(?)`, lockName)
	}()

	if existing, existingErr := r.findExistingGeneralDummyBookingTx(ctx, tx, patient.NoRM, ctxData.queueDate); existingErr == nil && existing != nil {
		existing.Existing = true
		if err = tx.Commit(); err != nil {
			return nil, err
		}
		return existing, nil
	} else if existingErr != nil && !errors.Is(existingErr, sql.ErrNoRows) {
		return nil, existingErr
	}

	if existing, existingErr := r.findExistingGeneralBookingTx(ctx, tx, target.PatientID, ctxData.queueDate); existingErr == nil && existing != nil {
		existing.Existing = true
		if err = tx.Commit(); err != nil {
			return nil, err
		}
		return existing, nil
	} else if existingErr != nil && !errors.Is(existingErr, sql.ErrNoRows) {
		return nil, existingErr
	}

	nextQueueNumber, err := r.nextDummyQueueNumberTx(ctx, tx, ctxData)
	if err != nil {
		return nil, err
	}

	regID := generateGeneralDummyQueueNumber(ctxData.queueDate, ctxData.poliCode, nextQueueNumber)
	queueNumber := strconv.Itoa(nextQueueNumber)
	queueCode := formatGeneralQueueCode(nextQueueNumber)
	queueDateString := ctxData.queueDate.Format("2006-01-02")
	jenisPasien := "umum"
	bayar := coalesceBookingString(target.Bayar, "2")
	requestPayload, err := generalDummyRequestPayload(ctxData, patient, target, regID, queueCode, queueNumber, bayar, jenisPasien)
	if err != nil {
		return nil, err
	}

	dummyResult, err := tx.ExecContext(ctx, `
		INSERT INTO registrasis_dummy (
			jenis_registrasi,
			nomorantrian,
			kodebooking,
			angkaantrian,
			nik,
			no_rm,
			no_hp,
			tglperiksa,
			kode_poli,
			jenisreferensi,
			jenisrequest,
			polieksekutif,
			jenisdaftar,
			kelamin,
			tmplahir,
			tgllahir,
			kode_cara_bayar,
			kode_dokter,
			dokter_id,
			nama,
			alamat,
			status,
			jampraktek,
			jeniskunjungan,
			keterangan,
			request,
			flag,
			created_at,
			updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`,
		generalDummyJenisRegistrasi,
		regID,
		nil,
		nextQueueNumber,
		patient.NIK,
		patient.NoRM,
		patient.Phone,
		queueDateString,
		ctxData.poliCode,
		0,
		nil,
		0,
		generalDummyJenisDaftar,
		nullableString(normalizeDummyGender(patient.Gender)),
		patient.BirthPlace,
		patient.BirthDate,
		bayar,
		coalesceBookingString(ctxData.doctorID, ctxData.doctorCode),
		nil,
		patient.Name,
		patient.Address,
		generalDummyDefaultStatus,
		ctxData.practiceTime,
		generalDummyDefaultKunjungan,
		"Booking antrian umum mobile",
		requestPayload,
		generalDummyFlag,
	)
	if err != nil {
		return nil, err
	}

	dummyID, err := dummyResult.LastInsertId()
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &model.BookingQueueResponse{
		RegistrationID:   0,
		DummyID:          dummyID,
		QueueID:          0,
		RegistrationCode: regID,
		QueueNumber:      queueNumber,
		QueueCode:        queueCode,
		QueueGroup:       coalesceBookingString(ctxData.queueGroup, ctxData.poliName),
		PoliID:           ctxData.poliID,
		PoliName:         ctxData.poliName,
		QueueDate:        queueDateString,
		ServiceMode:      ctxData.serviceMode,
		Source:           generalDummyFlag,
	}, nil
}

func (r *BookingRepository) ListPatientGeneralBookings(
	ctx context.Context,
	patientID int64,
	filter BookingListFilter,
) (*model.MobileBookingListResponse, error) {
	if patientID <= 0 {
		return nil, bookingValidationError("pasien tidak valid")
	}
	filter.patientID = patientID
	return r.listGeneralBookings(ctx, filter)
}

func (r *BookingRepository) listGeneralBookings(ctx context.Context, filter BookingListFilter) (*model.MobileBookingListResponse, error) {
	limit := normalizeBookingLimit(filter.Limit)
	args := []any{}
	conditions := []string{
		"r.deleted_at IS NULL",
		"r.input_from = 'mobile'",
		"(r.jkn IS NULL OR r.jkn <> 'Y')",
		"NOT EXISTS (SELECT 1 FROM registrasis_dummy rd WHERE rd.registrasi_id = r.id AND rd.flag = 'mobile_umum')",
	}

	if filter.patientID > 0 {
		conditions = append(conditions, "r.pasien_id = ?")
		args = append(args, filter.patientID)
	}

	if filter.Date != nil {
		conditions = append(conditions, "r.tgl_order = ?")
		args = append(args, filter.Date.Format("2006-01-02"))
	}

	status := strings.ToLower(strings.TrimSpace(filter.Status))
	if status != "" && status != "semua" && status != "all" {
		conditions = append(conditions, bookingStatusCondition())
		args = append(args, status, status, status, status, status, status, status, status, status)
	}

	query := fmt.Sprintf(`
	SELECT
		r.id,
		COALESCE(a.id, 0),
		COALESCE(r.reg_id, ''),
		COALESCE(NULLIF(r.nomorantrian, ''), COALESCE(CAST(a.nomor AS CHAR), '')),
		COALESCE(a.nomor, 0),
		COALESCE(a.kelompok, ''),
		COALESCE(DATE_FORMAT(r.tgl_order, '%%Y-%%m-%%d'), DATE_FORMAT(a.tanggal, '%%Y-%%m-%%d'), DATE_FORMAT(r.created_at, '%%Y-%%m-%%d'), ''),
		COALESCE(DATE_FORMAT(r.created_at, '%%Y-%%m-%%d %%H:%%i'), ''),
		COALESCE(r.status, ''),
		COALESCE(r.status_reg, ''),
		COALESCE(a.status, ''),
		COALESCE(a.panggil, ''),
		COALESCE(a.sudah_dipanggil, 0),
		p.id,
		COALESCE(p.no_rm, ''),
		COALESCE(p.nama, ''),
		COALESCE(po.id, 0),
		COALESCE(po.nama, ''),
		COALESCE(d.id, 0),
		COALESCE(d.nama, ''),
		COALESCE(r.input_from, ''),
		COALESCE(r.jenis_pasien, ''),
		COALESCE(r.bayar, ''),
		COALESCE(r.keterangan, '')
	FROM registrasis r
	INNER JOIN pasiens p ON p.id = r.pasien_id
	LEFT JOIN antrian_poli a ON a.id = r.antrian_poli_id
	LEFT JOIN polis po ON po.id = r.poli_id
	LEFT JOIN pegawais d ON d.id = CAST(r.dokter_id AS UNSIGNED)
	WHERE %s
	ORDER BY r.tgl_order DESC, r.id DESC
	LIMIT ?
	`, strings.Join(conditions, "\n\tAND "))
	args = append(args, limit)

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.MobileBookingRecord, 0)
	counts := map[string]int{}

	for rows.Next() {
		var record model.MobileBookingRecord
		var numericQueue int64
		var alreadyCalled int
		if err = rows.Scan(
			&record.RegistrationID,
			&record.QueueID,
			&record.RegistrationCode,
			&record.QueueNumber,
			&numericQueue,
			&record.QueueGroup,
			&record.QueueDate,
			&record.CreatedAt,
			&record.RegistrationStatus,
			&record.StatusReg,
			&record.QueueStatus,
			&record.CallStatus,
			&alreadyCalled,
			&record.PatientID,
			&record.NoRM,
			&record.PatientName,
			&record.PoliID,
			&record.PoliName,
			&record.DoctorID,
			&record.DoctorName,
			&record.InputFrom,
			&record.PatientType,
			&record.PaymentType,
			&record.Note,
		); err != nil {
			return nil, err
		}

		record.AlreadyCalled = alreadyCalled > 0
		if record.QueueNumber == "" && numericQueue > 0 {
			record.QueueNumber = strconv.FormatInt(numericQueue, 10)
		}
		if numericQueue > 0 {
			record.QueueCode = formatGeneralQueueCode(int(numericQueue))
		} else {
			record.QueueCode = record.QueueNumber
		}
		record.QueueStatusLabel = bookingStatusLabel(record.QueueStatus, record.CallStatus, record.AlreadyCalled, record.StatusReg)
		counts[record.QueueStatusLabel]++
		items = append(items, record)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	if err = rows.Close(); err != nil {
		return nil, err
	}

	if len(items) < limit {
		if err = r.appendGeneralDummyBookings(ctx, filter, limit-len(items), &items, counts); err != nil {
			return nil, err
		}
	}

	return &model.MobileBookingListResponse{
		Items:        items,
		StatusCounts: bookingStatusCounts(counts),
	}, nil
}

func (r *BookingRepository) appendGeneralDummyBookings(ctx context.Context, filter BookingListFilter, limit int, items *[]model.MobileBookingRecord, counts map[string]int) error {
	if limit <= 0 {
		return nil
	}

	args := []any{generalDummyFlag}
	conditions := []string{
		"rd.flag = ?",
	}

	if filter.patientID > 0 {
		conditions = append(conditions, "p.id = ?")
		args = append(args, filter.patientID)
	}

	if filter.Date != nil {
		conditions = append(conditions, "rd.tglperiksa = ?")
		args = append(args, filter.Date.Format("2006-01-02"))
	}

	status := strings.ToLower(strings.TrimSpace(filter.Status))
	if status != "" && status != "semua" && status != "all" {
		conditions = append(conditions, dummyBookingStatusCondition())
		args = append(args, status, status, status, status, status, status)
	}

	query := fmt.Sprintf(`
	SELECT
		rd.id,
		COALESCE(rd.registrasi_id, 0),
		COALESCE(NULLIF(rd.kodebooking, ''), COALESCE(rd.nomorantrian, '')),
		COALESCE(rd.nomorantrian, ''),
		COALESCE(rd.angkaantrian, 0),
		COALESCE(po_exact.nama, COALESCE(po_code.nama, '')),
		COALESCE(rd.tglperiksa, ''),
		COALESCE(DATE_FORMAT(rd.created_at, '%%Y-%%m-%%d %%H:%%i'), ''),
		COALESCE(rd.status, ''),
		COALESCE(p.id, 0),
		COALESCE(rd.no_rm, COALESCE(p.no_rm, '')),
		COALESCE(rd.nama, COALESCE(p.nama, '')),
		COALESCE(po_exact.id, COALESCE(po_code.id, 0)),
		COALESCE(po_exact.nama, COALESCE(po_code.nama, '')),
		COALESCE(d.id, 0),
		COALESCE(d.nama, ''),
		COALESCE(rd.flag, ''),
		COALESCE(rd.kode_cara_bayar, ''),
		COALESCE(rd.keterangan, '')
	FROM registrasis_dummy rd
	LEFT JOIN pasiens p ON p.no_rm = rd.no_rm
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
	LEFT JOIN pegawais d ON d.id = COALESCE(rd.dokter_id, CAST(NULLIF(rd.kode_dokter, '') AS UNSIGNED))
	WHERE %s
	ORDER BY rd.tglperiksa DESC, rd.id DESC
	LIMIT ?
	`, strings.Join(conditions, "\n\tAND "))
	args = append(args, limit)

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var record model.MobileBookingRecord
		var numericQueue int64
		if err = rows.Scan(
			&record.DummyID,
			&record.RegistrationID,
			&record.RegistrationCode,
			&record.QueueNumber,
			&numericQueue,
			&record.QueueGroup,
			&record.QueueDate,
			&record.CreatedAt,
			&record.RegistrationStatus,
			&record.PatientID,
			&record.NoRM,
			&record.PatientName,
			&record.PoliID,
			&record.PoliName,
			&record.DoctorID,
			&record.DoctorName,
			&record.InputFrom,
			&record.PaymentType,
			&record.Note,
		); err != nil {
			return err
		}

		record.Source = generalDummyFlag
		record.QueueID = 0
		record.StatusReg = record.RegistrationStatus
		record.QueueStatus = record.RegistrationStatus
		record.PatientType = "umum"
		if record.QueueGroup == "" {
			record.QueueGroup = record.PoliName
		}
		if record.QueueNumber == "" && numericQueue > 0 {
			record.QueueNumber = strconv.FormatInt(numericQueue, 10)
		}
		if record.QueueNumber != "" && strings.Contains(record.QueueNumber, "-") {
			record.QueueCode = record.QueueNumber
		} else if numericQueue > 0 {
			record.QueueCode = formatGeneralQueueCode(int(numericQueue))
		} else {
			record.QueueCode = record.QueueNumber
		}
		record.QueueStatusLabel = dummyBookingStatusLabel(record.RegistrationStatus)
		counts[record.QueueStatusLabel]++
		*items = append(*items, record)
	}

	return rows.Err()
}

func (r *BookingRepository) nextQueueNumberTx(ctx context.Context, tx *sql.Tx, poliID int64, queueDate time.Time) (int, error) {
	var next sql.NullInt64
	if err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(nomor), 0) + 1
		FROM antrian_poli
		WHERE poli_id = ?
		AND tanggal = ?
	`, poliID, queueDate.Format("2006-01-02")).Scan(&next); err != nil {
		return 0, err
	}

	if next.Valid && next.Int64 > 0 {
		return int(next.Int64), nil
	}

	return 1, nil
}

func (r *BookingRepository) nextDummyQueueNumberTx(ctx context.Context, tx *sql.Tx, ctxData *bookingContext) (int, error) {
	var next sql.NullInt64
	if err := tx.QueryRowContext(ctx, `
		SELECT GREATEST(
			COALESCE((
				SELECT MAX(ap.nomor)
				FROM antrian_poli ap
				WHERE ap.poli_id = ?
				AND ap.tanggal = ?
			), 0),
			COALESCE((
			SELECT MAX(rd.angkaantrian)
			FROM registrasis_dummy rd
			WHERE rd.kode_poli = ?
			AND rd.tglperiksa = ?
			AND LOWER(COALESCE(rd.status, '')) <> 'dibatalkan'
		), 0)
	) + 1
	`, ctxData.poliID, ctxData.queueDate.Format("2006-01-02"), ctxData.poliCode, ctxData.queueDate.Format("2006-01-02")).Scan(&next); err != nil {
		return 0, err
	}

	if next.Valid && next.Int64 > 0 {
		return int(next.Int64), nil
	}

	return 1, nil
}

func (r *BookingRepository) loadBookingPatientTx(ctx context.Context, tx *sql.Tx, patientID int64) (*bookingPatient, error) {
	var patient bookingPatient
	if err := tx.QueryRowContext(ctx, `
		SELECT
			id,
			COALESCE(no_rm, ''),
			COALESCE(nik, ''),
			COALESCE(nama, ''),
			COALESCE(nohp, COALESCE(notlp, '')),
			COALESCE(kelamin, ''),
			COALESCE(tmplahir, ''),
			COALESCE(tgllahir, ''),
			COALESCE(alamat, '')
		FROM pasiens
		WHERE id = ?
		LIMIT 1
	`, patientID).Scan(
		&patient.ID,
		&patient.NoRM,
		&patient.NIK,
		&patient.Name,
		&patient.Phone,
		&patient.Gender,
		&patient.BirthPlace,
		&patient.BirthDate,
		&patient.Address,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, bookingValidationError("data pasien tidak ditemukan")
		}
		return nil, err
	}

	if patient.NoRM == "" {
		return nil, bookingValidationError("pasien belum memiliki no rekam medis")
	}

	return &patient, nil
}

func (r *BookingRepository) findExistingGeneralBookingTx(ctx context.Context, tx *sql.Tx, patientID int64, queueDate time.Time) (*model.BookingQueueResponse, error) {
	var response model.BookingQueueResponse
	var numericQueue int64
	if err := tx.QueryRowContext(ctx, `
		SELECT
			r.id,
			COALESCE(a.id, 0),
			COALESCE(r.reg_id, ''),
			COALESCE(NULLIF(r.nomorantrian, ''), COALESCE(CAST(a.nomor AS CHAR), '')),
			COALESCE(a.nomor, 0),
			COALESCE(a.kelompok, ''),
			COALESCE(r.poli_id, 0),
			COALESCE(po.nama, ''),
			COALESCE(DATE_FORMAT(r.tgl_order, '%Y-%m-%d'), DATE_FORMAT(a.tanggal, '%Y-%m-%d'), DATE_FORMAT(r.created_at, '%Y-%m-%d'), '')
		FROM registrasis r
		LEFT JOIN antrian_poli a ON a.id = r.antrian_poli_id
		LEFT JOIN polis po ON po.id = r.poli_id
		WHERE r.pasien_id = ?
		AND r.input_from = 'mobile'
		AND (r.jkn IS NULL OR r.jkn <> 'Y')
		AND r.deleted_at IS NULL
		AND r.tgl_order = ?
		ORDER BY r.id DESC
		LIMIT 1
		FOR UPDATE
	`, patientID, queueDate.Format("2006-01-02")).Scan(
		&response.RegistrationID,
		&response.QueueID,
		&response.RegistrationCode,
		&response.QueueNumber,
		&numericQueue,
		&response.QueueGroup,
		&response.PoliID,
		&response.PoliName,
		&response.QueueDate,
	); err != nil {
		return nil, err
	}

	if response.QueueNumber == "" && numericQueue > 0 {
		response.QueueNumber = strconv.FormatInt(numericQueue, 10)
	}
	if numericQueue > 0 {
		response.QueueCode = formatGeneralQueueCode(int(numericQueue))
	} else {
		response.QueueCode = response.QueueNumber
	}
	response.ServiceMode = "umum"

	return &response, nil
}
func (r *BookingRepository) findExistingGeneralDummyBookingTx(ctx context.Context, tx *sql.Tx, noRM string, queueDate time.Time) (*model.BookingQueueResponse, error) {
	var response model.BookingQueueResponse
	var numericQueue int64
	if err := tx.QueryRowContext(ctx, `
		SELECT
			rd.id,
			COALESCE(rd.registrasi_id, 0),
			COALESCE(NULLIF(rd.kodebooking, ''), COALESCE(rd.nomorantrian, '')),
			COALESCE(rd.nomorantrian, ''),
			COALESCE(rd.angkaantrian, 0),
			COALESCE(po_exact.nama, COALESCE(po_code.nama, '')),
			COALESCE(po_exact.id, COALESCE(po_code.id, 0)),
			COALESCE(po_exact.nama, COALESCE(po_code.nama, '')),
			COALESCE(rd.tglperiksa, '')
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
		WHERE rd.no_rm = ?
		AND rd.tglperiksa = ?
		AND rd.flag = ?
		AND LOWER(COALESCE(rd.status, '')) <> 'dibatalkan'
		ORDER BY rd.id DESC
		LIMIT 1
		FOR UPDATE
	`, noRM, queueDate.Format("2006-01-02"), generalDummyFlag).Scan(
		&response.DummyID,
		&response.RegistrationID,
		&response.RegistrationCode,
		&response.QueueNumber,
		&numericQueue,
		&response.QueueGroup,
		&response.PoliID,
		&response.PoliName,
		&response.QueueDate,
	); err != nil {
		return nil, err
	}

	response.QueueID = 0
	if response.QueueNumber == "" && numericQueue > 0 {
		response.QueueNumber = strconv.FormatInt(numericQueue, 10)
	}
	if response.QueueNumber != "" && strings.Contains(response.QueueNumber, "-") {
		response.QueueCode = response.QueueNumber
	} else if numericQueue > 0 {
		response.QueueCode = formatGeneralQueueCode(int(numericQueue))
	} else {
		response.QueueCode = response.QueueNumber
	}
	response.ServiceMode = "umum"
	response.Source = generalDummyFlag

	return &response, nil
}

func acquireMySQLLock(ctx context.Context, tx *sql.Tx, lockName string) error {
	var lockStatus sql.NullInt64
	if err := tx.QueryRowContext(ctx, `SELECT GET_LOCK(?, 10)`, lockName).Scan(&lockStatus); err != nil {
		return err
	}
	if !lockStatus.Valid || lockStatus.Int64 != 1 {
		return bookingValidationError("gagal mengunci nomor antrian, silakan coba lagi")
	}
	return nil
}

func generateGeneralDummyQueueNumber(queueDate time.Time, poliCode string, queueNumber int) string {
	cleanPoliCode := strings.ToUpper(strings.TrimSpace(poliCode))
	cleanPoliCode = strings.ReplaceAll(cleanPoliCode, "-", "")
	cleanPoliCode = strings.ReplaceAll(cleanPoliCode, " ", "")
	if cleanPoliCode == "" {
		cleanPoliCode = "UM"
	}
	return fmt.Sprintf("%s%s%d", queueDate.Format("02012006"), cleanPoliCode, queueNumber)
}

func generateBookingRegistrationCode(queueDate time.Time, patientID int64, poliID int64) string {
	return fmt.Sprintf("%s%03d%03d", queueDate.Format("20060102"), patientID%1000, poliID%1000)
}

func generalDummyRequestPayload(ctxData *bookingContext, patient *bookingPatient, target BookingTarget, registrationNumber string, queueCode string, queueNumber string, bayar string, jenisPasien string) (string, error) {
	payload := map[string]any{
		"source":              generalDummyFlag,
		"flag":                generalDummyFlag,
		"jenisrequest":        nil,
		"jenisdaftar":         generalDummyJenisDaftar,
		"jenis_registrasi":    generalDummyJenisRegistrasi,
		"status":              generalDummyDefaultStatus,
		"pasien_id":           patient.ID,
		"no_rm":               patient.NoRM,
		"norm":                patient.NoRM,
		"nik":                 patient.NIK,
		"nama":                patient.Name,
		"no_hp":               patient.Phone,
		"tanggal_lahir":       patient.BirthDate,
		"tglperiksa":          ctxData.queueDate.Format("2006-01-02"),
		"tanggalperiksa":      ctxData.queueDate.Format("2006-01-02"),
		"poli_id":             ctxData.poliID,
		"kode_poli":           ctxData.poliCode,
		"nama_poli":           ctxData.poliName,
		"dokter_id":           ctxData.doctorID,
		"kode_dokter":         coalesceBookingString(ctxData.doctorID, ctxData.doctorCode),
		"nama_dokter":         ctxData.doctorName,
		"jampraktek":          ctxData.practiceTime,
		"kodebooking":         nil,
		"nomorantrian":        registrationNumber,
		"angkaantrian":        parseInt64Default(queueNumber, 0),
		"queue_code":          queueCode,
		"queue_group":         coalesceBookingString(ctxData.queueGroup, ctxData.poliName),
		"kode_cara_bayar":     bayar,
		"bayar":               bayar,
		"jenis_pasien":        jenisPasien,
		"jeniskunjungan":      generalDummyDefaultKunjungan,
		"booking_source_note": "Booking antrian umum dari aplikasi mobile",
	}
	if target.Email != "" {
		payload["email"] = target.Email
	}
	if target.QueueGroup != "" {
		payload["request_queue_group"] = target.QueueGroup
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func coalesceBookingString(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" && trimmed != "-" {
			return trimmed
		}
	}
	return ""
}

func firstCSVValue(value string) string {
	for _, part := range strings.Split(value, ",") {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" && trimmed != "-" {
			return trimmed
		}
	}
	return ""
}

func formatGeneralQueueCode(queueNumber int) string {
	if queueNumber <= 0 {
		return ""
	}
	return fmt.Sprintf("U-%03d", queueNumber)
}

func normalizeDummyGender(value string) string {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	normalized = strings.ReplaceAll(normalized, ".", "")
	switch normalized {
	case "L", "LK", "LKI", "LAKI", "LAKI-LAKI", "LAKI LAKI", "PRIA", "M", "MALE":
		return "L"
	case "P", "PR", "PEREMPUAN", "WANITA", "F", "FEMALE":
		return "P"
	default:
		return ""
	}
}

func parseInt64Default(value string, fallback int64) int64 {
	parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil {
		return fallback
	}
	return parsed
}

func normalizeBookingLimit(limit int) int {
	if limit < 1 {
		return 50
	}
	if limit > 200 {
		return 200
	}
	return limit
}

func bookingStatusCondition() string {
	return `(
		LOWER(COALESCE(r.status, '')) = ?
		OR LOWER(COALESCE(r.status_reg, '')) = ?
		OR LOWER(COALESCE(a.status, '')) = ?
		OR LOWER(COALESCE(a.panggil, '')) = ?
		OR (? = 'menunggu' AND (a.id IS NULL OR COALESCE(a.status, '0') = '0') AND COALESCE(a.panggil, '0') = '0' AND COALESCE(a.sudah_dipanggil, 0) = 0)
		OR (? = 'dipanggil' AND (COALESCE(a.panggil, '0') = '1' OR COALESCE(a.sudah_dipanggil, 0) = 1))
		OR (? = 'selesai' AND COALESCE(a.status, '') = '2')
		OR (? = 'batal' AND LOWER(COALESCE(r.batal, '')) IN ('1', 'y', 'ya', 'true'))
		OR (? = 'mobile' AND COALESCE(r.input_from, '') = 'mobile')
	)`
}

func dummyBookingStatusCondition() string {
	return `(
		LOWER(COALESCE(rd.status, '')) = ?
		OR (? = 'menunggu' AND LOWER(COALESCE(rd.status, '')) IN ('', 'pending', 'terdaftar', 'checkin'))
		OR (? = 'dipanggil' AND LOWER(COALESCE(rd.status, '')) = 'dilayani')
		OR (? = 'selesai' AND LOWER(COALESCE(rd.status, '')) IN ('selesai', 'selesai_dilayani'))
		OR (? = 'batal' AND LOWER(COALESCE(rd.status, '')) IN ('batal', 'dibatalkan'))
		OR (? = 'mobile' AND COALESCE(rd.flag, '') = 'mobile_umum')
	)`
}

func bookingStatusLabel(queueStatus string, callStatus string, alreadyCalled bool, statusReg string) string {
	normalizedQueueStatus := strings.ToLower(strings.TrimSpace(queueStatus))
	normalizedCallStatus := strings.ToLower(strings.TrimSpace(callStatus))
	normalizedStatusReg := strings.ToLower(strings.TrimSpace(statusReg))

	if normalizedQueueStatus == "2" || normalizedStatusReg == "selesai" {
		return "selesai"
	}
	if normalizedCallStatus == "1" || alreadyCalled || normalizedQueueStatus == "1" {
		return "dipanggil"
	}
	if normalizedStatusReg != "" && normalizedStatusReg != "-" && normalizedStatusReg != "0" {
		return normalizedStatusReg
	}
	return "menunggu"
}

func dummyBookingStatusLabel(status string) string {
	normalizedStatus := strings.ToLower(strings.TrimSpace(status))
	switch normalizedStatus {
	case "dibatalkan", "batal":
		return "batal"
	case "selesai", "selesai_dilayani":
		return "selesai"
	case "dilayani":
		return "dipanggil"
	case "", "pending", "terdaftar", "checkin":
		return "menunggu"
	default:
		return normalizedStatus
	}
}

func bookingStatusCounts(counts map[string]int) []model.MobileBookingStatusCount {
	order := []string{"menunggu", "dipanggil", "selesai", "batal"}
	results := make([]model.MobileBookingStatusCount, 0, len(counts))
	seen := map[string]bool{}

	for _, status := range order {
		count, ok := counts[status]
		if !ok {
			continue
		}
		results = append(results, model.MobileBookingStatusCount{Status: status, Count: count})
		seen[status] = true
	}

	for status, count := range counts {
		if seen[status] {
			continue
		}
		results = append(results, model.MobileBookingStatusCount{Status: status, Count: count})
	}

	return results
}
