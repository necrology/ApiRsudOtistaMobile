package handlers

import (
	"bytes"
	"os"
	"strings"
	"time"

	"apirusdotistamobile/internal/repository"

	"github.com/go-pdf/fpdf"
)

func renderMedicalResumePDF(doc *repository.MobilePatientMedicalResumeDocument) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetTitle(medicalResumeTitle(doc), false)
	pdf.SetAuthor("RSUD Oto Iskandar Di Nata", false)
	pdf.SetMargins(12, 12, 12)
	pdf.SetAutoPageBreak(true, 14)
	pdf.AddPage()
	pdf.SetDrawColor(30, 30, 30)
	pdf.SetLineWidth(0.2)

	if doc.IsInpatient {
		renderInpatientResume(pdf, doc)
	} else {
		renderOutpatientResume(pdf, doc)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func renderOutpatientResume(pdf *fpdf.Fpdf, doc *repository.MobilePatientMedicalResumeDocument) {
	pageW, _ := pdf.GetPageSize()
	left, _, right, _ := pdf.GetMargins()
	usableW := pageW - left - right

	x := pdf.GetX()
	y := pdf.GetY()
	headerH := 40.0
	logoW := 55.0
	pdf.Rect(x, y, usableW, headerH, "")
	pdf.Line(x+logoW, y, x+logoW, y+headerH)
	drawLogo(pdf, x+21, y+3, 13)
	pdf.SetXY(x, y+15)
	pdf.SetFont("Times", "B", 9.5)
	pdf.MultiCell(logoW, 4.2, "RSUD OTO ISKANDAR\nDI NATA", "", "C", false)
	pdf.SetFont("Times", "", 5.4)
	pdf.SetX(x)
	pdf.MultiCell(logoW, 2.8, "Jl. Gading Tutuka, RT. 01 RW. 01,\nKp. Cincin Kolot, Kec. Soreang,\nKab. Bandung, Jawa Barat", "", "C", false)
	pdf.SetXY(x+logoW, y+12)
	pdf.SetFont("Times", "B", 17)
	pdf.CellFormat(usableW-logoW, 8, "E-RESUME RAWAT JALAN", "", 0, "C", false, 0, "")
	pdf.SetY(y + headerH)

	rowOutpatientInfo(pdf, []pdfCell{
		{Label: "Nama Pasien", Value: doc.PatientName, W: 74},
		{Label: "Tgl. Lahir", Value: doc.BirthDate, W: 35},
		{Label: "Jenis\nKelamin", Value: genderShort(doc.Gender), W: 28},
		{Label: "No MR.", Value: doc.NoRM, W: usableW - 74 - 35 - 28},
	})
	rowOutpatientInfo(pdf, []pdfCell{
		{Label: "Alamat Lengkap", Value: doc.Address, W: 109},
		{Label: "Poli", Value: doc.Polyclinic, W: 28},
		{Label: "No Telp", Value: doc.PhoneNumber, W: usableW - 109 - 28},
	})
	rowOutpatientInfo(pdf, []pdfCell{
		{Label: "Tanggal Masuk", Value: doc.VisitDate, W: usableW / 2},
		{Label: "Tanggal Keluar", Value: doc.DischargeDate, W: usableW / 2},
	})

	resumeRow(pdf, "ANAMNESA", doc.Anamnesis, usableW, 15)
	resumeRow(pdf, "PEMERIKSAAN FISIK", doc.PhysicalExam, usableW, 16)
	resumeRow(pdf, "ASUHAN KEPERAWATAN", doc.NursingCare, usableW, 12)
	diagnosisRow(pdf, "DIAGNOSA UTAMA", doc.MainDiagnosis, "KODE ICD X", doc.MainICD, usableW, 25)
	diagnosisRow(pdf, "DIAGNOSA TAMBAHAN", doc.AdditionalDiagnosis, "KODE ICD X", doc.AdditionalICD, usableW, 25)
	diagnosisRow(pdf, "TINDAKAN", doc.Action, "KODE ICD IX", doc.ActionICD, usableW, 28)
	resumeRow(pdf, "PENGOBATAN", doc.Medication, usableW, 22)

	renderOutpatientDischargeSignature(pdf, doc)
}

func renderInpatientResume(pdf *fpdf.Fpdf, doc *repository.MobilePatientMedicalResumeDocument) {
	pageW, _ := pdf.GetPageSize()
	left, _, right, _ := pdf.GetMargins()
	usableW := pageW - left - right

	x := pdf.GetX()
	y := pdf.GetY()
	drawLogo(pdf, x+23, y+4, 10)
	pdf.SetY(y + 16)
	pdf.SetFont("Times", "B", 8.5)
	pdf.MultiCell(56, 3.8, "RSUD OTO ISKANDAR DI\nNATA", "", "C", false)
	pdf.SetFont("Times", "", 5.4)
	pdf.SetX(x)
	pdf.MultiCell(56, 2.8, "Jl. Gading Tutuka, RT. 01 RW. 01,\nKp. Cincin Kolot, Kec. Soreang, Kab. Bandung\n(022)5891355\nLaman: rsudotista.bandungkab.go.id / Email:\nrsudotista@bandungkab.go.id", "", "C", false)

	pdf.SetXY(x+60, y+10)
	pdf.SetFont("Times", "B", 16)
	pdf.MultiCell(60, 8, "RESUME MEDIS\nPASIEN PULANG", "", "C", false)

	boxX := x + usableW - 68
	pdf.RoundedRect(boxX, y+8, 68, 24, 3, "1234", "")
	pdf.SetXY(boxX+3, y+11)
	pdf.SetFont("Times", "", 9)
	pdf.MultiCell(62, 6, "Nama : "+pdfText(doc.PatientName)+"\nTanggal Lahir : "+pdfText(doc.BirthDate)+"\nNo. RM : "+pdfText(doc.NoRM), "", "L", false)
	pdf.SetY(y + 44)

	infoY := pdf.GetY()
	pdf.Rect(x, infoY, usableW, 15, "")
	pdf.SetFont("Times", "", 9)
	pdf.SetXY(x+2, infoY+2)
	pdf.CellFormat(34, 5, "Tanggal masuk", "", 0, "L", false, 0, "")
	pdf.CellFormat(4, 5, ":", "", 0, "L", false, 0, "")
	pdf.CellFormat(55, 5, pdfText(doc.VisitDate), "", 0, "L", false, 0, "")
	pdf.CellFormat(34, 5, "Tanggal kontrol", "", 0, "L", false, 0, "")
	pdf.CellFormat(4, 5, ":", "", 0, "L", false, 0, "")
	pdf.CellFormat(50, 5, pdfText(doc.ControlDate), "", 1, "L", false, 0, "")
	pdf.SetX(x + 2)
	pdf.CellFormat(34, 5, "Tanggal keluar", "", 0, "L", false, 0, "")
	pdf.CellFormat(4, 5, ":", "", 0, "L", false, 0, "")
	pdf.CellFormat(55, 5, pdfText(doc.DischargeDate), "", 0, "L", false, 0, "")
	pdf.CellFormat(34, 5, "Tujuan poliklinik", "", 0, "L", false, 0, "")
	pdf.CellFormat(4, 5, ":", "", 0, "L", false, 0, "")
	pdf.CellFormat(50, 5, pdfText(doc.ControlPolyclinic), "", 1, "L", false, 0, "")
	pdf.SetY(infoY + 15)

	startY := pdf.GetY()
	pdf.Rect(x, startY, usableW, 160, "")
	pdf.SetXY(x+2, startY+2)
	inpatientSection(pdf, "INDIKASI RAWAT INAP", doc.InpatientIndication)
	inpatientSection(pdf, "RINGKASAN RIWAYAT PENYAKIT", doc.DiseaseHistory)
	inpatientSection(pdf, "PEMERIKSAAN FISIK", doc.PhysicalExam)
	inpatientSection(pdf, "PEMERIKSAAN PENUNJANG", doc.SupportingExam)
	inpatientSection(pdf, "TERAPI / PENGOBATAN SELAMA DI RUMAH SAKIT", doc.Therapy)
	inpatientSection(pdf, "CATATAN", doc.Note)
	inpatientDiagnosis(pdf, "DIAGNOSA UTAMA", doc.MainDiagnosis, "ICD X", doc.MainICD, usableW)
	inpatientDiagnosis(pdf, "DIAGNOSA TAMBAHAN", doc.AdditionalDiagnosis, "ICD X", doc.AdditionalICD, usableW)
	inpatientDiagnosis(pdf, "TINDAKAN / PROSEDUR / OPERASI", doc.Action, "ICD IX", doc.ActionICD, usableW)
	inpatientSection(pdf, "INSTRUKSI PERAWAT LANJUTAN / EDUKASI", doc.FollowUpInstruction)

	renderInpatientDischargeSignature(pdf, doc)
}

func renderOutpatientDischargeSignature(pdf *fpdf.Fpdf, doc *repository.MobilePatientMedicalResumeDocument) {
	pdf.AddPage()
	pageW, _ := pdf.GetPageSize()
	left, _, right, _ := pdf.GetMargins()
	usableW := pageW - left - right
	x := pdf.GetX()
	y := pdf.GetY()
	rowH := 34.0
	labelW := 70.0

	pdf.Rect(x, y, usableW, rowH, "")
	pdf.Line(x+labelW, y, x+labelW, y+rowH)
	pdf.SetFont("Times", "B", 9)
	pdf.SetXY(x+4, y+rowH/2-2)
	pdf.CellFormat(labelW-8, 5, "CARA PULANG", "", 0, "L", false, 0, "")

	options := []string{
		"Kontrol ulang RS",
		"Kontrol PRB",
		"Dirawat",
		"Dirujuk",
		"Konsultasi selesai / tidak kontrol ulang",
		"Pulang Paksa",
		"Meninggal",
	}
	pdf.SetFont("Times", "", 8.5)
	optionX := x + labelW + 5
	optionY := y + 4
	for i, option := range options {
		drawCheckbox(pdf, optionX, optionY+float64(i)*4.1, checkboxMatches(doc.DischargeMethod, option))
		pdf.SetXY(optionX+4, optionY+float64(i)*4.1-0.7)
		pdf.CellFormat(90, 4, option, "", 0, "L", false, 0, "")
	}

	signY := y + 55
	drawSignatureBlock(pdf, x+30, signY, 58, "Tanda Tangan Pasien / Keluarga atau Wali", doc.PatientName)
	drawSignatureBlock(pdf, x+usableW-75, signY, 58, "Dokter", doc.Doctor)
}

func renderInpatientDischargeSignature(pdf *fpdf.Fpdf, doc *repository.MobilePatientMedicalResumeDocument) {
	pdf.AddPage()
	pageW, _ := pdf.GetPageSize()
	left, _, right, _ := pdf.GetMargins()
	usableW := pageW - left - right
	x := pdf.GetX()
	y := pdf.GetY()

	boxH := 60.0
	pdf.Rect(x, y, usableW, boxH, "")
	pdf.SetXY(x+2, y+3)
	pdf.SetFont("Times", "", 9)
	pdf.MultiCell(usableW-4, 5, pdfText(inpatientFollowUpText(doc)), "", "L", false)

	pdf.SetXY(x+2, y+14)
	pdf.CellFormat(32, 5, "Cara Pulang", "", 0, "L", false, 0, "")
	pdf.CellFormat(4, 5, ":", "", 0, "L", false, 0, "")
	dischargeOptions := []string{"Izin dokter", "Pindah RS", "APS", "Melarikan Diri"}
	optionX := x + 38
	for _, option := range dischargeOptions {
		drawCheckbox(pdf, optionX, y+16, checkboxMatches(doc.DischargeMethod, option))
		pdf.SetXY(optionX+4, y+14.7)
		pdf.CellFormat(23, 5, option, "", 0, "L", false, 0, "")
		optionX += 27
	}

	pdf.SetXY(x+2, y+31)
	pdf.CellFormat(42, 5, "Kondisi Saat Pulang", "", 0, "L", false, 0, "")
	pdf.CellFormat(4, 5, ":", "", 0, "L", false, 0, "")
	conditionOptions := []string{"Sembuh", "Perbaikan", "Tidak Sembuh", "Meninggal"}
	optionX = x + 4
	for i, option := range conditionOptions {
		if i == 0 {
			optionX = x + 4
		}
		drawCheckbox(pdf, optionX, y+39, checkboxMatches(doc.DischargeCondition, option))
		pdf.SetXY(optionX+4, y+37.7)
		pdf.CellFormat(25, 5, option, "", 0, "L", false, 0, "")
		optionX += 34
	}

	pdf.SetXY(x+2, y+48)
	pdf.SetFont("Times", "B", 10)
	pdf.CellFormat(0, 5, "TERAPI PULANG :", "", 1, "L", false, 0, "")
	pdf.SetFont("Times", "", 9)
	pdf.SetX(x + 2)
	pdf.MultiCell(usableW-4, 4.2, pdfText(coalescePDFText(doc.Medication, doc.Therapy)), "", "L", false)

	signY := y + 75
	pdf.SetFont("Times", "", 9)
	pdf.SetXY(x+usableW-55, signY-12)
	pdf.CellFormat(50, 5, "Soreang, "+resumeDate(doc), "", 0, "C", false, 0, "")
	drawSignatureBlock(pdf, x+38, signY, 62, "Tanda Tangan Pasien / Keluarga atau Wali", doc.PatientName)
	drawDoctorSignatureBlock(pdf, x+usableW-78, signY, 62, "Dokter penanggung jawab", doc.Doctor)
}

type pdfCell struct {
	Label string
	Value string
	W     float64
}

func rowOutpatientInfo(pdf *fpdf.Fpdf, cells []pdfCell) {
	x := pdf.GetX()
	y := pdf.GetY()
	rowH := 17.0
	offset := x
	pdf.SetFont("Times", "B", 8.5)
	for _, cell := range cells {
		pdf.Rect(offset, y, cell.W, rowH, "")
		pdf.SetXY(offset+4, y+4)
		pdf.MultiCell(cell.W-8, 4, pdfText(cell.Label), "", "L", false)
		pdf.SetFont("Times", "", 8.5)
		pdf.SetXY(offset+4, y+9)
		pdf.MultiCell(cell.W-8, 4, pdfText(cell.Value), "", "L", false)
		pdf.SetFont("Times", "B", 8.5)
		offset += cell.W
	}
	pdf.SetXY(x, y+rowH)
}

func resumeRow(pdf *fpdf.Fpdf, label string, value string, width float64, minH float64) {
	x := pdf.GetX()
	y := pdf.GetY()
	labelW := 55.0
	textW := width - labelW
	height := rowHeight(pdf, value, textW-8, minH)
	pdf.Rect(x, y, labelW, height, "")
	pdf.Rect(x+labelW, y, textW, height, "")
	pdf.SetFont("Times", "B", 8.5)
	pdf.SetXY(x+4, y+5)
	pdf.MultiCell(labelW-8, 4, label, "", "L", false)
	pdf.SetFont("Times", "", 8.5)
	pdf.SetXY(x+labelW+4, y+5)
	pdf.MultiCell(textW-8, 4, pdfText(value), "", "L", false)
	pdf.SetXY(x, y+height)
}

func diagnosisRow(pdf *fpdf.Fpdf, label string, value string, codeLabel string, codeValue string, width float64, minH float64) {
	x := pdf.GetX()
	y := pdf.GetY()
	labelW := 55.0
	valueW := 55.0
	codeLabelW := 28.0
	codeValueW := width - labelW - valueW - codeLabelW
	height := maxFloat(minH, rowHeight(pdf, value+"\n"+codeValue, maxFloat(valueW, codeValueW)-8, minH))
	pdf.Rect(x, y, labelW, height, "")
	pdf.Rect(x+labelW, y, valueW, height, "")
	pdf.Rect(x+labelW+valueW, y, codeLabelW, height, "")
	pdf.Rect(x+labelW+valueW+codeLabelW, y, codeValueW, height, "")
	pdf.SetFont("Times", "B", 8.5)
	pdf.SetXY(x+4, y+height/2-3)
	pdf.MultiCell(labelW-8, 4, label, "", "L", false)
	pdf.SetXY(x+labelW+valueW+4, y+5)
	pdf.MultiCell(codeLabelW-8, 4, codeLabel, "", "L", false)
	pdf.SetFont("Times", "", 8.5)
	pdf.SetXY(x+labelW+4, y+5)
	pdf.MultiCell(valueW-8, 4, pdfText(value), "", "L", false)
	pdf.SetXY(x+labelW+valueW+codeLabelW+4, y+5)
	pdf.MultiCell(codeValueW-8, 4, pdfText(codeValue), "", "L", false)
	pdf.SetXY(x, y+height)
}

func inpatientSection(pdf *fpdf.Fpdf, label string, value string) {
	pdf.SetFont("Times", "B", 10)
	pdf.CellFormat(0, 5, label+" :", "", 1, "L", false, 0, "")
	pdf.SetFont("Times", "", 9)
	pdf.MultiCell(0, 4.3, pdfText(value), "", "L", false)
	pdf.Ln(1)
}

func inpatientDiagnosis(pdf *fpdf.Fpdf, label string, value string, codeLabel string, codeValue string, width float64) {
	x := pdf.GetX()
	y := pdf.GetY()
	leftW := width * 0.48
	rightW := width - leftW
	pdf.SetFont("Times", "B", 10)
	pdf.SetXY(x, y)
	pdf.CellFormat(leftW, 5, label+" :", "", 0, "L", false, 0, "")
	pdf.CellFormat(rightW, 5, codeLabel+" :", "", 1, "L", false, 0, "")
	pdf.SetFont("Times", "", 9)
	y = pdf.GetY()
	pdf.SetXY(x, y)
	pdf.MultiCell(leftW-4, 4.3, pdfText(value), "", "L", false)
	leftY := pdf.GetY()
	pdf.SetXY(x+leftW, y)
	pdf.MultiCell(rightW-4, 4.3, pdfText(codeValue), "", "L", false)
	rightY := pdf.GetY()
	pdf.SetY(maxFloat(leftY, rightY) + 1)
}

func drawLogo(pdf *fpdf.Fpdf, x float64, y float64, w float64) {
	for _, path := range []string{
		"/app/assets/rsud_mark_source.png",
		"assets/rsud_mark_source.png",
		"/app/assets/logo_mark.png",
		"assets/logo_mark.png",
	} {
		if _, err := os.Stat(path); err == nil {
			pdf.ImageOptions(path, x, y, w, 0, false, fpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}, 0, "")
			return
		}
	}
}

