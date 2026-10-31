<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface GatewayTps {
		id: number;
		name: string;
		driver: string;
		limit: number;
		tps: number;
		latency_ms: number;
		is_enabled: boolean;
	}

	interface TpsData {
		current_tps: number;
		peak_today: number;
		capacity: number;
		gateways: GatewayTps[];
	}

	let data = $state<TpsData | null>(null);
	let loading = $state(true);

	const exportColumns: ExportColumn<GatewayTps>[] = [
		{ header: 'ID', getValue: (g) => g.id },
		{ header: 'Gateway Provider', getValue: (g) => g.name },
		{ header: 'Driver', getValue: (g) => g.driver.toUpperCase() },
		{ header: 'Current TPS', getValue: (g) => `${g.tps} req/sec` },
		{ header: 'Configured Limit', getValue: (g) => `${g.limit} req/sec` },
		{ header: 'Probe Latency', getValue: (g) => `${g.latency_ms} ms` },
		{ header: 'Status', getValue: (g) => g.is_enabled ? 'ONLINE' : 'DISABLED' }
	];

	async function loadTps() {
		loading = true;
		try {
			const res = await supermanApi.get<{ data: TpsData }>('/api/v1/superman/monitoring/tps');
			data = res.data ?? null;
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		loadTps();
	});
</script>

<div class="space-y-6">
	<!-- Page Header -->
	<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
		<div>
			<h1 class="text-xl font-bold text-gray-900 dark:text-white">Throughput & TPS Telemetry</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
				Real-time transaction per second (TPS) rates across upstream carrier pipes, peak bursts, and headroom.
			</p>
		</div>
		<div class="flex items-center gap-3">
			<button
				onclick={loadTps}
				class="inline-flex items-center gap-1.5 px-3 py-2 text-xs font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 dark:bg-gray-800 dark:text-gray-300 dark:border-gray-700 dark:hover:bg-gray-700/60 shadow-xs transition"
			>
				<svg class="w-3.5 h-3.5 {loading ? 'animate-spin' : ''}" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
				</svg>
				Refresh
			</button>
			<ExportDropdown
				title="Gateway Throughput and TPS"
				columns={exportColumns}
				getData={() => data?.gateways || []}
				disabled={!data || data.gateways.length === 0}
			/>
		</div>
	</div>

	{#if loading && !data}
		<div class="p-12 text-center text-xs text-gray-400">Loading TPS metrics...</div>
	{:else if data}
		<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
			<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-5 shadow-xs">
				<span class="text-xs font-medium text-emerald-600 uppercase tracking-wider">Current Pipeline Speed</span>
				<p class="text-3xl font-bold text-emerald-600 mt-1">{data.current_tps} TPS</p>
				<span class="text-[11px] text-gray-400 mt-1 block">Live messages dispatched / sec</span>
			</div>
			<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-5 shadow-xs">
				<span class="text-xs font-medium text-blue-600 uppercase tracking-wider">Peak Burst Recorded Today</span>
				<p class="text-3xl font-bold text-blue-600 mt-1">{data.peak_today} TPS</p>
				<span class="text-[11px] text-gray-400 mt-1 block">Highest instantaneous throughput</span>
			</div>
			<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-5 shadow-xs">
				<span class="text-xs font-medium text-purple-600 uppercase tracking-wider">Cluster Capacity Limit</span>
				<p class="text-3xl font-bold text-purple-600 mt-1">{data.capacity} TPS</p>
				<span class="text-[11px] text-emerald-600 mt-1 block">Aggregate active gateway bandwidth</span>
			</div>
		</div>

		<!-- Gateway Throughput Grid -->
		<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl shadow-xs overflow-hidden">
			<div class="px-5 py-4 border-b border-gray-200 dark:border-gray-800">
				<h2 class="text-sm font-bold text-gray-900 dark:text-white">Active Carrier Gateway Throughput & Headroom</h2>
				<p class="text-xs text-gray-400 mt-0.5">Real-time throughput relative to configured rate limits</p>
			</div>
			<div class="p-6 space-y-5">
				{#each data.gateways as gw}
					<div class="p-4 rounded-xl bg-gray-50 dark:bg-gray-800/40 border border-gray-100 dark:border-gray-800">
						<div class="flex items-center justify-between text-xs mb-2">
							<div class="flex items-center gap-2">
								<span class="font-bold text-gray-900 dark:text-white text-sm">{gw.name}</span>
								<span class="text-[11px] font-mono uppercase px-1.5 py-0.5 rounded bg-gray-200 dark:bg-gray-700 text-gray-600 dark:text-gray-300">{gw.driver}</span>
								{#if gw.is_enabled}
									<span class="h-2 w-2 rounded-full bg-emerald-500"></span>
								{:else}
									<span class="text-[10px] text-gray-400 font-semibold uppercase">(Disabled)</span>
								{/if}
							</div>
							<div class="flex items-center gap-4 text-xs font-mono">
								{#if gw.latency_ms > 0}
									<span class="text-gray-400">{gw.latency_ms}ms latency</span>
								{/if}
								<span class="font-bold text-gray-900 dark:text-white">{gw.tps} / {gw.limit} req/sec</span>
							</div>
						</div>
						<div class="h-2.5 w-full rounded-full bg-gray-200 dark:bg-gray-700 overflow-hidden">
							<div
								class="h-full rounded-full bg-emerald-500 transition-all duration-500"
								style="width: {Math.min(100, (gw.tps / (gw.limit || 1)) * 100)}%"
							></div>
						</div>
					</div>
				{/each}
			</div>
		</div>
	{/if}
</div>
