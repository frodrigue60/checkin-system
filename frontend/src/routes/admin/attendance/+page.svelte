<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { _, locale } from 'svelte-i18n';
	import { apiFetch } from '$lib/api';
	import type { AttendanceExportDTO, WorkCenter, Position, WorkShift } from '$lib/types/models';
	import {
		Clock,
		Calendar,
		Building2,
		User,
		Trash2,
		RefreshCw,
		Filter,
		ArrowLeft,
		Search,
		AlertTriangle,
		LogIn,
		LogOut,
		Timer,
		ChevronDown,
		Loader2,
		ShieldCheck
	} from 'lucide-svelte';
	import { fade, slide, fly } from 'svelte/transition';
	import { flip } from 'svelte/animate';
	import { quintOut } from 'svelte/easing';
	import { page as sveltePage } from '$app/state';

	let attendances = $state<AttendanceExportDTO[]>([]);
	let loading = $state(true);
	let loadingMore = $state(false);
	let errorMsg = $state('');
	let selectedIds = $state(new Set<number>());
	let bulkLoading = $state(false);

	import BatchActionBar from '$lib/components/BatchActionBar.svelte';

	// Filters State
	let searchQuery = $state('');
	let activeTab = $state('all'); // 'all', 'present', 'late', 'absence'
	let selectedCenter = $state('');
	let selectedPosition = $state('');
	let selectedShift = $state('');
	let startDate = $state('');
	let endDate = $state('');
	let selectedShiftType = $state('all');

	// Dropdowns Data
	let centers = $state<WorkCenter[]>([]);
	let positions = $state<Position[]>([]);
	let shifts = $state<WorkShift[]>([]);

	// Pagination State
	let currentPage = $state(1);
	let totalPages = $state(1);
	let hasMore = $derived(currentPage < totalPages);

	// Time for the Asymmetric Clock
	let currentTime = $state(new Date());
	let timerInterval: any;
	let mounted = $state(false);

	onMount(async () => {
		// timerInterval handles clock update
		timerInterval = setInterval(() => {
			currentTime = new Date();
		}, 1000);

		// Fetch filter options
		try {
			const [cRes, pRes, sRes] = await Promise.all([
				apiFetch<WorkCenter[]>('/admin/centers'),
				apiFetch<Position[]>('/admin/positions'),
				apiFetch<WorkShift[]>('/admin/shifts')
			]);
			if (cRes.ok) centers = await cRes.json();
			if (pRes.ok) positions = await pRes.json();
			if (sRes.ok) shifts = await sRes.json();
		} catch (e) {
			console.error('Error fetching filters:', e);
		}
		mounted = true;
	});

	onDestroy(() => {
		if (timerInterval) clearInterval(timerInterval);
	});

	const filteredAttendances = $derived(attendances); // Now server-side filtered

	// Consolidated effect for filters and search
	let debounceTimeout: any;
	$effect(() => {
		// Track all relevant dependencies
		const deps = [
			activeTab,
			selectedCenter,
			selectedPosition,
			selectedShift,
			startDate,
			endDate,
			searchQuery,
			selectedShiftType
		];

		clearTimeout(debounceTimeout);
		debounceTimeout = setTimeout(() => {
			loadAttendances(1, false);
		}, 300);
	});

	async function loadAttendances(page = 1, append = false) {
		if (append) {
			loadingMore = true;
		} else {
			loading = true;
			attendances = [];
		}

		errorMsg = '';
		try {
			const query = new URLSearchParams({
				page: page.toString(),
				limit: '50',
				status: activeTab,
				center_id: selectedCenter,
				position_id: selectedPosition,
				shift_id: selectedShift,
				start: startDate,
				end: endDate,
				search: searchQuery,
				shift_type: selectedShiftType
			});

			const res = await apiFetch<{
				data: AttendanceExportDTO[];
				page: number;
				total_pages: number;
			}>(`/admin/attendances?${query.toString()}`);
			if (res.ok) {
				const result = await res.json();
				if (append) {
					attendances = [...attendances, ...(result.data || [])];
				} else {
					attendances = result.data || [];
				}
				currentPage = result.page;
				totalPages = result.total_pages;
			} else {
				errorMsg = $_('admin.attendance.sync_error');
			}
		} catch (e) {
			errorMsg = $_('admin.attendance.connection_error');
		} finally {
			loading = false;
			loadingMore = false;
		}
	}

	function handleLoadMore() {
		if (hasMore && !loadingMore) {
			loadAttendances(currentPage + 1, true);
		}
	}

	async function exportData(format: 'pdf' | 'csv') {
		const query = new URLSearchParams({
			status: activeTab,
			center_id: selectedCenter,
			position_id: selectedPosition,
			shift_id: selectedShift,
			start: startDate,
			end: endDate,
			search: searchQuery,
			shift_type: selectedShiftType
		});

		try {
			const res = await apiFetch(`/admin/attendances/export/${format}?${query.toString()}`);
			if (res.ok) {
				const blob = await res.blob();
				const downloadUrl = window.URL.createObjectURL(blob);
				const a = document.createElement('a');
				a.href = downloadUrl;
				a.download = `asistencias_${new Date().toISOString().split('T')[0]}.${format}`;
				document.body.appendChild(a);
				a.click();
				a.remove();
			}
		} catch (e) {
			console.error('Export error:', e);
		}
	}

	function formatPreciseTime(date: Date) {
		return date.toLocaleTimeString('en-US', {
			hour12: false,
			hour: '2-digit',
			minute: '2-digit',
			second: '2-digit'
		});
	}

	function formatRecordTime(timeStr: string | null) {
		if (!timeStr || timeStr.includes('0000-01-01') || timeStr.includes('0001-01-01'))
			return '--:--';
		const d = new Date(timeStr);
		return d.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', hour12: true });
	}

	function formatDate(dateStr: string | null) {
		if (!dateStr || dateStr.includes('0000-01-01') || dateStr.includes('0001-01-01')) return '---';
		const d = new Date(dateStr);
		if (isNaN(d.getTime())) return '---';
		return d
			.toLocaleDateString('es-ES', {
				day: '2-digit',
				month: 'short',
				year: d.getFullYear() !== new Date().getFullYear() ? '2-digit' : undefined
			})
			.replace('.', '');
	}

	async function removeRecord(id: number) {
		if (!confirm($_('admin.attendance.confirm_delete'))) return;
		try {
			const res = await apiFetch(`/admin/attendances/${id}`, { method: 'DELETE' });
			if (res.ok) {
				attendances = attendances.filter((a) => a.id !== id);
			}
		} catch (e) {
			alert('Action aborted: Connection failure.');
		}
	}

	function toggleSelect(id: number) {
		if (selectedIds.has(id)) {
			selectedIds.delete(id);
		} else {
			selectedIds.add(id);
		}
		selectedIds = new Set(selectedIds);
	}

	async function handleBulkJustify() {
		if (!confirm(`¿Justificar ${selectedIds.size} registros seleccionados? Esta acción marcará las incidencias como justificadas y recalculará los totales financieros.`)) return;
		bulkLoading = true;
		try {
			const res = await apiFetch('/admin/bulk/attendances/justify', {
				method: 'POST',
				body: JSON.stringify({ 
					ids: Array.from(selectedIds),
					status: 'justified',
					note: 'Justificación masiva desde panel administrativo'
				})
			});
			if (res.ok) {
				selectedIds = new Set();
				loadAttendances(1, false);
			}
		} catch (e) {
			console.error(e);
		} finally {
			bulkLoading = false;
		}
	}
