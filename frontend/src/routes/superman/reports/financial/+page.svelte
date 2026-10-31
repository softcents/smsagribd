<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface TransactionItem {
		id: number;
		customer: string;
		reference: string;
		direction: string;
		amount: number;
		balance_before: number;
		balance_after: number;
		created_at: string;
	}

	let transactions = $state<TransactionItem[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let search = $state('');

	async function loadFinancialReport() {
		loading = true;
		error = null;
		try {
			const res = await supermanApi.get<{ data: TransactionItem[] }>('/api/v1/superman/transactions?per_page=100');
			transactions = res.data ?? [];
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Failed to load financial report.';
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		void loadFinancialReport();
	});

	let filtered = $derived(
		transactions.filter((t) => search === '' || t.customer.toLowerCase().includes(search.toLowerCase()) || (t.reference && t.reference.toLowerCase().includes(search.toLowerCase())))
	);

	let totalCredits = $derived(transactions.filter(t => t.direction === 'credit').reduce((sum, t) => sum + Number(t.amount), 0));
	let totalDebits = $derived(transactions.filter(t => t.direction === 'debit').reduce((sum, t) => sum + Number(t.amount), 0));

	const exportColumns: ExportColumn<TransactionItem>[] = [
		{ header: 'Customer', getValue: (t) => t.customer },
		{ header: 'Reference / Trx ID', getValue: (t) => t.reference },
		{ header: 'Type', getValue: (t) => t.direction.toUpperCase() },
		{ header: 'Amount', getValue: (t) => `BDT ${t.amount.toFixed(2)}` },
		{ header: 'Balance Before', getValue: (t) => `BDT ${t.balance_before.toFixed(2)}` },
		{ header: 'Balance After', getValue: (t) => `BDT ${t.balance_after.toFixed(2)}` },
		{ header: 'Date & Time', getValue: (t) => (t.created_at ? new Date(t.created_at).toLocaleString() : '—') }
	];
</script>

<svelte:head>
	<title>Financial Reports — Superman Administrator</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">Financial & Ledger Reports</h1>
			<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
				Audit trail of customer deposits, wallet top-ups, and per-message debit deductions.
			</p>
		</div>
		<div class="flex items-center gap-2">
			<ExportDropdown
				title="Financial and Ledger Reports"
				columns={exportColumns}
				getData={() => filtered}
				disabled={filtered.length === 0}
			/>
			<button onclick={loadFinancialReport} class="rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs font-medium text-gray-700 shadow-xs hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300">
				Refresh Ledger
			</button>
		</div>
	</div>

	<!-- Stats Bar -->
	<div class="grid grid-cols-2 gap-4 sm:grid-cols-3">
		<div class="rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
			<span class="text-[11px] font-semibold text-gray-500 uppercase tracking-wider">Total Top-Ups (Credits)</span>
			<p class="mt-1 text-2xl font-bold text-emerald-600">৳{totalCredits.toFixed(2)}</p>
		</div>
		<div class="rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
			<span class="text-[11px] font-semibold text-gray-500 uppercase tracking-wider">Total Usage (Debits)</span>
			<p class="mt-1 text-2xl font-bold text-rose-600">৳{totalDebits.toFixed(2)}</p>
		</div>
		<div class="col-span-2 rounded-xl border border-gray-200 bg-white p-4 shadow-xs sm:col-span-1 dark:border-gray-800 dark:bg-gray-900">
			<span class="text-[11px] font-semibold text-gray-500 uppercase tracking-wider">Total Audit Records</span>
			<p class="mt-1 text-2xl font-bold text-gray-900 dark:text-white">{transactions.length}</p>
		</div>
	</div>

	<div class="flex items-center gap-3 rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
		<input
			type="search"
			bind:value={search}
			placeholder="Search by customer name or reference..."
			class="w-full max-w-sm rounded-lg border border-gray-200 bg-gray-50/70 px-3 py-1.5 text-xs outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
		/>
	</div>

	<div class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-xs dark:border-gray-800 dark:bg-gray-900">
		{#if loading}
			<div class="py-16 text-center text-xs text-gray-500">Loading financial ledger...</div>
		{:else if error}
			<div class="p-6 text-center text-xs text-rose-600">{error}</div>
		{:else if filtered.length === 0}
			<div class="py-16 text-center text-xs text-gray-500">No financial transactions recorded.</div>
		{:else}
			<div class="overflow-x-auto">
				<table class="w-full text-left text-xs">
					<thead>
						<tr class="border-b border-gray-100 bg-gray-50/50 text-[10px] font-semibold uppercase tracking-wider text-gray-500 dark:border-gray-800 dark:bg-gray-800/50 dark:text-gray-400">
							<th class="px-4 py-3">Customer</th>
							<th class="px-4 py-3">Reference / Txn ID</th>
							<th class="px-4 py-3">Direction</th>
							<th class="px-4 py-3 text-right">Amount</th>
							<th class="px-4 py-3 text-right">Balance Before</th>
							<th class="px-4 py-3 text-right">Balance After</th>
							<th class="px-4 py-3 text-right">Timestamp</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
						{#each filtered as t}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30">
								<td class="px-4 py-3 font-semibold text-gray-950 dark:text-white">{t.customer}</td>
								<td class="px-4 py-3 font-mono text-[11px] text-gray-500">{t.reference || `TXN-${t.id}`}</td>
								<td class="px-4 py-3">
									<span class="inline-flex rounded-full px-2 py-0.5 text-[10px] font-bold uppercase {t.direction === 'credit' ? 'bg-emerald-50 text-emerald-700' : 'bg-rose-50 text-rose-700'}">
										{t.direction}
									</span>
								</td>
								<td class="px-4 py-3 text-right font-bold {t.direction === 'credit' ? 'text-emerald-600' : 'text-rose-600'}">
									{t.direction === 'credit' ? '+' : '-'}৳{Number(t.amount).toFixed(2)}
								</td>
								<td class="px-4 py-3 text-right text-gray-500">৳{Number(t.balance_before || 0).toFixed(2)}</td>
								<td class="px-4 py-3 text-right font-semibold text-gray-900 dark:text-white">৳{Number(t.balance_after || 0).toFixed(2)}</td>
								<td class="px-4 py-3 text-right text-gray-400">{new Date(t.created_at).toLocaleString()}</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</div>
</div>

