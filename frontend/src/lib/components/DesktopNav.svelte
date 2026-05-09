<script lang="ts">
    import { page } from '$app/state';
    import { _ } from 'svelte-i18n';
    import { authState } from '$lib/auth.svelte';
    import * as Avatar from '$lib/components/ui/avatar';
    import { env } from '$env/dynamic/public';
    import { fly } from 'svelte/transition';

    const tabs = [
        { label: 'nav.activity', icon: 'timer', href: '/' },
        { label: 'nav.history', icon: 'calendar_month', href: '/history' },
        { label: 'nav.schedule', icon: 'event_note', href: '/schedule' },
        { label: 'nav.profile', icon: 'person', href: '/profile' }
    ];

    const isActive = (href: string) => page.url.pathname === href;
</script>

<nav class="hidden lg:flex fixed top-0 left-0 w-full z-50 bg-white/80 backdrop-blur-md border-b border-outline-variant/10 px-8 h-20 items-center justify-between shadow-sm">
    <!-- Brand / Logo -->
    <div class="flex items-center gap-3">
        <div class="h-10 w-10 bg-primary text-white rounded-xl flex items-center justify-center shadow-lg shadow-primary/20">
            <span class="font-black text-xs tracking-tighter">
                {env.PUBLIC_APP_NAME.split(' ').map(n => n[0]).join('').slice(0, 3).toUpperCase()}
            </span>
        </div>
        <div class="flex flex-col">
            <span class="text-sm font-black tracking-tight text-primary leading-none uppercase">
                {env.PUBLIC_APP_NAME.split(' ')[0]}
            </span>
            <span class="text-[9px] font-bold text-outline-variant uppercase tracking-[0.2em] mt-1">
                Personal Portal
            </span>
        </div>
    </div>

    <!-- Navigation Links -->
    <div class="flex items-center gap-2 bg-surface-container-low p-1.5 rounded-2xl border border-outline-variant/5">
        {#each tabs as tab}
            <a 
                href={tab.href} 
                class="flex items-center gap-2 px-6 py-2.5 rounded-xl transition-all no-underline group relative
                {isActive(tab.href) ? 'bg-white text-primary shadow-sm' : 'text-outline-variant hover:text-primary hover:bg-white/50'}"
            >
                <span class="material-symbols-outlined text-xl transition-all" style="font-variation-settings: 'FILL' {isActive(tab.href) ? 1 : 0};">
                    {tab.icon}
                </span>
                <span class="text-xs font-black uppercase tracking-widest">{$_(tab.label)}</span>
                
                {#if isActive(tab.href)}
                    <div 
                        class="absolute -bottom-[19px] left-1/2 -translate-x-1/2 w-8 h-1 bg-primary rounded-t-full"
                    ></div>
                {/if}
            </a>
        {/each}
    </div>

    <!-- User Action / Profile -->
    <div class="flex items-center gap-4">
        {#if authState.user}
            <div class="flex flex-col items-end mr-2">
                <span class="text-xs font-black text-primary leading-none">
                    {authState.user?.profile?.name || authState.user?.name}
                </span>
                <span class="text-[9px] font-bold text-outline-variant uppercase tracking-widest mt-1">
                    {authState.user?.role_slug}
                </span>
            </div>
            <a href="/profile" class="relative group">
                <Avatar.Root class="h-10 w-10 border-2 border-white ring-1 ring-outline-variant/10 shadow-sm transition-all group-hover:ring-primary/20">
                    <Avatar.Fallback class="bg-primary/5 text-primary font-black text-xs uppercase">
                        {(authState.user?.profile?.name || authState.user?.name || 'U')[0]}
                    </Avatar.Fallback>
                </Avatar.Root>
            </a>
        {/if}
    </div>
</nav>

<style>
    .material-symbols-outlined {
        font-variation-settings: 'FILL' 0, 'wght' 400, 'GRAD' 0, 'opsz' 24;
    }
</style>
