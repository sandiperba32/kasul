package handlers

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"kas-app/db"
	"kas-app/models"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ─────────────────────────────────────────────
//  JSON response helpers
// ─────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// ─────────────────────────────────────────────
//  Auth: server-side password verification
//  Password is NEVER sent to the client.
// ─────────────────────────────────────────────

const appPassword = "781232"

// sessionStore holds valid session tokens in memory.
// Token -> expiry time. Sessions last 8 hours.
var (
	sessionStore   = make(map[string]time.Time)
	sessionMu      sync.RWMutex
	sessionTTL     = 8 * time.Hour
)

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func isValidSession(token string) bool {
	sessionMu.RLock()
	exp, ok := sessionStore[token]
	sessionMu.RUnlock()
	return ok && time.Now().Before(exp)
}

// Login verifies password and returns a session token on success.
// POST /api/auth/login  { "password": "..." }
func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	var body struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "Bad request")
		return
	}

	// Use constant-time comparison to prevent timing attacks
	match := subtle.ConstantTimeCompare([]byte(body.Password), []byte(appPassword)) == 1
	if !match {
		// Small delay to slow brute-force
		time.Sleep(400 * time.Millisecond)
		writeJSON(w, http.StatusUnauthorized, map[string]interface{}{"success": false, "error": "Password salah"})
		return
	}

	token, err := generateToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Gagal membuat sesi")
		return
	}
	sessionMu.Lock()
	sessionStore[token] = time.Now().Add(sessionTTL)
	// Cleanup old sessions
	for k, v := range sessionStore {
		if time.Now().After(v) {
			delete(sessionStore, k)
		}
	}
	sessionMu.Unlock()

	writeJSON(w, http.StatusOK, map[string]interface{}{"success": true, "token": token})
}

// CheckSession validates a session token.
// POST /api/auth/check  { "token": "..." }
func CheckSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	var body struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "Bad request")
		return
	}
	valid := isValidSession(body.Token)
	writeJSON(w, http.StatusOK, map[string]interface{}{"success": valid})
}


func getKasIDFromQuery(r *http.Request) int64 {
	kasIDStr := r.URL.Query().Get("kas_id")
	if kasIDStr == "" {
		return 1
	}
	kasID, _ := strconv.ParseInt(kasIDStr, 10, 64)
	if kasID == 0 {
		return 1
	}
	return kasID
}

// -------------------------------- KAS BOOKS HANDLERS --------------------------------

func GetKasBooks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	books, err := db.GetAllKasBooks()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Gagal mengambil data buku kas: "+err.Error())
		return
	}
	if books == nil {
		books = []models.KasBook{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"success": true, "data": books, "total": len(books)})
}

func CreateKasBook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	var input models.KasBookInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "Format JSON tidak valid: "+err.Error())
		return
	}
	if strings.TrimSpace(input.Name) == "" {
		writeError(w, http.StatusBadRequest, "Nama buku kas wajib diisi")
		return
	}
	if input.ModelType == 0 {
		input.ModelType = 1
	}
	id, err := db.CreateKasBook(input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Gagal menyimpan buku kas: "+err.Error())
		return
	}
	created, _ := db.GetKasBookByID(id)
	writeJSON(w, http.StatusCreated, map[string]interface{}{"success": true, "message": "Buku kas berhasil ditambahkan", "data": created})
}

func UpdateKasBook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	idStr := strings.TrimPrefix(r.URL.Path, "/api/kas_books/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ID tidak valid")
		return
	}
	var input models.KasBookInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "Format JSON tidak valid: "+err.Error())
		return
	}
	if strings.TrimSpace(input.Name) == "" {
		writeError(w, http.StatusBadRequest, "Nama buku kas wajib diisi")
		return
	}
	if err := db.UpdateKasBook(id, input); err != nil {
		writeError(w, http.StatusInternalServerError, "Gagal memperbarui buku kas: "+err.Error())
		return
	}
	updated, _ := db.GetKasBookByID(id)
	writeJSON(w, http.StatusOK, map[string]interface{}{"success": true, "message": "Buku kas berhasil diperbarui", "data": updated})
}

