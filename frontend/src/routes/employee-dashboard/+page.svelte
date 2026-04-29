<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';

	onMount(() => {
		goto('/');
	});
</script>

<div class="flex items-center justify-center min-h-[50vh]">
	<p class="text-slate-400 font-bold animate-pulse">Redirecting to Dashboard...</p>
</div>
