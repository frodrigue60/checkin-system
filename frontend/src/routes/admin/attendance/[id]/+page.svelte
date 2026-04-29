<script lang="ts">
	import { onMount } from 'svelte';
	import { fly, fade } from 'svelte/transition';
	import { quintOut } from 'svelte/easing';
	import { page } from '$app/state';
	import { apiFetch } from '$lib/api';
	import { Button } from '$lib/components/ui/button/index.js';
	import Badge from '$lib/components/ui/badge/badge.svelte';
	import { _ } from 'svelte-i18n';
	import {
		Loader2,
		Calendar,
		MapPin,
		Briefcase,
		Clock,
		Activity,
		DollarSign,
		AlertTriangle,
		ArrowLeft,
		ExternalLink,
		ShieldAlert,
		History,
		Navigation,
		User as UserIcon,
		CheckCircle2,
		XCircle,
		Timer,
		Edit3,
		Save,
		Trash2,
		Camera,
		Maximize2
	} from 'lucide-svelte';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';

	let attId = $derived(page.params.id);
	let data = $state<any>(null);
	let loading = $state(true);
	let errorMsg = $state('');

	async function loadDetails() {
		loading = true;
		try {
			const res = await apiFetch(`/admin/attendances/${attId}/details`);
			if (res.ok) {
				data = await res.json();
			} else {
				errorMsg = $_('admin.reports.load_details_error');
			}
		} catch (e) {
			errorMsg = $_('auth.error_server');
		} finally {
			loading = false;
		}
	}

	onMount(loadDetails);

	function formatTime(dateStr: string | null) {
		if (!dateStr) return '--:--';
		
		// Handle raw Year 0 timestamps from DB (e.g. 0000-01-01T08:00:00Z)
		if (dateStr.includes('0000-01-01') || dateStr.includes('0001-01-01')) {
			const timePart = dateStr.split('T')[1];
			if (timePart) {
				return timePart.substring(0, 5); // Return HH:MM
			}
		}

		// Handle standard HH:MM:SS or HH:MM
		if (dateStr.length <= 8 && dateStr.includes(':')) {
			return dateStr.substring(0, 5);
		}

		try {
			const d = new Date(dateStr);
			if (isNaN(d.getTime())) return dateStr;
			return d.toLocaleTimeString('es-ES', {
				hour: '2-digit',
				minute: '2-digit'
			});
		} catch (e) {
			return dateStr;
		}
	}

	function formatDate(dateStr: string | null) {
		if (!dateStr) return '---';
		return new Date(dateStr).toLocaleDateString('es-ES', {
			weekday: 'long',
			year: 'numeric',
			month: 'long',
			day: 'numeric'
		});
	}

	function formatCurrency(val: number) {
		return new Intl.NumberFormat('es-MX', {
			style: 'currency',
			currency: 'MXN'
		}).format(val);
	}

	function getGoogleMapsUrl(lat: number, lng: number) {
		return `https://www.google.com/maps?q=${lat},${lng}`;
	}

	async function resolveIncident(incidentId: number, status: string, note: string) {
		try {
			const res = await apiFetch(`/admin/incidents/${incidentId}`, {
				method: 'PATCH',
				body: JSON.stringify({
					status: status,
					resolution_note: note
				})
			});
			
			if (res.ok) {
				await loadDetails(); // Refresh to show updated earnings and statuses
			} else {
				const err = await res.json();
				alert('Error: ' + (err.error || 'No se pudo actualizar el incidente'));
			}
		} catch (e) {
			alert('Error de conexión');
		}
	}

	let editMode = $state(false);
	let editForm = $state({
		check_in: '',
		lunch_start: '',
		lunch_end: '',
		check_out: '',
		is_absence: false,
		absence_reason: ''
	});

	function enterEditMode() {
		// Helper to format for datetime-local input (YYYY-MM-DDTHH:mm)
		const getLocalISO = (dateStr: string | null) => {
			if (!dateStr) return '';
			try {
				return dateStr.split('.')[0].slice(0, 16); // Format: YYYY-MM-DDTHH:MM
			} catch (e) { return ''; }
		};

		editForm = {
			check_in: getLocalISO(data.attendance.check_in),
			lunch_start: getLocalISO(data.attendance.lunch_start),
			lunch_end: getLocalISO(data.attendance.lunch_end),
			check_out: getLocalISO(data.attendance.check_out),
			is_absence: data.attendance.is_absence,
			absence_reason: data.attendance.absence_reason || ''
		};

		// If times are empty but we have a shift, suggest current date with shift times
		if (!editForm.check_in && data.attendance.created_at) {
			const datePart = data.attendance.created_at.split('T')[0];
			if (data.shift) {
				editForm.check_in = `${datePart}T${data.shift.start_time.substring(0, 5)}`;
				editForm.check_out = `${datePart}T${data.shift.end_time.substring(0, 5)}`;
				// Also suggest a default lunch if needed
				editForm.lunch_start = `${datePart}T13:00`;
				editForm.lunch_end = `${datePart}T14:00`;
			}
		}
		
		editMode = true;
	}

	async function saveAttendance() {
		try {
			const payload = {
				check_in: editForm.check_in ? new Date(editForm.check_in).toISOString() : null,
				lunch_start: editForm.lunch_start ? new Date(editForm.lunch_start).toISOString() : null,
				lunch_end: editForm.lunch_end ? new Date(editForm.lunch_end).toISOString() : null,
				check_out: editForm.check_out ? new Date(editForm.check_out).toISOString() : null,
				is_absence: editForm.is_absence,
				absence_reason: editForm.absence_reason
			};

			const res = await apiFetch(`/admin/attendances/${attId}`, {
				method: 'PUT',
				body: JSON.stringify(payload)
			});
			
			if (res.ok) {
				editMode = false;
				await loadDetails();
			} else {
				const err = await res.json();
				alert('Error: ' + (err.error || 'No se pudo actualizar la asistencia'));
			}
		} catch (e) {
			alert('Error de conexión');
		}
	}

	async function recalculateIncidents() {
		try {
			const res = await apiFetch(`/admin/attendances/${attId}/recalculate`, {
				method: 'POST'
			});
			
			if (res.ok) {
				await loadDetails();
			} else {
				const err = await res.json();
				alert('Error: ' + (err.error || 'No se pudo recalcular'));
			}
		} catch (e) {
			alert('Error de conexión');
		}
	}
