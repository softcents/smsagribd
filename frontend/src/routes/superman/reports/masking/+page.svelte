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
		gateway?: string;
		created_at: string;
	}

	let messages = $state<MessageItem[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let search = $state('');

	async function loadReport() {
		loading = true;
		error = null;
		try {
			const res = await supermanApi.get<{ data: MessageItem[] }>('/api/v1/superman/messages?per_page=100');
			messages = res.data ?? [];
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Failed to load masking report.';
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		void loadReport();
	});

	let filtered = $derived(
		messages.filter((m) => search === '' || m.recipient.includes(search) || m.customer.toLowerCase().includes(search.toLowerCase()))
	);

	const exportColumns: ExportColumn<MessageItem>[] = [
		{ header: 'Customer', getValue: (m) => m.customer },
		{ header: 'Recipient', getValue: (m) => m.recipient },
		{ header: 'Message Body', getValue: (m) => m.body },
		{ header: 'Gateway', getValue: (m) => m.gateway || 'Auto' },
		{ header: 'Status', getValue: (m) => m.status.toUpperCase() },
		{ header: 'Cost (BDT)', getValue: (m) => `BDT ${Number(m.cost).toFixed(4)}` },
		{ header: 'Timestamp', getValue: (m) => new Date(m.created_at).toLocaleString() }
	];
</script>

<svelte:head>
	<title>Masking SMS Report — Superman Administrator</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">Masking SMS Delivery Report</h1>
			<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
				Filtered audit log for branded alphanumeric sender headers and shortcode dispatches.
			</p>
		</div>
		<div class="flex items-center gap-2">
			<button onclick={loadReport} class="rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs font-medium text-gray-700 shadow-xs hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300">
				Refresh
			</button>
			<ExportDropdown
				title="Masking SMS Report"
				columns={exportColumns}
				getData={() => filtered}
				disabled={filtered.length === 0}
			/>
		</div>
	</div>

	<div class="flex items-center gap-3 rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
		<input
			type="search"
			bind:value={search}
			placeholder="Search by recipient or customer..."
			class="w-full max-w-sm rounded-lg border border-gray-200 bg-gray-50/70 px-3 py-1.5 text-xs outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
		/>
	</div>

	<div class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-xs dark:border-gray-800 dark:bg-gray-900">
		{#if loading}
			<div class="py-16 text-center text-xs text-gray-500">Loading masking report...</div>
		{:else if error}
			<div class="p-6 text-center text-xs text-rose-600">{error}</div>
		{:else if filtered.length === 0}
			<div class="py-16 text-center text-xs text-gray-500">No masking SMS records found.</div>
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
							<th class="px-4 py-3 text-right">Cost</th>
							<th class="px-4 py-3 text-right">Timestamp</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
						{#each filtered as m}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30">
								<td class="px-4 py-3 font-semibold text-gray-950 dark:text-white">{m.customer}</td>
								<td class="px-4 py-3 font-mono">{m.recipient}</td>
								<td class="max-w-xs truncate px-4 py-3 text-gray-600 dark:text-gray-400">{m.body}</td>
								<td class="px-4 py-3 font-medium text-blue-600">{m.gateway || 'Auto'}</td>
								<td class="px-4 py-3 font-semibold text-emerald-600 uppercase text-[11px]">{m.status}</td>
								<td class="px-4 py-3 text-right font-medium">৳{Number(m.cost).toFixed(4)}</td>
								<td class="px-4 py-3 text-right text-gray-400">{new Date(m.created_at).toLocaleString()}</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</div>
</div>

