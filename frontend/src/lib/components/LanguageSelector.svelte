<script lang="ts">
	import { locale } from 'svelte-i18n';
	import { Globe, ChevronDown } from 'lucide-svelte';

	let { class: className = "" } = $props();

	// Persist language preference
	$effect(() => {
		if ($locale) {
			localStorage.setItem('locale', $locale);
		}
	});
</script>

<div class="relative flex items-center gap-2 px-3 py-2 bg-slate-50 rounded-md border border-slate-100 group hover:border-primary/20 transition-colors {className}">
	<Globe class="h-3.5 w-3.5 text-slate-400 group-hover:text-primary/50 transition-colors" />
	<select 
		bind:value={$locale} 
		class="bg-transparent text-[10px] font-black text-slate-500 focus:outline-none w-full cursor-pointer appearance-none uppercase tracking-[0.1em]"
	>
		<option value="es">Español</option>
		<option value="en">English</option>
	</select>
	<ChevronDown class="h-3 w-3 text-slate-400 pointer-events-none absolute right-3" />
</div>
