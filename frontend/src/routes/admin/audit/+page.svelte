<script lang="ts">
	import { onMount } from 'svelte';
	import { fly } from 'svelte/transition';
	import { quintOut } from 'svelte/easing';
	import { _ } from 'svelte-i18n';
	import { apiFetch } from '$lib/api';
	import type { AuditLog } from '$lib/types/models';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import {
		History,
		Search,
		User,
		Eye,
		ChevronLeft,
		ChevronRight,
		Shield,
		CalendarClock,
		Globe,
		Terminal
	} from 'lucide-svelte';

	let logs = $state<AuditLog[]>([]);
	let loading = $state(true);
	let page = $state(1);
	let totalPages = $state(1);
	let totalLogs = $state(0);
	
	let filterAction = $state('all');
	let filterEntity = $state('all');
	let filterStart = $state('');
	let filterEnd = $state('');

	let selectedLog = $state<(AuditLog & { old_parsed?: any; new_parsed?: any }) | null>(null);
	let showDetails = $state(false);
	let mounted = $state(false);

	async function loadLogs(p = 1) {
		loading = true;
		
		let startISO = '';
		let endISO = '';

		if (filterStart) {
			// Create date in local timezone at start of day
			const d = new Date(filterStart + 'T00:00:00');
			startISO = d.toISOString();
		}
		if (filterEnd) {
			// Create date in local timezone at end of day
			const d = new Date(filterEnd + 'T23:59:59');
			endISO = d.toISOString();
		}

		const query = new URLSearchParams({
			page: p.toString(),
			limit: '20',
			action: filterAction,
			entity: filterEntity,
			start: startISO,
			end: endISO
		});

		const res = await apiFetch<{
			data: AuditLog[];
			total_pages: number;
			total: number;
			page: number;
		}>(`/admin/audit-logs?${query.toString()}`);
		
		if (res.ok) {
			const data = await res.json();
			logs = data.data || [];
			totalPages = data.total_pages;
			totalLogs = data.total;
			page = data.page;
		}
		loading = false;
	}

	$effect(() => {
		if (mounted) {
			filterAction, filterEntity, filterStart, filterEnd;
			loadLogs(1);
		}
	});

	function viewDetails(log: AuditLog) {
		selectedLog = {
			...log,
			old_parsed: log.old_value ? JSON.parse(log.old_value) : null,
			new_parsed: log.new_value ? JSON.parse(log.new_value) : null
		};
		showDetails = true;
	}

	function getActionColor(action: string) {
		if (action.includes('create')) return 'bg-emerald-50 text-emerald-600 border-emerald-100';
		if (action.includes('update')) return 'bg-blue-50 text-blue-600 border-blue-100';
		if (action.includes('delete')) return 'bg-rose-50 text-rose-600 border-rose-100';
		if (action.includes('generate')) return 'bg-amber-50 text-amber-600 border-amber-100';
		return 'bg-slate-50 text-slate-600 border-slate-100';
	}

	onMount(() => {
		loadLogs(1);
		mounted = true;
	});
</script>

