package auth

import (
	"errors"

	"apirusdotistamobile/internal/config"
	"apirusdotistamobile/internal/model"
	"apirusdotistamobile/internal/repository"

	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
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

func (s *Service) Register(req RegisterRequest) error {

	if req.Username == "" ||
		req.Email == "" ||
		req.Password == "" {

		return errors.New("Username, email, and password are required")
	}

	existingUser, _ := s.Repo.FindByUsername(req.Username)

	if existingUser != nil {
		return errors.New("username already exists")
	}

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(req.Password),
		bcrypt.DefaultCost,
	)

	if err != nil {
		return err
	}

	token := uuid.NewString()
	user := model.User{
		Username:          req.Username,
		Email:             req.Email,
		Password:          string(hash),
		VerificationToken: token,
	}

	return s.Repo.Create(user)
}

func (s *Service) Login(req LoginRequest) error {

	user, err := s.Repo.FindByUsername(req.Username)

	if err != nil {
		return errors.New("username not found")
	}
	if user.EmailVerified {

		otp := GenerateOTP()

		err = s.Repo.CreateOTP(
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
	err = bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(req.Password),
	)

	if err != nil {
		return errors.New("wrong password")
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
