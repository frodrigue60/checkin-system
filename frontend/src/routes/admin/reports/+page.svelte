<script lang="ts">
	import { onMount } from 'svelte';
	import { fly } from 'svelte/transition';
	import { quintOut } from 'svelte/easing';
	import { _ } from 'svelte-i18n';
	import { apiFetch } from '$lib/api';
	import { authState } from '$lib/auth.svelte';
	import * as Table from '$lib/components/ui/table/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import * as Select from '$lib/components/ui/select/index.js';
	import {
		Loader2,
		Globe,
		User,
		Rocket,
		Info,
		FileText,
		Trash2,
		ExternalLink,
		Calendar,
		FileDown,
		ArrowLeft,
		Search,
		Users,
		PieChart,
		ChevronRight,
		PlayCircle,
		Download,
		Eye,
		LandPlot,
		Clock,
		Briefcase,
		MapPin
	} from 'lucide-svelte';

	let reportRanges = $state<any[]>([]);
	let employees = $state<Employee[]>([]);
	let shifts = $state<WorkShift[]>([]);
	let positions = $state<Position[]>([]);
	let workCenters = $state<WorkCenter[]>([]);
	let loading = $state(true);
	let generating = $state(false);
	let mounted = $state(false);
	let activeJobs = $state<ReportJobDTO[]>([]);

	let generationMode = $state('global'); // 'global' | 'individual'
	let selectedEmployeeId = $state('');
	let selectedShiftId = $state('');
	let selectedPositionId = $state('');
	let selectedCenterId = $state('');

	let filter = $state({
		startDate: new Date(new Date().setDate(new Date().getDate() - 30)).toISOString().split('T')[0],
		endDate: new Date().toISOString().split('T')[0]
	});

	let searchQuery = $state('');

	const canGenerate = $derived(authState.isAdmin);

	const filteredRanges = $derived(
		reportRanges.filter((range) => {
			const s = range.start_date.split('T')[0];
			const e = range.end_date.split('T')[0];
			return s.includes(searchQuery) || e.includes(searchQuery);
		})
	);

	async function loadReportRanges() {
		loading = true;
		try {
			const res = await apiFetch('/admin/reports');
			if (res.ok) {
				reportRanges = await res.json();
			}
		} catch (e) {
			console.error('Error loading report ranges:', e);
		} finally {
			loading = false;
		}
	}

	async function loadActiveJobs() {
		try {
			const res = await apiFetch<ReportJobDTO[]>('/admin/reports/jobs');
			if (res.ok) {
				const previousJobs = activeJobs;
				activeJobs = await res.json();
				
				// Detect completion to refresh the main list
				const hasNewCompletion = activeJobs.some(job => 
					job.status === 'completed' && 
					!previousJobs.find(pj => pj.id === job.id && pj.status === 'completed')
				);

				if (hasNewCompletion) {
					await loadReportRanges();
				}

				// If there are processing jobs, poll
				if (activeJobs.some((j) => j.status === 'processing' || j.status === 'pending' || j.status === 'generating_files')) {
					setTimeout(loadActiveJobs, 2000);
				}
			}
		} catch (e) {
			console.error('Error loading active jobs:', e);
		}
	}

	async function generateReport() {
		if (!canGenerate) return;

		if (generationMode === 'individual' && !selectedEmployeeId) {
			alert($_('admin.reports.select_emp_error'));
			return;
		}

		generating = true;
		try {
			const body: any = {
				start_date: filter.startDate,
				end_date: filter.endDate
			};

			if (generationMode === 'individual') {
				body.employee_id = parseInt(selectedEmployeeId);
			} else {
				if (selectedShiftId) body.work_shift_id = parseInt(selectedShiftId);
				if (selectedPositionId) body.position_id = parseInt(selectedPositionId);
				if (selectedCenterId) body.work_center_id = parseInt(selectedCenterId);
			}

			const res = await apiFetch('/admin/reports/generate', {
				method: 'POST',
				body: JSON.stringify(body)
			});

			if (res.ok) {
				// Since it's async, we just start polling jobs
				await loadActiveJobs();
			} else {
				const err = await res.json();
				alert(err.error || $_('admin.reports.gen_error'));
			}
		} catch (e) {
			console.error('Error generating report:', e);
			alert($_('admin.reports.gen_error'));
		} finally {
			generating = false;
		}
	}

	async function downloadReport(start: string, end: string) {
		const res = await apiFetch(`/admin/reports/export?start_date=${start}&end_date=${end}`);
		if (res.ok) {
			const blob = await res.blob();
			const url = window.URL.createObjectURL(blob);
			const a = document.createElement('a');
			a.href = url;
			a.download = `reporte_${start}_${end}.pdf`;
			document.body.appendChild(a);
			a.click();
			window.URL.revokeObjectURL(url);
			document.body.removeChild(a);
		} else {
			const err = await res.json();
			alert(err.error || $_('common.error_saving'));
		}
	}

	async function deleteRange(range: any) {
		const sDate = (range.start_date || '').split('T')[0];
		const eDate = (range.end_date || '').split('T')[0];

		if (
			!confirm($_('admin.reports.confirm_delete_cycle', { values: { start: sDate, end: eDate } }))
		)
			return;

		const res = await apiFetch(`/admin/reports?start_date=${sDate}&end_date=${eDate}`, {
			method: 'DELETE'
		});
		if (res.ok) {
			await loadReportRanges();
		} else {
			const err = await res.json();
			alert(err.error || $_('common.error_saving'));
		}
	}

	onMount(async () => {
		await loadReportRanges();
		await loadActiveJobs();

		// Fetch metadata for filters
		const [empRes, shiftRes, posRes, centerRes] = await Promise.all([
			apiFetch('/admin/employees'),
			apiFetch('/admin/shifts'),
			apiFetch('/admin/positions'),
			apiFetch('/admin/centers')
		]);

		if (empRes.ok) {
			const allEmps = await empRes.json();
			employees = allEmps.filter((e: any) => e.is_active);
		}
		if (shiftRes.ok) shifts = await shiftRes.json();
		if (posRes.ok) positions = await posRes.json();
		if (centerRes.ok) workCenters = await centerRes.json();
		mounted = true;
	});
