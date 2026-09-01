# Aplikasi Buku Kas & Dana Ikrom

Aplikasi pencatatan keuangan dan buku kas tabelar (multi-kolom) yang mengelola 2 pos dana: **Kas Utama** dan **Dana Ikrom (Bisaroh / Honor Guru & Ustadz)**.

Dibangun dengan **Golang** untuk backend performa tinggi dan **Vue 3** untuk frontend modern & reaktif.

---

## 🚀 Fitur Utama

1. **Buku Kas Tabelar (Multi-Kolom)**:
   - Kolom **KAS UTAMA**: Masuk, Keluar, Saldo Kas.
   - Kolom **DANA IKROM**: Masuk, Keluar, Saldo Ikrom.
   - Kolom **TOTAL SALDO**: Akumulasi kumulatif keseluruhan secara otomatis per baris transaksi ($Saldo Kas + Saldo Ikrom$).
   - Baris Total Footer: Rekapitulasi mutasi dan saldo akhir periode.

2. **Dashboard Statistik & Ringkasan Keuangan**:
   - Kartu Saldo Kas Utama + Total Masuk & Keluar.
   - Kartu Saldo Dana Ikrom + Total Masuk & Keluar.
   - Kartu Grand Total Saldo Gabungan.

3. **Formulir Transaksi Fleksibel**:
   - **Mode Input Cepat (Tunggal)**: Pilih Jenis (Masuk / Keluar) -> Pilih Pos (Kas / Ikrom) -> Isi Nominal Rupiah.
   - **Mode Multi-Pos (Majemuk)**: Mengisi langsung beberapa pos sekaligus dalam 1 nomor transaksi/kuitansi.

4. **Filter & Pencarian**:
   - Preset Periode: *Semua, Hari Ini, Bulan Ini, Bulan Lalu, Tahun Ini*.
   - Filter Rentang Tanggal (*Dari Tanggal s/d Sampai Tanggal*).
   - Filter Pos Dana (*Semua Pos, Kas Saja, Ikrom Saja*).
   - Filter Kategori.
   - Real-time search pada uraian keterangan dan nomor bukti kuitansi.

5. **Visualisasi Data & Grafik (Chart.js)**:
   - Grafik Pemasukan vs Pengeluaran per Bulan (Kas vs Ikrom).
   - Diagram Donat Distribusi Komposisi Saldo (Kas vs Ikrom).
   - Grafik Tren Pertumbuhan Saldo Kumulatif.

6. **Ekspor & Cetak Laporan (PDF / Print)**:
   - **Cetak Laporan Resmi (Print Preview / Save PDF)** format A4 Landscape lengkap dengan kop surat, tabel ringkasan per pos, mutasi transaksi detail, dan kolom tanda tangan (Ketua & Bendahara).
   - **Ekspor Excel / CSV** UTF-8 BOM yang langsung rapi saat dibuka di Microsoft Excel.

7. **Database SQLite Embedded**:
   - Berjalan mandiri tanpa perlu install database server eksternal (`kas.db`).
   - Mode WAL (Write-Ahead Logging) untuk performa cepat dan aman.

---

## 💻 Cara Menjalankan Aplikasi

Jalankan file executable:
```bash
./kas-app.exe
```
atau dari source code Go:
```bash
go run main.go
```

Setelah server berjalan, buka browser di:
👉 **[http://localhost:8080](http://localhost:8080)**
