package handlers

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"kas-app/db"
	"kas-app/models"
	"net/http"
	"strconv"
	"strings"
)

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// -------------------------------- STUDENT HANDLERS --------------------------------

// GetStudents handles GET /api/students
func GetStudents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	search := r.URL.Query().Get("search")

	students, err := db.GetAllStudents(search)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Gagal mengambil data siswa: "+err.Error())
		return
	}
	if students == nil {
		students = []models.Student{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"success": true, "data": students, "total": len(students)})
}

// CreateStudent handles POST /api/students
func CreateStudent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	var input models.StudentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "Format JSON tidak valid: "+err.Error())
		return
	}
	if strings.TrimSpace(input.Name) == "" {
		writeError(w, http.StatusBadRequest, "Nama siswa wajib diisi")
		return
	}
	id, err := db.CreateStudent(input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Gagal menyimpan siswa: "+err.Error())
		return
	}
	created, _ := db.GetStudentByID(id)
	writeJSON(w, http.StatusCreated, map[string]interface{}{"success": true, "message": "Data siswa berhasil ditambahkan", "data": created})
}

// UpdateStudent handles PUT /api/students/{id}
func UpdateStudent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	idStr := strings.TrimPrefix(r.URL.Path, "/api/students/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ID tidak valid")
		return
	}
	var input models.StudentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "Format JSON tidak valid: "+err.Error())
		return
	}
	if strings.TrimSpace(input.Name) == "" {
		writeError(w, http.StatusBadRequest, "Nama siswa wajib diisi")
		return
	}
	if err := db.UpdateStudent(id, input); err != nil {
		writeError(w, http.StatusInternalServerError, "Gagal memperbarui siswa: "+err.Error())
		return
	}
	updated, _ := db.GetStudentByID(id)
	writeJSON(w, http.StatusOK, map[string]interface{}{"success": true, "message": "Data siswa berhasil diperbarui", "data": updated})
}

// DeleteStudent handles DELETE /api/students/{id}
func DeleteStudent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	idStr := strings.TrimPrefix(r.URL.Path, "/api/students/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ID tidak valid")
		return
	}
	if err := db.DeleteStudent(id); err != nil {
		writeError(w, http.StatusInternalServerError, "Gagal menghapus siswa: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"success": true, "message": "Data siswa berhasil dihapus"})
}

// -------------------------------- TRANSACTION HANDLERS --------------------------------

// GetTransactions handles GET /api/transactions
func GetTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	q := r.URL.Query()
	txns, err := db.GetAllTransactions(q.Get("start_date"), q.Get("end_date"), q.Get("search"), q.Get("pos"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Gagal mengambil transaksi: "+err.Error())
		return
	}
	if txns == nil {
		txns = []models.Transaction{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"success": true, "data": txns, "total": len(txns)})
}

// CreateTransaction handles POST /api/transactions
func CreateTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	var input models.TransactionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "Format JSON tidak valid: "+err.Error())
		return
	}
	if strings.TrimSpace(input.Description) == "" {
		writeError(w, http.StatusBadRequest, "Keterangan transaksi wajib diisi")
		return
	}
	id, err := db.CreateTransaction(input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Gagal menyimpan transaksi: "+err.Error())
		return
	}
	created, _ := db.GetTransactionByID(id)
	writeJSON(w, http.StatusCreated, map[string]interface{}{"success": true, "message": "Transaksi berhasil disimpan", "data": created})
}

// UpdateTransaction handles PUT /api/transactions/{id}
func UpdateTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	idStr := strings.TrimPrefix(r.URL.Path, "/api/transactions/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ID tidak valid")
		return
	}
	var input models.TransactionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "Format JSON tidak valid: "+err.Error())
		return
	}
	if strings.TrimSpace(input.Description) == "" {
		writeError(w, http.StatusBadRequest, "Keterangan transaksi wajib diisi")
		return
	}
	if err := db.UpdateTransaction(id, input); err != nil {
		writeError(w, http.StatusInternalServerError, "Gagal memperbarui transaksi: "+err.Error())
		return
	}
	updated, _ := db.GetTransactionByID(id)
	writeJSON(w, http.StatusOK, map[string]interface{}{"success": true, "message": "Transaksi berhasil diperbarui", "data": updated})
}

