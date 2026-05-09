<script lang="ts">
	import { onMount } from 'svelte';
	import { _ } from 'svelte-i18n';
	import { apiFetch } from '$lib/api';
	import { authState } from '$lib/auth.svelte';
	import type { Employee, WorkCenter, WorkShift, Position, User } from '$lib/types/models';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import {
		UserPlus,
		Pencil,
		Trash2,
		ArrowLeft,
		Search,
		Users,
		UserCheck,
		MoreHorizontal,
		ChevronDown,
		Loader2,
		Building2,
		Clock,
		Filter,
		LayoutGrid,
		BadgeCheck,
		CheckCircle,
		XCircle,
		ArrowRightLeft
	} from 'lucide-svelte';
	import { fade, slide, fly } from 'svelte/transition';
	import { quintOut } from 'svelte/easing';

	let employees = $state<(Employee & { user_name: string; position_name: string; center_name: string; shift_name: string; phone?: string })[]>([]);
	let centers = $state<WorkCenter[]>([]);
	let shifts = $state<WorkShift[]>([]);
	let positions = $state<Position[]>([]);
	let unassignedUsers = $state<User[]>([]);
	let loading = $state(true);
	let showModal = $state(false);
	let editingItem = $state<Employee | null>(null);
	let mounted = $state(false);

	let searchQuery = $state('');
	let filterCenter = $state('all');
	let filterShift = $state('all');
	let filterPosition = $state('all');
	let filterShiftType = $state('all');
	let selectedIds = $state(new Set<number>());
	let selectedEmployeesData = $state<any[]>([]);
	let bulkLoading = $state(false);
	let showBulkTransfer = $state(false);
	let bulkTransferData = $state({
		work_center_id: 0,
		work_shift_id: 0
	});

	import BatchActionBar from '$lib/components/BatchActionBar.svelte';

	let formData = $state({
		user_id: 0,
		work_center_id: 0,
		work_shift_id: 0,
		position_id: 0,
		is_active: true
	});

	const canEdit = $derived(authState.isAdmin);

	async function loadData() {
		loading = true;
		const employeesEndpoint = authState.isManager ? '/manager/employees' : '/admin/employees';
		
		const query = new URLSearchParams({
			search: searchQuery,
			center_id: filterCenter,
			shift_id: filterShift,
			position_id: filterPosition,
			shift_type: filterShiftType
		});

		try {
			const [empRes, centRes, shiftRes, posRes, unassignedRes] = await Promise.all([
				apiFetch<any[]>(`${employeesEndpoint}?${query.toString()}`),
				apiFetch<WorkCenter[]>('/admin/centers'),
				apiFetch<WorkShift[]>('/admin/shifts'),
				apiFetch<Position[]>('/admin/positions'),
				apiFetch<User[]>('/admin/users/unassigned')
			]);

			if (empRes.ok) employees = (await empRes.json()) || [];
			if (centRes.ok) centers = (await centRes.json()) || [];
			if (shiftRes.ok) shifts = (await shiftRes.json()) || [];
			if (posRes.ok) positions = (await posRes.json()) || [];
			if (unassignedRes.ok) unassignedUsers = (await unassignedRes.json()) || [];
		} catch (e) {
			console.error('Error loading data:', e);
		} finally {
			loading = false;
		}
	}

	let debounceTimer: any;
	$effect(() => {
		// Track dependencies
		searchQuery, filterCenter, filterShift;
		
		if (!mounted) return;
		
		clearTimeout(debounceTimer);
		debounceTimer = setTimeout(() => {
			loadData();
		}, 300);
	});

	function openCreate() {
		if (!canEdit) return;
		editingItem = null;
		formData = {
			user_id: 0,
			work_center_id: centers[0]?.id || 0,
			work_shift_id: shifts[0]?.id || 0,
			position_id: positions[0]?.id || 0,
			is_active: true
		};
		showModal = true;
	}

	function openEdit(item: any) {
		if (!canEdit) return;
		editingItem = item;
		formData = { ...item };
		showModal = true;
	}

	async function save() {
		const method = editingItem ? 'PUT' : 'POST';
		const url = editingItem ? `/admin/employees/${editingItem.id}` : '/admin/employees';

		const res = await apiFetch(url, {
			method,
			body: JSON.stringify(formData)
		});

		if (res.ok) {
			showModal = false;
			loadData();
		} else {
			const err = await res.json();
			alert(err.error || $_('common.error_saving'));
		}
	}

	async function remove(item: any) {
		if (!canEdit) return;
		if (
			!confirm($_('admin.employees.confirm_delete_permanent', { values: { name: item.user_name } }))
		)
			return;
		const res = await apiFetch(`/admin/employees/${item.id}`, { method: 'DELETE' });
		if (res.ok) loadData();
	}

	const filteredEmployees = $derived(
		employees.filter((emp) => !selectedIds.has(emp.id))
	);

	function toggleSelect(emp: any) {
		if (selectedIds.has(emp.id)) {
			selectedIds.delete(emp.id);
			selectedEmployeesData = selectedEmployeesData.filter((e) => e.id !== emp.id);
		} else {
			selectedIds.add(emp.id);
			selectedEmployeesData = [...selectedEmployeesData, emp];
		}
		selectedIds = new Set(selectedIds);
	}

	async function bulkUpdate(data: any) {
		bulkLoading = true;
		const res = await apiFetch('/admin/bulk/employees/update', {
			method: 'POST',
			body: JSON.stringify({
				ids: Array.from(selectedIds),
				...data
			})
		});
		if (res.ok) {
			selectedIds = new Set();
			selectedEmployeesData = [];
			loadData();
		}
		bulkLoading = false;
	}

	async function bulkDelete() {
		if (!confirm($_('admin.employees.bulk_delete_confirm', { values: { count: selectedIds.size } }))) return;
		bulkLoading = true;
		const res = await apiFetch('/admin/bulk/employees/delete', {
			method: 'POST',
			body: JSON.stringify({ ids: Array.from(selectedIds) })
		});
		if (res.ok) {
			selectedIds = new Set();
			selectedEmployeesData = [];
			loadData();
		}
		bulkLoading = false;
	}

	onMount(() => {
		loadData();
		mounted = true;
	});
