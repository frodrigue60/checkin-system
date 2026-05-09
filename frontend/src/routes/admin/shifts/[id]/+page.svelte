<script lang="ts">
	import { onMount } from 'svelte';
	import { _ } from 'svelte-i18n';
	import { fly } from 'svelte/transition';
	import { quintOut } from 'svelte/easing';
	import { page } from '$app/state';
	import { apiFetch } from '$lib/api';
	import { Button } from '$lib/components/ui/button/index.js';
	import Badge from '$lib/components/ui/badge/badge.svelte';
	import { Avatar, AvatarFallback } from '$lib/components/ui/avatar/index.js';
	import {
		Loader2,
		LogIn,
		LogOut,
		History,
		AlertCircle,
		Users,
		Clock,
		Moon,
		Sun,
		Calendar,
		ArrowLeft,
		Activity,
		MoreVertical,
		User,
		ShieldCheck,
		MapPin
	} from 'lucide-svelte';

	let shiftId = $derived(page.params.id);
	let data = $state<any>(null);
	let loading = $state(true);
	let errorMsg = $state('');

	// Helper for time strings
	function formatTime(val: string) {
		if (!val || val === '') return '--:--';
		
		// If it's a full ISO string
		if (val.includes('T')) {
			const timePart = val.split('T')[1].substring(0, 5);
			// Only return --:-- if it's exactly the zero time of Go (0001-01-01 00:00)
			if (timePart === '00:00' && (val.includes('0001-01-01') || val.includes('0000-01-01'))) {
				// Special case: if we explicitly want 00:00 in a shift, this might be tricky,
				// but usually 00:00:00Z on year 1 is the null value.
				// For now, let's return the timePart anyway if it's a shift.
				return timePart;
			}
			return timePart;
		}
		return val.substring(0, 5);
	}

	function timeToMinutes(timeStr: string) {
		if (!timeStr) return 0;
		const clean = formatTime(timeStr);
		if (clean === '--:--') return 0;
		const parts = clean.split(':').map((val) => parseInt(val, 10));
		if (parts.length < 2 || isNaN(parts[0]) || isNaN(parts[1])) return 0;
		return parts[0] * 60 + parts[1];
	}

	function calculateDuration(start: string, end: string) {
		const sMin = timeToMinutes(start);
		const eMin = timeToMinutes(end);
		let diff = eMin - sMin;
		if (diff < 0) diff += 24 * 60; // Crosses midnight
		const h = Math.floor(diff / 60);
		const m = diff % 60;
		return `${h.toString().padStart(2, '0')}:${m.toString().padStart(2, '0')}h`;
	}

	// Timeline Helpers
	function getMarkerPosition(timeStr: string, start: string, end: string) {
		const sMin = timeToMinutes(start);
		const eMin = timeToMinutes(end);
		let targetMin = timeToMinutes(timeStr);

		let totalRange = eMin - sMin;
		if (totalRange <= 0) totalRange += 24 * 60;

		let relativePos = targetMin - sMin;
		if (relativePos < 0) relativePos += 24 * 60;

		return Math.min(100, Math.max(0, (relativePos / totalRange) * 100));
	}

	async function loadDetails() {
		loading = true;
		errorMsg = '';
		try {
			const res = await apiFetch(`/admin/shifts/${shiftId}/details`);
			if (res.ok) {
				data = await res.json();
			} else {
				errorMsg = 'No se pudo cargar la información del turno';
			}
		} catch (e) {
			console.error('Error loading shift details:', e);
			errorMsg = 'Error de conexión con el servidor';
		} finally {
			loading = false;
		}
	}

	onMount(loadDetails);
</script>

