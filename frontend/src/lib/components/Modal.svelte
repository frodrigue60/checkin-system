<script lang="ts">
	import * as Dialog from '$lib/components/ui/dialog/index.js';

	let {
		show = false,
		title = '',
		onclose,
		children
	} = $props<{
		show: boolean;
		title: string;
		onclose?: () => void;
		children?: any;
	}>();

	function handleOpenChange(open: boolean) {
		if (!open && onclose) {
			onclose();
		}
	}
</script>

<Dialog.Root open={show} onOpenChange={handleOpenChange}>
	<Dialog.Content class="max-w-5xl rounded-md border-none shadow-premium p-8">
		<Dialog.Header class="mb-6">
			<Dialog.Title class="text-2xl font-extrabold tracking-tight text-slate-900"
				>{title}</Dialog.Title
			>
		</Dialog.Header>

		<div class="max-h-[70vh] overflow-y-auto pr-2 custom-scrollbar">
			{#if children}
				{@render children()}
			{/if}
		</div>
	</Dialog.Content>
</Dialog.Root>

<style>
	.custom-scrollbar::-webkit-scrollbar {
		width: 6px;
	}
	.custom-scrollbar::-webkit-scrollbar-track {
		background: transparent;
	}
	.custom-scrollbar::-webkit-scrollbar-thumb {
		background-color: hsl(var(--border));
		border-radius: 10px;
	}
</style>
