<script>
	import { fade, fly } from 'svelte/transition';
	import { quintOut } from 'svelte/easing';
	import { _ } from 'svelte-i18n';
	let { isOpen = $bindable(false), onConfirm, loading = false } = $props();
	let reason = $state('');

	function handleConfirm() {
		if (!reason.trim()) return;
		onConfirm(reason);
		reason = '';
	}

	function close() {
		if (loading) return;
		isOpen = false;
		reason = '';
	}
</script>

{#if isOpen}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-[#002c60]/40 transition-all"
		in:fade={{ duration: 300 }}
		role="dialog"
		aria-modal="true"
	>
		<div
			class="w-full max-w-md bg-white rounded-xl shadow-2xl overflow-hidden"
			in:fly={{ y: 40, duration: 400, easing: quintOut, delay: 100 }}
		>
			<!-- Header -->
			<div class="px-6 py-5 border-b border-slate-100 flex items-center gap-4 bg-slate-50/50">
				<div class="w-10 h-10 rounded-full bg-red-100 flex items-center justify-center">
					<span class="material-symbols-outlined text-red-600">report_problem</span>
				</div>
				<div class="flex-1">
					<h3 class="text-xs font-black uppercase tracking-[0.2em] text-primary">{$_('absence.title')}</h3>
					<p class="text-[10px] font-bold text-slate-400 uppercase mt-0.5">
						{$_('common.today')}, {new Date().toLocaleDateString()}
					</p>
				</div>
				<button
					onclick={close}
					class="w-8 h-8 rounded-full flex items-center justify-center text-slate-400 hover:bg-slate-100 hover:text-slate-600 transition-all active:scale-95"
					disabled={loading}
				>
					<span class="material-symbols-outlined text-xl">close</span>
				</button>
			</div>

			<!-- Body -->
			<div class="p-6 space-y-4">
				<p class="text-sm text-slate-600 leading-relaxed font-medium">
					{$_('absence.description')}
					<span class="font-bold text-primary">{$_('dashboard.justified_absence')}</span>.
				</p>

				<div class="space-y-2">
					<label
						for="reason"
						class="block text-[10px] font-black uppercase tracking-widest text-slate-400"
					>
						{$_('absence.reason_label')}
					</label>
					<textarea
						id="reason"
						bind:value={reason}
						placeholder={$_('absence.placeholder')}
						class="w-full p-4 bg-slate-50 border border-slate-200 rounded-lg text-sm text-slate-700 min-h-[120px] focus:outline-none focus:ring-2 focus:ring-primary/10 focus:border-primary transition-all placeholder:text-slate-300"
						disabled={loading}
					></textarea>
				</div>
			</div>

			<!-- Footer -->
			<div class="px-6 py-4 bg-slate-50/50 border-t border-slate-100 flex gap-3 grid-cols-2">
				<button
					onclick={close}
					class="w-full py-3 text-[10px] font-black uppercase tracking-widest text-slate-400 hover:text-slate-600 transition-colors"
					disabled={loading}
				>
					{$_('common.cancel')}
				</button>
				<button
					onclick={handleConfirm}
					class="w-full py-3 bg-primary hover:bg-primary/80 text-white rounded-lg text-[10px] font-black uppercase tracking-widest shadow-lg shadow-red-500/20 active:scale-[0.98] transition-all disabled:opacity-50 disabled:grayscale"
					disabled={loading || !reason.trim()}
				>
					{loading ? $_('absence.submitting') : $_('absence.submit')}
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	/* Material symbols font fallback if not pre-loaded */
	@import url('https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:opsz,wght,FILL,GRAD@20..48,100..700,0..1,-50..200');
</style>