<div class="min-h-screen pb-32 font-sans bg-background">
	{#if loading}
		<div class="flex flex-col items-center justify-center min-h-[80vh] gap-6">
			<div class="relative">
				<div class="h-24 w-24 rounded-sm border-4 border-slate-100 animate-pulse"></div>
				<Loader2
					class="h-10 w-10 animate-spin text-primary absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2"
				/>
			</div>
			<div class="space-y-2 text-center">
				<p class="text-2xl font-black tracking-tighter text-slate-900 italic">
					Sincronizando Jornada...
				</p>
				<p class="text-[10px] font-black text-muted-foreground uppercase tracking-[0.3em]">
					Analizando parámetros operativos
				</p>
			</div>
		</div>
	{:else if errorMsg}
		<div class="max-w-md mx-auto p-12 text-center space-y-6">
			<div
				class="h-20 w-20 rounded-sm bg-white shadow-xl shadow-rose-200 flex items-center justify-center mx-auto"
			>
				<AlertCircle class="h-10 w-10 text-rose-500" />
			</div>
			<div class="space-y-2">
				<h2 class="text-3xl font-black text-rose-900 tracking-tighter">{errorMsg}</h2>
				<p class="font-bold text-rose-600/60 uppercase text-[10px] tracking-widest">
					Error de Referencia 404
				</p>
			</div>
			<Button variant="outline" href="/admin/shifts" class="rounded-sm font-black h-12 px-8">
				Volver a la Terminal
			</Button>
		</div>
	{:else if data}
		<main class="pt-6 px-6 max-w-7xl mx-auto space-y-12">
			<!-- Hero Header Section -->
			<section
				class="grid grid-cols-1 sm:grid-cols-1 md:grid-cols-2 md:items-end justify-between gap-8"
				in:fly={{ y: 20, duration: 800, easing: quintOut }}
			>
				<div class="space-y-2">
					<span class="text-[10px] font-black uppercase tracking-[0.3em] text-slate-400"
						>Total Shift Duration</span
					>
					<h2 class="text-7xl md:text-8xl font-black text-primary tracking-tighter leading-none">
						{calculateDuration(data.shift.start_time, data.shift.end_time)}
					</h2>
				</div>
				<div class="flex flex-col items-end gap-4">
					<!-- Night Shift Indicator -->
					<div
						class="flex items-center gap-2 px-4 py-2 rounded-full {data.shift.is_night_shift
							? 'bg-slate-900 text-white'
							: 'bg-amber-50 text-amber-700'} shadow-sm"
					>
						{#if data.shift.is_night_shift}
							<Moon class="h-3.5 w-3.5" />
							<span class="text-[10px] font-black uppercase tracking-widest"
								>Night Shift Active</span
							>
						{:else}
							<Sun class="h-3.5 w-3.5" />
							<span class="text-[10px] font-black uppercase tracking-widest">Day Cycle Active</span>
						{/if}
					</div>
					<div class="text-right space-y-2">
						<div class="flex justify-end gap-2">
							{#if data.shift.shift_type === 'field'}
								<Badge class="bg-amber-500 text-white border-none font-black text-[9px] tracking-widest px-3 py-1 flex items-center gap-1">
									<MapPin size={10} />
									{$_('admin.shifts.type_field')}
								</Badge>
							{:else if data.shift.shift_type === 'flexible'}
								<Badge class="bg-indigo-500 text-white border-none font-black text-[9px] tracking-widest px-3 py-1 flex items-center gap-1">
									<Activity size={10} />
									{$_('admin.shifts.type_flexible')}
								</Badge>
							{:else}
								<Badge class="bg-slate-900 text-white border-none font-black text-[9px] tracking-widest px-3 py-1 flex items-center gap-1">
									<ShieldCheck size={10} />
									{$_('admin.shifts.type_fixed')}
								</Badge>
							{/if}
						</div>
						<h3 class="text-3xl font-black text-slate-900 tracking-tighter">{data.shift.name}</h3>
						<p
							class="text-[10px] font-black text-muted-foreground uppercase tracking-widest flex items-center justify-end gap-2"
						>
							<Calendar class="h-3 w-3" /> Operational Window
						</p>
					</div>
				</div>
			</section>

			<!-- Bento Grid Layout -->
			<div class="grid grid-cols-1 md:grid-cols-12 gap-6">
				<!-- Card 1: Time Matrix (Large Span) -->
				<div
					class="md:col-span-12 lg:col-span-8 bg-white border border-slate-100 rounded-sm p-8 flex flex-col justify-between min-h-[400px] shadow-xl shadow-slate-200/50"
					in:fly={{ y: 20, duration: 800, easing: quintOut, delay: 100 }}
				>
					<div>
						<h3 class="text-3xl font-black text-primary tracking-tighter uppercase mb-1 italic">
							Time Matrix
						</h3>
						<p class="text-xs font-bold text-slate-400 uppercase tracking-widest">
							Visual operational window for the current rotation.
						</p>
					</div>

					<div class="relative w-full pt-16 pb-12">
						<!-- Timeline Track -->
						<div class="h-2 w-full bg-slate-50 rounded-full relative overflow-visible">
							<!-- Lunch Window Range (If applicable, here it's visualized as a duration) -->
							<div
								class="absolute left-[40%] right-[40%] h-full bg-emerald-500/10 rounded-none border-x border-emerald-500/20"
							></div>

							<!-- Start Marker -->
							<div class="absolute left-0 top-1/2 -translate-y-1/2 flex flex-col items-center">
								<div class="h-5 w-5 bg-primary rounded-full border-4 border-white shadow-lg"></div>
								<span class="absolute -top-10 font-black text-xs text-primary tabular-nums"
									>{formatTime(data.shift.start_time)}</span
								>
								<span
									class="absolute top-8 text-[9px] font-black text-slate-400 uppercase tracking-widest"
									>Check-in</span
								>
							</div>

							<!-- End Marker -->
							<div class="absolute right-0 top-1/2 -translate-y-1/2 flex flex-col items-center">
								<div
									class="h-5 w-5 bg-slate-300 rounded-full border-4 border-white shadow-lg"
								></div>
								<span class="absolute -top-10 font-black text-xs text-slate-400 tabular-nums"
									>{formatTime(data.shift.end_time)}</span
								>
								<span
									class="absolute top-8 text-[9px] font-black text-slate-400 uppercase tracking-widest text-right"
									>Check-out</span
								>
							</div>
						</div>
					</div>

					<div class="grid grid-cols-2 md:grid-cols-4 gap-4 mt-8">
						<div class="bg-slate-900 p-5 rounded-sm shadow-md border-b-4 border-primary text-white">
							<span
								class="text-[9px] font-black text-slate-500 uppercase block mb-1 tracking-widest"
								>Elapsed (WIP)</span
							>
							<span class="text-2xl font-black tabular-nums"
								>04:22 <span class="text-xs uppercase opacity-40">Hrs</span></span
							>
						</div>
						<div class="bg-slate-50 p-5 rounded-sm border border-slate-100">
							<span
								class="text-[9px] font-black text-slate-400 uppercase block mb-1 tracking-widest"
								>Remaining (WIP)</span
							>
							<span class="text-2xl font-black tabular-nums"
								>03:38 <span class="text-xs uppercase opacity-40">Hrs</span></span
							>
						</div>
						<div class="bg-slate-50 p-5 rounded-sm border border-slate-100">
							<span
								class="text-[9px] font-black text-slate-400 uppercase block mb-1 tracking-widest"
								>Assignments</span
							>
							<span class="text-2xl font-black tabular-nums"
								>{data.employees.length}
								<span class="text-xs uppercase opacity-40">Agents</span></span
							>
						</div>
						<div class="bg-slate-50 p-5 rounded-sm border border-slate-100">
							<span
								class="text-[9px] font-black text-slate-400 uppercase block mb-1 tracking-widest"
								>Overtime (WIP)</span
							>
							<span class="text-2xl font-black text-rose-500 italic">None</span>
						</div>
					</div>
				</div>

				<!-- Card 2: Compliance (Circle Chart) -->
				<!-- WIP: This UI is currently static until backend implementation of shift-specific compliance aggregates -->
				<div
					class="md:col-span-12 lg:col-span-4 bg-white border border-slate-100 p-10 rounded-sm shadow-xl shadow-slate-200/50 flex flex-col items-center justify-center text-center group"
					in:fly={{ y: 20, duration: 800, easing: quintOut, delay: 200 }}
				>
					<h3
						class="text-xl font-black text-primary uppercase tracking-tighter self-start italic mb-10"
					>
						Operational Pulse (WIP)
					</h3>
					<div class="relative w-56 h-56 flex items-center justify-center mb-8">
						<svg class="w-full h-full -rotate-90">
							<circle
								class="text-slate-50"
								cx="112"
								cy="112"
								fill="transparent"
								r="100"
								stroke="currentColor"
								stroke-width="12"
							></circle>
							<circle
								class="text-primary transition-all duration-1000"
								cx="112"
								cy="112"
								fill="transparent"
								r="100"
								stroke="currentColor"
								stroke-dasharray="628.3"
								stroke-dashoffset="37.7"
								stroke-width="12"
							></circle>
						</svg>
						<div class="absolute inset-0 flex flex-col items-center justify-center">
							<span
								class="text-6xl font-black text-primary tracking-tighter transition-transform group-hover:scale-110"
								>94%</span
							>
							<span class="text-[10px] font-black text-slate-400 uppercase tracking-widest mt-1"
								>Punctuality</span
							>
						</div>
					</div>
					<p class="text-xs font-bold text-slate-500 leading-relaxed max-w-[200px]">
						Performance for the last 30 days remains above the target of 90%.
					</p>
				</div>

				<!-- Card 3: Operative Roster -->
				<div
					class="md:col-span-12 lg:col-span-6 bg-white border border-slate-100 rounded-sm p-8 shadow-xl shadow-slate-200/50"
					in:fly={{ y: 20, duration: 800, easing: quintOut, delay: 300 }}
				>
					<div class="flex items-center justify-between mb-8">
						<h3 class="text-2xl font-black text-primary uppercase tracking-tighter italic">
							Operative Roster
						</h3>
						<Badge
							class="bg-primary/10 text-primary border-none font-black text-[9px] tracking-widest px-3 py-1"
							>{data.employees.length} INTEGRANTES</Badge
						>
					</div>
					<div class="space-y-4 max-h-[400px] overflow-y-auto no-scrollbar pr-2">
						{#each data.employees as emp (emp.id)}
							<div
								class="flex items-center justify-between p-4 bg-slate-50/50 rounded-sm hover:bg-slate-100 transition-colors group"
							>
								<div class="flex items-center gap-4">
									<Avatar
										class="h-10 w-10 border-2 border-white shadow-sm ring-2 ring-primary/5 rounded-sm"
									>
										<AvatarFallback
											class="bg-primary text-white font-black text-[10px] rounded-sm uppercase"
											>{emp.user_name[0]}</AvatarFallback
										>
									</Avatar>
									<div>
										<p class="font-black text-slate-900 leading-none text-sm uppercase italic">
											{emp.user_name}
										</p>
										<p class="text-[9px] font-black text-slate-400 uppercase tracking-widest mt-1">
											{emp.position_name || 'Agente'}
										</p>
									</div>
								</div>
								<div class="text-right">
									<span
										class="text-[8px] font-black text-primary bg-primary/5 px-2 py-1 rounded uppercase tracking-tighter"
										>{emp.center_name || 'HQ'}</span
									>
								</div>
							</div>
						{:else}
							<div class="py-12 text-center border border-dashed border-slate-200 rounded-sm">
								<p class="text-[10px] font-black text-slate-300 uppercase tracking-widest">
									Sin personal asignado
								</p>
							</div>
						{/each}
					</div>
				</div>

				<!-- Card 4: Recent Feed -->
				<div
					class="md:col-span-12 lg:col-span-6 bg-slate-900 text-white rounded-sm p-8 shadow-2xl shadow-slate-900/20"
					in:fly={{ y: 20, duration: 800, easing: quintOut, delay: 400 }}
				>
					<div class="flex items-center justify-between mb-8">
						<h3 class="text-2xl font-black text-primary uppercase tracking-tighter italic">
							Recent Audit Feed
						</h3>
						<Activity class="h-5 w-5 text-white/20" />
					</div>
					<div
						class="relative pl-6 space-y-8 before:content-[''] before:absolute before:left-[5px] before:top-2 before:bottom-2 before:w-[1px] before:bg-slate-200"
					>
						{#each data.recent_attendance.slice(0, 5) as log}
							<div class="relative group">
								{#if log.check_out}
									<div
										class="absolute -left-[25px] top-1 w-3 h-3 bg-slate-700 rounded-full border-2 border-slate-900 group-hover:bg-primary transition-colors"
									></div>
								{:else}
									<div
										class="absolute -left-[25px] top-1 w-3 h-3 bg-emerald-500 rounded-full border-2 border-slate-900 animate-pulse"
									></div>
								{/if}

								<div class="flex justify-between items-start">
									<div>
										<p class="font-black text-sm text-white uppercase italic tracking-tight italic">
											{log.check_out ? 'Cycle Closed' : 'Inbound Entry'}: {log.employee_name}
										</p>
										<p class="text-[9px] font-black text-slate-500 uppercase tracking-widest mt-1">
											Terminal {log.work_center_id} • {log.check_out ? 'Departure' : 'Arrival'}
										</p>
									</div>
									<span class="text-[9px] font-black text-primary tabular-nums"
										>{formatTime(log.check_in)}</span
									>
								</div>
							</div>
						{:else}
							<div class="py-12 text-center">
								<p class="text-[10px] font-black text-slate-600 uppercase tracking-widest italic">
									No events recorded in current window
								</p>
							</div>
						{/each}
					</div>

					<Button
						variant="outline"
						class="w-full mt-10 h-12 bg-slate-50 border-slate-200 hover:bg-white hover:text-slate-900 text-[10px] font-black uppercase tracking-widest rounded-sm transition-all"
					>
						Audit Full Activity
					</Button>
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
