<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface ProviderItem {
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

	let providers = $state<ProviderItem[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);

	async function loadProviders() {
		loading = true;
		error = null;
		try {
			const res = await supermanApi.get<{ data: { gateways: ProviderItem[] } }>('/api/v1/superman/providers/health');
			providers = res.data?.gateways ?? [];
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Failed to load provider reports.';
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		void loadProviders();
	});

	const exportColumns: ExportColumn<ProviderItem>[] = [
		{ header: 'Provider Gateway', getValue: (p) => p.name },
		{ header: 'Protocol / Driver', getValue: (p) => p.driver.toUpperCase() },
		{ header: 'Health State', getValue: (p) => p.health_status.toUpperCase() },
		{ header: 'Dispatches Processed', getValue: (p) => `${p.total_dispatches} msgs` },
		{ header: 'Success Ratio', getValue: (p) => `${p.success_rate_percent}%` },
		{ header: 'Round-Trip Latency', getValue: (p) => `${p.latency_ms} ms` }
	];
</script>

<svelte:head>
	<title>Provider Reports — Superman Administrator</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">Provider & Carrier Reports</h1>
			<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
				Gateway transmission volumes, delivery success ratios, and operating latencies across carriers.
			</p>
		</div>
		<div class="flex items-center gap-2">
			<ExportDropdown
				title="Provider and Carrier Reports"
				columns={exportColumns}
				getData={() => providers}
				disabled={providers.length === 0}
			/>
			<button onclick={loadProviders} class="rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs font-medium text-gray-700 shadow-xs hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300">
				Refresh
			</button>
		</div>
	</div>

	<div class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-xs dark:border-gray-800 dark:bg-gray-900">
		{#if loading}
			<div class="py-16 text-center text-xs text-gray-500">Loading provider reports...</div>
		{:else if error}
			<div class="p-6 text-center text-xs text-rose-600">{error}</div>
		{:else}
			<table class="w-full text-left text-xs">
				<thead>
					<tr class="border-b border-gray-100 bg-gray-50/50 text-[10px] font-semibold uppercase tracking-wider text-gray-500 dark:border-gray-800 dark:bg-gray-800/50 dark:text-gray-400">
						<th class="px-5 py-3.5">Provider Gateway</th>
						<th class="px-5 py-3.5">Protocol / Driver</th>
						<th class="px-5 py-3.5">Health State</th>
						<th class="px-5 py-3.5">Dispatches Processed</th>
						<th class="px-5 py-3.5">Success Ratio</th>
						<th class="px-5 py-3.5 text-right">Round-Trip Latency</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
					{#each providers as p}
						<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30">
							<td class="px-5 py-4 font-semibold text-gray-950 dark:text-white">{p.name}</td>
							<td class="px-5 py-4 font-mono uppercase text-blue-600">{p.driver}</td>
							<td class="px-5 py-4 font-bold uppercase text-emerald-600">{p.health_status}</td>
							<td class="px-5 py-4 font-semibold text-gray-900 dark:text-white">{p.total_dispatches} msgs</td>
							<td class="px-5 py-4 font-bold text-emerald-600">{p.success_rate_percent}%</td>
							<td class="px-5 py-4 text-right font-mono font-medium text-gray-700 dark:text-gray-300">{p.latency_ms} ms</td>
						</tr>
					{/each}
				</tbody>
			</table>
		{/if}
	</div>
</div>

