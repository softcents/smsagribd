<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface PendingSender {
		id: number;
		sender_id: string;
		type: string;
		requested_by_name: string;
		approval_status: string;
		created_at: string;
	}

	let senders = $state<PendingSender[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let processingId = $state<number | null>(null);

	async function loadPending() {
		loading = true;
		error = null;
		try {
			const res = await supermanApi.get<{ data: PendingSender[] }>('/api/v1/superman/senders?status=pending');
			senders = res.data ?? [];
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Failed to load pending sender approvals.';
		} finally {
			loading = false;
		}
	}

	async function approveSender(id: number) {
		processingId = id;
		try {
			await supermanApi.post(`/api/v1/superman/senders/${id}/approve`, {});
			await loadPending();
		} catch (err: unknown) {
			alert(err instanceof Error ? err.message : 'Sender approval failed.');
		} finally {
			processingId = null;
		}
	}

	async function rejectSender(id: number) {
		processingId = id;
		try {
			await supermanApi.post(`/api/v1/superman/senders/${id}/reject`, {});
			await loadPending();
		} catch (err: unknown) {
			alert(err instanceof Error ? err.message : 'Sender rejection failed.');
		} finally {
			processingId = null;
		}
	}

	onMount(() => {
		void loadPending();
	});

	const exportColumns: ExportColumn<PendingSender>[] = [
		{ header: 'Sender ID', getValue: (s) => s.sender_id },
		{ header: 'Type', getValue: (s) => s.type },
		{ header: 'Requested By', getValue: (s) => s.requested_by_name || 'Direct / System' },
		{ header: 'Status', getValue: (s) => s.approval_status.toUpperCase() },
		{ header: 'Submitted Date', getValue: (s) => (s.created_at ? new Date(s.created_at).toLocaleString() : '—') }
	];
</script>

<svelte:head>
	<title>Pending Sender Approvals — Superman Administrator</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">Pending Sender ID Approvals</h1>
			<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
				Audit and approve brand masking sender headers in accordance with national telecommunication authority guidelines.
			</p>
		</div>
		<div class="flex items-center gap-2">
			<button onclick={loadPending} class="rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs font-medium text-gray-700 shadow-xs hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300">
				Refresh Senders
			</button>
			<ExportDropdown
				title="Pending Sender Approvals"
				columns={exportColumns}
				getData={() => senders}
				disabled={senders.length === 0}
			/>
		</div>
	</div>

	<div class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-xs dark:border-gray-800 dark:bg-gray-900">
		{#if loading}
			<div class="py-16 text-center text-xs text-gray-500">Checking pending sender requests...</div>
		{:else if error}
			<div class="p-6 text-center text-xs text-rose-600">{error}</div>
		{:else if senders.length === 0}
			<div class="py-16 text-center text-xs text-gray-500">
				No sender headers waiting for approval.
			</div>
		{:else}
			<table class="w-full text-left text-xs">
				<thead>
					<tr class="border-b border-gray-100 bg-gray-50/50 text-[10px] font-semibold uppercase tracking-wider text-gray-500 dark:border-gray-800 dark:bg-gray-800/50 dark:text-gray-400">
						<th class="px-5 py-3.5">Sender ID Header</th>
						<th class="px-5 py-3.5">Header Type</th>
						<th class="px-5 py-3.5">Requested By Account</th>
						<th class="px-5 py-3.5">Submitted Date</th>
						<th class="px-5 py-3.5 text-right">Actions</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
					{#each senders as s}
						<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30">
							<td class="px-5 py-4 font-mono font-bold text-gray-950 dark:text-white text-sm">{s.sender_id}</td>
							<td class="px-5 py-4">
								<span class="inline-flex rounded-md bg-blue-50 px-2 py-0.5 text-[10px] font-bold uppercase text-blue-700">
									{s.type}
								</span>
							</td>
							<td class="px-5 py-4 font-semibold text-gray-800 dark:text-gray-200">{s.requested_by_name || 'Direct Customer'}</td>
							<td class="px-5 py-4 text-gray-400">{new Date(s.created_at).toLocaleDateString()}</td>
							<td class="px-5 py-4 text-right">
								<div class="flex items-center justify-end gap-2">
									<button
										disabled={processingId === s.id}
										onclick={() => approveSender(s.id)}
										class="rounded bg-emerald-600 px-2.5 py-1 text-xs font-semibold text-white shadow-xs hover:bg-emerald-700 disabled:opacity-50"
									>
										Approve Header
									</button>
									<button
										disabled={processingId === s.id}
										onclick={() => rejectSender(s.id)}
										class="rounded bg-rose-50 px-2.5 py-1 text-xs font-semibold text-rose-700 hover:bg-rose-100 dark:bg-rose-950/40 dark:text-rose-300 disabled:opacity-50"
									>
										Reject
									</button>
								</div>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		{/if}
	</div>
</div>

