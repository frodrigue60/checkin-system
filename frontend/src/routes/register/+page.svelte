<script lang="ts">
	import { goto } from '$app/navigation';
	import { apiFetch } from '$lib/api';
	import { authState } from '$lib/auth.svelte';
	import { Loader2 } from 'lucide-svelte';
	import { _ } from 'svelte-i18n';
	import { translateError } from '$lib/i18n/error-translator';
	import LanguageSelector from '$lib/components/LanguageSelector.svelte';

	let name = $state('');
	let email = $state('');
	let phone = $state(''); // New field from mockup, currently not saved to backend
	let password = $state('');
	let loading = $state(false);
	let errorMsg = $state('');
	let successMsg = $state('');

	async function handleRegister(e: Event) {
		e.preventDefault();
		loading = true;
		errorMsg = '';
		successMsg = '';

		try {
			// Backend currently only expects name, email, password
			const res = await apiFetch('/auth/register', {
				method: 'POST',
				body: JSON.stringify({ name, email, phone, password })
			});

			if (res.ok) {
				const data = await res.json();
				authState.login(data.token, data.user);
				successMsg = $_('auth.register_success');
				setTimeout(() => goto('/dashboard'), 1500);
			} else {
				const errData = await res.json();
				errorMsg = translateError(errData.error || 'Registration failed');
			}
		} catch (err) {
			errorMsg = $_('auth.error_server');
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>{$_('auth.register_title')} | {$_('landing.system_name')}</title>
</svelte:head>

<div
	class="bg-background font-body text-on-surface antialiased overflow-x-hidden selection:bg-primary/20 min-h-[100dvh]"
>
	<!-- Main Content Canvas -->
	<main class="min-h-screen flex flex-col md:flex-row relative overflow-hidden">
		<!-- Left Side: Editorial Brand Pillar (Desktop Only) -->
		<section class="hidden md:flex md:w-5/12 bg-primary p-16 flex-col justify-between relative">
			<!-- Decorative Background Element -->
			<div class="absolute inset-0 opacity-10 pointer-events-none overflow-hidden">
				<img
					class="w-full h-full object-cover"
					alt="Modern architectural glass building facade"
					src="./images/register-bg.png"
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
					{$_('auth.hero_title_reg')}
				</h1>
				<p class="text-on-primary/70 font-light text-lg max-w-sm leading-relaxed">
					{$_('auth.hero_subtitle_reg')}
				</p>
			</div>
			<div class="relative z-10">
				<div class="h-1 w-24 bg-primary-fixed-dim/30 mb-6"></div>
				<div class="text-on-primary/50 font-label text-xs uppercase tracking-widest">
					{$_('landing.system_name')} • v2.4.0
				</div>
			</div>
		</section>

		<!-- Right Side: Registration Portal -->
		<section
			class="flex-1 flex flex-col justify-center items-center px-6 py-12 md:px-12 lg:px-24 bg-surface relative"
		>
			<!-- Language Selector (Guest) -->
			<div class="absolute top-4 right-4 md:top-8 md:right-8">
				<LanguageSelector class="w-28 md:w-32" />
			</div>
			<!-- Mobile Brand Header -->
			<div class="md:hidden w-full max-w-md flex items-center gap-3 mb-12">
				<span class="material-symbols-outlined text-primary text-2xl">corporate_fare</span>
				<span class="font-display font-black text-primary tracking-tight text-xl"
					>{$_('landing.system_name')}</span
				>
			</div>

			<div class="w-full max-w-md">
				<header class="mb-10">
					<h2 class="font-display text-3xl font-bold text-primary tracking-tight mb-2">
						{$_('auth.register_title')}
					</h2>
					<p class="text-on-surface-variant font-medium">{$_('auth.register_subtitle')}</p>
				</header>

				<form class="space-y-6" onsubmit={handleRegister}>
					<!-- Full Name -->
					<div class="group">
						<label
							class="block font-label text-[10px] font-black uppercase tracking-wider mb-2 text-outline group-focus-within:text-primary transition-colors"
							for="name"
						>
							{$_('auth.name_label')}
						</label>
						<div class="relative">
							<input
								class="w-full bg-surface-container-lowest border-0 border-b-2 border-outline-variant/30 py-3 px-0 focus:ring-0 focus:border-primary text-on-surface font-body transition-all placeholder:text-outline/40 font-bold"
								id="name"
								name="name"
								placeholder="Arthur P. Morgan"
								type="text"
								bind:value={name}
								required
							/>
						</div>
					</div>

					<!-- Work Email -->
					<div class="group">
						<label
							class="block font-label text-[10px] font-black uppercase tracking-wider mb-2 text-outline group-focus-within:text-primary transition-colors"
							for="email"
						>
							{$_('common.email')}
						</label>
						<div class="relative">
							<input
								class="w-full bg-surface-container-lowest border-0 border-b-2 border-outline-variant/30 py-3 px-0 focus:ring-0 focus:border-primary text-on-surface font-body transition-all placeholder:text-outline/40 font-bold"
								id="email"
								name="email"
								placeholder="a.morgan@thebureau.com"
								type="email"
								bind:value={email}
								required
							/>
						</div>
					</div>

					<div class="grid grid-cols-1 sm:grid-cols-2 gap-6">
						<!-- Phone (UI only for now) -->
						<div class="group">
							<label
								class="block font-label text-[10px] font-black uppercase tracking-wider mb-2 text-outline group-focus-within:text-primary transition-colors"
								for="phone"
							>
								{$_('auth.phone_label')}
							</label>
							<input
								class="w-full bg-surface-container-lowest border-0 border-b-2 border-outline-variant/30 py-3 px-0 focus:ring-0 focus:border-primary text-on-surface font-body transition-all placeholder:text-outline/40 uppercase font-bold"
								id="phone"
								name="phone"
								placeholder="051 123 456"
								type="text"
								bind:value={phone}
							/>
						</div>
						<!-- Password -->
						<div class="group">
							<label
								class="block font-label text-[10px] font-black uppercase tracking-wider mb-2 text-outline group-focus-within:text-primary transition-colors"
								for="password"
							>
								{$_('auth.password_label')}
							</label>
							<input
								class="w-full bg-surface-container-lowest border-0 border-b-2 border-outline-variant/30 py-3 px-0 focus:ring-0 focus:border-primary text-on-surface font-body transition-all placeholder:text-outline/40 font-bold"
								id="password"
								name="password"
								placeholder="••••••••"
								type="password"
								bind:value={password}
								required
							/>
						</div>
					</div>

					{#if errorMsg}
						<div
							class="p-4 bg-error-container/20 text-destructive text-sm font-bold border border-destructive/20 rounded-md animate-in fade-in slide-in-from-top-1"
						>
							{errorMsg}
						</div>
					{/if}

					{#if successMsg}
						<div
							class="p-4 bg-emerald-50 text-emerald-600 text-sm font-bold border border-emerald-100 rounded-md animate-in fade-in slide-in-from-top-1"
						>
							{successMsg}
						</div>
					{/if}

					<div class="pt-6">
						<button
							class="w-full bg-primary text-primary-foreground font-display font-bold py-4 rounded-lg shadow-lg hover:shadow-xl active:scale-[0.98] transition-all tracking-widest uppercase text-sm flex items-center justify-center gap-3 disabled:opacity-50"
							type="submit"
							disabled={loading}
						>
							{#if loading}
								<Loader2 class="h-5 w-5 animate-spin" />
								{$_('auth.initializing')}
							{:else}
								{$_('auth.register_button')}
							{/if}
						</button>
					</div>
				</form>

				<footer class="mt-12 text-center">
					<p class="text-on-surface-variant font-medium text-sm">
						{$_('auth.has_account')}
						<a class="text-primary font-bold hover:underline ml-1" href="/login"
							>{$_('common.login')}</a
						>
					</p>
					<div class="mt-16 flex items-center justify-center gap-8 opacity-40">
						<!-- <img
							class="h-6 grayscale filter brightness-0"
							alt="Connectivity logo"
							src="https://lh3.googleusercontent.com/aida-public/AB6AXuB0I9uFOjgURibwwI25NlDU9jqhAb_jvcGARcvlZH5kF4k0j6_qboIVwxPPUaW7Wqo4Uqi1odbxxUWOEMBSnwiQGvzL5UYB7gBpX9R7K_KpSSy6IVtejhwx_Jlyth5Gxio_m4PF_4e5pxciAoYpMSSUzKdffDPpJT-EncvZ0BllFYjwvTeAG7bZ4LAhFx6Ayvq9MdZjKEBOqReiQgyPpslKq36t3CAcavUqK7mbvVPAyMkc5BZr_2Bvkqu1lPsvNrTMaPC67rsrKws"
						/> -->
						<span class="material-symbols-outlined text-primary text-5xl">phone</span>
						<!-- <img
							class="h-5 grayscale filter brightness-0"
							alt="Institutional strength logo"
							src="https://lh3.googleusercontent.com/aida-public/AB6AXuAvX5AOpdINq1MmPw8_7bNeeYBhGEBrm02cL7axkj2SdScvuonmPIQhdIDcMHTdB088BkIvO4bp4UvAU2ETr1Jejtx2f1zm2VLOWqk0Mxvq3n5KrZgHbpg35lDEO_5BSNC8EzNHfw5QS3MOx82Y90KPaAHNI-hlIiNl2c_ky0RmvpZ7yWaOWv5K-x8Hy6FtVNJ0ZjV0OjGj8sDP8ukb2Nwy0a97nxMhcz_kpcFBWSs08KfjwNOvRwgFQm0ZyYCrkBhtKF9NNVzmzEs"
						/> -->
						<span class="material-symbols-outlined text-primary text-5xl">corporate_fare</span>
					</div>
				</footer>
			</div>

			<!-- Asymmetric Anchor (Time Display) -->
			<div
				class="hidden lg:flex absolute top-12 right-0 bg-surface-container-highest px-8 py-4 items-end flex-col border-l-4 border-primary"
			>
				<span class="font-display text-4xl font-black text-primary tabular-nums">
					{new Date().toLocaleTimeString('en-US', {
						hour12: false,
						hour: '2-digit',
						minute: '2-digit'
					})}
				</span>
				<span
					class="font-label text-[10px] font-bold text-on-surface-variant uppercase tracking-[0.2em] -mt-1"
					>Attendance System Standard Time</span
				>
			</div>
		</section>

		<!-- Decorative Floating Gradient -->
		<div
			class="absolute -bottom-64 -left-64 w-[500px] h-[500px] bg-primary/5 rounded-full blur-[120px] pointer-events-none"
		></div>
	</main>
</div>

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
</style>
