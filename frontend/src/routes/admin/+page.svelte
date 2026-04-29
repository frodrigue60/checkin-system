<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';

	onMount(() => {
		const userStr = localStorage.getItem('user');
		if (userStr) {
			const user = JSON.parse(userStr);
			if (user.role_id !== 1) {
				goto('/dashboard');
				return;
			}
		} else {
			goto('/');
			return;
		}
		goto('/admin/centers');
	});
</script>

<div class="admin-portal">
	<p>Redirecting to administration panel...</p>
</div>
