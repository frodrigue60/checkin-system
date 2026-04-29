<script lang="ts">
	import { onMount } from 'svelte';
	import { fly } from 'svelte/transition';
	import { quintOut } from 'svelte/easing';
	import { _ } from 'svelte-i18n';
	import { apiFetch } from '$lib/api';
	import { resolve } from '$app/paths';
	import { authState } from '$lib/auth.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import {
		LogOut,
		Users,
		MapPin,
		Activity,
		Clock,
		DollarSign,
		AlertTriangle,
		ChevronRight,
		LayoutDashboard,
		CalendarDays,
		Briefcase,
		ShieldCheck,
		Bell
	} from 'lucide-svelte';
	import ComplianceChart from '$lib/components/ComplianceChart.svelte';

	// State for different roles
	let globalData = $state({ employees: [], centers: [] });
	let employeeStats = $state({
		stats: { total_hours: 0, total_earnings: 0, incidents_count: 0 },
		recent: []
	});
	let adminStats = $state({
		total_employees: 0,
		total_centers: 0,
		recent_incidents: [],
		compliance_rate: 0,
		compliance_trend: [],
		alerts: [],
		justifications: []
	});

	async function markAlertRead(id: number) {
		try {
			await apiFetch(`/admin/alerts/${id}/read`, { method: 'POST' });
			adminStats.alerts = adminStats.alerts.filter((a: any) => a.id !== id);
		} catch (e) {
			console.error(e);
		}
	}

	async function resolveJustification(id: number, approve: boolean) {
		try {
			const res = await apiFetch(`/admin/justifications/${id}/resolve`, {
				method: 'POST',
				body: JSON.stringify({ approve, note: 'Resuelto desde Dashboard' })
			});
			if (res.ok) {
				adminStats.justifications = adminStats.justifications.filter((j: any) => j.id !== id);
			}
		} catch (e) {
			console.error(e);
		}
	}

	let loading = $state(true);
	let errorMsg = $state('');

	const roleLabel = $derived(
		authState.isAdmin
			? $_('common.role_admin')
			: authState.isManager
				? $_('common.role_manager')
				: authState.isSupervisor
					? $_('common.role_supervisor')
					: authState.isBaseUser
						? $_('common.role_base_user')
						: authState.isEmployee
							? $_('common.role_employee')
							: $_('common.role_employee')
	);

	function formatHM(timeStr: string | null) {
		if (
			!timeStr ||
			timeStr === '' ||
			timeStr.includes('0001-01-01') ||
			timeStr.includes('0000-01-01')
		)
			return '--:--';
		if (timeStr.includes('T')) {
			const parts = timeStr.split('T');
			return parts[1].substring(0, 5);
		}
		return timeStr.substring(0, 5);
	}

	async function loadDashboardData() {
		if (!authState.token) return;

		loading = true;
		try {
			if (authState.isEmployee) {
				const res = await apiFetch('/user/stats');
				if (res.ok) employeeStats = await res.json();
			} else if (authState.isAdmin || authState.isManager || authState.isSupervisor) {
				// Admin / Manager / Supervisor fetch consolidated stats
				const [empRes, centRes, statsRes] = await Promise.all([
					apiFetch(authState.isManager ? '/manager/employees' : '/admin/employees'),
					apiFetch(authState.isManager ? '/manager/centers' : '/admin/centers'),
					apiFetch('/admin/stats')
				]);
				if (empRes.ok) globalData.employees = await empRes.json();
				if (centRes.ok) globalData.centers = await centRes.json();
				if (statsRes.ok) adminStats = await statsRes.json();
			} else {
				// Base User / User with no profile
				console.log('Skipping administrative data fetch for non-privileged role');
			}
		} catch (e) {
			errorMsg = $_('dashboard.sync_error');
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		loadDashboardData();
	});

	const managementModules = $derived(
		[
			{
				id: 'attendance',
				name: $_('sidebar.attendance'),
				href: '/admin/attendance',
				icon: CalendarDays,
				roles: ['admin', 'manager', 'supervisor'],
				desc: $_('dashboard.attendance_desc')
			},
			{
				id: 'directory',
				name: $_('sidebar.directory'),
				href: '/admin/employees',
				icon: Users,
				roles: ['admin', 'manager', 'supervisor'],
				desc: $_('dashboard.employees_desc')
			},
			{
				id: 'holidays',
				name: $_('sidebar.holidays'),
				href: '/admin/holidays',
				icon: CalendarDays,
				roles: ['admin', 'supervisor'],
				desc: $_('dashboard.holidays_desc')
			},
			{
				id: 'incidents',
				name: $_('common.incidents'),
				href: '/admin/attendance',
				icon: AlertTriangle,
				roles: ['admin', 'manager', 'supervisor'],
				desc: $_('dashboard.incidents_desc', { values: { count: adminStats.recent_incidents?.length || 0 } })
			},
			{
				id: 'shifts',
				name: $_('sidebar.shifts'),
				href: '/admin/shifts',
				icon: Clock,
				roles: ['admin', 'supervisor'],
				desc: $_('dashboard.shifts_desc')
			},
			{
				id: 'positions',
				name: $_('sidebar.positions'),
				href: '/admin/positions',
				icon: Briefcase,
				roles: ['admin', 'supervisor'],
				desc: $_('dashboard.positions_desc')
			},
			{
				id: 'centers',
				name: $_('sidebar.centers'),
				href: '/admin/centers',
				icon: MapPin,
				roles: ['admin', 'manager', 'supervisor'],
				desc: $_('dashboard.centers_desc')
			},
			{
				id: 'reports',
				name: $_('common.reports'),
				href: '/admin/reports',
				icon: Activity,
				roles: ['admin', 'supervisor'],
				desc: $_('dashboard.reports_desc')
			}
		].filter((module) => module.roles.includes(authState.user?.role_slug || ''))
	);