</script>

<svelte:head>
	<title>{$_('admin.attendance.detail.title')} | JGC</title>
</svelte:head>

<div class="container mx-auto p-4 max-w-5xl space-y-6">
	<!-- Top Navigation -->
	<div class="flex items-center justify-between mb-2">
		<Button variant="ghost" href="/admin/attendance" class="gap-2 text-muted-foreground hover:text-foreground">
			<ArrowLeft class="w-4 h-4" />
			{$_('admin.attendance.detail.back_to_history')}
		</Button>
		
		{#if data}
			<div class="flex gap-2">
				<Badge variant={data.attendance.is_absence ? "destructive" : "outline"} class="px-3 py-1">
					{data.attendance.is_absence ? $_('admin.attendance.absence') : $_('admin.attendance.detail.attendance')}
				</Badge>
				{#if data.attendance.is_late}
					<Badge variant="destructive" class="px-3 py-1 flex gap-1 items-center">
						<AlertTriangle class="w-3 h-3" />
						{$_('admin.incidents.types.late')}
					</Badge>
				{/if}
			</div>
		{/if}
	</div>

	{#if loading}
		<div class="flex flex-col items-center justify-center min-h-[400px] gap-4" in:fade>
			<Loader2 class="w-10 h-10 animate-spin text-primary" />
			<p class="text-muted-foreground animate-pulse">{$_('dashboard.loading')}</p>
		</div>
	{:else if errorMsg}
		<div class="flex flex-col items-center justify-center min-h-[400px] gap-4" in:fade>
			<div class="p-4 rounded-full bg-destructive/10 text-destructive">
				<ShieldAlert class="w-12 h-12" />
			</div>
			<p class="text-xl font-semibold">{errorMsg}</p>
			<Button onclick={loadDetails}>{$_('admin.shifts.recalculate')}</Button>
		</div>
	{:else if data}
		<!-- Header Info -->
		<div class="grid grid-cols-1 md:grid-cols-3 gap-6" in:fly={{ y: 20, duration: 800, easing: quintOut }}>
			<div class="md:col-span-2 space-y-2">
				<h1 class="text-3xl font-bold tracking-tight">{data.attendance.employee_name}</h1>
				<div class="flex flex-wrap gap-4 text-muted-foreground">
					<span class="flex items-center gap-1.5">
						<Calendar class="w-4 h-4" />
						{formatDate(data.attendance.check_in || data.attendance.date)}
					</span>
					<span class="flex items-center gap-1.5">
						<Briefcase class="w-4 h-4" />
						{data.attendance.position_name}
					</span>
				</div>
			</div>
			
			<div class="bg-card border rounded-xl p-6 shadow-sm flex flex-col justify-center items-end space-y-1">
				<p class="text-sm font-medium text-muted-foreground">{$_('admin.attendance.detail.daily_gain')}</p>
				<p class="text-3xl font-bold text-primary">{formatCurrency(data.attendance.daily_earnings)}</p>
			</div>
		</div>

		<div class="grid grid-cols-1 md:grid-cols-12 gap-6">
			<!-- Main Stats -->
			<div class="md:col-span-8 space-y-6">
				<!-- Temporal Analysis -->
				<section class="bg-card border rounded-xl overflow-hidden shadow-sm" in:fly={{ y: 20, duration: 800, easing: quintOut, delay: 100 }}>
					<div class="px-6 py-4 border-b bg-muted/30 flex items-center justify-between">
						<div class="flex items-center gap-2">
							<Timer class="w-5 h-5 text-primary" />
							<h2 class="font-semibold text-lg">{$_('admin.attendance.detail.temporal_analysis')}</h2>
						</div>
						<div class="flex items-center gap-2">
							<Button 
								size="sm" 
								variant="outline" 
								class="gap-1.5 h-8 text-orange-600 border-orange-200 hover:bg-orange-50" 
								onclick={recalculateIncidents}
								title={$_('admin.attendance.detail.recalculate_hint')}
							>
								<Activity class="w-3.5 h-3.5" />
								{$_('admin.attendance.detail.recalculate')}
							</Button>
							<Button size="sm" variant="outline" class="gap-1.5 h-8" onclick={editMode ? saveAttendance : enterEditMode}>
								{#if editMode}
									<Save class="w-3.5 h-3.5" />
									{$_('admin.attendance.detail.save_changes')}
								{:else}
									<Edit3 class="w-3.5 h-3.5" />
									{$_('admin.attendance.detail.edit_times')}
								{/if}
							</Button>
						</div>
					</div>
					
					<div class="p-6">
						{#if editMode}
							<div class="grid grid-cols-1 md:grid-cols-2 gap-6" in:fade>
								<div class="md:col-span-2 p-4 rounded-lg bg-orange-50 dark:bg-orange-950/20 border border-orange-200 dark:border-orange-900 mb-2">
									<div class="flex items-center gap-3 mb-4">
										<input 
											type="checkbox" 
											id="is_absence" 
											class="w-5 h-5 rounded border-gray-300 text-primary focus:ring-primary" 
											bind:checked={editForm.is_absence}
										/>
										<Label for="is_absence" class="text-base font-semibold cursor-pointer">{$_('admin.attendance.detail.mark_absence')}</Label>
									</div>
									
									{#if editForm.is_absence}
										<div class="space-y-2 ml-8" in:fade>
											<Label for="absence_reason">{$_('admin.attendance.detail.absence_reason_label')}</Label>
											<textarea 
												id="absence_reason" 
												class="flex min-h-[80px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
												placeholder={$_('admin.attendance.detail.absence_reason_placeholder')}
												bind:value={editForm.absence_reason}
											></textarea>
										</div>
									{/if}
								</div>

								{#if !editForm.is_absence}
									<div class="space-y-2">
										<Label for="check_in">{$_('admin.attendance.detail.check_in_label')}</Label>
										<Input type="datetime-local" id="check_in" bind:value={editForm.check_in} />
									</div>
									<div class="space-y-2">
										<Label for="check_out">{$_('admin.attendance.detail.check_out_label')}</Label>
										<Input type="datetime-local" id="check_out" bind:value={editForm.check_out} />
									</div>
									<div class="space-y-2">
										<Label for="lunch_start">{$_('admin.attendance.detail.lunch_start_label')}</Label>
										<Input type="datetime-local" id="lunch_start" bind:value={editForm.lunch_start} />
									</div>
									<div class="space-y-2">
										<Label for="lunch_end">{$_('admin.attendance.detail.lunch_end_label')}</Label>
										<Input type="datetime-local" id="lunch_end" bind:value={editForm.lunch_end} />
									</div>
								{/if}

								<div class="md:col-span-2 flex justify-end gap-2 pt-2">
									<Button variant="ghost" onclick={() => editMode = false}>{$_('common.cancel')}</Button>
									<Button onclick={saveAttendance}>
										<Save class="w-4 h-4 mr-2" />
										{$_('admin.attendance.detail.save_changes')}
									</Button>
								</div>
							</div>
						{:else if data.attendance.is_absence}
							<div class="flex flex-col items-center justify-center py-10 text-center space-y-4" in:fade>
								<div class="p-4 rounded-full bg-orange-100 dark:bg-orange-900/30 text-orange-600">
									<Activity class="w-12 h-12" />
								</div>
								<div class="max-w-md">
									<h3 class="text-xl font-bold">{$_('admin.attendance.detail.absence_reported')}</h3>
									<p class="text-muted-foreground italic mt-2 text-lg">"{data.attendance.absence_reason || $_('admin.attendance.detail.no_reason_specified')}"</p>
									
									<div class="flex justify-center gap-3 mt-6">
										<Button variant="outline" class="gap-2" onclick={enterEditMode}>
											<Edit3 class="w-4 h-4" />
											{$_('admin.attendance.detail.correct_or_edit')}
										</Button>
										<Button variant="secondary" class="gap-2" onclick={recalculateIncidents}>
											<Activity class="w-4 h-4" />
											{$_('admin.attendance.detail.recalculate')}
										</Button>
									</div>
								</div>
							</div>
						{:else}
							<div class="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
								<div class="space-y-1">
									<p class="text-xs font-medium text-muted-foreground uppercase">{$_('admin.attendance.detail.real_entrance')}</p>
									<p class="text-2xl font-bold">{formatTime(data.attendance.check_in)}</p>
								</div>
								<div class="space-y-1">
									<p class="text-xs font-medium text-muted-foreground uppercase">{$_('admin.attendance.detail.real_exit')}</p>
									<p class="text-2xl font-bold">{formatTime(data.attendance.check_out)}</p>
								</div>
								<div class="space-y-1">
									<p class="text-xs font-medium text-muted-foreground uppercase">{$_('admin.attendance.detail.net_hours')}</p>
									<p class="text-2xl font-bold text-primary">{data.attendance.net_hours_worked.toFixed(2)}h</p>
								</div>
								<div class="space-y-1">
									<p class="text-xs font-medium text-muted-foreground uppercase">{$_('admin.attendance.detail.incidents')}</p>
									<p class="text-2xl font-bold {data.incidents.length > 0 ? 'text-destructive' : 'text-green-500'}">
										{data.incidents.length}
									</p>
								</div>
							</div>

							<!-- Timeline visualization -->
							<div class="relative pt-8 pb-4 px-2">
								<div class="absolute top-1/2 left-0 w-full h-0.5 bg-muted -translate-y-1/2"></div>
								<div class="flex justify-between items-center relative">
									<!-- Expected Start -->
									<div class="flex flex-col items-center gap-2 group">
										<div class="w-3 h-3 rounded-full bg-muted-foreground relative z-10"></div>
										<div class="text-center">
											<p class="text-[10px] font-bold text-muted-foreground uppercase">{$_('admin.attendance.detail.scheduled')}</p>
											<p class="text-sm font-medium">{formatTime(data.shift?.start_time)}</p>
										</div>
									</div>

									<!-- Actual Check-In -->
									<div class="flex flex-col items-center gap-2 relative">
										<div class="w-5 h-5 rounded-full border-4 border-background z-10 {data.attendance.is_late ? 'bg-destructive' : 'bg-green-500'}"></div>
										<div class="text-center">
											<p class="text-[10px] font-bold text-muted-foreground uppercase">Check-In</p>
											<p class="text-sm font-bold">{formatTime(data.attendance.check_in)}</p>
										</div>
									</div>

									<!-- Lunch Start -->
									{#if data.attendance.lunch_start}
										<div class="flex flex-col items-center gap-2 relative">
											<div class="w-3 h-3 rounded-full bg-blue-500 relative z-10"></div>
											<div class="text-center">
												<p class="text-[10px] font-bold text-muted-foreground uppercase">{$_('common.lunch_start')}</p>
												<p class="text-sm font-medium">{formatTime(data.attendance.lunch_start)}</p>
											</div>
										</div>
									{/if}

									<!-- Actual Check-Out -->
									<div class="flex flex-col items-center gap-2 relative">
										<div class="w-5 h-5 rounded-full border-4 border-background z-10 {data.attendance.check_out ? 'bg-primary' : 'bg-muted'}"></div>
										<div class="text-center">
											<p class="text-[10px] font-bold text-muted-foreground uppercase">Check-Out</p>
											<p class="text-sm font-bold">{formatTime(data.attendance.check_out)}</p>
										</div>
									</div>

									<!-- Expected End -->
									<div class="flex flex-col items-center gap-2 group">
										<div class="w-3 h-3 rounded-full bg-muted-foreground relative z-10"></div>
										<div class="text-center">
											<p class="text-[10px] font-bold text-muted-foreground uppercase">{$_('admin.attendance.detail.scheduled')}</p>
											<p class="text-sm font-medium">{formatTime(data.shift?.end_time)}</p>
										</div>
									</div>
								</div>
							</div>
						{/if}
					</div>
				</section>

				<!-- Incidents -->
				<section class="bg-card border rounded-xl overflow-hidden shadow-sm" in:fly={{ y: 20, duration: 800, easing: quintOut, delay: 200 }}>
					<div class="px-6 py-4 border-b bg-muted/30 flex items-center justify-between">
						<div class="flex items-center gap-2">
							<AlertTriangle class="w-5 h-5 text-destructive" />
							<h2 class="font-semibold text-lg">{$_('admin.attendance.detail.detected_incidents')}</h2>
						</div>
						<Badge variant="outline">{data.incidents.length}</Badge>
					</div>
					
					<div class="divide-y">
						{#each data.incidents as incident, i}
							<div class="p-6 flex items-start gap-4 hover:bg-muted/10 transition-colors">
								<div class="mt-1">
									{#if incident.type === 'late'}
										<div class="p-2 rounded-lg bg-orange-100 dark:bg-orange-950 text-orange-600">
											<Clock class="w-5 h-5" />
										</div>
									{:else}
										<div class="p-2 rounded-lg bg-destructive/10 text-destructive">
											<MapPin class="w-5 h-5" />
										</div>
									{/if}
								</div>
								
								<div class="flex-1 space-y-2">
									<div class="flex items-center justify-between">
										<div class="flex items-center gap-2">
											<h3 class="font-bold text-lg">
												{#if incident.type === 'late'}
													{$_('admin.incidents.types.late')}
												{:else if incident.type === 'out_of_range'}
													{$_('admin.incidents.types.out_of_range')}
												{:else if incident.type === 'lunch_overstay'}
													{$_('admin.attendance.detail.lunch_overstay')}
												{:else if incident.type === 'absent'}
													{$_('admin.attendance.absence')}
												{:else}
													{incident.type}
												{/if}
											</h3>
											<Badge 
												variant={incident.status === 'approved' ? 'destructive' : incident.status === 'justified' ? 'secondary' : 'outline'}
												class="text-[10px] uppercase px-1.5 py-0"
											>
												{incident.status === 'approved' ? $_('admin.attendance.detail.approved') : incident.status === 'justified' ? $_('admin.attendance.detail.justified') : $_('admin.attendance.detail.pending')}
											</Badge>
										</div>
										<span class="text-xs text-muted-foreground font-mono">
											{new Date(incident.created_at).toLocaleTimeString()}
										</span>
									</div>
									
									<p class="text-muted-foreground text-sm">
										{#if incident.type === 'late'}
											{$_('admin.attendance.detail.minutes_late_desc', { values: { n: incident.delay_minutes } })}
										{:else if incident.type === 'out_of_range'}
											{$_('admin.attendance.detail.meters_away_desc', { values: { n: incident.distance } })}
										{:else if incident.type === 'lunch_overstay'}
											{$_('admin.attendance.detail.lunch_overstay_desc', { values: { n: incident.delay_minutes } })}
										{:else if incident.type === 'absent'}
											{$_('admin.attendance.detail.absence_reported_desc', { values: { desc: incident.description || $_('admin.attendance.detail.no_reason_specified') } })}
										{:else}
											{incident.description || $_('admin.attendance.detail.no_anomalies_detected')}
										{/if}
									</p>
									
									{#if incident.resolution_note}
										<div class="p-2 bg-blue-50 dark:bg-blue-900/20 border border-blue-100 dark:border-blue-800 rounded-lg text-xs italic">
											<span class="font-bold not-italic">{$_('admin.attendance.detail.note')}:</span> {incident.resolution_note}
										</div>
									{/if}

									{#if incident.status === 'pending' || !incident.status}
										<div class="flex gap-2 pt-2">
											<Button 
												size="sm" 
												variant="outline" 
												class="h-8 text-xs border-green-500 text-green-600 hover:bg-green-50"
												onclick={() => {
													const note = prompt($_('admin.attendance.detail.resolution_note_prompt'));
													if (note !== null) resolveIncident(incident.id, 'justified', note);
												}}
											>
												<CheckCircle2 class="w-3.5 h-3.5 mr-1" />
												{$_('admin.attendance.detail.justify_valid')}
											</Button>
											<Button 
												size="sm" 
												variant="outline" 
												class="h-8 text-xs border-destructive text-destructive hover:bg-destructive/10"
												onclick={() => {
													const note = prompt($_('admin.attendance.detail.resolution_note_prompt'));
													if (note !== null) resolveIncident(incident.id, 'approved', note);
												}}
											>
												<XCircle class="w-3.5 h-3.5 mr-1" />
												{$_('admin.attendance.detail.apply_penalty')}
											</Button>
										</div>
									{/if}
									
									{#if incident.metadata_json && incident.status === 'pending'}
										<details class="mt-2">
											<summary class="text-[10px] text-muted-foreground cursor-pointer hover:underline">Ver metadatos RAW</summary>
											<div class="mt-1 p-2 bg-muted rounded text-[10px] font-mono text-muted-foreground overflow-x-auto whitespace-pre">
												{JSON.stringify(JSON.parse(incident.metadata_json), null, 2)}
											</div>
										</details>
									{/if}
								</div>
							</div>
						{:else}
							<div class="flex flex-col items-center justify-center py-12 text-muted-foreground">
								<CheckCircle2 class="w-12 h-12 text-green-500/20 mb-2" />
								<p>{$_('admin.attendance.detail.no_anomalies_detected')}</p>
							</div>
						{/each}
					</div>
				</section>
			</div>

			<!-- Sidebar Info -->
			<div class="md:col-span-4 space-y-6">
				<!-- Evidence Photos Gallery -->
				{#if (data.attendance.evidence_urls && data.attendance.evidence_urls.length > 0) || data.attendance.evidence_url}
					<section class="bg-card border rounded-xl overflow-hidden shadow-sm" in:fly={{ y: 20, duration: 800, easing: quintOut, delay: 250 }}>
						<div class="px-6 py-4 border-b bg-muted/30 flex items-center gap-2">
							<Camera class="w-5 h-5 text-primary" />
							<h2 class="font-semibold text-lg">{$_('admin.attendance.detail.evidence_title')}</h2>
							{#if data.attendance.evidence_urls}
								<span class="ml-auto px-2 py-0.5 rounded-full bg-primary/10 text-primary text-[10px] font-black uppercase">
									{data.attendance.evidence_urls.length} Fotos
								</span>
							{/if}
						</div>
						<div class="p-4 space-y-4">
							<div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
								{#if data.attendance.evidence_urls && data.attendance.evidence_urls.length > 0}
									{#each data.attendance.evidence_urls as url}
										<div class="relative group rounded-lg overflow-hidden bg-slate-100 aspect-video flex items-center justify-center border shadow-inner">
											<img 
												src={url} 
												alt="Evidencia" 
												class="object-cover w-full h-full transition-transform duration-500 group-hover:scale-110"
											/>
											<div class="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center">
												<Button variant="secondary" size="icon" class="rounded-full" target="_blank" href={url}>
													<Maximize2 class="w-4 h-4" />
												</Button>
											</div>
										</div>
									{/each}
								{:else}
									<div class="relative group rounded-lg overflow-hidden bg-slate-100 aspect-video flex items-center justify-center border shadow-inner col-span-2">
										<img 
											src={data.attendance.evidence_url} 
											alt="Evidencia" 
											class="object-cover w-full h-full transition-transform duration-500 group-hover:scale-110"
										/>
										<div class="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center">
											<Button variant="secondary" size="sm" class="gap-2" target="_blank" href={data.attendance.evidence_url}>
												<Maximize2 class="w-4 h-4" />
												{$_('admin.attendance.detail.full_screen')}
											</Button>
										</div>
									</div>
								{/if}
							</div>

							{#if data.attendance.check_out_note}
								<div class="p-3 bg-primary/5 border border-primary/10 rounded-lg">
									<p class="text-[10px] font-black text-primary uppercase tracking-widest mb-1">
										{$_('dashboard.service_note_label')}
									</p>
									<p class="text-xs font-bold text-slate-700 leading-relaxed italic">
										"{data.attendance.check_out_note}"
									</p>
								</div>
							{/if}
							<p class="text-[10px] text-muted-foreground text-center uppercase tracking-widest font-bold">
								{$_('admin.attendance.detail.evidence_hint')}
							</p>
						</div>
					</section>
				{/if}

				<!-- Geolocation -->
				<section class="bg-card border rounded-xl overflow-hidden shadow-sm" in:fly={{ y: 20, duration: 800, easing: quintOut, delay: 300 }}>
					<div class="px-6 py-4 border-b bg-muted/30 flex items-center gap-2">
						<Navigation class="w-5 h-5 text-primary" />
						<h2 class="font-semibold text-lg">{$_('admin.attendance.detail.location')}</h2>
					</div>
					<div class="p-6 space-y-4">
						<div class="space-y-1">
							<p class="text-xs font-bold text-muted-foreground uppercase">{$_('admin.attendance.detail.assigned_center')}</p>
							<p class="font-semibold">{data.work_center.name}</p>
							<p class="text-sm text-muted-foreground">{data.work_center.address || $_('admin.attendance.detail.no_address')}</p>
						</div>
						
						{#if data.attendance.check_out_address}
							<div class="p-3 bg-muted/50 rounded-lg border border-slate-200">
								<p class="text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-1">
									{$_('dashboard.checkout_address_label')}
								</p>
								<p class="text-xs font-bold text-slate-900">
									{data.attendance.check_out_address}
								</p>
							</div>
						{/if}
						
						<div class="pt-2 flex flex-col gap-2">
							{#if data.attendance.check_in_latitude}
								<Button variant="outline" class="w-full gap-2 justify-start" target="_blank" href={getGoogleMapsUrl(data.attendance.check_in_latitude, data.attendance.check_in_longitude)}>
									<MapPin class="w-4 h-4 text-green-500" />
									{$_('admin.attendance.detail.entry_map')}
									<ExternalLink class="w-3 h-3 ml-auto opacity-50" />
								</Button>
							{/if}
							{#if data.attendance.check_out_latitude}
								<Button variant="outline" class="w-full gap-2 justify-start" target="_blank" href={getGoogleMapsUrl(data.attendance.check_out_latitude, data.attendance.check_out_longitude)}>
									<MapPin class="w-4 h-4 text-primary" />
									{$_('admin.attendance.detail.exit_map')}
									<ExternalLink class="w-3 h-3 ml-auto opacity-50" />
								</Button>
							{/if}
						</div>
						
						<div class="pt-2 p-3 bg-muted/50 rounded-lg space-y-2">
							<div class="flex justify-between text-xs">
								<span class="text-muted-foreground">{$_('admin.attendance.detail.tolerance_radius')}:</span>
								<span class="font-bold">{data.work_center.tolerance_radius}m</span>
							</div>
							<div class="flex justify-between text-xs">
								<span class="text-muted-foreground">{$_('admin.attendance.detail.timezone')}:</span>
								<span class="font-bold">{data.work_center.timezone || 'UTC'}</span>
							</div>
						</div>
					</div>
				</section>

				<!-- Shift Config -->
				<section class="bg-card border rounded-xl overflow-hidden shadow-sm" in:fly={{ y: 20, duration: 800, easing: quintOut, delay: 400 }}>
					<div class="px-6 py-4 border-b bg-muted/30 flex items-center gap-2">
						<Clock class="w-5 h-5 text-primary" />
						<h2 class="font-semibold text-lg">{$_('admin.attendance.detail.shift_config')}</h2>
					</div>
					<div class="p-6">
						{#if data.shift}
							<div class="space-y-4">
								<div class="flex justify-between items-center">
									<h3 class="font-bold">{data.shift.name}</h3>
									<Badge variant={data.shift.is_active ? "outline" : "secondary"}>
										{data.shift.is_active ? $_('common.active') : $_('common.inactive')}
									</Badge>
								</div>
								
								<div class="grid grid-cols-2 gap-4">
									<div class="space-y-1">
										<p class="text-[10px] font-bold text-muted-foreground uppercase">{$_('common.check_in')}</p>
										<p class="font-semibold">{formatTime(data.shift.start_time)}</p>
									</div>
									<div class="space-y-1">
										<p class="text-[10px] font-bold text-muted-foreground uppercase">{$_('common.check_out')}</p>
										<p class="font-semibold">{formatTime(data.shift.end_time)}</p>
									</div>
									<div class="space-y-1">
										<p class="text-[10px] font-bold text-muted-foreground uppercase">{$_('admin.shifts.tolerance_short')}</p>
										<p class="font-semibold">{formatTime(data.shift.grace_period)}</p>
									</div>
									<div class="space-y-1">
										<p class="text-[10px] font-bold text-muted-foreground uppercase">{$_('admin.shifts.lunch_short')}</p>
										<p class="font-semibold">{formatTime(data.shift.lunch_duration_limit)}</p>
									</div>
								</div>
								
								{#if data.shift.is_night_shift}
									<Badge variant="secondary" class="w-full justify-center gap-1.5">
										<Clock class="w-3 h-3" />
										{$_('admin.attendance.detail.night_shift')}
									</Badge>
								{/if}
							</div>
						{:else}
							<p class="text-sm text-muted-foreground italic">{$_('admin.attendance.detail.no_shift_config')}</p>
						{/if}
					</div>
				</section>
				
				<!-- Audit Metadata -->
				<section class="bg-muted/20 border rounded-xl p-4 text-xs space-y-3" in:fly={{ y: 20, duration: 800, easing: quintOut, delay: 500 }}>
					<div class="flex items-center gap-2 font-semibold text-muted-foreground uppercase mb-1">
						<History class="w-3 h-3" />
						{$_('admin.attendance.detail.audit_metadata')}
					</div>
					<div class="flex justify-between">
						<span>{$_('admin.attendance.detail.record_id')}:</span>
						<span class="font-mono">{data.attendance.id}</span>
					</div>
					<div class="flex justify-between">
						<span>{$_('admin.attendance.detail.created_at')}:</span>
						<span>{new Date(data.attendance.check_in || data.attendance.date).toLocaleString()}</span>
					</div>
					<div class="flex justify-between">
						<span>{$_('admin.attendance.detail.last_modified')}:</span>
						<span>{new Date(data.attendance.check_out || data.attendance.check_in || data.attendance.date).toLocaleString()}</span>
					</div>
				</section>
			</div>
		</div>
	{/if}
</div>

<style>
	:global(.container) {
		max-width: 1200px;
	}
</style>
