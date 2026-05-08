<script lang="ts">
	import { onMount } from 'svelte';
	import { fly } from 'svelte/transition';
	import { quintOut } from 'svelte/easing';
	import { _ } from 'svelte-i18n';
	import { page } from '$app/state';
	import { apiFetch } from '$lib/api';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import {
		Loader2,
		Mail,
		Badge as BadgeIcon,
		Calendar,
		MapPin,
		Briefcase,
		Clock,
		Activity,
		DollarSign,
		TrendingDown,
		AlertTriangle,
		History,
		ArrowLeft,
		MoreVertical,
		ExternalLink,
		ShieldAlert,
		Edit3,
		FileDown,
		ChevronRight
	} from 'lucide-svelte';

	let empId = $derived(page.params.id);
	let data = $state<any>(null);
	let loading = $state(true);
	let errorMsg = $state('');

	async function loadDetails() {
		loading = true;
		const res = await apiFetch(`/admin/employees/${empId}/details`);
		if (res.ok) {
			data = await res.json();
		} else {
			errorMsg = $_('admin.employees.details.not_found');
		}
		loading = false;
	}

	function formatCurrency(val: number) {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: 'USD'
		}).format(val);
	}

	const complianceScore = $derived(() => {
		if (!data || data.stats.total_attendances === 0) return 100;
		const rate = 100 - (data.stats.total_incidents / data.stats.total_attendances) * 100;
		return Math.max(0, Math.round(rate * 10) / 10);
	});

	onMount(loadDetails);
</script>

