<script lang="ts">
	import { onMount } from 'svelte';
	import { fly, fade } from 'svelte/transition';
	import { quintOut } from 'svelte/easing';
	import { _ } from 'svelte-i18n';
	import { apiFetch } from '$lib/api';
	import { authState } from '$lib/auth.svelte';
	import type { Position } from '$lib/types/models';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import {
		Briefcase,
		Pencil,
		Trash2,
		ArrowLeft,
		Plus,
		DollarSign,
		AlertCircle,
		Users,
		GanttChartSquare,
		Eye,
		ChevronRight,
		Loader2,
		Target,
		Search,
		MoreHorizontal
	} from 'lucide-svelte';

	let positions = $state<Position[]>([]);
	let loading = $state(true);
	let showModal = $state(false);
	let editingItem = $state<Position | null>(null);

	let formData = $state({
		name: '',
		base_pay: 0,
		late_penalty: 0,
		out_of_range_penalty: 0
	});

	const canEdit = $derived(authState.isAdmin);

	async function loadPositions() {
		loading = true;
		const res = await apiFetch<Position[]>('/admin/positions');
		if (res.ok) positions = await res.json();
		loading = false;
	}

	function openCreate() {
		if (!canEdit) return;
		editingItem = null;
		formData = { name: '', base_pay: 0, late_penalty: 0, out_of_range_penalty: 0 };
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
		const url = editingItem ? `/admin/positions/${editingItem.id}` : '/admin/positions';

		const res = await apiFetch(url, {
			method,
			body: JSON.stringify(formData)
		});

		if (res.ok) {
			showModal = false;
			loadPositions();
		} else {
			const err = await res.json();
			alert(err.error || $_('common.error_saving'));
		}
	}

	async function remove(item: any) {
		if (!canEdit) return;
		if (!confirm(`${$_('common.confirm_delete')} ${item.name}?`)) return;
		const res = await apiFetch(`/admin/positions/${item.id}`, { method: 'DELETE' });
		if (res.ok) loadPositions();
	}

	let searchQuery = $state('');

	const filteredPositions = $derived(
		positions.filter((p) => p.name.toLowerCase().includes(searchQuery.toLowerCase()))
	);

	onMount(loadPositions);
</script>

