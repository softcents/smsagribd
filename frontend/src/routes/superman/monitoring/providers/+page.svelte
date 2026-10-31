<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface GatewayItem {
		id: number;
		name: string;
		driver: string;
		health_status: string;
		priority: number;
		requests_per_second: number;
		latency_ms: number;
		total_dispatches: number;
		success_rate_percent: number;
	}

	interface HealthSummary {
		total_gateways: number;
		healthy_count: number;
		avg_latency_ms: number;
		gateways: GatewayItem[];
	}

	let summary = $state<HealthSummary | null>(null);
	let loading = $state(true);

	onMount(async () => {
		try {
			const res = await supermanApi.get<{ data: HealthSummary }>('/api/v1/superman/monitoring/providers');
			summary = res.data ?? null;
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	});

	const exportColumns: ExportColumn<GatewayItem>[] = [
		{ header: 'Carrier / Gateway', getValue: (g) => g.name },
		{ header: 'Driver', getValue: (g) => g.driver.toUpperCase() },
		{ header: 'Priority', getValue: (g) => `Pri ${g.priority}` },
		{ header: 'Latency', getValue: (g) => `${g.latency_ms} ms` },
		{ header: 'Throughput Limit', getValue: (g) => `${g.requests_per_second} req/s` },
		{ header: 'Success Rate', getValue: (g) => `${g.success_rate_percent}%` },
		{ header: 'Health State', getValue: (g) => g.health_status }
	];
</script>

<div class="space-y-6">
	<!-- Page Header -->
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-xl font-bold text-gray-900 dark:text-white">Provider Telemetry & Gateway Interconnect Monitor</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
				Real-time carrier latency monitoring, success rate ratios, and automated upstream connectivity telemetry.
			</p>
		</div>
		<div class="flex items-center gap-2">
			<ExportDropdown
				title="Provider Interconnect Monitor"
				columns={exportColumns}
				getData={() => summary?.gateways || []}
				disabled={!summary || summary.gateways.length === 0}
			/>
			<a
				href="/superman/providers/health"
				class="inline-flex items-center gap-2 rounded-lg bg-emerald-600 px-3 py-1.5 text-xs font-semibold text-white shadow-xs hover:bg-emerald-700 transition"
			>
				Live Health Checker
			</a>
		</div>
	</div>

	{#if loading}
		<div class="p-8 text-center text-xs text-gray-400">Loading gateway telemetry...</div>
	{:else if summary}
		<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
			<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-5 shadow-xs">
				<span class="text-xs font-medium text-gray-500 uppercase tracking-wider">Monitored Gateways</span>
				<p class="text-3xl font-bold text-gray-900 dark:text-white mt-1">{summary.total_gateways}</p>
				<span class="text-[11px] text-emerald-600 mt-1 block">{summary.healthy_count} currently healthy</span>
			</div>
			<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-5 shadow-xs">
				<span class="text-xs font-medium text-emerald-600 uppercase tracking-wider">Average Round-Trip Latency</span>
				<p class="text-3xl font-bold text-emerald-600 mt-1">{summary.avg_latency_ms} ms</p>
				<span class="text-[11px] text-gray-400 mt-1 block">HTTP REST socket ACK</span>
			</div>
			<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-5 shadow-xs">
				<span class="text-xs font-medium text-blue-600 uppercase tracking-wider">Network Interconnect</span>
				<p class="text-3xl font-bold text-blue-600 mt-1">Operational</p>
				<span class="text-[11px] text-gray-400 mt-1 block">Zero active packet drops</span>
			</div>
		</div>

		<!-- Gateways Table -->
		<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl shadow-xs overflow-hidden">
			<div class="overflow-x-auto">
				<table class="w-full text-left text-xs">
					<thead class="bg-gray-50/75 dark:bg-gray-800/50 text-gray-500 uppercase tracking-wider font-semibold border-b border-gray-200 dark:border-gray-800">
						<tr>
							<th class="px-5 py-3">Carrier / Gateway</th>
							<th class="px-5 py-3">Driver</th>
							<th class="px-5 py-3">Priority</th>
							<th class="px-5 py-3">Latency</th>
							<th class="px-5 py-3">Throughput Limit</th>
							<th class="px-5 py-3">Success Rate</th>
							<th class="px-5 py-3">State</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100 dark:divide-gray-800 text-gray-700 dark:text-gray-300">
						{#each summary.gateways as g}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30 transition">
								<td class="px-5 py-3 font-bold text-gray-900 dark:text-white">{g.name}</td>
								<td class="px-5 py-3 font-mono text-gray-500">{g.driver}</td>
								<td class="px-5 py-3">Pri {g.priority}</td>
								<td class="px-5 py-3 font-mono font-semibold text-emerald-600 dark:text-emerald-400">{g.latency_ms} ms</td>
								<td class="px-5 py-3">{g.requests_per_second} req/s</td>
								<td class="px-5 py-3 font-bold text-gray-900 dark:text-white">{g.success_rate_percent}%</td>
								<td class="px-5 py-3">
									<span class="inline-flex items-center gap-1.5 text-xs text-emerald-600 font-medium">
										<span class="h-1.5 w-1.5 rounded-full bg-emerald-600"></span>
										{g.health_status}
									</span>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>
	{/if}
</div>