<main class="pb-24 px-6 max-w-6xl mx-auto">
	{#if mounted}
	<!-- Hero Header -->
	<section
		class="mt-8 mb-10"
		in:fly={{ y: 20, duration: 800, easing: quintOut }}
	>
		<span class="text-primary font-bold tracking-widest text-[10px] uppercase mb-2 block"
			>{$_('admin.audit.header')}</span
		>
		<div class="flex justify-between items-end">
			<div>
				<h2 class="text-4xl font-black text-primary leading-none tracking-tighter mb-4">
					{$_('admin.audit.title')}
				</h2>
				<p class="text-sm font-bold text-slate-400 uppercase tracking-widest max-w-2xl">
					{$_('admin.audit.description')}
				</p>
			</div>
			<div class="hidden md:flex flex-col items-end">
				<span class="text-3xl font-black text-slate-900 leading-none">{totalLogs}</span>
				<span class="text-[9px] font-black text-slate-400 uppercase tracking-[0.2em] mt-1">Registros Totales</span>
			</div>
		</div>
		<div class="w-12 h-1 bg-primary/20 rounded-full mt-6"></div>
	</section>

	<!-- Search & Filters -->
	<div class="mb-10 grid grid-cols-1 md:grid-cols-4 gap-4" in:fly={{ y: 20, duration: 800, easing: quintOut, delay: 100 }}>
		<div class="bg-white border border-slate-100 p-4 rounded-sm shadow-sm flex flex-col justify-center">
			<label class="text-[9px] font-black uppercase text-slate-400 mb-1">Entidad</label>
			<select
				bind:value={filterEntity}
				class="bg-transparent border-none text-xs font-bold focus:ring-0 p-0 text-primary uppercase"
			>
				<option value="all">Todas las Entidades</option>
				<option value="employee">Empleados</option>
				<option value="attendance">Asistencias</option>
				<option value="work_center">Centros</option>
				<option value="work_shift">Turnos</option>
				<option value="position">Puestos</option>
				<option value="user">Usuarios</option>
				<option value="report">Reportes</option>
			</select>
		</div>

		<div class="bg-white border border-slate-100 p-4 rounded-sm shadow-sm flex flex-col justify-center">
			<label class="text-[9px] font-black uppercase text-slate-400 mb-1">Acción</label>
			<select
				bind:value={filterAction}
				class="bg-transparent border-none text-xs font-bold focus:ring-0 p-0 text-primary uppercase"
			>
				<option value="all">Todas las Acciones</option>
				<option value="create">Crear</option>
				<option value="update">Actualizar</option>
				<option value="delete">Eliminar</option>
				<option value="bulk_update">Masivo: Update</option>
				<option value="bulk_delete">Masivo: Delete</option>
				<option value="bulk_justify">Masivo: Justificar</option>
			</select>
		</div>

		<div class="bg-white border border-slate-100 p-4 rounded-sm shadow-sm flex flex-col justify-center">
			<label class="text-[9px] font-black uppercase text-slate-400 mb-1">Desde</label>
			<input 
				type="date" 
				bind:value={filterStart}
				class="bg-transparent border-none text-xs font-bold focus:ring-0 p-0 text-primary uppercase"
			/>
		</div>

		<div class="bg-white border border-slate-100 p-4 rounded-sm shadow-sm flex flex-col justify-center relative group">
			<label class="text-[9px] font-black uppercase text-slate-400 mb-1">Hasta</label>
			<div class="flex items-center justify-between">
				<input 
					type="date" 
					bind:value={filterEnd}
					class="bg-transparent border-none text-xs font-bold focus:ring-0 p-0 text-primary uppercase w-full"
				/>
				{#if filterStart || filterEnd || filterAction !== 'all' || filterEntity !== 'all'}
					<button 
						onclick={() => {
							filterAction = 'all';
							filterEntity = 'all';
							filterStart = '';
							filterEnd = '';
						}}
						class="text-[9px] font-black text-rose-500 uppercase hover:underline ml-2"
					>
						Limpiar
					</button>
				{/if}
			</div>
		</div>
	</div>

	<!-- Table Container -->
	<div class="bg-white rounded-md border border-slate-100 shadow-2xl shadow-slate-200/50 overflow-hidden" in:fly={{ y: 20, duration: 800, easing: quintOut, delay: 200 }}>
		<div class="overflow-x-auto">
			<table class="w-full text-left border-collapse">
				<thead>
					<tr class="bg-slate-50/50 border-b border-slate-100">
						<th class="px-6 py-4 text-[10px] font-black text-slate-400 uppercase tracking-[0.2em]">{$_('admin.audit.column_user')}</th>
						<th class="px-6 py-4 text-[10px] font-black text-slate-400 uppercase tracking-[0.2em]">{$_('admin.audit.column_action')}</th>
						<th class="px-6 py-4 text-[10px] font-black text-slate-400 uppercase tracking-[0.2em]">{$_('admin.audit.column_entity')}</th>
						<th class="px-6 py-4 text-[10px] font-black text-slate-400 uppercase tracking-[0.2em]">{$_('admin.audit.column_date')}</th>
						<th class="px-6 py-4 text-[10px] font-black text-slate-400 uppercase tracking-[0.2em]">{$_('admin.audit.column_ip')}</th>
						<th class="px-6 py-4"></th>
					</tr>
				</thead>
				<tbody class="divide-y divide-slate-50">
					{#if loading}
						{#each Array(5) as _}
							<tr class="animate-pulse">
								{#each Array(6) as _}
									<td class="px-6 py-4"><div class="h-4 bg-slate-100 rounded w-full"></div></td>
								{/each}
							</tr>
						{/each}
					{:else if logs.length === 0}
						<tr>
							<td colspan="6" class="px-6 py-20 text-center">
								<History size={40} class="mx-auto text-slate-200 mb-4" />
								<p class="text-slate-400 font-bold uppercase tracking-widest text-xs">{$_('admin.audit.no_logs')}</p>
							</td>
						</tr>
					{:else}
						{#each logs as log (log.id)}
							<tr class="hover:bg-slate-50/80 transition-colors group">
								<td class="px-6 py-4">
									<div class="flex items-center gap-3">
										<div class="h-8 w-8 rounded-full bg-slate-100 flex items-center justify-center border border-slate-200">
											<User size={14} class="text-slate-500" />
										</div>
										<span class="text-xs font-black text-slate-700 tracking-tight">{log.user_name}</span>
									</div>
								</td>
								<td class="px-6 py-4">
									<Badge variant="outline" class="font-black text-[9px] px-2 py-0.5 rounded-sm {getActionColor(log.action)}">
										{log.action.split('_').slice(1).join(' ').toUpperCase()}
									</Badge>
								</td>
								<td class="px-6 py-4">
									<div class="flex flex-col">
										<span class="text-[10px] font-black text-primary uppercase tracking-tighter">{log.entity_type}</span>
										<span class="text-[9px] font-bold text-slate-400 tracking-tight">ID: {log.entity_id || 'N/A'}</span>
									</div>
								</td>
								<td class="px-6 py-4">
									<div class="flex items-center gap-2 text-slate-500">
										<CalendarClock size={12} />
										<span class="text-[10px] font-bold">{new Date(log.created_at).toLocaleString()}</span>
									</div>
								</td>
								<td class="px-6 py-4">
									<div class="flex items-center gap-2 text-slate-400">
										<Globe size={12} />
										<span class="text-[10px] font-mono">{log.ip_address || '0.0.0.0'}</span>
									</div>
								</td>
								<td class="px-6 py-4 text-right">
									<Button 
										variant="ghost" 
										size="sm" 
										class="h-8 px-3 gap-2 text-[10px] font-black uppercase tracking-widest text-slate-400 hover:text-primary transition-all opacity-0 group-hover:opacity-100"
										onclick={() => viewDetails(log)}
									>
										<Eye size={12} />
										{$_('admin.audit.view_details')}
									</Button>
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
				Página {page} de {totalPages}
			</span>
			<div class="flex gap-2">
				<Button 
					variant="outline" 
					size="sm" 
					class="h-9 px-4 rounded-sm font-black text-[10px] uppercase tracking-widest gap-2 disabled:opacity-50"
					disabled={page === 1}
					onclick={() => loadLogs(page - 1)}
				>
					<ChevronLeft size={14} />
					Anterior
				</Button>
				<Button 
					variant="outline" 
					size="sm" 
					class="h-9 px-4 rounded-sm font-black text-[10px] uppercase tracking-widest gap-2 disabled:opacity-50"
					disabled={page >= totalPages}
					onclick={() => loadLogs(page + 1)}
				>
					Siguiente
					<ChevronRight size={14} />
				</Button>
			</div>
		</div>
	</div>
	{/if}
</main>

<!-- Details Modal -->
{#if showDetails}
	<Dialog.Root open={showDetails} onOpenChange={(o) => !o && (showDetails = false)}>
		<Dialog.Content class="sm:max-w-4xl max-h-[90vh] overflow-hidden flex flex-col p-12 border-none rounded-md">
			<Dialog.Header>
				<div class="flex items-center gap-4 mb-6 text-left">
					<div class="h-12 w-12 rounded-md bg-primary/10 text-primary flex items-center justify-center">
						<Shield size={24} />
					</div>
					<div>
						<Dialog.Title class="text-3xl font-black tracking-tighter text-slate-900">
							{$_('admin.audit.view_details')}
						</Dialog.Title>
						<Dialog.Description class="text-xs font-black text-slate-400 uppercase tracking-[0.2em]">
							{selectedLog.action} • {selectedLog.entity_type} #{selectedLog.entity_id || 'N/A'}
						</Dialog.Description>
					</div>
				</div>
			</Dialog.Header>

			<div class="flex-1 overflow-y-auto pr-4 space-y-8 pb-4">
				<div class="grid grid-cols-1 md:grid-cols-2 gap-8">
					<!-- Old State -->
					<div class="space-y-4">
						<div class="flex items-center gap-2 px-1">
							<Badge variant="outline" class="bg-slate-50 text-slate-400 border-slate-200 font-black text-[9px] uppercase">
								{$_('admin.audit.old_value')}
							</Badge>
						</div>
						<div class="bg-slate-950 rounded-md p-6 font-mono text-[11px] text-slate-400 shadow-inner overflow-x-auto min-h-[200px]">
							{#if selectedLog.old_parsed}
								<pre class="whitespace-pre-wrap">{JSON.stringify(selectedLog.old_parsed, null, 2)}</pre>
							{:else}
								<div class="h-full flex items-center justify-center italic opacity-30 text-xs">
									(NULL / EMPTY)
								</div>
							{/if}
						</div>
					</div>

					<!-- New State -->
					<div class="space-y-4">
						<div class="flex items-center gap-2 px-1">
							<Badge variant="outline" class="bg-primary/5 text-primary border-primary/10 font-black text-[9px] uppercase">
								{$_('admin.audit.new_value')}
							</Badge>
						</div>
						<div class="bg-slate-950 rounded-md p-6 font-mono text-[11px] text-emerald-400 shadow-inner overflow-x-auto min-h-[200px]">
							{#if selectedLog.new_parsed}
								<pre class="whitespace-pre-wrap">{JSON.stringify(selectedLog.new_parsed, null, 2)}</pre>
							{:else}
								<div class="h-full flex items-center justify-center italic opacity-30 text-xs text-slate-500">
									(DELETED / NULL)
								</div>
							{/if}
						</div>
					</div>
				</div>

				<!-- Metadata Section -->
				<div class="p-6 bg-slate-50 rounded-md border border-slate-100 flex justify-between items-center">
					<div class="flex gap-12">
						<div class="flex flex-col">
							<span class="text-[9px] font-black text-slate-300 uppercase tracking-widest mb-1">Operador</span>
							<span class="text-xs font-black text-slate-700 uppercase tracking-tight">{selectedLog.user_name}</span>
						</div>
						<div class="flex flex-col">
							<span class="text-[9px] font-black text-slate-300 uppercase tracking-widest mb-1">Terminal</span>
							<div class="flex items-center gap-2">
								<Terminal size={12} class="text-slate-400" />
								<span class="text-xs font-bold text-slate-500 font-mono">{selectedLog.ip_address || '0.0.0.0'}</span>
							</div>
						</div>
					</div>
					<div class="text-right">
						<span class="text-[9px] font-black text-slate-300 uppercase tracking-widest mb-1 block">Estampa de Tiempo</span>
						<span class="text-xs font-black text-primary tracking-tight">{new Date(selectedLog.created_at).toLocaleString()}</span>
					</div>
				</div>
			</div>

			<div class="flex justify-end pt-8 border-t border-slate-100">
				<Button 
					variant="secondary" 
					class="bg-slate-900 text-white hover:bg-slate-800 font-black uppercase text-[10px] tracking-widest h-12 px-8 rounded-sm"
					onclick={() => showDetails = false}
				>
					Cerrar Detalle
				</Button>
			</div>
		</Dialog.Content>
	</Dialog.Root>
{/if}

<style>
	:global(.lucide) {
		stroke-width: 2.5px;
	}
	pre {
		scrollbar-width: thin;
		scrollbar-color: rgba(255,255,255,0.1) transparent;
	}
</style>
