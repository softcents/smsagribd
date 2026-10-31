<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';

	interface QueueItem {
		name: string;
		connection: string;
		size: number;
		status: string;
	}

	let queues = $state<QueueItem[]>([]);
	let loading = $state(true);

	onMount(async () => {
		try {
			const res = await supermanApi.get<{ data: QueueItem[] }>('/api/v1/superman/monitoring/queue-health');
			queues = res.data ?? [];
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	});
</script>

<div class="space-y-6">
	<!-- Page Header -->
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-xl font-bold text-gray-900 dark:text-white">Queue Health & Backpressure Telemetry</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
				Real-time monitoring of job consumption pipelines, worker pool health, and delayed jobs.
			</p>
		</div>
		<div>
			<a
				href="/superman/system-settings/queue"
				class="inline-flex items-center gap-2 rounded-lg bg-emerald-600 px-3 py-1.5 text-xs font-semibold text-white shadow-xs hover:bg-emerald-700 transition"
			>
				Queue Settings
			</a>
		</div>
	</div>

	<!-- Queues Grid -->
	<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
		{#if loading}
			<div class="col-span-4 p-8 text-center text-xs text-gray-400">Loading queue health metrics...</div>
		{:else}
			{#each queues as q}
				<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-5 shadow-xs">
					<div class="flex items-center justify-between">
						<span class="text-xs font-mono font-bold text-gray-900 dark:text-white">{q.name}</span>
						<span class="inline-flex items-center gap-1.5 text-[11px] text-emerald-600 font-medium">
							<span class="h-1.5 w-1.5 rounded-full bg-emerald-600"></span>
							{q.status}
						</span>
					</div>
					<p class="text-3xl font-bold text-gray-900 dark:text-white mt-3">{q.size} <span class="text-sm font-normal text-gray-400">jobs</span></p>
					<span class="text-[11px] text-gray-400 mt-1 block">{q.connection}</span>
				</div>
			{/each}
		{/if}
	</div>
</div>

