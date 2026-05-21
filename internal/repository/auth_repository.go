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

func (r *AuthRepository) CreateVerifEmail(
	user model.User,
	otp string,
) error {

	query := `
	INSERT INTO otp_verif_email(
		username,
		email,
		otp_code,
		expired_at
	)
	VALUES(
		?,
		?,
		?,
		DATE_ADD(NOW(), INTERVAL 5 MINUTE)
	)
	`

	_, err := r.DB.Exec(
		query,
		user.Username,
		user.Email,
		otp,
	)

	return err
}

func (r *AuthRepository) FindVerifiedEmail(
	email string,
) (*model.User, error) {

	query := `
	SELECT username, email
	FROM otp_verif_email
	WHERE email = ?
	AND is_used = TRUE
	ORDER BY id DESC
	LIMIT 1
	`

	var user model.User

	err := r.DB.QueryRow(
		query,
		email,
	).Scan(
		&user.Username,
		&user.Email,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *AuthRepository) VerifyOTPNewUser(
	email string,
	otp string,
) (bool, error) {

	query := `
	SELECT id
	FROM otp_verif_email
	WHERE email = ?
	AND otp_code = ?
	AND is_used = FALSE
	AND expired_at > NOW()
	ORDER BY id DESC
	LIMIT 1
	`

	var otpID int64

	err := r.DB.QueryRow(
		query,
		email,
		otp,
	).Scan(&otpID)

	if err != nil {
		return false, err
	}

	updateQuery := `
	UPDATE otp_verif_email
	SET is_used = TRUE
	WHERE id = ?
	`

	_, err = r.DB.Exec(updateQuery, otpID)

	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *AuthRepository) Create(
	user model.User,
) error {

	query := `
	INSERT INTO user(
		username,
		email,
		password,
		email_verified,
		verification_token
	)
	VALUES(
		?,
		?,
		?,
		true,
		?
	)
	`

	_, err := r.DB.Exec(
		query,
		user.Username,
		user.Email,
		user.Password,
		user.VerificationToken,
	)

	return err
}

func (r *AuthRepository) FindByEmailNewUser(Email string) (*model.UserRegister, error) {

	query := `
	SELECT id,Username,email
	FROM otp_verif_email
	WHERE Email = ?
	`

	var UserRegister model.UserRegister

	err := r.DB.QueryRow(query, Email).
		Scan(
			&UserRegister.ID,
			&UserRegister.Username,
			&UserRegister.Email,
		)

	if err != nil {
		return nil, err
	}

	return &UserRegister, nil
}

func (r *AuthRepository) CreateUserNew(
	userID int64,
	otp string,
) error {

	query := `
	INSERT INTO user(
		email,
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

// func (r *Auth) InsertUserPass(
// 	c *fiber.Ctx,
// ) error {

// 	var req model.SetPasswordRequest

// 	if err := c.BodyParser(&req); err != nil {

// 		return c.Status(400).JSON(
// 			fiber.Map{
// 				"message": "invalid request body",
// 			},
// 		)
// 	}

// 	err := h.Service.SetPassword(req)

// 	if err != nil {

// 		return c.Status(400).JSON(
// 			fiber.Map{
// 				"message": err.Error(),
// 			},
// 		)
// 	}

// 	return c.JSON(fiber.Map{
// 		"message": "register success",
// 	})
// }

func (r *AuthRepository) FindByEmailUser(Email string) (*model.UserRegister, error) {

	query := `
	SELECT id,Username,email
	FROM user
	WHERE Email = ?
	`

	var UserRegister model.UserRegister

	err := r.DB.QueryRow(query, Email).
		Scan(
			&UserRegister.ID,
			&UserRegister.Username,
			&UserRegister.Email,
		)

	if err != nil {
		return nil, err
	}

	return &UserRegister, nil
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

func (r *AuthRepository) CreateOTPLogin(
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

func (r *AuthRepository) VerifyOTPLogin(
	userID int64,
	otp string,
) (bool, error) {

	query := `
	SELECT id
	FROM otp_user
	WHERE user_id = ?
	AND otp_code = ?
	AND is_used = FALSE
	AND expired_at > NOW()
	ORDER BY id DESC
	LIMIT 1
	`

	var otpID int64

	err := r.DB.QueryRow(
		query,
		userID,
		otp,
	).Scan(&otpID)

	if err != nil {
		return false, err
	}

	updateQuery := `
	UPDATE otp_user
	SET is_used = TRUE
	WHERE id = ?
	`

	_, err = r.DB.Exec(updateQuery, otpID)

	if err != nil {
		return false, err
	}

	return true, nil
}
