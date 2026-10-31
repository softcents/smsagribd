<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';

	interface ApiHealthData {
		status: string;
		avg_response_time: string;
		p95_response_time: string;
		status_2xx_pct: number;
		status_4xx_pct: number;
		status_5xx_pct: number;
		active_nodes: number;
	}

	let data = $state<ApiHealthData | null>(null);
	let loading = $state(true);

	onMount(async () => {
		try {
			const res = await supermanApi.get<{ data: ApiHealthData }>('/api/v1/superman/monitoring/api-health');
			data = res.data ?? null;
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
			<h1 class="text-xl font-bold text-gray-900 dark:text-white">API Health & HTTP Response Spectrum</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
				Real-time ingress HTTP code distribution, percentile latencies (p95/p99), and gateway socket responsiveness.
			</p>
		</div>
	</div>

	{#if loading}
		<div class="p-8 text-center text-xs text-gray-400">Loading API health metrics...</div>
	{:else if data}
		<div class="grid grid-cols-1 md:grid-cols-4 gap-4">
			<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-5 shadow-xs">
				<span class="text-xs font-medium text-emerald-600 uppercase tracking-wider">Overall API Health</span>
				<p class="text-2xl font-bold text-emerald-600 mt-1 capitalize">{data.status}</p>
				<span class="text-[11px] text-gray-400 mt-1 block">{data.active_nodes} Active Cluster Nodes</span>
			</div>
			<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-5 shadow-xs">
				<span class="text-xs font-medium text-blue-600 uppercase tracking-wider">Average Ingress Latency</span>
				<p class="text-3xl font-bold text-blue-600 mt-1">{data.avg_response_time}</p>
				<span class="text-[11px] text-gray-400 mt-1 block">p95 Percentile: {data.p95_response_time}</span>
			</div>
			<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-5 shadow-xs">
				<span class="text-xs font-medium text-emerald-600 uppercase tracking-wider">2xx Success Ratio</span>
				<p class="text-3xl font-bold text-emerald-600 mt-1">{data.status_2xx_pct}%</p>
				<span class="text-[11px] text-emerald-600 mt-1 block">Standard 200/202 responses</span>
			</div>
			<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-5 shadow-xs">
				<span class="text-xs font-medium text-amber-600 uppercase tracking-wider">4xx / 5xx Outliers</span>
				<p class="text-3xl font-bold text-amber-600 mt-1">{data.status_4xx_pct}%</p>
				<span class="text-[11px] text-red-500 mt-1 block">5xx Server Faults: {data.status_5xx_pct}%</span>
			</div>
		</div>
	{/if}
</div>

