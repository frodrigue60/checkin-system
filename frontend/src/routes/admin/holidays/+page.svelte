<script lang="ts">
	import { onMount } from 'svelte';
	import { _ } from 'svelte-i18n';
	import { apiFetch } from '$lib/api';
	import { authState } from '$lib/auth.svelte';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Textarea } from '$lib/components/ui/textarea';
	import { CalendarDays, Pencil, Trash2, ArrowLeft, Plus, Calendar, MoreHorizontal, ChevronDown, Loader2, GanttChartSquare, ShieldAlert } from 'lucide-svelte';
	import { fade, slide, fly } from 'svelte/transition';
	import { quintOut } from 'svelte/easing';

	let holidays = $state([]);
	let loading = $state(true);
	let showModal = $state(false);
	let editingItem = $state<any>(null);

	let selectedMonth = $state('all');
	const months = [
		{ id: '01', name: 'jan' },
		{ id: '02', name: 'feb' },
		{ id: '03', name: 'mar' },
		{ id: '04', name: 'apr' },
		{ id: '05', name: 'may' },
		{ id: '06', name: 'jun' },
		{ id: '07', name: 'jul' },
		{ id: '08', name: 'aug' },
		{ id: '09', name: 'sep' },
		{ id: '10', name: 'oct' },
		{ id: '11', name: 'nov' },
		{ id: '12', name: 'dec' }
	];

	let formData = $state({
		name: '',
		date: new Date().toISOString().split('T')[0],
		description: '',
		type: 'mandatory'
	});

	const canEdit = $derived(authState.isAdmin);
	const currentYear = new Date().getFullYear();
	const currentQuarter = $derived(`Q${Math.floor((new Date().getMonth() + 3) / 3)}`);

	async function loadHolidays() {
		loading = true;
		const res = await apiFetch('/admin/holidays');
		if (res.ok) {
			const data = await res.json();
			// Sort chronologically
			holidays = data.sort((a, b) => new Date(a.date).getTime() - new Date(b.date).getTime());
		}
		loading = false;
	}

	function openCreate() {
		if (!canEdit) return;
		editingItem = null;
		formData = {
			name: '',
			date: new Date().toISOString().split('T')[0],
			description: '',
			type: 'mandatory'
		};
		showModal = true;
	}

	function openEdit(item: any) {
		if (!canEdit) return;
		editingItem = item;
		formData = {
			...item,
			date: item.date.split('T')[0],
			description: item.description || ''
		};
		showModal = true;
	}

	async function save() {
		const method = editingItem ? 'PUT' : 'POST';
		const url = editingItem ? `/admin/holidays/${editingItem.id}` : '/admin/holidays';

		const res = await apiFetch(url, {
			method,
			body: JSON.stringify(formData)
		});

		if (res.ok) {
			showModal = false;
			loadHolidays();
		} else {
			const err = await res.json();
			alert(err.error || $_('common.error_saving'));
		}
	}

	async function remove(item: any) {
		if (!canEdit) return;
		if (!confirm($_('admin.holidays.confirm_delete', { values: { name: item.name } }))) return;
		const res = await apiFetch(`/admin/holidays/${item.id}`, { method: 'DELETE' });
		if (res.ok) loadHolidays();
	}

	const filteredHolidays = $derived(
		holidays.filter((h) => {
			if (selectedMonth === 'all') return true;
			const month = h.date.split('-')[1];
			return month === selectedMonth;
		})
	);

	// Groups holidays by month for the separators
	const groupedHolidays = $derived(
		filteredHolidays.reduce((acc, h) => {
			const monthIdx = h.date.split('-')[1];
			const monthKey = months.find((m) => m.id === monthIdx)?.name || 'unknown';
			if (!acc[monthKey]) acc[monthKey] = [];
			acc[monthKey].push(h);
			return acc;
		}, {})
	);

	onMount(loadHolidays);
</script>

