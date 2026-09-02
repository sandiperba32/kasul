package models

import "time"

// KasBook represents a main cash book
type KasBook struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	ModelType int       `json:"model_type"` // 1: With Ikrom & Halaqoh, 2: Standard
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type KasBookInput struct {
	Name      string `json:"name"`
	ModelType int    `json:"model_type"`
}

// Transaction represents a cash book entry for Kas & Ikrom with Nama and Keterangan
type Transaction struct {
	ID           int64     `json:"id"`
	KasID        int64     `json:"kas_id"`
	Date         string    `json:"date"`          // Format: YYYY-MM-DD
	RefNo        string    `json:"ref_no"`        // Nomor Bukti/Kwitansi (e.g. BKM-001)
	Name         string    `json:"name"`          // Nama Siswa / Donatur / Penerima
	Description  string    `json:"description"`   // Keterangan / Uraian
	KasIn        float64   `json:"kas_in"`        // Kas Masuk
	KasOut       float64   `json:"kas_out"`       // Kas Keluar
	IkromIn      float64   `json:"ikrom_in"`      // Ikrom Masuk
	IkromOut     float64   `json:"ikrom_out"`     // Ikrom Keluar
	KasBalance   float64   `json:"kas_balance"`   // Running balance for Kas
	IkromBalance float64   `json:"ikrom_balance"` // Running balance for Ikrom
	TotalBalance float64   `json:"total_balance"` // Running grand total (Kas + Ikrom)
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TransactionInput is the payload received when creating or updating a transaction
type TransactionInput struct {
	KasID       int64   `json:"kas_id"`
	Date        string  `json:"date"`
	RefNo       string  `json:"ref_no"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	KasIn       float64 `json:"kas_in"`
	KasOut      float64 `json:"kas_out"`
	IkromIn     float64 `json:"ikrom_in"`
	IkromOut    float64 `json:"ikrom_out"`
}

// Student represents Master Data Siswa - hanya Nama dan Halaqoh
type Student struct {
	ID        int64     `json:"id"`
	KasID     int64     `json:"kas_id"`
	Name      string    `json:"name"`       // Nama Lengkap Siswa
	Halaqoh   string    `json:"halaqoh"`    // Halaqoh
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// StudentInput is the payload received when creating or updating a student
type StudentInput struct {
	KasID   int64  `json:"kas_id"`
	Name    string `json:"name"`
	Halaqoh string `json:"halaqoh"`
}

// Summary provides aggregate metrics for the dashboard
type Summary struct {
	TotalKasIn    float64 `json:"total_kas_in"`
	TotalKasOut   float64 `json:"total_kas_out"`
	SaldoKas      float64 `json:"saldo_kas"`

	TotalIkromIn  float64 `json:"total_ikrom_in"`
	TotalIkromOut float64 `json:"total_ikrom_out"`
	SaldoIkrom    float64 `json:"saldo_ikrom"`

	TotalMasuk    float64 `json:"total_masuk"`
	TotalKeluar   float64 `json:"total_keluar"`
	TotalSaldo    float64 `json:"total_saldo"`

	// Filter / period specific aggregates
	PeriodKasIn    float64 `json:"period_kas_in"`
	PeriodKasOut   float64 `json:"period_kas_out"`
	PeriodIkromIn  float64 `json:"period_ikrom_in"`
	PeriodIkromOut float64 `json:"period_ikrom_out"`
	PeriodMasuk    float64 `json:"period_masuk"`
	PeriodKeluar   float64 `json:"period_keluar"`

	TransactionCount int `json:"transaction_count"`
	StudentCount     int `json:"student_count"`
}

// MonthlyStats provides monthly trend data for charts
type MonthlyStats struct {
	Month     string  `json:"month"` // e.g. "2026-01"
	KasIn     float64 `json:"kas_in"`
	KasOut    float64 `json:"kas_out"`
	IkromIn   float64 `json:"ikrom_in"`
	IkromOut  float64 `json:"ikrom_out"`
	TotalIn   float64 `json:"total_in"`
	TotalOut  float64 `json:"total_out"`
	NetChange float64 `json:"net_change"`
}

// ChartData encapsulates all data needed by the charts view
type ChartData struct {
	Monthly []MonthlyStats `json:"monthly"`
	PosDistribution struct {
		Kas   float64 `json:"kas"`
		Ikrom float64 `json:"ikrom"`
	} `json:"pos_distribution"`
}
