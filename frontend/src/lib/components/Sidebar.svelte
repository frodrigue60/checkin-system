<script lang="ts">
	import { authState } from '$lib/auth.svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { untrack } from 'svelte';
	import { fly, fade } from 'svelte/transition';
	import { Button } from '$lib/components/ui/button';
	import * as Avatar from '$lib/components/ui/avatar';
	import {
		LayoutDashboard,
		Users,
		BarChart3,
		LogOut,
		ChevronDown,
		Menu,
		X,
		Globe,
		Building2,
		Clock,
		Briefcase,
		CalendarDays,
		Settings,
		History,
		AlertTriangle
	} from 'lucide-svelte';
	import { PUBLIC_APP_NAME } from '$env/static/public';
	import { resolve } from '$app/paths';
	import { _, locale } from 'svelte-i18n';
	import LanguageSelector from '$lib/components/LanguageSelector.svelte';

	let { isMobileOpen = $bindable(false) } = $props();

	const isActive = (path: string) => page.url.pathname === path;
	const isParentActive = (path: string) => page.url.pathname.startsWith(path);

	function logout() {
		isMobileOpen = false;
		authState.logout();
		goto('/login');
	}

	// Close mobile sidebar on navigation
	$effect(() => {
		page.url.pathname;
		untrack(() => {
			isMobileOpen = false;
		});
	});

	// Structural Groups
	const adminGroups = [
		{
			label: 'sidebar.core_management',
			links: [
				{ label: 'sidebar.centers', href: '/admin/centers', icon: Building2 },
				{ label: 'sidebar.shifts', href: '/admin/shifts', icon: Clock },
				{ label: 'sidebar.positions', href: '/admin/positions', icon: Briefcase },
				{ label: 'sidebar.holidays', href: '/admin/holidays', icon: CalendarDays }
			]
		},
		{
			label: 'sidebar.personnel_activity',
			links: [
				{ label: 'sidebar.directory', href: '/admin/employees', icon: Users },
				{ label: 'sidebar.attendance', href: '/admin/attendance', icon: Globe },
				{ label: 'common.incidents', href: '/admin/incidents', icon: AlertTriangle }
			]
		},
		{
			label: 'sidebar.analytics',
			links: [
				{ label: 'common.reports', href: '/admin/reports', icon: BarChart3 },
				{ label: 'sidebar.audit', href: '/admin/audit', icon: History }
			]
		}
	];

	const employeeLinks = [{ label: 'sidebar.my_checkin', href: '/', icon: Globe }];

	// Persist language preference
	$effect(() => {
		if ($locale) {
			localStorage.setItem('locale', $locale);
		}
	});
</script>

