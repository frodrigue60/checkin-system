<script lang="ts">
	import { onMount } from 'svelte';
	import { apiFetch } from '$lib/api';
	import { authState } from '$lib/auth.svelte';
	import { fade, fly } from 'svelte/transition';
	import { _, locale } from 'svelte-i18n';

	let history: any[] = $state([]);
	let stats = $state({ total_hours: 0, attendance_rate: 0 });
	let loading = $state(true);
	let selectedDate = $state(new Date());

	// Justification Modal State
	let showModal = $state(false);
	let selectedIncident = $state<any>(null);
	let justificationMessage = $state('');
	let submitting = $state(false);

	function openJustifyModal(incident: any) {
		selectedIncident = incident;
		justificationMessage = '';
		showModal = true;
	}

	function closeModal() {
		showModal = false;
		selectedIncident = null;
	}

	async function submitJustification() {
		submitting = true;
		try {
			const res = await apiFetch('/user/justify', {
				method: 'POST',
				body: JSON.stringify({
					incident_id: selectedIncident.id,
					message: justificationMessage
				})
			});

			if (res.ok) {
				closeModal();
				loadHistory();
			} else {
				const data = await res.json();
				alert(data.error || 'Error al enviar justificación');
			}
		} catch (e) {
			console.error('Failed to submit justification', e);
		} finally {
			submitting = false;
		}
	}

	let monthName = $derived(
		new Intl.DateTimeFormat($locale || 'es-ES', { month: 'long' }).format(selectedDate)
	);

	async function loadHistory() {
		loading = true;
		try {
			const month = selectedDate.getMonth() + 1;
			const year = selectedDate.getFullYear();
			const res = await apiFetch(`/user/history?month=${month}&year=${year}`);
			if (res.ok) {
				const data = await res.json();
				history = data.history;
				stats = data.stats;
			}
		} catch (e) {
			console.error('Failed to load history', e);
		} finally {
			loading = false;
		}
	}

	function changeMonth(delta: number) {
		const newDate = new Date(selectedDate);
		newDate.setMonth(newDate.getMonth() + delta);
		selectedDate = newDate;
		loadHistory();
	}

	function formatTime(isoString: string | null) {
		if (!isoString) return '--:--';
		return new Date(isoString).toLocaleTimeString([], {
			hour: '2-digit',
			minute: '2-digit',
			hour12: false
		});
	}

	function formatDate(isoString: string) {
		const d = new Date(isoString);
		const currentLocale = $locale || 'es-ES';
		return {
			day: d.getDate(),
			weekday: d.toLocaleDateString(currentLocale, { weekday: 'short' }).replace('.', ''),
			full: d.toLocaleDateString(currentLocale, { day: 'numeric', month: 'short' })
		};
	}

	onMount(() => {
		loadHistory();
	});
</script>

<svelte:head>
	<title>{$_('history.title')} | {$_('landing.system_name')}</title>
</svelte:head>

