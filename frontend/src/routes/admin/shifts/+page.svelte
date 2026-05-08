<script lang="ts">
	import { onMount } from 'svelte';
	import { fly } from 'svelte/transition';
	import { quintOut } from 'svelte/easing';
	import { _ } from 'svelte-i18n';
	import { apiFetch } from '$lib/api';
	import { authState } from '$lib/auth.svelte';
	import type { WorkShift } from '$lib/types/models';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import {
		Clock,
		Pencil,
		Trash2,
		Plus,
		Timer,
		Coffee,
		Sun,
		Moon,
		Sunset,
		Search,
		MoreHorizontal,
		Loader2,
		CalendarDays,
		ArrowUpRight,
		ArrowDownRight,
		Settings2,
		CheckCircle2,
		XCircle,
		LayoutDashboard,
		ShieldCheck,
		Activity,
		MapPin
	} from 'lucide-svelte';

	let shifts = $state<WorkShift[]>([]);
	let loading = $state(true);
	let showModal = $state(false);
	let editingItem = $state<WorkShift | null>(null);
	let mounted = $state(false);
	let searchQuery = $state('');

	let formData = $state({
		name: '',
		start_time: '08:00',
		end_time: '17:00',
		grace_period: 15,
		lunch_duration_limit: 60,
		is_active: true,
		enforce_lateness: true,
		enforce_lunch_limit: true,
		enforce_geofence: true,
		work_days: [1, 2, 3, 4, 5]
	});

	const canEdit = $derived(authState.isAdmin);

	const shiftTypeOptions = [
		{ id: 'fixed', label: $_('admin.shifts.type_fixed'), hint: $_('admin.shifts.hint_fixed'), icon: ShieldCheck },
		{ id: 'flexible', label: $_('admin.shifts.type_flexible'), hint: $_('admin.shifts.hint_flexible'), icon: Activity },
		{ id: 'field', label: $_('admin.shifts.type_field'), hint: $_('admin.shifts.hint_field'), icon: MapPin }
	];

	function selectShiftType(type: string) {
		formData.shift_type = type;
		// Auto-set logical defaults
		if (type === 'fixed') {
			formData.enforce_lateness = true;
			formData.enforce_geofence = true;
		} else if (type === 'flexible') {
			formData.enforce_lateness = false;
			formData.enforce_geofence = true;
		} else if (type === 'field') {
			formData.enforce_lateness = false;
			formData.enforce_geofence = false;
		}
	}

	async function loadShifts() {
		loading = true;
		try {
			const res = await apiFetch<WorkShift[]>('/admin/shifts');
			if (res.ok) shifts = await res.json();
		} finally {
			loading = false;
		}
	}

	function openCreate() {
		if (!canEdit) return;
		editingItem = null;
		formData = {
			name: '',
			start_time: '08:00',
			end_time: '17:00',
			grace_period: 15,
			lunch_duration_limit: 60,
			is_active: true,
			enforce_lateness: true,
			enforce_lunch_limit: true,
			enforce_geofence: true,
			shift_type: 'fixed',
			work_days: [1, 2, 3, 4, 5]
		};
		showModal = true;
	}

	function openEdit(item: any) {
		if (!canEdit) return;
		editingItem = item;

		const cleanTime = (val: string) => {
			if (!val) return '';
			if (val.includes('T')) return val.split('T')[1].substring(0, 5);
			return val.substring(0, 5);
		};

		formData = {
			...item,
			start_time: cleanTime(item.start_time),
			end_time: cleanTime(item.end_time),
			grace_period: parseInt(formatMinutes(item.grace_period)),
			lunch_duration_limit: parseInt(formatMinutes(item.lunch_duration_limit)),
			is_active: item.is_active ?? true,
			enforce_lateness: item.enforce_lateness ?? true,
			enforce_lunch_limit: item.enforce_lunch_limit ?? true,
			enforce_geofence: item.enforce_geofence ?? true,
			shift_type: item.shift_type || 'fixed',
			work_days: item.work_days || []
		};
		showModal = true;
	}

	async function save() {
		if (formData.work_days.length === 0) {
			alert($_('admin.shifts.min_one_day_error'));
			return;
		}

		loading = true;
		const method = editingItem ? 'PUT' : 'POST';
		const url = editingItem ? `/admin/shifts/${editingItem.id}` : '/admin/shifts';

		// Helper to format minutes as HH:MM:SS for the backend string expectation
		const toTimeString = (min: number) => {
			const h = Math.floor(min / 60);
			const m = min % 60;
			return `${h.toString().padStart(2, '0')}:${m.toString().padStart(2, '0')}:00`;
		};

		// Explicit payload mapping to ensure correct JSON tags for the backend
		const payload = {
			name: formData.name,
			start_time: formData.start_time,
			end_time: formData.end_time,
			grace_period: toTimeString(formData.grace_period),
			lunch_duration_limit: toTimeString(formData.lunch_duration_limit),
			is_night_shift: formData.is_night_shift,
			is_active: formData.is_active,
			enforce_lateness: formData.enforce_lateness,
			enforce_lunch_limit: formData.enforce_lunch_limit,
			enforce_geofence: formData.enforce_geofence,
			shift_type: formData.shift_type,
			work_days: formData.work_days
		};

		try {
			const res = await apiFetch(url, {
				method,
				body: JSON.stringify(payload)
			});

			if (res.ok) {
				showModal = false;
				loadShifts();
			} else {
				const err = await res.json();
				alert(err.error || $_('admin.shifts.save_error'));
			}
		} finally {
			loading = false;
		}
	}

	async function remove(item: any) {
		if (!canEdit) return;
		if (!confirm(`${$_('common.confirm_delete')} ${item.name}?`)) return;
		const res = await apiFetch(`/admin/shifts/${item.id}`, { method: 'DELETE' });
		if (res.ok) loadShifts();
	}

	async function toggleStatus(item: any) {
		if (!canEdit) return;
		const res = await apiFetch(`/admin/shifts/${item.id}`, {
			method: 'PUT',
			body: JSON.stringify({
				...item,
				is_active: !item.is_active,
				start_time: formatTime(item.start_time),
				end_time: formatTime(item.end_time),
				lunch_duration_limit: formatTime(item.lunch_duration_limit),
				grace_period: formatTime(item.grace_period)
			})
		});
		if (res.ok) loadShifts();
	}

	const formatTime = (val: string) => {
		if (!val) return '00:00';
		if (val.includes('T')) {
			return val.split('T')[1].substring(0, 5);
		}
		return val.substring(0, 5);
	};

	const weekDays = $derived([
		{ id: 1, name: $_('common.days.1'), short: $_('common.days.mon_short') },
		{ id: 2, name: $_('common.days.2'), short: $_('common.days.tue_short') },
		{ id: 3, name: $_('common.days.3'), short: $_('common.days.wed_short') },
		{ id: 4, name: $_('common.days.4'), short: $_('common.days.thu_short') },
		{ id: 5, name: $_('common.days.5'), short: $_('common.days.fri_short') },
		{ id: 6, name: $_('common.days.6'), short: $_('common.days.sat_short') },
		{ id: 0, name: $_('common.days.0'), short: $_('common.days.sun_short') }
	]);

	function toggleDay(dayId: number) {
		if (formData.work_days.includes(dayId)) {
			formData.work_days = formData.work_days.filter((d) => d !== dayId);
		} else {
			formData.work_days = [...formData.work_days, dayId].sort((a, b) => {
				const sortValue = (x: number) => (x === 0 ? 7 : x);
				return sortValue(a) - sortValue(b);
			});
		}
	}

	function formatWorkDays(days: number[]) {
		if (!days || days.length === 0) return $_('admin.shifts.no_days');
		if (days.length === 7) return $_('admin.shifts.all_week');

		const sorted = [...days].sort((a, b) => {
			const sortValue = (x: number) => (x === 0 ? 7 : x);
			return sortValue(a) - sortValue(b);
		});

		// Basic consecutive check for "Lun - Vie"
		if (sorted.length === 5 && sorted[0] === 1 && sorted[4] === 5)
			return `${$_('common.days.mon_short')} - ${$_('common.days.fri_short')}`;

		return sorted.map((d) => weekDays.find((wd) => wd.id === d)?.short).join(', ');
	}

	const formatMinutes = (val: string | number) => {
		if (val === undefined || val === null) return '0';
		if (typeof val === 'number') return val.toString();

		let timeStr = val;
		if (val.includes('T')) {
			timeStr = val.split('T')[1];
		}
		const parts = timeStr.split(':');
		if (parts.length < 2) return '0';
		const [h, m] = parts;
		const total = parseInt(h) * 60 + parseInt(m);
		return total.toString();
	};

	function getHourFromTime(val: string) {
		if (!val) return 0;
		if (val.includes('T')) {
			return parseInt(val.split('T')[1].split(':')[0], 10);
		}
		return parseInt(val.split(':')[0], 10);
	}

	function getShiftIcon(startTime: string) {
		const hour = getHourFromTime(startTime);
		if (hour >= 6 && hour < 12) return Sun;
		if (hour >= 12 && hour < 18) return Sunset;
		return Moon;
	}

	function getShiftColor(startTime: string) {
		const hour = getHourFromTime(startTime);
		if (hour >= 6 && hour < 12) return 'text-[#b77000]';
		if (hour >= 12 && hour < 18) return 'text-[#a13b00]';
		return 'text-primary';
	}

	const filteredShifts = $derived(
		shifts.filter((s) => s.name.toLowerCase().includes(searchQuery.toLowerCase()))
	);

	const activeCount = $derived(shifts.filter((s) => s.is_active).length);

	onMount(() => {
		loadShifts();
		mounted = true;
	});
