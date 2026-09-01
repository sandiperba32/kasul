const { createApp, ref, reactive, computed, onMounted, watch, nextTick } = Vue;

createApp({
    setup() {
        // ─────────── State ───────────
        const transactions = ref([]);
        const students = ref([]);
        const summary = ref({
            total_kas_in: 0, total_kas_out: 0, saldo_kas: 0,
            total_ikrom_in: 0, total_ikrom_out: 0, saldo_ikrom: 0,
            total_masuk: 0, total_keluar: 0, total_saldo: 0,
            transaction_count: 0, student_count: 0
        });
        const chartData = ref(null);
        const activeTab = ref('ledger');  // 'ledger' | 'students' | 'charts'
        const loading = ref(false);
        const studentLoading = ref(false);
        const searchInput = ref('');

        // ─────────── Modals ───────────
        const showModal = ref(false);
        const isEditing = ref(false);
        const editingId = ref(null);
        const showPrintModal = ref(false);
        const showStudentModal = ref(false);
        const isEditingStudent = ref(false);
        const editingStudentId = ref(null);

        // ─────────── Filters ───────────
        const filters = reactive({ startDate: '', endDate: '', pos: 'all', periodPreset: 'all' });
        const studentSearch = ref('');

        // ─────────── Transaction Form ───────────
        const form = reactive({
            date: new Date().toISOString().split('T')[0],
            ref_no: '',
            name: '',
            description: '',
            type: 'in',
            kas_amount: 0,
            ikrom_amount: 0,
            kas_formatted: '',
            ikrom_formatted: ''
        });

        // ─────────── Student Form (hanya Nama & Orang Tua) ───────────
        const studentForm = reactive({ name: '', parent: '' });

        // ─────────── Report Settings ───────────
        const reportSettings = reactive({
            orgName: 'LEMBAGA / YAYASAN / PENGURUS KAS',
            address: 'Jl. Pemuda No. 123, Kota / Kabupaten',
            title: 'LAPORAN BUKU KAS & DANA IKROM',
            signerPlace: 'Indonesia',
            signerDate: new Date().toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' }),
            signer1Role: 'Ketua / Pimpinan', signer1Name: '( ........................... )',
            signer2Role: 'Bendahara',         signer2Name: '( ........................... )',
        });

        // Chart instances
        let monthlyChartInstance = null, posChartInstance = null, trendChartInstance = null;

        // ─────────── Helpers ───────────
        const formatRupiah = (val) => {
            if (!val && val !== 0) return 'Rp 0';
            return 'Rp ' + Math.round(Number(val)).toLocaleString('id-ID');
        };

        const parseFormatted = (val) => {
            if (!val) return 0;
            return parseInt(val.toString().replace(/[^0-9]/g, ''), 10) || 0;
        };

        const formatDate = (d) => {
            if (!d) return '-';
            const p = d.split('-');
            return p.length === 3 ? `${p[2]}/${p[1]}/${p[0]}` : d;
        };

        const formTotal = computed(() => (form.kas_amount || 0) + (form.ikrom_amount || 0));

        const tableTotals = computed(() => {
            let kasIn = 0, kasOut = 0, ikromIn = 0, ikromOut = 0;
            transactions.value.forEach(t => {
                kasIn += t.kas_in || 0; kasOut += t.kas_out || 0;
                ikromIn += t.ikrom_in || 0; ikromOut += t.ikrom_out || 0;
            });
            const last = transactions.value[transactions.value.length - 1];
            return {
                kasIn, kasOut, ikromIn, ikromOut,
                finalKasBalance:   last ? last.kas_balance   : summary.value.saldo_kas,
                finalIkromBalance: last ? last.ikrom_balance : summary.value.saldo_ikrom,
                finalTotalBalance: last ? last.total_balance : summary.value.total_saldo,
            };
        });

        // ─────────── Kas Input Masking ───────────
        const onKasInput = (e) => {
            form.kas_amount = parseFormatted(e.target.value);
            form.kas_formatted = form.kas_amount > 0 ? form.kas_amount.toLocaleString('id-ID') : '';
        };
        const onIkromInput = (e) => {
            form.ikrom_amount = parseFormatted(e.target.value);
            form.ikrom_formatted = form.ikrom_amount > 0 ? form.ikrom_amount.toLocaleString('id-ID') : '';
        };

        // ─────────── Select student name in transaction form ───────────
        const onSelectStudentName = (name) => {
            form.name = name;
            if (!form.description) form.description = `Iuran Kas / Donasi (${name})`;
        };

        // ─────────── Fetch API ───────────
        const fetchStudents = async () => {
            studentLoading.value = true;
            try {
                const params = new URLSearchParams();
                if (studentSearch.value.trim()) params.append('search', studentSearch.value.trim());
                const res = await fetch(`/api/students?${params}`);
                const json = await res.json();
                if (json.success) students.value = json.data || [];
            } catch (e) { console.error(e); }
            finally { studentLoading.value = false; }
        };

        const fetchTransactions = async () => {
            loading.value = true;
            try {
                const params = new URLSearchParams();
                if (filters.startDate) params.append('start_date', filters.startDate);
                if (filters.endDate)   params.append('end_date',   filters.endDate);
                if (searchInput.value.trim()) params.append('search', searchInput.value.trim());
                if (filters.pos !== 'all')    params.append('pos', filters.pos);
                const res = await fetch(`/api/transactions?${params}`);
                const json = await res.json();
                if (json.success) transactions.value = json.data || [];
            } catch (e) { toast('Gagal memuat transaksi', 'error'); }
            finally { loading.value = false; }
        };

        const fetchSummary = async () => {
            try {
                const params = new URLSearchParams();
                if (filters.startDate) params.append('start_date', filters.startDate);
                if (filters.endDate)   params.append('end_date',   filters.endDate);
                const res = await fetch(`/api/summary?${params}`);
                const json = await res.json();
                if (json.success) summary.value = json.data;
            } catch (e) { console.error(e); }
        };

        const fetchChartData = async () => {
            try {
                const res = await fetch('/api/chart-data');
                const json = await res.json();
                if (json.success) { chartData.value = json.data; nextTick(() => renderCharts()); }
            } catch (e) { console.error(e); }
        };

        const refreshAll = async () => {
            await Promise.all([fetchTransactions(), fetchStudents(), fetchSummary()]);
        };

        // ─────────── Period Presets ───────────
        const setPeriod = (preset) => {
            filters.periodPreset = preset;
            const now = new Date(), y = now.getFullYear(), mo = now.getMonth();
            if (preset === 'all')        { filters.startDate = ''; filters.endDate = ''; }
            else if (preset === 'today') { const t = now.toISOString().split('T')[0]; filters.startDate = t; filters.endDate = t; }
            else if (preset === 'this_month') {
                filters.startDate = new Date(y, mo, 1).toISOString().split('T')[0];
                filters.endDate   = new Date(y, mo + 1, 0).toISOString().split('T')[0];
            } else if (preset === 'last_month') {
                filters.startDate = new Date(y, mo - 1, 1).toISOString().split('T')[0];
                filters.endDate   = new Date(y, mo, 0).toISOString().split('T')[0];
            } else if (preset === 'this_year') {
                filters.startDate = new Date(y, 0, 1).toISOString().split('T')[0];
                filters.endDate   = new Date(y, 11, 31).toISOString().split('T')[0];
            }
            refreshAll();
        };

        const resetFilters = () => {
            Object.assign(filters, { startDate: '', endDate: '', pos: 'all', periodPreset: 'all' });
            searchInput.value = '';
            refreshAll();
        };

        // ─────────── Transaction Modal ───────────
        const generateRefNo = () => {
            const d = new Date();
            return `TRX-${d.getFullYear()}${String(d.getMonth()+1).padStart(2,'0')}-${Math.floor(100+Math.random()*900)}`;
        };

        const openCreateModal = () => {
            isEditing.value = false; editingId.value = null;
            Object.assign(form, {
                date: new Date().toISOString().split('T')[0],
                ref_no: generateRefNo(), name: '', description: '', type: 'in',
                kas_amount: 0, ikrom_amount: 0, kas_formatted: '', ikrom_formatted: ''
            });
            showModal.value = true;
        };

        const openEditModal = (t) => {
            isEditing.value = true; editingId.value = t.id;
            const isIn = t.kas_in > 0 || t.ikrom_in > 0;
            const kasAmt  = isIn ? (t.kas_in  || 0) : (t.kas_out  || 0);
            const ikromAmt = isIn ? (t.ikrom_in || 0) : (t.ikrom_out || 0);
            Object.assign(form, {
                date: t.date, ref_no: t.ref_no || '', name: t.name || '',
                description: t.description, type: isIn ? 'in' : 'out',
                kas_amount: kasAmt, ikrom_amount: ikromAmt,
                kas_formatted:  kasAmt  > 0 ? kasAmt.toLocaleString('id-ID')  : '',
                ikrom_formatted: ikromAmt > 0 ? ikromAmt.toLocaleString('id-ID') : '',
            });
            showModal.value = true;
        };

        const closeModal = () => { showModal.value = false; };

        const saveTransaction = async () => {
            if (!form.description.trim()) { toast('Keterangan wajib diisi!', 'warning'); return; }
            const kasVal   = parseFormatted(form.kas_formatted);
            const ikromVal = parseFormatted(form.ikrom_formatted);
            if (kasVal + ikromVal <= 0) { toast('Isi nominal Kas atau Ikrom terlebih dahulu', 'warning'); return; }

            const payload = {
                date: form.date, ref_no: form.ref_no, name: form.name.trim(),
                description: form.description.trim(),
                kas_in:    form.type === 'in'  ? kasVal  : 0,
                kas_out:   form.type === 'out' ? kasVal  : 0,
                ikrom_in:  form.type === 'in'  ? ikromVal : 0,
                ikrom_out: form.type === 'out' ? ikromVal : 0,
            };

            try {
                const res = await fetch(
                    isEditing.value ? `/api/transactions/${editingId.value}` : '/api/transactions',
                    { method: isEditing.value ? 'PUT' : 'POST', headers: {'Content-Type':'application/json'}, body: JSON.stringify(payload) }
                );
                const json = await res.json();
                if (json.success) { toast(json.message, 'success'); closeModal(); refreshAll(); }
                else toast(json.error || 'Gagal menyimpan', 'error');
            } catch (e) { toast('Terjadi kesalahan', 'error'); }
        };

        const deleteTransaction = (id, desc) => {
            Swal.fire({ title: 'Hapus Transaksi?', html: `<b>"${desc}"</b>`, icon: 'warning',
                showCancelButton: true, confirmButtonColor: '#ef4444', cancelButtonColor: '#64748b',
                confirmButtonText: 'Ya, Hapus!', cancelButtonText: 'Batal'
            }).then(async r => {
                if (!r.isConfirmed) return;
                const res = await fetch(`/api/transactions/${id}`, { method: 'DELETE' });
                const json = await res.json();
                if (json.success) { toast('Transaksi dihapus', 'success'); refreshAll(); }
                else toast(json.error, 'error');
            });
        };

        // ─────────── Student Modal (hanya Nama & Orang Tua) ───────────
        const openCreateStudentModal = () => {
            isEditingStudent.value = false; editingStudentId.value = null;
            Object.assign(studentForm, { name: '', parent: '' });
            showStudentModal.value = true;
        };

        const openEditStudentModal = (s) => {
            isEditingStudent.value = true; editingStudentId.value = s.id;
            Object.assign(studentForm, { name: s.name, parent: s.parent || '' });
            showStudentModal.value = true;
        };

        const closeStudentModal = () => { showStudentModal.value = false; };

        const saveStudent = async () => {
            if (!studentForm.name.trim()) { toast('Nama siswa wajib diisi!', 'warning'); return; }
            try {
                const res = await fetch(
                    isEditingStudent.value ? `/api/students/${editingStudentId.value}` : '/api/students',
                    { method: isEditingStudent.value ? 'PUT' : 'POST', headers: {'Content-Type':'application/json'}, body: JSON.stringify(studentForm) }
                );
                const json = await res.json();
                if (json.success) { toast(json.message, 'success'); closeStudentModal(); fetchStudents(); fetchSummary(); }
                else toast(json.error || 'Gagal menyimpan', 'error');
            } catch (e) { toast('Terjadi kesalahan', 'error'); }
        };

        const deleteStudent = (id, name) => {
            Swal.fire({ title: 'Hapus Siswa?', html: `<b>"${name}"</b>`, icon: 'warning',
                showCancelButton: true, confirmButtonColor: '#ef4444', cancelButtonColor: '#64748b',
                confirmButtonText: 'Ya, Hapus!', cancelButtonText: 'Batal'
            }).then(async r => {
                if (!r.isConfirmed) return;
                const res = await fetch(`/api/students/${id}`, { method: 'DELETE' });
                const json = await res.json();
                if (json.success) { toast('Siswa dihapus', 'success'); fetchStudents(); fetchSummary(); }
                else toast(json.error, 'error');
            });
        };

        const filterByStudent = (name) => {
            activeTab.value = 'ledger';
            searchInput.value = name;
            filters.periodPreset = 'all';
            filters.startDate = ''; filters.endDate = '';
            fetchTransactions();
        };

        // ─────────── Export / Seed / Reset ───────────
        const exportCSV = () => {
            const params = new URLSearchParams();
            if (filters.startDate) params.append('start_date', filters.startDate);
            if (filters.endDate)   params.append('end_date',   filters.endDate);
            if (searchInput.value.trim()) params.append('search', searchInput.value.trim());
            if (filters.pos !== 'all') params.append('pos', filters.pos);
            window.location.href = `/api/export/csv?${params}`;
        };

        const seedDemoData = () => {
            Swal.fire({ title: 'Muat Data Contoh?', icon: 'question', showCancelButton: true,
                confirmButtonColor: '#059669', confirmButtonText: 'Ya, Muat', cancelButtonText: 'Batal'
            }).then(async r => {
                if (!r.isConfirmed) return;
                loading.value = true;
                const res = await fetch('/api/seed', { method: 'POST' });
                const json = await res.json();
                if (json.success) { toast('Data demo dimuat!', 'success'); refreshAll(); }
                loading.value = false;
            });
        };

        const resetAllData = () => {
            Swal.fire({ title: 'Reset Semua Transaksi?', text: 'Semua transaksi akan dihapus!', icon: 'warning',
                showCancelButton: true, confirmButtonColor: '#dc2626', confirmButtonText: 'Ya, Hapus Semua!', cancelButtonText: 'Batal'
            }).then(async r => {
                if (!r.isConfirmed) return;
                const res = await fetch('/api/reset', { method: 'POST' });
                const json = await res.json();
                if (json.success) { toast('Data transaksi dibersihkan', 'success'); refreshAll(); }
            });
        };

        const openPrintModal = () => { showPrintModal.value = true; };
        const triggerPrint  = () => { window.print(); };

        const toast = (msg, icon = 'info') => {
            Swal.mixin({ toast: true, position: 'top-end', showConfirmButton: false, timer: 2500, timerProgressBar: true })
                .fire({ icon, title: msg });
        };

        // ─────────── Charts ───────────
        const renderCharts = () => {
            if (!chartData.value) return;
            const monthly = chartData.value.monthly || [];

            const monthlyCtx = document.getElementById('monthlyChart');
            if (monthlyCtx) {
                if (monthlyChartInstance) monthlyChartInstance.destroy();
                monthlyChartInstance = new Chart(monthlyCtx, {
                    type: 'bar',
                    data: {
                        labels: monthly.map(m => m.month) || ['Belum ada data'],
                        datasets: [
                            { label: 'Kas Masuk',   data: monthly.map(m => m.kas_in),   backgroundColor: '#10b981' },
                            { label: 'Kas Keluar',  data: monthly.map(m => m.kas_out),  backgroundColor: '#f87171' },
                            { label: 'Ikrom Masuk', data: monthly.map(m => m.ikrom_in), backgroundColor: '#3b82f6' },
                            { label: 'Ikrom Keluar',data: monthly.map(m => m.ikrom_out),backgroundColor: '#93c5fd' },
                        ]
                    },
                    options: { responsive: true, maintainAspectRatio: false,
                        plugins: { legend: { position: 'bottom' }, tooltip: { callbacks: { label: ctx => `${ctx.dataset.label}: ${formatRupiah(ctx.raw)}` } } },
                        scales: { y: { ticks: { callback: v => 'Rp ' + (v >= 1e6 ? (v/1e6).toFixed(1)+'jt' : v.toLocaleString('id-ID')) } } }
                    }
                });
            }

            const posCtx = document.getElementById('posChart');
            if (posCtx) {
                if (posChartInstance) posChartInstance.destroy();
                const pos = chartData.value.pos_distribution || {};
                posChartInstance = new Chart(posCtx, {
                    type: 'doughnut',
                    data: { labels: ['Saldo Kas', 'Saldo Ikrom'],
                        datasets: [{ data: [Math.max(0, pos.kas||0), Math.max(0, pos.ikrom||0)], backgroundColor: ['#10b981','#3b82f6'] }] },
                    options: { responsive: true, maintainAspectRatio: false,
                        plugins: { legend: { position: 'bottom' }, tooltip: { callbacks: { label: ctx => `${ctx.label}: ${formatRupiah(ctx.raw)}` } } }
                    }
                });
            }

            const trendCtx = document.getElementById('trendChart');
            if (trendCtx) {
                if (trendChartInstance) trendChartInstance.destroy();
                trendChartInstance = new Chart(trendCtx, {
                    type: 'line',
                    data: {
                        labels: transactions.value.map(t => formatDate(t.date)),
                        datasets: [{ label: 'Total Saldo Akumulasi', data: transactions.value.map(t => t.total_balance),
                            borderColor: '#8b5cf6', backgroundColor: 'rgba(139,92,246,0.1)', fill: true, tension: 0.3, pointRadius: 4 }]
                    },
                    options: { responsive: true, maintainAspectRatio: false,
                        plugins: { legend: { position: 'bottom' }, tooltip: { callbacks: { label: ctx => `Saldo: ${formatRupiah(ctx.raw)}` } } },
                        scales: { y: { ticks: { callback: v => 'Rp ' + (v >= 1e6 ? (v/1e6).toFixed(1)+'jt' : v.toLocaleString('id-ID')) } } }
                    }
                });
            }
        };

        // ─────────── Watchers ───────────
        watch(activeTab, (tab) => {
            if (tab === 'charts') nextTick(() => fetchChartData());
            if (tab === 'students') fetchStudents();
        });

        let txSearchTimer = null;
        watch(searchInput, () => {
            clearTimeout(txSearchTimer);
            txSearchTimer = setTimeout(fetchTransactions, 300);
        });

        let stuSearchTimer = null;
        watch(studentSearch, () => {
            clearTimeout(stuSearchTimer);
            stuSearchTimer = setTimeout(fetchStudents, 300);
        });

        onMounted(() => refreshAll());

        return {
            transactions, students, summary, chartData,
            activeTab, loading, studentLoading, searchInput, studentSearch,
            showModal, isEditing, showPrintModal, showStudentModal, isEditingStudent,
            form, formTotal, studentForm, filters, reportSettings, tableTotals,
            formatRupiah, formatDate,
            onKasInput, onIkromInput, onSelectStudentName,
            setPeriod, resetFilters, refreshAll, fetchTransactions, fetchStudents,
            openCreateModal, openEditModal, closeModal, saveTransaction, deleteTransaction,
            openCreateStudentModal, openEditStudentModal, closeStudentModal, saveStudent, deleteStudent,
            filterByStudent, exportCSV, seedDemoData, resetAllData,
            openPrintModal, triggerPrint,
        };
    }
}).mount('#app');
