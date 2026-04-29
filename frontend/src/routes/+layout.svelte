<script lang="ts">
	import '../app.css';
	import '$lib/i18n';
	import Sidebar from '$lib/components/Sidebar.svelte';
	import MobileHeader from '$lib/components/MobileHeader.svelte';
	import MobileBottomNav from '$lib/components/MobileBottomNav.svelte';
	import Toaster from '$lib/components/Toaster.svelte';
	import { page } from '$app/state';
	import { authState } from '$lib/auth.svelte';
	import { PUBLIC_APP_NAME } from '$env/static/public';
	import { waitLocale, isLoading } from 'svelte-i18n';

	let { children } = $props();
	let sidebarOpen = $state(false);

	// Visibility logic
	const isAuthPage = $derived(
		page.url.pathname === '/login' || 
		page.url.pathname === '/register'
	);
	
	const showShell = $derived(!isAuthPage && authState.isAuthenticated);
</script>

<svelte:head>
	<title>{PUBLIC_APP_NAME} | Attendance</title>
	<meta name="description" content="Precision Attendance Logistics" />
</svelte:head>

<Toaster />

{#if $isLoading}
	<div class="min-h-screen flex items-center justify-center bg-background">
		<div class="animate-pulse flex flex-col items-center gap-4">
			<div class="h-12 w-12 bg-primary/20 rounded-lg"></div>
			<div class="h-4 w-32 bg-slate-200 rounded"></div>
		</div>
	</div>
{:else}
	<div class="min-h-screen bg-background flex flex-col lg:flex-row selection:bg-primary/10">
		{#if showShell}
			<MobileHeader onMenuClick={() => sidebarOpen = true} />
			
			<Sidebar bind:isMobileOpen={sidebarOpen} />
		{/if}
		
		<main class="flex-1 flex flex-col min-w-0 {showShell ? 'lg:pl-[280px] pt-16 lg:pt-0' : ''}">
			<div class="flex-1">
				{@render children()}
			</div>
		</main>

		{#if showShell}
			<MobileBottomNav />
		{/if}
	</div>
{/if}

<style>
	:global(body) {
		overflow-x: hidden;
		min-height: 100dvh;
	}
</style>