func drawCheckbox(pdf *fpdf.Fpdf, x float64, y float64, checked bool) {
	size := 3.1
	pdf.Rect(x, y, size, size, "")
	if !checked {
		return
	}
	pdf.Line(x+0.5, y+1.7, x+1.25, y+2.6)
	pdf.Line(x+1.25, y+2.6, x+2.75, y+0.55)
}

func drawSignatureBlock(pdf *fpdf.Fpdf, x float64, y float64, w float64, title string, name string) {
	pdf.SetFont("Times", "", 9)
	pdf.SetXY(x, y)
	pdf.MultiCell(w, 5, title, "", "C", false)
	pdf.SetXY(x, y+43)
	pdf.SetFont("Times", "", 9)
	pdf.MultiCell(w, 5, pdfText(name), "", "C", false)
}

func drawDoctorSignatureBlock(pdf *fpdf.Fpdf, x float64, y float64, w float64, title string, name string) {
	pdf.SetFont("Times", "", 9)
	pdf.SetXY(x, y)
	pdf.MultiCell(w, 5, title, "", "C", false)
	qrX := x + w/2 - 10
	qrY := y + 18
	pdf.Rect(qrX, qrY, 20, 20, "")
	drawLogo(pdf, qrX+6, qrY+6, 8)
	pdf.SetXY(x, y+43)
	pdf.MultiCell(w, 5, pdfText(name), "", "C", false)
}