</script>

<div class="min-h-screen pb-24">
	{#if mounted}
		<main class="pt-12 px-8 max-w-7xl mx-auto space-y-12">
		<!-- Header Section -->
		<section
			class="flex flex-col md:flex-row md:items-end justify-between gap-10"
			in:fly={{ y: 20, duration: 800, easing: quintOut }}
		>
			<div class="space-y-4">
				<div class="space-y-1">
					<h2 class="text-7xl font-black text-slate-900 leading-none tracking-tighter">
						{$_('admin.reports.title')}
					</h2>
				</div>
			</div>
		</section>

		<!-- Main Content Grid -->
		<div
			class="grid grid-cols-1 lg:grid-cols-12 gap-10"
			in:fly={{ y: 20, duration: 800, easing: quintOut, delay: 100 }}
		>
			<!-- Generation Form Column -->
			<div
				class="lg:col-span-7 bg-white rounded-sm p-8 shadow-2xl shadow-slate-200/50 border border-slate-50 space-y-10"
			>
				<!-- <div class="flex items-center gap-3 border-b border-slate-50 pb-6">
					<div class="h-10 w-10 rounded-sm bg-primary text-white flex items-center justify-center">
						<PlayCircle class="h-6 w-6" />
					</div>
					<h3 class="text-xl font-black tracking-tight text-slate-900 uppercase">
						Processing Terminal
					</h3>
				</div> -->

				<div class="space-y-8">
					<!-- Segment Control (Mockup Inspired) -->
					<div class="space-y-3">
						<Label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1"
							>{$_('admin.reports.scope')}</Label
						>
						<div class="flex p-1.5 bg-slate-50 rounded-sm w-full sm:w-fit">
							<button
								onclick={() => (generationMode = 'global')}
								class="px-8 py-3 rounded-sm text-[10px] font-black uppercase tracking-widest transition-all {generationMode ===
								'global'
									? 'bg-white text-primary shadow-lg shadow-slate-200'
									: 'text-slate-400 hover:text-slate-600'}"
							>
								{$_('common.global')}
							</button>
							<button
								onclick={() => (generationMode = 'individual')}
								class="px-8 py-3 rounded-sm text-[10px] font-black uppercase tracking-widest transition-all {generationMode ===
								'individual'
									? 'bg-white text-primary shadow-lg shadow-slate-200'
									: 'text-slate-400 hover:text-slate-600'}"
							>
								{$_('common.individual')}
							</button>
						</div>
					</div>

					{#if generationMode === 'individual'}
						<div class="space-y-3">
							<Label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1"
								>{$_('admin.reports.assigned_personnel')}</Label
							>
							<div class="relative group">
								<div
									class="absolute inset-y-0 left-5 flex items-center pointer-events-none text-slate-300 group-focus-within:text-primary transition-colors"
								>
									<User class="h-5 w-5" />
								</div>
								<select
									bind:value={selectedEmployeeId}
									class="h-16 w-full rounded-sm bg-slate-50 border-none pl-14 pr-6 font-black text-sm text-slate-900 focus:ring-4 focus:ring-primary/5 appearance-none outline-none transition-all"
								>
									<option value="">{$_('admin.reports.select_colleague')}</option>
									{#each employees as emp}
										<option value={emp.id}>{emp.user_name}</option>
									{/each}
								</select>
								<div
									class="absolute right-5 top-1/2 -translate-y-1/2 pointer-events-none text-slate-300"
								>
									<ChevronRight rotation={90} class="h-4 w-4 rotate-90" />
								</div>
							</div>
						</div>
					{:else}
						<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
							<!-- Work Center Filter -->
							<div class="space-y-3">
								<Label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">
									{$_('admin.reports.filter_center')}
								</Label>
								<div class="relative group">
									<div
										class="absolute inset-y-0 left-4 flex items-center pointer-events-none text-slate-300 group-focus-within:text-primary"
									>
										<MapPin class="h-4 w-4" />
									</div>
									<select
										bind:value={selectedCenterId}
										class="h-12 w-full rounded-sm bg-slate-50 border-none pl-10 pr-4 font-bold text-xs text-slate-900 focus:ring-2 focus:ring-primary/5 appearance-none outline-none transition-all"
									>
										<option value="">{$_('common.all_centers')}</option>
										{#each workCenters as center}
											<option value={center.id}>{center.name}</option>
										{/each}
									</select>
								</div>
							</div>

							<!-- Position Filter -->
							<div class="space-y-3">
								<Label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">
									{$_('admin.reports.filter_position')}
								</Label>
								<div class="relative group">
									<div
										class="absolute inset-y-0 left-4 flex items-center pointer-events-none text-slate-300 group-focus-within:text-primary"
									>
										<Briefcase class="h-4 w-4" />
									</div>
									<select
										bind:value={selectedPositionId}
										class="h-12 w-full rounded-sm bg-slate-50 border-none pl-10 pr-4 font-bold text-xs text-slate-900 focus:ring-2 focus:ring-primary/5 appearance-none outline-none transition-all"
									>
										<option value="">{$_('common.all_positions')}</option>
										{#each positions as pos}
											<option value={pos.id}>{pos.name}</option>
										{/each}
									</select>
								</div>
							</div>

							<!-- Shift Filter -->
							<div class="space-y-3">
								<Label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">
									{$_('admin.reports.filter_shift')}
								</Label>
								<div class="relative group">
									<div
										class="absolute inset-y-0 left-4 flex items-center pointer-events-none text-slate-300 group-focus-within:text-primary"
									>
										<Clock class="h-4 w-4" />
									</div>
									<select
										bind:value={selectedShiftId}
										class="h-12 w-full rounded-sm bg-slate-50 border-none pl-10 pr-4 font-bold text-xs text-slate-900 focus:ring-2 focus:ring-primary/5 appearance-none outline-none transition-all"
									>
										<option value="">{$_('common.all_shifts')}</option>
										{#each shifts as shift}
											<option value={shift.id}>{shift.name}</option>
										{/each}
									</select>
								</div>
							</div>
						</div>
					{/if}

					<!-- Date Matrix -->
					<div class="space-y-3">
						<Label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1"
							>{$_('admin.reports.analysis_period')}</Label
						>
						<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
							<div class="relative group">
								<div
									class="absolute left-5 top-1/2 -translate-y-1/2 text-slate-300 pointer-events-none group-focus-within:text-primary transition-colors"
								>
									<Calendar class="h-5 w-5" />
								</div>
								<input
									type="date"
									bind:value={filter.startDate}
									class="h-16 w-full bg-slate-50 border-none rounded-sm font-black text-sm pl-14 pr-6 focus:ring-4 focus:ring-primary/5 transition-all text-slate-900"
								/>
							</div>
							<div class="relative group">
								<div
									class="absolute left-5 top-1/2 -translate-y-1/2 text-slate-300 pointer-events-none group-focus-within:text-primary transition-colors"
								>
									<Calendar class="h-5 w-5" />
								</div>
								<input
									type="date"
									bind:value={filter.endDate}
									class="h-16 w-full bg-slate-50 border-none rounded-sm font-black text-sm pl-14 pr-6 focus:ring-4 focus:ring-primary/5 transition-all text-slate-900"
								/>
							</div>
						</div>
					</div>

					<button
						class="h-20 w-full rounded-sm bg-primary text-white font-black text-lg gap-4 shadow-2xl shadow-primary/30 hover:scale-[1.02] active:scale-98 transition-all flex items-center justify-center disabled:opacity-50 disabled:scale-100"
						onclick={generateReport}
						disabled={generating}
					>
						{#if generating}
							<Loader2 class="h-6 w-6 animate-spin" />
							<span class="uppercase tracking-widest">{$_('common.processing')}</span>
						{:else}
							<Rocket class="h-6 w-6" />
							<span class="uppercase tracking-widest">{$_('admin.reports.generate_button')}</span>
						{/if}
					</button>
				</div>

				<div class="p-5 rounded-sm bg-slate-50 border border-slate-100 border-dashed flex gap-4">
					<div
						class="h-8 w-8 rounded-lg bg-white shadow-sm flex items-center justify-center text-primary shrink-0"
					>
						<Info class="h-4 w-4" />
					</div>
					<p
						class="text-[10px] font-black text-slate-400 uppercase tracking-[0.15em] leading-relaxed"
					>
						{$_('admin.reports.notice')}
					</p>
				</div>
				{#if activeJobs.length > 0}
					<div class="space-y-4">
						<h3 class="text-xs font-black uppercase tracking-[0.2em] text-slate-400 ml-1">
							{$_('admin.reports.active_tasks')}
						</h3>
						<div class="space-y-3">
							{#each activeJobs.slice(0, 5) as job (job.id)}
								<div class="p-4 bg-white border rounded-sm shadow-sm flex items-center gap-4 group/job relative overflow-hidden">
									{#if job.status === 'processing' || job.status === 'pending' || job.status === 'generating_files'}
										<div class="absolute left-0 top-0 bottom-0 w-1 bg-primary animate-pulse"></div>
									{:else if job.status === 'completed'}
										<div class="absolute left-0 top-0 bottom-0 w-1 bg-emerald-500"></div>
									{:else}
										<div class="absolute left-0 top-0 bottom-0 w-1 bg-rose-500"></div>
									{/if}
									
									<div class="h-10 w-10 rounded-full flex items-center justify-center shrink-0 bg-slate-50 text-slate-400">
										{#if job.status === 'processing' || job.status === 'generating_files'}
											<Loader2 class="h-5 w-5 animate-spin text-primary" />
										{:else if job.status === 'completed'}
											<Rocket class="h-5 w-5 text-emerald-500" />
										{:else if job.status === 'failed'}
											<Info class="h-5 w-5 text-rose-500" />
										{:else}
											<Clock class="h-5 w-5" />
										{/if}
									</div>
									
									<div class="flex-1 min-w-0">
										<div class="flex justify-between mb-1 items-center">
											<span class="text-[10px] font-black uppercase tracking-tight text-slate-700">
												{job.start_date.split('T')[0]} — {job.end_date.split('T')[0]}
											</span>
											{#if job.status === 'completed'}
												<div class="flex gap-2">
													{#if job.pdf_url}
														<a href={job.pdf_url} target="_blank" class="text-primary hover:scale-110 transition-transform">
															<FileText class="h-4 w-4" />
														</a>
													{/if}
													{#if job.excel_url}
														<a href={job.excel_url} target="_blank" class="text-emerald-600 hover:scale-110 transition-transform">
															<FileDown class="h-4 w-4" />
														</a>
													{/if}
												</div>
											{:else}
												<span class="text-[10px] font-black text-primary">{job.progress}%</span>
											{/if}
										</div>
										
										{#if job.status !== 'completed' && job.status !== 'failed'}
											<div class="w-full h-1.5 bg-slate-50 rounded-full overflow-hidden">
												<div 
													class="h-full bg-primary transition-all duration-700 ease-out" 
													style="width: {job.progress}%"
												></div>
											</div>
										{:else if job.status === 'failed'}
											<p class="text-[9px] text-rose-500 font-bold truncate">Error en generación</p>
										{:else}
											<p class="text-[9px] text-emerald-600 font-bold uppercase tracking-widest">{$_('common.completed')}</p>
										{/if}
									</div>
								</div>
							{/each}
						</div>
					</div>
				{/if}
			</div>

			<!-- History Column (Mockup Inspired) -->
			<div class="lg:col-span-5 space-y-6">
				<div class="flex items-center justify-between px-1">
					<div class="flex flex-col">
						<h3 class="text-2xl font-black tracking-tight text-slate-900 uppercase">
							{$_('admin.reports.recent_reports')}
						</h3>
						<span class="text-[10px] font-black text-slate-400 uppercase tracking-[0.2em] mt-1"
							>{$_('admin.reports.audit_archive')}</span
						>
					</div>
					<div class="relative w-32 group">
						<Search
							class="absolute left-3 top-1/2 -translate-y-1/2 h-3 w-3 text-slate-300 group-focus-within:text-primary transition-colors"
						/>
						<input
							bind:value={searchQuery}
							placeholder={$_('common.search_placeholder')}
							class="h-9 w-full rounded-lg bg-white border border-slate-100 pl-8 pr-2 text-[10px] font-black placeholder:text-slate-300 focus:ring-2 focus:ring-primary/5 transition-all outline-none"
						/>
					</div>
				</div>

				<div class="space-y-4 max-h-[700px] overflow-y-auto pr-2 no-scrollbar">
					{#if loading}
						{#each Array(3) as _}
							<div
								class="h-32 bg-white rounded-sm animate-pulse shadow-sm border border-slate-50"
							></div>
						{/each}
					{:else if filteredRanges.length === 0}
						<div
							class="h-[300px] flex flex-col items-center justify-center text-center space-y-4 bg-white rounded-sm border border-dashed border-slate-200"
						>
							<FileText class="h-10 w-10 text-slate-100" />
							<p class="text-[10px] font-black text-slate-300 uppercase tracking-widest">
								{$_('admin.reports.no_reports')}
							</p>
						</div>
					{:else}
						{#each filteredRanges as range}
							<div
								class="bg-white rounded-sm p-5 border border-transparent hover:border-primary/10 transition-all hover:shadow-xl hover:shadow-slate-200/50 group"
							>
								<div class="flex justify-between items-start mb-4">
									<div class="flex items-center gap-4">
										<div
											class="h-10 w-10 rounded-sm bg-slate-50 flex items-center justify-center text-slate-400 group-hover:bg-primary/5 group-hover:text-primary transition-all"
										>
											<Calendar class="h-5 w-5" />
										</div>
										<div>
											<h4 class="font-black text-slate-900 text-sm tracking-tighter leading-tight">
												{range.start_date.split('T')[0]} — {range.end_date.split('T')[0]}
											</h4>
											<span
												class="text-[9px] font-black text-slate-400 uppercase tracking-widest mt-1 block"
												>{$_('admin.reports.payroll_summary')}</span
											>
										</div>
									</div>
									<div class="flex flex-col items-end gap-1">
										<Badge
											variant="outline"
											class="h-7 px-3 border-emerald-100 bg-emerald-50 text-emerald-600 font-black text-[9px] uppercase tracking-widest rounded-lg"
										>
											{range.employee_count}
											{$_('common.personnel')}
										</Badge>
										{#if range.is_stale}
											<Badge
												variant="outline"
												class="h-5 px-2 border-amber-100 bg-amber-50 text-amber-600 font-black text-[8px] uppercase tracking-widest rounded-lg flex items-center gap-1"
											>
												<PlayCircle class="h-2 w-2 animate-pulse" />
												{$_('admin.reports.stale_warning')}
											</Badge>
										{/if}
									</div>
								</div>

								<div class="flex items-center justify-between pt-4 border-t border-slate-50">
									<div class="flex items-center gap-2">
										<a
											href={`/admin/reports/${range.start_date.split('T')[0]}/${range.end_date.split('T')[0]}`}
											class="flex items-center gap-1.5 px-3 py-2 rounded-lg text-[9px] font-black uppercase tracking-widest text-slate-400 hover:text-primary hover:bg-primary/5 transition-all"
										>
											<Eye size={12} />
											{$_('admin.reports.breakdown')}
										</a>
										<button
											onclick={() => downloadReport(range.start_date.split('T')[0], range.end_date.split('T')[0])}
											class="flex items-center gap-1.5 px-3 py-2 rounded-lg text-[9px] font-black uppercase tracking-widest text-slate-400 hover:text-emerald-600 hover:bg-emerald-50 transition-all"
										>
											<Download size={12} />
											{$_('common.download')}
										</button>
									</div>
									<button
										class="h-8 w-8 rounded-lg flex items-center justify-center text-slate-200 hover:text-rose-500 hover:bg-rose-50 transition-all"
										onclick={() => deleteRange(range)}
									>
										<Trash2 size={16} />
									</button>
								</div>
							</div>
						{/each}
					{/if}
				</div>
			</div>
		</div>
	</main>
	{/if}
</div>

<style>
	:global(.no-scrollbar::-webkit-scrollbar) {
		display: none;
	}
	:global(.no-scrollbar) {
		-ms-overflow-style: none;
		scrollbar-width: none;
	}
</style>
