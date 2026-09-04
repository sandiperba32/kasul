const { createApp, ref, reactive, computed, onMounted, watch, nextTick } = Vue;

createApp({
    setup() {
        // ─────────── State ───────────
        const kasBooks = ref([]);
        const selectedKas = ref(null);

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

        const showKasModal = ref(false);
        const isEditingKas = ref(false);
        const editingKasId = ref(null);

        // ─────────── Filters & Search ───────────
        const filters = reactive({ startDate: '', endDate: '', pos: 'all', periodPreset: 'all' });
        const studentSearch = ref('');

        // ─────────── Pagination ───────────
        const txPerPage = ref(20);
        const txCurrentPage = ref(1);

        const paginatedTransactions = computed(() => {
            if (!transactions.value || transactions.value.length === 0) return [];
            if (txPerPage.value === 'all' || !txPerPage.value) return transactions.value;
            const perPage = Number(txPerPage.value);
            const maxPage = Math.ceil(transactions.value.length / perPage) || 1;
            const page = Math.min(Math.max(1, txCurrentPage.value), maxPage);
            const start = (page - 1) * perPage;
            return transactions.value.slice(start, start + perPage);
        });

        const totalTxPages = computed(() => {
            if (txPerPage.value === 'all' || !txPerPage.value || !transactions.value || transactions.value.length === 0) return 1;
            return Math.ceil(transactions.value.length / Number(txPerPage.value));
        });

        const txStartItem = computed(() => {
            if (!transactions.value || transactions.value.length === 0) return 0;
            if (txPerPage.value === 'all') return 1;
            const perPage = Number(txPerPage.value);
            const maxPage = Math.ceil(transactions.value.length / perPage) || 1;
            const page = Math.min(Math.max(1, txCurrentPage.value), maxPage);
            return (page - 1) * perPage + 1;
        });

        const txEndItem = computed(() => {
            if (!transactions.value || transactions.value.length === 0) return 0;
            if (txPerPage.value === 'all') return transactions.value.length;
            const perPage = Number(txPerPage.value);
            const maxPage = Math.ceil(transactions.value.length / perPage) || 1;
            const page = Math.min(Math.max(1, txCurrentPage.value), maxPage);
            return Math.min(page * perPage, transactions.value.length);
        });

        const studentPerPage = ref(20);
        const studentCurrentPage = ref(1);

        const paginatedStudents = computed(() => {
            if (studentPerPage.value === 'all' || !studentPerPage.value) return students.value;
            const perPage = Number(studentPerPage.value);
            const start = (studentCurrentPage.value - 1) * perPage;
            return students.value.slice(start, start + perPage);
        });

        const totalStudentPages = computed(() => {
            if (studentPerPage.value === 'all' || !studentPerPage.value || students.value.length === 0) return 1;
            return Math.ceil(students.value.length / Number(studentPerPage.value));
        });

        const studentStartItem = computed(() => {
            if (students.value.length === 0) return 0;
            if (studentPerPage.value === 'all') return 1;
            return (studentCurrentPage.value - 1) * Number(studentPerPage.value) + 1;
        });

        const studentEndItem = computed(() => {
            if (studentPerPage.value === 'all') return students.value.length;
            return Math.min(studentCurrentPage.value * Number(studentPerPage.value), students.length);
        });

        const kasPerPage = ref(20);
        const kasCurrentPage = ref(1);

        const paginatedKasBooks = computed(() => {
            if (kasPerPage.value === 'all' || !kasPerPage.value) return kasBooks.value;
            const perPage = Number(kasPerPage.value);
            const start = (kasCurrentPage.value - 1) * perPage;
            return kasBooks.value.slice(start, start + perPage);
        });

        const totalKasPages = computed(() => {
            if (kasPerPage.value === 'all' || !kasPerPage.value || kasBooks.value.length === 0) return 1;
            return Math.ceil(kasBooks.value.length / Number(kasPerPage.value));
        });

        const kasStartItem = computed(() => {
            if (kasBooks.value.length === 0) return 0;
            if (kasPerPage.value === 'all') return 1;
            return (kasCurrentPage.value - 1) * Number(kasPerPage.value) + 1;
        });

        const kasEndItem = computed(() => {
            if (kasPerPage.value === 'all') return kasBooks.value.length;
            return Math.min(kasCurrentPage.value * Number(kasPerPage.value), kasBooks.value.length);
        });

        const getPageNumbers = (current, total) => {
            let pages = [];
            if (total <= 7) {
                for (let i = 1; i <= total; i++) pages.push(i);
            } else {
                if (current <= 4) {
                    pages = [1, 2, 3, 4, 5, '...', total];
                } else if (current >= total - 3) {
                    pages = [1, '...', total - 4, total - 3, total - 2, total - 1, total];
                } else {
                    pages = [1, '...', current - 1, current, current + 1, '...', total];
                }
            }
            return pages;
        };

        watch(txPerPage, () => { txCurrentPage.value = 1; });
        watch(studentPerPage, () => { studentCurrentPage.value = 1; });
        watch(kasPerPage, () => { kasCurrentPage.value = 1; });

        // ─────────── Forms ───────────
        const kasForm = reactive({ name: '', model_type: 1 });

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

        watch(() => form.name, (newName) => {
            const student = students.value.find(s => s.name === newName);
            if (student) {
                form.ref_no = student.halaqoh || '';
            }
        });

        const studentForm = reactive({ name: '', halaqoh: '' });

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

        const onKasInput = (e) => {
            form.kas_amount = parseFormatted(e.target.value);
            form.kas_formatted = form.kas_amount > 0 ? form.kas_amount.toLocaleString('id-ID') : '';
        };
        const onIkromInput = (e) => {
            form.ikrom_amount = parseFormatted(e.target.value);
            form.ikrom_formatted = form.ikrom_amount > 0 ? form.ikrom_amount.toLocaleString('id-ID') : '';
        };
        const onSelectStudentName = (name) => {
            form.name = name;
            if (!form.description) form.description = `Iuran Kas / Donasi (${name})`;
        };

        // ─────────── Kas Books API ───────────
        const fetchKasBooks = async () => {
            try {
                const res = await fetch('/api/kas_books');
                const json = await res.json();
                if (json.success) kasBooks.value = json.data || [];
            } catch (e) { console.error(e); }
        };

        const selectKas = (kas) => {
            selectedKas.value = kas;
            activeTab.value = 'ledger';
            resetFilters();
        };

        const backToMenu = () => {
            selectedKas.value = null;
        };

        const openCreateKasModal = () => {
            isEditingKas.value = false;
            editingKasId.value = null;
            kasForm.name = '';
            kasForm.model_type = 1;
            showKasModal.value = true;
        };

        const openEditKasModal = (kas) => {
            isEditingKas.value = true;
            editingKasId.value = kas.id;
            kasForm.name = kas.name;
            kasForm.model_type = kas.model_type;
            showKasModal.value = true;
        };

        const closeKasModal = () => { showKasModal.value = false; };

        const saveKasBook = async () => {
            if (!kasForm.name.trim()) { toast('Nama buku kas wajib diisi!', 'warning'); return; }
            try {
                const res = await fetch(
                    isEditingKas.value ? `/api/kas_books/${editingKasId.value}` : '/api/kas_books',
                    { method: isEditingKas.value ? 'PUT' : 'POST', headers: {'Content-Type':'application/json'}, body: JSON.stringify(kasForm) }
                );
                const json = await res.json();
                if (json.success) { toast(json.message, 'success'); closeKasModal(); fetchKasBooks(); }
                else toast(json.error || 'Gagal menyimpan', 'error');
            } catch (e) { toast('Terjadi kesalahan', 'error'); }
        };

        const deleteKasBook = (id, name) => {
            if (id === 1) { toast('Buku kas utama tidak bisa dihapus', 'error'); return; }
            Swal.fire({ title: 'Hapus Buku Kas?', html: `<b>"${name}"</b>`, icon: 'warning',
                showCancelButton: true, confirmButtonColor: '#ef4444', cancelButtonColor: '#64748b',
                confirmButtonText: 'Ya, Hapus!', cancelButtonText: 'Batal'
            }).then(async r => {
                if (!r.isConfirmed) return;
                const res = await fetch(`/api/kas_books/${id}`, { method: 'DELETE' });
                const json = await res.json();
                if (json.success) { toast('Buku kas dihapus', 'success'); fetchKasBooks(); }
                else toast(json.error, 'error');
            });
        };

        // ─────────── Fetch API ───────────
        const fetchStudents = async () => {
            if (!selectedKas.value) return;
            studentLoading.value = true;
            try {
                const params = new URLSearchParams();
                params.append('kas_id', selectedKas.value.id);
                if (studentSearch.value.trim()) params.append('search', studentSearch.value.trim());
                const res = await fetch(`/api/students?${params}`);
                const json = await res.json();
                if (json.success) students.value = json.data || [];
            } catch (e) { console.error(e); }
            finally { studentLoading.value = false; }
        };

        const fetchTransactions = async () => {
            if (!selectedKas.value) return;
            loading.value = true;
            try {
                const params = new URLSearchParams();
                params.append('kas_id', selectedKas.value.id);
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
            if (!selectedKas.value) return;
            try {
                const params = new URLSearchParams();
                params.append('kas_id', selectedKas.value.id);
                if (filters.startDate) params.append('start_date', filters.startDate);
                if (filters.endDate)   params.append('end_date',   filters.endDate);
                const res = await fetch(`/api/summary?${params}`);
                const json = await res.json();
                if (json.success) summary.value = json.data;
            } catch (e) { console.error(e); }
        };

        const fetchChartData = async () => {
            if (!selectedKas.value) return;
            try {
                const res = await fetch(`/api/chart-data?kas_id=${selectedKas.value.id}`);
                const json = await res.json();
                if (json.success) { chartData.value = json.data; nextTick(() => renderCharts()); }
            } catch (e) { console.error(e); }
        };

        const refreshAll = async () => {
            if (!selectedKas.value) return;
            await Promise.all([fetchTransactions(), fetchStudents(), fetchSummary()]);
        };

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
            const isModel1 = selectedKas.value.model_type === 1;
            const ikromVal = isModel1 ? parseFormatted(form.ikrom_formatted) : 0;
            if (isModel1) {
                if (kasVal + ikromVal <= 0) { toast('Isi nominal Kas atau Ikrom terlebih dahulu', 'warning'); return; }
            } else {
                if (kasVal <= 0) { toast('Isi nominal Kas terlebih dahulu', 'warning'); return; }
            }

            const payload = {
                kas_id: selectedKas.value.id,
                date: form.date, ref_no: form.ref_no, name: form.name.trim(),
                description: form.description.trim(),
                kas_in:    form.type === 'in'  ? kasVal  : 0,
                kas_out:   form.type === 'out' ? kasVal  : 0,
                ikrom_in:  isModel1 && form.type === 'in'  ? ikromVal : 0,
                ikrom_out: isModel1 && form.type === 'out' ? ikromVal : 0,
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

        // ─────────── Student Modal ───────────
        const openCreateStudentModal = () => {
            isEditingStudent.value = false; editingStudentId.value = null;
            Object.assign(studentForm, { name: '', halaqoh: '' });
            showStudentModal.value = true;
        };

        const openEditStudentModal = (s) => {
            isEditingStudent.value = true; editingStudentId.value = s.id;
            Object.assign(studentForm, { name: s.name, halaqoh: s.halaqoh || '' });
            showStudentModal.value = true;
        };

        const closeStudentModal = () => { showStudentModal.value = false; };

        const saveStudent = async () => {
            if (!studentForm.name.trim()) { toast('Nama siswa wajib diisi!', 'warning'); return; }
            const payload = { kas_id: selectedKas.value.id, ...studentForm };
            try {
                const res = await fetch(
                    isEditingStudent.value ? `/api/students/${editingStudentId.value}` : '/api/students',
                    { method: isEditingStudent.value ? 'PUT' : 'POST', headers: {'Content-Type':'application/json'}, body: JSON.stringify(payload) }
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
            params.append('kas_id', selectedKas.value.id);
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
                const res = await fetch('/api/seed', { method: 'POST', headers: {'Content-Type':'application/json'}, body: JSON.stringify({ kas_id: selectedKas.value.id }) });
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
                const res = await fetch('/api/reset', { method: 'POST', headers: {'Content-Type':'application/json'}, body: JSON.stringify({ kas_id: selectedKas.value.id }) });
                const json = await res.json();
                if (json.success) { toast('Data transaksi dibersihkan', 'success'); refreshAll(); }
            });
        };

        const openPrintModal = () => { showPrintModal.value = true; };
        const triggerPrint = () => {
            const origTitle = document.title;
            document.title = '';
            window.print();
            setTimeout(() => { document.title = origTitle; }, 500);
        };

        const toast = (msg, icon = 'info') => {
            Swal.mixin({ toast: true, position: 'top-end', showConfirmButton: false, timer: 2500, timerProgressBar: true })
                .fire({ icon, title: msg });
        };

        // ─────────── Charts ───────────
        const renderCharts = () => {
            if (!chartData.value) return;
            const monthly = chartData.value.monthly || [];
            
            const isModel1 = selectedKas.value.model_type === 1;

            const monthlyCtx = document.getElementById('monthlyChart');
            if (monthlyCtx) {
                if (monthlyChartInstance) monthlyChartInstance.destroy();
                let datasets = [
                    { label: 'Kas Masuk',   data: monthly.map(m => m.kas_in),   backgroundColor: '#10b981' },
                    { label: 'Kas Keluar',  data: monthly.map(m => m.kas_out),  backgroundColor: '#f87171' }
                ];
                if (isModel1) {
                    datasets.push({ label: 'Ikrom Masuk', data: monthly.map(m => m.ikrom_in), backgroundColor: '#3b82f6' });
                    datasets.push({ label: 'Ikrom Keluar',data: monthly.map(m => m.ikrom_out),backgroundColor: '#93c5fd' });
                }

                monthlyChartInstance = new Chart(monthlyCtx, {
                    type: 'bar',
                    data: {
                        labels: monthly.map(m => m.month) || ['Belum ada data'],
                        datasets: datasets
                    },
                    options: { responsive: true, maintainAspectRatio: false,
                        plugins: { legend: { position: 'bottom' }, tooltip: { callbacks: { label: ctx => `${ctx.dataset.label}: ${formatRupiah(ctx.raw)}` } } },
                        scales: { y: { ticks: { callback: v => 'Rp ' + (v >= 1e6 ? (v/1e6).toFixed(1)+'jt' : v.toLocaleString('id-ID')) } } }
                    }
                });
            }

            const posCtx = document.getElementById('posChart');
            if (posCtx && isModel1) {
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
                        datasets: [{ label: 'Total Saldo Akumulasi', data: transactions.value.map(t => isModel1 ? t.total_balance : t.kas_balance),
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
            if (tab === 'charts') nextTick(() => renderCharts());
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

        onMounted(() => fetchKasBooks());

        return {
            kasBooks, selectedKas, transactions, students, summary, chartData,
            activeTab, loading, studentLoading, searchInput, studentSearch,
            showModal, isEditing, showPrintModal, showStudentModal, isEditingStudent,
            showKasModal, isEditingKas, kasForm,
            form, formTotal, studentForm, filters, reportSettings, tableTotals,
            formatRupiah, formatDate,
            onKasInput, onIkromInput, onSelectStudentName,
            setPeriod, resetFilters, refreshAll, fetchTransactions, fetchStudents,
            openCreateModal, openEditModal, closeModal, saveTransaction, deleteTransaction,
            openCreateStudentModal, openEditStudentModal, closeStudentModal, saveStudent, deleteStudent,
            openCreateKasModal, openEditKasModal, closeKasModal, saveKasBook, deleteKasBook,
            selectKas, backToMenu,
            filterByStudent, exportCSV, seedDemoData, resetAllData,
            openPrintModal, triggerPrint,
            // Pagination
            txPerPage, txCurrentPage, paginatedTransactions, totalTxPages, txStartItem, txEndItem,
            studentPerPage, studentCurrentPage, paginatedStudents, totalStudentPages, studentStartItem, studentEndItem,
            kasPerPage, kasCurrentPage, paginatedKasBooks, totalKasPages, kasStartItem, kasEndItem,
            getPageNumbers,
        };
    }
}).mount('#app');
