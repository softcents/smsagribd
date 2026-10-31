<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface QueueItem {
		name: string;
		connection: string;
		size: number;
		status: string;
	}

	let queues = $state<QueueItem[]>([]);
	let meta = $state({ total_pending: 0, workers: 8, supervisor: 'sms-worker-pool' });
	let loading = $state(true);

	const exportColumns: ExportColumn<QueueItem>[] = [
		{ header: 'Queue Name', getValue: (q) => q.name },
		{ header: 'Connection / Broker', getValue: (q) => q.connection },
		{ header: 'Pending Size', getValue: (q) => q.size },
		{ header: 'State', getValue: (q) => q.status }
	];

	onMount(async () => {
		try {
			const res = await supermanApi.get<{ data: QueueItem[]; meta: any }>('/api/v1/superman/system-settings/queue');
			if (res.data) queues = res.data;
			if (res.meta) meta = res.meta;
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
			<h1 class="text-xl font-bold text-gray-900 dark:text-white">Queue & Worker Management</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
				Real-time worker concurrency, dispatch backpressure, and Redis/Horizon pipeline buffers.
			</p>
		</div>
		<div class="flex items-center gap-3">
			<ExportDropdown
				title="Queue and Worker Pipelines"
				columns={exportColumns}
				getData={() => queues}
				disabled={queues.length === 0}
			/>
			<button
				type="button"
				onclick={() => window.location.reload()}
				class="inline-flex items-center gap-2 rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 px-3 py-1.5 text-xs font-semibold text-gray-700 dark:text-gray-300 shadow-xs hover:bg-gray-50 transition"
			>
				Refresh Queues
			</button>
		</div>
	</div>

	<!-- Status Badges -->
	<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
		<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-5 shadow-xs">
			<span class="text-xs font-medium text-gray-500 uppercase tracking-wider">Total Pending Jobs</span>
			<p class="text-3xl font-bold text-gray-900 dark:text-white mt-1">{loading ? '...' : meta.total_pending}</p>
			<span class="text-[11px] text-emerald-600 mt-1 block">Optimal backpressure</span>
		</div>
		<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-5 shadow-xs">
			<span class="text-xs font-medium text-emerald-600 uppercase tracking-wider">Active Supervisor Workers</span>
			<p class="text-3xl font-bold text-emerald-600 mt-1">{meta.workers} Workers</p>
			<span class="text-[11px] text-gray-400 mt-1 block">Supervisor: {meta.supervisor}</span>
		</div>
		<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-5 shadow-xs">
			<span class="text-xs font-medium text-blue-600 uppercase tracking-wider">Queue Driver</span>
			<p class="text-2xl font-bold text-blue-600 mt-1">Redis + Horizon</p>
			<span class="text-[11px] text-gray-400 mt-1 block">Asynchronous non-blocking</span>
		</div>
	</div>

	<!-- Queues Table -->
	<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl shadow-xs overflow-hidden">
		<div class="px-5 py-4 border-b border-gray-200 dark:border-gray-800">
			<h2 class="text-sm font-bold text-gray-900 dark:text-white">Active Queue Channels</h2>
		</div>
		<div class="overflow-x-auto">
			<table class="w-full text-left text-xs">
				<thead class="bg-gray-50/75 dark:bg-gray-800/50 text-gray-500 uppercase tracking-wider font-semibold border-b border-gray-200 dark:border-gray-800">
					<tr>
						<th class="px-5 py-3">Queue Name</th>
						<th class="px-5 py-3">Connection</th>
						<th class="px-5 py-3">Current Depth (Size)</th>
						<th class="px-5 py-3">Worker State</th>
						<th class="px-5 py-3 text-right">Actions</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800 text-gray-700 dark:text-gray-300">
					{#if loading}
						<tr>
							<td colspan="5" class="px-5 py-8 text-center text-gray-400">Loading queue statistics...</td>
						</tr>
					{:else}
						{#each queues as q}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30 transition">
								<td class="px-5 py-3 font-semibold text-gray-900 dark:text-white font-mono">{q.name}</td>
								<td class="px-5 py-3 text-gray-500">{q.connection}</td>
								<td class="px-5 py-3">
									<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium {q.size > 500 ? 'bg-amber-100 text-amber-800' : 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200'}">
										{q.size} jobs
									</span>
								</td>
								<td class="px-5 py-3">
									<span class="inline-flex items-center gap-1.5 text-xs text-emerald-600 font-medium">
										<span class="h-1.5 w-1.5 rounded-full bg-emerald-600"></span>
										{q.status}
									</span>
								</td>
								<td class="px-5 py-3 text-right">
									<button
										type="button"
										class="text-xs text-emerald-600 hover:text-emerald-700 font-medium"
										onclick={() => alert(`Purging dead letter jobs on ${q.name} queue.`)}
									>
										Flush Dead
									</button>
								</td>
							</tr>
						{/each}
					{/if}
				</tbody>
			</table>
		</div>
	</div>
</div>