</script>

<div class="min-h-screen pb-24">
	{#if mounted}
		<main class="pt-8 px-6 max-w-7xl mx-auto space-y-12">
		<!-- Editorial Hero Header -->
		<section
			class="flex flex-col md:flex-row md:items-end justify-between gap-8"
			in:fly={{ y: 20, duration: 800, easing: quintOut }}
		>
			<div class="space-y-3">
				<span class="text-primary font-black tracking-[0.3em] uppercase text-[10px]"
					>{$_('admin.attendance.intelligence')}</span
				>
				<h2 class="text-6xl font-black text-primary leading-none tracking-tighter">
					{$_('admin.attendance.title')}.
				</h2>
				<p class="text-slate-500 max-w-sm font-medium text-sm">
					{$_('admin.attendance.description')}
					{new Date().toLocaleDateString($locale || 'es-ES', {
						month: 'long',
						day: 'numeric',
						year: 'numeric'
					})}.
				</p>
			</div>

			<!-- Asymmetric Clock Component -->
			<div
				class="bg-white flex items-center px-8 py-6 rounded-sm border-l-12 border-primary relative overflow-hidden shadow-sm shadow-blue-900/5"
			>
				<div class="absolute -right-6 -top-6 opacity-[0.03] pointer-events-none">
					<Clock size={160} class="text-primary stroke-4" />
				</div>
				<div class="flex flex-col items-end relative z-10">
					<span class="text-primary font-black text-5xl tracking-tight leading-none tabular-nums">
						{formatPreciseTime(currentTime)}
					</span>
					<span class="text-[10px] font-black uppercase tracking-[0.2em] text-slate-400 mt-1"
						>{$_('admin.attendance.precise_time')}</span
					>
				</div>
			</div>
		</section>

		<!-- Filters & High-Density UI -->
		<div class="space-y-8" in:fly={{ y: 20, duration: 800, easing: quintOut, delay: 100 }}>
			<!-- Search & Filters Bar -->
			<div class="space-y-4">
				<div class="grid grid-cols-1 md:grid-cols-12 gap-4">
					<div class="md:col-span-8 relative group">
						<div
							class="absolute inset-y-0 left-5 flex items-center pointer-events-none text-slate-300 group-focus-within:text-primary"
						>
							<Search size={20} strokeWidth={3} />
						</div>
						<input
							type="text"
							placeholder={$_('admin.attendance.search_placeholder')}
							bind:value={searchQuery}
							class="w-full h-16 pl-14 pr-6 bg-white rounded-sm border-none shadow-sm focus:ring-4 focus:ring-primary/5 text-sm font-black tracking-tight placeholder:text-slate-300 uppercase"
						/>
					</div>
					<div class="md:col-span-4 flex gap-2">
						<button
							onclick={() => exportData('pdf')}
							class="flex-1 py-2 px-3 bg-white hover:bg-slate-50 text-primary border-none shadow-sm rounded-sm font-black text-[10px] uppercase tracking-widest flex items-center justify-center gap-2 active:scale-95"
						>
							<Filter size={14} /> PDF
						</button>
						<button
							onclick={() => exportData('csv')}
							class="flex-1 py-2 px-3 bg-white hover:bg-slate-50 text-emerald-600 border-none shadow-sm rounded-sm font-black text-[10px] uppercase tracking-widest flex items-center justify-center gap-2 active:scale-95"
						>
							<Filter size={14} /> CSV
						</button>
					</div>
				</div>

				<!-- Advanced Filters Row -->
				<div class="grid grid-cols-1 md:grid-cols-5 gap-3">
					<div class="bg-white rounded-sm shadow-sm px-4 py-2 flex flex-col">
						<label class="text-[9px] font-black uppercase text-slate-400 mb-1"
							>{$_('common.center')}</label
						>
						<select
							id="centerFilter"
							bind:value={selectedCenter}
							class="bg-transparent border-none text-xs font-bold focus:ring-0 p-0 text-primary uppercase"
						>
							<option value="">{$_('common.all_centers')}</option>
							{#each centers as c}
								<option value={c.id}>{c.name}</option>
							{/each}
						</select>
					</div>

					<div class="bg-white rounded-sm shadow-sm px-4 py-2 flex flex-col">
						<label class="text-[9px] font-black uppercase text-slate-400 mb-1"
							>{$_('common.position')}</label
						>
						<select
							bind:value={selectedPosition}
							class="bg-transparent border-none text-xs font-bold focus:ring-0 p-0 text-primary uppercase"
						>
							<option value="">{$_('common.all_positions')}</option>
							{#each positions as p}
								<option value={p.id}>{p.name}</option>
							{/each}
						</select>
					</div>

					<div class="bg-white rounded-sm shadow-sm px-4 py-2 flex flex-col">
						<label class="text-[9px] font-black uppercase text-slate-400 mb-1"
							>{$_('common.shift')}</label
						>
						<select
							bind:value={selectedShift}
							class="bg-transparent border-none text-xs font-bold focus:ring-0 p-0 text-primary uppercase"
						>
							<option value="">{$_('common.all_shifts')}</option>
							{#each shifts as s}
								<option value={s.id}>{s.name}</option>
							{/each}
						</select>
					</div>

					<div class="bg-white rounded-sm shadow-sm px-4 py-2 flex flex-col">
						<label class="text-[9px] font-black uppercase text-slate-400 mb-1"
							>{$_('common.from')}</label
						>
						<input
							type="date"
							bind:value={startDate}
							class="bg-transparent border-none text-xs font-bold focus:ring-0 p-0 text-primary"
						/>
					</div>

					<div class="bg-white rounded-sm shadow-sm px-4 py-2 flex flex-col">
						<label class="text-[9px] font-black uppercase text-slate-400 mb-1"
							>{$_('common.to')}</label
						>
						<input
							type="date"
							bind:value={endDate}
							class="bg-transparent border-none text-xs font-bold focus:ring-0 p-0 text-primary"
						/>
					</div>

					<div class="bg-white rounded-sm shadow-sm px-4 py-2 flex flex-col">
						<label class="text-[9px] font-black uppercase text-slate-400 mb-1">Tipo</label>
						<select
							bind:value={selectedShiftType}
							class="bg-transparent border-none text-xs font-bold focus:ring-0 p-0 text-primary uppercase"
						>
							<option value="all">Todos</option>
							<option value="fixed">Fijo</option>
							<option value="flexible">Flexible</option>
							<option value="field">Campo</option>
						</select>
					</div>
				</div>
			</div>

			<!-- Navigation Tabs -->
			<nav class="flex items-center gap-3 overflow-x-auto pb-2 no-scrollbar">
				{#each ['all', 'present', 'late', 'absence'] as tab (tab)}
					<button
						class="px-8 py-3 rounded-full text-[10px] font-black uppercase tracking-widest transition-all duration-300 active:scale-95
                   {activeTab === tab
							? 'bg-primary text-white shadow-lg shadow-primary/20'
							: 'bg-slate-100 text-slate-400 hover:bg-slate-200 hover:text-slate-600'}"
						onclick={() => (activeTab = tab)}
					>
						{tab === 'all'
							? $_('common.all')
							: tab === 'present'
								? $_('admin.attendance.tab_present')
								: tab === 'late'
									? $_('admin.attendance.tab_late')
									: $_('admin.attendance.tab_absence')}
					</button>
				{/each}
			</nav>

			<div
				class="space-y-3"
				in:fly={{ y: 20, duration: 800, easing: quintOut, delay: 200 }}
			>
				{#if loading && attendances.length === 0}
					<!-- 
					<div class="py-24 flex flex-col items-center justify-center space-y-4 bg-white rounded-sm border border-slate-100 shadow-sm" in:fade>
						<Loader2 class="w-10 h-10 text-primary animate-spin" />
						<p class="text-[10px] font-black text-slate-400 uppercase tracking-[0.2em]">{$_('common.loading') || 'Cargando registros...'}</p>
					</div>
					-->
					{#each Array(5) as _}
						<div
							class="h-24 bg-white rounded-sm animate-pulse shadow-sm border border-slate-50"
						></div>
					{/each}
				{:else if filteredAttendances.length === 0}
					<div class="py-20 text-center space-y-2 bg-white rounded-sm shadow-sm italic">
						<p class="text-slate-400 font-bold uppercase tracking-widest text-xs">
							{$_('admin.attendance.no_anomalies')}
						</p>
						<p class="text-[10px] text-slate-300 font-medium">
							{$_('admin.attendance.no_records_hint')}
						</p>
					</div>
				{:else}
					{#each filteredAttendances as record (record.id)}
						<div animate:flip={{ duration: 400 }}>
							<div
								class="group flex items-center justify-between p-5 bg-white rounded-sm border border-transparent shadow-sm hover:shadow-2xl hover:shadow-primary/5 transition-all duration-300 relative {selectedIds.has(record.id) ? 'ring-2 ring-primary bg-primary/5' : ''}"
							>
								<!-- Bulk Checkbox -->
								<div class="absolute -top-2 -left-2 z-20 transition-transform hover:scale-110">
									<input 
										type="checkbox" 
										checked={selectedIds.has(record.id)}
										onchange={() => toggleSelect(record.id)}
										class="h-6 w-6 rounded-lg border-2 border-slate-200 text-primary focus:ring-primary cursor-pointer shadow-lg bg-white checked:bg-primary"
									/>
								</div>

								<div class="flex items-center gap-5">
									<!-- Initials Placeholder -->
									<div
										class="w-14 h-14 rounded-sm bg-slate-50 flex items-center justify-center text-primary font-black text-xl shadow-inner group-hover:bg-primary group-hover:text-white"
									>
										{record.employee_name
											.split(' ')
											.map((n) => n[0])
											.join('')}
									</div>

									<div>
										<h4
											class="font-black text-primary text-lg tracking-tight"
										>
											<a href="/admin/attendance/{record.id}" class="hover:text-primary-600 hover:underline">
												{record.employee_name}
											</a>
										</h4>
										<div class="flex items-center gap-2 mt-0.5">
											<span
												class="text-[10px] text-primary font-black uppercase tracking-wider bg-primary/5 px-1.5 py-0.5 rounded"
												>{formatDate(record.check_in || record.created_at)}</span
											>
											<span class="w-1 h-1 rounded-full bg-slate-200"></span>
											<span class="text-[10px] text-slate-400 font-black uppercase tracking-wider"
												>{record.position_name || $_('common.personnel')}</span
											>
											<span class="w-1 h-1 rounded-full bg-slate-200"></span>
											<span class="text-[10px] text-slate-400 font-black uppercase tracking-wider"
												>{record.work_center_name}</span
											>
										</div>
									</div>
								</div>

								<div class="flex items-center gap-8">
									<div class="flex flex-col items-end">
										<span
											class="text-[10px] font-black uppercase tracking-widest text-slate-300 mb-1 leading-none"
											>{$_('common.check_in')}</span
										>
										<span class="font-black text-lg text-primary tabular-nums tracking-tighter"
											>{formatRecordTime(record.check_in)}</span
										>
									</div>

									<div class="flex flex-col items-end min-w-[100px]">
										{#if record.is_absence}
											<span
												class="bg-orange-50 text-orange-600 px-3 py-1.5 rounded-lg text-[10px] font-black uppercase tracking-widest border border-orange-100"
											>
												{$_('admin.attendance.absence')}
											</span>
										{:else if record.is_late}
											<span
												class="bg-rose-50 text-rose-600 px-3 py-1.5 rounded-lg text-[10px] font-black uppercase tracking-widest border border-rose-100 animate-pulse"
											>
												{$_('admin.attendance.late_arrival')}
											</span>
										{:else if !record.check_out}
											<span
												class="bg-emerald-50 text-emerald-600 px-3 py-1.5 rounded-lg text-[10px] font-black uppercase tracking-widest border border-emerald-100"
											>
												{$_('admin.attendance.on_duty')}
											</span>
										{:else}
											<span
												class="bg-slate-50 text-slate-400 px-3 py-1.5 rounded-lg text-[10px] font-black uppercase tracking-widest border border-slate-100"
											>
												{$_('admin.attendance.closed')}
											</span>
										{/if}
									</div>

									<button
										class="p-2 text-slate-200 hover:text-rose-500"
										onclick={(e) => {
											e.preventDefault();
											e.stopPropagation();
											removeRecord(record.id);
										}}
									>
										<Trash2 size={18} />
									</button>
								</div>
							</div>
						</div>
					{/each}
				{/if}
			</div>

			<!-- Batch Actions Hub -->
			<BatchActionBar 
				selectedCount={selectedIds.size} 
				onClear={() => selectedIds = new Set()}
			>
				<div class="flex items-center gap-2">
					<button 
						onclick={handleBulkJustify}
						disabled={bulkLoading}
						class="bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl h-11 px-6 flex items-center gap-2 shadow-lg shadow-indigo-900/10 font-black text-[10px] uppercase tracking-widest transition-all active:scale-95 disabled:opacity-50"
					>
						<ShieldCheck size={16} /> 
						<span class="hidden md:inline">{$_('admin.attendance.bulk_justify')}</span>
					</button>
				</div>
			</BatchActionBar>

			<!-- Pagination / Load More -->
			<div class="pt-8 flex justify-center">
				{#if hasMore}
					<button
						class="flex flex-col items-center gap-2 group outline-none disabled:opacity-50"
						onclick={handleLoadMore}
						disabled={loadingMore}
					>
						<div
							class="w-12 h-12 rounded-full border-2 border-slate-100 flex items-center justify-center text-slate-300 group-hover:border-primary group-hover:text-primary group-hover:bg-primary/5 transition-all active:scale-95"
						>
							{#if loadingMore}
								<Loader2 size={24} class="animate-spin" />
							{:else}
								<ChevronDown size={24} />
							{/if}
						</div>
						<span
							class="text-[9px] font-black tracking-[0.3em] uppercase text-slate-300 group-hover:text-primary"
							>{loadingMore ? $_('common.loading') : $_('admin.attendance.archive_access')}</span
						>
					</button>
				{:else if attendances.length > 0}
					<p class="text-[9px] font-black tracking-[0.3em] uppercase text-slate-300">
						{$_('admin.attendance.end_of_archive')}
					</p>
				{/if}
			</div>
		</div>
		</main>
	{/if}
</div>
