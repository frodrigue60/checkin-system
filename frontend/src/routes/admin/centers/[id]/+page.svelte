<script lang="ts">
	import { onMount } from 'svelte';
	import { fly } from 'svelte/transition';
	import { quintOut } from 'svelte/easing';
	import { page } from '$app/state';
	import { apiFetch } from '$lib/api';
	import * as Table from '$lib/components/ui/table/index.js';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { Avatar, AvatarFallback } from '$lib/components/ui/avatar/index.js';
	import {
		Loader2,
		LogIn,
		LogOut,
		History,
		Mail,
		Phone,
		AlertCircle,
		ShieldCheck,
		Users,
		Target,
		User,
		Maximize2,
		MapPin
	} from 'lucide-svelte';

	let centerId = $derived(page.params.id);
	let data = $state<any>(null);
	let loading = $state(true);
	let errorMsg = $state('');

	let showMapModal = $state(false);

	let map: any;
	let modalMap: any;

	// Preview Map Effect (Static)
	$effect(() => {
		if (data && data.center) {
			const timer = setTimeout(() => {
				const mapEl = document.getElementById('details-map');
				if (mapEl && typeof L !== 'undefined') {
					if (map) map.remove();
					map = L.map('details-map', {
						zoomControl: false,
						attributionControl: false,
						dragging: false,
						touchZoom: false,
						scrollWheelZoom: false,
						doubleClickZoom: false
					}).setView([data.center.latitude, data.center.longitude], 16);

					L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png').addTo(map);

					L.circle([data.center.latitude, data.center.longitude], {
						color: '#3b82f6',
						fillColor: '#3b82f6',
						fillOpacity: 0.15,
						weight: 2,
						radius: data.center.tolerance_radius
					}).addTo(map);

					L.circleMarker([data.center.latitude, data.center.longitude], {
						radius: 5,
						color: '#fff',
						weight: 2,
						fillColor: '#3b82f6',
						fillOpacity: 1
					}).addTo(map);
				}
			}, 300);
			return () => {
				clearTimeout(timer);
				if (map) {
					map.remove();
					map = null;
				}
			};
		}
	});

	// Expanded Modal Map Effect (Interactive)
	$effect(() => {
		if (showMapModal && data?.center) {
			const timer = setTimeout(() => {
				const mapEl = document.getElementById('expanded-map');
				if (mapEl && typeof L !== 'undefined') {
					if (modalMap) modalMap.remove();
					modalMap = L.map('expanded-map').setView(
						[data.center.latitude, data.center.longitude],
						17
					);

					L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
						attribution: '&copy; OpenStreetMap contributors'
					}).addTo(modalMap);

					L.circle([data.center.latitude, data.center.longitude], {
						color: '#3b82f6',
						fillColor: '#3b82f6',
						fillOpacity: 0.1,
						weight: 1,
						radius: data.center.tolerance_radius
					}).addTo(modalMap);

					L.marker([data.center.latitude, data.center.longitude])
						.addTo(modalMap)
						.bindPopup(data.center.name)
						.openPopup();
				}
			}, 100);

			return () => {
				clearTimeout(timer);
				if (modalMap) {
					modalMap.remove();
					modalMap = null;
				}
			};
		}
	});

	function formatHM(timeStr: string | null) {
		if (
			!timeStr ||
			timeStr === '' ||
			timeStr.includes('0001-01-01') ||
			timeStr.includes('0000-01-01')
		)
			return '--:--';
		try {
			const d = new Date(timeStr);
			if (isNaN(d.getTime())) return '--:--';
			return d.toLocaleTimeString('es-MX', { hour: '2-digit', minute: '2-digit', hour12: false });
		} catch (e) {
			return '--:--';
		}
	}

	async function loadDetails() {
		loading = true;
		const res = await apiFetch(`/admin/centers/${centerId}/details`);
		if (res.ok) {
			data = await res.json();
		} else {
			errorMsg = 'No se pudo cargar la información del centro';
		}
		loading = false;
	}

	onMount(loadDetails);
</script>

