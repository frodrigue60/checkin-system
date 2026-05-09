<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { fly } from 'svelte/transition';
	import { apiFetch } from '$lib/api';
	import { authState, type User } from '$lib/auth.svelte';
	import { Button } from '$lib/components/ui/button';
	import { Loader2 } from 'lucide-svelte';
	import AbsenceModal from '$lib/components/AbsenceModal.svelte';
	import { notifications } from '$lib/notifications.svelte';
	import { _, locale } from 'svelte-i18n';
	import { translateError } from '$lib/i18n/error-translator';
	import * as Dialog from '$lib/components/ui/dialog';
	import Badge from '$lib/components/ui/badge/badge.svelte';
	import type { Attendance, WorkShift, WorkCenter } from '$lib/types/models';
	import { getCurrentLocation, reverseGeocode } from '$lib/services/location';

	let loading = $state(false);
	let error = $state('');
	let checkingAuth = $state(true);
	let currentUser = $state<User | null>(null);
	let mounted = $state(false);

	// Dashboard State
	let todayAttendance = $state<Attendance | null>(null);
	let todayHistory = $state<Attendance[]>([]);
	let currentShift = $state<WorkShift | null>(null);
	let goalSeconds = $state(8 * 3600);
	let isEmployee = $state(false);
	let currentTime = $state(new Date());
	let showAbsenceModal = $state(false);
	let absenceLoading = $state(false);
	let workCenters = $state<WorkCenter[]>([]);
	let selectedWorkCenterId = $state<number | null>(null);
	let isFieldWork = $state(false);
	let fieldWorkNote = $state('');

	let showEvidenceModal = $state(false);
	let evidencePreviews = $state<string[]>([]);
	let evidenceFiles = $state<File[]>([]);
	let fileInput: HTMLInputElement;
	let evidenceUploading = $state(false);
	let serviceNote = $state('');
	let stats = $derived.by(() => {
		const formatMs = (ms: number) => {
			if (isNaN(ms) || ms < 0) return '00:00';
			const h = Math.floor(ms / 3600000);
			const m = Math.floor((ms % 3600000) / 60000);
			return `${h.toString().padStart(2, '0')}:${m.toString().padStart(2, '0')}`;
		};

		if (!todayAttendance) return { worked: '00:00', remaining: formatMs(goalSeconds * 1000) };

		// Robust Date parsing
		const parseDate = (val: any) => {
			if (!val) return null;
			const d = new Date(val);
			return isNaN(d.getTime()) ? null : d;
		};

		const checkIn = parseDate(todayAttendance.check_in);
		if (!checkIn) return { worked: '00:00', remaining: formatMs(goalSeconds * 1000) };

		const checkOut = parseDate(todayAttendance.check_out);
		const now = checkOut || currentTime;
		let diffMs = now.getTime() - checkIn.getTime();

		// Subtract lunch if exists
		const lStart = parseDate(todayAttendance.lunch_start);
		const lEnd = parseDate(todayAttendance.lunch_end);

		if (lStart && lEnd) {
			diffMs -= lEnd.getTime() - lStart.getTime();
		} else if (lStart) {
			diffMs -= currentTime.getTime() - lStart.getTime();
		}

		let goalMs = goalSeconds * 1000;
		// If overtime, goal effectively increases to avoid negative remaining
		if (diffMs > goalMs) goalMs = diffMs;

		const remMs = Math.max(0, goalMs - diffMs);
		return { worked: formatMs(diffMs), remaining: formatMs(remMs) };
	});

	onMount(() => {
		mounted = true;
		if (authState.token && authState.user) {
			currentUser = authState.user;
			fetchTodayStatus();
			fetchWorkCenters();
		}
		checkingAuth = false;

		const timer = setInterval(() => {
			currentTime = new Date();
		}, 1000);

		return () => clearInterval(timer);
	});

	async function fetchTodayStatus() {
		try {
			const res = await apiFetch<{
				is_employee: boolean;
				attendance: Attendance;
				history: Attendance[];
				shift: WorkShift;
				goal_seconds: number;
			}>('/attendance/today');
			if (res.ok) {
				const data = await res.json();
				isEmployee = data.is_employee;
				todayAttendance = data.attendance;
				todayHistory = data.history || [];
				currentShift = data.shift;
				goalSeconds = data.goal_seconds || 8 * 3600;
			}
		} catch (e) {
			console.error('Failed to fetch attendance status');
		}
	}

	async function fetchWorkCenters() {
		try {
			const res = await apiFetch('/attendance/centers'); // New endpoint or public centers list
			if (res.ok) {
				workCenters = await res.json();
			}
		} catch (e) {
			console.error('Failed to fetch centers');
		}
	}

	async function performAction(actionType: string) {
		if (actionType === 'report-absence') {
			showAbsenceModal = true;
			return;
		}

		loading = true;
		error = '';
		try {
			// Layered location retrieval
			const pos = await getCurrentLocation();

			let address = '';
			if (actionType === 'check-in' && isFieldWork) {
				address = await reverseGeocode(pos.latitude, pos.longitude);
			}

			const body: any = {
				employee_id: authState.user?.employee_id || authState.user?.id,
				latitude: pos.latitude,
				longitude: pos.longitude,
				work_center_id: actionType === 'check-in' ? selectedWorkCenterId : undefined,
				is_field_work: isFieldWork,
				check_in_note: isFieldWork ? fieldWorkNote : undefined,
				address: address || undefined
			};

			if (actionType === 'check-out') {
				loading = false;
				// Pre-fill check-out note with check-in note as a draft if it exists
				if (todayAttendance?.check_in_note) {
					serviceNote = todayAttendance.check_in_note;
				}
				showEvidenceModal = true;
				return;
			}

			const res = await apiFetch<{ attendance: Attendance; error?: string }>(`/attendance/${actionType}`, {
				method: 'POST',
				body: JSON.stringify(body)
			});

			const data = await res.json();
			if (!res.ok) throw new Error(translateError(data.error || 'Action failed'));

			// Success feedback
			switch (actionType) {
				case 'check-in':
					notifications.success($_('success.checkin'));
					break;
				case 'check-out':
					notifications.success($_('success.checkout'));
					break;
				case 'lunch-start':
					notifications.success($_('success.lunch_start'));
					break;
				case 'lunch-end':
					notifications.success($_('success.lunch_end'));
					break;
			}

			// Apply instant reactivity
			if (data.attendance) {
				todayAttendance = data.attendance;
				isFieldWork = false; // Reset toggle
				fieldWorkNote = '';  // Reset note
			}

			// Fallback background sync
			fetchTodayStatus();
		} catch (err: any) {
			error = err.message;
			notifications.error(error);
		} finally {
			loading = false;
		}
	}

	async function handleEvidenceConfirm() {
		if (evidencePreviews.length < 2) {
			notifications.error($_('dashboard.min_evidence_required', { min: 2 }));
			return;
		}

		evidenceUploading = true;
		try {
			// 1. Get Presigned URLs from Backend
			let finalUrls: string[] = [];
			if (evidenceFiles.length > 0) {
				const fileData = evidenceFiles.map(f => ({ name: f.name, type: f.type }));
				
				const uploadRes = await apiFetch<{ items: { upload_url: string, public_url: string, key: string, file_name: string }[] }>('/attendance/upload', {
					method: 'POST',
					body: JSON.stringify({ files: fileData })
				});

				if (!uploadRes.ok) {
					const errData = await uploadRes.json();
					throw new Error(errData.error || 'Failed to get upload permissions');
				}
				
				const { items } = await uploadRes.json();

				// 2. Upload directly to Cloudflare R2 using PUT
				const uploadPromises = items.map(async (item, index) => {
					const file = evidenceFiles[index];
					const res = await fetch(item.upload_url, {
						method: 'PUT',
						body: file
					});
					if (!res.ok) throw new Error(`Failed to upload ${file.name}`);
					return item.key;
				});

				finalUrls = await Promise.all(uploadPromises);
			}

			// Get position again for precision
			const pos = await getCurrentLocation();

			const body = {
				employee_id: authState.user?.employee_id || authState.user?.id,
				latitude: pos.latitude,
				longitude: pos.longitude,
				evidence_urls: finalUrls,
				check_out_note: serviceNote,
				address: "" 
			};

			const res = await apiFetch<{ attendance: Attendance }>(`/attendance/check-out`, {
				method: 'POST',
				body: JSON.stringify(body)
			});

			const data = await res.json();
			if (!res.ok) throw new Error(translateError(data.error || 'Check-out failed'));

			notifications.success($_('success.checkout'));
			todayAttendance = data.attendance;
			showEvidenceModal = false;
			evidencePreviews = [];
			evidenceFiles = [];
			serviceNote = '';
			fetchTodayStatus();
		} catch (err: any) {
			notifications.error(err.message);
		} finally {
			evidenceUploading = false;
		}
	}

	function addEvidence() {
		fileInput?.click();
	}

	function handleFileChange(event: Event) {
		const target = event.target as HTMLInputElement;
		if (!target.files) return;

		const files = Array.from(target.files);
		
		for (const file of files) {
			if (evidencePreviews.length >= 4) break;
			
			// Validate type
			if (!file.type.startsWith('image/')) {
				notifications.error(`${file.name} is not an image`);
				continue;
			}

			// Max 5MB per image
			if (file.size > 5 * 1024 * 1024) {
				notifications.error(`${file.name} exceeds 5MB`);
				continue;
			}

			const reader = new FileReader();
			reader.onload = (e) => {
				const preview = e.target?.result as string;
				evidencePreviews = [...evidencePreviews, preview];
				evidenceFiles = [...evidenceFiles, file];
			};
			reader.readAsDataURL(file);
		}
		
		// Reset input to allow selecting the same file again
		target.value = '';
	}

	function removeEvidence(index: number) {
		evidencePreviews = evidencePreviews.filter((_, i) => i !== index);
		evidenceFiles = evidenceFiles.filter((_, i) => i !== index);
	}

	async function handleAbsenceConfirm(reason: string) {
		absenceLoading = true;
		try {
			const body = {
				employee_id: authState.user?.employee_id || authState.user?.id,
				reason: reason
			};

			const res = await apiFetch('/attendance/report-absence', {
				method: 'POST',
				body: JSON.stringify(body)
			});

			const data = await res.json();
			if (!res.ok) throw new Error(translateError(data.error || 'Failed to report absence'));

			if (data.attendance) {
				todayAttendance = data.attendance;
			}

			showAbsenceModal = false;
			notifications.success($_('success.absence'));
			fetchTodayStatus();
		} catch (err: any) {
			notifications.error(err.message);
		} finally {
			absenceLoading = false;
		}
	}

	function formatTime(isoString: string | null) {
		if (!isoString) return '--:--';
		const date = new Date(isoString);
		return date.toLocaleTimeString($locale || 'es-ES', {
			hour12: false,
			hour: '2-digit',
			minute: '2-digit'
		});
	}