<div class="min-h-screen pb-24">
	<main class="pt-8 px-6 max-w-5xl mx-auto space-y-12">
		<!-- Timeline Hero (Asymmetric Style) -->
		<section
			class="flex justify-between items-end"
			in:fly={{ y: 20, duration: 800, easing: quintOut }}
		>
			<div class="space-y-1">
				<p class="text-[10px] font-black uppercase tracking-[0.3em] text-primary/50">
					{$_('admin.holidays.timeline_focus')}
				</p>
				<h2 class="text-6xl font-black text-primary leading-none tracking-tighter">
					{currentYear}
				</h2>
				<p class="text-[10px] font-bold text-slate-400 uppercase tracking-widest pt-2 italic">
					{holidays.length} {$_('admin.holidays.programmed_events')}
				</p>
			</div>

			<div class="flex items-center gap-8">
				{#if canEdit}
					<Button
						onclick={openCreate}
						class="h-14 px-8 rounded-sm bg-primary hover:bg-primary/90 text-white font-black text-xs uppercase tracking-[0.2em] shadow-xl shadow-primary/20 transition-all active:scale-95 flex items-center gap-2"
					>
						<Plus size={18} strokeWidth={3} />
						{$_('admin.holidays.publish_event')}
					</Button>
				{/if}
				<!-- <div class="bg-slate-100 w-1.5 h-16 rounded-full hidden md:block"></div>
				<div class="text-right hidden sm:block">
					<span class="text-6xl font-black text-primary/10 select-none tracking-tighter"
						>{currentQuarter}</span
					>
				</div> -->
			</div>
		</section>

		<!-- Month Switcher -->
		<nav
			class="flex overflow-x-auto no-scrollbar gap-2 -mx-6 px-6 py-2"
			in:fly={{ y: 20, duration: 800, easing: quintOut, delay: 100 }}
		>
			<button
				onclick={() => (selectedMonth = 'all')}
				class="flex-none px-6 py-3 rounded-full text-xs font-black uppercase tracking-widest transition-all duration-300 {selectedMonth ===
				'all'
					? 'bg-primary text-white shadow-lg shadow-primary/30 scale-105'
					: 'bg-white text-slate-400 hover:bg-slate-100'}"
			>
				{$_('common.all')}
			</button>
			{#each months as month}
				<button
					onclick={() => (selectedMonth = month.id)}
					class="flex-none px-6 py-3 rounded-full text-xs font-black uppercase tracking-widest transition-all duration-300 {selectedMonth ===
					month.id
						? 'bg-primary text-white shadow-lg shadow-primary/30 scale-105'
						: 'bg-white text-slate-400 hover:bg-slate-100'}"
				>
					{$_(`common.months.${month.name}`)}
				</button>
			{/each}
		</nav>

		<!-- Vertical Timeline -->
		<div
			class="space-y-10 relative"
			in:fly={{ y: 20, duration: 800, easing: quintOut, delay: 200 }}
		>
			{#if loading}
				{#each Array(3) as _}
					<div class="flex gap-6 animate-pulse">
						<div class="w-16 h-12 bg-white rounded-sm"></div>
						<div class="flex-grow h-32 bg-white rounded-sm"></div>
					</div>
				{/each}
			{:else if filteredHolidays.length === 0}
				<div
					class="py-20 text-center bg-white rounded-sm shadow-sm border border-slate-50 border-dashed"
				>
					<p class="text-slate-400 font-black uppercase tracking-[0.2em] text-xs">
						{$_('admin.holidays.no_entries')}
					</p>
				</div>
			{:else}
				{#each Object.entries(groupedHolidays) as [monthName, items], mIdx}
					{#if selectedMonth === 'all' && mIdx > 0}
						<div class="flex items-center gap-4 py-6" in:fade>
							<div class="flex-grow h-px bg-slate-100"></div>
							<span class="text-[10px] font-black uppercase tracking-[0.3em] text-slate-300"
								>{$_(`common.months.${monthName}`)}</span
							>
							<div class="flex-grow h-px bg-slate-100"></div>
						</div>
					{/if}

					{#each items as holiday}
						<article class="flex gap-6 group" in:fade>
							<!-- Date Block -->
							<div class="w-16 flex-none text-right pt-2">
								<p
									class="text-3xl font-black text-primary leading-none tracking-tighter transition-all group-hover:scale-110"
								>
									{new Date(holiday.date).getUTCDate().toString().padStart(2, '0')}
								</p>
								<p
									class="text-[10px] uppercase font-black text-slate-300 tracking-[0.2em] mt-1 italic"
								>
									{$_(`common.months.${months.find((m) => m.id === holiday.date.split('-')[1])?.name}`)}
								</p>
							</div>

							<!-- Content Card -->
							<div
								class="flex-grow bg-white rounded-sm p-6 relative transition-all duration-500 border border-transparent shadow-sm hover:shadow-2xl hover:shadow-primary/5 hover:border-primary/10"
							>
								<div class="flex justify-between items-start mb-4">
									<div class="flex gap-2">
										{#if holiday.type === 'mandatory'}
											<span
												class="px-4 py-1.5 rounded-full bg-rose-50 text-rose-600 text-[9px] font-black uppercase tracking-widest border border-rose-100/50"
											>
												{$_('admin.holidays.mandatory')}
											</span>
										{:else}
											<span
												class="px-4 py-1.5 rounded-full bg-slate-50 text-slate-400 text-[9px] font-black uppercase tracking-widest"
											>
												{$_('admin.holidays.informative')}
											</span>
										{/if}
									</div>

									<DropdownMenu.Root>
										<DropdownMenu.Trigger
											class="p-2 text-slate-200 hover:text-primary transition-colors focus:outline-none h-8 w-8 flex items-center justify-center rounded-full hover:bg-slate-50"
										>
											<MoreHorizontal size={18} />
										</DropdownMenu.Trigger>
										<DropdownMenu.Content
											class="bg-white border-none shadow-2xl rounded-sm p-2 min-w-[180px]"
										>
											<DropdownMenu.Item
												class="flex items-center gap-3 px-4 py-3 text-[10px] font-black uppercase tracking-widest text-slate-500 hover:text-primary hover:bg-slate-50 rounded-sm cursor-pointer"
												onSelect={() => openEdit(holiday)}
											>
												<Pencil size={14} />
												{$_('admin.holidays.configure_event')}
											</DropdownMenu.Item>
											<DropdownMenu.Item
												class="flex items-center gap-3 px-4 py-3 text-[10px] font-black uppercase tracking-widest text-rose-500 hover:bg-rose-50 rounded-sm cursor-pointer"
												onSelect={() => remove(holiday)}
											>
												<Trash2 size={14} />
												{$_('admin.holidays.de_publish')}
											</DropdownMenu.Item>
										</DropdownMenu.Content>
									</DropdownMenu.Root>
								</div>

								<h3 class="text-xl font-black text-primary leading-tight mb-2 tracking-tight">
									{holiday.name}
								</h3>
								<p class="text-sm font-medium text-slate-400 leading-relaxed italic">
									{holiday.description || $_('admin.holidays.no_description')}
								</p>
							</div>
						</article>
					{/each}
				{/each}
			{/if}
		</div>
	</main>

	<!-- Create/Edit Dialog -->
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
					<div class="space-y-1">
						<Dialog.Title class="text-4xl font-black tracking-tighter text-primary">
							{editingItem ? $_('common.edit') : $_('common.add')} <span class="italic opacity-50">Event.</span>
						</Dialog.Title>
						<Dialog.Description
							class="text-[10px] font-black uppercase tracking-widest text-slate-400"
						>
							{$_('admin.holidays.modal_description')}
						</Dialog.Description>
					</div>
				</Dialog.Header>

				<form
					onsubmit={(e) => {
						e.preventDefault();
						save();
					}}
					class="pt-8 space-y-10"
				>
					<div class="space-y-6">
						<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
							<div class="space-y-2">
								<Label class="text-[10px] font-black uppercase tracking-widest text-primary ml-1"
									>{$_('admin.holidays.event_marker')}</Label
								>
								<Input
									bind:value={formData.name}
									placeholder="e.g. Founder's Day"
									class="h-14 bg-slate-50 border-none rounded-sm font-black text-xs px-5 focus:ring-4 focus:ring-primary/5 transition-all text-primary"
									required
								/>
							</div>
							<div class="space-y-2">
								<Label class="text-[10px] font-black uppercase tracking-widest text-primary ml-1"
									>{$_('admin.holidays.temporal_entry')}</Label
								>
								<div class="relative group">
									<Calendar
										class="absolute left-5 top-1/2 -translate-y-1/2 text-slate-300 group-focus-within:text-primary transition-colors h-4 w-4"
									/>
									<Input
										type="date"
										bind:value={formData.date}
										class="h-14 bg-slate-50 border-none rounded-sm font-black text-xs pl-12 pr-5 text-primary"
										required
									/>
								</div>
							</div>
						</div>

						<div class="space-y-2">
							<Label class="text-[10px] font-black uppercase tracking-widest text-primary ml-1"
								>{$_('admin.holidays.logistical_brief')}</Label
							>
							<Textarea
								bind:value={formData.description}
								placeholder="Document operational impact..."
								class="h-28 bg-slate-50 border-none rounded-sm font-black text-xs p-5 focus:ring-4 focus:ring-primary/5 transition-all resize-none text-primary"
							/>
						</div>

						<div
							class="p-6 bg-slate-900 rounded-sm space-y-5 border-4 border-slate-800 text-white shadow-2xl"
						>
							<div class="flex items-center gap-2">
								<ShieldAlert class="h-4 w-4 text-primary" />
								<span class="text-[9px] font-black uppercase tracking-widest text-slate-400"
									>{$_('admin.holidays.security_perimeter')}</span
								>
							</div>
							<div class="grid grid-cols-2 gap-3">
								<button
									type="button"
									onclick={() => (formData.type = 'mandatory')}
									class="flex flex-col p-4 rounded-sm border-2 transition-all text-left {formData.type ===
									'mandatory'
										? 'bg-primary/20 border-primary'
										: 'bg-slate-800/50 border-transparent hover:bg-slate-800'}"
								>
									<span class="text-[10px] font-black tracking-widest">{$_('admin.holidays.mandatory')}</span>
									<span
										class="text-[8px] font-medium text-slate-400 uppercase tracking-tight italic pt-1"
										>{$_('admin.holidays.operational_block')}</span
									>
								</button>
								<button
									type="button"
									onclick={() => (formData.type = 'optional')}
									class="flex flex-col p-4 rounded-sm border-2 transition-all text-left {formData.type ===
									'optional'
										? 'bg-primary/20 border-primary'
										: 'bg-slate-800/50 border-transparent hover:bg-slate-800'}"
								>
									<span class="text-[10px] font-black tracking-widest text-slate-400"
										>{$_('admin.holidays.informative')}</span
									>
									<span
										class="text-[8px] font-medium text-slate-500 uppercase tracking-tight italic pt-1"
										>{$_('admin.holidays.reference_only')}</span
									>
								</button>
							</div>
						</div>
					</div>

					<div class="flex flex-col gap-3">
						<Button
							type="submit"
							class="h-16 rounded-sm font-black text-xs uppercase tracking-widest shadow-2xl shadow-primary/30 transition-all hover:scale-[1.01] active:scale-95 text-white"
							disabled={loading}
						>
							{#if loading}
								<Loader2 class="h-4 w-4 animate-spin mr-2" />
								{$_('admin.holidays.publishing')}
							{:else}
								{editingItem ? $_('admin.holidays.sync_event') : $_('admin.holidays.publish_event')}
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
</div>
