package auth

import (
	"fmt"

	"apirusdotistamobile/internal/config"

	"gopkg.in/gomail.v2"
)

func SendOTPEmail(
	cfg config.SMTPConfig,
	to string,
	otp string,
) error {

	fmt.Println("SMTP HOST:", cfg.Host)
	fmt.Println("SMTP PORT:", cfg.Port)
	fmt.Println("SMTP EMAIL:", cfg.Email)

	m := gomail.NewMessage()

	m.SetHeader("From", cfg.Email)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Kode OTP Login")

	body := fmt.Sprintf(`
		<h2>Kode OTP Anda</h2>
		<h1>%s</h1>
		<p>OTP berlaku 5 menit</p>
	`, otp)

	m.SetBody("text/html", body)

	d := gomail.NewDialer(
		"smtp.gmail.com",
		587,
		cfg.Email,
		cfg.Password,
	)

	return d.DialAndSend(m)
}