func DeleteKasBook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	idStr := strings.TrimPrefix(r.URL.Path, "/api/kas_books/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ID tidak valid")
		return
	}
	if id == 1 {
		writeError(w, http.StatusBadRequest, "Buku kas default tidak bisa dihapus")
		return
	}
	if err := db.DeleteKasBook(id); err != nil {
		writeError(w, http.StatusInternalServerError, "Gagal menghapus buku kas: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"success": true, "message": "Buku kas berhasil dihapus"})
}

// -------------------------------- STUDENT HANDLERS --------------------------------

func GetStudents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	kasID := getKasIDFromQuery(r)
	search := r.URL.Query().Get("search")

	students, err := db.GetAllStudents(kasID, search)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Gagal mengambil data siswa: "+err.Error())
		return
	}
	if students == nil {
		students = []models.Student{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"success": true, "data": students, "total": len(students)})
}

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
	if input.KasID == 0 {
		input.KasID = 1
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
	if input.KasID == 0 {
		input.KasID = 1
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

func GetTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	q := r.URL.Query()
	kasID := getKasIDFromQuery(r)
	txns, err := db.GetAllTransactions(kasID, q.Get("start_date"), q.Get("end_date"), q.Get("search"), q.Get("pos"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Gagal mengambil transaksi: "+err.Error())
		return
	}
	if txns == nil {
		txns = []models.Transaction{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"success": true, "data": txns, "total": len(txns)})
}

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
	if input.KasID == 0 {
		input.KasID = 1
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
	if input.KasID == 0 {
		input.KasID = 1
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

func GetSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	q := r.URL.Query()
	kasID := getKasIDFromQuery(r)
	summary, err := db.GetSummary(kasID, q.Get("start_date"), q.Get("end_date"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Gagal menghitung ringkasan: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"success": true, "data": summary})
}

func GetChartData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	kasID := getKasIDFromQuery(r)
	chartData, err := db.GetChartData(kasID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Gagal mengambil data grafik: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"success": true, "data": chartData})
}

func ExportCSV(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	q := r.URL.Query()
	kasID := getKasIDFromQuery(r)
	txns, err := db.GetAllTransactions(kasID, q.Get("start_date"), q.Get("end_date"), q.Get("search"), q.Get("pos"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Gagal export: "+err.Error())
		return
	}

	filename := "buku_kas.csv"
	if q.Get("start_date") != "" && q.Get("end_date") != "" {
		filename = fmt.Sprintf("buku_kas_%s_sd_%s.csv", q.Get("start_date"), q.Get("end_date"))
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Write([]byte("\xEF\xBB\xBF")) // UTF-8 BOM for Excel

	writer := csv.NewWriter(w)
	defer writer.Flush()

	kasBook, _ := db.GetKasBookByID(kasID)
	isModel1 := kasBook == nil || kasBook.ModelType == 1

	if isModel1 {
		_ = writer.Write([]string{"No", "Tanggal", "No. Bukti", "Nama", "Keterangan",
			"Kas Masuk", "Kas Keluar", "Saldo Kas",
			"Ikrom Masuk", "Ikrom Keluar", "Saldo Ikrom", "Total Saldo"})
	} else {
		_ = writer.Write([]string{"No", "Tanggal", "No. Bukti", "Nama", "Keterangan",
			"Kas Masuk", "Kas Keluar", "Saldo Kas"})
	}

	var totalKasIn, totalKasOut, totalIkromIn, totalIkromOut float64
	for i, t := range txns {
		totalKasIn += t.KasIn
		totalKasOut += t.KasOut
		totalIkromIn += t.IkromIn
		totalIkromOut += t.IkromOut
		if isModel1 {
			_ = writer.Write([]string{
				strconv.Itoa(i + 1), t.Date, t.RefNo, t.Name, t.Description,
				fmt.Sprintf("%.0f", t.KasIn), fmt.Sprintf("%.0f", t.KasOut), fmt.Sprintf("%.0f", t.KasBalance),
				fmt.Sprintf("%.0f", t.IkromIn), fmt.Sprintf("%.0f", t.IkromOut), fmt.Sprintf("%.0f", t.IkromBalance),
				fmt.Sprintf("%.0f", t.TotalBalance),
			})
		} else {
			_ = writer.Write([]string{
				strconv.Itoa(i + 1), t.Date, t.RefNo, t.Name, t.Description,
				fmt.Sprintf("%.0f", t.KasIn), fmt.Sprintf("%.0f", t.KasOut), fmt.Sprintf("%.0f", t.KasBalance),
			})
		}
	}
	if isModel1 {
		_ = writer.Write([]string{"TOTAL", "", "", "", "TOTAL AKUMULASI",
			fmt.Sprintf("%.0f", totalKasIn), fmt.Sprintf("%.0f", totalKasOut), fmt.Sprintf("%.0f", totalKasIn-totalKasOut),
			fmt.Sprintf("%.0f", totalIkromIn), fmt.Sprintf("%.0f", totalIkromOut), fmt.Sprintf("%.0f", totalIkromIn-totalIkromOut),
			fmt.Sprintf("%.0f", (totalKasIn-totalKasOut)+(totalIkromIn-totalIkromOut)),
		})
	} else {
		_ = writer.Write([]string{"TOTAL", "", "", "", "TOTAL AKUMULASI",
			fmt.Sprintf("%.0f", totalKasIn), fmt.Sprintf("%.0f", totalKasOut), fmt.Sprintf("%.0f", totalKasIn-totalKasOut),
		})
	}
}

func SeedData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	var input struct {
		KasID int64 `json:"kas_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&input)
	if input.KasID == 0 {
		input.KasID = 1
	}

	if err := db.SeedSampleData(input.KasID); err != nil {
		writeError(w, http.StatusInternalServerError, "Gagal seed data: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"success": true, "message": "Data contoh berhasil dimuat"})
}

func ResetData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	var input struct {
		KasID int64 `json:"kas_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&input)
	if input.KasID == 0 {
		input.KasID = 1
	}
	if err := db.ResetAllData(input.KasID); err != nil {
		writeError(w, http.StatusInternalServerError, "Gagal reset data: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"success": true, "message": "Semua transaksi berhasil dihapus"})
}
