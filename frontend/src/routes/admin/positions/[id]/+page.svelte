<script lang="ts">
	import { onMount } from 'svelte';
	import { fly, fade } from 'svelte/transition';
	import { quintOut } from 'svelte/easing';
	import { page } from '$app/state';
	import { apiFetch } from '$lib/api';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import {
		Loader2,
		DollarSign,
		Scale,
		Users,
		AlertTriangle,
		ArrowLeft,
		ChevronRight,
		MoreVertical,
		Activity,
		ShieldAlert,
		History,
		Wallet,
		Briefcase
	} from 'lucide-svelte';

	let posId = $derived(page.params.id);
	let data = $state<any>(null);
	let loading = $state(true);
	let errorMsg = $state('');

	async function loadDetails() {
		loading = true;
		const res = await apiFetch(`/admin/positions/${posId}/details`);
		if (res.ok) {
			data = await res.json();
		} else {
			errorMsg = 'No se pudo cargar la información del puesto';
		}
		loading = false;
	}

	function formatCurrency(val: number) {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: 'USD'
		}).format(val);
	}

	onMount(loadDetails);
</script>

<div class="min-h-screen bg-white p-6 space-y-6">
	{#if loading}
		<div class="flex flex-col items-center justify-center min-h-[80vh] gap-6">
			<div class="relative">
				<div class="h-20 w-20 rounded-sm border-4 border-slate-100 animate-pulse"></div>
				<Loader2
					class="h-8 w-8 animate-spin text-primary absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2"
				/>
			</div>
			<div class="space-y-1 text-center">
				<p class="text-xl font-black tracking-tight text-slate-900 uppercase italic">
					Sincronizando Cargo
				</p>
				<p class="text-[9px] font-black text-slate-400 uppercase tracking-widest">
					Validando Esquema Financiero
				</p>
			</div>
		</div>
	{:else if errorMsg}
		<div class="max-w-md mx-auto p-12 text-center space-y-6">
			<div class="h-20 w-20 bg-rose-50 rounded-sm flex items-center justify-center mx-auto">
				<AlertTriangle class="h-10 w-10 text-rose-500" />
			</div>
			<h2 class="text-2xl font-black text-rose-900 tracking-tighter">{errorMsg}</h2>
			<Button variant="outline" href="/admin/positions" class="font-black h-12 px-8"
				>Volver al Directorio</Button
			>
		</div>
	{:else if data}
		<!-- Asymmetric Header -->
		<div
			class="flex flex-col md:flex-row justify-between items-start md:items-end gap-6 px-6 max-w-xl mx-auto w-full"
			in:fly={{ y: 20, duration: 800, easing: quintOut }}
		>
			<div class="space-y-1">
				<p class="text-[10px] font-black uppercase tracking-[0.3em] text-primary/50">
					Position Profile
				</p>
				<h1 class="text-5xl font-black text-primary tracking-tighter leading-none italic uppercase">
					{data.position.name}.
				</h1>
				<p class="text-[10px] font-bold text-slate-400 uppercase tracking-widest pt-1 italic">
					Data-Driven Identity Matrix
				</p>
			</div>
			<!-- Hourly Rate Score -->
			<section class="space-y-2">
				<p class="text-slate-400 font-black text-[10px] uppercase tracking-[0.2em]">
					Standard Base Rate
				</p>
				<div class="flex items-baseline gap-2">
					<h2 class="font-black text-7xl tracking-tighter text-primary">
						{formatCurrency(data.position.base_pay || 0).replace('$', '')}
					</h2>
					<div class="flex flex-col">
						<span class="font-black text-xl text-slate-400 uppercase italic leading-none"
							>$ USD</span
						>
						<span class="font-black text-sm text-slate-300 uppercase tracking-widest">/ HR</span>
					</div>
				</div>
			</section>
		</div>

		<main
			class="px-6 max-w-4xl mx-auto space-y-10"
			in:fly={{ y: 20, duration: 800, easing: quintOut, delay: 100 }}
		>
			<!-- Budgetary Impact Bento Card -->
			<section>
				<div
					class="bg-slate-50 rounded-sm p-8 border border-slate-100 shadow-xl shadow-slate-200/40 relative overflow-hidden group"
				>
					<div class="flex justify-between items-start mb-10">
						<div>
							<p class="text-slate-400 font-black text-[10px] uppercase tracking-wider mb-2">
								Weekly Operational Weight (WIP)
							</p>
							<h3 class="font-black text-4xl text-primary tracking-tighter italic">
								{formatCurrency((data.position.base_pay || 0) * data.employees.length * 48)}
								<span class="text-xs font-black text-slate-300 uppercase not-italic ml-2"
									>/ Week</span
								>
							</h3>
						</div>
						<div
							class="bg-primary/5 p-4 rounded-sm group-hover:scale-110 transition-transform duration-300"
						>
							<Wallet class="h-6 w-6 text-primary" />
						</div>
					</div>

					<div class="flex items-center gap-6 bg-white border border-slate-100 p-5 rounded-sm">
						<div class="flex -space-x-3">
							{#each data.employees.slice(0, 3) as emp (emp.id)}
								<div
									class="w-10 h-10 rounded-sm border-2 border-white bg-primary flex items-center justify-center text-[10px] font-black text-white uppercase italic"
								>
									{emp.user_name[0]}
								</div>
							{/each}
							{#if data.employees.length > 3}
								<div
									class="w-10 h-10 rounded-sm border-2 border-white bg-slate-100 flex items-center justify-center text-[10px] font-black text-slate-400 uppercase tracking-tighter"
								>
									+{data.employees.length - 3}
								</div>
							{/if}
						</div>
						<div>
							<p class="text-slate-900 font-black text-sm uppercase italic leading-tight">
								{data.employees.length} Active Staff Members
							</p>
							<p class="text-[9px] font-black text-slate-400 uppercase tracking-widest mt-1">
								Current Roster Allocation
							</p>
						</div>
					</div>
				</div>
			</section>

			<!-- Penalty Engine Grid -->
			<section class="grid grid-cols-2 gap-4">
				<div class="bg-slate-50 rounded-sm p-6 border border-slate-100 group">
					<div class="flex items-center gap-2 mb-4">
						<div class="p-2 bg-rose-50 rounded-sm">
							<AlertTriangle class="h-4 w-4 text-rose-500" />
						</div>
						<span class="text-slate-400 font-black text-[9px] uppercase tracking-wider"
							>Late Entry</span
						>
					</div>
					<div class="font-black text-3xl text-slate-900 tracking-tighter italic mb-2">
						{formatCurrency(data.position.late_penalty || 0)}
					</div>
					<p class="text-[9px] font-black text-slate-400 uppercase tracking-widest leading-tight">
						Fixed deduction applied per occurrence.
					</p>
				</div>

				<div class="bg-slate-50 rounded-sm p-6 border border-slate-100 group">
					<div class="flex items-center gap-2 mb-4">
						<div class="p-2 bg-primary/5 rounded-sm">
							<ShieldAlert class="h-4 w-4 text-primary" />
						</div>
						<span class="text-slate-400 font-black text-[9px] uppercase tracking-wider"
							>Geocerca</span
						>
					</div>
					<div class="font-black text-3xl text-slate-900 tracking-tighter italic mb-2">
						{formatCurrency(data.position.out_of_range_penalty || 0)}
					</div>
					<p class="text-[9px] font-black text-slate-400 uppercase tracking-widest leading-tight">
						Violation of proximity perimeter perimeter.
					</p>
				</div>
			</section>

			<!-- Talent Roster -->
			<section class="space-y-6">
				<div class="flex items-center justify-between">
					<div class="flex items-center gap-3">
						<Users class="h-5 w-5 text-primary" />
						<h4 class="font-black text-xl text-primary uppercase italic tracking-tighter">
							Talent Roster
						</h4>
					</div>
					<Badge
						class="bg-slate-900 text-white font-black text-[9px] px-3 py-1 rounded-sm uppercase tracking-widest"
					>
						{data.employees.length} TOTAL
					</Badge>
				</div>

				<div class="space-y-3">
					{#each data.employees as emp (emp.id)}
						<div
							class="flex items-center justify-between p-4 rounded-sm bg-slate-50/50 hover:bg-slate-100 transition-all group border border-transparent hover:border-slate-200"
						>
							<div class="flex items-center gap-4">
								<div
									class="w-12 h-12 rounded-sm bg-white border border-slate-100 flex items-center justify-center font-black text-lg text-primary italic shadow-sm group-hover:scale-105 transition-transform"
								>
									{emp.user_name[0]}
								</div>
								<div>
									<p class="font-black text-sm text-slate-900 uppercase italic tracking-tight">
										{emp.user_name}
									</p>
									<p class="font-black text-[9px] text-slate-400 uppercase tracking-[0.2em] mt-1">
										{emp.center_name || 'Terminal HQ'}
									</p>
								</div>
							</div>
							<div class="flex flex-col items-end gap-2">
								<Badge
									variant="outline"
									class="h-6 rounded-full font-black text-[8px] uppercase tracking-widest border-none {emp.is_active
										? 'bg-emerald-50 text-emerald-600'
										: 'bg-slate-200 text-slate-600'}"
								>
									<div
										class="w-1 h-1 rounded-full mr-2 {emp.is_active
											? 'bg-emerald-500 animate-pulse'
											: 'bg-slate-400'}"
									></div>
									{emp.is_active ? 'ACTIVE' : 'OFF SHIFT'}
								</Badge>
							</div>
						</div>
					{:else}
						<div class="py-20 text-center border-2 border-dashed border-slate-100 rounded-sm">
							<p class="text-[10px] font-black text-slate-300 uppercase tracking-widest italic">
								No personnel assigned to this position
							</p>
						</div>
					{/each}
				</div>
			</section>
		</main>

		<!-- Bottom Navigation Bar -->
		<nav
			class="fixed bottom-0 w-full z-50 border-t border-slate-100 bg-white flex justify-around items-center h-20 px-4"
		>
			<button class="flex flex-col items-center justify-center gap-1 group">
				<Wallet class="h-5 w-5 text-primary transition-transform group-hover:-translate-y-1" />
				<span class="text-[9px] font-black text-primary uppercase tracking-widest">Financials</span>
				<div class="h-1 w-6 bg-primary rounded-full mt-1"></div>
			</button>
			<button class="flex flex-col items-center justify-center gap-1 group opacity-40">
				<Scale class="h-5 w-5 text-slate-400 transition-transform group-hover:-translate-y-1" />
				<span class="text-[9px] font-black text-slate-400 uppercase tracking-widest"
					>Analytics (WIP)</span
				>
			</button>
			<button class="flex flex-col items-center justify-center gap-1 group opacity-40">
				<ShieldAlert
					class="h-5 w-5 text-slate-400 transition-transform group-hover:-translate-y-1"
				/>
				<span class="text-[9px] font-black text-slate-400 uppercase tracking-widest"
					>Compliance (WIP)</span
				>
			</button>
			<button class="flex flex-col items-center justify-center gap-1 group opacity-40">
				<Activity class="h-5 w-5 text-slate-400 transition-transform group-hover:-translate-y-1" />
				<span class="text-[9px] font-black text-slate-400 uppercase tracking-widest"
					>Audit (WIP)</span
				>
			</button>
		</nav>
	{/if}
</div>