func rowHeight(pdf *fpdf.Fpdf, value string, width float64, minH float64) float64 {
	lines := pdf.SplitLines([]byte(pdfText(value)), width)
	height := float64(len(lines))*4 + 10
	if height < minH {
		return minH
	}
	return height
}

func checkboxMatches(value string, option string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	option = strings.ToLower(strings.TrimSpace(option))
	if value == "" || value == "-" {
		return false
	}

	switch option {
	case "kontrol ulang rs":
		return strings.Contains(value, "kontrol") || strings.Contains(value, "izin dokter")
	case "dirawat":
		return strings.Contains(value, "dirawat") || strings.Contains(value, "rawat")
	case "izin dokter":
		return strings.Contains(value, "izin dokter") || strings.Contains(value, "kontrol")
	case "pindah rs":
		return strings.Contains(value, "pindah") || strings.Contains(value, "rujuk")
	case "aps":
		return strings.Contains(value, "aps") || strings.Contains(value, "paksa")
	case "melarikan diri":
		return strings.Contains(value, "lari") || strings.Contains(value, "melarikan")
	case "perbaikan":
		return strings.Contains(value, "perbaikan") || strings.Contains(value, "membaik")
	case "tidak sembuh":
		return strings.Contains(value, "tidak sembuh")
	default:
		return strings.Contains(value, option)
	}
}

