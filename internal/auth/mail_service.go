package auth

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"apirusdotistamobile/internal/config"

	"gopkg.in/gomail.v2"
)

func SendOTPEmail(
	cfg config.SMTPConfig,
	to string,
	otp string,
) error {
	return SendOTPEmailWithPurpose(cfg, to, otp, "Kode OTP Login", "Masuk ke RSUD Otista Mobile")
}

func SendOTPEmailWithPurpose(
	cfg config.SMTPConfig,
	to string,
	otp string,
	subject string,
	purpose string,
) error {
	if strings.TrimSpace(cfg.Host) == "" || strings.TrimSpace(cfg.Port) == "" || strings.TrimSpace(cfg.Email) == "" || strings.TrimSpace(cfg.Password) == "" {
		return errors.New("smtp config is incomplete")
	}

	port, err := strconv.Atoi(strings.TrimSpace(cfg.Port))
	if err != nil {
		return fmt.Errorf("invalid smtp port: %w", err)
	}

	m := gomail.NewMessage()

	m.SetHeader("From", cfg.Email)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)

	body := fmt.Sprintf(`
		<div style="margin:0;padding:24px;background:#edf7f5;font-family:Arial,sans-serif;color:#173d39">
			<div style="max-width:560px;margin:0 auto;background:#ffffff;border-radius:18px;overflow:hidden;border:1px solid #d7ebe7">
				<div style="padding:24px;background:linear-gradient(135deg,#41AB9F,#1A4540);color:#ffffff">
					<div style="font-size:13px;letter-spacing:.04em;text-transform:uppercase">RSUD Oto Iskandar Di Nata</div>
					<h1 style="margin:8px 0 0;font-size:24px;line-height:1.25">Verifikasi Akun Mobile</h1>
				</div>
				<div style="padding:24px">
					<p style="font-size:16px;line-height:1.6;margin:0 0 16px">Gunakan kode berikut untuk %s.</p>
					<div style="margin:20px 0;padding:18px;border-radius:14px;background:#e7f5f2;text-align:center;border:1px solid #c4e5df">
						<div style="font-size:32px;font-weight:800;letter-spacing:8px;color:#1A4540">%s</div>
					</div>
					<p style="font-size:14px;line-height:1.6;margin:0;color:#5d6675">Kode berlaku selama 5 menit. Jangan bagikan kode ini kepada siapa pun, termasuk petugas rumah sakit.</p>
					<hr style="border:none;border-top:1px solid #d7ebe7;margin:24px 0">
					<p style="font-size:13px;line-height:1.6;margin:0;color:#7a8494">Email ini dikirim otomatis oleh layanan RSUD Otista Mobile. Abaikan email ini jika Anda tidak meminta kode verifikasi.</p>
				</div>
			</div>
		</div>
	`, strings.ToLower(purpose), otp)

	m.SetBody("text/html", body)

	d := gomail.NewDialer(
		strings.TrimSpace(cfg.Host),
		port,
		cfg.Email,
		cfg.Password,
	)

	return d.DialAndSend(m)
}
