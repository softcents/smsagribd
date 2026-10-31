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

	interface Meta {
		page: number;
		per_page: number;
		total: number;
	}

	let logs = $state<ApiLog[]>([]);
	let meta = $state<Meta>({ page: 1, per_page: 25, total: 0 });
	let loading = $state(true);
	let error = $state<string | null>(null);
	let search = $state('');

	async function loadLogs(targetPage = 1) {
		loading = true;
		error = null;
		try {
			const params = new URLSearchParams({
				page: String(targetPage),
				per_page: String(meta.per_page)
			});
			if (search.trim()) params.set('search', search.trim());

			const res = await supermanApi.get<{ data: ApiLog[]; meta: Meta }>(
				`/api/v1/superman/providers/api-logs?${params.toString()}`
			);
			logs = res.data ?? [];
			if (res.meta) meta = res.meta;
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Failed to load transmission logs.';
		} finally {
			loading = false;
		}
	}

	function handleSearch(e: Event) {
		e.preventDefault();
		void loadLogs(1);
	}

	onMount(() => {
		void loadLogs(1);
	});

	let totalPages = $derived(Math.ceil((meta.total || 1) / meta.per_page));

	const exportColumns: ExportColumn<ApiLog>[] = [
		{ header: 'Log ID', getValue: (l) => `#${l.id}` },
		{ header: 'Provider / Gateway', getValue: (l) => l.gateway_name },
		{ header: 'Event', getValue: (l) => l.event },
		{ header: 'Recipient', getValue: (l) => l.phone },
		{ header: 'HTTP Status', getValue: (l) => l.status_code },
		{ header: 'Outcome', getValue: (l) => l.status.toUpperCase() },
		{ header: 'Latency', getValue: (l) => `${l.latency_ms} ms` },
		{ header: 'Cost (BDT)', getValue: (l) => `BDT ${Number(l.cost).toFixed(4)}` },
		{ header: 'Timestamp', getValue: (l) => new Date(l.created_at).toLocaleString() }
	];
</script>

<svelte:head>
	<title>Provider API Logs — Superman Administrator</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">Provider Transmission & API Logs</h1>
			<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
				Audit upstream SMS submission requests, HTTP response status codes, and telco round-trip latencies.
			</p>
		</div>
		<div class="flex items-center gap-2">
			<button
				onclick={() => loadLogs(meta.page)}
				class="inline-flex items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs font-medium text-gray-700 shadow-xs hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300 dark:hover:bg-gray-800"
			>
				Refresh Logs
			</button>
			<ExportDropdown
				title="Provider API Logs"
				columns={exportColumns}
				getData={() => logs}
				disabled={logs.length === 0}
			/>
		</div>
	</div>

	<!-- Filter / Search -->
	<form onsubmit={handleSearch} class="flex items-center gap-3 rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
		<input
			type="search"
			bind:value={search}
			placeholder="Search by recipient MSISDN..."
			class="w-full max-w-sm rounded-lg border border-gray-200 bg-gray-50/70 px-3 py-2 text-xs outline-none focus:border-emerald-500 focus:bg-white dark:border-gray-700 dark:bg-gray-800 dark:text-white"
		/>
		<button type="submit" class="rounded-lg bg-emerald-600 px-4 py-2 text-xs font-semibold text-white shadow-xs hover:bg-emerald-700">
			Filter
		</button>
	</form>

	<!-- Table -->
	<div class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-xs dark:border-gray-800 dark:bg-gray-900">
		{#if loading}
			<div class="flex items-center justify-center py-16 text-xs text-gray-500">
				<svg class="mr-2 h-4 w-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path></svg>
				Loading provider logs...
			</div>
		{:else if error}
			<div class="p-6 text-center text-xs text-rose-600">{error}</div>
		{:else if logs.length === 0}
			<div class="py-12 text-center text-xs text-gray-500">No transmission logs recorded yet.</div>
		{:else}
			<div class="overflow-x-auto">
				<table class="w-full text-left text-xs">
					<thead>
						<tr class="border-b border-gray-100 bg-gray-50/50 text-[10px] font-semibold uppercase tracking-wider text-gray-500 dark:border-gray-800 dark:bg-gray-800/50 dark:text-gray-400">
							<th class="px-4 py-3">Timestamp</th>
							<th class="px-4 py-3">Supplier / Gateway</th>
							<th class="px-4 py-3">Event Type</th>
							<th class="px-4 py-3">Recipient MSISDN</th>
							<th class="px-4 py-3">Status Code</th>
							<th class="px-4 py-3">Latency</th>
							<th class="px-4 py-3 text-right">Cost</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
						{#each logs as item}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30">
								<td class="px-4 py-3 text-gray-500">
									{new Date(item.created_at).toLocaleString()}
								</td>
								<td class="px-4 py-3 font-semibold text-gray-900 dark:text-white">
									{item.gateway_name}
								</td>
								<td class="px-4 py-3">
									<span class="rounded bg-gray-100 px-2 py-0.5 text-[10px] font-mono font-medium text-gray-700 dark:bg-gray-800 dark:text-gray-300">
										{item.event}
									</span>
								</td>
								<td class="px-4 py-3 font-mono text-gray-800 dark:text-gray-200">
									{item.phone}
								</td>
								<td class="px-4 py-3">
									<span class="inline-flex rounded px-1.5 py-0.5 text-[10px] font-bold {item.status_code === 200 || item.status_code === 202 ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-400' : 'bg-rose-50 text-rose-700 dark:bg-rose-950/40 dark:text-rose-400'}">
										{item.status_code}
									</span>
								</td>
								<td class="px-4 py-3 font-mono text-gray-600 dark:text-gray-400">
									{item.latency_ms} ms
								</td>
								<td class="px-4 py-3 text-right font-medium text-gray-900 dark:text-white">
									৳{Number(item.cost).toFixed(4)}
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>

			<!-- Pagination -->
			{#if totalPages > 1}
				<div class="flex items-center justify-between border-t border-gray-100 px-4 py-3 text-xs text-gray-500 dark:border-gray-800">
					<div>
						Showing {(meta.page - 1) * meta.per_page + 1}–{Math.min(meta.page * meta.per_page, meta.total)} of {meta.total} records
					</div>
					<div class="flex items-center gap-1">
						<button
							onclick={() => loadLogs(meta.page - 1)}
							disabled={meta.page <= 1}
							class="rounded border border-gray-200 px-2.5 py-1 text-xs hover:bg-gray-50 disabled:opacity-40 dark:border-gray-700 dark:hover:bg-gray-800"
						>
							Previous
						</button>
						<span class="px-2 text-xs font-semibold text-gray-700 dark:text-gray-300">
							Page {meta.page} of {totalPages}
						</span>
						<button
							onclick={() => loadLogs(meta.page + 1)}
							disabled={meta.page >= totalPages}
							class="rounded border border-gray-200 px-2.5 py-1 text-xs hover:bg-gray-50 disabled:opacity-40 dark:border-gray-700 dark:hover:bg-gray-800"
						>
							Next
						</button>
					</div>
				</div>
			{/if}
		{/if}
	</div>
</div>

