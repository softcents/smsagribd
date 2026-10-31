<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface GatewayHealthItem {
		id: number;
		name: string;
		driver: string;
		health_status: string;
		priority: number;
		requests_per_second: number;
		latency_ms: number;
		last_checked_at?: string;
		is_enabled: boolean;
		is_otp_enabled: boolean;
		total_dispatches: number;
		success_rate_percent: number;
	}

	interface HealthSummary {
		total_gateways: number;
		healthy_count: number;
		degraded_count: number;
		down_count: number;
		avg_latency_ms: number;
		gateways: GatewayHealthItem[];
	}

	let summary = $state<HealthSummary | null>(null);
	let loading = $state(true);
	let checking = $state(false);
	let error = $state<string | null>(null);
	let successMsg = $state<string | null>(null);

	async function loadHealth() {
		loading = true;
		error = null;
		try {
			const res = await supermanApi.get<{ data: HealthSummary }>('/api/v1/superman/providers/health');
			summary = res.data;
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Failed to load provider health data.';
		} finally {
			loading = false;
		}
	}

	async function checkAll() {
		checking = true;
		successMsg = null;
		try {
			await supermanApi.post('/api/v1/superman/providers/health/check-all', {});
			successMsg = 'Live health check triggered successfully across all gateways.';
			await loadHealth();
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Failed to trigger health check.';
		} finally {
			checking = false;
		}
	}

	onMount(() => {
		void loadHealth();
	});

	const exportColumns: ExportColumn<GatewayHealthItem>[] = [
		{ header: 'Provider / Gateway', getValue: (g) => g.name },
		{ header: 'Driver', getValue: (g) => g.driver.toUpperCase() },
		{ header: 'Health Status', getValue: (g) => g.health_status.toUpperCase() },
		{ header: 'Priority', getValue: (g) => g.priority },
		{ header: 'Capacity (RPS)', getValue: (g) => g.requests_per_second },
		{ header: 'Latency (ms)', getValue: (g) => `${g.latency_ms} ms` },
		{ header: 'Success Rate', getValue: (g) => `${g.success_rate_percent.toFixed(1)}%` },
		{ header: 'Total Dispatches', getValue: (g) => g.total_dispatches },
		{ header: 'Operating State', getValue: (g) => (g.is_enabled ? 'Enabled' : 'Disabled') }
	];
</script>

<svelte:head>
	<title>Provider Health Monitor — Superman Administrator</title>
</svelte:head>

