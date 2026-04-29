<script lang="ts">
	import { onMount } from 'svelte';
	import { apiFetch } from '$lib/api';
	import { authState } from '$lib/auth.svelte';
	import { fade, fly } from 'svelte/transition';
	import { _ } from 'svelte-i18n';

	let profile = $state<any>(null);
	let loading = $state(true);
	let editing = $state(false);
	let saving = $state(false);
	let editForm = $state({ name: '', email: '', phone: '' });

	async function loadProfile() {
		try {
			const res = await apiFetch('/user/profile');
			if (res.ok) {
				profile = await res.json();
				editForm = {
					name: profile.name,
					email: profile.email,
					phone: profile.phone || ''
				};
			}
		} catch (e) {
			console.error('Failed to load profile', e);
		} finally {
			loading = false;
		}
	}

	async function saveProfile() {
		saving = true;
		try {
			const res = await apiFetch('/user/profile', {
				method: 'PUT',
				body: JSON.stringify(editForm)
			});
			if (res.ok) {
				await loadProfile();
				editing = false;
			} else {
				const err = await res.json();
				alert(err.error || 'Error updating profile');
			}
		} catch (e) {
			console.error('Error saving profile', e);
			alert('Network error');
		} finally {
			saving = false;
		}
	}

	onMount(() => {
		loadProfile();
	});

	function logout() {
		authState.logout();
		window.location.href = '/login';
	}

	function goBack() {
		window.history.back();
	}

	function formatShiftTime(t: string | null) {
		if (!t) return '--:--';
		if (t.includes('T')) return t.substring(11, 16);
		return t.substring(0, 5);
	}
</script>

<svelte:head>
	<title>{$_('profile.title')} | {$_('landing.system_name')}</title>
</svelte:head>

<div
	class="min-h-screen bg-background text-on-surface font-body selection:bg-primary-container selection:text-white"
