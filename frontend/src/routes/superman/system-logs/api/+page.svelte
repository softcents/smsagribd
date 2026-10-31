<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface ApiLog {
		id: number;
		gateway_name: string;
		event: string;
		phone: string;
		status: string;
		status_code: number;
		latency_ms: number;
		cost: number;
		created_at: string;
	}

	let logs = $state<ApiLog[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);

	const exportColumns: ExportColumn<ApiLog>[] = [
		{ header: 'ID', getValue: (l) => l.id },
		{ header: 'Gateway / Route', getValue: (l) => l.gateway_name },
		{ header: 'Event', getValue: (l) => l.event },
		{ header: 'Phone / Recipient', getValue: (l) => l.phone },
		{ header: 'Status', getValue: (l) => l.status },
		{ header: 'HTTP Code', getValue: (l) => l.status_code },
		{ header: 'Latency (ms)', getValue: (l) => l.latency_ms },
		{ header: 'Cost (BDT)', getValue: (l) => `BDT ${l.cost.toFixed(4)}` },
		{ header: 'Timestamp', getValue: (l) => l.created_at }
	];

	async function loadLogs() {
		loading = true;
		error = null;
		try {
			const res = await supermanApi.get<{ data: ApiLog[] }>('/api/v1/superman/providers/api-logs?per_page=50');
			logs = res.data ?? [];
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Failed to load API logs.';
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		void loadLogs();
	});
</script>

<svelte:head>
	<title>API Traffic Logs — Superman Administrator</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">API Traffic & Request Logs</h1>
			<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
				Live inbound customer API requests, HTTP status codes, and execution latencies.
			</p>
		</div>
		<div class="flex items-center gap-2">
			<ExportDropdown
				title="API Traffic and Request Logs"
				columns={exportColumns}
				getData={() => logs}
				disabled={logs.length === 0}
			/>
			<button onclick={loadLogs} class="rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs font-medium text-gray-700 shadow-xs hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300">
				Refresh Logs
			</button>
		</div>
	</div>

	<div class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-xs dark:border-gray-800 dark:bg-gray-900">
		{#if loading}
			<div class="py-16 text-center text-xs text-gray-500">Loading API requests...</div>
		{:else if error}
			<div class="p-6 text-center text-xs text-rose-600">{error}</div>
		{:else if logs.length === 0}
			<div class="py-16 text-center text-xs text-gray-500">No external API logs recorded yet.</div>
		{:else}
			<div class="overflow-x-auto">
				<table class="w-full text-left text-xs">
					<thead>
						<tr class="border-b border-gray-100 bg-gray-50/50 text-[10px] font-semibold uppercase tracking-wider text-gray-500 dark:border-gray-800 dark:bg-gray-800/50 dark:text-gray-400">
							<th class="px-5 py-3.5">Endpoint / Event</th>
							<th class="px-5 py-3.5">Recipient MSISDN</th>
							<th class="px-5 py-3.5">HTTP Code</th>
							<th class="px-5 py-3.5">Response Time</th>
							<th class="px-5 py-3.5 text-right">Timestamp</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
						{#each logs as l}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30">
								<td class="px-5 py-4 font-mono font-medium text-gray-900 dark:text-white">POST /api/v1/sms/send</td>
								<td class="px-5 py-4 font-mono text-gray-700 dark:text-gray-300">{l.phone}</td>
								<td class="px-5 py-4">
									<span class="inline-flex rounded px-2 py-0.5 text-[10px] font-bold {l.status_code === 200 || l.status_code === 202 ? 'bg-emerald-50 text-emerald-700' : 'bg-rose-50 text-rose-700'}">
										{l.status_code}
									</span>
								</td>
								<td class="px-5 py-4 font-mono text-gray-500">{l.latency_ms} ms</td>
								<td class="px-5 py-4 text-right text-gray-400">{new Date(l.created_at).toLocaleString()}</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</div>
</div>