<main class="pt-8 px-6 max-w-5xl mx-auto space-y-10 pb-24">
	<!-- Asymmetric Header -->
	<div
		class="flex justify-between items-end"
		in:fly={{ y: 20, duration: 800, easing: quintOut }}
	>
		<div class="space-y-1">
			<p class="text-[10px] font-black uppercase tracking-[0.3em] text-primary/50">
				{$_('admin.positions.infrastructure')}
			</p>
			<h2 class="text-5xl font-black text-primary tracking-tighter leading-none">{$_('admin.positions.title')}.</h2>
			<p class="text-[10px] font-bold text-slate-400 uppercase tracking-widest pt-1 italic">
				{positions.length} {$_('admin.positions.detected_roles')}
			</p>
		</div>
		<div class="flex items-center gap-6">
			{#if canEdit}
				<Button
					onclick={openCreate}
					class="h-14 px-8 rounded-sm bg-primary hover:bg-primary/90 text-white font-black text-xs uppercase tracking-[0.2em] shadow-xl shadow-primary/20 transition-all active:scale-95 flex items-center gap-2"
				>
					<Plus size={18} strokeWidth={3} />
					{$_('admin.positions.register_button')}
				</Button>
			{/if}
			<div class="bg-slate-100 w-1.5 h-16 rounded-full hidden sm:block"></div>
		</div>
	</div>
	<!-- Search and Filter Bar -->
	<div
		class="relative group"
		in:fly={{ y: 20, duration: 800, easing: quintOut, delay: 100 }}
	>
		<div
			class="flex items-center bg-white border-b-2 border-slate-200 focus-within:border-primary transition-all p-4 rounded-t-2xl shadow-sm"
		>
			<Search class="text-slate-300 mr-3" size={20} />
			<input
				type="text"
				placeholder={$_('admin.positions.search_placeholder')}
				bind:value={searchQuery}
				class="bg-transparent border-none focus:ring-0 w-full text-sm font-bold text-slate-900 placeholder:text-slate-300 uppercase tracking-tight"
			/>
		</div>
	</div>
	<!-- List Container -->
	<div class="space-y-4" in:fly={{ y: 20, duration: 800, easing: quintOut, delay: 200 }}>
		{#if loading}
			{#each Array(3) as _}
				<div class="h-32 bg-slate-50 rounded-sm animate-pulse"></div>
			{/each}
		{:else if filteredPositions.length === 0}
			<div class="py-20 text-center bg-white rounded-sm border border-slate-100 border-dashed">
				<p class="text-slate-400 font-black uppercase tracking-[0.2em] text-[10px]">
					{$_('admin.positions.no_results')}
				</p>
			</div>
		{:else}
			{#each filteredPositions as position (position.id)}
				<div
					class="bg-white p-6 rounded-sm border border-slate-50 shadow-xl shadow-slate-200/40 flex flex-col gap-4 transition-all hover:scale-[1.01] active:scale-[0.98] group relative overflow-hidden"
					in:fade={{ duration: 200 }}
				>
					<div class="flex justify-between items-start gap-4">
						<div class="flex items-center gap-4">
							<div
								class="w-12 h-12 rounded-sm bg-slate-50 flex items-center justify-center text-slate-400 group-hover:bg-primary group-hover:text-white transition-all duration-500"
							>
								<Briefcase size={22} />
							</div>
							<div>
								<h3 class="font-black text-xl text-primary tracking-tight leading-tight">
									<a href="/admin/positions/{position.id}">{position.name}</a>
								</h3>
								<div class="flex items-center gap-2 mt-1">
									<Users size={12} class="text-slate-300" />
									<p class="text-[10px] font-black text-slate-400 uppercase tracking-widest">
										{position.employees_count || 0} {$_('admin.positions.assigned')}
									</p>
								</div>
							</div>
						</div>

						<DropdownMenu.Root>
							<DropdownMenu.Trigger
								class="p-2 text-slate-300 hover:text-primary rounded-full hover:bg-slate-50 transition-all focus:outline-none"
							>
								<MoreHorizontal size={20} />
							</DropdownMenu.Trigger>
							<DropdownMenu.Content
								class="bg-white border-none shadow-premium rounded-sm p-2 min-w-[180px]"
							>
								<DropdownMenu.Item
									class="flex items-center gap-3 px-4 py-3 text-[10px] font-black uppercase tracking-widest text-slate-500 hover:text-primary hover:bg-slate-50 rounded-sm cursor-pointer"
									onSelect={() => openEdit(position)}
								>
									<Pencil size={14} />
									{$_('admin.positions.configure_structure')}
								</DropdownMenu.Item>
								<DropdownMenu.Item
									class="flex items-center gap-3 px-4 py-3 text-[10px] font-black uppercase tracking-widest text-rose-500 hover:bg-rose-50 rounded-sm cursor-pointer"
									onSelect={() => remove(position)}
								>
									<Trash2 size={14} />
									{$_('admin.positions.delete_role')}
								</DropdownMenu.Item>
							</DropdownMenu.Content>
						</DropdownMenu.Root>
					</div>

					<div class="flex items-center justify-between pt-2">
						<div class="flex gap-2">
							{#if position.late_penalty > 0}
								<span
									class="bg-rose-50 text-rose-600 text-[9px] font-black px-3 py-1.5 rounded-full uppercase tracking-widest border border-rose-100/50"
									>{$_('admin.positions.late')}</span
								>
							{/if}
							{#if position.out_of_range_penalty > 0}
								<span
									class="bg-amber-50 text-amber-600 text-[9px] font-black px-3 py-1.5 rounded-full uppercase tracking-widest border border-amber-100/50"
									>{$_('admin.positions.gps_breach')}</span
								>
							{/if}
							{#if position.late_penalty === 0 && position.out_of_range_penalty === 0}
								<span class="text-[9px] font-black text-slate-300 uppercase tracking-widest italic"
									>{$_('admin.positions.no_penalties')}</span
								>
							{/if}
						</div>
						<div class="text-right">
							<span class="font-black text-2xl text-primary tracking-tighter">
								<span class="text-xs align-top mr-0.5">$</span>{position.base_pay?.toFixed(2)}
								<span class="text-[10px] text-slate-300 font-black uppercase tracking-widest ml-1"
									>{$_('admin.positions.per_hour')}</span
								>
							</span>
						</div>
					</div>
				</div>
			{/each}
		{/if}
	</div>
</main>

<!-- Editor Dialog -->
{#if showModal}
	<Dialog.Root
		open={showModal}
		onOpenChange={(o) => {
			if (!o) showModal = false;
		}}
	>
		<Dialog.Content
			class="rounded-md border-none shadow-premium bg-white p-12 sm:w-full md:max-w-4xl"
		>
			<Dialog.Header class="space-y-6">
				<div class="space-y-2 text-left">
					<Dialog.Title class="text-4xl font-black tracking-tighter text-slate-900">
						{editingItem ? $_('admin.positions.edit_title') : $_('admin.positions.create_title')} <span class="text-primary italic">Puesto</span>
					</Dialog.Title>
					<Dialog.Description
						class="text-sm font-bold text-slate-400 uppercase tracking-widest leading-relaxed"
					>
						{$_('admin.positions.description')}
					</Dialog.Description>
				</div>
			</Dialog.Header>

			<form
				onsubmit={(e) => {
					e.preventDefault();
					save();
				}}
				class="py-8 space-y-8"
			>
				<div class="space-y-6">
					<div class="space-y-2">
						<Label class="text-sm font-black text-slate-900 ml-1">{$_('admin.positions.name_label')}</Label>
						<Input
							bind:value={formData.name}
							placeholder="Ej: Analista de Planta III"
							class="h-14 bg-slate-50 border-none rounded-md font-bold px-5 focus-visible:ring-2 focus-visible:ring-primary/20 transition-all font-sans"
							required
						/>
					</div>

					<div class="space-y-2">
						<Label class="text-sm font-black text-slate-900 ml-1">{$_('admin.positions.pay_label')}</Label>
						<div class="relative group">
							<div
								class="absolute left-5 top-1/2 -translate-y-1/2 text-slate-400 group-focus-within:text-primary transition-colors"
							>
								<DollarSign class="h-5 w-5" />
							</div>
							<Input
								type="number"
								step="0.01"
								bind:value={formData.base_pay}
								class="h-14 bg-slate-50 border-none rounded-md font-mono font-bold pl-12 pr-5 font-sans"
								required
							/>
						</div>
					</div>

					<div class="grid grid-cols-2 gap-4">
						<div class="space-y-2">
							<Label
								class="text-sm font-black text-slate-900 ml-1 uppercase tracking-tighter text-[10px]"
								>{$_('admin.positions.late_deduction')}</Label
							>
							<Input
								type="number"
								step="0.01"
								bind:value={formData.late_penalty}
								class="h-14 bg-slate-50 border-none rounded-md font-mono font-bold px-5 text-rose-500"
								required
							/>
						</div>
						<div class="space-y-2">
							<Label
								class="text-sm font-black text-slate-900 ml-1 uppercase tracking-tighter text-[10px]"
								>{$_('admin.positions.gps_deduction')}</Label
							>
							<Input
								type="number"
								step="0.01"
								bind:value={formData.out_of_range_penalty}
								class="h-14 bg-slate-50 border-none rounded-md font-mono font-bold px-5 text-rose-500"
								required
							/>
						</div>
					</div>
				</div>

				<div class="flex flex-col gap-3 pt-4">
					<Button
						type="submit"
						class="h-16 rounded-md font-black text-lg gap-2 shadow-xl shadow-primary/20 hover:scale-[1.02] active:scale-95 transition-all text-white font-sans"
						disabled={loading}
					>
						{#if loading}
							<Loader2 class="h-5 w-5 animate-spin" />
							{$_('common.syncing')}
						{:else}
							{editingItem ? $_('admin.positions.update_definition') : $_('admin.positions.register_button')}
						{/if}
					</Button>
					<Button
						type="button"
						variant="ghost"
						class="h-12 rounded-md font-black text-slate-400 hover:text-slate-600 hover:bg-slate-50 font-sans"
						onclick={() => (showModal = false)}
					>
						{$_('common.cancel')}
					</Button>
				</div>
			</form>
		</Dialog.Content>
	</Dialog.Root>
{/if}