</script>

<svelte:head>
	<title
		>{authState.isAuthenticated
			? `${$_('nav.activity')} | ${$_('landing.system_name')}`
			: `${$_('common.welcome')} | ${$_('landing.system_name')}`}</title
	>
</svelte:head>

{#if !checkingAuth && mounted}
	{#if !authState.isAuthenticated}
		<!-- GUEST LANDING VIEW -->
		<main class="min-h-screen flex flex-col md:flex-row relative overflow-hidden bg-surface">
			<!-- Left Side: Editorial Brand Pillar (Desktop Only) -->
			<section class="hidden md:flex md:w-5/12 bg-primary p-16 flex-col justify-between relative">
				<div class="absolute inset-0 opacity-10 pointer-events-none overflow-hidden">
					<img
						class="w-full h-full object-cover"
						alt="Corporate architecture"
						src="./images/login-bg.png"
					/>
				</div>
				<div class="relative z-10">
					<div class="flex items-center gap-3 mb-12">
						<div class="w-8 h-8 bg-on-primary rounded-lg flex items-center justify-center">
							<span
								class="material-symbols-outlined text-primary text-xl"
								style="font-variation-settings: 'FILL' 1;">corporate_fare</span
							>
						</div>
						<span class="font-display font-black text-on-primary tracking-tighter text-2xl"
							>{$_('landing.system_name')}</span
						>
					</div>
					<h1
						class="font-display text-5xl lg:text-7xl font-extrabold text-on-primary leading-tight tracking-tighter mb-8"
					>
						{$_('landing.hero_title')}
					</h1>
				</div>
				<div class="relative z-10">
					<div class="h-1 w-24 bg-primary-fixed-dim/30 mb-6"></div>
					<div class="text-on-primary/50 font-label text-xs uppercase tracking-widest">
						{$_('landing.authorized_access')}
					</div>
				</div>
			</section>

			<!-- Right Side: Interaction Portal -->
			<section
				class="flex-1 flex flex-col justify-center items-center px-6 py-12 bg-white relative"
			>
				<div class="w-full max-w-sm text-center space-y-12">
					<header class="space-y-4">
						<div class="md:hidden flex justify-center mb-8">
							<span class="material-symbols-outlined text-primary text-5xl">corporate_fare</span>
						</div>
						<h2 class="font-display text-4xl font-black text-primary tracking-tight">
							{$_('landing.access_terminal')}
						</h2>
						<p class="text-on-surface-variant font-medium">
							{$_('landing.auth_description')}
						</p>
					</header>

					<nav class="space-y-4">
						<Button
							href="/login"
							class="w-full h-16 rounded-sm font-display font-black text-lg tracking-widest uppercase shadow-xl shadow-primary/20"
						>
							{$_('common.login')}
						</Button>
						<Button
							href="/register"
							variant="outline"
							class="w-full h-16 rounded-sm font-display font-black text-lg tracking-widest uppercase border-2 text-primary hover:bg-primary/5"
						>
							{$_('common.register')}
						</Button>
					</nav>

					<footer class="pt-12 opacity-30 select-none pointer-events-none">
						<div class="flex items-center justify-center gap-8">
							<!-- <img
								class="h-6 grayscale brightness-0"
								alt="Attendance System Logo"
								src="https://lh3.googleusercontent.com/aida-public/AB6AXuB0I9uFOjgURibwwI25NlDU9jqhAb_jvcGARcvlZH5kF4k0j6_qboIVwxPPUaW7Wqo4Uqi1odbxxUWOEMBSnwiQGvzL5UYB7gBpX9R7K_KpSSy6IVtejhwx_Jlyth5Gxio_m4PF_4e5pxciAoYpMSSUzKdffDPpJT-EncvZ0BllFYjwvTeAG7bZ4LAhFx6Ayvq9MdZjKEBOqReiQgyPpslKq36t3CAcavUqK7mbvVPAyMkc5BZr_2Bvkqu1lPsvNrTMaPC67rsrKws"
							/> -->
							<span class="material-symbols-outlined text-primary text-5xl">corporate_fare</span>
						</div>
					</footer>
				</div>
			</section>
		</main>
	{:else}
		<!-- LOGGED IN ACTIVITY DASHBOARD VIEW -->
		<div class="bg-background text-on-surface font-body pb-32 pt-6">
			<main class="px-6 pt-4 max-w-2xl mx-auto space-y-12">
				<!-- Minimal Hero Section -->
				<section class="text-center pt-4">
					<div class="flex flex-col items-center">
						<div class="flex items-baseline gap-2 mb-2">
							<h2 class="font-display font-black text-7xl text-primary tracking-tighter">
								{currentTime.toLocaleTimeString($locale || 'es-ES', {
									hour12: false,
									hour: '2-digit',
									minute: '2-digit'
								})}
							</h2>
							<span class="font-display font-bold text-2xl text-primary/30 uppercase">
								{currentTime.toLocaleTimeString($locale || 'es-ES', { hour12: true }).slice(-2)}
							</span>
						</div>
						<div class="flex items-center gap-2 px-3 py-1 bg-slate-100 rounded-full">
							<div
								class="w-1.5 h-1.5 rounded-full {todayAttendance?.is_absence
									? 'bg-amber-500'
									: todayAttendance?.check_in && !todayAttendance?.check_out
										? 'bg-emerald-500 animate-pulse'
										: 'bg-slate-400'}"
							></div>
							<p
								class="font-label text-[10px] font-bold uppercase tracking-[0.15em] {todayAttendance?.is_absence
									? 'text-amber-600'
									: todayAttendance?.check_in && !todayAttendance?.check_out
										? 'text-emerald-600'
										: 'text-slate-500'}"
							>
								{#if todayAttendance?.is_absence}
									{$_('dashboard.justified_absence')}
								{:else if todayAttendance?.check_in && !todayAttendance?.check_out}
									{$_('dashboard.active_session')} • {$_('dashboard.in_office')}
								{:else}
									{$_('dashboard.not_started')} • {$_('dashboard.off_duty')}
								{/if}
							</p>
						</div>
					</div>
				</section>

				<!-- Primary Action Center -->
				<section class="space-y-6">
					{#if todayAttendance?.is_absence}
						<!-- Absence Status Card -->
						<div
							class="w-full py-10 bg-amber-50 border border-amber-100 text-amber-900 rounded-sm font-bold flex flex-col items-center justify-center gap-4 transition-all shadow-xl shadow-amber-500/5"
						>
							<div class="w-12 h-12 bg-amber-100 rounded-full flex items-center justify-center">
								<span class="material-symbols-outlined text-amber-600 text-3xl">event_busy</span>
							</div>
							<div class="text-center space-y-1">
								<span class="text-xl uppercase tracking-[0.2em] block"
									>{$_('dashboard.absence_reported')}</span
								>
								<p class="text-[10px] text-amber-600/70 font-bold max-w-[200px] italic">
									"{todayAttendance.absence_reason}"
								</p>
							</div>
						</div>
					{:else if !todayAttendance?.check_in || !!todayAttendance?.check_out}
						<!-- Check In Button (Show even if previously checked out today) -->
						<div class="space-y-4">
							{#if currentShift?.shift_type === 'field'}
								<div class="space-y-2">
									<label class="text-[10px] font-black uppercase tracking-widest text-slate-400 pl-1" for="center-select">
										{$_('dashboard.select_location')}
									</label>
									<select 
										id="center-select"
										bind:value={selectedWorkCenterId}
										class="w-full p-4 bg-white border-2 border-slate-100 rounded-sm font-bold text-slate-700 focus:border-primary outline-none transition-colors"
									>
										<option value={null}>{$_('dashboard.default_center')}</option>
										{#each workCenters as center}
											<option value={center.id}>{center.name}</option>
										{/each}
									</select>
								</div>
							{/if}

							<!-- Field Work Toggle -->
							<div class="flex items-center justify-between p-4 bg-slate-50 border border-slate-100 rounded-sm">
								<div class="flex items-center gap-3">
									<div class="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center">
										<span class="material-symbols-outlined text-primary text-lg">distance</span>
									</div>
									<div class="flex flex-col">
										<span class="text-[10px] font-black uppercase tracking-widest text-slate-700">{$_('dashboard.field_work_mode')}</span>
										<span class="text-[8px] font-bold text-slate-400">{$_('dashboard.field_work_description')}</span>
									</div>
								</div>
								<button 
									class="w-12 h-6 rounded-full relative transition-colors {isFieldWork ? 'bg-primary' : 'bg-slate-200'}"
									onclick={() => isFieldWork = !isFieldWork}
								>
									<div class="absolute top-1 left-1 w-4 h-4 bg-white rounded-full transition-transform {isFieldWork ? 'translate-x-6' : ''}"></div>
								</button>
							</div>
							
							{#if isFieldWork}
								<div class="space-y-2" in:fly={{ y: -10, duration: 300 }}>
									<label class="text-[10px] font-black uppercase tracking-widest text-slate-400 pl-1">
										{$_('dashboard.field_work_note_label')}
									</label>
									<textarea 
										bind:value={fieldWorkNote}
										class="w-full min-h-[80px] p-4 bg-white border-2 border-slate-100 rounded-sm text-xs font-bold focus:border-primary outline-none transition-colors font-sans resize-none"
										placeholder={$_('dashboard.field_work_note_placeholder')}
									></textarea>
								</div>
							{/if}

							<button
								class="w-full py-8 bg-primary text-primary-foreground rounded-sm font-bold flex flex-col items-center justify-center gap-3 active:scale-[0.98] transition-all shadow-xl shadow-green-500/10 disabled:opacity-50"
								onclick={() => performAction('check-in')}
								disabled={loading}
							>
								<span
									class="material-symbols-outlined text-4xl"
									style="font-variation-settings: 'FILL' 1;">login</span>
								<span class="text-xl uppercase tracking-widest">
									{todayAttendance?.check_out ? $_('dashboard.new_checkin') : $_('common.check_in')}
								</span>
							</button>
						</div>
					{:else if !todayAttendance?.check_out}
						<!-- Clock Out Button -->
						<button
							class="w-full py-8 bg-[#dbeafe] text-[#1e40af] rounded-sm font-bold flex flex-col items-center justify-center gap-3 active:scale-[0.98] transition-all shadow-xl shadow-blue-500/10 disabled:opacity-50"
							onclick={() => performAction('check-out')}
							disabled={loading}
						>
							<span
								class="material-symbols-outlined text-4xl"
								style="font-variation-settings: 'FILL' 1;">logout</span
							>
							<span class="text-xl uppercase tracking-widest">{$_('dashboard.clock_out')}</span>
						</button>
					{:else}
						<!-- Day Completed -->
						<div
							class="w-full py-8 bg-slate-100 text-slate-400 rounded-sm font-bold flex flex-col items-center justify-center gap-3 border border-slate-200 cursor-not-allowed"
						>
							<span class="material-symbols-outlined text-4xl">task_alt</span>
							<span class="text-xl uppercase tracking-widest"
								>{$_('dashboard.shift_completed')}</span
							>
						</div>
					{/if}
					<!-- Secondary Actions -->
					<div class="grid grid-cols-2 gap-4">
						<!-- Lunch Start -->
						<button
							class="py-6 rounded-sm font-bold flex flex-col items-center justify-center gap-2 transition-all {!todayAttendance?.lunch_start &&
							todayAttendance?.check_in &&
							!todayAttendance?.check_out
								? 'bg-[#FEF3C7] text-[#B45309] active:scale-95'
								: 'bg-slate-50 text-slate-400 border border-slate-200'}"
							onclick={() => performAction('lunch-start')}
							disabled={loading ||
								!!todayAttendance?.lunch_start ||
								!todayAttendance?.check_in ||
								!!todayAttendance?.check_out ||
								todayAttendance?.is_absence}
						>
							<span class="material-symbols-outlined text-2xl font-bold">restaurant</span>
							<span class="text-[10px] uppercase tracking-widest font-bold"
								>{$_('common.lunch_start')}</span
							>
						</button>

						<!-- Lunch End -->
						<button
							class="py-6 rounded-sm font-bold flex flex-col items-center justify-center gap-2 transition-all {!todayAttendance?.lunch_end &&
							todayAttendance?.lunch_start &&
							!todayAttendance?.check_out
								? 'bg-[#D1FAE5] text-[#059669] active:scale-95'
								: 'bg-slate-50 text-slate-400 border border-slate-200'}"
							onclick={() => performAction('lunch-end')}
							disabled={loading ||
								!!todayAttendance?.lunch_end ||
								!todayAttendance?.lunch_start ||
								!!todayAttendance?.check_out ||
								todayAttendance?.is_absence}
						>
							<span class="material-symbols-outlined text-2xl font-bold">restaurant</span>
							<span class="text-[10px] uppercase tracking-widest font-bold"
								>{$_('common.lunch_end')}</span
							>
						</button>
					</div>
					<div>
						<button
							class="w-full py-2 {todayAttendance
								? 'bg-slate-100 text-slate-400 border border-slate-200'
								: 'bg-red-400 text-red-900 shadow-xl shadow-red-500/10'} rounded-sm font-bold flex items-center justify-center gap-3 active:scale-[0.98] transition-all disabled:opacity-50"
							onclick={() => performAction('report-absence')}
							disabled={loading || !!todayAttendance}
						>
							<span
								class="material-symbols-outlined text-xl"
								style="font-variation-settings: 'FILL' {todayAttendance ? 0 : 1};"
							>
								{todayAttendance?.is_absence ? 'check_circle' : 'report_problem'}
							</span>
							<span class="text-[10px] uppercase tracking-widest">
								{todayAttendance?.is_absence
									? $_('dashboard.absence_reported')
									: todayAttendance
										? $_('dashboard.day_recorded')
										: $_('dashboard.report_absence')}
							</span>
						</button>
					</div>
				</section>

				<!-- Simplified Stats & Context -->
				<section class="grid grid-cols-2 gap-8 py-8 border-y border-slate-200">
					<div class="text-center">
						<p class="text-4xl font-display font-black text-primary tabular-nums">{stats.worked}</p>
						<p class="text-[10px] text-slate-400 font-black uppercase tracking-widest mt-2">
							{$_('dashboard.logged_hours')}
						</p>
					</div>
					<div class="text-center">
						{#if todayAttendance?.check_out}
							<p class="text-4xl font-display font-black text-emerald-500 tabular-nums">
								<span class="material-symbols-outlined text-4xl">check_circle</span>
							</p>
							<p class="text-[10px] text-emerald-600 font-black uppercase tracking-widest mt-2">
								{$_('dashboard.shift_completed')}
							</p>
						{:else}
							<p class="text-4xl font-display font-black text-slate-500 tabular-nums">
								{stats.remaining}
							</p>
							<p class="text-[10px] text-slate-400 font-black uppercase tracking-widest mt-2">
								{$_('dashboard.remaining')}
							</p>
						{/if}
					</div>
				</section>

				<!-- Timeline Summary - Minimalist -->
				<section class="pb-8">
					<div class="flex justify-between items-center mb-8">
						<h3 class="font-display font-black text-xs text-primary uppercase tracking-[0.2em]">
							{$_('dashboard.today_timeline')}
						</h3>
						<span class="font-label text-[10px] font-bold text-slate-400 uppercase tracking-widest">
							{new Date().toLocaleDateString($locale || 'es-ES', {
								month: 'short',
								day: 'numeric',
								year: 'numeric'
							})}
						</span>
					</div>
					<div class="space-y-4">
						{#if todayHistory.length > 0}
							{#each todayHistory as session (session.id)}
								<div class="border rounded-sm p-4 bg-white shadow-sm space-y-3">
									<div class="flex justify-between items-center border-b pb-2">
										<span class="text-[10px] font-black uppercase tracking-widest text-primary">
											{session.is_absence ? $_('dashboard.absence') : $_('dashboard.session')} #{session.id}
										</span>
										{#if session.check_in && session.check_out}
											<Badge variant="outline" class="text-[9px] bg-emerald-50 text-emerald-700 border-emerald-100">
												{$_('dashboard.completed')}
											</Badge>
										{:else if session.check_in}
											<Badge variant="outline" class="text-[9px] bg-blue-50 text-blue-700 border-blue-100 animate-pulse">
												{$_('dashboard.active')}
											</Badge>
										{/if}
									</div>

									{#if session.is_absence}
										<div class="flex items-center gap-4 py-2">
											<span class="material-symbols-outlined text-amber-500">event_busy</span>
											<div class="flex flex-col">
												<span class="text-xs font-black text-on-surface uppercase tracking-tight">{$_('dashboard.justified_absence')}</span>
												<span class="text-[10px] font-bold text-slate-400 italic">"{session.absence_reason}"</span>
											</div>
										</div>
									{:else}
										<div class="grid grid-cols-2 gap-4">
											<div class="flex items-center gap-3">
												<span class="text-xs font-black text-primary tabular-nums w-10">{formatTime(session.check_in)}</span>
												<span class="text-[10px] font-bold uppercase text-slate-400">{$_('common.check_in')}</span>
											</div>
											{#if session.check_out}
												<div class="flex items-center gap-3">
													<span class="text-xs font-black text-rose-500 tabular-nums w-10">{formatTime(session.check_out)}</span>
													<span class="text-[10px] font-bold uppercase text-slate-400">{$_('common.check_out')}</span>
												</div>
											{/if}
										</div>

										{#if session.lunch_start}
											<div class="flex items-center gap-3 pt-1 opacity-70">
												<span class="material-symbols-outlined text-xs text-amber-600">restaurant</span>
												<span class="text-[9px] font-bold uppercase text-slate-400">
													{$_('common.lunch')}: {formatTime(session.lunch_start)} 
													{session.lunch_end ? ` - ${formatTime(session.lunch_end)}` : '...'}
												</span>
											</div>
										{/if}
									{/if}
								</div>
							{/each}
						{:else}
							<div class="py-12 text-center bg-white rounded-sm border border-slate-200">
								<span class="material-symbols-outlined text-3xl text-slate-300 mb-2">pending_actions</span>
								<p class="text-[10px] font-bold uppercase tracking-widest text-slate-400">
									{$_('dashboard.no_activity')}
								</p>
							</div>
						{/if}
					</div>
				</section>
			</main>
		</div>
	{/if}

	<!-- Evidence Modal -->
	<Dialog.Root bind:open={showEvidenceModal}>
		<Dialog.Content class="sm:max-w-md bg-white border-none shadow-2xl p-0 overflow-hidden">
			<div class="p-8 space-y-6">
				<div class="space-y-2">
					<h3 class="text-2xl font-black tracking-tighter text-primary uppercase">{$_('dashboard.checkout_evidence')}</h3>
					<p class="text-[10px] font-black text-slate-400 uppercase tracking-widest">
						Se requiere evidencia fotográfica para cerrar la sesión.
					</p>
				</div>

				<div class="space-y-3">
					<div class="flex justify-between items-end px-1">
						<div class="flex flex-col">
							<p class="text-[10px] font-black text-slate-500 uppercase tracking-widest">
								{$_('dashboard.checkout_evidence')}
							</p>
							<p class="text-[9px] font-bold text-slate-400">
								{$_('dashboard.min_evidence_required', { min: 2 })}
							</p>
						</div>
						<div class="px-2 py-0.5 rounded-full bg-slate-100 text-[9px] font-black text-slate-500 uppercase">
							{$_('dashboard.evidence_count', { count: evidencePreviews.length, min: 2 })}
						</div>
					</div>

					<div class="grid grid-cols-2 gap-2">
						{#each evidencePreviews as preview, i}
							<div class="aspect-video bg-slate-50 rounded-sm border border-slate-100 overflow-hidden relative group" in:fly={{ y: 10, duration: 400, delay: i * 100 }}>
								<img src={preview} class="w-full h-full object-cover" alt="Evidence {i+1}" />
								<div class="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center">
									<button 
										class="bg-rose-500 text-white p-1.5 rounded-full shadow-lg transform scale-75 group-hover:scale-100 transition-transform"
										onclick={() => removeEvidence(i)}
									>
										<span class="material-symbols-outlined text-sm">close</span>
									</button>
								</div>
							</div>
						{/each}

						{#if evidencePreviews.length < 4}
							<button 
								class="aspect-video border-2 border-dashed border-slate-200 rounded-sm flex flex-col items-center justify-center gap-1 text-slate-400 hover:border-primary hover:text-primary transition-all group"
								onclick={addEvidence}
							>
								<span class="material-symbols-outlined text-2xl group-hover:scale-110 transition-transform">add_a_photo</span>
								<span class="text-[8px] font-black uppercase tracking-widest">{$_('dashboard.add_another_photo')}</span>
							</button>
						{/if}

						<input
							type="file"
							accept="image/*"
							multiple
							class="hidden"
							bind:this={fileInput}
							onchange={handleFileChange}
						/>
					</div>
				</div>

				<div class="space-y-2">
					<label class="text-[10px] font-black text-slate-500 uppercase tracking-widest px-1">
						{$_('dashboard.service_note_label')}
					</label>
					<textarea 
						bind:value={serviceNote}
						class="w-full min-h-[80px] p-4 bg-slate-50 border border-slate-100 rounded-sm text-xs font-bold focus:ring-2 focus:ring-primary/20 transition-all font-sans resize-none"
						placeholder={$_('dashboard.service_note_placeholder')}
					></textarea>
				</div>

				<div class="flex gap-3 pt-2">
					<Button 
						variant="ghost" 
						class="flex-1 h-14 font-black text-[10px] uppercase tracking-widest text-slate-400"
						onclick={() => {
							showEvidenceModal = false;
							evidencePreviews = [];
						}}
					>
						Cancelar
					</Button>
					<Button 
						class="flex-1 h-14 bg-primary text-white font-black text-[10px] uppercase tracking-widest shadow-xl shadow-primary/20 disabled:opacity-50 disabled:grayscale transition-all"
						disabled={evidencePreviews.length < 2 || evidenceUploading}
						onclick={handleEvidenceConfirm}
					>
						{#if evidenceUploading}
							<Loader2 class="animate-spin mr-2" size={16} />
						{/if}
						{$_('dashboard.confirm_checkout_final')}
					</Button>
				</div>
			</div>
		</Dialog.Content>
	</Dialog.Root>
{/if}

<AbsenceModal
	bind:isOpen={showAbsenceModal}
	onConfirm={handleAbsenceConfirm}
	loading={absenceLoading}
/>

<style>
	:global(.font-display) {
		font-family: 'Public Sans', sans-serif;
	}

	.material-symbols-outlined {
		font-variation-settings:
			'FILL' 0,
			'wght' 400,
			'GRAD' 0,
			'opsz' 24;
	}

	button:disabled {
		cursor: not-allowed;
	}
</style>