<div class="min-h-screen bg-surface pb-32">
	<!-- Main Content -->
	<main class="max-w-2xl mx-auto px-4 pt-8 space-y-8">
		<!-- Month Selector -->
		<section class="flex items-center justify-between" in:fly={{ y: -20, duration: 600 }}>
			<div>
				<p class="text-[10px] font-black uppercase tracking-[0.2em] text-on-surface-variant">
					{$_('history.monthly_log')}
				</p>
				<div class="flex items-center gap-3 mt-1">
					<h2 class="text-2xl font-black text-primary font-headline capitalize">
						{monthName}
						{selectedDate.getFullYear()}
					</h2>
					<div class="flex gap-1">
						<button
							class="p-1 hover:bg-surface-container-high rounded-full transition-colors text-primary"
							onclick={() => changeMonth(-1)}
						>
							<span class="material-symbols-outlined text-lg">chevron_left</span>
						</button>
						<button
							class="p-1 hover:bg-surface-container-high rounded-full transition-colors text-primary"
							onclick={() => changeMonth(1)}
						>
							<span class="material-symbols-outlined text-lg">chevron_right</span>
						</button>
					</div>
				</div>
			</div>

			<div class="flex gap-2 bg-surface-container-low p-1.5 rounded-sm">
				<button
					class="w-10 h-10 flex items-center justify-center bg-white shadow-sm rounded-sm text-primary"
				>
					<span class="material-symbols-outlined">calendar_today</span>
				</button>
				<button
					class="w-10 h-10 flex items-center justify-center text-outline-variant hover:text-primary transition-colors"
				>
					<span class="material-symbols-outlined">bar_chart</span>
				</button>
			</div>
		</section>

		<!-- Summary Stats -->
		<section class="grid grid-cols-2 gap-4" in:fade={{ delay: 200, duration: 600 }}>
			<div class="bg-[#F2F4F5] p-6 rounded-sm space-y-1 group hover:bg-[#E6E8E9] transition-colors">
				<p class="text-[10px] font-black uppercase tracking-widest text-on-surface-variant">
					{$_('history.total_hours')}
				</p>
				<p class="text-3xl font-black text-primary tracking-tighter tabular-nums">
					{Math.floor(stats.total_hours)}h {Math.round((stats.total_hours % 1) * 60)}m
				</p>
			</div>
			<div class="bg-[#F2F4F5] p-6 rounded-sm space-y-1 group hover:bg-[#E6E8E9] transition-colors">
				<p class="text-[10px] font-black uppercase tracking-widest text-on-surface-variant">
					{$_('history.attendance')}
				</p>
				<div class="flex items-end gap-2">
					<p class="text-3xl font-black text-primary tracking-tighter tabular-nums">
						{Math.round(stats.attendance_rate)}%
					</p>
					<span class="text-[10px] font-bold text-emerald-600 mb-1.5">+2%</span>
				</div>
			</div>
		</section>

		<!-- History List -->
		<section class="space-y-4">
			{#if loading}
				<div class="flex flex-col items-center justify-center py-20 gap-4 opacity-40">
					<div
						class="w-8 h-8 border-4 border-primary/20 border-t-primary rounded-full animate-spin"
					></div>
					<p class="text-[10px] font-black uppercase tracking-widest">{$_('history.syncing')}</p>
				</div>
			{:else if history.length === 0}
				<div
					class="flex flex-col items-center justify-center py-20 text-center space-y-4 bg-surface-container-low/30 rounded-sm border border-dashed border-outline-variant/30"
				>
					<span class="material-symbols-outlined text-4xl text-outline-variant opacity-20"
						>history</span
					>
					<div class="space-y-1">
						<p class="font-bold text-outline-variant">{$_('history.no_records')}</p>
						<p class="text-xs text-outline-variant/60 italic">
							{$_('history.no_records_subtitle')}
						</p>
					</div>
				</div>
			{:else}
				<div class="space-y-3">
					{#each history as item, i (item.id)}
						{@const dateInfo = formatDate(item.check_in)}
						<div
							class="group bg-white p-5 rounded-sm flex items-center justify-between shadow-sm hover:shadow-xl hover:shadow-primary/5 transition-all duration-300 border border-outline-variant/5"
							in:fly={{ x: -20, delay: i * 50, duration: 500 }}
						>
							<div class="flex items-center gap-5">
								<div
									class="w-14 h-14 bg-surface-container-low rounded-sm flex flex-col items-center justify-center group-hover:bg-primary/5 transition-colors"
								>
									<span
										class="text-[10px] font-black uppercase text-on-surface-variant leading-none"
										>{dateInfo.weekday}</span
									>
									<span class="text-xl font-black text-primary mt-0.5">{dateInfo.day}</span>
								</div>
								<div>
									<p class="text-sm font-bold text-primary">
										{formatTime(item.check_in)} - {formatTime(item.check_out)}
									</p>
									<p class="text-[11px] font-medium text-on-surface-variant italic mt-0.5">
										{item.work_center_name || $_('history.unknown_location')}
									</p>
								</div>
							</div>

							<div class="text-right space-y-1.5">
								<p class="text-sm font-black text-primary tabular-nums">
									{Math.floor(item.net_hours_worked)}h {Math.round(
										(item.net_hours_worked % 1) * 60
									)}m
								</p>

								{#if item.is_absence}
									<span
										class="inline-flex items-center gap-1.5 px-3 py-1 bg-surface-container-high text-on-surface-variant rounded-full text-[9px] font-black uppercase tracking-tighter"
									>
										<span class="material-symbols-outlined text-[12px]">event_busy</span>
										{$_('history.absence')}
									</span>
								{:else if item.incidents && item.incidents.length > 0}
									<div class="flex flex-col items-end gap-1">
										{#each item.incidents as incident (incident.id)}
											<div class="flex items-center gap-2">
												{#if incident.status === 'pending'}
													<button
														class="text-[9px] font-black uppercase text-primary hover:underline"
														onclick={() => openJustifyModal(incident)}
													>
														{$_('history.justify')}
													</button>
												{:else if incident.status === 'pending_review'}
													<span class="text-[9px] font-black uppercase text-amber-600 italic">
														{$_('history.justification_pending')}
													</span>
												{:else if incident.status === 'justified'}
													<span class="text-[9px] font-black uppercase text-emerald-600">
														{$_('history.justification_approved')}
													</span>
												{/if}
												<span
													class="inline-flex items-center gap-1.5 px-3 py-1 {incident.status ===
													'justified'
														? 'bg-emerald-50 text-emerald-700'
														: incident.status === 'pending_review'
															? 'bg-amber-50 text-amber-700'
															: 'bg-[#FEF3C7] text-[#B45309]'} rounded-full text-[9px] font-black uppercase tracking-tighter"
												>
													<div
														class="w-1 h-1 rounded-full {incident.status === 'justified'
															? 'bg-emerald-600'
															: incident.status === 'pending_review'
																? 'bg-amber-500'
																: 'bg-[#B45309]'}"
													></div>
													{$_(`incidents.${incident.type}`)}
												</span>
											</div>
										{/each}
									</div>
								{:else if !item.check_out && new Date(item.check_in).toDateString() === new Date().toDateString()}
									<span
										class="inline-flex items-center gap-1.5 px-3 py-1 bg-primary/5 text-primary rounded-full text-[9px] font-black uppercase tracking-tighter"
									>
										<div class="w-1 h-1 rounded-full bg-primary animate-pulse"></div>
										{$_('history.in_progress')}
									</span>
								{:else if !item.check_out}
									<span
										class="inline-flex items-center gap-1.5 px-3 py-1 bg-error-container/20 text-error rounded-full text-[9px] font-black uppercase tracking-tighter"
									>
										<div class="w-1 h-1 rounded-full bg-error"></div>
										{$_('history.incomplete')}
									</span>
								{:else}
									<span
										class="inline-flex items-center gap-1.5 px-3 py-1 bg-[#D1FAE5] text-[#059669] rounded-full text-[9px] font-black uppercase tracking-tighter"
									>
										<div class="w-1 h-1 rounded-full bg-[#059669]"></div>
										{$_('history.completed')}
									</span>
								{/if}
							</div>
						</div>
					{/each}
				</div>

				<div class="pt-4 text-center">
					<button
						class="px-8 py-3 bg-surface-container-low text-primary text-[10px] font-black uppercase tracking-widest rounded-sm hover:bg-surface-container-high transition-all active:scale-95 shadow-sm"
						onclick={loadHistory}
					>
						{$_('history.load_more')}
					</button>
				</div>
			{/if}
		</section>
	</main>
</div>

<!-- Justification Modal -->
{#if showModal}
	<div
		class="fixed inset-0 bg-black/60 z-50 flex items-end sm:items-center justify-center p-0 sm:p-4"
		transition:fade
	>
		<div
			class="bg-white w-full max-w-lg rounded-t-2xl sm:rounded-sm overflow-hidden flex flex-col shadow-2xl"
			in:fly={{ y: 100, duration: 500 }}
		>
			<div class="p-6 border-b border-outline-variant/10 flex items-center justify-between">
				<div class="space-y-1">
					<p class="text-[10px] font-black uppercase tracking-widest text-primary">
						{$_('history.send_justification')}
					</p>
					<h3 class="text-xl font-headline font-black text-primary">
						{$_(`incidents.${selectedIncident.type}`)}
					</h3>
				</div>
				<button class="p-2 hover:bg-surface-container-low rounded-full" onclick={closeModal}>
					<span class="material-symbols-outlined">close</span>
				</button>
			</div>

			<div class="p-6 space-y-6">
				<div class="space-y-2">
					<label
						for="msg"
						class="text-[10px] font-black uppercase tracking-widest text-outline-variant"
						>{$_('history.absence_reason')}</label
					>
					<textarea
						id="msg"
						bind:value={justificationMessage}
						class="w-full bg-surface-container-low p-4 rounded-sm border-none focus:ring-2 focus:ring-primary/20 text-sm font-medium h-32 placeholder:italic"
						placeholder={$_('history.justification_placeholder')}
					></textarea>
				</div>

				<button
					class="w-full py-4 bg-primary text-on-primary text-[11px] font-black uppercase tracking-[0.2em] rounded-sm hover:bg-primary/90 transition-all active:scale-[0.98] disabled:opacity-50"
					disabled={!justificationMessage || submitting}
					onclick={submitJustification}
				>
					{#if submitting}
						<div class="flex items-center justify-center gap-2">
							<div
								class="w-3 h-3 border-2 border-white/30 border-t-white rounded-full animate-spin"
							></div>
							<span>{$_('common.sending')}</span>
						</div>
					{:else}
						{$_('history.send_justification')}
					{/if}
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	:global(body) {
		background-color: #f8fafb;
	}
</style>
