<script lang="ts">
	import { notifications } from '$lib/notifications.svelte';
	import { flip } from 'svelte/animate';
	import { fly, fade } from 'svelte/transition';

	// Svelte 5 pattern for accessing the store items
	const items = $derived(notifications.items);

	function getColors(type: string) {
		switch (type) {
			case 'success':
				return {
					bg: 'bg-emerald-50/90',
					border: 'border-emerald-200',
					icon: 'text-emerald-600',
					accent: 'bg-emerald-500',
					symbol: 'check_circle'
				};
			case 'error':
				return {
					bg: 'bg-rose-50/90',
					border: 'border-rose-200',
					icon: 'text-rose-600',
					accent: 'bg-rose-500',
					symbol: 'error'
				};
			case 'warning':
				return {
					bg: 'bg-amber-50/90',
					border: 'border-amber-200',
					icon: 'text-amber-600',
					accent: 'bg-amber-500',
					symbol: 'warning'
				};
			default:
				return {
					bg: 'bg-slate-50/90',
					border: 'border-slate-200',
					icon: 'text-primary',
					accent: 'bg-primary',
					symbol: 'info'
				};
		}
	}
</script>

<div class="fixed top-6 right-6 z-[100] flex flex-col gap-3 w-full max-w-sm pointer-events-none">
	{#each items as notification (notification.id)}
		{@const colors = getColors(notification.type)}
		<div
			class="pointer-events-auto flex items-start gap-4 p-4 rounded-xl border shadow-2xl {colors.bg} {colors.border} overflow-hidden relative group"
			in:fly={{ x: 50, duration: 400 }}
			out:fade={{ duration: 200 }}
		>
			<!-- Accent Bar -->
			<div class="absolute left-0 top-0 bottom-0 w-1 {colors.accent}"></div>

			<!-- Icon -->
			<div class="mt-0.5">
				<span class="material-symbols-outlined {colors.icon} text-2xl font-bold">
					{colors.symbol}
				</span>
			</div>

			<!-- Message -->
			<div class="flex-1 pr-4">
				<p class="text-[11px] font-black uppercase tracking-[0.1em] {colors.icon} mb-0.5">
					{notification.type}
				</p>
				<p class="text-sm font-medium text-slate-700 leading-tight">
					{notification.message}
				</p>
			</div>

			<!-- Close Button -->
			<button
				onclick={() => notifications.dismiss(notification.id)}
				class="text-slate-400 hover:text-slate-600 transition-colors p-1"
			>
				<span class="material-symbols-outlined text-lg">close</span>
			</button>

			<!-- Progress Bar (Internal Auto-Dismiss Visualiczer) -->
			{#if notification.duration > 0}
				<div class="absolute bottom-0 left-0 right-0 h-0.5 bg-slate-200/20">
					<div
						class="h-full {colors.accent} opacity-40 shrink-animation"
						style="animation-duration: {notification.duration}ms"
					></div>
				</div>
			{/if}
		</div>
	{/each}
</div>

<style>
	/* Ensure Material Symbols are available */
	@import url('https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:opsz,wght,FILL,GRAD@20..48,100..700,0..1,-50..200');

	@keyframes shrink {
		from {
			width: 100%;
		}
		to {
			width: 0%;
		}
	}

	.shrink-animation {
		animation-name: shrink;
		animation-timing-function: linear;
		animation-fill-mode: forwards;
	}
</style>
