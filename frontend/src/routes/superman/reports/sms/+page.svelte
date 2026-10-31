<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface MessageItem {
		id: number;
		customer: string;
		recipient: string;
		body: string;
		status: string;
		cost: number;
		operator_cost?: number;
		gateway?: string;
		created_at: string;
	}

	interface Meta {
		page: number;
		per_page: number;
		total: number;
	}

	let messages = $state<MessageItem[]>([]);
	let meta = $state<Meta>({ page: 1, per_page: 25, total: 0 });
	let loading = $state(true);
	let error = $state<string | null>(null);

	let search = $state('');
	let statusFilter = $state('all');
	let perPageChoice = $state(25);

	async function loadMessages(targetPage = 1) {
		loading = true;
		error = null;
		try {
			const params = new URLSearchParams({
				page: String(targetPage),
				per_page: String(perPageChoice)
			});
			if (search.trim()) params.set('search', search.trim());
			if (statusFilter !== 'all') params.set('status', statusFilter);

			const res = await supermanApi.get<{ data: MessageItem[]; meta?: Meta }>(
				`/api/v1/superman/messages?${params.toString()}`
			);
			messages = res.data ?? [];
			if (res.meta) {
				meta = res.meta;
			} else {
				meta = { page: targetPage, per_page: perPageChoice, total: messages.length };
			}
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Failed to load SMS report.';
		} finally {
			loading = false;
		}
	}

	function handleFilter(e: Event) {
		e.preventDefault();
		void loadMessages(1);
	}

	onMount(() => {
		void loadMessages(1);
	});

	let totalPages = $derived(Math.ceil((meta.total || 1) / meta.per_page));

	const exportColumns: ExportColumn<MessageItem>[] = [
		{ header: 'Recipient', getValue: (m) => m.recipient },
		{ header: 'Customer', getValue: (m) => m.customer },
		{ header: 'Gateway', getValue: (m) => m.gateway || '—' },
		{ header: 'Message Body', getValue: (m) => m.body },
		{ header: 'Status', getValue: (m) => m.status.toUpperCase() },
		{ header: 'Customer Cost', getValue: (m) => `BDT ${m.cost.toFixed(4)}` },
		{ header: 'Operator Cost', getValue: (m) => (m.operator_cost !== undefined ? `BDT ${m.operator_cost.toFixed(4)}` : '—') },
		{ header: 'Timestamp', getValue: (m) => (m.created_at ? new Date(m.created_at).toLocaleString() : '—') }
	];

	async function getAllMessagesForExport(): Promise<MessageItem[]> {
		try {
			const params = new URLSearchParams({
				page: '1',
				per_page: '10000'
			});
			if (search.trim()) params.set('search', search.trim());
			if (statusFilter !== 'all') params.set('status', statusFilter);
			const res = await supermanApi.get<{ data: MessageItem[] }>(
				`/api/v1/superman/messages?${params.toString()}`
			);
			return res.data ?? messages;
		} catch {
			return messages;
		}
	}
</script>

