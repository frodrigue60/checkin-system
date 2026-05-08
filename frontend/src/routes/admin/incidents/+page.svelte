<script lang="ts">
	import { onMount } from 'svelte';
	import { fly, fade } from 'svelte/transition';
	import { flip } from 'svelte/animate';
	import { quintOut } from 'svelte/easing';
	import { _ } from 'svelte-i18n';
	import { apiFetch } from '$lib/api';
	import type { IncidentRichDTO } from '$lib/types/models';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import {
		AlertTriangle,
		Search,
		User,
		Eye,
		ChevronLeft,
		ChevronRight,
		Clock,
		MapPin,
		CheckCircle2,
		XCircle,
		MoreHorizontal,
		Filter,
		ShieldAlert,
		ShieldCheck,
		Undo2,
		Loader2
	} from 'lucide-svelte';
	import BatchActionBar from '$lib/components/BatchActionBar.svelte';
	import { notifications } from '$lib/notifications.svelte';

	let incidents = $state<IncidentRichDTO[]>([]);
	let loading = $state(true);
	let page = $state(1);
	let totalPages = $state(1);
	let totalIncidents = $state(0);
	
	let searchQuery = $state('');
	let filterType = $state('all');
	let filterStatus = $state('pending');
	let filterStart = $state('');
	let filterEnd = $state('');
	let filterShift = $state('all');
	let filterCenter = $state('all');
	let filterPosition = $state('all');
	let filterShiftType = $state('all');

	let shifts = $state<any[]>([]);
	let centers = $state<any[]>([]);
	let positions = $state<any[]>([]);

	let selectedIds = $state(new Set<number>());
	let bulkLoading = $state(false);
	let mounted = $state(false);

	let showJustifyDialog = $state(false);
	let justificationNote = $state('');
	let targetsToJustify = $state<number[]>([]);

	async function loadIncidents(p = 1) {
		loading = true;
		
		let startISO = '';
		let endISO = '';

		if (filterStart) {
			const d = new Date(filterStart + 'T00:00:00');
			startISO = d.toISOString();
		}
		if (filterEnd) {
			const d = new Date(filterEnd + 'T23:59:59');
			endISO = d.toISOString();
		}

		const query = new URLSearchParams({
			page: p.toString(),
			limit: '20',
			search: searchQuery,
			type: filterType,
			status: filterStatus,
			start: startISO,
			end: endISO,
			shift_id: filterShift,
			center_id: filterCenter,
			position_id: filterPosition,
			shift_type: filterShiftType
		});

		try {
			const res = await apiFetch<{
				data: IncidentRichDTO[];
				total_pages: number;
				total: number;
				page: number;
			}>(`/admin/incidents?${query.toString()}`);
			
			if (res.ok) {
				const data = await res.json();
				incidents = data.data || [];
				totalPages = data.total_pages;
				totalIncidents = data.total;
				page = data.page;
			}
		} catch (e) {
			console.error('Error loading incidents:', e);
		} finally {
			loading = false;
		}
	}

	// Debounced search
	let searchTimeout: any;
	$effect(() => {
		if (mounted) {
			clearTimeout(searchTimeout);
			searchTimeout = setTimeout(() => {
				loadIncidents(1);
			}, 300);
		}
	});

	$effect(() => {
		if (mounted) {
			filterType, filterStatus, filterStart, filterEnd, filterShift, filterCenter, filterPosition;
			loadIncidents(1);
		}
	});

	function toggleSelectAll() {
		if (selectedIds.size === incidents.length) {
			selectedIds.clear();
		} else {
			incidents.forEach(i => selectedIds.add(i.id));
		}
		selectedIds = new Set(selectedIds);
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
		if (selectedIds.size === 0) return;
		targetsToJustify = Array.from(selectedIds);
		justificationNote = '';
		showJustifyDialog = true;
	}

	function clearSelection() {
		selectedIds.clear();
		selectedIds = new Set();
	}

	async function confirmJustification() {
		if (targetsToJustify.length === 0) return;
		
		bulkLoading = true;
		try {
			const res = await apiFetch('/bulk/incidents/justify', {
				method: 'POST',
				body: JSON.stringify({
					incident_ids: targetsToJustify,
					note: justificationNote
				})
			});

			if (res.ok) {
				notifications.success($_('admin.incidents.justify_success'));
				selectedIds.clear();
				showJustifyDialog = false;
				loadIncidents(page);
			} else {
				const err = await res.json();
				notifications.error(err.message || $_('admin.incidents.justify_error'));
			}
		} catch (e) {
			notifications.error($_('errors.server_error'));
		} finally {
			bulkLoading = false;
		}
	}

	function getStatusBadge(status: string) {
		switch (status) {
			case 'pending': return 'bg-amber-50 text-amber-600 border-amber-100';
			case 'justified': return 'bg-emerald-50 text-emerald-600 border-emerald-100';
			case 'approved': return 'bg-blue-50 text-blue-600 border-blue-100';
			case 'rejected': return 'bg-rose-50 text-rose-600 border-rose-100';
			default: return 'bg-slate-50 text-slate-600 border-slate-100';
		}
	}

	async function loadOptions() {
		try {
			const [sRes, cRes, pRes] = await Promise.all([
				apiFetch<any[]>('/admin/shifts'),
				apiFetch<any[]>('/admin/centers'),
				apiFetch<any[]>('/admin/positions')
			]);
			if (sRes.ok) shifts = await sRes.json();
			if (cRes.ok) centers = await cRes.json();
			if (pRes.ok) positions = await pRes.json();
		} catch (e) {
			console.error('Error loading filter options:', e);
		}
	}

	onMount(() => {
		loadOptions();
		loadIncidents(1);
		mounted = true;
	});
