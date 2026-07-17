package auth

import (
	"context"
	"database/sql"
	"errors"
	"net/mail"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"apirusdotistamobile/internal/config"
	"apirusdotistamobile/internal/model"
	"apirusdotistamobile/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	Repo *repository.AuthRepository
	SMTP config.SMTPConfig
	Mail *OTPEmailDispatcher
}

var errInvalidLoginCredentials = errors.New("email atau password tidak sesuai")

var ErrAccountAlreadyRegistered = errors.New("account already registered")

const dummyPasswordHash = "$2y$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2uheWG/igi."

func NewService(
	repo *repository.AuthRepository,
	smtp config.SMTPConfig,
) *Service {

	return &Service{
		Repo: repo,
		SMTP: smtp,
		Mail: newProductionOTPEmailDispatcher(smtp),
	}
}

func (s *Service) Close(ctx context.Context) error {
	if s == nil || s.Mail == nil {
		return nil
	}
	return s.Mail.Close(ctx)
}

func (s *Service) RegisterUser(
	req model.RegisterRequest,
) error {
	req.Username = strings.TrimSpace(req.Username)
	email, err := normalizeEmail(req.Email)
	if err != nil {
		return errors.New("email tidak valid")
	}
	req.Email = email
	req.NoRM = ""
	req.Phone = strings.TrimSpace(req.Phone)
	req.FullName = strings.TrimSpace(req.FullName)

	if req.Username == "" && req.FullName != "" {
		req.Username = req.FullName
	}

	if req.Username == "" || req.Email == "" {

		return errors.New(
			"username and email are required",
		)
	}
	if utf8.RuneCountInString(req.Username) > 100 ||
		utf8.RuneCountInString(req.Phone) > 30 ||
		utf8.RuneCountInString(req.FullName) > 150 {
		return errors.New("data registrasi tidak valid")
	}

	existingEmail, err := s.Repo.FindUserByEmail(req.Email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	if existingEmail != nil {
		if err := consumeDummyOTPWork(); err != nil {
			return err
		}
		return ErrAccountAlreadyRegistered
	}

	otp, err := GenerateOTP()
	if err != nil {
		return err
	}
	otpHash, err := hashOTP(otp)
	if err != nil {
		return err
	}

	err = s.Repo.CreateVerifEmail(
		model.User{
			Username: req.Username,
			Email:    req.Email,
			NoRM:     req.NoRM,
			Phone:    req.Phone,
			FullName: req.FullName,
		},
		otpHash,
	)

	if err != nil {
		return err
	}

	return s.enqueueOTPEmail(
		req.Email,
		otp,
		"Verifikasi Email RSUD Otista Mobile",
		"verifikasi email akun RSUD Otista Mobile",
	)
}

func (s *Service) VerifyOTPNewUser(
	ctx context.Context,
	req model.VerifyOTPNewUser,
) (*RegistrationTicket, error) {
	email, err := normalizeEmail(req.Email)
	if err != nil {
		return nil, errors.New("invalid otp")
	}
	if !validOTPFormat(req.OTP) {
		return nil, errors.New("invalid otp")
	}

	verificationID, valid, err := s.Repo.VerifyOTPNewUser(
		email,
		req.OTP,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("invalid otp")
		}
		return nil, err
	}
	if !valid {
		return nil, errors.New("invalid otp")
	}

	ticket, ticketHash, err := newRegistrationTicket()
	if err != nil {
		return nil, err
	}
	if err = s.Repo.CreateAuthTicket(
		ctx,
		verificationID,
		email,
		registrationTicketPurpose,
		ticketHash,
		ticket.ExpiresAt,
	); err != nil {
		return nil, err
	}

	return ticket, nil
}

func (s *Service) SetPassword(
	ctx context.Context,
	req model.SetPasswordRequest,
) (*model.User, error) {
	if err := validatePassword(req.Password); err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.RegistrationTicket) == "" {
		return nil, errors.New("registration ticket tidak valid atau kedaluwarsa")
	}

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(req.Password),
		bcrypt.DefaultCost,
	)

	if err != nil {
		return nil, err
	}

	ticketHash, err := hashPresentedAuthTicket(req.RegistrationTicket)
	if err != nil {
		return nil, errors.New("registration ticket tidak valid atau kedaluwarsa")
	}
	user, err := s.Repo.CompleteRegistration(
		ctx,
		registrationTicketPurpose,
		ticketHash,
		string(hash),
		time.Now(),
	)
	if err != nil {
		if errors.Is(err, repository.ErrInvalidAuthTicket) {
			return nil, errors.New("registration ticket tidak valid atau kedaluwarsa")
		}
		return nil, err
	}

	return user, nil
}

