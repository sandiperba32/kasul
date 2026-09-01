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

// Helper to write JSON responses
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// Helper to write JSON error responses
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// GetTransactions handles GET /api/transactions
func GetTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	q := r.URL.Query()
	startDate := q.Get("start_date")
	endDate := q.Get("end_date")
	category := q.Get("category")
	search := q.Get("search")
	pos := q.Get("pos")

	transactions, err := db.GetAllTransactions(startDate, endDate, category, search, pos)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Gagal mengambil data transaksi: "+err.Error())
		return
	}

	if transactions == nil {
		transactions = []models.Transaction{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    transactions,
		"total":   len(transactions),
	})
}

// GetTransactionByID handles GET /api/transactions/{id}
func GetTransactionByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/transactions/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ID transaksi tidak valid")
		return
	}

	t, err := db.GetTransactionByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "Transaksi tidak ditemukan")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    t,
	})
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
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"message": "Transaksi berhasil disimpan",
		"data":    created,
	})
}

// UpdateTransaction handles PUT /api/transactions/{id}
func UpdateTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/transactions/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ID transaksi tidak valid")
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
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Transaksi berhasil diperbarui",
		"data":    updated,
	})
}

// DeleteTransaction handles DELETE /api/transactions/{id}
func DeleteTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/transactions/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ID transaksi tidak valid")
		return
	}

	if err := db.DeleteTransaction(id); err != nil {
		writeError(w, http.StatusInternalServerError, "Gagal menghapus transaksi: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Transaksi berhasil dihapus",
	})
}

// GetSummary handles GET /api/summary
func GetSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	q := r.URL.Query()
	startDate := q.Get("start_date")
	endDate := q.Get("end_date")

	summary, err := db.GetSummary(startDate, endDate)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Gagal menghitung ringkasan kas: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    summary,
	})
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

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    chartData,
	})
}

// CategoriesHandler handles GET & POST /api/categories
func CategoriesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		cats, err := db.GetCategories()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Gagal mengambil kategori: "+err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"data":    cats,
		})
	case http.MethodPost:
		var c models.Category
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			writeError(w, http.StatusBadRequest, "Format JSON tidak valid")
			return
		}
		if strings.TrimSpace(c.Name) == "" {
			writeError(w, http.StatusBadRequest, "Nama kategori wajib diisi")
			return
		}
		if c.Type == "" {
			c.Type = "both"
		}
		if c.Pos == "" {
			c.Pos = "all"
		}
		id, err := db.CreateCategory(c)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Gagal menambah kategori: "+err.Error())
			return
		}
		c.ID = id
		writeJSON(w, http.StatusCreated, map[string]interface{}{
			"success": true,
			"message": "Kategori berhasil ditambahkan",
			"data":    c,
		})
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// DeleteCategoryHandler handles DELETE /api/categories/{id}
func DeleteCategoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/categories/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ID kategori tidak valid")
		return
	}

	if err := db.DeleteCategory(id); err != nil {
		writeError(w, http.StatusInternalServerError, "Gagal menghapus kategori: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Kategori berhasil dihapus",
	})
}

// ExportCSV handles GET /api/export/csv
func ExportCSV(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	q := r.URL.Query()
	startDate := q.Get("start_date")
	endDate := q.Get("end_date")
	category := q.Get("category")
	search := q.Get("search")
	pos := q.Get("pos")

	transactions, err := db.GetAllTransactions(startDate, endDate, category, search, pos)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Gagal mengekspor data: "+err.Error())
		return
	}

	filename := "buku_kas_tabelar.csv"
	if startDate != "" && endDate != "" {
		filename = fmt.Sprintf("buku_kas_%s_sd_%s.csv", startDate, endDate)
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	// UTF-8 BOM for Excel compatibility in Windows
	w.Write([]byte("\xEF\xBB\xBF"))

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// CSV Headers
	headers := []string{
		"No", "Tanggal", "No. Bukti", "Keterangan", "Kategori",
		"Kas Masuk", "Kas Keluar", "Saldo Kas",
		"Ikrom Masuk", "Ikrom Keluar", "Saldo Ikrom",
		"Pen Masuk", "Pen Keluar", "Saldo Pen",
		"Total Saldo Akumulasi",
	}
	_ = writer.Write(headers)

	var totalKasIn, totalKasOut float64
	var totalIkromIn, totalIkromOut float64
	var totalPenIn, totalPenOut float64

	for i, t := range transactions {
		totalKasIn += t.KasIn
		totalKasOut += t.KasOut
		totalIkromIn += t.IkromIn
		totalIkromOut += t.IkromOut
		totalPenIn += t.PenIn
		totalPenOut += t.PenOut

		row := []string{
			strconv.Itoa(i + 1),
			t.Date,
			t.RefNo,
			t.Description,
			t.Category,
			fmt.Sprintf("%.0f", t.KasIn),
			fmt.Sprintf("%.0f", t.KasOut),
			fmt.Sprintf("%.0f", t.KasBalance),
			fmt.Sprintf("%.0f", t.IkromIn),
			fmt.Sprintf("%.0f", t.IkromOut),
			fmt.Sprintf("%.0f", t.IkromBalance),
			fmt.Sprintf("%.0f", t.PenIn),
			fmt.Sprintf("%.0f", t.PenOut),
			fmt.Sprintf("%.0f", t.PenBalance),
			fmt.Sprintf("%.0f", t.TotalBalance),
		}
		_ = writer.Write(row)
	}

	// Summary Footer Row
	finalKasBal := totalKasIn - totalKasOut
	finalIkromBal := totalIkromIn - totalIkromOut
	finalPenBal := totalPenIn - totalPenOut
	finalTotalBal := finalKasBal + finalIkromBal + finalPenBal

	footerRow := []string{
		"TOTAL", "", "", "TOTAL AKUMULASI PERIODE INI", "",
		fmt.Sprintf("%.0f", totalKasIn),
		fmt.Sprintf("%.0f", totalKasOut),
		fmt.Sprintf("%.0f", finalKasBal),
		fmt.Sprintf("%.0f", totalIkromIn),
		fmt.Sprintf("%.0f", totalIkromOut),
		fmt.Sprintf("%.0f", finalIkromBal),
		fmt.Sprintf("%.0f", totalPenIn),
		fmt.Sprintf("%.0f", totalPenOut),
		fmt.Sprintf("%.0f", finalPenBal),
		fmt.Sprintf("%.0f", finalTotalBal),
	}
	_ = writer.Write(footerRow)
}

// SeedData handles POST /api/seed
func SeedData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if err := db.SeedSampleData(); err != nil {
		writeError(w, http.StatusInternalServerError, "Gagal mengisi data contoh: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Data contoh Kas, Ikrom, dan Pen berhasil dimasukkan",
	})
}

// ResetData handles POST /api/reset
func ResetData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if err := db.ResetAllData(); err != nil {
		writeError(w, http.StatusInternalServerError, "Gagal mereset data: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Semua data transaksi berhasil dibersihkan",
	})
}
