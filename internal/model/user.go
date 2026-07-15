package model

type User struct {
	ID            int64
	Username      string
	Email         string
	NoRM          string
	PatientID     int64
	Phone         string
	FullName      string
	Password      string
	EmailVerified bool
}

type PatientMedicalRecord struct {
	ID       int64
	NoRM     string
	FullName string
	Phone    string
}
