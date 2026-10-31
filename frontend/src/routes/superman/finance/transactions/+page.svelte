<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';

	onMount(() => {
		goto('/superman/billing/transactions', { replaceState: true });
	});
</script>

<div class="flex h-64 items-center justify-center">
	<div class="flex items-center gap-2 text-xs text-gray-400">
		<div class="h-4 w-4 animate-spin rounded-full border-2 border-emerald-500 border-t-transparent"></div>
		<span>Navigating to Transactions…</span>
	</div>
</div>
