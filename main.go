package main

import (
	"fmt"
	"kas-app/db"
	"kas-app/handlers"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// corsMiddleware adds standard CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs incoming requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		if strings.HasPrefix(r.URL.Path, "/api/") {
			log.Printf("[%s] %s - %v", r.Method, r.URL.Path, time.Since(start))
		}
	})
}

func main() {
	dbPath := "kas.db"
	if customDB := os.Getenv("DB_PATH"); customDB != "" {
		dbPath = customDB
	}

	log.Printf("Menginisialisasi database SQLite: %s ...", dbPath)
	_, err := db.InitDB(dbPath)
	if err != nil {
		log.Fatalf("Gagal inisialisasi database: %v", err)
	}
	log.Println("Database SQLite berhasil dihubungkan.")

	// Auto-seed sample data on first run if database is empty
	summary, err := db.GetSummary(1, "", "")
	if err == nil && summary.TransactionCount == 0 {
		log.Println("Database kosong, mengisi data demo awal...")
		_ = db.SeedSampleData(1)
	}

	mux := http.NewServeMux()

	// Student API Routes
	mux.HandleFunc("/api/students", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetStudents(w, r)
		case http.MethodPost:
			handlers.CreateStudent(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/students/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			handlers.UpdateStudent(w, r)
		case http.MethodDelete:
			handlers.DeleteStudent(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Transaction API Routes
	mux.HandleFunc("/api/transactions", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetTransactions(w, r)
		case http.MethodPost:
			handlers.CreateTransaction(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/transactions/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			handlers.UpdateTransaction(w, r)
		case http.MethodDelete:
			handlers.DeleteTransaction(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})


	// KasBooks API Routes
	mux.HandleFunc("/api/kas_books", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetKasBooks(w, r)
		case http.MethodPost:
			handlers.CreateKasBook(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/kas_books/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			handlers.UpdateKasBook(w, r)
		case http.MethodDelete:
			handlers.DeleteKasBook(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/summary", handlers.GetSummary)
	mux.HandleFunc("/api/chart-data", handlers.GetChartData)
	mux.HandleFunc("/api/export/csv", handlers.ExportCSV)
	mux.HandleFunc("/api/seed", handlers.SeedData)
	mux.HandleFunc("/api/reset", handlers.ResetData)

	// Auth Routes
	mux.HandleFunc("/api/auth/login", handlers.Login)
	mux.HandleFunc("/api/auth/check", handlers.CheckSession)

	// Static Frontend Server
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/", fs)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	handler := loggingMiddleware(corsMiddleware(mux))

	fmt.Println("==========================================================")
	fmt.Println("  APLIKASI BUKU KAS & DANA IKROM + MASTER DATA SISWA")
	fmt.Println("  Backend: Golang (net/http + SQLite)")
	fmt.Println("  Frontend: Vue 3 + Tailwind CSS + Chart.js")
	fmt.Printf("  Aplikasi berjalan di: http://localhost:%s\n", port)
	fmt.Println("==========================================================")

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}
