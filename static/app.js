const { createApp, ref, reactive, computed, onMounted, watch, nextTick } = Vue;

createApp({
    setup() {
        // State
        const transactions = ref([]);
        const summary = ref({
            total_kas_in: 0, total_kas_out: 0, saldo_kas: 0,
            total_ikrom_in: 0, total_ikrom_out: 0, saldo_ikrom: 0,
            total_pen_in: 0, total_pen_out: 0, saldo_pen: 0,
            total_masuk: 0, total_keluar: 0, total_saldo: 0,
            period_kas_in: 0, period_kas_out: 0,
            period_ikrom_in: 0, period_ikrom_out: 0,
            period_pen_in: 0, period_pen_out: 0,
            period_masuk: 0, period_keluar: 0,
            transaction_count: 0
        });
        const categories = ref([]);
        const chartData = ref(null);
        const activeTab = ref('ledger'); // 'ledger', 'charts', 'categories'
        const loading = ref(false);
        const searchInput = ref('');

        // Modals
        const showModal = ref(false);
        const isEditing = ref(false);
        const editingId = ref(null);
        const showPrintModal = ref(false);
        const showCategoryModal = ref(false);

        // Filter state
        const filters = reactive({
            startDate: '',
            endDate: '',
            category: 'Semua',
            pos: 'all',
            periodPreset: 'all'
        });

        // Form state
        const inputMode = ref('single'); // 'single' (Cepat) or 'multi' (Tabelar Lengkap)
        const singleForm = reactive({
            type: 'in',      // 'in' or 'out'
            pos: 'kas',      // 'kas', 'ikrom', 'pen'
            amount: ''
        });

        const form = reactive({
            date: new Date().toISOString().split('T')[0],
            ref_no: '',
            description: '',
            category: 'Umum',
            kas_in: 0,
            kas_out: 0,
            ikrom_in: 0,
            ikrom_out: 0,
            pen_in: 0,
            pen_out: 0
        });

        const newCategory = reactive({
            name: '',
            type: 'both',
            pos: 'all'
        });

        // Report Settings for Printing
        const reportSettings = reactive({
            orgName: 'LEMBAGA / YAYASAN / PENGURUS KAS',
            address: 'Jl. Pemuda No. 123, Kota / Kabupaten',
            title: 'LAPORAN REKAPITULASI BUKU KAS TABELAR',
            subtitle: 'Pos Kas Umum, Dana Ikrom, dan Dana Pen',
            signerPlace: 'Indonesia',
            signerDate: new Date().toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' }),
            signer1Role: 'Ketua / Pimpinan',
            signer1Name: '( ........................................ )',
            signer2Role: 'Bendahara',
            signer2Name: '( ........................................ )'
        });

        // Chart instances
        let monthlyChartInstance = null;
        let posChartInstance = null;
        let trendChartInstance = null;

        // Number Formatter
        const formatRupiah = (val) => {
            if (val === undefined || val === null || isNaN(val)) return 'Rp 0';
            const num = Number(val);
            return 'Rp ' + Math.round(num).toLocaleString('id-ID');
        };

        const formatNumberInput = (val) => {
            if (!val && val !== 0) return '';
            return Number(val).toLocaleString('id-ID');
        };

        const parseNumber = (val) => {
            if (!val) return 0;
            if (typeof val === 'number') return val;
            const cleaned = val.toString().replace(/[^0-9.-]+/g, '');
            return parseFloat(cleaned) || 0;
        };

        const formatDate = (dateStr) => {
            if (!dateStr) return '-';
            try {
                const parts = dateStr.split('-');
                if (parts.length === 3) {
                    return `${parts[2]}/${parts[1]}/${parts[0]}`;
                }
                return dateStr;
            } catch (e) {
                return dateStr;
            }
        };

        // Computed Totals on Filtered Table
        const tableTotals = computed(() => {
            let kasIn = 0, kasOut = 0;
            let ikromIn = 0, ikromOut = 0;
            let penIn = 0, penOut = 0;

            transactions.value.forEach(t => {
                kasIn += t.kas_in || 0;
                kasOut += t.kas_out || 0;
                ikromIn += t.ikrom_in || 0;
                ikromOut += t.ikrom_out || 0;
                penIn += t.pen_in || 0;
                penOut += t.pen_out || 0;
            });

            const kasNet = kasIn - kasOut;
            const ikromNet = ikromIn - ikromOut;
            const penNet = penIn - penOut;
            const totalNet = kasNet + ikromNet + penNet;

            const lastRow = transactions.value.length > 0 ? transactions.value[transactions.value.length - 1] : null;
            const finalKasBalance = lastRow ? lastRow.kas_balance : (summary.value.saldo_kas || 0);
            const finalIkromBalance = lastRow ? lastRow.ikrom_balance : (summary.value.saldo_ikrom || 0);
            const finalPenBalance = lastRow ? lastRow.pen_balance : (summary.value.saldo_pen || 0);
            const finalTotalBalance = lastRow ? lastRow.total_balance : (summary.value.total_saldo || 0);

            return {
                kasIn, kasOut, kasNet,
                ikromIn, ikromOut, ikromNet,
                penIn, penOut, penNet,
                totalNet,
                finalKasBalance,
                finalIkromBalance,
                finalPenBalance,
                finalTotalBalance
            };
        });

        // API Calls
        const fetchTransactions = async () => {
            loading.value = true;
            try {
                const params = new URLSearchParams();
                if (filters.startDate) params.append('start_date', filters.startDate);
                if (filters.endDate) params.append('end_date', filters.endDate);
                if (filters.category && filters.category !== 'Semua') params.append('category', filters.category);
                if (searchInput.value.trim()) params.append('search', searchInput.value.trim());
                if (filters.pos && filters.pos !== 'all') params.append('pos', filters.pos);

                const res = await fetch(`/api/transactions?${params.toString()}`);
                const json = await res.json();
                if (json.success) {
                    transactions.value = json.data || [];
                }
            } catch (err) {
                console.error('Error fetching transactions:', err);
                toast('Gagal memuat data transaksi', 'error');
            } finally {
                loading.value = false;
            }
        };

        const fetchSummary = async () => {
            try {
                const params = new URLSearchParams();
                if (filters.startDate) params.append('start_date', filters.startDate);
                if (filters.endDate) params.append('end_date', filters.endDate);

                const res = await fetch(`/api/summary?${params.toString()}`);
                const json = await res.json();
                if (json.success) {
                    summary.value = json.data;
                }
            } catch (err) {
                console.error('Error fetching summary:', err);
            }
        };

        const fetchCategories = async () => {
            try {
                const res = await fetch('/api/categories');
                const json = await res.json();
                if (json.success) {
                    categories.value = json.data || [];
                }
            } catch (err) {
                console.error('Error fetching categories:', err);
            }
        };

        const fetchChartData = async () => {
            try {
                const res = await fetch('/api/chart-data');
                const json = await res.json();
                if (json.success) {
                    chartData.value = json.data;
                    if (activeTab.value === 'charts') {
                        nextTick(() => renderCharts());
                    }
                }
            } catch (err) {
                console.error('Error fetching chart data:', err);
            }
        };

        const refreshAll = async () => {
            await Promise.all([
                fetchTransactions(),
                fetchSummary(),
                fetchCategories(),
                fetchChartData()
            ]);
        };

        // Filter Presets
        const setPeriod = (preset) => {
            filters.periodPreset = preset;
            const now = new Date();
            const year = now.getFullYear();
            const month = now.getMonth();

            if (preset === 'all') {
                filters.startDate = '';
                filters.endDate = '';
            } else if (preset === 'today') {
                const todayStr = now.toISOString().split('T')[0];
                filters.startDate = todayStr;
                filters.endDate = todayStr;
            } else if (preset === 'this_month') {
                const firstDay = new Date(year, month, 1);
                const lastDay = new Date(year, month + 1, 0);
                filters.startDate = firstDay.toISOString().split('T')[0];
                filters.endDate = lastDay.toISOString().split('T')[0];
            } else if (preset === 'last_month') {
                const firstDay = new Date(year, month - 1, 1);
                const lastDay = new Date(year, month, 0);
                filters.startDate = firstDay.toISOString().split('T')[0];
                filters.endDate = lastDay.toISOString().split('T')[0];
            } else if (preset === 'this_year') {
                const firstDay = new Date(year, 0, 1);
                const lastDay = new Date(year, 11, 31);
                filters.startDate = firstDay.toISOString().split('T')[0];
                filters.endDate = lastDay.toISOString().split('T')[0];
            }

            refreshAll();
        };

        const resetFilters = () => {
            filters.startDate = '';
            filters.endDate = '';
            filters.category = 'Semua';
            filters.pos = 'all';
            filters.periodPreset = 'all';
            searchInput.value = '';
            refreshAll();
        };

        // Modal Form Actions
        const openCreateModal = () => {
            isEditing.value = false;
            editingId.value = null;
            inputMode.value = 'single';

            // Reset form
            form.date = new Date().toISOString().split('T')[0];
            form.ref_no = generateRefNo();
            form.description = '';
            form.category = categories.value.length > 0 ? categories.value[0].name : 'Umum';
            form.kas_in = 0;
            form.kas_out = 0;
            form.ikrom_in = 0;
            form.ikrom_out = 0;
            form.pen_in = 0;
            form.pen_out = 0;

            singleForm.type = 'in';
            singleForm.pos = 'kas';
            singleForm.amount = '';

            showModal.value = true;
        };

        const openEditModal = (t) => {
            isEditing.value = true;
            editingId.value = t.id;
            inputMode.value = 'multi'; // default multi for full fidelity edit

            form.date = t.date;
            form.ref_no = t.ref_no || '';
            form.description = t.description;
            form.category = t.category || 'Umum';
            form.kas_in = t.kas_in || 0;
            form.kas_out = t.kas_out || 0;
            form.ikrom_in = t.ikrom_in || 0;
            form.ikrom_out = t.ikrom_out || 0;
            form.pen_in = t.pen_in || 0;
            form.pen_out = t.pen_out || 0;

            showModal.value = true;
        };

        const generateRefNo = () => {
            const date = new Date();
            const ym = `${date.getFullYear()}${String(date.getMonth() + 1).padStart(2, '0')}`;
            const rnd = Math.floor(100 + Math.random() * 900);
            return `TRX-${ym}-${rnd}`;
        };

        const closeModal = () => {
            showModal.value = false;
        };

        const saveTransaction = async () => {
            if (!form.description || form.description.trim() === '') {
                toast('Keterangan transaksi harus diisi!', 'warning');
                return;
            }

            // Sync single input mode to form values if single mode is active
            if (inputMode.value === 'single') {
                const amount = parseNumber(singleForm.amount);
                form.kas_in = 0;
                form.kas_out = 0;
                form.ikrom_in = 0;
                form.ikrom_out = 0;
                form.pen_in = 0;
                form.pen_out = 0;

                if (singleForm.pos === 'kas') {
                    if (singleForm.type === 'in') form.kas_in = amount;
                    else form.kas_out = amount;
                } else if (singleForm.pos === 'ikrom') {
                    if (singleForm.type === 'in') form.ikrom_in = amount;
                    else form.ikrom_out = amount;
                } else if (singleForm.pos === 'pen') {
                    if (singleForm.type === 'in') form.pen_in = amount;
                    else form.pen_out = amount;
                }
            } else {
                // Ensure numbers
                form.kas_in = parseNumber(form.kas_in);
                form.kas_out = parseNumber(form.kas_out);
                form.ikrom_in = parseNumber(form.ikrom_in);
                form.ikrom_out = parseNumber(form.ikrom_out);
                form.pen_in = parseNumber(form.pen_in);
                form.pen_out = parseNumber(form.pen_out);
            }

            const totalNominal = (form.kas_in + form.kas_out + form.ikrom_in + form.ikrom_out + form.pen_in + form.pen_out);
            if (totalNominal <= 0) {
                toast('Harap masukkan nominal transaksi (tidak boleh Rp 0)', 'warning');
                return;
            }

            try {
                let res;
                if (isEditing.value) {
                    res = await fetch(`/api/transactions/${editingId.value}`, {
                        method: 'PUT',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify(form)
                    });
                } else {
                    res = await fetch('/api/transactions', {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify(form)
                    });
                }

                const json = await res.json();
                if (json.success) {
                    toast(json.message || 'Transaksi berhasil disimpan', 'success');
                    closeModal();
                    refreshAll();
                } else {
                    toast(json.error || 'Gagal menyimpan transaksi', 'error');
                }
            } catch (err) {
                console.error('Error saving transaction:', err);
                toast('Terjadi kesalahan saat menyimpan', 'error');
            }
        };

        const deleteTransaction = async (id, desc) => {
            Swal.fire({
                title: 'Hapus Transaksi?',
                html: `Apakah Anda yakin ingin menghapus transaksi:<br><b>"${desc}"</b>?`,
                icon: 'warning',
                showCancelButton: true,
                confirmButtonColor: '#ef4444',
                cancelButtonColor: '#64748b',
                confirmButtonText: 'Ya, Hapus!',
                cancelButtonText: 'Batal'
            }).then(async (result) => {
                if (result.isConfirmed) {
                    try {
                        const res = await fetch(`/api/transactions/${id}`, { method: 'DELETE' });
                        const json = await res.json();
                        if (json.success) {
                            toast('Transaksi berhasil dihapus', 'success');
                            refreshAll();
                        } else {
                            toast(json.error || 'Gagal menghapus transaksi', 'error');
                        }
                    } catch (err) {
                        toast('Terjadi kesalahan saat menghapus', 'error');
                    }
                }
            });
        };

        // Export and Seed Actions
        const exportCSV = () => {
            const params = new URLSearchParams();
            if (filters.startDate) params.append('start_date', filters.startDate);
            if (filters.endDate) params.append('end_date', filters.endDate);
            if (filters.category && filters.category !== 'Semua') params.append('category', filters.category);
            if (searchInput.value.trim()) params.append('search', searchInput.value.trim());
            if (filters.pos && filters.pos !== 'all') params.append('pos', filters.pos);

            window.location.href = `/api/export/csv?${params.toString()}`;
        };

        const seedDemoData = async () => {
            Swal.fire({
                title: 'Muat Data Contoh?',
                text: 'Data saat ini akan digantikan dengan data simulasi Kas, Ikrom, dan Pen.',
                icon: 'question',
                showCancelButton: true,
                confirmButtonColor: '#059669',
                cancelButtonColor: '#64748b',
                confirmButtonText: 'Ya, Muat Contoh',
                cancelButtonText: 'Batal'
            }).then(async (result) => {
                if (result.isConfirmed) {
                    loading.value = true;
                    try {
                        const res = await fetch('/api/seed', { method: 'POST' });
                        const json = await res.json();
                        if (json.success) {
                            toast('Data demo berhasil dimuat!', 'success');
                            refreshAll();
                        }
                    } catch (err) {
                        toast('Gagal memuat data demo', 'error');
                    } finally {
                        loading.value = false;
                    }
                }
            });
        };

        const resetAllData = async () => {
            Swal.fire({
                title: 'Reset Semua Data?',
                text: 'PERINGATAN: Semua transaksi akan dihapus secara permanen!',
                icon: 'warning',
                showCancelButton: true,
                confirmButtonColor: '#dc2626',
                cancelButtonColor: '#64748b',
                confirmButtonText: 'Ya, Kosongkan Data',
                cancelButtonText: 'Batal'
            }).then(async (result) => {
                if (result.isConfirmed) {
                    try {
                        const res = await fetch('/api/reset', { method: 'POST' });
                        const json = await res.json();
                        if (json.success) {
                            toast('Semua data berhasil dibersihkan', 'success');
                            refreshAll();
                        }
                    } catch (err) {
                        toast('Gagal mereset data', 'error');
                    }
                }
            });
        };

        // Category Management
        const addCategory = async () => {
            if (!newCategory.name.trim()) {
                toast('Nama kategori harus diisi', 'warning');
                return;
            }
            try {
                const res = await fetch('/api/categories', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(newCategory)
                });
                const json = await res.json();
                if (json.success) {
                    toast('Kategori berhasil ditambahkan', 'success');
                    newCategory.name = '';
                    fetchCategories();
                } else {
                    toast(json.error || 'Gagal menambah kategori', 'error');
                }
            } catch (err) {
                toast('Terjadi kesalahan', 'error');
            }
        };

        const deleteCategory = async (id, name) => {
            try {
                const res = await fetch(`/api/categories/${id}`, { method: 'DELETE' });
                const json = await res.json();
                if (json.success) {
                    toast(`Kategori "${name}" dihapus`, 'success');
                    fetchCategories();
                }
            } catch (err) {
                toast('Gagal menghapus kategori', 'error');
            }
        };

        // Print & PDF
        const openPrintModal = () => {
            showPrintModal.value = true;
        };

        const triggerPrint = () => {
            window.print();
        };

        // Toast notification utility
        const toast = (msg, icon = 'info') => {
            const Toast = Swal.mixin({
                toast: true,
                position: 'top-end',
                showConfirmButton: false,
                timer: 2500,
                timerProgressBar: true
            });
            Toast.fire({
                icon: icon,
                title: msg
            });
        };

        // Chart.js Visualizations
        const renderCharts = () => {
            if (!chartData.value) return;

            // Monthly In/Out Bar Chart
            const monthlyCtx = document.getElementById('monthlyChart');
            if (monthlyCtx) {
                if (monthlyChartInstance) monthlyChartInstance.destroy();

                const labels = chartData.value.monthly.map(m => m.month);
                const kasInData = chartData.value.monthly.map(m => m.kas_in);
                const kasOutData = chartData.value.monthly.map(m => m.kas_out);
                const ikromInData = chartData.value.monthly.map(m => m.ikrom_in);
                const ikromOutData = chartData.value.monthly.map(m => m.ikrom_out);
                const penInData = chartData.value.monthly.map(m => m.pen_in);
                const penOutData = chartData.value.monthly.map(m => m.pen_out);

                monthlyChartInstance = new Chart(monthlyCtx, {
                    type: 'bar',
                    data: {
                        labels: labels.length > 0 ? labels : ['Belum ada data'],
                        datasets: [
                            { label: 'Kas Masuk', data: kasInData, backgroundColor: '#10b981' },
                            { label: 'Kas Keluar', data: kasOutData, backgroundColor: '#f87171' },
                            { label: 'Ikrom Masuk', data: ikromInData, backgroundColor: '#3b82f6' },
                            { label: 'Ikrom Keluar', data: ikromOutData, backgroundColor: '#93c5fd' },
                            { label: 'Pen Masuk', data: penInData, backgroundColor: '#f59e0b' },
                            { label: 'Pen Keluar', data: penOutData, backgroundColor: '#fde68a' }
                        ]
                    },
                    options: {
                        responsive: true,
                        maintainAspectRatio: false,
                        plugins: {
                            legend: { position: 'bottom' },
                            tooltip: {
                                callbacks: {
                                    label: (ctx) => `${ctx.dataset.label}: ${formatRupiah(ctx.raw)}`
                                }
                            }
                        },
                        scales: {
                            y: {
                                ticks: {
                                    callback: (val) => 'Rp ' + (val >= 1000000 ? (val / 1000000).toFixed(1) + ' jt' : val.toLocaleString('id-ID'))
                                }
                            }
                        }
                    }
                });
            }

            // Pos Distribution Doughnut Chart
            const posCtx = document.getElementById('posChart');
            if (posCtx) {
                if (posChartInstance) posChartInstance.destroy();

                const pos = chartData.value.pos_distribution || { kas: 0, ikrom: 0, pen: 0 };
                posChartInstance = new Chart(posCtx, {
                    type: 'doughnut',
                    data: {
                        labels: ['Saldo Kas Utama', 'Saldo Dana Ikrom', 'Saldo Dana Pen'],
                        datasets: [{
                            data: [Math.max(0, pos.kas), Math.max(0, pos.ikrom), Math.max(0, pos.pen)],
                            backgroundColor: ['#10b981', '#3b82f6', '#f59e0b']
                        }]
                    },
                    options: {
                        responsive: true,
                        maintainAspectRatio: false,
                        plugins: {
                            legend: { position: 'bottom' },
                            tooltip: {
                                callbacks: {
                                    label: (ctx) => `${ctx.label}: ${formatRupiah(ctx.raw)}`
                                }
                            }
                        }
                    }
                });
            }

            // Cumulative Balance Trend Line Chart
            const trendCtx = document.getElementById('trendChart');
            if (trendCtx) {
                if (trendChartInstance) trendChartInstance.destroy();

                let running = 0;
                const trendLabels = [];
                const trendValues = [];

                transactions.value.forEach(t => {
                    trendLabels.push(`${formatDate(t.date)}`);
                    trendValues.push(t.total_balance);
                });

                trendChartInstance = new Chart(trendCtx, {
                    type: 'line',
                    data: {
                        labels: trendLabels.length > 0 ? trendLabels : ['Data Kosong'],
                        datasets: [{
                            label: 'Total Saldo Akumulasi',
                            data: trendValues.length > 0 ? trendValues : [0],
                            borderColor: '#8b5cf6',
                            backgroundColor: 'rgba(139, 92, 246, 0.1)',
                            fill: true,
                            tension: 0.3,
                            pointRadius: 4
                        }]
                    },
                    options: {
                        responsive: true,
                        maintainAspectRatio: false,
                        plugins: {
                            legend: { position: 'bottom' },
                            tooltip: {
                                callbacks: {
                                    label: (ctx) => `Total Saldo: ${formatRupiah(ctx.raw)}`
                                }
                            }
                        },
                        scales: {
                            y: {
                                ticks: {
                                    callback: (val) => 'Rp ' + (val >= 1000000 ? (val / 1000000).toFixed(1) + ' jt' : val.toLocaleString('id-ID'))
                                }
                            }
                        }
                    }
                });
            }
        };

        // Watchers
        watch(activeTab, (newTab) => {
            if (newTab === 'charts') {
                nextTick(() => {
                    fetchChartData();
                });
            }
        });

        // Search debounce
        let searchTimeout = null;
        watch(searchInput, () => {
            clearTimeout(searchTimeout);
            searchTimeout = setTimeout(() => {
                fetchTransactions();
            }, 300);
        });

        // Lifecycle
        onMounted(() => {
            refreshAll();
        });

        return {
            transactions,
            summary,
            categories,
            chartData,
            activeTab,
            loading,
            searchInput,
            showModal,
            isEditing,
            showPrintModal,
            showCategoryModal,
            filters,
            form,
            singleForm,
            inputMode,
            newCategory,
            reportSettings,
            tableTotals,
            formatRupiah,
            formatNumberInput,
            formatDate,
            setPeriod,
            resetFilters,
            fetchTransactions,
            refreshAll,
            openCreateModal,
            openEditModal,
            closeModal,
            saveTransaction,
            deleteTransaction,
            exportCSV,
            seedDemoData,
            resetAllData,
            addCategory,
            deleteCategory,
            openPrintModal,
            triggerPrint
        };
    }
}).mount('#app');