</script>

<div class="min-h-screen pb-24">
	{#if mounted}
		<main class="pt-8 px-6 max-w-5xl mx-auto">
			<header
				class="mb-12 flex justify-between items-end"
				in:fly={{ y: 20, duration: 800, easing: quintOut }}
			>
				<div class="max-w-md">
					<h2 class="text-5xl font-black text-primary leading-none tracking-tighter">
						{$_('admin.shifts.title')}
					</h2>
					<p class="text-[10px] font-bold text-slate-400 uppercase tracking-widest pt-2 italic">
						{activeCount.toString().padStart(1, '0')}
						{$_('admin.shifts.detected_shifts')}
					</p>
				</div>
				<div class="flex items-center gap-6">
					{#if canEdit}
						<Button
							onclick={openCreate}
							class="h-14 px-8 rounded-sm bg-primary hover:bg-primary/90 text-white font-black text-xs uppercase tracking-[0.2em] shadow-xl shadow-primary/20 transition-all active:scale-95 flex items-center gap-2"
						>
							<Plus size={18} strokeWidth={3} />
							{$_('admin.shifts.register_button')}
						</Button>
					{/if}
					<div class="bg-slate-100 w-1.5 h-16 rounded-full hidden md:block"></div>
				</div>
			</header>

			<!-- Search Bar -->
			<div
				class="mb-8 flex items-center bg-white border-b-2 border-slate-200 focus-within:border-primary transition-all p-4 rounded-t-2xl shadow-sm"
				in:fly={{ y: 20, duration: 800, easing: quintOut, delay: 100 }}
			>
				<Search class="text-slate-300 mr-3" size={20} />
				<input
					type="text"
					placeholder={$_('admin.shifts.search_placeholder')}
					bind:value={searchQuery}
					class="bg-transparent border-none focus:ring-0 w-full text-sm font-bold text-slate-900 placeholder:text-slate-300 uppercase tracking-tight"
				/>
			</div>

			<!-- Shifts Vertical Ledger -->
			<div class="space-y-3" in:fly={{ y: 20, duration: 800, easing: quintOut, delay: 200 }}>
				{#if loading}
					{#each Array(3) as _}
						<div class="h-28 w-full bg-surface-container-lowest rounded-sm animate-pulse"></div>
					{/each}
				{:else}
					{#each filteredShifts as shift (shift.id)}
						{@const Icon = getShiftIcon(shift.start_time)}
						<div
							class="group relative flex flex-col md:flex-row items-start md:items-center justify-between p-6 bg-surface-container-lowest rounded-sm hover:bg-surface-container-low transition-all duration-300 gap-4"
						>
							<div class="flex items-center gap-6">
								<div
									class="flex flex-col items-center justify-center w-12 h-12 bg-primary-container/10 rounded-lg"
								>
									<Icon class="h-6 w-6 {getShiftColor(shift.start_time)}" />
								</div>
								<div>
									<div class="flex items-center gap-2">
										<h3 class=" font-bold text-lg text-primary tracking-tight">
											<a href="/admin/shifts/{shift.id}">{shift.name}</a>
										</h3>
										{#if shift.shift_type === 'field'}
											<span class="text-[9px] font-black bg-amber-500 text-white px-2 py-0.5 rounded-sm uppercase tracking-widest flex items-center gap-1">
												<MapPin size={10} />
												{$_('admin.shifts.type_field')}
											</span>
										{:else if shift.shift_type === 'flexible'}
											<span class="text-[9px] font-black bg-indigo-500 text-white px-2 py-0.5 rounded-sm uppercase tracking-widest flex items-center gap-1">
												<Activity size={10} />
												{$_('admin.shifts.type_flexible')}
											</span>
										{:else}
											<span class="text-[9px] font-black bg-slate-900 text-white px-2 py-0.5 rounded-sm uppercase tracking-widest flex items-center gap-1">
												<ShieldCheck size={10} />
												{$_('admin.shifts.type_fixed')}
											</span>
										{/if}
									</div>
									<div class="flex items-center gap-3 mt-1 flex-wrap">
										<span class=" text-sm font-medium text-on-surface-variant"
											>{formatTime(shift.start_time)} - {formatTime(shift.end_time)}</span
										>
										<span class="w-1 h-1 bg-outline-variant rounded-full hidden sm:block"></span>
										<span
											class=" text-[11px] uppercase tracking-widest text-on-surface-variant font-semibold"
											>{$_('admin.shifts.tolerance_short')}: {formatMinutes(
												shift.grace_period
											)}min</span
										>
										<span class="w-1 h-1 bg-outline-variant rounded-full hidden sm:block"></span>
										<span
											class=" text-[11px] uppercase tracking-widest text-on-surface-variant font-semibold flex items-center gap-1"
											>{$_('admin.shifts.lunch_short')}: {formatMinutes(
												shift.lunch_duration_limit
											)}m</span
										>
										<span class="w-1 h-1 bg-outline-variant rounded-full hidden sm:block"></span>
										<span
											class=" text-[11px] uppercase tracking-widest text-primary font-bold flex items-center gap-1"
										>
											<CalendarDays size={12} />
											{formatWorkDays(shift.work_days)}
										</span>
										<span class="w-1 h-1 bg-outline-variant rounded-full hidden sm:block"></span>
										{#if shift.enforce_lateness && shift.enforce_lunch_limit && shift.enforce_geofence}
											<span
												class="text-[10px] font-black bg-primary/10 text-primary px-2 py-0.5 rounded-sm uppercase tracking-widest"
											>
												{$_('admin.shifts.strict_policy')}
											</span>
										{:else}
											<span
												class="text-[10px] font-black bg-orange-500/10 text-orange-600 px-2 py-0.5 rounded-sm uppercase tracking-widest"
											>
												{$_('admin.shifts.flexible_policy')}
											</span>
										{/if}
									</div>
								</div>
							</div>

							<div class="flex items-center gap-4 self-end md:self-auto">
								{#if shift.is_active}
									<div class="px-3 py-1 bg-green-500/25 rounded-full">
										<span class="text-[10px] font-bold text-green-500 uppercase tracking-widest"
											>{$_('common.active')}</span
										>
									</div>
								{:else}
									<div class="px-3 py-1 bg-red-500/25 rounded-full">
										<span class="text-[10px] font-bold text-red-500 uppercase tracking-widest"
											>{$_('common.inactive')}</span
										>
									</div>
								{/if}

								<DropdownMenu.Root>
									<DropdownMenu.Trigger
										class="text-on-surface-variant hover:text-primary transition-colors p-2 rounded-full hover:bg-surface-container-high focus:outline-none flex items-center justify-center"
									>
										<MoreHorizontal class="h-5 w-5" />
									</DropdownMenu.Trigger>
									<DropdownMenu.Content
										align="end"
										class="w-48 p-1.5 rounded-sm border-none bg-surface-container-lowest"
									>
										<DropdownMenu.Item
											onSelect={() => openEdit(shift)}
											class="rounded-lg gap-3 font-semibold py-2.5 cursor-pointer hover:bg-surface-container-high"
										>
											<Pencil class="h-4 w-4 text-primary" />
											<span>{$_('admin.shifts.configure_shift')}</span>
										</DropdownMenu.Item>
										<DropdownMenu.Item
											onSelect={() => toggleStatus(shift)}
											class="rounded-lg gap-3 font-semibold py-2.5 cursor-pointer hover:bg-surface-container-high"
										>
											{#if shift.is_active}
												<XCircle class="h-4 w-4 text-error" />
												<span>{$_('admin.shifts.suspend_shift')}</span>
											{:else}
												<CheckCircle2 class="h-4 w-4 text-tertiary-container" />
												<span>{$_('admin.shifts.activate_shift')}</span>
											{/if}
										</DropdownMenu.Item>
										<DropdownMenu.Separator />
										<DropdownMenu.Item
											onSelect={() => remove(shift)}
											class="rounded-lg gap-3 font-semibold py-2.5 text-error focus:bg-error-container cursor-pointer hover:bg-error-container"
										>
											<Trash2 class="h-4 w-4" />
											<span>{$_('admin.shifts.delete_record')}</span>
										</DropdownMenu.Item>
									</DropdownMenu.Content>
								</DropdownMenu.Root>
							</div>
						</div>
					{/each}

					{#if filteredShifts.length === 0}
						<div
							class="flex flex-col items-center justify-center p-12 bg-surface-container-lowest border border-dashed border-outline-variant/30 rounded-sm"
						>
							<Clock class="h-12 w-12 text-outline-variant mb-4 opacity-50" />
							<p class="text-on-surface-variant font-medium text-center">
								{$_('admin.shifts.no_results')}
							</p>
						</div>
					{/if}
				{/if}
			</div>
		</main>
	{/if}

	<!-- Control Dialog (Precision Modern) -->
	{#if showModal}
		<Dialog.Root
			open={showModal}
			onOpenChange={(o) => {
				if (!o) showModal = false;
			}}
		>
			<Dialog.Content
				class="max-h-[calc(100vh-2rem)] overflow-y-auto sm:max-w-4xl p-0 bg-white border-none shadow-2xl rounded-sm"
			>
				<div class="p-12 text-primary relative h-48 flex items-end">
					<div class="z-10 space-y-2">
						<div
							class="inline-flex items-center gap-2 text-xs font-black uppercase tracking-[0.3em] opacity-70"
						>
							{editingItem ? $_('admin.shifts.edit_prefix') : $_('admin.shifts.create_prefix')}
						</div>
						<h2 class="text-4xl font-black tracking-tight">
							{$_('admin.shifts.shift_operative')}
							<span class="italic text-primary-fixed-dim/70">Operativo</span>
						</h2>
					</div>
				</div>

				<form
					onsubmit={(e) => {
						e.preventDefault();
						save();
					}}
					class="p-10 space-y-8"
				>
					<div class="space-y-6">
						<div class="space-y-2">
							<Label
								class="text-[10px] font-black uppercase tracking-widest text-on-surface-variant ml-1"
								>{$_('admin.shifts.name_label')}</Label
							>
							<Input
								bind:value={formData.name}
								placeholder="Ej: Jornada Nocturna"
								required
								class="h-14 bg-surface-container-low border-none rounded-sm px-6 font-bold text-base focus-visible:ring-2 focus-visible:ring-primary/20 transition-all font-sans"
							/>
						</div>

						<div class="grid grid-cols-2 gap-4">
							<div class="space-y-2">
								<Label
									class="text-[10px] font-black uppercase tracking-widest text-on-surface-variant ml-1"
									>{$_('admin.shifts.start_label')}</Label
								>
								<div class="relative group">
									<Sun
										class="absolute left-5 top-1/2 -translate-y-1/2 h-4 w-4 text-outline transition-colors group-focus-within:text-primary"
									/>
									<Input
										type="time"
										bind:value={formData.start_time}
										required
										class="h-14 bg-surface-container-low border-none rounded-sm pl-12 pr-6 font-mono font-bold text-base tabular-nums"
									/>
								</div>
							</div>
							<div class="space-y-2">
								<Label
									class="text-[10px] font-black uppercase tracking-widest text-on-surface-variant ml-1"
									>{$_('admin.shifts.end_label')}</Label
								>
								<div class="relative group">
									<Moon
										class="absolute left-5 top-1/2 -translate-y-1/2 h-4 w-4 text-outline transition-colors group-focus-within:text-primary"
									/>
									<Input
										type="time"
										bind:value={formData.end_time}
										required
										class="h-14 bg-surface-container-low border-none rounded-sm pl-12 pr-6 font-mono font-bold text-base tabular-nums"
									/>
								</div>
							</div>
						</div>

						<div class="grid grid-cols-2 gap-4 sm:gap-6 pt-2">
							<div class="space-y-2">
								<Label
									class="text-[10px] font-black uppercase tracking-widest text-on-surface-variant ml-1"
									>{$_('admin.shifts.grace_label')}</Label
								>
								<div class="relative">
									<Timer
										class="absolute left-5 top-1/2 -translate-y-1/2 h-4 w-4 text-emerald-500"
									/>
									<Input
										type="number"
										bind:value={formData.grace_period}
										required
										class="h-14 bg-surface-container-low border-none rounded-sm pl-12 pr-6 font-bold text-base tabular-nums"
									/>
								</div>
							</div>
							<div class="space-y-2">
								<Label
									class="text-[10px] font-black uppercase tracking-widest text-on-surface-variant ml-1"
									>{$_('admin.shifts.lunch_label')}</Label
								>
								<div class="relative">
									<Coffee class="absolute left-5 top-1/2 -translate-y-1/2 h-4 w-4 text-primary" />
									<Input
										type="number"
										bind:value={formData.lunch_duration_limit}
										required
										class="h-14 bg-surface-container-low border-none rounded-sm pl-12 pr-6 font-bold text-base tabular-nums"
									/>
								</div>
							</div>
						</div>
						<div class="space-y-3">
							<Label
								class="text-[10px] font-black uppercase tracking-widest text-on-surface-variant ml-1"
								>{$_('admin.shifts.work_days_label')}</Label
							>
							<div class="flex flex-wrap gap-2">
								{#each weekDays as day}
									<button
										type="button"
										onclick={() => toggleDay(day.id)}
										class="px-4 py-3 rounded-sm text-xs font-bold transition-all border-2 {formData.work_days.includes(
											day.id
										)
											? 'border-primary bg-primary text-white shadow-lg shadow-primary/20'
											: 'border-slate-100 bg-slate-50 text-slate-400 hover:border-slate-200'}"
									>
										{day.name}
									</button>
								{/each}
							</div>
						</div>

						<div class="space-y-3">
							<Label class="text-[10px] font-black uppercase tracking-widest text-on-surface-variant ml-1">
								{$_('admin.shifts.type_label')}
							</Label>
							<div class="grid grid-cols-1 md:grid-cols-3 gap-3">
								{#each shiftTypeOptions as option}
									{@const Icon = option.icon}
									<button
										type="button"
										onclick={() => selectShiftType(option.id)}
										class="flex flex-col p-4 rounded-sm border-2 text-left transition-all {formData.shift_type === option.id ? 'border-primary bg-primary/5' : 'border-slate-100 hover:border-slate-200'}"
									>
										<div class="flex items-center justify-between mb-2">
											<Icon size={18} class={formData.shift_type === option.id ? 'text-primary' : 'text-slate-400'} />
											{#if formData.shift_type === option.id}
												<CheckCircle2 size={14} class="text-primary" />
											{/if}
										</div>
										<span class="text-xs font-black uppercase tracking-tight {formData.shift_type === option.id ? 'text-primary' : 'text-slate-600'}">{option.label}</span>
										<p class="text-[9px] leading-tight text-slate-400 mt-1">{option.hint}</p>
									</button>
								{/each}
							</div>
						</div>

						<div class="pt-4 border-t border-slate-100">
							<Label
								class="text-[10px] font-black uppercase tracking-[0.2em] text-primary mb-4 block"
								>{$_('admin.shifts.compliance_policies')}</Label
							>
							<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
								<button
									type="button"
									onclick={() => (formData.enforce_lateness = !formData.enforce_lateness)}
									class="flex flex-col p-4 rounded-sm border-2 text-left transition-all {formData.enforce_lateness
										? 'border-primary bg-primary/5'
										: 'border-slate-100 opacity-60'}"
								>
									<div class="flex items-center justify-between mb-1">
										<Clock
											size={16}
											class={formData.enforce_lateness ? 'text-primary' : 'text-slate-400'}
										/>
										{#if formData.enforce_lateness}<CheckCircle2
												size={14}
												class="text-primary"
											/>{/if}
									</div>
									<span class="text-xs font-black uppercase tracking-tight"
										>{$_('admin.shifts.enforce_lateness')}</span
									>
									<p class="text-[10px] leading-tight text-slate-400 mt-1">
										{$_('admin.shifts.enforce_lateness_hint')}
									</p>
								</button>

								<button
									type="button"
									onclick={() => (formData.enforce_lunch_limit = !formData.enforce_lunch_limit)}
									class="flex flex-col p-4 rounded-sm border-2 text-left transition-all {formData.enforce_lunch_limit
										? 'border-primary bg-primary/5'
										: 'border-slate-100 opacity-60'}"
								>
									<div class="flex items-center justify-between mb-1">
										<Coffee
											size={16}
											class={formData.enforce_lunch_limit ? 'text-primary' : 'text-slate-400'}
										/>
										{#if formData.enforce_lunch_limit}<CheckCircle2
												size={14}
												class="text-primary"
											/>{/if}
									</div>
									<span class="text-xs font-black uppercase tracking-tight"
										>{$_('admin.shifts.enforce_lunch')}</span
									>
									<p class="text-[10px] leading-tight text-slate-400 mt-1">
										{$_('admin.shifts.enforce_lunch_hint')}
									</p>
								</button>

								<button
									type="button"
									onclick={() => (formData.enforce_geofence = !formData.enforce_geofence)}
									class="flex flex-col p-4 rounded-sm border-2 text-left transition-all {formData.enforce_geofence
										? 'border-primary bg-primary/5'
										: 'border-slate-100 opacity-60'}"
								>
									<div class="flex items-center justify-between mb-1">
										<LayoutDashboard
											size={16}
											class={formData.enforce_geofence ? 'text-primary' : 'text-slate-400'}
										/>
										{#if formData.enforce_geofence}<CheckCircle2
												size={14}
												class="text-primary"
											/>{/if}
									</div>
									<span class="text-xs font-black uppercase tracking-tight"
										>{$_('admin.shifts.enforce_gps')}</span
									>
									<p class="text-[10px] leading-tight text-slate-400 mt-1">
										{$_('admin.shifts.enforce_gps_hint')}
									</p>
								</button>
							</div>
						</div>

						<div class="pt-2">
							<Label
								class="text-[10px] font-black uppercase tracking-widest text-on-surface-variant ml-1 mb-3 block"
								>{$_('admin.shifts.status_label')}</Label
							>
							<div class="flex items-center gap-4">
								<!-- Custom Select Styled Toggle -->
								<button
									type="button"
									onclick={() => (formData.is_active = !formData.is_active)}
									class="flex-1 flex items-center justify-between p-4 rounded-sm border-2 transition-all duration-300 {formData.is_active
										? 'border-primary bg-primary/5'
										: 'border-outline-variant/30 opacity-50'}"
								>
									<div class="flex items-center gap-3">
										<div
											class="w-2 h-2 rounded-full {formData.is_active
												? 'bg-primary'
												: 'bg-outline-variant'}"
										></div>
										<span class="font-bold text-sm">{$_('common.active')}</span>
									</div>
									{#if formData.is_active}
										<CheckCircle2 class="h-5 w-5 text-primary" />
									{/if}
								</button>

								<button
									type="button"
									onclick={() => (formData.is_active = !formData.is_active)}
									class="flex-1 flex items-center justify-between p-4 rounded-sm border-2 transition-all duration-300 {!formData.is_active
										? 'border-rose-500 bg-rose-50'
										: 'border-outline-variant/30 opacity-50'}"
								>
									<div class="flex items-center gap-3">
										<div
											class="w-1.5 h-1.5 rounded-full {!formData.is_active
												? 'bg-rose-500'
												: 'bg-outline-variant'}"
										></div>
										<span class="font-bold text-sm">{$_('common.inactive')}</span>
									</div>
									{#if !formData.is_active}
										<XCircle class="h-5 w-5 text-rose-500" />
									{/if}
								</button>
							</div>
						</div>
					</div>

					<div class="flex flex-col gap-3 pt-6">
						<Button
							type="submit"
							class="h-16 rounded-sm bg-primary hover:bg-primary-variant text-white text-lg font-black shadow-xl shadow-primary/20 transition-all active:scale-[0.98] disabled:opacity-50"
							disabled={loading}
						>
							{#if loading}
								<Loader2 class="h-6 w-6 animate-spin" />
							{:else}
								{$_('admin.shifts.save_button')}
							{/if}
						</Button>
						<Button
							type="button"
							variant="ghost"
							onclick={() => (showModal = false)}
							class="h-12 font-bold text-on-surface-variant hover:text-primary transition-colors uppercase text-[10px] tracking-widest"
						>
							{$_('admin.shifts.cancel_operation')}
						</Button>
					</div>
				</form>
			</Dialog.Content>
		</Dialog.Root>
	{/if}
</div>