>
	<main class="pb-32 px-6 max-w-3xl mx-auto">
		{#if loading}
			<div class="flex flex-col items-center justify-center py-20 gap-4 opacity-40">
				<div
					class="w-8 h-8 border-4 border-primary/20 border-t-primary rounded-full animate-spin"
				></div>
				<p class="text-[10px] font-black uppercase tracking-widest">{$_('profile.syncing')}</p>
			</div>
		{:else if profile}
			<!-- Top Section: Identity Anchor -->
			<section class="mb-12 flex flex-col md:flex-row items-center md:items-end gap-8" in:fade>
				<div class="relative group">
					<div
						class="w-32 h-32 rounded-sm overflow-hidden shadow-[0px_12px_32px_rgba(25,28,29,0.06)] bg-surface-container-high flex items-center justify-center text-primary bg-gradient-to-br from-surface-container-low to-surface-container-high"
					>
						<span class="text-5xl font-black font-headline">
							{profile.name?.charAt(0) || 'U'}
						</span>
					</div>
					<button
						onclick={() => (editing = true)}
						class="absolute -bottom-2 -right-2 bg-primary text-white p-2 rounded-lg shadow-lg hover:bg-primary-container transition-all"
					>
						<span class="material-symbols-outlined text-sm">edit</span>
					</button>
				</div>
				<div class="text-center md:text-left flex-1">
					<h2 class="font-headline text-3xl font-extrabold tracking-tight text-primary mb-1">
						{profile.name}
					</h2>
					<div class="flex flex-wrap justify-center md:justify-start gap-3 items-center">
						<span
							class="font-label text-xs font-bold uppercase tracking-widest text-on-surface-variant bg-surface-container px-3 py-1 rounded-full"
							>ID: {profile.id || '---'}</span
						>

						{#if profile.is_employee}
							<span
								class="flex items-center gap-1 text-tertiary-container font-semibold text-xs bg-tertiary-fixed-dim/20 px-3 py-1 rounded-full"
							>
								<span
									class="material-symbols-outlined text-[14px]"
									style="font-variation-settings: 'FILL' 1;">verified</span
								>
								{profile.is_active ? $_('common.active') : $_('common.inactive')}
							</span>
						{:else}
							<span
								class="flex items-center gap-1 text-amber-600 font-semibold text-xs bg-amber-50 px-3 py-1 rounded-full border border-amber-100 italic"
							>
								<span class="material-symbols-outlined text-[14px]">info</span>
								{$_('common.pending')}
							</span>
						{/if}
					</div>
				</div>
			</section>

			<!-- Section 1 - Personal Information -->
			<section class="mb-8" in:fly={{ y: 20, delay: 100 }}>
				<div class="flex items-baseline justify-between mb-4 px-2">
					<h3 class="font-headline text-lg font-bold text-primary">
						{$_('profile.personal_info')}
					</h3>
					<span class="h-[1px] flex-1 mx-4 bg-outline-variant/20"></span>
				</div>
				<div class="bg-surface-container-low rounded-sm p-1 space-y-1">
					<div
						class="bg-surface-container-lowest p-5 rounded-lg flex justify-between items-center group transition-all hover:bg-white shadow-sm"
					>
						<div class="space-y-1">
							<p class="font-label text-[10px] uppercase tracking-widest text-outline font-bold">
								{$_('auth.name_label')}
							</p>
							<p class="font-body text-on-surface font-medium">{profile.name}</p>
						</div>
						<button
							onclick={() => (editing = true)}
							class="material-symbols-outlined text-outline opacity-0 group-hover:opacity-100 transition-opacity cursor-pointer bg-transparent border-none"
							>edit</button
						>
					</div>
					<div
						class="bg-surface-container-lowest p-5 rounded-lg flex justify-between items-center group transition-all hover:bg-white shadow-sm"
					>
						<div class="space-y-1">
							<p class="font-label text-[10px] uppercase tracking-widest text-outline font-bold">
								{$_('common.email')}
							</p>
							<p class="font-body text-on-surface font-medium">{profile.email}</p>
						</div>
						<button
							onclick={() => (editing = true)}
							class="material-symbols-outlined text-outline opacity-0 group-hover:opacity-100 transition-opacity cursor-pointer bg-transparent border-none"
							>edit</button
						>
					</div>
					<div
						class="bg-surface-container-lowest p-5 rounded-lg flex justify-between items-center group transition-all hover:bg-white shadow-sm"
					>
						<div class="space-y-1">
							<p class="font-label text-[10px] uppercase tracking-widest text-outline font-bold">
								{$_('auth.phone_label')}
							</p>
							{#if profile.phone}
								<p class="font-body text-on-surface font-medium">{profile.phone}</p>
							{:else}
								<p class="font-body text-on-surface font-medium italic opacity-50">
									{$_('common.to_be_defined')}
								</p>
							{/if}
						</div>
						<button
							onclick={() => (editing = true)}
							class="material-symbols-outlined text-outline opacity-0 group-hover:opacity-100 transition-opacity cursor-pointer bg-transparent border-none"
							>edit</button
						>
					</div>
				</div>
			</section>

			<!-- Section 2 - Professional Information -->
			<section class="mb-12" in:fly={{ y: 20, delay: 200 }}>
				<div class="flex items-baseline justify-between mb-4 px-2">
					<h3 class="font-headline text-lg font-bold text-primary">
						{$_('profile.professional_info')}
					</h3>
					<span class="h-[1px] flex-1 mx-4 bg-outline-variant/20"></span>
				</div>
				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<div class="bg-surface-container-low p-6 rounded-sm space-y-4">
						<div class="space-y-1">
							<p class="font-label text-[10px] uppercase tracking-widest text-outline font-bold">
								{$_('profile.position')}
							</p>
							<p class="font-headline text-xl font-bold text-primary flex items-center gap-2">
								{profile.position_name || $_('common.unassigned')}
								{#if profile.hourly_rate}
									<span
										class="bg-emerald-50 text-emerald-600 px-2 py-0.5 rounded-sm text-[10px] font-black tracking-wider uppercase"
									>
										${profile.hourly_rate}/hr
									</span>
								{/if}
							</p>
						</div>
						<div class="space-y-1">
							<p class="font-label text-[10px] uppercase tracking-widest text-outline font-bold">
								{$_('profile.work_center')}
							</p>
							<p class="font-body text-on-surface">
								{profile.work_center_name || $_('common.unassigned')}
							</p>
						</div>
					</div>
					<div
						class="bg-primary text-white p-6 rounded-sm flex flex-col justify-between shadow-[0px_12px_32px_rgba(0,44,96,0.15)]"
					>
						<div class="space-y-1">
							<p
								class="font-label text-[10px] uppercase tracking-widest text-primary-fixed-dim/80 font-bold"
							>
								{$_('profile.shift_schedule')}
							</p>
							<p class="font-headline text-2xl font-black">
								{formatShiftTime(profile.expected_check_in)} — {formatShiftTime(
									profile.expected_check_out
								)}
							</p>
						</div>
						<div class="flex items-center gap-2 mt-4 text-xs font-medium text-primary-fixed">
							<span class="material-symbols-outlined text-sm">schedule</span>
							{$_('schedule.weekdays')}
						</div>
					</div>
				</div>
			</section>

			<!-- Section 3 - Account Settings -->
			<section class="space-y-4" in:fly={{ y: 20, delay: 300 }}>
				<button
					onclick={() => alert('Próximamente')}
					class="w-full bg-surface-container-lowest border border-outline-variant/20 py-4 px-6 rounded-sm flex items-center justify-between group hover:bg-surface-container transition-colors"
				>
					<div class="flex items-center gap-4">
						<div
							class="w-10 h-10 rounded-full bg-secondary-container flex items-center justify-center text-primary"
						>
							<span class="material-symbols-outlined">lock_reset</span>
						</div>
						<span class="font-headline font-bold text-primary">{$_('profile.change_password')}</span
						>
					</div>
					<span
						class="material-symbols-outlined text-outline group-hover:translate-x-1 transition-transform"
						>chevron_right</span
					>
				</button>
				<button
					onclick={logout}
					class="w-full bg-error-container/10 border border-error/10 py-4 px-6 rounded-sm flex items-center justify-between group hover:bg-error-container transition-colors"
				>
					<div class="flex items-center gap-4">
						<div
							class="w-10 h-10 rounded-full bg-error-container flex items-center justify-center text-on-error-container"
						>
							<span class="material-symbols-outlined">logout</span>
						</div>
						<span class="font-headline font-bold text-on-error-container"
							>{$_('common.logout')}</span
						>
					</div>
					<span class="material-symbols-outlined text-on-error-container/40">arrow_forward</span>
				</button>
			</section>
		{/if}

		{#if editing}
			<div
				class="fixed inset-0 z-50 flex items-center justify-center p-6 bg-slate-900/60"
				transition:fade
			>
				<div
					class="w-full max-w-md bg-white rounded-2xl shadow-2xl overflow-hidden"
					in:fly={{ y: 20 }}
				>
					<div class="p-8 border-b border-slate-100">
						<h3 class="font-headline text-2xl font-black text-slate-900 tracking-tight">
							{$_('profile.edit_title') || 'Editar Perfil'}
						</h3>
						<p class="text-slate-500 text-sm font-medium mt-1">
							{$_('profile.edit_description') || 'Actualiza tu información de contacto.'}
						</p>
					</div>

					<div class="p-8 space-y-6">
						<div class="space-y-2">
							<label
								for="name"
								class="text-[10px] font-black uppercase tracking-widest text-slate-400"
								>{$_('auth.name_label')}</label
							>
							<input
								id="name"
								type="text"
								bind:value={editForm.name}
								placeholder="Tu nombre completo"
								class="w-full h-14 px-5 rounded-xl bg-slate-50 border-2 border-transparent focus:border-primary focus:bg-white transition-all font-bold text-slate-900 outline-none"
							/>
						</div>

						<div class="space-y-2">
							<label
								for="email"
								class="text-[10px] font-black uppercase tracking-widest text-slate-400"
								>{$_('common.email')}</label
							>
							<input
								id="email"
								type="email"
								bind:value={editForm.email}
								placeholder="correo@ejemplo.com"
								class="w-full h-14 px-5 rounded-xl bg-slate-50 border-2 border-transparent focus:border-primary focus:bg-white transition-all font-bold text-slate-900 outline-none"
							/>
						</div>

						<div class="space-y-2">
							<label
								for="phone"
								class="text-[10px] font-black uppercase tracking-widest text-slate-400"
								>{$_('auth.phone_label')}</label
							>
							<input
								id="phone"
								type="tel"
								bind:value={editForm.phone}
								placeholder="+1 (555) 000-0000"
								class="w-full h-14 px-5 rounded-xl bg-slate-50 border-2 border-transparent focus:border-primary focus:bg-white transition-all font-bold text-slate-900 outline-none"
							/>
						</div>
					</div>

					<div class="p-8 bg-slate-50/50 flex gap-4">
						<button
							onclick={() => (editing = false)}
							class="flex-1 h-14 rounded-xl font-black text-slate-400 uppercase tracking-widest hover:bg-slate-100 transition-all"
						>
							{$_('common.cancel')}
						</button>
						<button
							onclick={saveProfile}
							disabled={saving}
							class="flex-1 h-14 rounded-xl bg-primary text-white font-black uppercase tracking-widest shadow-lg shadow-primary/20 hover:scale-[1.02] active:scale-95 transition-all flex items-center justify-center gap-2"
						>
							{#if saving}
								<div
									class="w-5 h-5 border-2 border-white/20 border-t-white rounded-full animate-spin"
								></div>
							{:else}
								{$_('common.save') || 'Guardar'}
							{/if}
						</button>
					</div>
				</div>
			</div>
		{/if}
	</main>
</div>

<style>
	:global(body) {
		background-color: #f8fafb;
	}
	.material-symbols-outlined {
		font-variation-settings:
			'FILL' 0,
			'wght' 400,
			'GRAD' 0,
			'opsz' 24;
	}
</style>
