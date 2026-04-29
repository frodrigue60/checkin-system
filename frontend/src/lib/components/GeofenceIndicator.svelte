<script lang="ts">
	import { _ } from 'svelte-i18n';
	let { inRange = false, distance = 0 } = $props<{
		inRange: boolean;
		distance: number;
	}>();

	let pulseClass = $derived(inRange ? 'pulse-success' : 'pulse-error');
</script>

<div class="geofence-container card-premium">
	<div class="header">
		<div class="dot {pulseClass}"></div>
		<span class="label">{inRange ? $_('geofence.inside') : $_('geofence.outside')}</span>
	</div>
	<div class="footer">
		<span class="distance">
			{#if distance > 0}
				{$_('geofence.distance_msg', { values: { distance: distance.toFixed(0) } })}
			{:else}
				{$_('geofence.detecting')}
			{/if}
		</span>
	</div>
</div>

<style>
	.geofence-container {
		padding: 1.25rem;
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
		max-width: 260px;
		border-radius: var(--radius-md);
	}

	.header {
		display: flex;
		align-items: center;
		gap: 0.875rem;
	}

	.dot {
		width: 12px;
		height: 12px;
		border-radius: 50%;
		position: relative;
	}

	.dot::after {
		content: '';
		position: absolute;
		top: -4px;
		left: -4px;
		right: -4px;
		bottom: -4px;
		border-radius: 50%;
		opacity: 0.4;
		animation: pulse 2s infinite;
	}

	.pulse-success { background-color: var(--accent-success); }
	.pulse-success::after { background-color: var(--accent-success); }

	.pulse-error { background-color: var(--accent-danger); }
	.pulse-error::after { background-color: var(--accent-danger); }

	.label {
		font-size: 0.875rem;
		font-weight: 700;
		color: var(--text-primary);
	}

	.footer {
		margin-top: 0.25rem;
	}

	.distance {
		font-size: 0.775rem;
		color: var(--text-secondary);
		font-weight: 500;
	}

	@keyframes pulse {
		0% { transform: scale(0.8); opacity: 0.5; }
		100% { transform: scale(2.2); opacity: 0; }
	}
</style>
