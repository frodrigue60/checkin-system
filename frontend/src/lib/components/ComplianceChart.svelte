<script lang="ts">
	import { Chart, Svg, Axis, Line, Area, Tooltip, Highlight } from 'layerchart';
	import { scaleTime } from 'd3-scale';
	import { curveMonotoneX } from 'd3-shape';
	import { format } from 'date-fns';
	import { _ } from 'svelte-i18n';

	let { data = [] } = $props();

	// Parse dates for d3-scale
	const chartData = $derived(
		data.map((d: any) => ({
			...d,
			date: new Date(d.date)
		}))
	);
</script>

<div class="h-[300px] w-full bg-white p-4 rounded-xl border border-slate-100 shadow-sm">
	<div class="mb-4 flex items-center justify-between">
		<h3 class="text-sm font-black uppercase tracking-wider text-slate-400">
			{$_('dashboard.compliance_trend')}
		</h3>
		<div class="flex items-center gap-4">
			<div class="flex items-center gap-2">
				<div class="h-3 w-3 rounded-full bg-primary"></div>
				<span class="text-[10px] font-bold text-slate-500">{$_('dashboard.compliance')} (%)</span>
			</div>
		</div>
	</div>

	{#if chartData.length > 0}
		<Chart
			data={chartData}
			x="date"
			xScale={scaleTime()}
			y="compliance"
			yDomain={[0, 100]}
			padding={{ left: 30, bottom: 20, right: 10, top: 10 }}
			tooltip={{ mode: 'bisect-x' }}
		>
			<Svg>
				<Axis placement="left" grid={{ stroke: '#f1f5f9' }} rule={{ stroke: '#f1f5f9' }} 
					ticks={5} format={(v) => `${v}%`} 
					classes={{ label: 'text-[10px] font-bold text-slate-400' }} />
				<Axis placement="bottom" grid={{ stroke: '#f1f5f9' }} rule={{ stroke: '#f1f5f9' }}
					format={(v) => format(v, 'dd MMM')}
					classes={{ label: 'text-[10px] font-bold text-slate-400' }} />
				
				<Area
					line={{ stroke: 'var(--color-primary)', strokeWidth: 3 }}
					fill="url(#area-gradient)"
					curve={curveMonotoneX}
				/>
				
				<Line
					stroke="var(--color-primary)"
					strokeWidth={3}
					curve={curveMonotoneX}
				/>
				
				<Highlight points lines={{ stroke: 'var(--color-primary)', strokeDasharray: '4 4' }} />
				
				<defs>
					<linearGradient id="area-gradient" x1="0" y1="0" x2="0" y2="1">
						<stop offset="0%" stop-color="var(--color-primary)" stop-opacity="0.15" />
						<stop offset="100%" stop-color="var(--color-primary)" stop-opacity="0" />
					</linearGradient>
				</defs>
			</Svg>

			<Tooltip.Root 
				classes={{ container: 'bg-white border border-slate-200 p-4 rounded-xl shadow-[0_20px_50px_rgba(0,0,0,0.15)] min-w-[180px]' }}
				let:data
			>
				<div class="space-y-2">
					<div class="text-[10px] font-black text-slate-400 uppercase tracking-widest border-b border-slate-100 pb-2 mb-2">
						{format(data.date, 'EEEE, dd MMMM')}
					</div>
					<Tooltip.Item label={$_('dashboard.compliance')} value={`${data.compliance.toFixed(1)}%`} classes={{ label: 'text-slate-500 text-[10px] font-bold uppercase', value: 'text-slate-900 font-black text-sm' }} />
					<Tooltip.Item label={$_('common.attendance')} value={data.attendance} classes={{ label: 'text-slate-500 text-[10px] font-bold uppercase', value: 'text-slate-900 font-black text-sm' }} />
					<Tooltip.Item label={$_('common.incidents')} value={data.incidents} color="#f43f5e" classes={{ label: 'text-slate-500 text-[10px] font-bold uppercase', value: 'text-rose-600 font-black text-sm' }} />
				</div>
			</Tooltip.Root>
		</Chart>
	{:else}
		<div class="h-full flex items-center justify-center">
			<p class="text-xs font-bold text-slate-400 uppercase tracking-widest italic">
				{$_('dashboard.no_data_available')}
			</p>
		</div>
	{/if}
</div>

<style>
	:global(.layerchart-axis-label) {
		font-family: inherit;
	}
</style>
