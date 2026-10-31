<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';

	interface SystemHealthData {
		status: string;
		uptime_seconds: number;
		goroutines: number;
		memory_alloc_mb: number;
		memory_sys_mb: number;
		database_state: string;
		open_db_conns: number;
		in_use_db_conns: number;
		idle_db_conns: number;
		queues: Record<string, number>;
		gateways_health: Record<string, string>;
		environment: string;
		server_time: string;
	}

	let data = $state<SystemHealthData | null>(null);
	let loading = $state(true);

	onMount(async () => {
		try {
			const res = await supermanApi.get<{ data: SystemHealthData }>('/api/v1/superman/monitoring/system-health');
			data = res.data ?? null;
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	});

	function formatUptime(seconds: number): string {
		const d = Math.floor(seconds / 86400);
		const h = Math.floor((seconds % 86400) / 3600);
		const m = Math.floor((seconds % 3600) / 60);
		return `${d}d ${h}h ${m}m`;
	}
</script>

<div class="space-y-6">
	<!-- Page Header -->
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-xl font-bold text-gray-900 dark:text-white">System Health & Runtime Diagnostics</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
				Real-time Go runtime telemetry, memory footprints, MySQL connection pool stats, and server vitals.
			</p>
		</div>
		<button
			type="button"
			onclick={() => window.location.reload()}
			class="inline-flex items-center gap-2 rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 px-3 py-1.5 text-xs font-semibold text-gray-700 dark:text-gray-300 shadow-xs hover:bg-gray-50 transition"
		>
			Refresh Metrics
		</button>
	</div>

	{#if loading}
		<div class="p-8 text-center text-xs text-gray-400">Loading system runtime telemetry...</div>
	{:else if data}
		<!-- KPI Vitals -->
		<div class="grid grid-cols-1 md:grid-cols-4 gap-4">
			<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-5 shadow-xs">
				<span class="text-xs font-medium text-emerald-600 uppercase tracking-wider">System State</span>
				<p class="text-2xl font-bold text-emerald-600 mt-1 capitalize">{data.status}</p>
				<span class="text-[11px] text-gray-400 mt-1 block">Uptime: {formatUptime(data.uptime_seconds)}</span>
			</div>
			<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-5 shadow-xs">
				<span class="text-xs font-medium text-blue-600 uppercase tracking-wider">Memory Allocation</span>
				<p class="text-3xl font-bold text-blue-600 mt-1">{data.memory_alloc_mb.toFixed(1)} MB</p>
				<span class="text-[11px] text-gray-400 mt-1 block">Sys: {data.memory_sys_mb.toFixed(1)} MB</span>
			</div>
			<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-5 shadow-xs">
				<span class="text-xs font-medium text-purple-600 uppercase tracking-wider">Go Goroutines</span>
				<p class="text-3xl font-bold text-purple-600 mt-1">{data.goroutines}</p>
				<span class="text-[11px] text-gray-400 mt-1 block">Active lightweight threads</span>
			</div>
			<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-5 shadow-xs">
				<span class="text-xs font-medium text-amber-600 uppercase tracking-wider">MySQL Connections</span>
				<p class="text-3xl font-bold text-amber-600 mt-1">{data.open_db_conns}</p>
				<span class="text-[11px] text-gray-400 mt-1 block">In-use: {data.in_use_db_conns} | Idle: {data.idle_db_conns}</span>
			</div>
		</div>

		<!-- Details Card -->
		<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
			<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-6 shadow-xs space-y-4">
				<h3 class="text-sm font-bold text-gray-900 dark:text-white">Active Dispatch Queues</h3>
				<div class="space-y-3">
					{#each Object.entries(data.queues) as [qName, qSize]}
						<div class="flex items-center justify-between text-xs p-3 rounded-lg bg-gray-50 dark:bg-gray-800/40">
							<span class="font-mono font-semibold text-gray-800 dark:text-gray-200">{qName}</span>
							<span class="font-bold text-emerald-600">{qSize} pending</span>
						</div>
					{/each}
				</div>
			</div>

			<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-6 shadow-xs space-y-4">
				<h3 class="text-sm font-bold text-gray-900 dark:text-white">Gateway Interconnect States</h3>
				<div class="space-y-3">
					{#each Object.entries(data.gateways_health) as [gwName, gwState]}
						<div class="flex items-center justify-between text-xs p-3 rounded-lg bg-gray-50 dark:bg-gray-800/40">
							<span class="font-semibold text-gray-800 dark:text-gray-200">{gwName}</span>
							<span class="inline-flex items-center gap-1.5 font-medium {gwState === 'healthy' ? 'text-emerald-600' : 'text-amber-600'}">
								<span class="h-1.5 w-1.5 rounded-full {gwState === 'healthy' ? 'bg-emerald-600' : 'bg-amber-600'}"></span>
								{gwState}
							</span>
						</div>
					{/each}
				</div>
			</div>
		</div>
	{/if}
</div>

