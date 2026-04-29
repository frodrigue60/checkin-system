<script lang="ts">
    import { page } from '$app/state';
    import { _ } from 'svelte-i18n';
    // Using Material Symbols directly as per mockup
    const tabs = [
        { label: 'nav.activity', icon: 'timer', href: '/' },
        { label: 'nav.history', icon: 'calendar_month', href: '/history' },
        { label: 'nav.schedule', icon: 'event_note', href: '/schedule' },
        { label: 'nav.profile', icon: 'person', href: '/profile' }
    ];

    const isActive = (href: string) => page.url.pathname === href;
</script>

<nav class="lg:hidden fixed bottom-0 left-0 w-full z-50 bg-white shadow-[0px_-12px_32px_rgba(25,28,29,0.06)] flex justify-around items-center px-4 py-4 pb-[calc(1rem+env(safe-area-inset-bottom))] border-t border-outline-variant/10">
    {#each tabs as tab}
        <a 
            href={tab.href} 
            class="flex flex-col items-center justify-center transition-all active:scale-90 no-underline {isActive(tab.href) ? 'text-primary' : 'text-outline-variant'}"
        >
            <span class="material-symbols-outlined transition-all" style="font-variation-settings: 'FILL' {isActive(tab.href) ? 1 : 0};">
                {tab.icon}
            </span>
            <span class="text-[10px] font-bold uppercase tracking-tighter mt-1">{$_(tab.label)}</span>
        </a>
    {/each}
</nav>

<style>
    .material-symbols-outlined {
        font-variation-settings: 'FILL' 0, 'wght' 400, 'GRAD' 0, 'opsz' 24;
    }
</style>