func coalescePDFText(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" && value != "-" {
			return value
		}
	}
	return "-"
}

func inpatientFollowUpText(doc *repository.MobilePatientMedicalResumeDocument) string {
	if strings.TrimSpace(doc.FollowUpInstruction) != "" && strings.TrimSpace(doc.FollowUpInstruction) != "-" {
		return doc.FollowUpInstruction
	}
	if strings.TrimSpace(doc.ControlPolyclinic) != "" && strings.TrimSpace(doc.ControlPolyclinic) != "-" {
		return "kontrol " + strings.ToLower(doc.ControlPolyclinic)
	}
	return "-"
}

func resumeDate(doc *repository.MobilePatientMedicalResumeDocument) string {
	value := strings.TrimSpace(doc.DischargeDate)
	if value == "" || value == "-" {
		return time.Now().Format("02-01-2006")
	}
	parts := strings.Fields(value)
	if len(parts) > 0 {
		return parts[0]
	}
	return value
}

func medicalResumeTitle(doc *repository.MobilePatientMedicalResumeDocument) string {
	if doc.IsInpatient {
		return "Resume Medis Pasien Pulang"
	}
	return "E-Resume Rawat Jalan"
}

func pdfText(value string) string {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	value = strings.TrimSpace(value)
	if value == "" {
		return "-"
	}
	return value
}

func genderShort(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case "laki-laki", "laki laki", "l":
		return "L"
	case "perempuan", "p":
		return "P"
	default:
		return value
	}
}

func maxFloat(a float64, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