</script>

<div class="min-h-screen pb-24">
	{#if mounted}
		<main class="pt-8 px-6 max-w-5xl mx-auto space-y-8">
			<!-- Asymmetric Header -->
			<section
				class="flex justify-between items-end"
				in:fly={{ y: 20, duration: 800, easing: quintOut }}
			>
				<div class="space-y-1">
					<h2 class="text-5xl font-black text-primary leading-none tracking-tighter">
						{$_('admin.employees.title')}.
					</h2>
					<p class="text-[10px] font-bold text-slate-400 uppercase tracking-widest pt-2 italic">
						{employees.length}
						{$_('admin.employees.registered_agents')}
					</p>
				</div>

				<div class="flex items-center gap-6">
					{#if canEdit}
						<Button
							onclick={openCreate}
							class="h-14 px-8 rounded-sm bg-primary hover:bg-primary/90 text-white font-black text-xs uppercase tracking-[0.2em] shadow-xl shadow-primary/20 transition-all active:scale-95 flex items-center gap-2"
						>
							<UserPlus size={18} strokeWidth={3} />
							{$_('admin.employees.add_personnel')}
						</Button>
					{/if}
					<div class="bg-slate-100 w-1.5 h-16 rounded-full hidden md:block"></div>
				</div>
			</section>

			<!-- Search & Filters Bar -->
			<div class="space-y-4" in:fly={{ y: 20, duration: 800, easing: quintOut, delay: 100 }}>
				<div class="relative group">
					<div
						class="absolute inset-y-0 left-5 flex items-center pointer-events-none text-slate-300 group-focus-within:text-primary"
					>
						<Search size={20} strokeWidth={3} />
					</div>
					<input
						type="text"
						placeholder={$_('admin.employees.search_placeholder')}
						bind:value={searchQuery}
						class="w-full h-16 pl-14 pr-6 bg-white rounded-sm border border-slate-100 shadow-sm focus:border-primary focus:ring-4 focus:ring-primary/5 text-sm font-black tracking-tight placeholder:text-slate-300 uppercase transition-all"
					/>
					{#if loading}
						<div class="absolute right-5 top-1/2 -translate-y-1/2">
							<Loader2 size={20} class="animate-spin text-primary" />
						</div>
					{/if}
				</div>

				<div class="grid grid-cols-1 md:grid-cols-4 gap-4">
					<div class="relative">
							<select
							bind:value={filterCenter}
							class="appearance-none w-full bg-white border border-slate-100 rounded-sm py-4 px-6 text-xs font-black uppercase tracking-widest text-primary focus:ring-4 focus:ring-primary/5 cursor-pointer transition-all outline-none"
						>
							<option value="all">{$_('common.all_centers')}</option>
							{#each centers as c}
								<option value={c.id.toString()}>{c.name}</option>
							{/each}
						</select>
						<ChevronDown class="absolute right-5 top-1/2 -translate-y-1/2 text-slate-300 pointer-events-none" size={16} />
					</div>
					<div class="relative">
						<select
							bind:value={filterShift}
							class="appearance-none w-full bg-white border border-slate-100 rounded-sm py-4 px-6 text-xs font-black uppercase tracking-widest text-primary focus:ring-4 focus:ring-primary/5 cursor-pointer transition-all outline-none"
						>
							<option value="all">{$_('common.all_shifts')}</option>
							{#each shifts as s}
								<option value={s.id.toString()}>{s.name}</option>
							{/each}
						</select>
						<ChevronDown class="absolute right-5 top-1/2 -translate-y-1/2 text-slate-300 pointer-events-none" size={16} />
					</div>
					<div class="relative">
						<select
							bind:value={filterPosition}
							class="appearance-none w-full bg-white border border-slate-100 rounded-sm py-4 px-6 text-xs font-black uppercase tracking-widest text-primary focus:ring-4 focus:ring-primary/5 cursor-pointer transition-all outline-none"
						>
							<option value="all">{$_('common.all_positions')}</option>
							{#each positions as p}
								<option value={p.id.toString()}>{p.name}</option>
							{/each}
						</select>
						<ChevronDown class="absolute right-5 top-1/2 -translate-y-1/2 text-slate-300 pointer-events-none" size={16} />
					</div>
					<div class="relative">
						<select
							bind:value={filterShiftType}
							class="appearance-none w-full bg-white border border-slate-100 rounded-sm py-4 px-6 text-xs font-black uppercase tracking-widest text-primary focus:ring-4 focus:ring-primary/5 cursor-pointer transition-all outline-none"
						>
							<option value="all">{$_('common.type')}</option>
							<option value="fixed">{$_('common.fixed')}</option>
							<option value="flexible">{$_('common.flexible')}</option>
							<option value="field">{$_('common.field')}</option>
						</select>
						<ChevronDown class="absolute right-5 top-1/2 -translate-y-1/2 text-slate-300 pointer-events-none" size={16} />
					</div>
				</div>
			</div>

			<!-- Selected Container (The "Bucket") -->
			{#if selectedEmployeesData.length > 0}
				<section class="space-y-6 pt-4" in:slide>
					<div class="flex items-center gap-4">
						<div class="h-px flex-1 bg-slate-100"></div>
						<h3 class="text-[10px] font-black text-slate-400 uppercase tracking-[0.2em] flex items-center gap-3">
							<Users size={14} />
							{$_('common.items_selected')} ({selectedEmployeesData.length})
						</h3>
						<div class="h-px flex-1 bg-slate-100"></div>
					</div>
					<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
						{#each selectedEmployeesData as emp (emp.id)}
							<div animate:flip={{ duration: 400 }}>
								{@render employeeCard(emp)}
							</div>
						{/each}
					</div>
				</section>
			{/if}

			<div
				class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6"
				in:fly={{ y: 20, duration: 800, easing: quintOut, delay: 200 }}
			>
				{#if loading && employees.length === 0}
					{#each Array(6) as _}
						<div
							class="h-24 bg-white rounded-sm animate-pulse shadow-sm border border-slate-50"
						></div>
					{/each}
				{:else if filteredEmployees.length === 0}
					<div
						class="col-span-full py-24 text-center space-y-4 bg-white rounded-sm border-2 border-dashed border-slate-100"
					>
						<div class="w-16 h-16 bg-slate-50 rounded-full flex items-center justify-center mx-auto">
							<Search size={32} class="text-slate-200" />
						</div>
						<div class="space-y-1">
							<p class="text-slate-400 font-black uppercase tracking-widest text-xs">
								{$_('admin.employees.no_results')}
							</p>
							<p class="text-[10px] text-slate-300 font-medium">
								{$_('common.try_adjusting_filters')}
							</p>
						</div>
					</div>
				{:else}
					{#each filteredEmployees as emp (emp.id)}
						<div animate:flip={{ duration: 400 }}>
							{@render employeeCard(emp)}
						</div>
					{/each}
				{/if}
			</div>
		</main>

		{#snippet employeeCard(emp)}
			<div
				class="flex items-center justify-between p-4 bg-white rounded-sm border border-slate-100 shadow-sm hover:shadow-md transition-all group relative {selectedIds.has(emp.id) ? 'ring-2 ring-primary bg-primary/5' : ''}"
			>
				<!-- Bulk Checkbox -->
				<div class="absolute -top-2 -left-2 z-20 transition-transform hover:scale-110">
					<input 
						type="checkbox" 
						checked={selectedIds.has(emp.id)}
						onchange={() => toggleSelect(emp)}
						class="h-6 w-6 rounded-lg border-2 border-slate-200 text-primary focus:ring-primary cursor-pointer shadow-lg bg-white checked:bg-primary"
					/>
				</div>

				<div class="flex items-center gap-4">
					<div
						class="w-12 h-12 rounded-sm bg-slate-50 flex items-center justify-center text-primary font-black text-lg border border-slate-100 group-hover:bg-primary group-hover:text-white transition-all duration-500 overflow-hidden"
					>
						{#if emp.photo_url}
							<img 
								src={emp.photo_url} 
								alt={emp.user_name} 
								class="w-full h-full object-cover"
							/>
						{:else}
							{emp.user_name
								.split(' ')
								.map((n) => n[0])
								.join('')}
						{/if}
					</div>
					<a href="/admin/employees/{emp.id}" class="flex flex-col">
						<h3 class="text-base font-bold text-primary leading-tight">{emp.user_name}</h3>
						<div class="flex flex-col">
							<span class="text-[10px] font-bold text-slate-400 uppercase tracking-tight"
								>{emp.position_name}</span
							>
							{#if emp.phone}
								<span
									class="text-[9px] font-black text-primary/40 flex items-center gap-1 uppercase tracking-tighter"
								>
									<div class="w-1 h-1 bg-primary/20 rounded-full"></div>
									{emp.phone}
								</span>
							{/if}
							<span class="text-[10px] font-medium text-slate-400 italic"
								>{emp.shift_name || $_('admin.employees.no_shift')}</span
							>
						</div>
						<span class="text-[10px] font-medium text-slate-400 italic"
							>{emp.center_name}</span
						>
					</a>
				</div>

				<div class="flex items-center gap-3">
					<!-- Status Indicator -->
					{#if emp.is_active}
						<span
							class="bg-emerald-50 text-emerald-600 px-3 py-1 rounded-full text-[9px] font-black uppercase tracking-wider"
						>
							{$_('common.active')}
						</span>
					{:else}
						<span
							class="bg-rose-50 text-rose-600 px-3 py-1 rounded-full text-[9px] font-black uppercase tracking-wider"
						>
							{$_('common.inactive')}
						</span>
					{/if}

					<!-- Action Menu -->
					<DropdownMenu.Root>
						<DropdownMenu.Trigger
							class="p-2 text-slate-300 hover:text-primary transition-all focus:outline-none"
						>
							<MoreHorizontal size={20} />
						</DropdownMenu.Trigger>
						<DropdownMenu.Content
							class="bg-white border-none shadow-premium rounded-sm p-2 min-w-[180px]"
						>
							<DropdownMenu.Item
								class="flex items-center gap-3 px-4 py-3 text-[10px] font-black uppercase tracking-widest text-slate-500 hover:text-primary hover:bg-slate-50 rounded-sm cursor-pointer"
								onSelect={() => openEdit(emp)}
							>
								<Pencil size={14} />
								{$_('admin.employees.config_record')}
							</DropdownMenu.Item>
							{#if canEdit}
								<DropdownMenu.Item
									class="flex items-center gap-3 px-4 py-3 text-[10px] font-black uppercase tracking-widest text-rose-500 hover:bg-rose-50 rounded-sm cursor-pointer"
									onSelect={() => remove(emp)}
								>
									<Trash2 size={14} />
									{$_('admin.employees.de_authorize')}
								</DropdownMenu.Item>
							{/if}
						</DropdownMenu.Content>
					</DropdownMenu.Root>
				</div>
			</div>
		{/snippet}


		<!-- Batch Actions Hub -->
		<BatchActionBar 
			selectedCount={selectedIds.size} 
			onClear={() => {
				selectedIds = new Set();
				selectedEmployeesData = [];
			}}
		>
			<div class="flex items-center gap-2">
				<Button 
					size="sm" 
					onclick={() => bulkUpdate({ is_active: true })}
					disabled={bulkLoading}
					class="bg-emerald-600 hover:bg-emerald-700 text-white rounded-xl h-11 px-6 flex items-center gap-2 shadow-lg shadow-emerald-900/10"
				>
					<CheckCircle size={16} /> 
					<span class="hidden md:inline">{$_('common.activate')}</span>
				</Button>
				
				<Button 
					size="sm" 
					onclick={() => bulkUpdate({ is_active: false })}
					disabled={bulkLoading}
					class="bg-amber-600 hover:bg-amber-700 text-white rounded-xl h-11 px-6 flex items-center gap-2 shadow-lg shadow-amber-900/10"
				>
					<XCircle size={16} /> 
					<span class="hidden md:inline">{$_('common.deactivate')}</span>
				</Button>

				<Button 
					size="sm" 
					onclick={() => {
						bulkTransferData = { work_center_id: centers[0]?.id || 0, work_shift_id: shifts[0]?.id || 0 };
						showBulkTransfer = true;
					}}
					disabled={bulkLoading}
					class="bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl h-11 px-6 flex items-center gap-2 shadow-lg shadow-indigo-900/10"
				>
					<ArrowRightLeft size={16} /> 
					<span class="hidden md:inline">{$_('admin.employees.reassign')}</span>
				</Button>

				<Button 
					size="sm" 
					onclick={bulkDelete}
					disabled={bulkLoading}
					class="bg-rose-600 hover:bg-rose-700 text-white rounded-xl h-11 px-6 flex items-center gap-2 shadow-lg shadow-rose-900/10"
				>
					<Trash2 size={16} /> 
					<span class="hidden md:inline">{$_('common.delete')}</span>
				</Button>
			</div>
		</BatchActionBar>

		<!-- Create/Edit Dialog (Architecture Ledger Style) -->
		{#if showModal}
			<Dialog.Root
				open={showModal}
				onOpenChange={(o) => {
					if (!o) showModal = false;
				}}
			>
				<Dialog.Content
					class="bg-white border-none shadow-premium p-10 sm:w-full md:max-w-4xl rounded-sm"
				>
					<Dialog.Header class="space-y-6">
						<div
							class="w-14 h-14 bg-primary/10 text-primary rounded-sm flex items-center justify-center rotate-3 shadow-lg shadow-primary/5"
						>
							<LayoutGrid size={28} />
						</div>
						<div class="space-y-1">
							<Dialog.Title class="text-4xl font-black tracking-tighter text-primary">
								{editingItem
									? $_('admin.employees.edit_title')
									: $_('admin.employees.create_title')}
								<span class="italic opacity-50">Identity.</span>
							</Dialog.Title>
							<Dialog.Description
								class="text-[10px] font-black uppercase tracking-widest text-slate-400"
							>
								{$_('admin.employees.description')}
							</Dialog.Description>
						</div>
					</Dialog.Header>

					<form
						onsubmit={(e) => {
							e.preventDefault();
							save();
						}}
						class="pt-8 space-y-8"
					>
						<div class="grid grid-cols-1 md:grid-cols-2 gap-8">
							<!-- Identity Module -->
							<div class="space-y-6">
								<div class="space-y-2">
									<Label class="text-[10px] font-black uppercase tracking-widest text-primary ml-1"
										>{$_('admin.employees.credential_access')}</Label
									>
									{#if editingItem}
										<div
											class="h-14 w-full bg-slate-100/50 border-2 border-slate-50 flex items-center px-5 rounded-sm font-black text-slate-400 text-xs"
										>
											{editingItem.user_name}
										</div>
									{:else}
										<select
											bind:value={formData.user_id}
											class="h-14 w-full bg-slate-50 border-none rounded-sm font-black text-xs px-5 focus:ring-4 focus:ring-primary/5 transition-all appearance-none cursor-pointer outline-none uppercase tracking-tight"
										>
											<option value={0}>{$_('admin.employees.match_user')}</option>
											{#each unassignedUsers as user}
												<option value={user.id}>{user.name}</option>
											{/each}
										</select>
									{/if}
								</div>

								<div class="space-y-2">
									<Label class="text-[10px] font-black uppercase tracking-widest text-primary ml-1"
										>{$_('admin.employees.designated_position')}</Label
									>
									<select
										bind:value={formData.position_id}
										class="h-14 w-full bg-slate-50 border-none rounded-sm font-black text-xs px-5 focus:ring-4 focus:ring-primary/5 transition-all appearance-none cursor-pointer outline-none uppercase tracking-tight"
									>
										{#each positions as pos}
											<option value={pos.id}>{pos.name}</option>
										{/each}
									</select>
								</div>
							</div>

							<!-- Logistic Module -->
							<div class="space-y-6">
								<div class="space-y-2">
									<Label class="text-[10px] font-black uppercase tracking-widest text-primary ml-1"
										>{$_('admin.employees.station_facility')}</Label
									>
									<select
										bind:value={formData.work_center_id}
										class="h-14 w-full bg-slate-50 border-none rounded-sm font-black text-xs px-5 focus:ring-4 focus:ring-primary/5 transition-all appearance-none cursor-pointer outline-none uppercase tracking-tight"
									>
										{#each centers as center}
											<option value={center.id}>{center.name}</option>
										{/each}
									</select>
								</div>

								<div class="space-y-2">
									<Label class="text-[10px] font-black uppercase tracking-widest text-primary ml-1"
										>{$_('admin.employees.operative_cycle')}</Label
									>
									<select
										bind:value={formData.work_shift_id}
										class="h-14 w-full bg-slate-50 border-none rounded-sm font-black text-xs px-5 focus:ring-4 focus:ring-primary/5 transition-all appearance-none cursor-pointer outline-none uppercase tracking-tight"
									>
										<option value={0}>{$_('admin.employees.manual_override')}</option>
										{#each shifts as shift}
											<option value={shift.id}>{shift.name}</option>
										{/each}
									</select>
								</div>
							</div>
						</div>

						<!-- Status Ledger -->
						<div
							class="p-6 bg-slate-900 rounded-sm flex items-center justify-between border-4 border-slate-800 shadow-2xl"
						>
							<div class="flex items-center gap-3 text-white">
								<BadgeCheck size={24} class="text-primary" />
								<div class="flex flex-col">
									<span class="text-[10px] font-black uppercase tracking-[0.2em] leading-none"
										>{$_('admin.employees.operative_status')}</span
									>
									<span class="text-[9px] font-medium text-slate-500 italic pt-1 uppercase"
										>{$_('admin.employees.status_hint')}</span
									>
								</div>
							</div>
							<div class="flex gap-2">
								<button
									type="button"
									onclick={() => (formData.is_active = true)}
									class="px-5 py-2 rounded-sm text-[10px] font-black tracking-[.2em] transition-all {formData.is_active
										? 'bg-primary text-white shadow-xl'
										: 'bg-slate-800 text-slate-500 hover:text-white'}"
								>
									{$_('common.active')}
								</button>
								<button
									type="button"
									onclick={() => (formData.is_active = false)}
									class="px-5 py-2 rounded-sm text-[10px] font-black tracking-[.2em] transition-all {!formData.is_active
										? 'bg-rose-500 text-white shadow-xl'
										: 'bg-slate-800 text-slate-500 hover:text-white'}"
								>
									{$_('common.disabled')}
								</button>
							</div>
						</div>

						<div class="flex flex-col gap-3 pt-4">
							<Button
								type="submit"
								class="h-16 rounded-sm font-black text-xs uppercase tracking-widest shadow-2xl shadow-primary/20 transition-all hover:scale-[1.01] active:scale-95 text-white"
								disabled={loading}
							>
								{#if loading}
									<Loader2 class="h-4 w-4 animate-spin mr-2" />
									{$_('admin.employees.syncing_identities')}
								{:else}
									{editingItem
										? $_('admin.employees.update_file')
										: $_('admin.employees.authorize_agent')}
								{/if}
							</Button>
							<button
								type="button"
								class="h-12 rounded-sm font-black text-[9px] uppercase tracking-[0.4em] text-slate-300 hover:text-primary transition-all underline underline-offset-8"
								onclick={() => (showModal = false)}
							>
								{$_('common.cancel')}
							</button>
						</div>
					</form>
				</Dialog.Content>
			</Dialog.Root>
		{/if}

		<!-- Bulk Transfer Dialog -->
		{#if showBulkTransfer}
			<Dialog.Root
				open={showBulkTransfer}
				onOpenChange={(o) => {
					if (!o) showBulkTransfer = false;
				}}
			>
				<Dialog.Content
					class="bg-white border-none shadow-premium p-10 sm:w-full md:max-w-2xl rounded-sm"
				>
					<Dialog.Header class="space-y-4">
						<div class="w-12 h-12 bg-indigo-50 text-indigo-600 rounded-xl flex items-center justify-center shadow-lg shadow-indigo-900/5">
							<ArrowRightLeft size={24} />
						</div>
						<div class="space-y-1">
							<Dialog.Title class="text-3xl font-black tracking-tighter text-primary">
								{$_('admin.employees.bulk_reassign_title')}
							</Dialog.Title>
							<Dialog.Description class="text-[10px] font-black uppercase tracking-widest text-slate-400">
								{$_('admin.employees.bulk_reassign_desc', { values: { count: selectedIds.size } })}
							</Dialog.Description>
						</div>
					</Dialog.Header>

					<div class="grid grid-cols-1 gap-6 pt-6">
						<div class="space-y-2">
							<Label class="text-[10px] font-black uppercase tracking-widest text-primary ml-1">{$_('common.center')}</Label>
							<select
								bind:value={bulkTransferData.work_center_id}
								class="h-14 w-full bg-slate-50 border-none rounded-sm font-black text-xs px-5 focus:ring-4 focus:ring-primary/5 transition-all outline-none"
							>
								{#each centers as c}
									<option value={c.id}>{c.name}</option>
								{/each}
							</select>
						</div>

						<div class="space-y-2">
							<Label class="text-[10px] font-black uppercase tracking-widest text-primary ml-1">{$_('common.shift')}</Label>
							<select
								bind:value={bulkTransferData.work_shift_id}
								class="h-14 w-full bg-slate-50 border-none rounded-sm font-black text-xs px-5 focus:ring-4 focus:ring-primary/5 transition-all outline-none"
							>
								<option value={0}>{$_('common.no_change')}</option>
								{#each shifts as s}
									<option value={s.id}>{s.name}</option>
								{/each}
							</select>
						</div>
					</div>

					<div class="flex flex-col gap-3 pt-8">
						<Button
							onclick={() => {
								bulkUpdate(bulkTransferData);
								showBulkTransfer = false;
							}}
							class="h-14 rounded-sm font-black text-xs uppercase tracking-widest shadow-2xl shadow-indigo-900/20 text-white bg-indigo-600 hover:bg-indigo-700"
							disabled={bulkLoading}
						>
							{$_('admin.employees.apply_bulk_changes')}
						</Button>
						<Button variant="ghost" onclick={() => showBulkTransfer = false} class="text-[10px] font-black uppercase tracking-widest text-slate-400">
							{$_('common.cancel')}
						</Button>
					</div>
				</Dialog.Content>
			</Dialog.Root>
		{/if}
	{/if}
</div>
