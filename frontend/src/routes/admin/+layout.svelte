<script lang="ts">
	import { onMount } from 'svelte';
	import { authState } from '$lib/auth.svelte';
	import { goto } from '$app/navigation';

	let { children } = $props();

	onMount(() => {
		if (!authState.isAuthenticated) {
			goto('/login');
			return;
		}

		// Basic check: Employees should not be in /admin/*
		if (authState.isEmployee) {
			goto('/dashboard');
		}
	});
</script>

{#if !authState.isEmployee && authState.isAuthenticated}
	{@render children()}
{/if}
