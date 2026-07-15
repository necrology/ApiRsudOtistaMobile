package model

type RegisterRequest struct {
	Username         string
	Email            string
	NoRM             string
	Phone            string
	FullName         string
	HasMedicalRecord bool
}

type LoginRequest struct {
	Username   string
	Identifier string
	Email      string
	NoRM       string
	Password   string
}

type VerifyOTPRequestLogin struct {
	Username   string
	Identifier string
	Email      string
	OTP        string
}

type VerifyOTPNewUser struct {
	Email string
	OTP   string
}

type SetPasswordRequest struct {
	Password           string `json:"password"`
	RegistrationTicket string `json:"registration_ticket"`
}

type RefreshSessionRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type ForgotPasswordRequest struct {
	Identifier string
	Email      string
	NoRM       string
}

type ResetPasswordRequest struct {
	Identifier string
	Email      string
	NoRM       string
	OTP        string
	Password   string
}

type RequestMedicalRecordClaim struct {
	Password  string
	NoRM      string
	NIK       string
	BirthDate string
}

type ConfirmMedicalRecordClaim struct {
	OTP string
}

type BookingQueueResponse struct {
	RegistrationID   int64  `json:"registration_id"`
	DummyID          int64  `json:"dummy_id,omitempty"`
	QueueID          int64  `json:"queue_id"`
	RegistrationCode string `json:"registration_code"`
	QueueNumber      string `json:"queue_number"`
	QueueCode        string `json:"queue_code"`
	QueueGroup       string `json:"queue_group"`
	PoliID           int64  `json:"poli_id"`
	PoliName         string `json:"poli_name"`
	QueueDate        string `json:"queue_date"`
	ServiceMode      string `json:"service_mode"`
	Source           string `json:"source,omitempty"`
	Existing         bool   `json:"existing"`
}

type MobileBookingListResponse struct {
	Items        []MobileBookingRecord      `json:"items"`
	StatusCounts []MobileBookingStatusCount `json:"status_counts"`
}

type MobileBookingRecord struct {
	RegistrationID     int64  `json:"registration_id"`
	DummyID            int64  `json:"dummy_id,omitempty"`
	QueueID            int64  `json:"queue_id"`
	RegistrationCode   string `json:"registration_code"`
	QueueNumber        string `json:"queue_number"`
	QueueCode          string `json:"queue_code"`
	QueueGroup         string `json:"queue_group"`
	QueueDate          string `json:"queue_date"`
	CreatedAt          string `json:"created_at"`
	RegistrationStatus string `json:"registration_status"`
	StatusReg          string `json:"status_reg"`
	QueueStatus        string `json:"queue_status"`
	QueueStatusLabel   string `json:"queue_status_label"`
	CallStatus         string `json:"call_status"`
	AlreadyCalled      bool   `json:"already_called"`
	PatientID          int64  `json:"patient_id"`
	NoRM               string `json:"no_rm"`
	PatientName        string `json:"nama_pasien"`
	PoliID             int64  `json:"poli_id"`
	PoliName           string `json:"poli_name"`
	DoctorID           int64  `json:"dokter_id"`
	DoctorName         string `json:"nama_dokter"`
	InputFrom          string `json:"input_from"`
	Source             string `json:"source,omitempty"`
	PatientType        string `json:"jenis_pasien"`
	PaymentType        string `json:"bayar"`
	Note               string `json:"keterangan"`
}

type MobileBookingStatusCount struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

type BookingDoctorOption struct {
	ID          int64                   `json:"id"`
	Nama        string                  `json:"nama"`
	KodeAntrian string                  `json:"kode_antrian"`
	KodeBpjs    string                  `json:"kode_bpjs"`
	GeneralCode string                  `json:"general_code"`
	Schedules   []BookingScheduleOption `json:"jadwal,omitempty"`
}

type BookingScheduleOption struct {
	Source     string `json:"source"`
	DayIndex   int    `json:"day_index"`
	Hari       string `json:"hari"`
	Buka       string `json:"buka"`
	Tutup      string `json:"tutup"`
	Praktik    bool   `json:"praktik"`
	Kuota      int64  `json:"kuota,omitempty"`
	PoliID     int64  `json:"poli_id,omitempty"`
	PoliName   string `json:"nama_poli,omitempty"`
	DoctorID   int64  `json:"dokter_id,omitempty"`
	DoctorName string `json:"nama_dokter,omitempty"`
}

type BookingPoliOption struct {
	ID             int64                   `json:"id"`
	Nama           string                  `json:"nama"`
	Kelompok       string                  `json:"kelompok"`
	Politype       string                  `json:"politype"`
	BpjsCode       string                  `json:"bpjs"`
	KodeRuangan    string                  `json:"kode_ruangan"`
	Buka           string                  `json:"buka"`
	Tutup          string                  `json:"tutup"`
	Praktik        string                  `json:"praktik"`
	Kuota          int64                   `json:"kuota"`
	KuotaOnline    int64                   `json:"kuota_online"`
	Terisi         int64                   `json:"terisi"`
	QueueGroup     string                  `json:"queue_group"`
	QueueGroupHint string                  `json:"queue_group_hint"`
	Doctors        []BookingDoctorOption   `json:"dokter_list"`
	Schedules      []BookingScheduleOption `json:"jadwal_list"`
}

type BookingOptionsResponse struct {
	Poli BookingPoliOption `json:"poli"`
}

type BookingCalendarResponse struct {
	Year  int                  `json:"year"`
	Month int                  `json:"month"`
	Days  []BookingCalendarDay `json:"days"`
}

type BookingCalendarDay struct {
	Date              string                  `json:"date"`
	DayName           string                  `json:"day_name"`
	IsSunday          bool                    `json:"is_sunday"`
	HasSundaySchedule bool                    `json:"has_sunday_schedule"`
	HasPoliSchedule   bool                    `json:"has_poli_schedule"`
	HasDoctorSchedule bool                    `json:"has_doctor_schedule"`
	IsHoliday         bool                    `json:"is_holiday"`
	IsLeave           bool                    `json:"is_leave"`
	IsOffDay          bool                    `json:"is_off_day"`
	IsOpen            bool                    `json:"is_open"`
	Reason            string                  `json:"reason"`
	HolidayNames      []string                `json:"holiday_names"`
	Schedules         []BookingScheduleOption `json:"jadwal,omitempty"`
}