<svelte:head>
	<title>SMS Reports — Superman Administrator</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">Historical SMS Reports & Logs</h1>
			<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
				Comprehensive audit log of all processed messages, upstream gateway routing, and customer billing deductions.
			</p>
		</div>
		<div class="flex items-center gap-2">
			<ExportDropdown
				title="SMS Reports"
				columns={exportColumns}
				getData={getAllMessagesForExport}
				disabled={meta.total === 0 && messages.length === 0}
			/>
			<button
				onclick={() => loadMessages(meta.page)}
				class="inline-flex items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs font-medium text-gray-700 shadow-xs hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300 dark:hover:bg-gray-800"
			>
				<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0 3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182m0-4.991v4.99" />
				</svg>
				Refresh
			</button>
		</div>
	</div>

	<!-- Search & Filters -->
	<form onsubmit={handleFilter} class="flex flex-wrap items-center gap-3 rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
		<div class="w-full max-w-xs">
			<input
				type="search"
				bind:value={search}
				placeholder="Search by recipient or customer..."
				class="w-full rounded-lg border border-gray-200 bg-gray-50/70 px-3 py-1.5 text-xs outline-none focus:border-emerald-500 focus:bg-white dark:border-gray-700 dark:bg-gray-800 dark:text-white"
			/>
		</div>
		<select
			bind:value={statusFilter}
			onchange={() => loadMessages(1)}
			class="rounded-lg border border-gray-200 bg-white px-3 py-1.5 text-xs text-gray-700 outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-300"
		>
			<option value="all">All Delivery States</option>
			<option value="delivered">Delivered</option>
			<option value="sent">Sent (Pending DLR)</option>
			<option value="pending">Queued / Pending</option>
			<option value="failed">Failed / Rejected</option>
		</select>
		<select
			bind:value={perPageChoice}
			onchange={() => loadMessages(1)}
			class="rounded-lg border border-gray-200 bg-white px-3 py-1.5 text-xs text-gray-700 outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-300"
		>
			<option value={10}>10 per page</option>
			<option value={25}>25 per page</option>
			<option value={50}>50 per page</option>
			<option value={100}>100 per page</option>
		</select>
		<button
			type="submit"
			class="rounded-lg bg-emerald-600 px-4 py-1.5 text-xs font-semibold text-white shadow-xs hover:bg-emerald-700"
		>
			Apply Filters
		</button>
	</form>

	<!-- Table -->
	<div class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-xs dark:border-gray-800 dark:bg-gray-900">
		{#if loading}
			<div class="flex items-center justify-center py-20 text-xs text-gray-500">
				<svg class="mr-2 h-4 w-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path></svg>
				Loading SMS logs...
			</div>
		{:else if error}
			<div class="p-6 text-center text-xs text-rose-600">{error}</div>
		{:else if messages.length === 0}
			<div class="py-16 text-center text-xs text-gray-500">No message records found matching your filters.</div>
		{:else}
			<div class="overflow-x-auto">
				<table class="w-full text-left text-xs">
					<thead>
						<tr class="border-b border-gray-100 bg-gray-50/50 text-[10px] font-semibold uppercase tracking-wider text-gray-500 dark:border-gray-800 dark:bg-gray-800/50 dark:text-gray-400">
							<th class="px-4 py-3">Customer</th>
							<th class="px-4 py-3">Recipient</th>
							<th class="px-4 py-3">Message Body</th>
							<th class="px-4 py-3">Gateway</th>
							<th class="px-4 py-3">Status</th>
							<th class="px-4 py-3 text-right">Billed Cost</th>
							<th class="px-4 py-3 text-right">Timestamp</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
						{#each messages as m}
							<tr class="transition-colors hover:bg-gray-50/50 dark:hover:bg-gray-800/30">
								<td class="px-4 py-3 font-semibold text-gray-950 dark:text-white">
									{m.customer || 'Customer'}
								</td>
								<td class="px-4 py-3 font-mono font-medium text-gray-800 dark:text-gray-200">
									{m.recipient}
								</td>
								<td class="max-w-xs truncate px-4 py-3 text-gray-600 dark:text-gray-400" title={m.body}>
									{m.body}
								</td>
								<td class="px-4 py-3">
									<span class="rounded bg-blue-50 px-2 py-0.5 text-[10px] font-bold text-blue-700 uppercase dark:bg-blue-950/40 dark:text-blue-300">
										{m.gateway || 'Auto'}
									</span>
								</td>
								<td class="px-4 py-3">
									{#if m.status === 'delivered'}
										<span class="inline-flex rounded-full bg-emerald-50 px-2.5 py-0.5 text-[10px] font-bold text-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-400">
											DELIVERED
										</span>
									{:else if m.status === 'sent'}
										<span class="inline-flex rounded-full bg-blue-50 px-2.5 py-0.5 text-[10px] font-bold text-blue-700 dark:bg-blue-950/40 dark:text-blue-400">
											SENT
										</span>
									{:else if m.status === 'pending' || m.status === 'sending'}
										<span class="inline-flex rounded-full bg-amber-50 px-2.5 py-0.5 text-[10px] font-bold text-amber-700 dark:bg-amber-950/40 dark:text-amber-400">
											PENDING
										</span>
									{:else}
										<span class="inline-flex rounded-full bg-rose-50 px-2.5 py-0.5 text-[10px] font-bold text-rose-700 dark:bg-rose-950/40 dark:text-rose-400">
											FAILED
										</span>
									{/if}
								</td>
								<td class="px-4 py-3 text-right font-medium text-gray-900 dark:text-white">
									৳{Number(m.cost).toFixed(4)}
								</td>
								<td class="px-4 py-3 text-right text-gray-400">
									{new Date(m.created_at).toLocaleString()}
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>

			<!-- Pagination Footer -->
			{#if totalPages > 1}
				<div class="flex items-center justify-between border-t border-gray-100 px-4 py-3 text-xs text-gray-500 dark:border-gray-800">
					<div>
						Showing {(meta.page - 1) * meta.per_page + 1}–{Math.min(meta.page * meta.per_page, meta.total)} of {meta.total} records
					</div>
					<div class="flex items-center gap-1">
						<button
							onclick={() => loadMessages(meta.page - 1)}
							disabled={meta.page <= 1}
							class="rounded border border-gray-200 px-2.5 py-1 text-xs hover:bg-gray-50 disabled:opacity-40 dark:border-gray-700 dark:hover:bg-gray-800"
						>
							Previous
						</button>
						<span class="px-2 text-xs font-semibold text-gray-700 dark:text-gray-300">
							Page {meta.page} of {totalPages}
						</span>
						<button
							onclick={() => loadMessages(meta.page + 1)}
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

