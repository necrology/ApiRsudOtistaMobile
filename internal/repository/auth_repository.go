package repository

import (
	"database/sql"

	"apirusdotistamobile/internal/model"
)

type AuthRepository struct {
	DB *sql.DB
}

func NewAuthRepository(db *sql.DB) *AuthRepository {
	return &AuthRepository{
		DB: db,
	}
}

func (r *AuthRepository) Create(user model.User) error {

	query := `
	INSERT INTO user(Username,email,password,email_verified,verification_token)
	VALUES(?,?,?,?,?)
	`

	_, err := r.DB.Exec(
		query,
		user.Username,
		user.Email,
		user.Password,
		user.EmailVerified,
		user.VerificationToken,
	)

	return err
}

func (r *AuthRepository) FindByUsername(Username string) (*model.User, error) {

	query := `
	SELECT id,Username,email,password,email_verified
	FROM user
	WHERE Username = ?
	`

	var user model.User

	err := r.DB.QueryRow(query, Username).
		Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Password,
			&user.EmailVerified,
			//&user.VerificationToken,
		)

	if err != nil {
		return nil, err
	}

	return &user, nil
}
func (r *AuthRepository) VerifyEmail(token string) error {

	query := `
	UPDATE user
	SET
		email_verified = true,
		verified_at = NOW(),
		verification_token = NULL
	WHERE verification_token = ?
	`

	_, err := r.DB.Exec(query, token)

	return err
}

func (r *AuthRepository) CreateOTP(
	userID int64,
	otp string,
) error {

	query := `
	INSERT INTO otp_user(
		user_id,
		otp_code,
		expired_at
	)
	VALUES(
		?,
		?,
		DATE_ADD(NOW(), INTERVAL 5 MINUTE)
	)
	`

	_, err := r.DB.Exec(
		query,
		userID,
		otp,
	)

	return err
}
