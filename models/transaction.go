package models

import "time"

// Transaction represents a multi-column cash book entry
type Transaction struct {
	ID           int64     `json:"id"`
	Date         string    `json:"date"`          // Format: YYYY-MM-DD
	RefNo        string    `json:"ref_no"`        // Nomor Bukti/Kwitansi (e.g. BKM-001)
	Description  string    `json:"description"`   // Keterangan / Uraian
	Category     string    `json:"category"`      // Kategori (e.g. Donasi, SPP, Bisaroh, Operasional)
	KasIn        float64   `json:"kas_in"`        // Kas Masuk
	KasOut       float64   `json:"kas_out"`       // Kas Keluar
	IkromIn      float64   `json:"ikrom_in"`      // Ikrom Masuk
	IkromOut     float64   `json:"ikrom_out"`     // Ikrom Keluar
	PenIn        float64   `json:"pen_in"`        // Pen Masuk
	PenOut       float64   `json:"pen_out"`       // Pen Keluar
	KasBalance   float64   `json:"kas_balance"`   // Running balance for Kas
	IkromBalance float64   `json:"ikrom_balance"` // Running balance for Ikrom
	PenBalance   float64   `json:"pen_balance"`   // Running balance for Pen
	TotalBalance float64   `json:"total_balance"` // Running grand total (Kas + Ikrom + Pen)
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TransactionInput is the payload received when creating or updating a transaction
type TransactionInput struct {
	Date        string  `json:"date"`
	RefNo       string  `json:"ref_no"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	KasIn       float64 `json:"kas_in"`
	KasOut      float64 `json:"kas_out"`
	IkromIn     float64 `json:"ikrom_in"`
	IkromOut    float64 `json:"ikrom_out"`
	PenIn       float64 `json:"pen_in"`
	PenOut      float64 `json:"pen_out"`
}

// Summary provides aggregate metrics for the dashboard
type Summary struct {
	TotalKasIn    float64 `json:"total_kas_in"`
	TotalKasOut   float64 `json:"total_kas_out"`
	SaldoKas      float64 `json:"saldo_kas"`

	TotalIkromIn  float64 `json:"total_ikrom_in"`
	TotalIkromOut float64 `json:"total_ikrom_out"`
	SaldoIkrom    float64 `json:"saldo_ikrom"`

	TotalPenIn    float64 `json:"total_pen_in"`
	TotalPenOut   float64 `json:"total_pen_out"`
	SaldoPen      float64 `json:"saldo_pen"`

	TotalMasuk    float64 `json:"total_masuk"`
	TotalKeluar   float64 `json:"total_keluar"`
	TotalSaldo    float64 `json:"total_saldo"`

	// Filter / period specific aggregates
	PeriodKasIn    float64 `json:"period_kas_in"`
	PeriodKasOut   float64 `json:"period_kas_out"`
	PeriodIkromIn  float64 `json:"period_ikrom_in"`
	PeriodIkromOut float64 `json:"period_ikrom_out"`
	PeriodPenIn    float64 `json:"period_pen_in"`
	PeriodPenOut   float64 `json:"period_pen_out"`
	PeriodMasuk    float64 `json:"period_masuk"`
	PeriodKeluar   float64 `json:"period_keluar"`

	TransactionCount int `json:"transaction_count"`
}

// Category represents transaction categories
type Category struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"` // "in", "out", "both"
	Pos  string `json:"pos"`  // "kas", "ikrom", "pen", "all"
}

// MonthlyStats provides monthly trend data for charts
type MonthlyStats struct {
	Month     string  `json:"month"` // e.g. "2026-01"
	KasIn     float64 `json:"kas_in"`
	KasOut    float64 `json:"kas_out"`
	IkromIn   float64 `json:"ikrom_in"`
	IkromOut  float64 `json:"ikrom_out"`
	PenIn     float64 `json:"pen_in"`
	PenOut    float64 `json:"pen_out"`
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
		Pen   float64 `json:"pen"`
	} `json:"pos_distribution"`
}
