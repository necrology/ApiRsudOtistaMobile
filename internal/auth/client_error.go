package auth

// ClientErrorMessage memisahkan pesan validasi/credential yang aman untuk
// dikirim ke klien dari error repository, SQL, SMTP, atau crypto. Handler tidak
// boleh meneruskan err.Error() yang tidak ada dalam allowlist ini.
func ClientErrorMessage(err error) (string, bool) {
	if err == nil {
		return "", false
	}

	_, ok := safeClientErrorMessages[err.Error()]
	return err.Error(), ok
}

var safeClientErrorMessages = map[string]struct{}{
	"email tidak valid":                {},
	"invalid email":                    {},
	"email is required":                {},
	"username and email are required":  {},
	"username already exists":          {},
	"email already exists":             {},
	"email not verified":               {},
	"email atau password tidak sesuai": {},
	"email atau no rm wajib diisi":     {},
	"No. RM belum terhubung ke akun mobile. Silakan login dengan email lalu hubungkan No. RM di profil.": {},
	"No. RM terhubung ke lebih dari satu akun. Silakan login dengan email.":                              {},
	"invalid otp": {},
	"registration ticket tidak valid atau kedaluwarsa":       {},
	"email, otp, and password are required":                  {},
	"password minimal 8 karakter":                            {},
	"password harus berisi huruf dan angka":                  {},
	"password maksimal 72 byte":                              {},
	"password tidak boleh diawali atau diakhiri spasi":       {},
	"data registrasi tidak valid":                            {},
	"akun belum valid untuk mengubah no rm":                  {},
	"data rekam medis tidak cocok atau tidak bisa digunakan": {},
	"otp klaim no rm tidak valid":                            {},
}