<svelte:head>
	<link
		rel="stylesheet"
		href="https://unpkg.com/leaflet@1.9.4/dist/leaflet.css"
		integrity="sha256-p4NxAoJBhIIN+hmNHrzRCf9tD/miZyoHS5obTRR9BMY="
		crossorigin=""
	/>
	<script
		src="https://unpkg.com/leaflet@1.9.4/dist/leaflet.js"
		integrity="sha256-20nQCchB9co0qIjJZRGuk2/Z9VM+kNiyxNV1lvTlZBo="
		crossorigin=""
	></script>
</svelte:head>

<div class="min-h-screen pb-24 font-sans bg-background">
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
					Sincronizando Matriz...
				</p>
				<p class="text-[10px] font-black text-muted-foreground uppercase tracking-[0.3em]">
					Extrayendo coordenadas geoespaciales
				</p>
			</div>
		</div>
	{:else if errorMsg}
		<div class="max-w-md mx-auto p-12 text-center space-y-6">
			<div
				class="h-20 w-20 rounded-sm bg-white shadow-xl shadow-rose-200 flex items-center justify-center mx-auto"
			>
				<span class="text-4xl text-rose-500">⚠️</span>
			</div>
			<div class="space-y-2">
				<h2 class="text-3xl font-black text-rose-900 tracking-tighter">{errorMsg}</h2>
				<p class="font-bold text-rose-600/60 uppercase text-[10px] tracking-widest">
					Error de Referencia 404
				</p>
			</div>
			<Button variant="outline" href="/admin/centers" class="rounded-sm font-black h-12 px-8">
				Volver a la Terminal
			</Button>
		</div>
	{:else if data}
		<main
			class="px-6 py-10 max-w-5xl mx-auto space-y-10"
			in:fly={{ y: 20, duration: 800, easing: quintOut }}
		>
			<!-- Header & Identity -->
			<section class="flex flex-col md:flex-row justify-between items-start md:items-end gap-8">
				<div class="space-y-4">
					<div class="space-y-1">
						<div class="flex items-center gap-3">
							<div
								class="w-2.5 h-2.5 rounded-full bg-emerald-500 pulse-dot shadow-lg shadow-emerald-200"
							></div>
							<span class="text-[10px] font-black uppercase tracking-[0.2em] text-muted-foreground"
								>Nodo Activo</span
							>
						</div>
						<h1 class="text-6xl font-black tracking-tighter text-slate-900 leading-tight">
							{data.center.name}
						</h1>
						<h2
							class="text-2xl font-black tracking-tighter text-slate-900 uppercase italic opacity-70"
						>
							{data.center.address}
						</h2>
						<p
							class="text-muted-foreground font-bold italic border-l-2 border-primary/20 pl-4 py-2 mt-2"
						>
							{data.center.description ?? 'Sin descripción disponible'}
						</p>
					</div>
				</div>
			</section>

			<!-- Bento Grid Layout -->
			<div class="grid grid-cols-1 md:grid-cols-12 gap-6">
				<!-- KPI Card: Personnel Capacity -->
				<div
					class="md:col-span-8 bg-slate-900 rounded-sm p-8 text-white flex justify-between items-center relative overflow-hidden group"
				>
					<div class="relative z-10 space-y-2">
						<span
							class="text-[10px] font-black uppercase tracking-[0.3em] text-primary transition-colors group-hover:text-primary-fixed-dim"
						>
							Personnel Capacity
						</span>
						<div class="flex items-baseline gap-3">
							<span class="text-7xl font-black tracking-tighter">
								{data.employees.length}
							</span>
							<span class="text-2xl font-black text-slate-500 uppercase tracking-tighter italic"
								>/ Nominal</span
							>
						</div>
						<p class="text-[10px] font-black uppercase tracking-widest text-slate-400">
							Sincronización en tiempo real con la nómina activa
						</p>
					</div>
					<div
						class="relative z-10 h-24 w-24 rounded-sm bg-slate-50 border border-slate-200 flex items-center justify-center group-hover:scale-105 transition-transform duration-500"
					>
						<Users class="h-10 w-10 text-primary" />
					</div>
					<!-- Background decoration -->
					<div
						class="absolute -right-10 -bottom-10 w-64 h-64 bg-primary/10 rounded-full blur-3xl transition-all group-hover:bg-primary/20"
					></div>
				</div>

				<!-- Snapshot Card: Manager -->
				<div
					class="md:col-span-4 bg-white rounded-sm p-6 flex flex-col justify-between border border-slate-100 shadow-xl shadow-slate-200/50"
				>
					<div class="space-y-4">
						<span
							class="text-[10px] font-black uppercase tracking-widest text-muted-foreground flex items-center gap-2"
						>
							<ShieldCheck class="h-3 w-3" /> Custodio del Nodo
						</span>
						{#if data.manager}
							<div class="flex items-center gap-4">
								<Avatar class="h-14 w-14 rounded-sm border-2 border-slate-50 shadow-sm">
									<AvatarFallback
										class="bg-slate-100 text-slate-900 font-black text-xl rounded-sm lowercase"
									>
										{data.manager.name.charAt(0)}
									</AvatarFallback>
								</Avatar>
								<div class="space-y-0.5">
									<h3 class="text-lg font-black tracking-tight text-slate-900 leading-none">
										{data.manager.name}
									</h3>
									<p class="text-[9px] font-black text-primary uppercase tracking-widest italic">
										{data.manager.email}
									</p>
								</div>
							</div>
						{:else}
							<div
								class="p-6 rounded-sm bg-slate-50 border border-dashed border-slate-200 text-center"
							>
								<p class="text-[9px] font-black text-muted-foreground uppercase tracking-widest">
									Sin gestión asignada
								</p>
							</div>
						{/if}
					</div>
					<div class="mt-4 flex gap-2">
						<Button
							variant="outline"
							size="sm"
							class="flex-1 rounded-sm border-slate-100 bg-slate-50/50 text-[10px] font-black uppercase tracking-widest"
						>
							<Mail class="h-3 w-3 mr-2" /> Email
						</Button>
						<Button
							variant="outline"
							size="sm"
							class="flex-1 rounded-sm border-slate-100 bg-slate-50/50 text-[10px] font-black uppercase tracking-widest"
						>
							<Phone class="h-3 w-3 mr-2" /> Call
						</Button>
					</div>
				</div>

				<!-- Geofence Card -->
				<button
					type="button"
					onclick={() => (showMapModal = true)}
					class="md:col-span-4 bg-white rounded-sm overflow-hidden border border-slate-100 shadow-xl shadow-slate-200/50 flex flex-col text-left group transition-all hover:scale-[1.02] active:scale-95"
				>
					<div class="sm:h-40 md:h-full w-full relative bg-slate-50">
						<!-- Live Leaflet Map Preview -->
						<div id="details-map" class="h-full w-full grayscale-[0.2] contrast-[1.1]"></div>
						<div
							class="absolute inset-0 bg-gradient-to-tr from-primary/5 to-transparent pointer-events-none z-[400] opacity-0 group-hover:opacity-100 transition-opacity"
						></div>

						<!-- Overlay Controls -->
						<div
							class="absolute top-4 left-4 bg-white px-3 py-1.5 rounded-sm flex items-center gap-2 shadow-sm border border-slate-100 z-[401]"
						>
							<Target class="h-3.5 w-3.5 text-primary" />
							<span class="text-[10px] font-black uppercase tracking-widest text-slate-900">
								{data.center.tolerance_radius}m Radius
							</span>
						</div>
						<div
							class="absolute top-4 right-4 bg-slate-900/80 text-white p-2 rounded-sm opacity-0 group-hover:opacity-100 transition-all z-[401] shadow-xl"
						>
							<Maximize2 size={14} />
						</div>
					</div>
					<div class="p-5 space-y-1">
						<span class="text-[10px] font-black uppercase text-muted-foreground tracking-widest"
							>Perímetro Digital</span
						>
						<p class="text-sm font-black text-slate-900 italic tracking-tight">
							Geofencing Crítico Activo
						</p>
					</div>
				</button>

				<!-- Activity Log -->
				<div
					class="md:col-span-8 bg-white rounded-sm p-8 shadow-2xl shadow-slate-200/50 border border-slate-100 space-y-8"
				>
					<div class="flex justify-between items-center pb-6 border-b border-slate-50">
						<div>
							<h3 class="text-2xl font-black tracking-tighter text-slate-900 uppercase">
								Operational Chronology
							</h3>
							<p
								class="text-[9px] font-black text-muted-foreground uppercase tracking-[0.2em] mt-1 italic"
							>
								Last synchronized events
							</p>
						</div>
						<History class="h-5 w-5 text-slate-300" />
					</div>

					<div class="space-y-6">
						{#each data.recent_attendance.slice(0, 5) as log (log.id)}
							<div class="flex items-center gap-5 group">
								<div
									class="h-12 w-12 rounded-sm bg-slate-50 border border-slate-100 flex items-center justify-center shrink-0 group-hover:bg-primary/5 group-hover:border-primary/20 transition-all duration-300"
								>
									{#if log.check_out}
										<LogOut class="h-5 w-5 text-slate-400 group-hover:text-primary" />
									{:else}
										<LogIn class="h-5 w-5 text-emerald-500 animate-pulse" />
									{/if}
								</div>
								<div class="flex-1 space-y-0.5">
									<div class="flex items-center gap-2">
										<p class="text-base font-black tracking-tight text-slate-900 leading-none">
											{log.employee_name}
										</p>
										{#if !log.check_out}
											<Badge
												class="h-4 px-2 py-0 bg-emerald-100 text-emerald-700 border-none font-black text-[8px] uppercase tracking-widest"
												>Live</Badge
											>
										{/if}
									</div>
									<p class="text-[9px] font-black text-slate-400 uppercase tracking-widest">
										{new Date(log.check_in).toLocaleDateString('es-MX', {
											day: '2-digit',
											month: 'short'
										})} • Terminal {centerId} • {formatHM(log.check_in)}
									</p>
								</div>
								<div class="text-right">
									<p class="text-[10px] font-black text-slate-900 uppercase tabular-nums">
										{log.net_hours_worked ? log.net_hours_worked.toFixed(2) : '--'}
									</p>
									<p class="text-[8px] font-black text-slate-300 uppercase tracking-tighter">
										Hours
									</p>
								</div>
							</div>
							<div class="h-px bg-slate-100 last:hidden"></div>
						{/each}

						{#if data.recent_attendance.length === 0}
							<div class="py-12 text-center space-y-3">
								<AlertCircle class="h-10 w-10 text-slate-100 mx-auto" />
								<p class="text-[10px] font-black text-slate-300 uppercase tracking-widest">
									No activity logged in current cycle
								</p>
							</div>
						{/if}
					</div>

					<Button
						variant="outline"
						class="w-full h-12 rounded-sm border-slate-100 text-[10px] font-black uppercase tracking-widest hover:bg-slate-900 hover:text-white transition-all"
					>
						View Full Audit Archive
					</Button>
				</div>
			</div>

			<!-- Team Assigned -->
			<section class="space-y-6 pt-10 border-t border-slate-100">
				<div class="flex items-end justify-between px-2">
					<div class="space-y-1">
						<h3 class="text-3xl font-black text-slate-900 tracking-tighter italic">
							Equipo <span class="text-primary not-italic">Asignado</span>
						</h3>
						<p class="text-[10px] font-black text-muted-foreground uppercase tracking-widest">
							Personal operativo fijo en sede
						</p>
					</div>
					<Badge
						variant="secondary"
						class="rounded-sm px-4 py-1.5 bg-slate-100 text-slate-900 font-black text-[10px] shadow-sm border-none"
					>
						{data.employees.length} INTEGRANTES
					</Badge>
				</div>

				<div
					class="border-none shadow-2xl shadow-slate-200/50 rounded-sm bg-white overflow-hidden min-h-[300px]"
				>
					<Table.Root>
						<Table.Header class="bg-slate-50/50 border-b border-slate-100">
							<Table.Row class="hover:bg-transparent border-none">
								<Table.Head
									class="h-14 px-8 text-[9px] font-black uppercase tracking-widest text-slate-400"
									>Colaborador / Perfil</Table.Head
								>
								<Table.Head
									class="h-14 px-8 text-[9px] font-black uppercase tracking-widest text-slate-400 text-center"
									>Turno</Table.Head
								>
								<Table.Head
									class="h-14 px-8 text-[9px] font-black uppercase tracking-widest text-slate-400 text-right"
									>Estado</Table.Head
								>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#each data.employees as emp (emp.id)}
								<Table.Row class="group border-slate-50 hover:bg-slate-50/50 transition-colors">
									<Table.Cell class="px-8 py-4">
										<div class="flex items-center gap-4">
											<div
												class="h-10 w-10 rounded-sm bg-slate-100 flex items-center justify-center group-hover:bg-primary/10 transition-colors"
											>
												<User class="h-4 w-4 text-slate-400 group-hover:text-primary" />
											</div>
											<div class="space-y-0.5">
												<p class="text-base font-black tracking-tight text-slate-900 leading-none">
													{emp.user_name}
												</p>
												<p
													class="text-[9px] font-black text-slate-400 uppercase tracking-widest italic"
												>
													{emp.position_name}
												</p>
											</div>
										</div>
									</Table.Cell>
									<Table.Cell class="px-8 py-4 text-center">
										<Badge
											variant="outline"
											class="border-slate-100 bg-white font-black text-[9px] px-3 py-1 rounded-sm"
										>
											{emp.shift_name || 'SIN ASIGNAR'}
										</Badge>
									</Table.Cell>
									<Table.Cell class="px-8 py-4 text-right">
										{#if emp.is_active}
											<span
												class="h-2 w-2 rounded-full bg-emerald-500 inline-block shadow-lg shadow-emerald-200"
											></span>
										{:else}
											<span class="h-2 w-2 rounded-full bg-rose-500 inline-block"></span>
										{/if}
									</Table.Cell>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				</div>
			</section>
		</main>
	{/if}
</div>

<!-- Map Modal -->
{#if showMapModal}
	<Dialog.Root
		open={showMapModal}
		onOpenChange={(o) => {
			if (!o) showMapModal = false;
		}}
	>
		<Dialog.Portal>
			<Dialog.Content
				class="rounded-md border-none p-12 sm:max-w-full md:max-w-4xl max-h-[90vh] bg-white shadow-2xl"
			>
				<Dialog.Header class="mb-8">
					<div class="flex items-center gap-4">
						<div
							class="h-12 w-12 rounded-md bg-primary/10 text-primary flex items-center justify-center"
						>
							<MapPin class="h-6 w-6" />
						</div>
						<div class="space-y-1 text-left">
							<Dialog.Title class="text-3xl font-black tracking-tighter text-slate-900">
								Monitor Geoespacial
							</Dialog.Title>
							<Dialog.Description
								class="text-xs font-black text-slate-400 uppercase tracking-widest"
							>
								{data.center.address ?? 'Sin dirección registrada'}
							</Dialog.Description>
						</div>
					</div>
				</Dialog.Header>

				<div
					class="relative w-full aspect-video rounded-md overflow-hidden bg-slate-100 border-2 border-slate-50 shadow-inner"
				>
					<div id="expanded-map" class="h-full w-full"></div>
				</div>

				<div class="flex justify-end pt-8">
					<Button
						variant="secondary"
						class="rounded-sm font-black gap-2 h-12 px-8 bg-slate-900 text-white hover:bg-slate-800 uppercase text-[10px] tracking-widest shadow-xl shadow-slate-900/20 transition-all border-none"
						onclick={() => (showMapModal = false)}
					>
						Finalizar Auditoría
					</Button>
				</div>
			</Dialog.Content>
		</Dialog.Portal>
	</Dialog.Root>
{/if}

<style>
	:global(.no-scrollbar::-webkit-scrollbar) {
		display: none;
	}
	:global(.no-scrollbar) {
		-ms-overflow-style: none;
		scrollbar-width: none;
	}

	:global(.leaflet-container) {
		background: #f8fafc !important;
	}
</style>
