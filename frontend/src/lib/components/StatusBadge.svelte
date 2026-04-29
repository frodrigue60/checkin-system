<script lang="ts">
	let { status, type = 'info' } = $props<{
		status: string;
		type?: 'success' | 'warning' | 'error' | 'info';
	}>();

	const colors = {
		success: 'badge-success',
		warning: 'badge-warning',
		error: 'badge-error',
		info: 'badge-info'
	};
</script>

<span class="badge {colors[type]}">
	{status}
</span>

<style>
	.badge {
		display: inline-flex;
		align-items: center;
		padding: 0.25rem 0.875rem;
		border-radius: 9999px;
		font-size: 0.725rem;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.025em;
		border: 1px solid transparent;
	}

	.badge-success {
		background-color: rgba(16, 185, 129, 0.08);
		color: var(--accent-success);
		border-color: rgba(16, 185, 129, 0.15);
	}

	.badge-warning {
		background-color: rgba(245, 158, 11, 0.08);
		color: var(--accent-warning);
		border-color: rgba(245, 158, 11, 0.15);
	}

	.badge-error {
		background-color: rgba(239, 68, 68, 0.08);
		color: var(--accent-danger);
		border-color: rgba(239, 68, 68, 0.15);
	}

	.badge-info {
		background-color: rgba(79, 70, 229, 0.08);
		color: var(--primary);
		border-color: rgba(79, 70, 229, 0.15);
	}
</style>