</script>

<div
	class="p-6 md:p-10 space-y-10"
	in:fly={{ y: 20, duration: 800, easing: quintOut }}
>
	<header class="flex flex-col md:flex-row justify-between items-start md:items-center gap-6">
		<div class="space-y-1">
			<h1 class="text-5xl font-black tracking-tighter text-slate-900">
				{authState.isEmployee ? $_('dashboard.my_folder') : $_('dashboard.management_hub')}
			</h1>
			<div class="flex items-center gap-2 text-sm font-medium text-muted-foreground">
				<span>{$_('dashboard.session_as')}</span>
				<Badge
					variant="outline"
					class="bg-primary/5 text-primary border-primary/20 font-bold px-2 py-0"
				>
					{roleLabel}
				</Badge>
				<span class="text-slate-300">•</span>
				<span class="font-bold text-slate-700">{authState.user?.email}</span>
			</div>
		</div>
	</header>

	{#if loading}
		<div class="grid grid-cols-1 md:grid-cols-3 gap-6">
			{#each Array(3) as _, i (i)}
				<div class="h-40 rounded-md bg-white animate-pulse border shadow-sm"></div>
			{/each}
		</div>
	{:else if authState.isBaseUser}
		<!-- USER HUB: Unassigned state -->
		<section
			class="py-16 flex flex-col items-center text-center space-y-8 max-w-2xl mx-auto"
			in:fly={{ y: 20, duration: 800, easing: quintOut, delay: 100 }}
		>
			<div
				class="h-24 w-24 rounded-md bg-primary/10 flex items-center justify-center text-primary shadow-xl shadow-primary/10 ring-1 ring-primary/20"
			>
				<ShieldCheck class="h-12 w-12" />
			</div>
			<div class="space-y-3">
				<h2 class="text-4xl font-black tracking-tight text-slate-900">
					{$_('common.welcome_user', { values: { name: authState.user?.name } })}
				</h2>
				<p class="text-lg text-slate-500 font-medium leading-relaxed">
					{$_('dashboard.base_user_welcome')} <span
						class="text-primary font-bold italic underline decoration-2">{$_('dashboard.base_user_level')}</span
					>.
				</p>
				<p class="text-slate-400 font-medium">
					{$_('dashboard.base_user_hint')}
				</p>
			</div>

			<div
				class="p-6 bg-primary/5 rounded-md border border-primary/20 flex items-center gap-4 text-left"
			>
				<div
					class="h-10 w-10 rounded-full bg-primary text-white flex items-center justify-center animate-bounce"
				>
					<Bell class="h-5 w-5" />
				</div>
				<div>
					<p class="text-primary font-black text-sm uppercase tracking-wider">
						{$_('dashboard.config_notice')}
					</p>
					<p class="text-slate-600 text-sm font-medium">
						{$_('dashboard.config_notice_desc')}
					</p>
				</div>
			</div>
		</section>
	{:else if authState.isEmployee}
		<!-- Employee View -->
		<section class="grid grid-cols-1 md:grid-cols-3 gap-6">
			<Card.Root class="border-none shadow-premium rounded-md overflow-hidden group">
				<Card.Content class="p-6 flex items-center gap-6">
					<div
						class="h-16 w-16 rounded-md bg-indigo-50 text-indigo-600 flex items-center justify-center group-hover:scale-110 transition-transform"
					>
						<Clock class="h-8 w-8" />
					</div>
					<div class="space-y-1">
						<p
							class="text-[10px] font-bold uppercase tracking-[0.2em] text-muted-foreground italic"
						>
							{$_('dashboard.hours_this_month')}
						</p>
						<p class="text-4xl font-black text-slate-900">
							{employeeStats.stats.total_hours.toFixed(1)}
							<span class="text-lg font-bold text-slate-400">hrs</span>
						</p>
					</div>
				</Card.Content>
			</Card.Root>

			<Card.Root class="border-none shadow-premium rounded-md overflow-hidden group">
				<Card.Content class="p-6 flex items-center gap-6">
					<div
						class="h-16 w-16 rounded-md bg-emerald-50 text-emerald-600 flex items-center justify-center group-hover:scale-110 transition-transform"
					>
						<DollarSign class="h-8 w-8" />
					</div>
					<div class="space-y-1">
						<p
							class="text-[10px] font-bold uppercase tracking-[0.2em] text-muted-foreground italic"
						>
							{$_('dashboard.estimated_earnings')}
						</p>
						<p class="text-4xl font-black text-slate-900">
							${employeeStats.stats.total_earnings.toLocaleString()}
							<span class="text-lg font-bold text-slate-400">MXN</span>
						</p>
					</div>
				</Card.Content>
			</Card.Root>

			<Card.Root class="border-none shadow-premium rounded-md overflow-hidden group">
				<Card.Content class="p-6 flex items-center gap-6">
					<div
						class="h-16 w-16 rounded-md bg-rose-50 text-rose-600 flex items-center justify-center group-hover:scale-110 transition-transform"
					>
						<AlertTriangle class="h-8 w-8" />
					</div>
					<div class="space-y-1">
						<p
							class="text-[10px] font-bold uppercase tracking-[0.2em] text-muted-foreground italic"
						>
							{$_('common.incidents')}
						</p>
						<p class="text-4xl font-black text-slate-900">{employeeStats.stats.incidents_count}</p>
					</div>
				</Card.Content>
			</Card.Root>
		</section>

		<section class="space-y-6">
			<div class="flex justify-between items-end">
				<div class="space-y-1">
					<h2 class="text-2xl font-black tracking-tight text-slate-800">{$_('dashboard.recent_activity')}</h2>
					<p class="text-sm text-muted-foreground font-medium uppercase tracking-[0.2em]">
						{$_('dashboard.latest_records')}
					</p>
				</div>
				<Button
					href={resolve('/')}
					class="rounded-md font-bold px-8 shadow-xl shadow-primary/20 transition-all hover:-translate-y-1 active:scale-95"
				>
					{$_('dashboard.go_to_checkin')}
				</Button>
			</div>

			<div class="space-y-4">
				{#if employeeStats.recent.length === 0}
					<div
						class="p-20 text-center border-2 border-dashed rounded-md text-muted-foreground font-bold uppercase tracking-widest bg-slate-50/50"
					>
						{$_('dashboard.no_recent_records')}
					</div>
				{:else}
					{#each employeeStats.recent as attendance (attendance.id)}
						<Card.Root
							class="border-none shadow-premium rounded-md overflow-hidden hover:shadow-xl transition-shadow border-l-4 border-l-primary"
						>
							<Card.Content class="p-6 flex flex-col md:flex-row items-center gap-8">
								<div
									class="flex flex-col items-center min-w-[4rem] border-r pr-8 h-full justify-center"
								>
									<span class="text-3xl font-black text-slate-900 leading-none"
										>{new Date(attendance.date).toLocaleDateString(undefined, {
											day: 'numeric'
										})}</span
									>
									<span
										class="text-[10px] font-black uppercase text-muted-foreground tracking-widest"
										>{new Date(attendance.date).toLocaleDateString(undefined, {
											month: 'short'
										})}</span
									>
								</div>
								<div class="flex-1 space-y-2">
									<div class="flex items-center gap-4 text-slate-700">
										<div class="flex items-center gap-1.5">
											<Clock class="h-3 w-3 text-primary" />
											<span class="text-sm font-bold">{$_('common.check_in')}:</span>
											<span class="text-sm font-medium">{formatHM(attendance.check_in)}</span>
										</div>
										<span class="text-slate-300">|</span>
										<div class="flex items-center gap-1.5">
											<Clock class="h-3 w-3 text-rose-500" />
											<span class="text-sm font-bold">{$_('common.check_out')}:</span>
											<span class="text-sm font-medium">{formatHM(attendance.check_out)}</span>
										</div>
									</div>
									<div
										class="flex items-center gap-2 text-xs text-muted-foreground font-medium italic"
									>
										<MapPin class="h-3 w-3" />
										{$_('dashboard.gps_confirmed_hint')}
									</div>
								</div>
								<div class="text-center md:text-right">
									<p class="text-2xl font-black text-primary leading-none">
										{attendance.net_hours_worked.toFixed(1)}
									</p>
									<p class="text-[10px] font-bold text-muted-foreground uppercase tracking-widest">
										{$_('common.hours')}
									</p>
								</div>
							</Card.Content>
						</Card.Root>
					{/each}
				{/if}
			</div>
		</section>
	{:else}
		<!-- Admin / Manager / Supervisor View (Precision Attendance System) -->
		<!-- Compact Metrics Bar -->
		<div class="bg-slate-50 px-4 py-4 grid grid-cols-3 gap-2 border rounded-sm mb-8">
			<div class="flex flex-col">
				<span class="text-[10px] uppercase font-black tracking-wider text-slate-400"
					>{$_('dashboard.system_status')}</span
				>
				<div class="flex items-center gap-1.5">
					<span class="text-xl font-black text-slate-900">{adminStats.compliance_rate}%</span>
					<span class="text-[10px] text-emerald-600 font-bold">+2.1%</span>
				</div>
			</div>
			<div class="flex flex-col border-l border-slate-200 pl-4">
				<span class="text-[10px] uppercase font-black tracking-wider text-slate-400"
					>{$_('common.personnel')}</span
				>
				<span class="text-xl font-black text-slate-900">{adminStats.total_employees}</span>
			</div>
			<div class="flex flex-col border-l border-slate-200 pl-4">
				<span class="text-[10px] uppercase font-black tracking-wider text-slate-400"
					>{$_('common.incidents')}</span
				>
				<div class="flex items-center gap-1.5">
					<span class="text-xl font-black text-slate-900"
						>{adminStats.recent_incidents?.length || 0}</span
					>
					{#if (adminStats.recent_incidents?.length || 0) > 0}
						<span class="w-2 h-2 rounded-full bg-rose-500 animate-pulse"></span>
					{/if}
				</div>
			</div>
		</div>

		<ComplianceChart data={adminStats.compliance_trend} />

		<section class="space-y-8">
			<div class="space-y-1 px-1">
				<h2 class="text-xs font-black uppercase tracking-[0.2em] text-slate-400">
					{$_('dashboard.management_modules')}
				</h2>
			</div>
			<div class="grid grid-cols-2 md:grid-cols-4 gap-4">
				{#each managementModules as module (module.id)}
					<a
						href={resolve(module.href)}
						class="bg-white border-none shadow-sm rounded-sm p-4 flex flex-col gap-3 hover:shadow-md hover:-translate-y-1 transition-all active:scale-[0.98] no-underline group"
					>
						<div
							class="w-10 h-10 rounded bg-slate-50 text-slate-600 flex items-center justify-center group-hover:bg-primary group-hover:text-white transition-colors"
						>
							<module.icon class="h-5 w-5" />
						</div>
						<div>
							<h3
								class="text-sm font-black text-slate-900 tracking-tight group-hover:text-primary transition-colors"
							>
								{module.name}
							</h3>
							<p class="text-[11px] text-slate-500 font-medium leading-tight mt-1 line-clamp-2">
								{module.desc}
							</p>
						</div>
					</a>
				{/each}
			</div>
		</section>

		<div class="grid grid-cols-1 md:grid-cols-2 gap-8 mt-10">
			<!-- Critical Events -->
			<section class="space-y-4">
				<div class="flex items-center justify-between px-1">
					<h2 class="text-xs font-black uppercase tracking-[0.2em] text-slate-400">
						{$_('dashboard.critical_events')}
					</h2>
					<button class="text-[10px] font-black text-primary uppercase hover:underline"
						>{$_('common.view_all')}</button
					>
				</div>
				<div
					class="bg-white rounded-sm shadow-sm border-none overflow-hidden divide-y divide-slate-50"
				>
					{#if (adminStats.recent_incidents?.length || 0) === 0 && (adminStats.alerts?.length || 0) === 0 && (adminStats.justifications?.length || 0) === 0}
						<div
							class="p-8 text-center text-slate-400 text-xs font-bold uppercase tracking-widest italic"
						>
							{$_('dashboard.no_anomalies')}
						</div>
					{:else}
						<!-- Justifications -->
						{#each adminStats.justifications || [] as just (just.id)}
							<div class="p-4 flex items-start gap-4 group bg-indigo-50/20">
								<div class="w-6 h-6 rounded-full bg-indigo-100 text-indigo-600 flex items-center justify-center mt-0.5">
									<Activity class="h-3 w-3" />
								</div>
								<div class="flex-1 min-w-0">
									<p class="text-xs font-black text-slate-900 uppercase tracking-tight">
										{$_('history.send_justification')}: Emp #{just.employee_id}
									</p>
									<p class="text-[10px] text-indigo-700 font-medium line-clamp-2 italic">
										"{just.message}"
									</p>
								</div>
								<div class="flex flex-col gap-1 items-end">
									<div class="flex gap-1">
										<button 
											onclick={() => resolveJustification(just.id, true)}
											class="p-1 text-emerald-600 hover:bg-emerald-50 rounded"
											title="Aprobar"
										>
											<ShieldCheck class="h-4 w-4" />
										</button>
										<button 
											onclick={() => resolveJustification(just.id, false)}
											class="p-1 text-rose-600 hover:bg-rose-50 rounded"
											title="Rechazar"
										>
											<AlertTriangle class="h-4 w-4" />
										</button>
									</div>
									<span class="text-[9px] font-bold text-slate-400">
										{new Date(just.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
									</span>
								</div>
							</div>
						{/each}

						<!-- Incidents -->
						{#each adminStats.recent_incidents || [] as incident (incident.id)}
							<div class="p-4 flex items-start gap-4 group">
								<div
									class="w-6 h-6 rounded-full bg-rose-50 text-rose-500 flex items-center justify-center mt-0.5 group-hover:bg-rose-500 group-hover:text-white transition-colors"
								>
									<AlertTriangle class="h-3 w-3" />
								</div>
								<div class="flex-1 min-w-0">
									<p class="text-xs font-black text-slate-900 truncate uppercase tracking-tight">
										{incident.type === 'late' ? $_('incidents.late') : $_('incidents.out_of_range')}: {incident.employee_name}
									</p>
									<p class="text-[10px] text-slate-500 font-medium line-clamp-1">
										{incident.center_name || $_('dashboard.unidentified_center')}
									</p>
								</div>
								<span class="text-[10px] font-black text-slate-400 bg-slate-50 px-2 py-1 rounded">
									{new Date(incident.created_at).toLocaleTimeString([], {
										hour: '2-digit',
										minute: '2-digit'
									})}
								</span>
							</div>
						{/each}

						<!-- Alerts -->
						{#each adminStats.alerts || [] as alert (alert.id)}
							<div class="p-4 flex items-start gap-4 group bg-amber-50/30">
								<div
									class="w-6 h-6 rounded-full {alert.severity === 'critical' ? 'bg-rose-100 text-rose-600' : 'bg-amber-100 text-amber-600'} flex items-center justify-center mt-0.5"
								>
									<Bell class="h-3 w-3" />
								</div>
								<div class="flex-1 min-w-0">
									<p class="text-xs font-black text-slate-900 uppercase tracking-tight">
										{$_('dashboard.system_alert')}: {alert.type}
									</p>
									<p class="text-[10px] text-slate-600 font-medium italic">
										{alert.message}
									</p>
								</div>
								<button 
									onclick={() => markAlertRead(alert.id)}
									class="text-[10px] font-black text-primary hover:underline"
								>
									OK
								</button>
							</div>
						{/each}
					{/if}
				</div>
			</section>

			<!-- Quick Snapshot -->
			<section
				class="bg-slate-900 rounded-2xl p-6 text-white shadow-xl shadow-slate-900/20 flex flex-col justify-between"
			>
				<div class="space-y-4">
					<div class="flex items-center justify-between">
						<h3 class="text-lg font-black tracking-tight">{$_('dashboard.quick_snapshot')}</h3>
						<Activity class="h-5 w-5 text-indigo-400" />
					</div>
					<div class="space-y-2">
						{#each globalData.centers.slice(0, 2) as center (center.id)}
							<a
								href={resolve(`/admin/centers/${center.id}`)}
								class="p-3 rounded-lg bg-slate-800 border border-slate-700 flex items-center justify-between hover:bg-slate-700 transition-colors no-underline group"
							>
								<span
									class="text-xs font-bold text-slate-300 group-hover:text-white transition-colors"
									>{center.name}</span
								>
								<ChevronRight
									class="h-4 w-4 text-slate-500 group-hover:translate-x-1 transition-transform"
								/>
							</a>
						{/each}
					</div>
				</div>
				<Button
					href={resolve('/admin/reports')}
					variant="secondary"
					class="mt-6 w-full py-6 font-black text-xs uppercase tracking-widest bg-white text-slate-900 hover:bg-slate-100 border-none transition-all active:scale-95 gap-2"
				>
					<Activity class="h-4 w-4" />
					{$_('dashboard.generate_global_report')}
				</Button>
			</section>
		</div>
	{/if}
</div>
