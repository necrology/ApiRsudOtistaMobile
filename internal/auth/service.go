package auth

import (
	"errors"

	"apirusdotistamobile/internal/config"
	"apirusdotistamobile/internal/model"
	"apirusdotistamobile/internal/repository"

	"fmt"
	"math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	Repo *repository.AuthRepository
	SMTP config.SMTPConfig
}

func NewService(
	repo *repository.AuthRepository,
	smtp config.SMTPConfig,
) *Service {

	return &Service{
		Repo: repo,
		SMTP: smtp,
	}
}

func (s *Service) RegisterUser(
	req model.RegisterRequest,
) error {

	if req.Username == "" ||
		req.Email == "" {

		return errors.New(
			"username and email are required",
		)
	}

	existingUser, _ := s.Repo.FindByUsername(
		req.Username,
	)

	if existingUser != nil {
		return errors.New(
			"username already exists",
		)
	}
	existingEmail, _ := s.Repo.FindByEmailNewUser(
		req.Email,
	)

	if existingEmail != nil {
		return errors.New(
			"email already exists",
		)
	}

	otp := GenerateOTP()

	err := s.Repo.CreateVerifEmail(
		model.User{
			Username: req.Username,
			Email:    req.Email,
		},
		otp,
	)

	if err != nil {
		return err
	}

	err = SendOTPEmail(
		s.SMTP,
		req.Email,
		otp,
	)

	if err != nil {
		return err
	}

	return nil
}

func (s *Service) VerifyOTPNewUser(
	req model.VerifyOTPNewUser,
) error {

	user, err := s.Repo.FindByEmailNewUser(
		req.Email,
	)

	if err != nil {
		return errors.New("Email not found")
	}

	valid, err := s.Repo.VerifyOTPNewUser(
		user.Email,
		req.OTP,
	)

	if err != nil || !valid {
		return errors.New("invalid otp")
	}

	return nil
}

func (s *Service) SetPassword(
	req model.SetPasswordRequest,
) error {

	data, err := s.Repo.FindVerifiedEmail(
		req.Email,
	)

	if err != nil {
		return errors.New("email not verified")
	}

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(req.Password),
		bcrypt.DefaultCost,
	)

	if err != nil {
		return err
	}

	user := model.User{
		Username: data.Username,
		Email:    data.Email,
		Password: string(hash),
	}

	return s.Repo.Create(user)
}

func (s *Service) Login(req model.LoginRequest) error {

	user, err := s.Repo.FindByUsername(req.Username)

	if err != nil {
		return errors.New("username not found")
	}

	// cek password dulu
	err = bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(req.Password),
	)

	if err != nil {
		return errors.New("wrong password")
	}

	// baru OTP
	if user.EmailVerified {

		otp := GenerateOTP()

		err = s.Repo.CreateOTPLogin(
			user.ID,
			otp,
		)

		if err != nil {
			return err
		}

		err = SendOTPEmail(
			s.SMTP,
			user.Email,
			otp,
		)

		if err != nil {
			return err
		}

		return errors.New("otp sent to email")
	}

	return nil
}
func (s *Service) VerifyEmail(token string) error {

	return s.Repo.VerifyEmail(token)
}
func GenerateOTP() string {

	rand.Seed(time.Now().UnixNano())

	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func (s *Service) VerifyOTPLogin(
	req model.VerifyOTPRequestLogin,
) error {

	user, err := s.Repo.FindByUsername(
		req.Username,
	)

	if err != nil {
		return errors.New("username not found")
	}

	valid, err := s.Repo.VerifyOTPLogin(
		user.ID,
		req.OTP,
	)

	if err != nil || !valid {
		return errors.New("invalid otp")
	}

	return nil
}
