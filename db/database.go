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

// InitDB initializes SQLite connection, runs migrations, and seeds initial data
func InitDB(dataSourceName string) (*sql.DB, error) {
	var err error
	DB, err = sql.Open("sqlite", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	DB.SetMaxOpenConns(1)
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

	return DB, nil
}

func createTables() error {
	queries := []string{
		// Students table: hanya nama dan nama orang tua
		`CREATE TABLE IF NOT EXISTS students (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			parent TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE INDEX IF NOT EXISTS idx_student_name ON students(name);`,

		// Transactions table
		`CREATE TABLE IF NOT EXISTS transactions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			date TEXT NOT NULL,
			ref_no TEXT DEFAULT '',
			name TEXT DEFAULT '',
			description TEXT NOT NULL,
			kas_in REAL NOT NULL DEFAULT 0.0,
			kas_out REAL NOT NULL DEFAULT 0.0,
			ikrom_in REAL NOT NULL DEFAULT 0.0,
			ikrom_out REAL NOT NULL DEFAULT 0.0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE INDEX IF NOT EXISTS idx_trans_date ON transactions(date);`,
	}

	for _, q := range queries {
		if _, err := DB.Exec(q); err != nil {
			return err
		}
	}

	// Migration: add name column to transactions if not exists
	_, _ = DB.Exec("ALTER TABLE transactions ADD COLUMN name TEXT DEFAULT '';")

	return nil
}

// -------------------------------- STUDENT CRUD --------------------------------

// GetAllStudents returns students with optional search filter
func GetAllStudents(search string) ([]models.Student, error) {
	var conditions []string
	var args []interface{}

	if search != "" {
		likePattern := "%" + search + "%"
		conditions = append(conditions, "(name LIKE ? OR parent LIKE ?)")
		args = append(args, likePattern, likePattern)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	query := fmt.Sprintf(`
		SELECT id, name, parent, created_at, updated_at
		FROM students
		%s
		ORDER BY name ASC
	`, whereClause)

	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []models.Student
	for rows.Next() {
		var s models.Student
		var createdAt, updatedAt string
		err := rows.Scan(&s.ID, &s.Name, &s.Parent, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}
		s.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		s.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
		students = append(students, s)
	}
	return students, nil
}

// GetStudentByID fetches a single student by ID
func GetStudentByID(id int64) (*models.Student, error) {
	var s models.Student
	var createdAt, updatedAt string
	err := DB.QueryRow(
		`SELECT id, name, parent, created_at, updated_at FROM students WHERE id = ?`, id,
	).Scan(&s.ID, &s.Name, &s.Parent, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	s.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	s.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
	return &s, nil
}

// CreateStudent inserts a new student
func CreateStudent(input models.StudentInput) (int64, error) {
	result, err := DB.Exec(
		`INSERT INTO students (name, parent, created_at, updated_at)
		 VALUES (?, ?, datetime('now','localtime'), datetime('now','localtime'))`,
		input.Name, input.Parent,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// UpdateStudent updates an existing student
func UpdateStudent(id int64, input models.StudentInput) error {
	_, err := DB.Exec(
		`UPDATE students SET name = ?, parent = ?, updated_at = datetime('now','localtime') WHERE id = ?`,
		input.Name, input.Parent, id,
	)
	return err
}

// DeleteStudent deletes a student
func DeleteStudent(id int64) error {
	_, err := DB.Exec("DELETE FROM students WHERE id = ?", id)
	return err
}

// -------------------------------- TRANSACTIONS --------------------------------

// GetInitialBalance calculates balance of all transactions before startDate
func GetInitialBalance(beforeDate string) (kasBal, ikromBal float64, err error) {
	if beforeDate == "" {
		return 0, 0, nil
	}
	err = DB.QueryRow(`
		SELECT COALESCE(SUM(kas_in - kas_out), 0), COALESCE(SUM(ikrom_in - ikrom_out), 0)
		FROM transactions WHERE date < ?
	`, beforeDate).Scan(&kasBal, &ikromBal)
	return
}

// GetAllTransactions retrieves filtered transactions with running balances
func GetAllTransactions(startDate, endDate, search, posFilter string) ([]models.Transaction, error) {
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
	if search != "" {
		like := "%" + search + "%"
		conditions = append(conditions, "(name LIKE ? OR description LIKE ? OR ref_no LIKE ?)")
		args = append(args, like, like, like)
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

	initKas, initIkrom, err := GetInitialBalance(startDate)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`
		SELECT id, date, ref_no, COALESCE(name,''), description,
		       kas_in, kas_out, ikrom_in, ikrom_out, created_at, updated_at
		FROM transactions %s ORDER BY date ASC, id ASC
	`, whereClause)

	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txns []models.Transaction
	runningKas := initKas
	runningIkrom := initIkrom

	for rows.Next() {
		var t models.Transaction
		var ca, ua string
		if err := rows.Scan(&t.ID, &t.Date, &t.RefNo, &t.Name, &t.Description,
			&t.KasIn, &t.KasOut, &t.IkromIn, &t.IkromOut, &ca, &ua); err != nil {
			return nil, err
		}
		t.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", ca)
		t.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", ua)

		runningKas += t.KasIn - t.KasOut
		runningIkrom += t.IkromIn - t.IkromOut
		t.KasBalance = runningKas
		t.IkromBalance = runningIkrom
		t.TotalBalance = runningKas + runningIkrom
		txns = append(txns, t)
	}
	return txns, rows.Err()
}

// GetTransactionByID fetches a single transaction by ID
func GetTransactionByID(id int64) (*models.Transaction, error) {
	var t models.Transaction
	var ca, ua string
	err := DB.QueryRow(`
		SELECT id, date, ref_no, COALESCE(name,''), description,
		       kas_in, kas_out, ikrom_in, ikrom_out, created_at, updated_at
		FROM transactions WHERE id = ?
	`, id).Scan(&t.ID, &t.Date, &t.RefNo, &t.Name, &t.Description,
		&t.KasIn, &t.KasOut, &t.IkromIn, &t.IkromOut, &ca, &ua)
	if err != nil {
		return nil, err
	}
	t.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", ca)
	t.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", ua)
	return &t, nil
}

// CreateTransaction inserts a new transaction
func CreateTransaction(input models.TransactionInput) (int64, error) {
	if input.Date == "" {
		input.Date = time.Now().Format("2006-01-02")
	}
	result, err := DB.Exec(`
		INSERT INTO transactions (date, ref_no, name, description,
		    kas_in, kas_out, ikrom_in, ikrom_out, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'))
	`, input.Date, input.RefNo, input.Name, input.Description,
		input.KasIn, input.KasOut, input.IkromIn, input.IkromOut)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// UpdateTransaction updates an existing transaction
func UpdateTransaction(id int64, input models.TransactionInput) error {
	_, err := DB.Exec(`
		UPDATE transactions SET date=?, ref_no=?, name=?, description=?,
		    kas_in=?, kas_out=?, ikrom_in=?, ikrom_out=?,
		    updated_at=datetime('now','localtime')
		WHERE id=?
	`, input.Date, input.RefNo, input.Name, input.Description,
		input.KasIn, input.KasOut, input.IkromIn, input.IkromOut, id)
	return err
}

// DeleteTransaction deletes a transaction by ID
func DeleteTransaction(id int64) error {
	_, err := DB.Exec("DELETE FROM transactions WHERE id = ?", id)
	return err
}

// GetSummary calculates overall and period totals
func GetSummary(startDate, endDate string) (*models.Summary, error) {
	var s models.Summary

	err := DB.QueryRow(`
		SELECT
			COALESCE(SUM(kas_in),0), COALESCE(SUM(kas_out),0), COALESCE(SUM(kas_in-kas_out),0),
			COALESCE(SUM(ikrom_in),0), COALESCE(SUM(ikrom_out),0), COALESCE(SUM(ikrom_in-ikrom_out),0),
			COALESCE(SUM(kas_in+ikrom_in),0), COALESCE(SUM(kas_out+ikrom_out),0),
			COALESCE(SUM((kas_in-kas_out)+(ikrom_in-ikrom_out)),0), COUNT(*)
		FROM transactions
	`).Scan(&s.TotalKasIn, &s.TotalKasOut, &s.SaldoKas,
		&s.TotalIkromIn, &s.TotalIkromOut, &s.SaldoIkrom,
		&s.TotalMasuk, &s.TotalKeluar, &s.TotalSaldo, &s.TransactionCount)
	if err != nil {
		return nil, err
	}

	_ = DB.QueryRow("SELECT COUNT(*) FROM students").Scan(&s.StudentCount)

	var conds []string
	var args []interface{}
	if startDate != "" {
		conds = append(conds, "date >= ?")
		args = append(args, startDate)
	}
	if endDate != "" {
		conds = append(conds, "date <= ?")
		args = append(args, endDate)
	}
	where := ""
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}

	err = DB.QueryRow(fmt.Sprintf(`
		SELECT COALESCE(SUM(kas_in),0), COALESCE(SUM(kas_out),0),
		       COALESCE(SUM(ikrom_in),0), COALESCE(SUM(ikrom_out),0),
		       COALESCE(SUM(kas_in+ikrom_in),0), COALESCE(SUM(kas_out+ikrom_out),0)
		FROM transactions %s
	`, where), args...).Scan(&s.PeriodKasIn, &s.PeriodKasOut,
		&s.PeriodIkromIn, &s.PeriodIkromOut, &s.PeriodMasuk, &s.PeriodKeluar)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

// GetChartData aggregates data for charts
func GetChartData() (*models.ChartData, error) {
	var cd models.ChartData

	rows, err := DB.Query(`
		SELECT strftime('%Y-%m', date) AS month,
			COALESCE(SUM(kas_in),0), COALESCE(SUM(kas_out),0),
			COALESCE(SUM(ikrom_in),0), COALESCE(SUM(ikrom_out),0),
			COALESCE(SUM(kas_in+ikrom_in),0), COALESCE(SUM(kas_out+ikrom_out),0)
		FROM transactions GROUP BY month ORDER BY month ASC LIMIT 12
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var m models.MonthlyStats
		rows.Scan(&m.Month, &m.KasIn, &m.KasOut, &m.IkromIn, &m.IkromOut, &m.TotalIn, &m.TotalOut)
		m.NetChange = m.TotalIn - m.TotalOut
		cd.Monthly = append(cd.Monthly, m)
	}

	_ = DB.QueryRow(`
		SELECT COALESCE(SUM(kas_in-kas_out),0), COALESCE(SUM(ikrom_in-ikrom_out),0)
		FROM transactions
	`).Scan(&cd.PosDistribution.Kas, &cd.PosDistribution.Ikrom)

	return &cd, nil
}

// ResetAllData wipes all transactions
func ResetAllData() error {
	_, err := DB.Exec("DELETE FROM transactions")
	if err != nil {
		return err
	}
	_, _ = DB.Exec("DELETE FROM sqlite_sequence WHERE name='transactions'")
	return nil
}

// SeedSampleData populates sample records using student names
func SeedSampleData() error {
	_ = ResetAllData()

	now := time.Now()
	y, m := now.Year(), int(now.Month())

	// Ensure some default students exist
	var studentCount int
	_ = DB.QueryRow("SELECT COUNT(*) FROM students").Scan(&studentCount)
	if studentCount == 0 {
		defaultStudents := []models.StudentInput{
			{Name: "Muhammad Rizky Pratama", Parent: "Bpk. Hendra Saputra"},
			{Name: "Siti Fatimah Azzahra", Parent: "Bpk. Fauzi Rahman"},
			{Name: "Ahmad Fauzan Mubarok", Parent: "Bpk. Abdullah Hasan"},
			{Name: "Nur Aini Rahmawati", Parent: "Bpk. Wahyu Pratomo"},
			{Name: "Bilal Ramadhan", Parent: "Bpk. Supriyadi"},
		}
		for _, s := range defaultStudents {
			_, _ = CreateStudent(s)
		}
	}

	samples := []struct {
		Date, RefNo, Name, Desc       string
		KasIn, KasOut, IkromIn, IkromOut float64
	}{
		{fmt.Sprintf("%04d-%02d-01", y, m), "BKM-001", "Muhammad Rizky Pratama", "Iuran Kas & Infaq Bulanan Siswa", 350000, 0, 150000, 0},
		{fmt.Sprintf("%04d-%02d-02", y, m), "BKM-002", "Siti Fatimah Azzahra", "Iuran Kas & Infaq Bulanan Siswa", 350000, 0, 150000, 0},
		{fmt.Sprintf("%04d-%02d-03", y, m), "BKM-003", "Ahmad Fauzan Mubarok", "Iuran Kas & Infaq Bulanan Siswa", 350000, 0, 150000, 0},
		{fmt.Sprintf("%04d-%02d-05", y, m), "BKM-004", "Bpk. H. Ahmad Dahlan", "Titipan Khusus Dana Ikrom / Bisaroh Guru", 0, 0, 3000000, 0},
		{fmt.Sprintf("%04d-%02d-08", y, m), "BKK-001", "Ustadz Fauzi & Pengajar", "Bisaroh / Honor Guru Pengajar", 0, 0, 0, 2500000},
		{fmt.Sprintf("%04d-%02d-10", y, m), "BKK-002", "PLN & PDAM", "Pembayaran Listrik & Air Kelas", 0, 650000, 0, 0},
		{fmt.Sprintf("%04d-%02d-12", y, m), "BKM-005", "Nur Aini Rahmawati", "Iuran Kas & Infaq Bulanan Siswa", 350000, 0, 150000, 0},
		{fmt.Sprintf("%04d-%02d-15", y, m), "BKM-006", "Bilal Ramadhan", "Iuran Kas & Infaq Bulanan Siswa", 350000, 0, 150000, 0},
		{fmt.Sprintf("%04d-%02d-18", y, m), "BKK-003", "Toko ATK Berkah", "Pembelian Perlengkapan Kelas", 0, 350000, 0, 0},
		{fmt.Sprintf("%04d-%02d-20", y, m), "BKK-004", "Ustadz Tamu", "Bisaroh Pembicara Kajian", 0, 0, 0, 400000},
	}

	for _, item := range samples {
		_, err := CreateTransaction(models.TransactionInput{
			Date: item.Date, RefNo: item.RefNo, Name: item.Name, Description: item.Desc,
			KasIn: item.KasIn, KasOut: item.KasOut, IkromIn: item.IkromIn, IkromOut: item.IkromOut,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