{#if isMobileOpen}
	<div
		class="fixed inset-0 bg-slate-900/40 z-50 lg:hidden"
		onclick={() => (isMobileOpen = false)}
		onkeydown={(e) => e.key === 'Escape' && (isMobileOpen = false)}
		role="button"
		tabindex="-1"
		aria-label="Cerrar menú"
		transition:fade
	></div>
{/if}

<!-- Sidebar Container -->
<aside
	class="fixed top-0 left-0 h-full bg-white border-r border-slate-200 z-50 transition-transform duration-300 transform
  {isMobileOpen ? 'translate-x-0' : '-translate-x-full'} lg:translate-x-0 w-[280px] flex flex-col"
>
	<!-- Header / Brand -->
	<div class="p-6 flex items-center justify-between">
		<div class="flex items-center gap-3">
			<div
				class="h-10 w-10 bg-primary text-white rounded-md flex items-center justify-center shadow-sm"
			>
				<span class="font-black text-xs tracking-tighter"
					>{PUBLIC_APP_NAME.split(' ').map(n => n[0]).join('').slice(0, 3).toUpperCase()}</span
				>
			</div>
			<div class="flex flex-col min-w-0">
				<span class="text-lg font-black tracking-tight text-slate-900 leading-none truncate"
					>{PUBLIC_APP_NAME.split(' ')[0]}</span
				>
				<span class="text-[10px] font-bold text-slate-400 uppercase tracking-widest mt-1 truncate">
					{PUBLIC_APP_NAME.split(' ').slice(1).join(' ') || 'Enterprise'}
				</span>
			</div>
		</div>
		<Button variant="ghost" size="icon" class="lg:hidden" onclick={() => (isMobileOpen = false)}>
			<X class="h-5 w-5" />
		</Button>
	</div>

	<!-- Navigation Scroll Area -->
	<div class="flex-1 overflow-y-auto px-4 py-4 space-y-8">
		{#if authState.isAdmin || authState.isManager || authState.isSupervisor}
			<!-- Dashboard Link -->
			<div>
				<a
					href={resolve('/dashboard')}
					class="flex items-center gap-3 px-3 py-2.5 rounded-md font-bold text-sm transition-all {isActive(
						'/dashboard'
					)
						? 'bg-primary/10 text-primary'
						: 'text-slate-600 hover:bg-slate-50'}"
				>
					<LayoutDashboard class="h-4 w-4" />
					{$_('common.dashboard')}
				</a>
			</div>

			<!-- Grouped Admin Links -->
			{#each adminGroups as group (group.label)}
				<div class="space-y-2">
					<h4 class="px-3 text-[10px] font-black uppercase tracking-[0.2em] text-slate-400">
						{$_(group.label)}
					</h4>
					<div class="space-y-1">
						{#each group.links as link (link.href)}
							{@const Icon = link.icon}
							<a
								href={resolve(link.href)}
								class="flex items-center gap-3 px-3 py-2.5 rounded-md font-bold text-sm transition-all {isParentActive(
									link.href
								)
									? 'bg-primary/10 text-primary'
									: 'text-slate-600 hover:bg-slate-50'}"
							>
								<Icon class="h-4 w-4" />
								{$_(link.label)}
							</a>
						{/each}
					</div>
				</div>
			{/each}
		{:else if authState.isEmployee}
			<div class="space-y-2">
				<h4 class="px-3 text-[10px] font-black uppercase tracking-[0.2em] text-slate-400">
					{$_('sidebar.personal_portal')}
				</h4>
				<div class="space-y-1">
					{#each employeeLinks as link (link.href)}
						{@const Icon = link.icon}
						<a
							href={resolve(link.href)}
							class="flex items-center gap-3 px-3 py-2.5 rounded-md font-bold text-sm transition-all {isActive(
								link.href
							)
								? 'bg-primary/10 text-primary'
								: 'text-slate-600 hover:bg-slate-50'}"
						>
							<Icon class="h-4 w-4" />
							{$_(link.label)}
						</a>
					{/each}
				</div>
			</div>
		{/if}
	</div>

	<!-- Footer / User Info -->
	<div class="p-4 border-t border-slate-100 space-y-4">
		{#if authState.user}
			<div class="flex items-center gap-3 px-2">
				<Avatar.Root class="h-10 w-10 border border-slate-200">
					<Avatar.Fallback class="bg-slate-100 text-slate-600 font-bold text-xs uppercase">
						{(authState.user?.profile?.name || authState.user?.name || 'U')[0]}
					</Avatar.Fallback>
				</Avatar.Root>
				<div class="flex flex-col min-w-0">
					<span class="text-sm font-black text-slate-900 truncate">
						{authState.user?.profile?.name || authState.user?.name}
					</span>
					<span class="text-[10px] font-bold text-primary uppercase tracking-wider truncate">
						{authState.user?.role_slug}
					</span>
				</div>
			</div>
		{/if}

		<div class="px-2 pt-2 pb-1">
			<LanguageSelector />
		</div>
		<div class="grid grid-cols-2 gap-2 pb-2">
			<!-- <Button variant="ghost" class="h-9 gap-2 text-slate-500 hover:text-slate-900 justify-start px-2 font-bold text-xs">
        <Settings class="h-4 w-4" />
        Config
      </Button> -->
			<a
				href={resolve('/')}
				class="flex items-center gap-3 px-3 py-2.5 rounded-md font-bold text-sm transition-all {isActive(
					'/'
				)
					? 'bg-primary/10 text-primary'
					: 'text-slate-600 hover:bg-slate-50'}"
			>
				<LayoutDashboard class="h-4 w-4" />
				{$_('common.site')}
			</a>
			<Button
				variant="ghost"
				onclick={logout}
				class="h-9 gap-2 text-rose-500 hover:bg-rose-50 hover:text-rose-600 justify-start px-2 font-bold text-xs"
			>
				<LogOut class="h-4 w-4" />
				{$_('common.logout')}
			</Button>
		</div>
	</div>
</aside>