<div class="min-h-screen bg-[#f8fafb] pb-32">
	{#if loading}
		<div class="flex flex-col items-center justify-center min-h-[80vh] gap-6">
			<div class="relative">
				<div class="h-24 w-24 rounded-full border-4 border-slate-100 animate-pulse"></div>
				<Loader2
					class="h-10 w-10 animate-spin text-primary absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2"
				/>
			</div>
			<p class="font-black text-primary uppercase italic tracking-tighter text-xl">
				{$_('admin.employees.details.accessing')}
			</p>
		</div>
	{:else if errorMsg}
		<div class="max-w-md mx-auto p-12 text-center space-y-6">
			<AlertTriangle class="h-16 w-16 text-rose-500 mx-auto" />
			<h2 class="text-3xl font-black text-slate-900 tracking-tighter uppercase italic">
				{$_('admin.employees.details.not_found')}
			</h2>
			<Button
				href="/admin/employees"
				variant="outline"
				class="h-12 px-8 font-black uppercase italic">{$_('admin.employees.details.close_audit')}</Button
			>
		</div>
	{:else if data}
		<main class="max-w-5xl mx-auto px-6 py-12 space-y-16">
			<!-- Identity Hero -->
			<section
				class="flex flex-col items-center text-center space-y-6"
				in:fly={{ y: 20, duration: 800, easing: quintOut }}
			>
				<div class="relative">
					<div
						class="h-40 w-40 rounded-full bg-white border-8 border-slate-50 flex items-center justify-center shadow-2xl shadow-slate-200/50"
					>
						<span class="text-6xl font-black text-primary italic uppercase"
							>{data.employee.user_name[0]}</span
						>
					</div>
					<div
						class="absolute bottom-2 right-2 h-8 w-8 bg-white rounded-full flex items-center justify-center border-4 border-slate-50"
					>
						<div
							class="h-3 w-3 rounded-full {data.employee.is_active
								? 'bg-emerald-500 animate-pulse'
								: 'bg-slate-300'}"
						></div>
					</div>
				</div>

				<div class="space-y-4">
					<h2
						class="text-7xl font-black tracking-tighter text-primary italic uppercase leading-none"
					>
						{data.employee.user_name}
					</h2>
					<div class="flex flex-wrap items-center justify-center gap-6 text-slate-400">
						<span class="flex items-center gap-2 font-black text-[10px] uppercase tracking-widest">
							<Mail class="h-3 w-3" />
							{data.employee.email}
						</span>
						{#if data.employee.phone}
							<span class="flex items-center gap-2 font-black text-[10px] uppercase tracking-widest text-primary">
								<svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-phone"><path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.5 19.5 0 0 1-6-6 19.79 19.79 0 0 1-3.07-8.67A2 2 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72 12.84 12.84 0 0 0 .7 2.81 2 2 0 0 1-.45 2.11L8.09 9.91a16 16 0 0 0 6 6l2.19-1.32a2 2 0 0 1 2.11-.45 12.84 12.84 0 0 0 2.81.7A2 2 0 0 1 22 16.92z"/></svg>
								{data.employee.phone}
							</span>
						{/if}
						<span class="flex items-center gap-2 font-black text-[10px] uppercase tracking-widest">
							<BadgeIcon class="h-3 w-3" /> ID: {data.employee.id}
						</span>
						<span class="flex items-center gap-2 font-black text-[10px] uppercase tracking-widest">
							<Calendar class="h-3 w-3" /> {$_('admin.employees.details.joined')}: {data.employee.joined_at}
						</span>
					</div>
					<div
						class="inline-flex px-4 py-1.5 rounded-full {data.employee.is_active
							? 'bg-emerald-50 text-emerald-600'
							: 'bg-slate-100 text-slate-500'} font-black text-[10px] uppercase tracking-[0.2em] italic"
					>
						{data.employee.is_active ? $_('admin.employees.details.active_status') : $_('admin.employees.details.inactive_status')}
					</div>
				</div>
			</section>

			<!-- Assignment Matrix -->
			<section
				class="space-y-8"
				in:fly={{ y: 20, duration: 800, easing: quintOut, delay: 100 }}
			>
				<div class="flex items-center gap-3">
					<div class="h-10 w-10 bg-primary/5 rounded-sm flex items-center justify-center">
						<Briefcase class="h-5 w-5 text-primary" />
					</div>
					<h3 class="text-xl font-black text-primary uppercase italic tracking-tighter">
						{$_('admin.employees.details.assignment_matrix')}
					</h3>
				</div>

				<div class="grid grid-cols-1 md:grid-cols-3 gap-6">
					<a
						href="/admin/centers/{data.employee.work_center_id}"
						class="group bg-white p-8 rounded-sm border border-slate-100 shadow-xl shadow-slate-200/40 hover:border-primary/20 transition-all"
					>
						<p class="text-slate-400 font-black text-[10px] uppercase tracking-[0.2em] mb-4">
							{$_('admin.employees.details.cost_center')}
						</p>
						<h4
							class="text-2xl font-black text-primary italic uppercase tracking-tighter mb-2 group-hover:translate-x-1 transition-transform"
						>
							{data.employee.center_name}
						</h4>
						<span class="flex items-center gap-2 text-slate-400 font-black text-[9px] uppercase">
							<MapPin class="h-3 w-3" /> {$_('admin.employees.details.hq_perimeter')}
						</span>
					</a>

					<a
						href="/admin/positions/{data.employee.position_id}"
						class="group bg-white p-8 rounded-sm border border-slate-100 shadow-xl shadow-slate-200/40 hover:border-primary/20 transition-all"
					>
						<p class="text-slate-400 font-black text-[10px] uppercase tracking-[0.2em] mb-4">
							{$_('admin.employees.details.position')}
						</p>
						<h4
							class="text-2xl font-black text-primary italic uppercase tracking-tighter mb-2 group-hover:translate-x-1 transition-transform"
						>
							{data.employee.position_name}
						</h4>
						<span class="flex items-center gap-2 text-slate-400 font-black text-[9px] uppercase">
							<BadgeIcon class="h-3 w-3" /> {$_('admin.employees.details.management_tier')}
						</span>
					</a>

					<a
						href="/admin/shifts/{data.employee.work_shift_id}"
						class="group bg-white p-8 rounded-sm border border-slate-100 shadow-xl shadow-slate-200/40 hover:border-primary/20 transition-all"
					>
						<p class="text-slate-400 font-black text-[10px] uppercase tracking-[0.2em] mb-4">
							{$_('admin.employees.details.shift')}
						</p>
						<h4
							class="text-2xl font-black text-primary italic uppercase tracking-tighter mb-2 group-hover:translate-x-1 transition-transform"
						>
							{data.employee.shift_name || $_('admin.employees.details.individual_load')}
						</h4>
						<span class="flex items-center gap-2 text-slate-400 font-black text-[9px] uppercase">
							<Clock class="h-3 w-3" /> {$_('admin.employees.details.standard_cycle')}
						</span>
					</a>
				</div>
			</section>

			<!-- Operational Pulse -->
			<section
				class="space-y-8"
				in:fly={{ y: 20, duration: 800, easing: quintOut, delay: 200 }}
			>
				<div class="flex items-center gap-3">
					<div class="h-10 w-10 bg-primary/5 rounded-sm flex items-center justify-center">
						<Activity class="h-5 w-5 text-primary" />
					</div>
					<h3 class="text-xl font-black text-primary uppercase italic tracking-tighter">
						{$_('admin.employees.details.operational_pulse')}
					</h3>
				</div>

				<div class="grid grid-cols-1 md:grid-cols-3 gap-8">
					<!-- Compliance -->
					<div class="space-y-4">
						<div class="flex justify-between items-end">
							<p class="text-slate-400 font-black text-[10px] uppercase tracking-[0.2em]">
								{$_('admin.employees.details.compliance')}
							</p>
							<span class="text-2xl font-black text-primary italic tracking-tighter"
								>{complianceScore()}%</span
							>
						</div>
						<div class="h-2 w-full bg-slate-100 rounded-full overflow-hidden">
							<div
								class="h-full bg-primary transition-all duration-1000"
								style="width: {complianceScore()}%"
							></div>
						</div>
					</div>

					<!-- Financial Impact -->
					<div
						class="bg-primary p-8 rounded-sm text-white shadow-2xl shadow-primary/30 relative group"
					>
						<DollarSign
							class="absolute top-4 right-4 h-12 w-12 text-white/10 group-hover:scale-125 transition-transform"
						/>
						<p class="text-white/40 font-black text-[10px] uppercase tracking-widest mb-4">
							{$_('admin.employees.details.financial_impact')}
						</p>
						<div class="flex items-baseline gap-2">
							<h4 class="text-5xl font-black italic tracking-tighter">
								{formatCurrency(data.stats.total_earnings).replace('$', '')}
							</h4>
							<span class="text-xs font-black opacity-40 uppercase">USD</span>
						</div>
						{#if data.stats.total_incidents > 0}
							<div class="mt-6 flex items-center gap-2 text-rose-300">
								<TrendingDown class="h-4 w-4" />
								<span class="font-black text-[10px] uppercase tracking-widest italic leading-none"
									>-{data.stats.total_incidents} {$_('admin.employees.details.deductions_applied')}</span
								>
							</div>
						{/if}
					</div>

					<!-- Velocity (WIP) -->
					<div
						class="border-2 border-dashed border-slate-100 p-8 rounded-sm flex flex-col justify-center items-center text-center space-y-2 opacity-50"
					>
						<Activity class="h-6 w-6 text-slate-300" />
						<p class="text-slate-300 font-black text-[10px] uppercase tracking-widest">
							{$_('admin.employees.details.workload_velocity')}
						</p>
						<p class="text-[9px] text-slate-200 uppercase font-bold italic">
							{$_('admin.employees.details.analytics_offline')}
						</p>
					</div>
				</div>
			</section>

			<div
				class="grid grid-cols-1 lg:grid-cols-2 gap-16 pt-8"
				in:fly={{ y: 20, duration: 800, easing: quintOut, delay: 300 }}
			>
				<!-- Incidents -->
				<section class="space-y-8">
					<h3
						class="text-xl font-black text-primary uppercase italic tracking-tighter flex items-center gap-3"
					>
						<AlertTriangle class="h-5 w-5 text-rose-500" /> {$_('admin.employees.details.active_incidents')}
					</h3>

					<div class="space-y-4">
						{#each data.recent_attendance.filter((a) => a.is_late).slice(0, 3) as incident}
							<div
								class="bg-rose-50/50 border-l-4 border-rose-500 p-6 rounded-sm space-y-3 hover:bg-rose-50 transition-colors"
							>
								<div class="flex justify-between items-start">
									<h5 class="text-sm font-black text-rose-700 uppercase italic tracking-tight">
										{$_('admin.employees.details.late_detected')}
									</h5>
									<span class="text-[9px] font-black text-rose-300 uppercase tracking-widest"
										>{new Date(incident.check_in).toLocaleDateString()}</span
									>
								</div>
								<p class="text-xs text-rose-900/60 font-medium leading-relaxed italic">
									{$_('admin.employees.details.late_hint')}
								</p>
								<div class="flex gap-2">
									<Badge
										class="bg-rose-100 text-rose-600 font-black text-[8px] uppercase tracking-widest border-none"
										>{$_('admin.employees.details.gravity_low')}</Badge
									>
									<Badge
										class="bg-white text-rose-400 font-black text-[8px] uppercase tracking-widest border border-rose-100"
										>{$_('admin.employees.details.auto_logged')}</Badge
									>
								</div>
							</div>
						{:else}
							<div class="py-20 text-center border-2 border-dashed border-slate-50 rounded-sm">
								<p class="text-[10px] font-black text-slate-300 uppercase tracking-widest italic">
									{$_('admin.employees.details.pristine_record')}
								</p>
							</div>
						{/each}
					</div>
				</section>

				<!-- Audit history -->
				<section class="space-y-8">
					<h3
						class="text-xl font-black text-primary uppercase italic tracking-tighter flex items-center gap-3"
					>
						<History class="h-5 w-5 text-primary" /> {$_('admin.employees.details.audit_timeline')}
					</h3>

					<div class="relative pl-8 border-l border-slate-100 space-y-10 py-2">
						{#each data.recent_attendance as event}
							<div class="relative">
								<div
									class="absolute -left-[41px] top-0 h-4 w-4 rounded-full bg-white border-4 {event.is_late
										? 'border-rose-500 ring-4 ring-rose-50'
										: 'border-emerald-500 ring-4 ring-emerald-50'}"
								></div>
								<div class="space-y-3">
									<div class="text-[10px] font-black text-slate-400 uppercase tracking-widest">
										{new Date(event.check_in).toLocaleString()}
									</div>
									<div
										class="bg-white p-6 rounded-sm border border-slate-100 shadow-xl shadow-slate-200/20 group hover:border-primary/10 transition-all"
									>
										<div class="flex justify-between items-center mb-1">
											<span class="font-black text-sm text-primary uppercase italic tracking-tight">
												{event.is_late ? $_('admin.employees.details.late_checkin') : $_('admin.employees.details.standard_log')}
											</span>
											<ChevronRight
												class="h-4 w-4 text-slate-200 group-hover:text-primary transition-colors"
											/>
										</div>
										<p class="text-[9px] font-black text-slate-400 uppercase tracking-widest">
											{$_('admin.employees.details.terminal')}: {event.work_center_name}
										</p>
									</div>
								</div>
							</div>
						{:else}
							<div class="text-center py-20">
								<p class="text-[10px] font-black text-slate-300 uppercase tracking-widest italic">
									{$_('admin.employees.details.empty_ledger')}
								</p>
							</div>
						{/each}
					</div>
				</section>
			</div>
		</main>

		<!-- Bottom Executive Bar -->
		<div
			class="fixed bottom-0 w-full z-50 bg-white border-t border-slate-100 px-8 py-5 hidden md:flex items-center justify-between"
		>
			<div class="flex items-center gap-3 opacity-40">
				<ShieldAlert class="h-5 w-5 text-slate-400" />
				<span class="font-black text-[9px] text-slate-400 uppercase tracking-[0.2em]"
					>{$_('admin.employees.details.clearance_required')}</span
				>
			</div>
			<div class="flex items-center gap-4">
				<Button
					variant="ghost"
					class="font-black text-[10px] uppercase tracking-widest text-primary gap-2 h-12 px-6 bg-slate-50 opacity-50 cursor-not-allowed"
				>
					<FileDown class="h-4 w-4" /> {$_('admin.employees.details.export_ledger')}
				</Button>
				<Button
					class="bg-primary text-white font-black text-[10px] uppercase tracking-widest gap-2 h-12 px-8 shadow-xl shadow-primary/20 opacity-50 cursor-not-allowed"
				>
					<Edit3 class="h-4 w-4" /> {$_('admin.employees.details.modify_profile')}
				</Button>
			</div>
		</div>
	{/if}
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
