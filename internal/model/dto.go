package model

type RegisterRequest struct {
	Username string
	Email    string
	Password string
}

type LoginRequest struct {
	Username string
	Password string
}

type VerifyOTPRequestLogin struct {
	Username string
	OTP      string
}

type UserRegister struct {
	ID       int64
	Username string
	Email    string
	Otp      string
}

type VerifyEmailUser struct {
	Username string
	OTP      string
}

type VerifyOTPNewUser struct {
	Email string
	OTP   string
}

type SetPasswordRequest struct {
	Email    string
	Password string
}