<div class="space-y-6">
	<!-- Header -->
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">Provider & Gateway Health Monitor</h1>
			<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
				Monitor live HTTP latency, network reachability, and delivery success rates across all configured SMS gateways.
			</p>
		</div>
		<div class="flex items-center gap-2">
			<button
				onclick={loadHealth}
				disabled={loading || checking}
				class="inline-flex items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs font-medium text-gray-700 shadow-xs hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300 dark:hover:bg-gray-800 disabled:opacity-60"
			>
				Refresh
			</button>
			<ExportDropdown
				title="Provider Health Monitor"
				columns={exportColumns}
				getData={() => summary?.gateways || []}
				disabled={!summary || summary.gateways.length === 0}
			/>
			<button
				onclick={checkAll}
				disabled={checking}
				class="inline-flex items-center gap-1.5 rounded-lg bg-emerald-600 px-3.5 py-2 text-xs font-semibold text-white shadow-xs hover:bg-emerald-700 disabled:opacity-60"
			>
				<svg class="h-4 w-4 {checking ? 'animate-spin' : ''}" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0 3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182m0-4.991v4.99" />
				</svg>
				{checking ? 'Checking All...' : 'Check All Providers'}
			</button>
		</div>
	</div>

	<!-- Information Banner -->
	<div class="rounded-xl border border-gray-200 bg-white p-5 shadow-xs dark:border-gray-800 dark:bg-gray-900">
		<h3 class="text-sm font-bold text-gray-900 dark:text-white">Provider Status & Real-time Connectivity</h3>
		<p class="mt-1 text-xs text-gray-600 dark:text-gray-400">
			Monitor HTTP latency, network reachability, and credit balances across all configured SMS gateways.
			Providers experiencing high latency or downtime will trigger automatic failovers when priority routing is enabled.
		</p>
	</div>

	{#if successMsg}
		<div class="rounded-xl bg-emerald-50 p-4 text-xs font-medium text-emerald-800 dark:bg-emerald-950/40 dark:text-emerald-300 border border-emerald-200 dark:border-emerald-900">
			{successMsg}
		</div>
	{/if}

	<!-- KPI Stat Cards -->
	{#if summary}
		<div class="grid grid-cols-2 gap-4 sm:grid-cols-4">
			<div class="rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
				<span class="text-[11px] font-medium text-gray-500 uppercase tracking-wider">Total Gateways</span>
				<p class="mt-1 text-2xl font-bold text-gray-950 dark:text-white">{summary.total_gateways}</p>
			</div>
			<div class="rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
				<span class="text-[11px] font-medium text-gray-500 uppercase tracking-wider">Healthy & Online</span>
				<p class="mt-1 text-2xl font-bold text-emerald-600">{summary.healthy_count}</p>
			</div>
			<div class="rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
				<span class="text-[11px] font-medium text-gray-500 uppercase tracking-wider">Degraded / Down</span>
				<p class="mt-1 text-2xl font-bold {summary.down_count > 0 ? 'text-rose-600' : 'text-gray-400'}">
					{summary.down_count}
				</p>
			</div>
			<div class="rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
				<span class="text-[11px] font-medium text-gray-500 uppercase tracking-wider">Average Latency</span>
				<p class="mt-1 text-2xl font-bold text-blue-600">{summary.avg_latency_ms} <span class="text-xs font-normal text-gray-500">ms</span></p>
			</div>
		</div>
	{/if}

	<!-- Gateways Health Matrix Table -->
	<div class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-xs dark:border-gray-800 dark:bg-gray-900">
		{#if loading}
			<div class="flex items-center justify-center py-16 text-xs text-gray-500">
				<svg class="mr-2 h-4 w-4 animate-spin" fill="none" viewBox="0 0 24 24">
					<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
					<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
				</svg>
				Verifying provider latency & status...
			</div>
		{:else if error}
			<div class="p-6 text-center text-xs text-rose-600">{error}</div>
		{:else if summary && summary.gateways.length > 0}
			<div class="overflow-x-auto">
				<table class="w-full text-left text-xs">
					<thead>
						<tr class="border-b border-gray-100 bg-gray-50/50 text-[10px] font-semibold uppercase tracking-wider text-gray-500 dark:border-gray-800 dark:bg-gray-800/50 dark:text-gray-400">
							<th class="px-4 py-3">Provider / Gateway</th>
							<th class="px-4 py-3">Driver</th>
							<th class="px-4 py-3">Health Status</th>
							<th class="px-4 py-3">Round-Trip Latency</th>
							<th class="px-4 py-3">Delivery Success</th>
							<th class="px-4 py-3">Capacity</th>
							<th class="px-4 py-3">Last Checked</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
						{#each summary.gateways as g}
							<tr class="transition-colors hover:bg-gray-50/50 dark:hover:bg-gray-800/30">
								<td class="px-4 py-3.5">
									<div class="font-semibold text-gray-950 dark:text-white">{g.name}</div>
									<div class="text-[11px] text-gray-400">Priority Tier #{g.priority}</div>
								</td>
								<td class="px-4 py-3.5">
									<span class="inline-flex rounded-md bg-blue-50 px-2 py-0.5 text-[10px] font-bold text-blue-700 uppercase dark:bg-blue-950/40 dark:text-blue-300">
										{g.driver}
									</span>
								</td>
								<td class="px-4 py-3.5">
									{#if g.health_status.toLowerCase() === 'healthy'}
										<span class="inline-flex items-center gap-1.5 rounded-full bg-emerald-50 px-2.5 py-1 text-[11px] font-semibold text-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-400">
											<span class="h-2 w-2 rounded-full bg-emerald-500 animate-pulse"></span>
											HEALTHY
										</span>
									{:else if g.health_status.toLowerCase() === 'degraded'}
										<span class="inline-flex items-center gap-1.5 rounded-full bg-amber-50 px-2.5 py-1 text-[11px] font-semibold text-amber-700 dark:bg-amber-950/40 dark:text-amber-400">
											<span class="h-2 w-2 rounded-full bg-amber-500"></span>
											DEGRADED
										</span>
									{:else}
										<span class="inline-flex items-center gap-1.5 rounded-full bg-rose-50 px-2.5 py-1 text-[11px] font-semibold text-rose-700 dark:bg-rose-950/40 dark:text-rose-400">
											<span class="h-2 w-2 rounded-full bg-rose-500"></span>
											DOWN
										</span>
									{/if}
								</td>
								<td class="px-4 py-3.5 font-mono text-xs">
									<span class="{g.latency_ms < 150 ? 'text-emerald-600 font-bold' : g.latency_ms < 400 ? 'text-amber-600' : 'text-rose-600'}">
										{g.latency_ms} ms
									</span>
								</td>
								<td class="px-4 py-3.5 font-medium text-gray-900 dark:text-white">
									{g.success_rate_percent}%
								</td>
								<td class="px-4 py-3.5 text-gray-600 dark:text-gray-400">
									{g.requests_per_second} req/s
								</td>
								<td class="px-4 py-3.5 text-gray-400">
									{g.last_checked_at ? new Date(g.last_checked_at).toLocaleTimeString() : 'Just now'}
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</div>
</div>

