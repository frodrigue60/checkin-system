<script lang="ts">
	import { goto } from '$app/navigation';
	import { apiFetch } from '$lib/api';
	import { authState } from '$lib/auth.svelte';
	import { Loader2 } from 'lucide-svelte';
	import { _ } from 'svelte-i18n';
	import { translateError } from '$lib/i18n/error-translator';
	import LanguageSelector from '$lib/components/LanguageSelector.svelte';

	let email = $state('');
	let password = $state('');
	let loading = $state(false);
	let errorMsg = $state('');

	async function handleLogin(e: Event) {
		e.preventDefault();
		loading = true;
		errorMsg = '';

		try {
			const res = await apiFetch('/auth/login', {
				method: 'POST',
				body: JSON.stringify({ email, password })
			});

			if (res.ok) {
				const data = await res.json();
				authState.login(data.token, data.user);
				goto('/dashboard');
			} else {
				const errData = await res.json();
				errorMsg = translateError(errData.error || 'Invalid credentials');
			}
		} catch (err) {
			errorMsg = $_('auth.error_server');
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>{$_('common.welcome')} | {$_('auth.login_title')}</title>
</svelte:head>

<div
	class="bg-background text-on-surface min-h-[100dvh] flex items-center justify-center p-0 md:p-6 lg:p-12 overflow-hidden selection:bg-primary/20"
>
	<!-- Login Container -->
	<main
		class="relative w-full max-w-[1440px] md:min-h-[795px] bg-surface-container flex overflow-hidden lg:rounded-sm shadow-[0px_12px_32px_rgba(25,28,29,0.06)]"
	>
		<!-- Left Side: Visual Anchor & Branding -->
		<section class="hidden lg:flex lg:w-3/5 relative bg-primary overflow-hidden">
			<img
				class="absolute inset-0 w-full h-full object-cover mix-blend-overlay opacity-40"
				alt="Modern corporate office interior"
				src="./images/login-bg.png"
			/>
			<div class="relative z-10 p-16 flex flex-col justify-between h-full w-full">
				<!-- Branding Area -->
				<div
					class="brand-logo text-on-primary font-black uppercase tracking-widest text-2xl font-display"
				>
					{$_('landing.system_name')}
				</div>
				<!-- Editorial Content -->
				<div class="max-w-md">
					<span
						class="text-secondary-fixed font-semibold tracking-widest uppercase text-xs mb-4 block"
						>{$_('common.welcome')}</span
					>
					<h1
						class="text-on-primary text-6xl font-black leading-tight tracking-tighter mb-6 font-display"
					>
						{$_('auth.hero_title')}
					</h1>
					<p class="text-primary-fixed-dim text-lg leading-relaxed opacity-80">
						{$_('auth.hero_subtitle')}
					</p>
				</div>
				<!-- Time-Stamp Hero (Asymmetric Anchor) -->
				<div
					class="absolute top-0 right-0 h-full w-24 bg-primary-container/20 flex items-end justify-center pb-12 border-l border-white/5"
				>
					<div class="rotate-90 origin-center translate-y-[-100%] whitespace-nowrap">
						<span
							class="text-on-primary-container font-display font-bold text-4xl tracking-tighter opacity-40"
						>
							{new Date().toLocaleTimeString('en-US', {
								hour12: true,
								hour: '2-digit',
								minute: '2-digit',
								second: '2-digit'
							})}
						</span>
					</div>
				</div>
			</div>
		</section>

		<!-- Right Side: Interaction Canvas -->
		<section
			class="w-full lg:w-2/5 flex flex-col bg-surface-container-lowest p-8 md:p-16 lg:p-20 relative"
		>
			<!-- Language Selector (Guest) -->
			<div class="absolute top-4 right-4 md:top-8 md:right-8">
				<LanguageSelector class="w-28 md:w-32" />
			</div>
			<!-- Mobile Branding -->
			<div class="lg:hidden mb-12">
				<div
					class="brand-logo text-primary font-black uppercase tracking-widest text-xl font-display"
				>
					{$_('landing.system_name')}
				</div>
			</div>

			<!-- Login Form -->
			<div class="w-full max-w-sm mx-auto my-auto">
				<header class="mb-10">
					<h2 class="text-primary text-3xl font-bold tracking-tight mb-2 font-display">
						{$_('auth.login_title')}
					</h2>
					<p class="text-on-surface-variant font-medium">{$_('auth.login_subtitle')}</p>
				</header>

				<form class="space-y-8" onsubmit={handleLogin}>
					<!-- Username Input -->
					<div class="relative group">
						<label
							class="block text-xs font-semibold uppercase tracking-widest text-on-surface-variant mb-2 ml-1"
							for="email"
						>
							{$_('auth.email_label')}
						</label>
						<div class="relative">
							<input
								class="w-full bg-surface-container-low border-0 border-b-2 border-outline/20 focus:border-primary focus:ring-0 px-4 py-4 text-on-surface transition-all duration-300 placeholder:text-outline-variant font-bold"
								id="email"
								name="email"
								placeholder="e.g. adrian.stone@bureau.com"
								type="email"
								bind:value={email}
								required
							/>
							<span
								class="material-symbols-outlined absolute right-4 top-1/2 -translate-y-1/2 text-outline-variant"
								>alternate_email</span
							>
						</div>
					</div>

					<!-- Password Input -->
					<div class="relative group">
						<div class="flex justify-between items-end mb-2 ml-1">
							<label
								class="text-xs font-semibold uppercase tracking-widest text-on-surface-variant"
								for="password"
							>
								{$_('auth.password_label')}
							</label>
							<a
								href="#"
								class="text-[11px] font-bold text-primary hover:text-primary/80 transition-colors uppercase tracking-wider"
							>
								{$_('auth.forgot_password')}
							</a>
						</div>
						<div class="relative">
							<input
								class="w-full bg-surface-container-low border-0 border-b-2 border-outline/20 focus:border-primary focus:ring-0 px-4 py-4 text-on-surface transition-all duration-300 placeholder:text-outline-variant font-bold"
								id="password"
								name="password"
								placeholder="••••••••••••"
								type="password"
								bind:value={password}
								required
							/>
							<span
								class="material-symbols-outlined absolute right-4 top-1/2 -translate-y-1/2 text-outline-variant"
								>lock</span
							>
						</div>
					</div>

					{#if errorMsg}
						<div
							class="p-4 bg-error-container/20 text-destructive text-sm font-bold border border-destructive/20 rounded-md animate-in fade-in slide-in-from-top-1"
						>
							{errorMsg}
						</div>
					{/if}

					<!-- Options -->
					<div class="flex items-center">
						<input
							class="w-5 h-5 rounded-md border-outline-variant text-primary focus:ring-primary transition-all cursor-pointer"
							id="remember"
							name="remember"
							type="checkbox"
						/>
						<label
							class="ml-3 text-sm font-medium text-on-surface-variant cursor-pointer select-none"
							for="remember"
						>
							{$_('auth.remember_me')}
						</label>
					</div>

					<!-- CTA -->
					<div class="pt-4">
						<button
							class="w-full bg-primary text-primary-foreground font-bold text-sm uppercase tracking-[0.2em] py-5 rounded-lg shadow-lg shadow-primary/10 hover:shadow-primary/20 hover:scale-[1.02] active:scale-95 transition-all duration-300 flex items-center justify-center gap-3 disabled:opacity-50"
							type="submit"
							disabled={loading}
						>
							{#if loading}
								<Loader2 class="h-5 w-5 animate-spin" />
								{$_('auth.authenticating')}
							{:else}
								{$_('auth.login_button')}
								<span class="material-symbols-outlined text-[20px]">login</span>
							{/if}
						</button>
					</div>
				</form>

				<!-- Footer -->
				<footer
					class="mt-16 pt-8 border-t border-surface-container flex flex-col items-center gap-4"
				>
					<p
						class="text-[10px] text-on-surface-variant/60 font-medium text-center leading-relaxed max-w-[280px]"
					>
						{$_('auth.footer_note')}
					</p>
					<div class="flex gap-6">
						<a
							href="/register"
							class="text-[10px] uppercase tracking-tighter font-bold text-primary hover:underline transition-colors"
							>{$_('common.register')}</a
						>
						<a
							href="#"
							class="text-[10px] uppercase tracking-tighter font-bold text-on-surface-variant/40 hover:text-primary transition-colors"
							>Security</a
						>
						<a
							href="#"
							class="text-[10px] uppercase tracking-tighter font-bold text-on-surface-variant/40 hover:text-primary transition-colors"
							>Support</a
						>
					</div>
				</footer>
			</div>
			<!-- Floating Decorative Element -->
			<div
				class="absolute bottom-12 right-0 w-32 h-1 bg-primary/10 rounded-l-full hidden lg:block"
			></div>
		</section>
	</main>

	<!-- Glass Overlay Background -->
	<div class="fixed top-0 left-0 w-full h-full -z-10 pointer-events-none overflow-hidden">
		<div
			class="absolute top-[-10%] right-[-10%] w-[50%] h-[50%] bg-primary/5 blur-[120px] rounded-full"
		></div>
		<div
			class="absolute bottom-[-10%] left-[-10%] w-[40%] h-[40%] bg-primary/10 blur-[120px] rounded-full"
		></div>
	</div>
</div>

<style>
	.brand-logo {
		font-family: 'Public Sans', sans-serif;
	}

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