</script>

<main class="pb-24 px-6 max-w-7xl mx-auto">
	{#if mounted}
	<!-- Hero Header -->
	<section
		class="mt-8 mb-10"
		in:fly={{ y: 20, duration: 800, easing: quintOut }}
	>
		
		<div class="flex justify-between items-end">
			<div>
				<h2 class="text-4xl font-black text-primary leading-none tracking-tighter mb-4">
					{$_('admin.incidents.header')}
				</h2>
			</div>
			<div class="hidden md:flex flex-col items-end">
				<span class="text-3xl font-black text-slate-900 leading-none">{totalIncidents}</span>
				<span class="text-[9px] font-black text-slate-400 uppercase tracking-[0.2em] mt-1">{$_('admin.incidents.total_pending')}</span>
			</div>
		</div>
	</section>

	<!-- Filters & Search -->
	<div class="mb-8 space-y-4" in:fly={{ y: 20, duration: 800, easing: quintOut, delay: 100 }}>
		<div class="flex flex-wrap gap-4 items-end">
			<div class="flex-1 min-w-[300px] relative group">
				<Search class="absolute left-4 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400 group-focus-within:text-primary transition-colors" />
				<Input 
					bind:value={searchQuery}
					placeholder={$_('admin.incidents.search_placeholder')}
					class="pl-12 h-14 bg-white border-slate-100 shadow-sm focus:ring-primary/20 rounded-sm font-bold text-sm"
				/>
			</div>

			<div class="bg-white border border-slate-100 p-3 rounded-sm shadow-sm flex flex-col justify-center min-w-[180px]">
				<label class="text-[9px] font-black uppercase text-slate-400 mb-1">{$_('admin.incidents.filter_type')}</label>
				<select
					bind:value={filterType}
					class="bg-transparent border-none text-xs font-bold focus:ring-0 p-0 text-primary uppercase"
				>
					<option value="all">{$_('common.all')}</option>
					<option value="late">{$_('admin.incidents.types.late')}</option>
					<option value="out_of_range">{$_('admin.incidents.types.out_of_range')}</option>
				</select>
			</div>

			<div class="bg-white border border-slate-100 p-3 rounded-sm shadow-sm flex flex-col justify-center min-w-[180px]">
				<label class="text-[9px] font-black uppercase text-slate-400 mb-1">{$_('admin.incidents.filter_status')}</label>
				<select
					bind:value={filterStatus}
					class="bg-transparent border-none text-xs font-bold focus:ring-0 p-0 text-primary uppercase"
				>
					<option value="pending">{$_('admin.incidents.status_pending')}</option>
					<option value="justified">{$_('admin.incidents.status_justified')}</option>
					<option value="approved">{$_('admin.incidents.status_approved')}</option>
					<option value="rejected">{$_('admin.incidents.status_rejected')}</option>
					<option value="all">{$_('admin.incidents.view_all')}</option>
				</select>
			</div>

			<div class="bg-white border border-slate-100 p-3 rounded-sm shadow-sm flex flex-col justify-center">
				<label class="text-[9px] font-black uppercase text-slate-400 mb-1">{$_('admin.incidents.filter_date')}</label>
				<input 
					type="date" 
					bind:value={filterStart}
					class="bg-transparent border-none text-xs font-bold focus:ring-0 p-0 text-primary uppercase"
				/>
			</div>

			<div class="bg-white border border-slate-100 p-3 rounded-sm shadow-sm flex flex-col justify-center min-w-[150px]">
				<label class="text-[9px] font-black uppercase text-slate-400 mb-1">{$_('admin.incidents.filter_center')}</label>
				<select
					bind:value={filterCenter}
					class="bg-transparent border-none text-xs font-bold focus:ring-0 p-0 text-primary uppercase"
				>
					<option value="all">{$_('common.all_centers')}</option>
					{#each centers as center}
						<option value={center.id}>{center.name}</option>
					{/each}
				</select>
			</div>

			<div class="bg-white border border-slate-100 p-3 rounded-sm shadow-sm flex flex-col justify-center min-w-[150px]">
				<label class="text-[9px] font-black uppercase text-slate-400 mb-1">{$_('admin.incidents.filter_shift')}</label>
				<select
					bind:value={filterShift}
					class="bg-transparent border-none text-xs font-bold focus:ring-0 p-0 text-primary uppercase"
				>
					<option value="all">{$_('common.all_shifts')}</option>
					{#each shifts as shift}
						<option value={shift.id}>{shift.name}</option>
					{/each}
				</select>
			</div>

			<div class="bg-white border border-slate-100 p-3 rounded-sm shadow-sm flex flex-col justify-center min-w-[150px]">
				<label class="text-[9px] font-black uppercase text-slate-400 mb-1">{$_('admin.incidents.filter_position')}</label>
				<select
					bind:value={filterPosition}
					class="bg-transparent border-none text-xs font-bold focus:ring-0 p-0 text-primary uppercase"
				>
					<option value="all">{$_('common.all_positions')}</option>
					{#each positions as pos}
						<option value={pos.id}>{pos.name}</option>
					{/each}
				</select>
			</div>

			<div class="bg-white border border-slate-100 p-3 rounded-sm shadow-sm flex flex-col justify-center min-w-[150px]">
				<label class="text-[9px] font-black uppercase text-slate-400 mb-1">{$_('admin.incidents.filter_type_label')}</label>
				<select
					bind:value={filterShiftType}
					class="bg-transparent border-none text-xs font-bold focus:ring-0 p-0 text-primary uppercase"
				>
					<option value="all">{$_('common.all')}</option>
					<option value="fixed">{$_('admin.shifts.type_fixed')}</option>
					<option value="flexible">{$_('admin.shifts.type_flexible')}</option>
					<option value="field">{$_('admin.shifts.type_field')}</option>
				</select>
			</div>
		</div>
	</div>

	<!-- Table -->
	<div class="bg-white rounded-md border border-slate-100 shadow-2xl shadow-slate-200/50 overflow-hidden" in:fly={{ y: 20, duration: 800, easing: quintOut, delay: 200 }}>
		<div class="overflow-x-auto">
			<table class="w-full text-left border-collapse">
				<thead>
					<tr class="bg-slate-50/50 border-b border-slate-100">
						<th class="px-6 py-4 w-10">
							<input 
								type="checkbox"
								checked={selectedIds.size === incidents.length && incidents.length > 0}
								onchange={toggleSelectAll}
								class="h-5 w-5 rounded border-slate-200 text-primary focus:ring-primary cursor-pointer transition-all"
							/>
						</th>
						<th class="px-6 py-4 text-[10px] font-black text-slate-400 uppercase tracking-[0.2em]">{$_('admin.incidents.table.employee')}</th>
						<th class="px-6 py-4 text-[10px] font-black text-slate-400 uppercase tracking-[0.2em]">{$_('admin.incidents.table.type')}</th>
						<th class="px-6 py-4 text-[10px] font-black text-slate-400 uppercase tracking-[0.2em]">{$_('admin.incidents.table.details')}</th>
						<th class="px-6 py-4 text-[10px] font-black text-slate-400 uppercase tracking-[0.2em]">{$_('admin.incidents.table.date')}</th>
						<th class="px-6 py-4 text-[10px] font-black text-slate-400 uppercase tracking-[0.2em]">{$_('admin.incidents.table.status')}</th>
						<th class="px-6 py-4"></th>
					</tr>
				</thead>
				<tbody class="divide-y divide-slate-50">
					{#if loading}
						{#each Array(5) as _}
							<tr class="animate-pulse">
								<td class="px-6 py-4"><div class="h-4 w-4 bg-slate-100 rounded"></div></td>
								{#each Array(5) as _}
									<td class="px-6 py-4"><div class="h-4 bg-slate-100 rounded w-full"></div></td>
								{/each}
								<td class="px-6 py-4"></td>
							</tr>
						{/each}
					{:else if incidents.length === 0}
						<tr>
							<td colspan="7" class="px-6 py-20 text-center">
								<ShieldAlert size={40} class="mx-auto text-slate-200 mb-4" />
								<p class="text-slate-400 font-bold uppercase tracking-widest text-xs">{$_('admin.incidents.no_results')}</p>
							</td>
						</tr>
					{:else}
						{#each incidents as incident (incident.id)}
							<tr class="hover:bg-slate-50/80 transition-colors group {selectedIds.has(incident.id) ? 'bg-primary/5' : ''}">
								<td class="px-6 py-4">
									<input 
										type="checkbox"
										checked={selectedIds.has(incident.id)}
										onchange={() => toggleSelect(incident.id)}
										class="h-5 w-5 rounded border-slate-200 text-primary focus:ring-primary cursor-pointer transition-all"
									/>
								</td>
								<td class="px-6 py-4">
									<div class="flex items-center gap-3">
										<div class="h-8 w-8 rounded-full bg-slate-100 flex items-center justify-center border border-slate-200 group-hover:border-primary/30 transition-all">
											<User size={14} class="text-slate-500 group-hover:text-primary" />
										</div>
										<div class="flex flex-col">
											<span class="text-xs font-black text-slate-700 tracking-tight">{incident.employee_name}</span>
											<span class="text-[9px] font-bold text-slate-400 uppercase tracking-tighter">{incident.work_center_name}</span>
										</div>
									</div>
								</td>
								<td class="px-6 py-4">
									<Badge variant="outline" class="font-black text-[9px] px-2 py-0.5 rounded-sm {incident.type === 'late' ? 'bg-rose-50 text-rose-600 border-rose-100' : 'bg-orange-50 text-orange-600 border-orange-100'}">
										{$_(`admin.incidents.types.${incident.type}`)}
									</Badge>
								</td>
								<td class="px-6 py-4">
									<div class="flex items-center gap-2 text-slate-600">
										{#if incident.type === 'late'}
											<Clock size={12} class="text-rose-400" />
											<span class="text-[11px] font-bold">{$_('admin.incidents.details.minutes_late', { values: { n: incident.delay_minutes } })}</span>
										{:else}
											<MapPin size={12} class="text-orange-400" />
											<span class="text-[11px] font-bold">{$_('admin.incidents.details.meters_away', { values: { n: Math.round(incident.distance) } })}</span>
										{/if}
									</div>
								</td>
								<td class="px-6 py-4">
									<span class="text-[10px] font-black text-slate-500 uppercase">{new Date(incident.attendance_date + 'T12:00:00').toLocaleDateString(undefined, { day: 'numeric', month: 'short', year: 'numeric' })}</span>
								</td>
								<td class="px-6 py-4">
									<Badge variant="outline" class="font-black text-[9px] px-2 py-0.5 rounded-sm {getStatusBadge(incident.status)}">
										{$_(`admin.incidents.status_${incident.status}`)}
									</Badge>
								</td>
								<td class="px-6 py-4 text-right">
									<div class="flex items-center justify-end gap-2">
										{#if incident.status === 'pending'}
											<Button 
												variant="ghost" 
												size="icon" 
												class="h-8 w-8 text-emerald-500 hover:bg-emerald-50 opacity-0 group-hover:opacity-100 transition-all"
												onclick={() => {
													targetsToJustify = [incident.id];
													justificationNote = '';
													showJustifyDialog = true;
												}}
											>
												<ShieldCheck size={16} />
											</Button>
										{/if}
										<a 
											href="/admin/attendance/{incident.attendance_id}"
											class="h-8 w-8 flex items-center justify-center rounded-md text-slate-400 hover:text-primary hover:bg-slate-100 opacity-0 group-hover:opacity-100 transition-all"
										>
											<Eye size={16} />
										</a>
									</div>
								</td>
							</tr>
						{/each}
					{/if}
				</tbody>
			</table>
		</div>

		<!-- Pagination -->
		<div class="px-6 py-4 bg-slate-50/50 border-t border-slate-100 flex justify-between items-center">
			<span class="text-[10px] font-black text-slate-400 uppercase tracking-widest">
				{$_('common.pagination', { values: { current: page, total: totalPages } })}
			</span>
			<div class="flex gap-2">
				<Button 
					variant="outline" 
					size="sm" 
					class="h-9 px-4 rounded-sm font-black text-[10px] uppercase tracking-widest gap-2 disabled:opacity-50"
					disabled={page === 1}
					onclick={() => loadIncidents(page - 1)}
				>
					<ChevronLeft size={14} />
					{$_('common.previous')}
				</Button>
				<Button 
					variant="outline" 
					size="sm" 
					class="h-9 px-4 rounded-sm font-black text-[10px] uppercase tracking-widest gap-2 disabled:opacity-50"
					disabled={page >= totalPages}
					onclick={() => loadIncidents(page + 1)}
				>
					{$_('common.next')}
					<ChevronRight size={14} />
				</Button>
			</div>
		</div>
	</div>
	{/if}
</main>

<!-- Bulk Action Bar -->
{#if selectedIds.size > 0}
	<BatchActionBar 
		selectedCount={selectedIds.size}
		onClear={clearSelection}
	>
		<Button 
			variant="default" 
			class="bg-slate-900 hover:bg-slate-800 text-white font-black uppercase text-[10px] tracking-widest h-10 px-6 rounded-xl shadow-lg shadow-slate-900/20"
			onclick={handleBulkJustify}
			disabled={bulkLoading}
		>
			{#if bulkLoading}
				<Loader2 class="mr-2 h-3 w-3 animate-spin" />
			{:else}
				<ShieldCheck class="mr-2 h-3 w-3" />
			{/if}
			{$_('admin.incidents.bulk_justify')}
		</Button>
	</BatchActionBar>
{/if}

<!-- Justify Dialog -->
<Dialog.Root open={showJustifyDialog} onOpenChange={(o) => !o && (showJustifyDialog = false)}>
	<Dialog.Content class="sm:max-w-md border-none p-10">
		<Dialog.Header>
			<div class="flex items-center gap-3 mb-4">
				<div class="h-10 w-10 rounded-md bg-emerald-50 text-emerald-600 flex items-center justify-center">
					<ShieldCheck size={20} />
				</div>
				<Dialog.Title class="text-2xl font-black tracking-tighter">{$_('admin.incidents.justify_dialog_title')}</Dialog.Title>
			</div>
			<Dialog.Description class="text-xs font-black text-slate-400 uppercase tracking-[0.15em] mb-6">
				{$_('admin.incidents.justify_dialog_description', { values: { count: targetsToJustify.length } })}
			</Dialog.Description>
		</Dialog.Header>

		<div class="space-y-4 py-4">
			<div class="space-y-2">
				<label class="text-[10px] font-black uppercase text-slate-500 tracking-widest">{$_('admin.incidents.justification_note')}</label>
				<textarea
					bind:value={justificationNote}
					placeholder={$_('admin.incidents.justification_placeholder')}
					class="w-full min-h-[120px] bg-slate-50 border border-slate-100 rounded-sm p-4 text-sm font-bold focus:ring-primary/20 focus:border-primary transition-all"
				></textarea>
			</div>
		</div>

		<Dialog.Footer class="mt-8 flex gap-3">
			<Button 
				variant="ghost" 
				class="flex-1 font-black uppercase text-[10px] tracking-widest h-12"
				onclick={() => showJustifyDialog = false}
			>
				{$_('common.cancel')}
			</Button>
			<Button 
				class="flex-1 bg-slate-900 text-white hover:bg-slate-800 font-black uppercase text-[10px] tracking-widest h-12 shadow-lg shadow-slate-900/20"
				onclick={confirmJustification}
				disabled={bulkLoading}
			>
				{bulkLoading ? $_('admin.incidents.processing') : $_('admin.incidents.confirm')}
			</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>

<style>
	:global(.lucide) {
		stroke-width: 2.5px;
	}
</style>
