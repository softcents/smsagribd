<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface GatewayLoadBalance {
		id: number;
		name: string;
		gateway: string;
		driver: string;
		requests_per_second: number;
		capacity_cap: string;
		priority: number;
		supports_bulk: boolean;
		timeout_seconds: number;
	}

	interface LoadBalancingResponse {
		strategy: string;
		gateways: GatewayLoadBalance[];
	}

	let gateways = $state<GatewayLoadBalance[]>([]);
	let strategy = $state('Weighted Round Robin');
	let loading = $state(true);

	const exportColumns: ExportColumn<GatewayLoadBalance>[] = [
		{ header: 'ID', getValue: (g) => g.id },
		{ header: 'Active Gateway', getValue: (g) => g.name },
		{ header: 'Driver', getValue: (g) => g.driver.toUpperCase() },
		{ header: 'Capacity Cap (TPS)', getValue: (g) => `${g.requests_per_second} TPS` },
		{ header: 'Base Priority', getValue: (g) => `P${g.priority}` },
		{ header: 'Supports Bulk Parallelism', getValue: (g) => g.supports_bulk ? 'YES' : 'NO' },
		{ header: 'Connection Timeout', getValue: (g) => `${g.timeout_seconds}s` }
	];

	async function loadGateways() {
		loading = true;
		try {
			const res = await supermanApi.get<{ data: LoadBalancingResponse }>('/api/v1/superman/routing/load-balancing');
			gateways = res.data?.gateways ?? [];
			if (res.data?.strategy) {
				strategy = res.data.strategy;
			}
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		loadGateways();
	});
</script>

<div class="space-y-6">
	<!-- Page Header -->
	<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
		<div>
			<h1 class="text-xl font-bold text-gray-900 dark:text-white">Gateway Load Balancing & Traffic Ratios</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
				Real-time capacity caps, concurrency limits, and priority weighting for enabled interconnect gateways.
			</p>
		</div>
		<div class="flex items-center gap-3">
			<button
				onclick={loadGateways}
				class="inline-flex items-center gap-1.5 px-3 py-2 text-xs font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 dark:bg-gray-800 dark:text-gray-300 dark:border-gray-700 dark:hover:bg-gray-700/60 shadow-xs transition"
			>
				<svg class="w-3.5 h-3.5 {loading ? 'animate-spin' : ''}" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
				</svg>
				Refresh
			</button>
			<ExportDropdown
				title="Gateway Load Balancing"
				columns={exportColumns}
				getData={() => gateways}
				disabled={gateways.length === 0}
			/>
		</div>
	</div>

	<!-- Strategy Banner -->
	<div class="bg-emerald-50/60 dark:bg-emerald-950/20 border border-emerald-200/60 dark:border-emerald-800/50 rounded-xl p-4 flex items-center justify-between">
		<div class="flex items-center gap-3">
			<div class="w-8 h-8 rounded-lg bg-emerald-600 text-white flex items-center justify-center font-bold">
				<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 6l3 1m0 0l-3 9a5.002 5.002 0 006.001 0M6 7l3 9M6 7l6-2m6 2l3-1m-3 1l-3 9a5.002 5.002 0 006.001 0M18 7l3 9m-3-9l-6-2m0-2v2m0 16V5m0 16H9m3 0h3" />
				</svg>
			</div>
			<div>
				<div class="text-xs font-bold text-gray-900 dark:text-white">Active Distribution Strategy: {strategy}</div>
				<div class="text-[11px] text-gray-500 dark:text-gray-400">Total aggregate bandwidth: {gateways.reduce((sum, g) => sum + g.requests_per_second, 0)} TPS across {gateways.length} active gateways</div>
			</div>
		</div>
	</div>

	<!-- Table -->
	<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl shadow-xs overflow-hidden">
		<div class="overflow-x-auto">
			<table class="w-full text-left text-xs">
				<thead class="bg-gray-50/75 dark:bg-gray-800/50 text-gray-500 uppercase tracking-wider font-semibold border-b border-gray-200 dark:border-gray-800">
					<tr>
						<th class="px-5 py-3">Active Gateway</th>
						<th class="px-5 py-3 text-right">Capacity Cap (TPS)</th>
						<th class="px-5 py-3 text-center">Base Priority</th>
						<th class="px-5 py-3 text-center">Supports Bulk Parallelism</th>
						<th class="px-5 py-3 text-right">Connection Timeout</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800 text-gray-700 dark:text-gray-300">
					{#if loading}
						<tr>
							<td colspan="5" class="px-5 py-10 text-center text-gray-400">Loading gateway load balance telemetry...</td>
						</tr>
					{:else if gateways.length === 0}
						<tr>
							<td colspan="5" class="px-5 py-10 text-center text-gray-400">No active gateways enabled.</td>
						</tr>
					{:else}
						{#each gateways as g}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30 transition">
								<td class="px-5 py-3.5">
									<div class="font-bold text-gray-900 dark:text-white">{g.name}</div>
									<div class="text-[11px] text-gray-400 uppercase font-mono">{g.driver} driver</div>
								</td>
								<td class="px-5 py-3.5 text-right font-mono font-bold text-emerald-600 dark:text-emerald-400">
									{g.requests_per_second} TPS
								</td>
								<td class="px-5 py-3.5 text-center">
									<span class="inline-flex items-center px-2 py-0.5 rounded-full text-[11px] font-semibold bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300 font-mono">
										P{g.priority}
									</span>
								</td>
								<td class="px-5 py-3.5 text-center">
									{#if g.supports_bulk}
										<span class="inline-flex items-center text-emerald-600 dark:text-emerald-400">
											<svg class="w-5 h-5 mx-auto" fill="none" viewBox="0 0 24 24" stroke="currentColor">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7" />
											</svg>
										</span>
									{:else}
										<span class="inline-flex items-center text-gray-300 dark:text-gray-600">
											<svg class="w-5 h-5 mx-auto" fill="none" viewBox="0 0 24 24" stroke="currentColor">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
											</svg>
										</span>
									{/if}
								</td>
								<td class="px-5 py-3.5 text-right font-mono text-gray-600 dark:text-gray-400">
									{g.timeout_seconds}s
								</td>
							</tr>
						{/each}
					{/if}
				</tbody>
			</table>
		</div>
	</div>
</div>
