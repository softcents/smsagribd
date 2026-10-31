<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface BatchItem {
		id: number;
		customer: string;
		sender_id: string;
		source: string;
		status: string;
		approval: string;
		total_recipients: number;
		remaining: number;
		progress_percent: number;
		created_at: string;
	}

	let batches = $state<BatchItem[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);

	async function loadBatches() {
		loading = true;
		error = null;
		try {
			const res = await supermanApi.get<{ data: BatchItem[] }>('/api/v1/superman/batches?per_page=100');
			batches = res.data ?? [];
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Failed to load message batches.';
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		void loadBatches();
	});

	const exportColumns: ExportColumn<BatchItem>[] = [
		{ header: 'Batch ID', getValue: (b) => `#${b.id}` },
		{ header: 'Customer', getValue: (b) => b.customer },
		{ header: 'Sender ID', getValue: (b) => b.sender_id },
		{ header: 'Source', getValue: (b) => b.source.toUpperCase() },
		{ header: 'Total Recipients', getValue: (b) => b.total_recipients },
		{ header: 'Progress', getValue: (b) => `${b.progress_percent}%` },
		{ header: 'Approval', getValue: (b) => b.approval.toUpperCase() },
		{ header: 'Status', getValue: (b) => b.status.toUpperCase() },
		{ header: 'Created Date', getValue: (b) => new Date(b.created_at).toLocaleString() }
	];
</script>

<svelte:head>
	<title>Batches — Superman Administrator</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">Bulk Message Processing Batches</h1>
			<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
				Monitor asynchronous bulk SMS jobs, recipient progress bars, and campaign dispatch health.
			</p>
		</div>
		<div class="flex items-center gap-2">
			<button onclick={loadBatches} class="rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs font-medium text-gray-700 shadow-xs hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300">
				Refresh Batches
			</button>
			<ExportDropdown
				title="Message Batches"
				columns={exportColumns}
				getData={() => batches}
				disabled={batches.length === 0}
			/>
		</div>
	</div>

	<div class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-xs dark:border-gray-800 dark:bg-gray-900">
		{#if loading}
			<div class="py-16 text-center text-xs text-gray-500">Loading batch queues...</div>
		{:else if error}
			<div class="p-6 text-center text-xs text-rose-600">{error}</div>
		{:else if batches.length === 0}
			<div class="py-16 text-center text-xs text-gray-500">No active or historical batches recorded.</div>
		{:else}
			<div class="overflow-x-auto">
				<table class="w-full text-left text-xs">
					<thead>
						<tr class="border-b border-gray-100 bg-gray-50/50 text-[10px] font-semibold uppercase tracking-wider text-gray-500 dark:border-gray-800 dark:bg-gray-800/50 dark:text-gray-400">
							<th class="px-4 py-3">Customer</th>
							<th class="px-4 py-3">Sender ID</th>
							<th class="px-4 py-3">Source Channel</th>
							<th class="px-4 py-3">Recipients</th>
							<th class="px-4 py-3">Progress</th>
							<th class="px-4 py-3">Approval Status</th>
							<th class="px-4 py-3 text-right">Created</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
						{#each batches as b}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30">
								<td class="px-4 py-3 font-semibold text-gray-950 dark:text-white">{b.customer}</td>
								<td class="px-4 py-3 font-mono font-medium">{b.sender_id || 'Default'}</td>
								<td class="px-4 py-3 uppercase text-[11px] text-gray-500">{b.source || 'Web Portal'}</td>
								<td class="px-4 py-3 font-medium text-gray-900 dark:text-white">{b.total_recipients} msgs</td>
								<td class="px-4 py-3 w-44">
									<div class="flex items-center gap-2">
										<div class="h-2 w-full rounded-full bg-gray-100 overflow-hidden dark:bg-gray-800">
											<div class="h-full bg-emerald-500 rounded-full" style="width: {b.progress_percent || 100}%"></div>
										</div>
										<span class="text-[10px] font-bold text-gray-600 dark:text-gray-400">{b.progress_percent || 100}%</span>
									</div>
								</td>
								<td class="px-4 py-3">
									<span class="inline-flex rounded-full px-2 py-0.5 text-[10px] font-bold uppercase {b.approval === 'approved' ? 'bg-emerald-50 text-emerald-700' : 'bg-amber-50 text-amber-700'}">
										{b.approval || 'Approved'}
									</span>
								</td>
								<td class="px-4 py-3 text-right text-gray-400">{new Date(b.created_at).toLocaleString()}</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</div>
</div>

