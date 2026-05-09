<script lang="ts">
	import StatusBadge from '$lib/components/StatusBadge.svelte';
	import { onMount } from 'svelte';
	import { fly } from 'svelte/transition';
	import { quintOut } from 'svelte/easing';
	import { goto } from '$app/navigation';
	import { Button } from "$lib/components/ui/button/index.js";
	import * as Card from "$lib/components/ui/card/index.js";
	import * as Table from "$lib/components/ui/table/index.js";
	import Badge from "$lib/components/ui/badge/badge.svelte";
	import { Input } from "$lib/components/ui/input/index.js";
	import { Label } from "$lib/components/ui/label/index.js";
	import * as Select from "$lib/components/ui/select/index.js";
	import * as Dialog from "$lib/components/ui/dialog/index.js";
	import { 
		Users, 
		Building2, 
		ArrowLeft, 
		Users2, 
		ShieldCheck, 
		Search, 
		ArrowRightLeft,
		Loader2,
		Briefcase
	} from "lucide-svelte";

	interface ManagedEmployee {
		id: number;
		user_name: string;
		user_email: string;
		center_name: string;
		shift_name: string | null;
		work_center_id: number;
		work_shift_id: number | null;
		is_active: boolean;
	}

	interface Center {
		id: number;
		name: string;
	}

	interface Shift {
		id: number;
		name: string;
	}

	let employees = $state<ManagedEmployee[]>([]);
	let centers = $state<Center[]>([]);
	let shifts = $state<Shift[]>([]);
	let loading = $state(true);
	let error = $state('');
	let selectedEmployee = $state<ManagedEmployee | null>(null);

	// Reassignment state
	let targetCenterId = $state(0);
	let targetShiftId = $state(0);
	let saving = $state(false);
	let searchQuery = $state("");

	const filteredEmployees = $derived(
		employees.filter(emp => 
			emp.user_name.toLowerCase().includes(searchQuery.toLowerCase()) ||
			emp.user_email.toLowerCase().includes(searchQuery.toLowerCase()) ||
			emp.center_name.toLowerCase().includes(searchQuery.toLowerCase())
		)
	);

	const selectedCenterLabel = $derived(
		centers.find(c => c.id === targetCenterId)?.name || "Select Center"
	);

	const selectedShiftLabel = $derived(
		targetShiftId === 0 ? "Unassigned" : shifts.find(s => s.id === targetShiftId)?.name || "Select Shift"
	);

	onMount(async () => {
		const token = localStorage.getItem('auth_token');
		if (!token) {
			goto('/');
			return;
		}
		
		await Promise.all([
			fetchEmployees(token),
			fetchCenters(token),
			fetchShifts(token)
		]);
		loading = false;
	});

	async function fetchEmployees(token: string) {
		const res = await fetch('http://localhost:3000/api/v1/manager/employees', {
			headers: { 'Authorization': `Bearer ${token}` }
		});
		if (res.ok) employees = await res.json();
	}

	async function fetchCenters(token: string) {
		const res = await fetch('http://localhost:3000/api/v1/manager/centers', {
			headers: { 'Authorization': `Bearer ${token}` }
		});
		if (res.ok) centers = await res.json();
	}

	async function fetchShifts(token: string) {
		const res = await fetch('http://localhost:3000/api/v1/admin/shifts', {
			headers: { 'Authorization': `Bearer ${token}` }
		});
		if (res.ok) shifts = await res.json();
	}

	function openReassignModal(emp: ManagedEmployee) {
		selectedEmployee = emp;
		targetCenterId = emp.work_center_id;
		targetShiftId = emp.work_shift_id || 0;
	}

	async function handleReassign() {
		if (!selectedEmployee) return;
		saving = true;
		const token = localStorage.getItem('auth_token');

		try {
			const res = await fetch(`http://localhost:3000/api/v1/manager/assign/${selectedEmployee.id}`, {
				method: 'POST',
				headers: { 
					'Authorization': `Bearer ${token}`,
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({
					work_center_id: targetCenterId,
					work_shift_id: targetShiftId || null
				})
			});

			if (res.ok) {
				await fetchEmployees(token!);
				selectedEmployee = null;
			} else {
				const data = await res.json();
				error = data.error || 'Failed to reassign employee';
			}
		} finally {
			saving = false;
		}
	}
</script>

<div class="min-h-screen bg-slate-50 p-6 md:p-12 max-w-7xl mx-auto space-y-12">
	<!-- Header -->
	<div in:fly={{ y: 20, duration: 800, easing: quintOut }}>
		<header
			class="flex flex-col md:flex-row justify-between items-start md:items-center gap-8"
		>
		<div class="space-y-3">
			<Button variant="ghost" href="/dashboard" class="rounded-md font-bold -ml-3 text-slate-500 hover:text-primary gap-2 font-sans">
				<ArrowLeft class="h-4 w-4" />
				Dashboard
			</Button>
			<h1 class="text-5xl font-black tracking-tighter text-slate-900 leading-tight">
				Portal de <span class="text-primary italic">Gestión</span>
			</h1>
			<p class="text-lg font-bold text-muted-foreground max-w-xl">
				Administra y reasigna personal a través de tus sedes operativas con validación inmediata.
			</p>
		</div>

		<div class="flex flex-wrap gap-4">
			<div class="flex -space-x-3">
				{#each employees.slice(0, 5) as emp}
					<div class="h-10 w-10 rounded-full border-2 border-white bg-slate-200 flex items-center justify-center text-[10px] font-black text-slate-600 shadow-sm">
						{emp.user_name[0]}
					</div>
				{/each}
				{#if employees.length > 5}
					<div class="h-10 w-10 rounded-full border-2 border-white bg-primary text-white flex items-center justify-center text-[10px] font-black shadow-sm">
						+{employees.length - 5}
					</div>
				{/if}
			</div>
			<Badge variant="outline" class="bg-primary/5 text-primary border-primary/20 font-black h-10 px-4 rounded-md italic">
				{employees.length} Operativos Activos
			</Badge>
		</div>
		</header>
	</div>

	{#if loading}
		<div class="flex flex-col items-center justify-center p-20 gap-4">
			<Loader2 class="h-12 w-12 text-primary animate-spin" />
			<p class="text-sm font-black text-muted-foreground uppercase tracking-widest">Sincronizando Nómina...</p>
		</div>
	{:else}
		<!-- Stats Row -->
		<div
			class="grid grid-cols-1 md:grid-cols-3 gap-6"
			in:fly={{ y: 20, duration: 800, easing: quintOut, delay: 100 }}
		>
			<Card.Root class="border-none shadow-premium rounded-md overflow-hidden bg-white group hover:scale-[1.02] transition-transform duration-300 text-card-foreground">
				<Card.Content class="p-8 space-y-4">
					<div class="h-12 w-12 rounded-md bg-indigo-50 text-indigo-600 flex items-center justify-center group-hover:bg-indigo-600 group-hover:text-white transition-colors duration-300">
						<Users2 class="h-6 w-6" />
					</div>
					<div>
						<p class="text-sm font-black text-muted-foreground uppercase tracking-wider">Total Personal</p>
						<p class="text-4xl font-black text-slate-900 leading-none mt-1">{employees.length}</p>
					</div>
				</Card.Content>
			</Card.Root>

			<Card.Root class="border-none shadow-premium rounded-md overflow-hidden bg-white group hover:scale-[1.02] transition-transform duration-300 text-card-foreground">
				<Card.Content class="p-8 space-y-4">
					<div class="h-12 w-12 rounded-md bg-emerald-50 text-emerald-600 flex items-center justify-center group-hover:bg-emerald-600 group-hover:text-white transition-colors duration-300">
						<Building2 class="h-6 w-6" />
					</div>
					<div>
						<p class="text-sm font-black text-muted-foreground uppercase tracking-wider">Sedes Gestionadas</p>
						<p class="text-4xl font-black text-slate-900 leading-none mt-1">{centers.length}</p>
					</div>
				</Card.Content>
			</Card.Root>

			<Card.Root class="border-none shadow-premium rounded-md overflow-hidden bg-white group hover:scale-[1.02] transition-transform duration-300 text-card-foreground">
				<Card.Content class="p-8 space-y-4">
					<div class="h-12 w-12 rounded-md bg-amber-50 text-amber-600 flex items-center justify-center group-hover:bg-amber-600 group-hover:text-white transition-colors duration-300">
						<ShieldCheck class="h-6 w-6" />
					</div>
					<div>
						<p class="text-sm font-black text-muted-foreground uppercase tracking-wider">Cumplimiento Activo</p>
						<p class="text-4xl font-black text-slate-900 leading-none mt-1">{employees.filter(e => e.is_active).length}</p>
					</div>
				</Card.Content>
			</Card.Root>
		</div>

		<!-- Workforce Section -->
		<div in:fly={{ y: 20, duration: 800, easing: quintOut, delay: 200 }}>
			<Card.Root
				class="border-none shadow-premium rounded-md overflow-hidden bg-white text-card-foreground"
			>
			<Card.Header class="p-8 md:p-10 border-b border-slate-100/50 space-y-6">
				<div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
					<div class="space-y-1">
						<Card.Title class="text-3xl font-black tracking-tight text-slate-900">Nómina Bajo Supervisión</Card.Title>
						<Card.Description class="text-sm font-bold text-slate-500 uppercase tracking-widest">Control y monitoreo de desplazamientos</Card.Description>
					</div>
					<div class="relative w-full md:w-80 group">
						<Search class="absolute left-4 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400 group-focus-within:text-primary transition-colors" />
						<Input 
							placeholder="Buscar por nombre, email o sede..." 
							bind:value={searchQuery}
							class="pl-11 h-12 bg-slate-50 border-none rounded-md font-bold focus-visible:ring-2 focus-visible:ring-primary/20 transition-all shadow-inner font-sans"
						/>
					</div>
				</div>
			</Card.Header>

			<Card.Content class="p-0">
				{#if error}
					<div class="m-8 p-4 rounded-md bg-rose-50 border border-rose-100 text-rose-600 text-sm font-bold flex items-center gap-3">
						<AlertCircle class="h-5 w-5" />
						{error}
					</div>
				{/if}

				<Table.Root>
					<Table.Header class="bg-slate-50/50">
						<Table.Row class="hover:bg-transparent border-slate-100">
							<Table.Head class="h-16 px-8 font-black text-slate-500 uppercase tracking-widest text-[10px]">Empleado</Table.Head>
							<Table.Head class="h-16 px-8 font-black text-slate-500 uppercase tracking-widest text-[10px]">Asignación Actual</Table.Head>
							<Table.Head class="h-16 px-8 font-black text-slate-500 uppercase tracking-widest text-[10px]">Turno</Table.Head>
							<Table.Head class="h-16 px-8 font-black text-slate-500 uppercase tracking-widest text-[10px]">Estado</Table.Head>
							<Table.Head class="h-16 px-8 text-right font-black text-slate-500 uppercase tracking-widest text-[10px]">Operaciones</Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each filteredEmployees as emp}
							<Table.Row class="hover:bg-slate-50/50 transition-colors border-slate-100 group">
								<Table.Cell class="px-8 py-6">
									<div class="flex items-center gap-4">
										<div class="h-12 w-12 rounded-md bg-primary/10 text-primary flex items-center justify-center font-black text-lg italic">
											{emp.user_name[0]}
										</div>
										<div class="flex flex-col">
											<span class="font-black text-slate-900 uppercase tracking-tight leading-tight">{emp.user_name}</span>
											<span class="text-xs font-bold text-slate-400">{emp.user_email}</span>
										</div>
									</div>
								</Table.Cell>
								<Table.Cell class="px-8 py-6">
									<Badge class="bg-slate-100 text-slate-600 hover:bg-slate-200 border-none rounded-lg font-bold px-3 py-1">
										<Building2 class="h-3 w-3 mr-2 opacity-50" />
										{emp.center_name}
									</Badge>
								</Table.Cell>
								<Table.Cell class="px-8 py-6">
									<Badge variant="outline" class="border-primary/20 text-primary font-bold px-3 py-1 bg-primary/5 rounded-lg">
										<Clock class="h-3 w-3 mr-2 opacity-50" />
										{emp.shift_name || 'Sin Turno'}
									</Badge>
								</Table.Cell>
								<Table.Cell class="px-8 py-6">
									<StatusBadge status={emp.is_active ? 'Activo' : 'Inactivo'} type={emp.is_active ? 'success' : 'error'} />
								</Table.Cell>
								<Table.Cell class="px-8 py-6 text-right">
									<Button 
										variant="ghost" 
										class="h-10 rounded-md font-black text-primary hover:bg-primary/10 transition-all opacity-0 group-hover:opacity-100 scale-95 group-hover:scale-100 gap-2 font-sans"
										onclick={() => openReassignModal(emp)}
									>
										<ArrowRightLeft class="h-4 w-4" />
										Reasignar
									</Button>
								</Table.Cell>
							</Table.Row>
						{:else}
							<Table.Row>
								<Table.Cell colspan={5} class="h-60 text-center">
									<div class="flex flex-col items-center justify-center gap-4 text-slate-300">
										<Users class="h-16 w-16 opacity-20" />
										<p class="font-black text-lg tracking-tight">No se encontraron colaboradores</p>
									</div>
								</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
			</Card.Content>
		</Card.Root>
	</div>
{/if}
</div>

<!-- Reassignment Dialog -->
{#if selectedEmployee}
	<Dialog.Root open={!!selectedEmployee} onOpenChange={(o) => { if(!o) selectedEmployee = null; }}>
		<Dialog.Content class="rounded-md border-none shadow-premium bg-white p-10 max-w-xl">
			<Dialog.Header class="space-y-4">
				<div class="h-16 w-16 bg-primary text-white rounded-md flex items-center justify-center shadow-xl shadow-primary/20 rotate-3">
					<ArrowRightLeft class="h-8 w-8" />
				</div>
				<div>
					<Dialog.Title class="text-4xl font-black tracking-tighter text-slate-900 leading-tight">
						Reasentamiento de <span class="text-primary italic">{selectedEmployee.user_name.split(' ')[0]}</span>
					</Dialog.Title>
					<Dialog.Description class="text-sm font-bold text-slate-500 uppercase tracking-widest mt-2 px-1">
						Sincronización Operativa Inmediata
					</Dialog.Description>
				</div>
			</Dialog.Header>

			<div class="py-8 space-y-6">
				<div class="space-y-3">
					<Label class="text-sm font-black text-slate-900 ml-1">Sede de Operaciones</Label>
					<Select.Root portal={null} onSelectedChange={(v: any) => targetCenterId = v.value}>
						<Select.Trigger class="h-14 bg-slate-50 border-none rounded-md font-bold px-5 focus:ring-2 focus:ring-primary/20 transition-all font-sans">
							<div class="flex items-center gap-3">
								<Building2 class="h-5 w-5 text-primary opacity-50" />
								<Select.Value placeholder={selectedCenterLabel} />
							</div>
						</Select.Trigger>
						<Select.Content class="rounded-md border-none shadow-premium p-2 animate-in fade-in zoom-in-95 duration-200">
							{#each centers as center}
								<Select.Item value={center.id} class="rounded-md font-bold px-4 py-3 hover:bg-slate-50 transition-colors cursor-pointer font-sans">
									{center.name}
								</Select.Item>
							{/each}
						</Select.Content>
					</Select.Root>
				</div>

				<div class="space-y-3">
					<Label class="text-sm font-black text-slate-900 ml-1">Configuración de Turno</Label>
					<Select.Root portal={null} onSelectedChange={(v: any) => targetShiftId = v.value}>
						<Select.Trigger class="h-14 bg-slate-50 border-none rounded-md font-bold px-5 focus:ring-2 focus:ring-primary/20 transition-all font-sans">
							<div class="flex items-center gap-3">
								<Briefcase class="h-5 w-5 text-primary opacity-50" />
								<Select.Value placeholder={selectedShiftLabel} />
							</div>
						</Select.Trigger>
						<Select.Content class="rounded-md border-none shadow-premium p-2 animate-in fade-in zoom-in-95 duration-200">
							<Select.Item value={0} class="rounded-md font-bold px-4 py-3 hover:bg-slate-50 transition-colors cursor-pointer text-slate-400 italic font-sans">
								Sin Asignación
							</Select.Item>
							{#each shifts as shift}
								<Select.Item value={shift.id} class="rounded-md font-bold px-4 py-3 hover:bg-slate-50 transition-colors cursor-pointer font-sans">
									{shift.name}
								</Select.Item>
							{/each}
						</Select.Content>
					</Select.Root>
				</div>
			</div>

			<Dialog.Footer class="flex flex-col sm:flex-row gap-3 pt-4 border-t border-slate-50">
				<Button 
					variant="ghost" 
					class="flex-1 h-14 rounded-md font-black text-slate-400 hover:text-slate-600 hover:bg-slate-50 font-sans" 
					onclick={() => selectedEmployee = null}
				>
					Cancelar
				</Button>
				<Button 
					class="flex-[2] h-14 rounded-md font-black gap-2 shadow-xl shadow-primary/20 text-lg hover:scale-[1.02] active:scale-95 transition-all font-sans text-white" 
					onclick={handleReassign} 
					disabled={saving}
				>
					{#if saving}
						<Loader2 class="h-5 w-5 animate-spin" />
						Actualizando...
					{:else}
						Confirmar Cambios
					{/if}
				</Button>
			</Dialog.Footer>
		</Dialog.Content>
	</Dialog.Root>
{/if}
