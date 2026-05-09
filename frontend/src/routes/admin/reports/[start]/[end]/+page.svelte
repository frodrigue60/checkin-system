<script lang="ts">
	import { onMount } from 'svelte';
	import { _, locale } from 'svelte-i18n';
	import { page } from '$app/stores';
	import { apiFetch } from '$lib/api';
	import { fade, fly } from 'svelte/transition';
	import { quintOut } from 'svelte/easing';
	import { get } from 'svelte/store';
	import {
		Search,
		FileText,
		Settings2,
		Download,
		ChevronRight,
		ArrowLeft,
		Trash2
	} from 'lucide-svelte';
	import { env } from '$env/dynamic/public';

	let currentDetails = $state([]);
	let loading = $state(true);
	let searchQuery = $state('');
	let pdfOrientation = $state('vertical');

	const startDate = $page.params.start;
	const endDate = $page.params.end;

	let totalNetDistribution = $derived(
		currentDetails.reduce((sum, report) => sum + ((report.total_earnings || 0) - (report.total_deductions || 0)), 0)
	);

	const filteredDetails = $derived(
		currentDetails.filter((report) => {
			const nameMatch = report.employee_name?.toLowerCase().includes(searchQuery.toLowerCase());
			return nameMatch;
		})
	);

	async function loadDetails() {
		loading = true;
		try {
			const res = await apiFetch(
				`/admin/reports/details?start_date=${startDate}&end_date=${endDate}`
			);
			if (res.ok) {
				currentDetails = await res.json();
			} else {
				const err = await res.json();
				alert(err.error || $_('admin.reports.load_details_error'));
			}
		} catch (e) {
			console.error(e);
			alert('Error fetching report details.');
		} finally {
			loading = false;
		}
	}

	async function viewIndividualPDF(report: any) {
		const res = await apiFetch(`/admin/reports/${report.id}/export?orientation=${pdfOrientation}&lang=${get(locale)}`);
		if (res.ok) {
			const blob = await res.blob();
			const url = window.URL.createObjectURL(blob);
			window.open(url, '_blank');
		}
	}

	async function viewBatchPDF() {
		const res = await apiFetch(
			`/admin/reports/export?start_date=${startDate}&end_date=${endDate}&orientation=${pdfOrientation}&lang=${get(locale)}`
		);
		if (res.ok) {
			const blob = await res.blob();
			const url = window.URL.createObjectURL(blob);
			window.open(url, '_blank');
		} else {
			alert($_('common.error_saving'));
		}
	}

	async function deleteReport(report: any) {
		if (!confirm($_('admin.reports.confirm_delete_report', { values: { name: report.employee_name } })))
			return;

		const res = await apiFetch(`/admin/reports/${report.id}`, { method: 'DELETE' });
		if (res.ok) {
			currentDetails = currentDetails.filter((r: any) => r.id !== report.id);
		} else {
			const err = await res.json();
			alert(err.error || $_('common.error_saving'));
		}
	}

	onMount(loadDetails);
</script>

