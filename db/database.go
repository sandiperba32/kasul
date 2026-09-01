package db

import (
	"database/sql"
	"fmt"
	"kas-app/models"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

// InitDB initializes SQLite connection, runs migrations, and inserts default categories
func InitDB(dataSourceName string) (*sql.DB, error) {
	var err error
	DB, err = sql.Open("sqlite", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool & SQLite pragmas
	DB.SetMaxOpenConns(1) // SQLite single writer safety
	_, err = DB.Exec(`
		PRAGMA journal_mode = WAL;
		PRAGMA synchronous = NORMAL;
		PRAGMA foreign_keys = ON;
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to set pragmas: %w", err)
	}

	if err := createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	if err := seedDefaultCategories(); err != nil {
		return nil, fmt.Errorf("failed to seed default categories: %w", err)
	}

	return DB, nil
}

func createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			type TEXT NOT NULL DEFAULT 'both', -- 'in', 'out', 'both'
			pos TEXT NOT NULL DEFAULT 'all'    -- 'kas', 'ikrom', 'all'
		);`,
		`CREATE TABLE IF NOT EXISTS transactions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			date TEXT NOT NULL,                -- YYYY-MM-DD
			ref_no TEXT DEFAULT '',
			description TEXT NOT NULL,
			category TEXT DEFAULT 'Umum',
			kas_in REAL NOT NULL DEFAULT 0.0,
			kas_out REAL NOT NULL DEFAULT 0.0,
			ikrom_in REAL NOT NULL DEFAULT 0.0,
			ikrom_out REAL NOT NULL DEFAULT 0.0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE INDEX IF NOT EXISTS idx_trans_date ON transactions(date);`,
		`CREATE INDEX IF NOT EXISTS idx_trans_category ON transactions(category);`,
	}

	for _, q := range queries {
		if _, err := DB.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

func seedDefaultCategories() error {
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM categories").Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	defaultCats := []struct {
		Name string
		Type string
		Pos  string
	}{
		{"Infaq / Shodaqoh", "in", "kas"},
		{"Donasi Khusus", "in", "kas"},
		{"Dana Bantuan Operasional", "in", "kas"},
		{"Penerimaan Ikrom / Bisaroh", "in", "ikrom"},
		{"Bisaroh Guru / Ustadz", "out", "ikrom"},
		{"Honor Karyawan / Pengurus", "out", "ikrom"},
		{"Operasional & Listrik / Air", "out", "kas"},
		{"Konsumsi & Jamuan", "out", "kas"},
		{"Pengadaan ATK & Perlengkapan", "out", "kas"},
		{"Pemeliharaan Fasilitas", "out", "kas"},
		{"Lain-lain", "both", "all"},
	}

	stmt, err := DB.Prepare("INSERT INTO categories (name, type, pos) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, c := range defaultCats {
		_, _ = stmt.Exec(c.Name, c.Type, c.Pos)
	}
	return nil
}

// GetInitialBalance calculates the balance of all transactions before a given date
func GetInitialBalance(beforeDate string) (kasBal, ikromBal float64, err error) {
	if beforeDate == "" {
		return 0, 0, nil
	}

	query := `
		SELECT 
			COALESCE(SUM(kas_in - kas_out), 0),
			COALESCE(SUM(ikrom_in - ikrom_out), 0)
		FROM transactions
		WHERE date < ?
	`
	err = DB.QueryRow(query, beforeDate).Scan(&kasBal, &ikromBal)
	return kasBal, ikromBal, err
}

// GetAllTransactions retrieves filtered transactions and calculates accurate running balances for Kas & Ikrom
func GetAllTransactions(startDate, endDate, category, search, posFilter string) ([]models.Transaction, error) {
	var conditions []string
	var args []interface{}

	if startDate != "" {
		conditions = append(conditions, "date >= ?")
		args = append(args, startDate)
	}
	if endDate != "" {
		conditions = append(conditions, "date <= ?")
		args = append(args, endDate)
	}
	if category != "" && category != "Semua" {
		conditions = append(conditions, "category = ?")
		args = append(args, category)
	}
	if search != "" {
		likePattern := "%" + search + "%"
		conditions = append(conditions, "(description LIKE ? OR ref_no LIKE ? OR category LIKE ?)")
		args = append(args, likePattern, likePattern, likePattern)
	}
	if posFilter != "" && posFilter != "all" {
		switch strings.ToLower(posFilter) {
		case "kas":
			conditions = append(conditions, "(kas_in > 0 OR kas_out > 0)")
		case "ikrom":
			conditions = append(conditions, "(ikrom_in > 0 OR ikrom_out > 0)")
		}
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Calculate initial balance prior to startDate
	initKas, initIkrom, err := GetInitialBalance(startDate)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`
		SELECT id, date, ref_no, description, category,
		       kas_in, kas_out, ikrom_in, ikrom_out,
		       created_at, updated_at
		FROM transactions
		%s
		ORDER BY date ASC, id ASC
	`, whereClause)

	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.Transaction
	runningKas := initKas
	runningIkrom := initIkrom

	for rows.Next() {
		var t models.Transaction
		var createdAt, updatedAt string
		err := rows.Scan(
			&t.ID, &t.Date, &t.RefNo, &t.Description, &t.Category,
			&t.KasIn, &t.KasOut, &t.IkromIn, &t.IkromOut,
			&createdAt, &updatedAt,
		)
		if err != nil {
			return nil, err
		}

		t.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		t.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)

		runningKas += (t.KasIn - t.KasOut)
		runningIkrom += (t.IkromIn - t.IkromOut)

		t.KasBalance = runningKas
		t.IkromBalance = runningIkrom
		t.TotalBalance = runningKas + runningIkrom

		transactions = append(transactions, t)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

// GetTransactionByID fetches a single transaction by ID
func GetTransactionByID(id int64) (*models.Transaction, error) {
	query := `
		SELECT id, date, ref_no, description, category,
		       kas_in, kas_out, ikrom_in, ikrom_out,
		       created_at, updated_at
		FROM transactions
		WHERE id = ?
	`
	var t models.Transaction
	var createdAt, updatedAt string
	err := DB.QueryRow(query, id).Scan(
		&t.ID, &t.Date, &t.RefNo, &t.Description, &t.Category,
		&t.KasIn, &t.KasOut, &t.IkromIn, &t.IkromOut,
		&createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}

	t.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	t.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
	return &t, nil
}

// CreateTransaction inserts a new transaction
func CreateTransaction(input models.TransactionInput) (int64, error) {
	if input.Date == "" {
		input.Date = time.Now().Format("2006-01-02")
	}
	if input.Category == "" {
		input.Category = "Umum"
	}

	query := `
		INSERT INTO transactions (
			date, ref_no, description, category,
			kas_in, kas_out, ikrom_in, ikrom_out,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, datetime('now', 'localtime'), datetime('now', 'localtime'))
	`
	result, err := DB.Exec(query,
		input.Date, input.RefNo, input.Description, input.Category,
		input.KasIn, input.KasOut, input.IkromIn, input.IkromOut,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// UpdateTransaction updates an existing transaction
func UpdateTransaction(id int64, input models.TransactionInput) error {
	query := `
		UPDATE transactions
		SET date = ?, ref_no = ?, description = ?, category = ?,
		    kas_in = ?, kas_out = ?, ikrom_in = ?, ikrom_out = ?,
		    updated_at = datetime('now', 'localtime')
		WHERE id = ?
	`
	_, err := DB.Exec(query,
		input.Date, input.RefNo, input.Description, input.Category,
		input.KasIn, input.KasOut, input.IkromIn, input.IkromOut,
		id,
	)
	return err
}

// DeleteTransaction deletes a transaction by ID
func DeleteTransaction(id int64) error {
	_, err := DB.Exec("DELETE FROM transactions WHERE id = ?", id)
	return err
}

// GetSummary calculates overall and filtered totals for Kas & Ikrom
func GetSummary(startDate, endDate string) (*models.Summary, error) {
	var s models.Summary

	// Overall Lifetime Summary
	err := DB.QueryRow(`
		SELECT 
			COALESCE(SUM(kas_in), 0), COALESCE(SUM(kas_out), 0), COALESCE(SUM(kas_in - kas_out), 0),
			COALESCE(SUM(ikrom_in), 0), COALESCE(SUM(ikrom_out), 0), COALESCE(SUM(ikrom_in - ikrom_out), 0),
			COALESCE(SUM(kas_in + ikrom_in), 0),
			COALESCE(SUM(kas_out + ikrom_out), 0),
			COALESCE(SUM((kas_in - kas_out) + (ikrom_in - ikrom_out)), 0),
			COUNT(*)
		FROM transactions
	`).Scan(
		&s.TotalKasIn, &s.TotalKasOut, &s.SaldoKas,
		&s.TotalIkromIn, &s.TotalIkromOut, &s.SaldoIkrom,
		&s.TotalMasuk, &s.TotalKeluar, &s.TotalSaldo,
		&s.TransactionCount,
	)
	if err != nil {
		return nil, err
	}

	// Filter Period Summary
	var conditions []string
	var args []interface{}
	if startDate != "" {
		conditions = append(conditions, "date >= ?")
		args = append(args, startDate)
	}
	if endDate != "" {
		conditions = append(conditions, "date <= ?")
		args = append(args, endDate)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	periodQuery := fmt.Sprintf(`
		SELECT 
			COALESCE(SUM(kas_in), 0), COALESCE(SUM(kas_out), 0),
			COALESCE(SUM(ikrom_in), 0), COALESCE(SUM(ikrom_out), 0),
			COALESCE(SUM(kas_in + ikrom_in), 0),
			COALESCE(SUM(kas_out + ikrom_out), 0)
		FROM transactions
		%s
	`, whereClause)

	err = DB.QueryRow(periodQuery, args...).Scan(
		&s.PeriodKasIn, &s.PeriodKasOut,
		&s.PeriodIkromIn, &s.PeriodIkromOut,
		&s.PeriodMasuk, &s.PeriodKeluar,
	)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

// GetChartData aggregates data for visualizations
func GetChartData() (*models.ChartData, error) {
	var chartData models.ChartData

	// Query last 12 months aggregates
	rows, err := DB.Query(`
		SELECT 
			strftime('%Y-%m', date) AS month,
			COALESCE(SUM(kas_in), 0) as kas_in,
			COALESCE(SUM(kas_out), 0) as kas_out,
			COALESCE(SUM(ikrom_in), 0) as ikrom_in,
			COALESCE(SUM(ikrom_out), 0) as ikrom_out,
			COALESCE(SUM(kas_in + ikrom_in), 0) as total_in,
			COALESCE(SUM(kas_out + ikrom_out), 0) as total_out
		FROM transactions
		GROUP BY strftime('%Y-%m', date)
		ORDER BY month ASC
		LIMIT 12
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var m models.MonthlyStats
		err := rows.Scan(
			&m.Month,
			&m.KasIn, &m.KasOut,
			&m.IkromIn, &m.IkromOut,
			&m.TotalIn, &m.TotalOut,
		)
		if err != nil {
			return nil, err
		}
		m.NetChange = m.TotalIn - m.TotalOut
		chartData.Monthly = append(chartData.Monthly, m)
	}

	// Pos Distribution (Current Saldo Breakdown: Kas vs Ikrom)
	_ = DB.QueryRow(`
		SELECT 
			COALESCE(SUM(kas_in - kas_out), 0),
			COALESCE(SUM(ikrom_in - ikrom_out), 0)
		FROM transactions
	`).Scan(
		&chartData.PosDistribution.Kas,
		&chartData.PosDistribution.Ikrom,
	)

	return &chartData, nil
}

// GetCategories returns all category records
func GetCategories() ([]models.Category, error) {
	rows, err := DB.Query("SELECT id, name, type, pos FROM categories ORDER BY name ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Type, &c.Pos); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, nil
}

// CreateCategory adds a category
func CreateCategory(c models.Category) (int64, error) {
	res, err := DB.Exec("INSERT INTO categories (name, type, pos) VALUES (?, ?, ?)", c.Name, c.Type, c.Pos)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// DeleteCategory removes a category by ID
func DeleteCategory(id int64) error {
	_, err := DB.Exec("DELETE FROM categories WHERE id = ?", id)
	return err
}

// ResetAllData wipes all transactions and resets category table
func ResetAllData() error {
	_, err := DB.Exec("DELETE FROM transactions")
	if err != nil {
		return err
	}
	_, _ = DB.Exec("DELETE FROM sqlite_sequence WHERE name='transactions'")
	return nil
}

// SeedSampleData populates rich sample records for Kas and Ikrom
func SeedSampleData() error {
	_ = ResetAllData()

	now := time.Now()
	thisYear := now.Year()
	thisMonth := int(now.Month())

	sampleItems := []struct {
		Date        string
		RefNo       string
		Description string
		Category    string
		KasIn       float64
		KasOut      float64
		IkromIn     float64
		IkromOut    float64
	}{
		{
			Date:        fmt.Sprintf("%04d-%02d-01", thisYear, thisMonth),
			RefNo:       "BKM-001",
			Description: "Saldo Awal Kas & Titipan Ikrom Awal Bulan",
			Category:    "Infaq / Shodaqoh",
			KasIn:       6500000, KasOut: 0,
			IkromIn:     3500000, IkromOut: 0,
		},
		{
			Date:        fmt.Sprintf("%04d-%02d-03", thisYear, thisMonth),
			RefNo:       "BKM-002",
			Description: "Penerimaan Infaq Jamaah & Donasi Rutin",
			Category:    "Infaq / Shodaqoh",
			KasIn:       2800000, KasOut: 0,
			IkromIn:     0, IkromOut: 0,
		},
		{
			Date:        fmt.Sprintf("%04d-%02d-06", thisYear, thisMonth),
			RefNo:       "BKM-003",
			Description: "Titipan Khusus Dana Ikrom dari Donatur H. Ahmad",
			Category:    "Penerimaan Ikrom / Bisaroh",
			KasIn:       0, KasOut: 0,
			IkromIn:     2000000, IkromOut: 0,
		},
		{
			Date:        fmt.Sprintf("%04d-%02d-08", thisYear, thisMonth),
			RefNo:       "BKK-001",
			Description: "Bisaroh / Honor Guru Pengajar & Tenaga Pendidik",
			Category:    "Bisaroh Guru / Ustadz",
			KasIn:       0, KasOut: 0,
			IkromIn:     0, IkromOut: 2500000,
		},
		{
			Date:        fmt.Sprintf("%04d-%02d-10", thisYear, thisMonth),
			RefNo:       "BKK-002",
			Description: "Pembayaran Listrik PLN, Air PDAM, dan Internet",
			Category:    "Operasional & Listrik / Air",
			KasIn:       0, KasOut: 650000,
			IkromIn:     0, IkromOut: 0,
		},
		{
			Date:        fmt.Sprintf("%04d-%02d-14", thisYear, thisMonth),
			RefNo:       "BKK-003",
			Description: "Pembelian ATK & Perlengkapan Kantor / Inventaris",
			Category:    "Pengadaan ATK & Perlengkapan",
			KasIn:       0, KasOut: 450000,
			IkromIn:     0, IkromOut: 0,
		},
		{
			Date:        fmt.Sprintf("%04d-%02d-17", thisYear, thisMonth),
			RefNo:       "BKM-004",
			Description: "Penerimaan Infaq Kotak Amal Jum'at",
			Category:    "Infaq / Shodaqoh",
			KasIn:       1950000, KasOut: 0,
			IkromIn:     0, IkromOut: 0,
		},
		{
			Date:        fmt.Sprintf("%04d-%02d-20", thisYear, thisMonth),
			RefNo:       "BKK-004",
			Description: "Bisaroh / Uang Saku Ustadz Tamu Pengajian Ahad Pagi",
			Category:    "Bisaroh Guru / Ustadz",
			KasIn:       0, KasOut: 0,
			IkromIn:     0, IkromOut: 600000,
		},
		{
			Date:        fmt.Sprintf("%04d-%02d-22", thisYear, thisMonth),
			RefNo:       "BKK-005",
			Description: "Biaya Konsumsi Rapat Pengurus & Jamuan Tamu",
			Category:    "Konsumsi & Jamuan",
			KasIn:       0, KasOut: 350000,
			IkromIn:     0, IkromOut: 0,
		},
		{
			Date:        fmt.Sprintf("%04d-%02d-26", thisYear, thisMonth),
			RefNo:       "BKK-006",
			Description: "Pemeliharaan AC, Lampu, dan Kebersihan Fasilitas",
			Category:    "Pemeliharaan Fasilitas",
			KasIn:       0, KasOut: 500000,
			IkromIn:     0, IkromOut: 0,
		},
	}

	for _, item := range sampleItems {
		_, err := CreateTransaction(models.TransactionInput{
			Date:        item.Date,
			RefNo:       item.RefNo,
			Description: item.Description,
			Category:    item.Category,
			KasIn:       item.KasIn,
			KasOut:      item.KasOut,
			IkromIn:     item.IkromIn,
			IkromOut:    item.IkromOut,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