func (s *Service) Login(req model.LoginRequest) error {
	user, err := s.resolveLoginUser(req.Identifier, req.Email, req.NoRM, req.Username)
	if err != nil {
		if errors.Is(err, errInvalidLoginCredentials) {
			_ = bcrypt.CompareHashAndPassword([]byte(dummyPasswordHash), []byte(req.Password))
			return errInvalidLoginCredentials
		}
		return err
	}

	// cek password dulu
	err = bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(req.Password),
	)

	if err != nil {
		return errInvalidLoginCredentials
	}

	// baru OTP
	if user.EmailVerified {

		otp, err := GenerateOTP()
		if err != nil {
			return err
		}
		otpHash, err := hashOTP(otp)
		if err != nil {
			return err
		}

		err = s.Repo.CreateOTPLogin(
			user.ID,
			otpHash,
		)

		if err != nil {
			return err
		}

		err = s.enqueueOTPEmail(
			user.Email,
			otp,
			"Kode OTP Login RSUD Otista Mobile",
			"login ke RSUD Otista Mobile",
		)

		if err != nil {
			return err
		}

		return nil
	}

	return errors.New("email not verified")
}
func (s *Service) VerifyOTPLogin(
	req model.VerifyOTPRequestLogin,
) (*model.User, error) {
	if !validOTPFormat(req.OTP) {
		return nil, errors.New("invalid otp")
	}
	user, err := s.resolveLoginUser(req.Identifier, req.Email, req.Username)
	if err != nil {
		if !errors.Is(err, errInvalidLoginCredentials) {
			return nil, err
		}
		return nil, errors.New("invalid otp")
	}

	valid, err := s.Repo.VerifyOTPLogin(
		user.ID,
		req.OTP,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("invalid otp")
		}
		return nil, err
	}
	if !valid {
		return nil, errors.New("invalid otp")
	}

	return user, nil
}

func (s *Service) ForgotPassword(req model.ForgotPasswordRequest) error {
	identifier, err := normalizeEmailIdentifier(req.Identifier, req.Email)
	if err != nil {
		return errors.New("email is required")
	}

	user, err := s.Repo.FindByIdentifier(identifier)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return consumeDummyOTPWork()
		}
		return err
	}

	otp, err := GenerateOTP()
	if err != nil {
		return err
	}
	otpHash, err := hashOTP(otp)
	if err != nil {
		return err
	}
	if err = s.Repo.CreatePasswordResetOTP(user.ID, otpHash); err != nil {
		return err
	}

	return s.enqueueOTPEmail(
		user.Email,
		otp,
		"Reset Password RSUD Otista Mobile",
		"mengatur ulang password akun RSUD Otista Mobile",
	)
}

func consumeDummyOTPWork() error {
	otp, err := GenerateOTP()
	if err != nil {
		return err
	}
	_, err = hashOTP(otp)
	return err
}

func (s *Service) ResetPassword(req model.ResetPasswordRequest) error {
	identifier, emailErr := normalizeEmailIdentifier(req.Identifier, req.Email)

	if emailErr != nil || strings.TrimSpace(req.OTP) == "" || strings.TrimSpace(req.Password) == "" {
		return errors.New("email, otp, and password are required")
	}
	if !validOTPFormat(req.OTP) {
		return errors.New("invalid otp")
	}

	if err := validatePassword(req.Password); err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user, err := s.Repo.FindByIdentifier(identifier)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("invalid otp")
		}
		return err
	}

	valid, err := s.Repo.ResetPassword(user.ID, strings.TrimSpace(req.OTP), string(hash))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("invalid otp")
		}
		return err
	}
	if !valid {
		return errors.New("invalid otp")
	}

	return nil
}

func (s *Service) RequestMedicalRecordClaim(
	userID int64,
	req model.RequestMedicalRecordClaim,
) error {
	if userID <= 0 {
		return errors.New("akun belum valid untuk mengubah no rm")
	}

	password := strings.TrimSpace(req.Password)
	noRM := strings.TrimSpace(req.NoRM)
	nik := normalizeDigits(req.NIK)
	birthDate, err := normalizeBirthDate(req.BirthDate)
	if err != nil {
		return errors.New("data rekam medis tidak cocok atau tidak bisa digunakan")
	}

	if password == "" || noRM == "" || nik == "" || utf8.RuneCountInString(noRM) > 20 {
		return errors.New("data rekam medis tidak cocok atau tidak bisa digunakan")
	}

	if len(nik) != 16 {
		return errors.New("data rekam medis tidak cocok atau tidak bisa digunakan")
	}

	user, err := s.Repo.FindUserByID(userID)
	if err != nil || !user.EmailVerified {
		return errors.New("akun belum valid untuk mengubah no rm")
	}

	if err = bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(password),
	); err != nil {
		return errors.New("akun belum valid untuk mengubah no rm")
	}

	patient, err := s.Repo.FindPatientByMedicalRecordClaim(
		noRM,
		nik,
		birthDate,
	)
	if err != nil {
		return errors.New("data rekam medis tidak cocok atau tidak bisa digunakan")
	}

	linked, err := s.Repo.IsPatientLinkedToAnotherUser(patient.ID, user.ID)
	if err != nil {
		return err
	}
	if linked {
		return errors.New("data rekam medis tidak cocok atau tidak bisa digunakan")
	}

	otp, err := GenerateOTP()
	if err != nil {
		return err
	}
	otpHash, err := hashOTP(otp)
	if err != nil {
		return err
	}
	if err = s.Repo.CreateMedicalRecordClaimOTP(user.ID, *patient, otpHash); err != nil {
		return err
	}

	return s.enqueueOTPEmail(
		user.Email,
		otp,
		"Verifikasi No. RM RSUD Otista Mobile",
		"menghubungkan No. RM ke akun RSUD Otista Mobile",
	)
}

