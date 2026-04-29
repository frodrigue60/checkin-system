<script lang="ts">
    import { fade, fly } from 'svelte/transition';
    import { _ } from 'svelte-i18n';
</script>

<svelte:head>
    <title>{$_('schedule.title')} | {$_('landing.system_name')}</title>
</svelte:head>

<div class="min-h-screen bg-surface flex flex-col items-center justify-center p-6 text-center">
    <div in:fly={{ y: 20, duration: 800 }} class="space-y-8 max-w-sm">
        <!-- Visual Icon -->
        <div class="relative inline-flex">
            <div class="w-24 h-24 bg-primary/5 rounded-[2.5rem] flex items-center justify-center rotate-12 absolute -inset-1 -z-10 animate-pulse"></div>
            <div class="w-24 h-24 bg-white shadow-2xl shadow-primary/10 rounded-[2rem] flex items-center justify-center text-primary relative">
                <span class="material-symbols-outlined text-5xl" style="font-variation-settings: 'FILL' 1;">event_note</span>
            </div>
        </div>

        <!-- Text Content -->
        <div class="space-y-3">
            <h1 class="text-3xl font-black text-primary font-headline tracking-tighter">{$_('schedule.header')}</h1>
            <p class="text-sm font-medium text-outline-variant leading-relaxed">
                {$_('schedule.description')}
            </p>
        </div>

        <!-- Status Badge -->
        <div class="inline-flex items-center gap-2 px-4 py-2 bg-surface-container-low rounded-full border border-outline-variant/10">
            <div class="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></div>
            <span class="text-[10px] font-black uppercase tracking-widest text-primary">{$_('common.coming_soon')}</span>
        </div>

        <!-- Back Home -->
        <div class="pt-8">
            <a href="/" class="text-[10px] font-black uppercase tracking-widest text-outline-variant hover:text-primary transition-colors no-underline flex items-center justify-center gap-2">
                <span class="material-symbols-outlined text-lg">arrow_back</span>
                {$_('common.back_to_dashboard')}
            </a>
        </div>
    </div>
</div>
