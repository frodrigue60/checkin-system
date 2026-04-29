<script lang="ts">
  import { authState } from '$lib/auth.svelte';
  import { _ } from 'svelte-i18n';
  import { page } from '$app/state';
  import { goto } from '$app/navigation';
  import { untrack } from 'svelte';
  import { fly, fade } from 'svelte/transition';
  import { Button } from "$lib/components/ui/button";
  import * as DropdownMenu from "$lib/components/ui/dropdown-menu";
  import { Avatar, AvatarFallback, AvatarImage } from "$lib/components/ui/avatar";
  import { 
    LayoutDashboard, 
    Settings, 
    Users, 
    BarChart3, 
    LogOut, 
    ChevronDown, 
    Menu, 
    X, 
    Globe,
    Building2,
    Clock,
    UserCircle2,
    Briefcase
  } from "lucide-svelte";
  import { PUBLIC_APP_NAME } from '$env/static/public';
  import Badge from "$lib/components/ui/badge/badge.svelte";

  let isMenuOpen = $state(false);
  let activeDropdown = $state<string | null>(null);

  const isActive = (path: string) => page.url.pathname === path;
  const isParentActive = (path: string) => page.url.pathname.startsWith(path);

  function logout() {
    isMenuOpen = false;
    activeDropdown = null;
    authState.logout();
    goto('/login');
  }

  function toggleMenu() {
    isMenuOpen = !isMenuOpen;
  }

  function toggleDropdown(name: string) {
    if (activeDropdown === name) {
      activeDropdown = null;
    } else {
      activeDropdown = name;
    }
  }

  // Close everything on navigation
  $effect(() => {
    page.url.pathname;
    untrack(() => {
      isMenuOpen = false;
      activeDropdown = null;
    });
  });

  // Groups for Admin/Manager
  const managementLinks = $derived([
    { label: $_('nav.centers'), href: '/admin/centers', active: isParentActive('/admin/centers') },
    { label: $_('nav.shifts'), href: '/admin/shifts', active: isParentActive('/admin/shifts') },
    { label: $_('nav.positions'), href: '/admin/positions', active: isParentActive('/admin/positions') },
  ]);

  const hrLinks = $derived([
    { label: $_('nav.directory'), href: '/admin/employees', active: isParentActive('/admin/employees') },
    { label: $_('nav.attendance'), href: '/admin/attendance', active: isParentActive('/admin/attendance') },
  ]);

  const analysisLinks = $derived([
    { label: $_('nav.reports'), href: '/admin/reports', active: isParentActive('/admin/reports') },
  ]);

  // Flat list for Mobile
  const mobileLinks = $derived([
    ...(authState.isAdmin || authState.isManager || authState.isSupervisor ? [
      { label: $_('nav.dashboard'), href: '/dashboard', active: isActive('/dashboard'), icon: LayoutDashboard },
      { label: $_('nav.centers'), href: '/admin/centers', active: isParentActive('/admin/centers'), icon: Building2 },
      { label: $_('nav.shifts'), href: '/admin/shifts', active: isParentActive('/admin/shifts'), icon: Clock },
      { label: $_('nav.positions'), href: '/admin/positions', active: isParentActive('/admin/positions'), icon: Briefcase },
      { label: $_('nav.directory'), href: '/admin/employees', active: isParentActive('/admin/employees'), icon: Users },
      { label: $_('nav.attendance'), href: '/admin/attendance', active: isParentActive('/admin/attendance'), icon: Globe },
      { label: $_('nav.reports'), href: '/admin/reports', active: isParentActive('/admin/reports'), icon: BarChart3 }
    ] : []),
    ...(authState.isEmployee ? [{ label: $_('nav.my_checkin'), href: '/employee-dashboard', active: isActive('/employee-dashboard'), icon: Globe }] : [])
  ]);
</script>