// DeleteTransaction handles DELETE /api/transactions/{id}
func DeleteTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	idStr := strings.TrimPrefix(r.URL.Path, "/api/transactions/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ID tidak valid")
		return
	}
	if err := db.DeleteTransaction(id); err != nil {
		writeError(w, http.StatusInternalServerError, "Gagal menghapus transaksi: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"success": true, "message": "Transaksi berhasil dihapus"})
}

// GetSummary handles GET /api/summary
func GetSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	q := r.URL.Query()
	summary, err := db.GetSummary(q.Get("start_date"), q.Get("end_date"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Gagal menghitung ringkasan: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"success": true, "data": summary})
}

// GetChartData handles GET /api/chart-data
func GetChartData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	chartData, err := db.GetChartData()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Gagal mengambil data grafik: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"success": true, "data": chartData})
}

// ExportCSV handles GET /api/export/csv
func ExportCSV(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	q := r.URL.Query()
	txns, err := db.GetAllTransactions(q.Get("start_date"), q.Get("end_date"), q.Get("search"), q.Get("pos"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Gagal export: "+err.Error())
		return
	}

	filename := "buku_kas_ikrom.csv"
	if q.Get("start_date") != "" && q.Get("end_date") != "" {
		filename = fmt.Sprintf("buku_kas_%s_sd_%s.csv", q.Get("start_date"), q.Get("end_date"))
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Write([]byte("\xEF\xBB\xBF")) // UTF-8 BOM for Excel

	writer := csv.NewWriter(w)
	defer writer.Flush()

	_ = writer.Write([]string{"No", "Tanggal", "No. Bukti", "Nama", "Keterangan",
		"Kas Masuk", "Kas Keluar", "Saldo Kas",
		"Ikrom Masuk", "Ikrom Keluar", "Saldo Ikrom", "Total Saldo"})

	var totalKasIn, totalKasOut, totalIkromIn, totalIkromOut float64
	for i, t := range txns {
		totalKasIn += t.KasIn
		totalKasOut += t.KasOut
		totalIkromIn += t.IkromIn
		totalIkromOut += t.IkromOut
		_ = writer.Write([]string{
			strconv.Itoa(i + 1), t.Date, t.RefNo, t.Name, t.Description,
			fmt.Sprintf("%.0f", t.KasIn), fmt.Sprintf("%.0f", t.KasOut), fmt.Sprintf("%.0f", t.KasBalance),
			fmt.Sprintf("%.0f", t.IkromIn), fmt.Sprintf("%.0f", t.IkromOut), fmt.Sprintf("%.0f", t.IkromBalance),
			fmt.Sprintf("%.0f", t.TotalBalance),
		})
	}
	_ = writer.Write([]string{"TOTAL", "", "", "", "TOTAL AKUMULASI",
		fmt.Sprintf("%.0f", totalKasIn), fmt.Sprintf("%.0f", totalKasOut), fmt.Sprintf("%.0f", totalKasIn-totalKasOut),
		fmt.Sprintf("%.0f", totalIkromIn), fmt.Sprintf("%.0f", totalIkromOut), fmt.Sprintf("%.0f", totalIkromIn-totalIkromOut),
		fmt.Sprintf("%.0f", (totalKasIn-totalKasOut)+(totalIkromIn-totalIkromOut)),
	})
}

// SeedData handles POST /api/seed
func SeedData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	if err := db.SeedSampleData(); err != nil {
		writeError(w, http.StatusInternalServerError, "Gagal seed data: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"success": true, "message": "Data contoh berhasil dimuat"})
}

// ResetData handles POST /api/reset
func ResetData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	if err := db.ResetAllData(); err != nil {
		writeError(w, http.StatusInternalServerError, "Gagal reset data: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"success": true, "message": "Semua transaksi berhasil dihapus"})
}