func (s *Service) enqueueOTPEmail(to string, otp string, subject string, purpose string) error {
	if s == nil || s.Mail == nil {
		return ErrMailUnavailable
	}
	return s.Mail.Enqueue(to, otp, subject, purpose)
}

func (s *Service) ConfirmMedicalRecordClaim(
	userID int64,
	req model.ConfirmMedicalRecordClaim,
) (*model.User, error) {
	otp := strings.TrimSpace(req.OTP)
	if userID <= 0 || !validOTPFormat(otp) {
		return nil, errors.New("otp klaim no rm tidak valid")
	}

	user, valid, err := s.Repo.ConfirmMedicalRecordClaim(userID, otp)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("otp klaim no rm tidak valid")
		}
		return nil, err
	}
	if !valid {
		return nil, errors.New("otp klaim no rm tidak valid")
	}

	return user, nil
}

func normalizeDigits(value string) string {
	var builder strings.Builder
	for _, char := range value {
		if unicode.IsDigit(char) {
			builder.WriteRune(char)
		}
	}

	return builder.String()
}

func (s *Service) resolveLoginUser(values ...string) (*model.User, error) {
	identifier := firstNonEmpty(values...)
	if identifier == "" {
		return nil, errors.New("email atau no rm wajib diisi")
	}
	if utf8.RuneCountInString(identifier) > 100 {
		return nil, errInvalidLoginCredentials
	}

	if email, err := normalizeEmail(identifier); err == nil {
		user, err := s.Repo.FindUserByEmail(email)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errInvalidLoginCredentials
		}
		if err != nil {
			return nil, err
		}
		return user, nil
	}

	user, err := s.Repo.FindLinkedUserByNoRM(normalizeMedicalRecordNumber(identifier))
	if err == nil {
		return user, nil
	}
	if errors.Is(err, repository.ErrNoRMNotLinked) || errors.Is(err, repository.ErrNoRMAmbiguous) {
		return nil, errInvalidLoginCredentials
	}
	return nil, err
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}

	return ""
}

func normalizeMedicalRecordNumber(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func normalizeEmailIdentifier(values ...string) (string, error) {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		return normalizeEmail(value)
	}

	return "", errors.New("email is required")
}

func normalizeEmail(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", errors.New("email is required")
	}

	address, err := mail.ParseAddress(value)
	if err != nil || address.Address == "" || address.Name != "" ||
		!strings.EqualFold(address.Address, value) || utf8.RuneCountInString(address.Address) > 100 {
		return "", errors.New("invalid email")
	}

	return strings.ToLower(address.Address), nil
}

func validatePassword(password string) error {
	if password != strings.TrimSpace(password) {
		return errors.New("password tidak boleh diawali atau diakhiri spasi")
	}
	if utf8.RuneCountInString(password) < 8 {
		return errors.New("password minimal 8 karakter")
	}
	if len([]byte(password)) > 72 {
		return errors.New("password maksimal 72 byte")
	}

	hasLetter := false
	hasDigit := false
	for _, char := range password {
		if unicode.IsLetter(char) {
			hasLetter = true
		}
		if unicode.IsDigit(char) {
			hasDigit = true
		}
	}

	if !hasLetter || !hasDigit {
		return errors.New("password harus berisi huruf dan angka")
	}

	return nil
}

func validOTPFormat(value string) bool {
	value = strings.TrimSpace(value)
	if len(value) != 6 {
		return false
	}
	for _, character := range value {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}

func normalizeBirthDate(value string) (string, error) {
	value = strings.TrimSpace(value)
	for _, layout := range []string{"2006-01-02", "02-01-2006", "02/01/2006"} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed.Format("2006-01-02"), nil
		}
	}

	return "", errors.New("invalid birth date")
}