<nav class="fixed top-0 left-0 w-full h-[72px] z-50 bg-white border-b border-slate-200 selection:bg-primary/20">
  <div class="max-w-7xl mx-auto h-full px-6 flex items-center justify-between gap-8">
    <!-- Brand -->
    <a href="/" class="flex items-center gap-3 no-underline group">
      <div class="h-10 w-10 bg-primary text-white rounded-md flex items-center justify-center shadow-lg shadow-primary/20 group-hover:scale-105 transition-transform">
        <span class="font-black text-xs tracking-tighter">{PUBLIC_APP_NAME.split(' ').map(n => n[0]).join('').slice(0, 3).toUpperCase()}</span>
      </div>
      <span class="text-lg font-black tracking-tight text-slate-900 group-hover:text-primary transition-colors">{PUBLIC_APP_NAME}</span>
    </a>

    <!-- Desktop Navigation -->
    <div class="hidden md:flex items-center gap-1">
      {#if authState.isAdmin || authState.isManager || authState.isSupervisor}
        <Button variant={isActive('/dashboard') ? 'secondary' : 'ghost'} href="/dashboard" class="rounded-md font-bold gap-2 {isActive('/dashboard') ? 'bg-primary/5 text-primary hover:bg-primary/10' : 'text-slate-600'}">
          {$_('nav.dashboard')}
        </Button>
        
        <!-- Gestión Dropdown -->
        <DropdownMenu.Root>
          <DropdownMenu.Trigger>
            {#snippet child({ props: builder })}
              <Button builders={[builder]} variant={managementLinks.some(l => l.active) ? 'secondary' : 'ghost'} class="rounded-md font-bold gap-2 {managementLinks.some(l => l.active) ? 'bg-primary/5 text-primary hover:bg-primary/10' : 'text-slate-600'}">
                {$_('nav.management')}
                <ChevronDown class="h-4 w-4 opacity-50" />
              </Button>
            {/snippet}
          </DropdownMenu.Trigger>
          <DropdownMenu.Content class="rounded-md p-2 border-none shadow-premium min-w-[200px]" transition={fly} transitionConfig={{ y: 8, duration: 200 }}>
            {#each managementLinks as link}
              <DropdownMenu.Item asChild>
                <a href={link.href} class="flex items-center gap-2 p-3 rounded-md font-bold text-sm {link.active ? 'bg-primary/5 text-primary' : 'text-slate-600 hover:bg-slate-50'} no-underline transition-colors">
                  {link.label}
                </a>
              </DropdownMenu.Item>
            {/each}
          </DropdownMenu.Content>
        </DropdownMenu.Root>

        <!-- Personnel Dropdown -->
        <DropdownMenu.Root>
          <DropdownMenu.Trigger>
            {#snippet child({ props: builder })}
              <Button builders={[builder]} variant={hrLinks.some(l => l.active) ? 'secondary' : 'ghost'} class="rounded-md font-bold gap-2 {hrLinks.some(l => l.active) ? 'bg-primary/5 text-primary hover:bg-primary/10' : 'text-slate-600'}">
                {$_('nav.personnel')}
                <ChevronDown class="h-4 w-4 opacity-50" />
              </Button>
            {/snippet}
          </DropdownMenu.Trigger>
          <DropdownMenu.Content class="rounded-md p-2 border-none shadow-premium min-w-[200px]" transition={fly} transitionConfig={{ y: 8, duration: 200 }}>
            {#each hrLinks as link}
              <DropdownMenu.Item asChild>
                <a href={link.href} class="flex items-center gap-2 p-3 rounded-md font-bold text-sm {link.active ? 'bg-primary/5 text-primary' : 'text-slate-600 hover:bg-slate-50'} no-underline transition-colors">
                  {link.label}
                </a>
              </DropdownMenu.Item>
            {/each}
          </DropdownMenu.Content>
        </DropdownMenu.Root>

        <Button variant={isParentActive('/admin/reports') ? 'secondary' : 'ghost'} href="/admin/reports" class="rounded-md font-bold gap-2 {isParentActive('/admin/reports') ? 'bg-primary/5 text-primary hover:bg-primary/10' : 'text-slate-600'}">
          {$_('nav.reports')}
        </Button>
      {:else if authState.isEmployee}
        <Button variant={isActive('/employee-dashboard') ? 'secondary' : 'ghost'} href="/employee-dashboard" class="rounded-md font-bold gap-2 {isActive('/employee-dashboard') ? 'bg-primary/5 text-primary hover:bg-primary/10' : 'text-slate-600'}">
          {$_('nav.my_checkin')}
        </Button>
      {/if}
    </div>

    <!-- Right Section -->
    <div class="flex items-center gap-4">
      {#if authState.user}
        <DropdownMenu.Root>
          <DropdownMenu.Trigger>
            {#snippet child({ props: builder })}
              <Button builders={[builder]} variant="ghost" class="hidden md:flex items-center gap-3 h-12 px-2 hover:bg-slate-50 rounded-md transition-all group font-sans">
                <div class="text-right flex flex-col items-end">
                  <span class="text-sm font-black text-slate-900 leading-tight">{authState.user?.profile?.name || authState.user?.name || $_('common.user')}</span>
                  <span class="text-[10px] font-black uppercase text-primary tracking-widest">{authState.user?.role_slug || $_('common.role')}</span>
                </div>
                <Avatar class="h-9 w-9 border-2 border-white shadow-sm ring-2 ring-primary/20">
                  <AvatarFallback class="bg-primary text-white font-black text-xs">
                    { (authState.user?.profile?.name || authState.user?.name || 'U')[0].toUpperCase() }
                  </AvatarFallback>
                </Avatar>
              </Button>
            {/snippet}
          </DropdownMenu.Trigger>
          <DropdownMenu.Content class="rounded-md p-2 border-none shadow-premium min-w-[240px]" align="end" transition={fly} transitionConfig={{ y: 8, duration: 200 }}>
            <div class="px-4 py-3 pb-4">
              <p class="text-sm font-black text-slate-900">{authState.user?.profile?.name}</p>
              <p class="text-xs font-medium text-muted-foreground truncate">{authState.user?.profile?.email || authState.user?.email}</p>
            </div>
            <DropdownMenu.Separator class="bg-slate-100 mx-2" />
            <DropdownMenu.Item onSelect={logout} class="flex items-center gap-3 p-3 rounded-md font-bold text-sm text-rose-500 hover:bg-rose-50 hover:text-rose-600 transition-colors cursor-pointer font-sans">
              <LogOut class="h-4 w-4" />
              {$_('common.logout')}
            </DropdownMenu.Item>
          </DropdownMenu.Content>
        </DropdownMenu.Root>
      {/if}

      <!-- Menu Toggle (Mobile) -->
      <Button variant="ghost" size="icon" class="md:hidden rounded-md h-10 w-10 active:scale-95 transition-all" onclick={toggleMenu}>
        {#if isMenuOpen}
          <X class="h-6 w-6 text-slate-600" />
        {:else}
          <Menu class="h-6 w-6 text-slate-600" />
        {/if}
      </Button>
    </div>
  </div>

  <!-- Mobile Drawer -->
  {#if isMenuOpen}
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div class="fixed inset-0 bg-slate-900/40 z-40 md:hidden" onclick={() => isMenuOpen = false} transition:fade></div>
    
    <div class="fixed top-0 right-0 h-full w-[300px] bg-white z-50 md:hidden shadow-2xl flex flex-col animate-in slide-in-from-right duration-300">
      <div class="p-8 pb-6 space-y-6">
        <div class="flex justify-between items-center">
          <div class="h-14 w-14 bg-primary text-white rounded-md flex items-center justify-center shadow-lg shadow-primary/20 rotate-3 font-black text-lg tracking-tighter">{PUBLIC_APP_NAME.split(' ').map(n => n[0]).join('').slice(0, 3).toUpperCase()}</div>
          <Button variant="ghost" size="icon" class="rounded-md h-10 w-10 font-sans" onclick={() => isMenuOpen = false}>
            <X class="h-6 w-6 text-slate-400" />
          </Button>
        </div>

        {#if authState.user}
          <div class="space-y-1">
            <h3 class="text-xl font-black text-slate-900">{authState.user?.profile?.name || authState.user?.name}</h3>
            <Badge variant="outline" class="bg-primary/5 text-primary border-primary/20 font-bold px-2 py-0">
              {authState.user?.role_slug}
            </Badge>
          </div>
        {/if}
      </div>

      <div class="flex-1 overflow-y-auto px-4 space-y-2">
        {#each mobileLinks as link}
          {@const Icon = link.icon}
          <a href={link.href} class="flex items-center gap-4 p-4 rounded-md font-black text-sm no-underline transition-all {link.active ? 'bg-primary text-white shadow-lg shadow-primary/20' : 'text-slate-600 hover:bg-slate-50'}">
            <Icon class="h-5 w-5" />
            {link.label}
          </a>
        {/each}
      </div>

      <div class="p-6 border-t border-slate-100">
        <Button variant="ghost" class="w-full h-12 rounded-md font-black text-rose-500 hover:bg-rose-50 hover:text-rose-600 gap-3 justify-start font-sans" onclick={logout}>
          <LogOut class="h-5 w-5" />
          {$_('common.logout')}
        </Button>
      </div>
    </div>
  {/if}
</nav>
