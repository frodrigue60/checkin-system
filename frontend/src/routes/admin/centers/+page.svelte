<script lang="ts">
	import { onMount } from 'svelte';
	import { fly } from 'svelte/transition';
	import { quintOut } from 'svelte/easing';
	import { _ } from 'svelte-i18n';
	import { apiFetch } from '$lib/api';
	import { authState } from '$lib/auth.svelte';
	import type { WorkCenter } from '$lib/types/models';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import {
		MapPin,
		Pencil,
		Trash2,
		ArrowLeft,
		Plus,
		Search,
		Navigation,
		Settings2,
		Building2,
		Eye,
		ChevronRight,
		Loader2,
		Link,
		Zap,
		MoreHorizontal,
		Activity
	} from 'lucide-svelte';

	let centers = $state<WorkCenter[]>([]);
	let loading = $state(true);
	let showModal = $state(false);
	let editingItem = $state<WorkCenter | null>(null);
	let mounted = $state(false);

	let showMapModal = $state(false);
	let mapCoords = $state({ lat: 0, lng: 0, name: '', radius: 100 });

	let map: any;
	$effect(() => {
		if (showMapModal && mapCoords.lat !== 0) {
			// Small delay to ensure DOM is ready
			const timer = setTimeout(() => {
				const mapEl = document.getElementById('leaflet-map');
				if (mapEl && typeof L !== 'undefined') {
					map = L.map('leaflet-map').setView([mapCoords.lat, mapCoords.lng], 16);
					L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
						attribution: '&copy; OpenStreetMap contributors'
					}).addTo(map);

					L.marker([mapCoords.lat, mapCoords.lng]).addTo(map).bindPopup(mapCoords.name).openPopup();

					L.circle([mapCoords.lat, mapCoords.lng], {
						color: '#3b82f6',
						fillColor: '#3b82f6',
						fillOpacity: 0.1,
						radius: mapCoords.radius
					}).addTo(map);
				}
			}, 100);

			return () => {
				clearTimeout(timer);
				if (map) {
					map.remove();
					map = null;
				}
			};
		}
	});

	let formData = $state({
		name: '',
		address: '',
		latitude: 0,
		longitude: 0,
		tolerance_radius: 50,
		manager_id: null,
		timezone: 'UTC'
	});

	let isDetectingTz = $state(false);
	$effect(() => {
		if (showModal && formData.latitude !== 0 && formData.longitude !== 0) {
			const detect = async () => {
				isDetectingTz = true;
				try {
					const res = await apiFetch(`/admin/utils/detect-timezone?lat=${formData.latitude}&lng=${formData.longitude}`);
					if (res.ok) {
						const data = await res.json();
						formData.timezone = data.timezone;
					}
				} finally {
					isDetectingTz = false;
				}
			};
			detect();
		}
	});

	let mapsUrl = $state('');
	let isParsing = $state(false);

	const canEdit = $derived(authState.isAdmin);

	async function loadCenters() {
		loading = true;
		const endpoint = authState.isManager ? '/manager/centers' : '/admin/centers';
		const res = await apiFetch<WorkCenter[]>(endpoint);
		if (res.ok) centers = await res.json();
		loading = false;
	}

	function openCreate() {
		if (!canEdit) return;
		editingItem = null;
		formData = {
			name: '',
			address: '',
			latitude: 0,
			longitude: 0,
			tolerance_radius: 50,
			manager_id: null,
			timezone: 'UTC'
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
		const url = editingItem ? `/admin/centers/${editingItem.id}` : '/admin/centers';

		const res = await apiFetch(url, {
			method,
			body: JSON.stringify(formData)
		});

		if (res.ok) {
			showModal = false;
			loadCenters();
		} else {
			const err = await res.json();
			alert(err.error || 'Error saving center');
		}
	}

	async function remove(item: any) {
		if (!canEdit) return;
		if (!confirm(`${$_('common.confirm_delete')} ${item.name}?`)) return;
		const res = await apiFetch(`/admin/centers/${item.id}`, { method: 'DELETE' });
		if (res.ok) loadCenters();
	}

	function openLocation(lat: number, lng: number, name: string, radius: number) {
		mapCoords = { lat, lng, name, radius };
		showMapModal = true;
	}

	async function importFromMaps() {
		if (!mapsUrl) return;

		// 1. Detectar si el usuario pegó coordenadas directas (e.g. "44.4668, -73.1630")
		const coordRegex = /^([-+]?\d{1,2}(?:\.\d+)?),\s*([-+]?\d{1,3}(?:\.\d+)?)$/;
		const match = mapsUrl.trim().match(coordRegex);

		if (match) {
			formData.latitude = parseFloat(match[1]);
			formData.longitude = parseFloat(match[2]);
			mapsUrl = ''; // Limpiar campo tras éxito
			return;
		}

		// 2. Si no son coordenadas, proceder con el parsing de URL de Google Maps
		isParsing = true;
		try {
			const res = await apiFetch('/admin/utils/parse-maps-url', {
				method: 'POST',
				body: JSON.stringify({ url: mapsUrl })
			});
			if (res.ok) {
				const data = await res.json();
				formData.latitude = parseFloat(data.latitude);
				formData.longitude = parseFloat(data.longitude);
				mapsUrl = ''; // Limpiar campo tras éxito
			} else {
				const err = await res.json();
				alert(err.error || 'No se pudieron extraer coordenadas del enlace.');
			}
		} catch (e) {
			alert('Error de conexión con el servicio de mapas.');
		} finally {
			isParsing = false;
		}
	}

	let searchQuery = $state('');

	const filteredCenters = $derived(
		centers.filter((c) => c.name.toLowerCase().includes(searchQuery.toLowerCase()))
	);

	onMount(() => {
		loadCenters();
		mounted = true;
	});
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

<main class="pb-24 px-6 max-w-5xl mx-auto">
	{#if mounted}
	<!-- Hero Section / Editorial Scale -->
	<section
		class="mt-8 mb-10 flex justify-between items-end"
		in:fly={{ y: 20, duration: 800, easing: quintOut }}
	>
		<div>
			<h2 class="text-4xl font-black text-primary leading-none tracking-tighter mb-4">
				{$_('admin.centers.title')}
			</h2>
		</div>
		{#if canEdit}
			<Button
				onclick={openCreate}
				class="h-14 px-8 rounded-sm bg-primary hover:bg-primary/90 text-white font-black text-xs uppercase tracking-[0.2em] shadow-xl shadow-primary/20 transition-all active:scale-95 flex items-center gap-2"
			>
				<Plus size={18} strokeWidth={3} />
				{$_('admin.centers.register_button')}
			</Button>
		{/if}
	</section>

	<!-- Search Bar (Integrated before list) -->
	<div
		class="mb-10 relative group"
		in:fly={{ y: 20, duration: 800, easing: quintOut, delay: 100 }}
	>
		<div
			class="flex items-center bg-white border-b-2 border-slate-200 focus-within:border-primary transition-all p-4 rounded-t-2xl shadow-sm"
		>
			<Search class="text-slate-300 mr-3" size={20} />
			<input
				type="text"
				placeholder={$_('admin.centers.search_placeholder')}
				bind:value={searchQuery}
				class="bg-transparent border-none focus:ring-0 w-full text-sm font-bold text-slate-900 placeholder:text-slate-300 uppercase tracking-tight"
			/>
		</div>
	</div>

	<!-- Work Centers List -->
	<div class="space-y-6" in:fly={{ y: 20, duration: 800, easing: quintOut, delay: 200 }}>
		{#if loading}
			{#each Array(3) as _}
				<div class="h-48 bg-white rounded-sm animate-pulse"></div>
			{/each}
		{:else if filteredCenters.length === 0}
			<div class="py-20 text-center bg-white rounded-sm border border-slate-100 border-dashed">
				<p class="text-slate-400 font-black uppercase tracking-[0.2em] text-[10px]">
					{$_('admin.centers.no_results')}
				</p>
			</div>
		{:else}
			{#each filteredCenters as center (center.id)}
				<div
					class="bg-white rounded-sm p-6 flex flex-col gap-6 border border-slate-50 shadow-xl shadow-slate-200/40 hover:scale-[1.01] transition-all duration-300 group"
				>
					<div class="flex justify-between items-start">
						<div>
							<h3 class="text-2xl font-black text-primary tracking-tighter leading-tight">
								<a href="/admin/centers/{center.id}">{center.name}</a>
							</h3>
							<div class="flex flex-col gap-1 mt-1 text-slate-400">
								<div class="flex items-center gap-1.5">
									<MapPin size={14} />
									<span
										class="text-[10px] font-black uppercase tracking-widest leading-none pt-0.5"
									>
										{center.address || $_('admin.centers.no_address')}
									</span>
								</div>
								<div class="flex items-center gap-1.5 opacity-60">
									<Navigation size={12} />
									<span class="text-[9px] font-bold tracking-tight leading-none">
										{center.latitude.toFixed(4)}, {center.longitude.toFixed(4)}
									</span>
								</div>
							</div>
						</div>

						<DropdownMenu.Root>
							<DropdownMenu.Trigger
								class="p-2 text-slate-300 hover:text-primary rounded-full hover:bg-slate-50 transition-all focus:outline-none"
							>
								<MoreHorizontal size={20} />
							</DropdownMenu.Trigger>
							<DropdownMenu.Content class="bg-white border-none  rounded-sm p-2 min-w-[200px]">
								<DropdownMenu.Item
									class="flex items-center gap-3 px-4 py-3 text-[10px] font-black uppercase tracking-widest text-slate-500 hover:text-primary hover:bg-slate-50 rounded-sm cursor-pointer"
									onSelect={() =>
										openLocation(
											center.latitude,
											center.longitude,
											center.name,
											center.tolerance_radius
										)}
								>
									<Navigation size={14} />
									{$_('admin.centers.view_map')}
								</DropdownMenu.Item>
								{#if canEdit}
									<DropdownMenu.Separator class="bg-slate-50 my-1" />
									<DropdownMenu.Item
										class="flex items-center gap-3 px-4 py-3 text-[10px] font-black uppercase tracking-widest text-slate-500 hover:text-primary hover:bg-slate-50 rounded-sm cursor-pointer"
										onSelect={() => openEdit(center)}
									>
										<Pencil size={14} />
										{$_('admin.centers.sync_changes')}
									</DropdownMenu.Item>
									<DropdownMenu.Item
										class="flex items-center gap-3 px-4 py-3 text-[10px] font-black uppercase tracking-widest text-rose-500 hover:bg-rose-50 rounded-sm cursor-pointer"
										onSelect={() => remove(center)}
									>
										<Trash2 size={14} />
										{$_('admin.centers.delete_node')}
									</DropdownMenu.Item>
								{/if}
							</DropdownMenu.Content>
						</DropdownMenu.Root>
					</div>

					<!-- Additional Professional Details -->
					<div class="grid grid-cols-2 gap-6 border-t border-b border-slate-50 py-4">
						<div class="flex flex-col">
							<span class="text-[9px] font-black uppercase tracking-[0.2em] text-slate-300 mb-1"
								>{$_('admin.centers.status_report')}</span
							>
							<div class="flex items-center gap-2">
								<div class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></div>
								<span class="text-xs font-bold text-slate-600">{$_('common.online')}</span>
							</div>
						</div>
						<div class="flex flex-col">
							<span class="text-[9px] font-black uppercase tracking-[0.2em] text-slate-300 mb-1"
								>{$_('admin.centers.geofencing')}</span
							>
							<div class="flex items-center gap-1.5 mt-0.5">
								<Activity size={12} class="text-primary" />
								<span class="text-xs font-black text-primary tracking-tight"
									>{$_('admin.centers.radius')}: {center.tolerance_radius}m</span
								>
							</div>
						</div>
					</div>
				</div>
			{/each}
		{/if}
	</div>

	<!-- Utility Stats -->
	<section
		class="mt-16 mb-8 bg-primary rounded-sm p-8 relative overflow-hidden shadow-2xl shadow-primary/20"
	>
		<div class="relative z-10 flex justify-between items-end">
			<div class="space-y-4">
				<h4 class="text-white/50 font-black text-[10px] uppercase tracking-[0.3em]">
					{$_('admin.centers.network_status')}
				</h4>
				<div class="flex gap-12">
					<div class="flex flex-col">
						<span class="text-4xl font-black text-white tracking-tighter">{centers.length}</span>
						<span class="text-[9px] font-black text-white/50 uppercase tracking-widest"
							>{$_('admin.centers.active_centers')}</span
						>
					</div>
					<div class="flex flex-col">
						<span class="text-4xl font-black text-white tracking-tighter">Operational</span>
						<span class="text-[9px] font-black text-white/50 uppercase tracking-widest"
							>{$_('admin.centers.nodes_mode')}</span
						>
					</div>
				</div>
			</div>
			<div class="bg-primary/10 p-3 rounded-sm">
				<Activity class="text-white h-6 w-6" />
			</div>
		</div>
	</section>
	{/if}
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
			class="rounded-md border-none p-12 sm:max-w-full md:max-w-4xl max-h-[90vh] overflow-y-auto"
		>
			<Dialog.Header class="space-y-6">
				<div class="space-y-2 text-left">
					<Dialog.Title class="text-4xl font-black tracking-tighter text-slate-900">
						{editingItem ? $_('admin.centers.edit_title') : $_('admin.centers.create_title')} <span class="text-primary italic">Sede</span>
					</Dialog.Title>
					<Dialog.Description
						class="text-sm font-bold text-slate-400 uppercase tracking-widest leading-relaxed"
					>
						{$_('admin.centers.description')}
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
					<div class="p-4 rounded-md bg-primary/5 border border-primary/10 space-y-3">
						<Label
							class="text-[10px] font-black text-primary uppercase tracking-widest flex items-center gap-2"
						>
							<Zap class="h-3 w-3" />
							{$_('admin.centers.magic_import')}
						</Label>
						<div class="flex gap-2">
							<Input
								bind:value={mapsUrl}
								placeholder={$_('admin.centers.maps_placeholder')}
								class="h-11 bg-white border-slate-200 rounded-md font-bold px-4 text-xs font-sans"
							/>
							<Button
								type="button"
								variant="secondary"
								class="h-11 px-4 rounded-md font-black text-xs gap-2 font-sans"
								onclick={importFromMaps}
								disabled={isParsing || !mapsUrl}
							>
								{#if isParsing}
									<Loader2 class="h-3 w-3 animate-spin" />
								{:else}
									<Link class="h-3 w-3" />
								{/if}
								{$_('common.import')}
							</Button>
						</div>
						<p class="text-[9px] font-bold text-slate-400 italic px-1">
							{$_('admin.centers.maps_hint')}
						</p>
					</div>

					<div class="space-y-2">
						<Label class="text-sm font-black text-slate-900 ml-1">{$_('admin.centers.name_label')}</Label>
						<Input
							bind:value={formData.name}
							placeholder="Ej: Planta Industrial Norte"
							class="h-14 bg-slate-50 border-none rounded-md font-bold px-5 focus-visible:ring-2 focus-visible:ring-primary/20 transition-all"
							required
						/>
					</div>

					<div class="space-y-2">
						<Label class="text-sm font-black text-slate-900 ml-1">{$_('admin.centers.address_label')}</Label>
						<Input
							bind:value={formData.address}
							placeholder="Ej: Av. Industrial 123, Sector Norte"
							class="h-14 bg-slate-50 border-none rounded-md font-bold px-5 focus-visible:ring-2 focus-visible:ring-primary/20 transition-all"
						/>
					</div>

					<div class="grid grid-cols-2 gap-4">
						<div class="space-y-2">
							<Label
								class="text-sm font-black text-slate-900 ml-1 uppercase tracking-tighter text-[10px]"
								>{$_('admin.centers.lat_label')}</Label
							>
							<Input
								type="number"
								step="any"
								bind:value={formData.latitude}
								class="h-14 bg-slate-50 border-none rounded-md font-mono font-bold px-5 font-sans"
								required
							/>
						</div>
						<div class="space-y-2">
							<Label
								class="text-sm font-black text-slate-900 ml-1 uppercase tracking-tighter text-[10px]"
								>{$_('admin.centers.lng_label')}</Label
							>
							<Input
								type="number"
								step="any"
								bind:value={formData.longitude}
								class="h-14 bg-slate-50 border-none rounded-md font-mono font-bold px-5 font-sans"
								required
							/>
						</div>
					</div>

					<div class="p-4 rounded-md bg-slate-50 border-2 border-dashed border-slate-200 flex justify-between items-center group transition-all hover:border-primary/30">
						<div class="flex flex-col">
							<span class="text-[10px] font-black text-slate-400 uppercase tracking-widest">Zona Horaria Detectada</span>
							<span class="text-sm font-bold text-slate-900 font-mono mt-0.5">
								{#if isDetectingTz}
									<span class="flex items-center gap-2 text-slate-300">
										<Loader2 class="h-3 w-3 animate-spin" />
										Analizando...
									</span>
								{:else}
									{formData.timezone}
								{/if}
							</span>
						</div>
						<div class="px-3 py-1.5 rounded-full bg-primary/10 text-primary text-[10px] font-black uppercase tracking-tighter">
							AUTO-DETECT
						</div>
					</div>

					<div class="space-y-2">
						<Label class="text-sm font-black text-slate-900 ml-1">{$_('admin.centers.radius_label')}</Label>
						<Input
							type="number"
							bind:value={formData.tolerance_radius}
							class="h-14 bg-slate-50 border-none rounded-md font-bold px-5"
							required
						/>
						<p class="text-[10px] font-bold text-slate-400 px-1 italic">
							{$_('admin.centers.radius_hint')}
						</p>
					</div>
				</div>

				<div class="flex flex-col gap-3 pt-4">
					<Button
						type="submit"
						class="h-16 rounded-md font-black text-lg gap-2 shadow-xl shadow-primary/20 hover:scale-[1.02] active:scale-95 transition-all"
						disabled={loading}
					>
						{#if loading}
							<Loader2 class="h-5 w-5 animate-spin" />
							{$_('common.processing')}
						{:else}
							{editingItem ? $_('admin.centers.sync_changes') : $_('admin.centers.register_button')}
						{/if}
					</Button>
					<Button
						type="button"
						variant="ghost"
						class="h-12 rounded-md font-black text-slate-400 hover:text-slate-600 hover:bg-slate-50"
						onclick={() => (showModal = false)}
					>
						{$_('common.cancel')}
					</Button>
				</div>
			</form>
		</Dialog.Content>
	</Dialog.Root>
{/if}

<!-- Map Modal -->
{#if showMapModal}
	<Dialog.Root
		open={showMapModal}
		onOpenChange={(o) => {
			if (!o) showMapModal = false;
		}}
	>
		<Dialog.Content
			class="rounded-md border-none  bg-white p-12 sm:max-w-full md:max-w-4xl max-h-[90vh]"
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
							{$_('admin.centers.map_monitor')}
						</Dialog.Title>
						<Dialog.Description class="text-xs font-black text-slate-400 uppercase tracking-widest">
							{mapCoords.name} • {mapCoords.lat.toFixed(6)}, {mapCoords.lng.toFixed(6)}
						</Dialog.Description>
					</div>
				</div>
			</Dialog.Header>

			<div
				class="relative w-full aspect-video rounded-md overflow-hidden bg-slate-100 border-2 border-slate-50 shadow-inner"
			>
				<div id="leaflet-map" class="h-full w-full"></div>
			</div>

			<div class="flex justify-end pt-8">
				<Button
					variant="secondary"
					class="rounded-md font-black gap-2 h-12 px-8 bg-slate-900 text-white hover:bg-slate-800"
					onclick={() => (showMapModal = false)}
				>
					{$_('admin.centers.finish_audit')}
				</Button>
			</div>
		</Dialog.Content>
	</Dialog.Root>
{/if}
