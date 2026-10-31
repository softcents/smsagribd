<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface BatchApprovalItem {
		id: number;
		customer: string;
		sender_id: string;
		total_recipients: number;
		status: string;
		created_at: string;
	}

	let batches = $state<BatchApprovalItem[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);

	async function loadApprovals() {
		loading = true;
		error = null;
		try {
			const res = await supermanApi.get<{ data: BatchApprovalItem[] }>('/api/v1/superman/batches?approval=pending');
			batches = res.data ?? [];
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Failed to load approvals queue.';
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		void loadApprovals();
	});

	const exportColumns: ExportColumn<BatchApprovalItem>[] = [
		{ header: 'Batch ID', getValue: (b) => `#${b.id}` },
		{ header: 'Customer', getValue: (b) => b.customer },
		{ header: 'Sender ID', getValue: (b) => b.sender_id },
		{ header: 'Recipients Count', getValue: (b) => b.total_recipients },
		{ header: 'Status', getValue: (b) => b.status.toUpperCase() },
		{ header: 'Submitted Date', getValue: (b) => (b.created_at ? new Date(b.created_at).toLocaleString() : '—') }
	];
</script>

<svelte:head>
	<title>Messaging Approvals — Superman Administrator</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">Message & Campaign Approvals</h1>
			<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
				Review high-volume promotional batches and outbound campaigns waiting for manual compliance sign-off.
			</p>
		</div>
		<div class="flex items-center gap-2">
			<button onclick={loadApprovals} class="rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs font-medium text-gray-700 shadow-xs hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300">
				Refresh Queue
			</button>
			<ExportDropdown
				title="Pending Messaging Approvals"
				columns={exportColumns}
				getData={() => batches}
				disabled={batches.length === 0}
			/>
		</div>
	</div>

	<div class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-xs dark:border-gray-800 dark:bg-gray-900">
		{#if loading}
			<div class="py-16 text-center text-xs text-gray-500">Scanning pending approval queue...</div>
		{:else if error}
			<div class="p-6 text-center text-xs text-rose-600">{error}</div>
		{:else if batches.length === 0}
			<div class="py-16 text-center text-xs text-gray-500">
				No campaigns or batches pending review. All dispatches approved.
			</div>
		{:else}
			<div class="overflow-x-auto">
				<table class="w-full text-left text-xs">
					<thead>
						<tr class="border-b border-gray-100 bg-gray-50/50 text-[10px] font-semibold uppercase tracking-wider text-gray-500 dark:border-gray-800 dark:bg-gray-800/50 dark:text-gray-400">
							<th class="px-4 py-3">Batch ID</th>
							<th class="px-4 py-3">Customer Account</th>
							<th class="px-4 py-3">Sender ID</th>
							<th class="px-4 py-3">Audience Count</th>
							<th class="px-4 py-3">Current State</th>
							<th class="px-4 py-3 text-right">Submitted</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
						{#each batches as b}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30">
								<td class="px-4 py-3 font-mono font-bold text-gray-900 dark:text-white">BATCH-{b.id}</td>
								<td class="px-4 py-3 font-semibold text-gray-900 dark:text-white">{b.customer}</td>
								<td class="px-4 py-3 font-mono font-medium text-blue-600">{b.sender_id || 'Default'}</td>
								<td class="px-4 py-3 font-bold text-gray-900 dark:text-white">{b.total_recipients} Recipients</td>
								<td class="px-4 py-3">
									<span class="inline-flex rounded-full bg-amber-50 px-2.5 py-0.5 text-[10px] font-bold uppercase text-amber-700">
										{b.status || 'Pending'}
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