<div class="min-h-screen text-foreground font-body">
	<main class="pt-8 pb-32 px-4 md:px-8 max-w-5xl mx-auto space-y-8">
		<a
			in:fly={{ y: 20, duration: 800, easing: quintOut }}
			href="/admin/reports"
			class="flex items-center gap-2 text-slate-400 hover:text-primary transition-colors text-[10px] font-black uppercase tracking-widest w-fit"
		>
			<ArrowLeft size={14} />
			{$_('admin.reports.back_to_registry')}
		</a>

		<!-- Main Panel -->
		<div
			class="bg-card rounded-sm overflow-hidden shadow-[0px_12px_32px_rgba(25,28,29,0.06)] ring-1 ring-border/50"
			in:fly={{ y: 20, duration: 800, easing: quintOut, delay: 100 }}
		>
			<!-- Header -->
			<div
				class="px-6 py-8 bg-card flex flex-col md:flex-row md:items-end justify-between gap-4 border-b border-border/40"
			>
				<div>
					<span
						class="font-sans text-[10px] font-black tracking-widest text-primary uppercase mb-2 block animate-pulse"
					>
						{env.PUBLIC_APP_NAME} — {$_('admin.reports.payroll_preview')}
					</span>
					<h2 class="font-sans text-4xl font-extrabold text-primary tracking-tighter">
						{$_('admin.reports.cycle')}: {startDate} — {endDate}
					</h2>
				</div>
				<div class="flex flex-col items-end">
					<span class="font-sans text-[10px] text-muted-foreground uppercase tracking-widest">
						{$_('admin.reports.total_net_distribution')}
					</span>
					<span class="font-sans text-2xl font-bold text-primary tabular-nums">
						${totalNetDistribution.toLocaleString('es-MX', { minimumFractionDigits: 2 })}
					</span>
				</div>
			</div>

			<!-- List Search & Filter -->
			<div
				class="px-6 py-4 flex items-center justify-between bg-subtle-container border-b border-border/40"
			>
				<div class="flex items-center gap-3 text-muted-foreground w-full md:w-auto">
					<Search size={16} />
					<input
						type="text"
						bind:value={searchQuery}
						class="bg-transparent border-none focus:ring-0 font-sans font-bold text-xs placeholder:text-muted-foreground/60 w-full md:w-48 outline-none uppercase tracking-wider"
						placeholder={$_('admin.reports.filter_employee_placeholder')}
					/>
				</div>
				<div class="hidden md:flex items-center gap-4">
					<div class="flex items-center bg-card rounded-lg border border-border/50 p-1 shadow-sm">
						<button
							onclick={() => (pdfOrientation = 'vertical')}
							class="px-3 py-1 rounded text-[10px] font-black uppercase tracking-widest transition-all {pdfOrientation ===
							'vertical'
								? 'bg-primary text-primary-foreground shadow-sm'
								: 'text-muted-foreground hover:text-foreground'}"
						>
							{$_('common.vertical')}
						</button>
						<button
							onclick={() => (pdfOrientation = 'horizontal')}
							class="px-3 py-1 rounded text-[10px] font-black uppercase tracking-widest transition-all {pdfOrientation ===
							'horizontal'
								? 'bg-primary text-primary-foreground shadow-sm'
								: 'text-muted-foreground hover:text-foreground'}"
						>
							{$_('common.horizontal')}
						</button>
					</div>
					<button
						class="flex items-center gap-2 hover:bg-card px-3 py-1.5 rounded-lg transition-colors shadow-sm border border-transparent hover:border-border/50 group"
						onclick={viewBatchPDF}
					>
						<FileText size={16} class="text-muted-foreground group-hover:text-primary" />
						<span
							class="text-[10px] font-black text-muted-foreground group-hover:text-primary uppercase tracking-widest"
						>
							{$_('admin.reports.print_batch')}
						</span>
					</button>
					<div class="h-6 w-[1px] bg-border/80"></div>
					<Settings2
						size={18}
						class="text-muted-foreground cursor-pointer hover:text-primary transition-colors"
					/>
				</div>
			</div>

			<!-- Employee Cards Container -->
			<div class="p-4 bg-subtle-container/30 space-y-3">
				{#if loading}
					{#each Array(5) as _}
						<div
							class="bg-card rounded-lg h-24 w-full animate-pulse shadow-sm border border-border/40"
						></div>
					{/each}
				{:else if filteredDetails.length === 0}
					<div
						class="p-10 text-center text-muted-foreground font-black text-[10px] uppercase tracking-widest"
					>
						{$_('admin.reports.no_match')}
					</div>
				{:else}
					{#each filteredDetails as report}
						<!-- Card -->
						<div
							class="bg-card rounded-sm border border-border/40 p-5 flex flex-col xl:flex-row xl:items-center gap-6 hover:shadow-lg transition-shadow group relative overflow-hidden"
						>
							<div class="flex-1 flex items-center gap-4">
								<div
									class="w-12 h-12 rounded-sm bg-primary/5 flex items-center justify-center text-primary font-black text-sm shrink-0 uppercase shadow-sm"
								>
									{report.employee_name[0]}
								</div>
								<div class="min-w-0">
									<h3
										class="font-sans text-base font-black text-foreground truncate tracking-tight"
									>
										{report.employee_name}
									</h3>
									<p
										class="font-sans text-[9px] text-muted-foreground uppercase tracking-[0.2em] mt-1"
									>
										ID: {report.employee_id}
									</p>
								</div>
							</div>

							<!-- Metrics Cluster -->
							<div class="grid grid-cols-2 md:grid-cols-4 gap-6 flex-[2]">
								<div class="space-y-1">
									<span
										class="block font-sans text-[9px] text-muted-foreground/80 uppercase tracking-widest font-black"
									>
										{$_('common.hours')}
									</span>
									<span class="block font-sans text-sm font-black text-primary tabular-nums">
										{report.total_hours_worked?.toFixed(2)}h
									</span>
								</div>
								<div class="space-y-1">
									<span
										class="block font-sans text-[9px] text-muted-foreground/80 uppercase tracking-widest font-black"
									>
										{$_('common.status')}
									</span>
									<div class="flex items-center pt-0.5">
										{#if report.total_incidents > 0 || report.total_deductions > 0}
											<span
												class="bg-destructive/10 text-destructive text-[9px] px-2.5 py-1 rounded-md font-black uppercase tracking-widest flex items-center gap-1.5 border border-destructive/20 shadow-[inset_0_1px_2px_rgba(0,0,0,0.05)]"
											>
												{report.total_incidents} {$_('admin.reports.alerts')}
											</span>
										{:else}
											<span
												class="bg-emerald-500/10 text-emerald-600 text-[9px] px-2.5 py-1 rounded-md font-black uppercase tracking-widest flex items-center gap-1.5 border border-emerald-500/20 shadow-[inset_0_1px_2px_rgba(0,0,0,0.05)]"
											>
												{$_('admin.reports.clear_status')}
											</span>
										{/if}
									</div>
								</div>
								<div class="space-y-1">
									<span
										class="block font-sans text-[9px] text-muted-foreground/80 uppercase tracking-widest font-black"
									>
										{$_('admin.reports.deductions')}
									</span>
									<span
										class="block font-sans text-xs font-black text-destructive tabular-nums mt-0.5"
									>
										-${report.total_deductions?.toFixed(2)}
									</span>
								</div>
								<div class="space-y-1 text-right md:text-left">
									<span
										class="block font-sans text-[9px] text-muted-foreground/80 uppercase tracking-widest font-black"
									>
										{$_('admin.reports.net_pay')}
									</span>
									<span
										class="block font-sans text-lg font-black text-primary tabular-nums tracking-tighter"
									>
										${((report.total_earnings || 0) - (report.total_deductions || 0)).toLocaleString('es-MX', { minimumFractionDigits: 2 })}
									</span>
									<span class="block font-sans text-[8px] text-muted-foreground uppercase font-black tracking-widest mt-0.5">
										{$_('admin.reports.gross')}: ${report.total_earnings?.toFixed(2)}
									</span>
								</div>
							</div>

							<div class="hidden xl:flex items-center justify-end gap-1">
								<button
									onclick={() => viewIndividualPDF(report)}
									class="p-2.5 rounded-sm hover:bg-subtle-container transition-colors text-muted-foreground hover:text-primary focus:outline-none focus:ring-2 focus:ring-primary/20"
									title="Download Individual PDF"
								>
									<Download size={18} />
								</button>
								<button
									onclick={() => deleteReport(report)}
									class="p-2.5 rounded-sm hover:bg-destructive/5 transition-colors text-muted-foreground hover:text-destructive focus:outline-none"
									title="Delete Report"
								>
									<Trash2 size={18} />
								</button>
							</div>
						</div>
					{/each}
				{/if}
			</div>

			<!-- Sticky Footer CTA -->
			<div
				class="p-6 bg-card flex flex-col-reverse md:flex-row justify-between items-center gap-4 border-t border-border/40"
			>
				<a
					href="/admin/reports"
					class="font-sans text-[10px] font-black text-muted-foreground hover:text-primary transition-colors uppercase tracking-[0.2em]"
				>
					{$_('admin.reports.cancel_review')}
				</a>
				<button
					class="w-full md:w-auto bg-primary text-primary-foreground px-10 py-3 rounded-sm font-sans font-black text-[10px] uppercase tracking-[0.2em] shadow-xl shadow-primary/20 hover:scale-[1.02] active:scale-[0.98] transition-all"
				>
					{$_('admin.reports.approve_archive')}
				</button>
			</div>
		</div>

		<!-- Attendance System Notice -->
		<div class="max-w-xl" in:fly={{ y: 20, duration: 800, easing: quintOut, delay: 200 }}>
			<div
				class="bg-card border-l-4 border-primary rounded-r-xl p-6 flex flex-col justify-between shadow-sm"
			>
				<div>
					<h4
						class="font-sans text-[10px] font-black uppercase tracking-widest text-muted-foreground mb-2"
					>
						{$_('admin.reports.notice_title')}
					</h4>
					<p class="font-sans text-xs text-muted-foreground/80 leading-relaxed font-semibold">
						{$_('admin.reports.notice_body')}
					</p>
				</div>
			</div>
		</div>
	</main>
</div>

<style>
	.no-scrollbar::-webkit-scrollbar {
		display: none;
	}
	.no-scrollbar {
		-ms-overflow-style: none;
		scrollbar-width: none;
	}
</style>
