# Aplikasi Buku Kas Terpadu (Kas, Ikrom, Pen)

Aplikasi pencatatan keuangan dan buku kas tabelar (multi-kolom) yang mengelola 3 pos dana secara simultan: **Kas Utama**, **Dana Ikrom (Bisaroh / Honor)**, dan **Dana Pen (Pendidikan / Pembangunan)**.

Dibangun dengan **Golang** untuk backend performa tinggi dan **Vue 3** untuk frontend modern & reaktif.

---

## 🚀 Fitur Utama

1. **Buku Kas Tabelar (Multi-Kolom)**:
   - Kolom **KAS**: Masuk, Keluar, Saldo Kas.
   - Kolom **IKROM**: Masuk, Keluar, Saldo Ikrom.
   - Kolom **PEN**: Masuk, Keluar, Saldo Pen.
   - Kolom **TOTAL SALDO**: Akumulasi kumulatif keseluruhan secara otomatis per baris transaksi.
   - Baris Total Footer: Rekapitulasi mutasi dan saldo akhir periode.

2. **Dashboard Statistik & Ringkasan Keuangan**:
   - Kartu Saldo Kas Utama + Total Masuk & Keluar.
   - Kartu Saldo Dana Ikrom + Total Masuk & Keluar.
   - Kartu Saldo Dana Pen + Total Masuk & Keluar.
   - Kartu Grand Total Saldo Gabungan.

3. **Formulir Transaksi Fleksibel**:
   - **Mode Input Cepat (Tunggal)**: Pilih Jenis (Masuk / Keluar) -> Pilih Pos (Kas / Ikrom / Pen) -> Isi Nominal Rupiah.
   - **Mode Multi-Pos (Majemuk)**: Mengisi langsung beberapa pos sekaligus dalam 1 nomor transaksi/kuitansi.

4. **Filter & Pencarian**:
   - Preset Periode: *Semua, Hari Ini, Bulan Ini, Bulan Lalu, Tahun Ini*.
   - Filter Rentang Tanggal (*Dari Tanggal s/d Sampai Tanggal*).
   - Filter Pos Dana (*Kas Saja, Ikrom Saja, Pen Saja*).
   - Filter Kategori.
   - Real-time search pada uraian keterangan dan nomor bukti kuitansi.

5. **Visualisasi Data & Grafik (Chart.js)**:
   - Grafik Pemasukan vs Pengeluaran per Bulan.
   - Diagram Donat Distribusi Komposisi Saldo.
   - Grafik Tren Pertumbuhan Saldo Kumulatif.

6. **Ekspor & Cetak Laporan (PDF / Print)**:
   - **Cetak Laporan Resmi (Print Preview / Save PDF)** format A4 Landscape lengkap dengan kop surat, tabel ringkasan per pos, mutasi transaksi detail, dan kolom tanda tangan (Ketua & Bendahara).
   - **Ekspor Excel / CSV** UTF-8 BOM yang langsung rapi saat dibuka di Microsoft Excel.

7. **Database SQLite Embedded**:
   - Berjalan mandiri tanpa perlu install database server eksternal (`kas.db`).
   - Mode WAL (Write-Ahead Logging) untuk performa cepat dan aman.

---

## 🛠️ Struktur Project

```
D:\kas/
├── db/
│   └── database.go        # Koneksi SQLite, skema tabel, kalkulasi saldo kumulatif, seeding
├── handlers/
│   └── handlers.go        # Handler REST API (CRUD, ringkasan, ekspor CSV, kategori)
├── models/
│   └── transaction.go     # Struct model Transaksi, Ringkasan, Kategori, Grafik
├── static/
│   ├── index.html         # Single Page Application (SPA)
│   ├── app.js             # Logika Vue 3 (Composition API, state, chart render)
│   └── style.css          # CSS kustom & aturan cetak cetak/print styling
├── go.mod
├── go.sum
├── main.go                # Server entrypoint & static router
├── kas-app.exe            # Binary executable
└── README.md
```

---

## 💻 Cara Menjalankan Aplikasi

### Opsi 1: Menjalankan Binary Langsung
Cukup jalankan file executable:
```bash
./kas-app.exe
```

### Opsi 2: Menjalankan dari Source Code Go
```bash
go run main.go
```

Setelah server berjalan, buka browser di:
👉 **[http://localhost:8080](http://localhost:8080)**

---

## 📡 Dokumentasi Endpoint REST API

| Method | Endpoint | Deskripsi |
|---|---|---|
| `GET` | `/api/transactions` | Mengambil daftar transaksi (mendukung query `start_date`, `end_date`, `category`, `search`, `pos`) |
| `GET` | `/api/transactions/{id}` | Mengambil detail 1 transaksi |
| `POST` | `/api/transactions` | Menambah transaksi baru |
| `PUT` | `/api/transactions/{id}` | Mengubah transaksi |
| `DELETE` | `/api/transactions/{id}` | Menghapus transaksi |
| `GET` | `/api/summary` | Mendapatkan ringkasan saldo per pos dan grand total |
| `GET` | `/api/chart-data` | Mendapatkan data statistik untuk grafik |
| `GET` | `/api/categories` | Mengambil daftar kategori |
| `POST` | `/api/categories` | Menambah kategori baru |
| `DELETE` | `/api/categories/{id}` | Menghapus kategori |
| `GET` | `/api/export/csv` | Download file CSV untuk Microsoft Excel |
| `POST` | `/api/seed` | Mengisi data simulasi/demo awal |
| `POST` | `/api/reset` | Mengosongkan data transaksi |
