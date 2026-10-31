<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface Transaction {
		id: number;
		user_id: number;
		customer_name: string;
		customer_phone: string;
		reference: string;
		direction: string;
		amount: number;
		balance_before: number;
		balance_after: number;
		description: string | null;
		created_at: string;
	}

	interface ReportSummary {
		total_credits: number;
		total_debits: number;
		net_flow: number;
		total_tx_count: number;
	}

	let transactions = $state<Transaction[]>([]);
	let summary = $state<ReportSummary>({
		total_credits: 0,
		total_debits: 0,
		net_flow: 0,
		total_tx_count: 0
	});

	let loading = $state(true);
	let search = $state('');
	let directionFilter = $state('all');
	let fromDate = $state('');
	let untilDate = $state('');

	// Pagination
	let currentPage = $state(1);
	let perPage = $state(25);
	let totalRecords = $state(0);
	let totalPages = $derived(Math.ceil(totalRecords / perPage) || 1);

	// Selected transaction for details modal
	let selectedTx = $state<Transaction | null>(null);

	async function loadTransactions() {
		loading = true;
		try {
			const query = new URLSearchParams({
				page: String(currentPage),
				per_page: String(perPage)
			});
			if (search) query.set('search', search);
			if (directionFilter !== 'all') query.set('direction', directionFilter);
			if (fromDate) query.set('from', fromDate);
			if (untilDate) query.set('until', untilDate);

			const res = await supermanApi.get<{
				data: Transaction[];
				meta: { page: number; per_page: number; total: number };
				summary: ReportSummary;
			}>(`/api/v1/superman/billing/reports?${query.toString()}`);

			transactions = res.data || [];
			totalRecords = res.meta?.total || 0;
			if (res.summary) {
				summary = res.summary;
			}
		} catch (err) {
			console.error('Failed to load transactions:', err);
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		loadTransactions();
	});

	function formatCurrency(val: number): string {
		return '৳' + (val || 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
	}

	const exportColumns: ExportColumn<Transaction>[] = [
		{ header: 'Date & Time', getValue: (t) => new Date(t.created_at).toLocaleString() },
		{ header: 'Customer', getValue: (t) => t.customer_name },
		{ header: 'Phone', getValue: (t) => t.customer_phone || '—' },
		{ header: 'Type', getValue: (t) => t.direction.toUpperCase() },
		{ header: 'Amount (BDT)', getValue: (t) => `BDT ${t.amount.toFixed(2)}` },
		{ header: 'Balance Before', getValue: (t) => `BDT ${t.balance_before.toFixed(2)}` },
		{ header: 'Balance After', getValue: (t) => `BDT ${t.balance_after.toFixed(2)}` },
		{ header: 'Reference', getValue: (t) => t.reference || '—' },
		{ header: 'Description', getValue: (t) => t.description || '—' }
	];
</script>

<svelte:head>
	<title>Transactions Ledger — SMS Whiteinfo Administrator</title>
</svelte:head>

<div class="space-y-6">
	<!-- Page Header -->
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<div class="flex items-center gap-2">
				<a href="/superman/billing" class="text-xs font-semibold text-gray-500 hover:text-emerald-600 dark:hover:text-emerald-400">
					Billing
				</a>
				<span class="text-xs text-gray-400">/</span>
				<span class="text-xs font-semibold text-gray-900 dark:text-white">Transactions</span>
			</div>
			<h1 class="mt-1 text-xl font-bold tracking-tight text-gray-900 dark:text-white sm:text-2xl">
				Financial Transactions Ledger
			</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400">
				Double-entry balance audit ledger tracking credits, message deductions, top-ups, and refunds.
			</p>
		</div>

		<div class="flex items-center gap-2">
			<button
				type="button"
				onclick={loadTransactions}
				class="inline-flex items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs font-semibold text-gray-700 transition hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300 dark:hover:bg-gray-800"
			>
				<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0 3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182m0-4.991v4.99" />
				</svg>
				Refresh
			</button>
			<ExportDropdown
				title="Transactions Ledger"
				columns={exportColumns}
				getData={() => transactions}
				disabled={transactions.length === 0}
			/>
			<a
				href="/superman/billing"
				class="inline-flex items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs font-semibold text-gray-700 transition hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300 dark:hover:bg-gray-800"
			>
				<svg class="h-4 w-4 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 17l-5-5m0 0l5-5m-5 5h12" />
				</svg>
				Billing Control Center
			</a>
			<a
				href="/superman/billing/deposits"
				class="inline-flex items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs font-semibold text-gray-700 transition hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300 dark:hover:bg-gray-800"
			>
				<svg class="h-4 w-4 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
				</svg>
				Deposits
			</a>
		</div>
	</div>

	<!-- Summary Cards -->
	<div class="grid grid-cols-2 gap-3 sm:grid-cols-4">
		<div class="rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
			<span class="text-[11px] font-medium text-gray-400 uppercase tracking-wider">Total Credits</span>
			<p class="mt-1 text-lg font-bold text-emerald-600 dark:text-emerald-400">+{formatCurrency(summary.total_credits)}</p>
			<p class="text-[10px] text-gray-400">Total top-ups & refunds</p>
		</div>

		<div class="rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
			<span class="text-[11px] font-medium text-gray-400 uppercase tracking-wider">Total Debits</span>
			<p class="mt-1 text-lg font-bold text-rose-600 dark:text-rose-400">-{formatCurrency(summary.total_debits)}</p>
			<p class="text-[10px] text-gray-400">Total SMS usages deducted</p>
		</div>

		<div class="rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
			<span class="text-[11px] font-medium text-gray-400 uppercase tracking-wider">Net Cash Flow</span>
			<p class="mt-1 text-lg font-bold {summary.net_flow >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-rose-600'}">
				{formatCurrency(summary.net_flow)}
			</p>
			<p class="text-[10px] text-gray-400">Credit minus debit balance</p>
		</div>

		<div class="rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
			<span class="text-[11px] font-medium text-gray-400 uppercase tracking-wider">Transactions Count</span>
			<p class="mt-1 text-lg font-bold text-gray-900 dark:text-white">{summary.total_tx_count.toLocaleString()}</p>
			<p class="text-[10px] text-gray-400">Matching filtered subset</p>
		</div>
	</div>

	<!-- Toolbar & Filters -->
	<div class="flex flex-col gap-3 rounded-xl border border-gray-200 bg-white p-3 shadow-xs dark:border-gray-800 dark:bg-gray-900 sm:flex-row sm:items-center sm:justify-between">
		<div class="relative flex-1 max-w-sm">
			<svg class="pointer-events-none absolute left-3 top-2.5 h-4 w-4 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
			</svg>
			<input
				type="text"
				bind:value={search}
				onkeydown={(e) => {
					if (e.key === 'Enter') {
						currentPage = 1;
						loadTransactions();
					}
				}}
				placeholder="Search customer, note… (Enter)"
				class="w-full rounded-lg border border-gray-200 bg-gray-50 py-1.5 pl-9 pr-3 text-xs text-gray-900 focus:border-emerald-500 focus:bg-white focus:outline-hidden dark:border-gray-800 dark:bg-gray-800 dark:text-white"
			/>
		</div>

		<div class="flex flex-wrap items-center gap-2">
			<!-- Direction Filter -->
			<select
				bind:value={directionFilter}
				class="rounded-lg border border-gray-200 bg-white px-2.5 py-1.5 text-xs text-gray-700 focus:border-emerald-500 focus:outline-hidden dark:border-gray-800 dark:bg-gray-800 dark:text-gray-300"
			>
				<option value="all">All Directions</option>
				<option value="credit">Credits Only (+)</option>
				<option value="debit">Debits Only (-)</option>
			</select>

			<!-- Date range -->
			<input
				type="date"
				bind:value={fromDate}
				class="rounded-lg border border-gray-200 bg-white px-2 py-1 text-xs font-mono text-gray-700 focus:border-emerald-500 focus:outline-hidden dark:border-gray-800 dark:bg-gray-800 dark:text-gray-300"
			/>
			<span class="text-xs text-gray-400">to</span>
			<input
				type="date"
				bind:value={untilDate}
				class="rounded-lg border border-gray-200 bg-white px-2 py-1 text-xs font-mono text-gray-700 focus:border-emerald-500 focus:outline-hidden dark:border-gray-800 dark:bg-gray-800 dark:text-gray-300"
			/>

			<button
				type="button"
				onclick={() => {
					currentPage = 1;
					loadTransactions();
				}}
				class="rounded-lg bg-gray-900 px-3 py-1.5 text-xs font-semibold text-white hover:bg-gray-800 dark:bg-white dark:text-gray-900 dark:hover:bg-gray-100"
			>
				Filter
			</button>
		</div>
	</div>

	<!-- Transactions Table -->
	<div class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-xs dark:border-gray-800 dark:bg-gray-900">
		<div class="overflow-x-auto">
			<table class="w-full text-left text-xs">
				<thead>
					<tr class="border-b border-gray-100 bg-gray-50/50 text-[10px] font-semibold uppercase tracking-wider text-gray-400 dark:border-gray-800 dark:bg-gray-800/40">
						<th class="py-3 px-4">TX ID</th>
						<th class="py-3 px-3">Date</th>
						<th class="py-3 px-3">Customer</th>
						<th class="py-3 px-3">Type</th>
						<th class="py-3 px-3 text-right">Amount</th>
						<th class="py-3 px-3 text-right">Balance Before</th>
						<th class="py-3 px-3 text-right">Balance After</th>
						<th class="py-3 px-3">Reference / Description</th>
						<th class="py-3 px-4 text-right">Action</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
					{#each transactions as t}
						<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30 transition-colors">
							<td class="py-3 px-4 font-mono font-medium text-gray-500">
								#{t.id}
							</td>
							<td class="py-3 px-3 font-mono text-[11px] text-gray-400">
								{t.created_at}
							</td>
							<td class="py-3 px-3">
								<p class="font-bold text-gray-900 dark:text-white">{t.customer_name}</p>
								<p class="font-mono text-[10px] text-gray-400">{t.customer_phone}</p>
							</td>
							<td class="py-3 px-3">
								<span class="inline-flex rounded-md px-1.5 py-0.5 text-[9px] font-bold uppercase tracking-wider {t.direction === 'credit' ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300' : 'bg-rose-100 text-rose-800 dark:bg-rose-950 dark:text-rose-300'}">
									{t.direction}
								</span>
							</td>
							<td class="py-3 px-3 text-right font-mono font-bold {t.direction === 'credit' ? 'text-emerald-600 dark:text-emerald-400' : 'text-rose-600 dark:text-rose-400'}">
								{t.direction === 'credit' ? '+' : '-'}{formatCurrency(t.amount)}
							</td>
							<td class="py-3 px-3 text-right font-mono text-gray-500">
								{formatCurrency(t.balance_before)}
							</td>
							<td class="py-3 px-3 text-right font-mono font-semibold text-gray-900 dark:text-white">
								{formatCurrency(t.balance_after)}
							</td>
							<td class="py-3 px-3 max-w-xs truncate text-gray-600 dark:text-gray-300" title={t.description || ''}>
								{t.description || '—'}
							</td>
							<td class="py-3 px-4 text-right">
								<button
									type="button"
									onclick={() => (selectedTx = t)}
									class="rounded-lg border border-gray-200 bg-white px-2.5 py-1 text-xs font-medium text-gray-700 hover:bg-gray-50 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-300"
								>
									View
								</button>
							</td>
						</tr>
					{:else}
						<tr>
							<td colspan="9" class="py-8 text-center text-xs text-gray-400">
								{loading ? 'Loading ledger transactions…' : 'No transactions matching filters.'}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>

		<!-- Pagination -->
		{#if totalPages > 1}
			<div class="flex items-center justify-between border-t border-gray-100 px-4 py-3 dark:border-gray-800">
				<p class="text-xs text-gray-500">
					Page <strong class="font-semibold text-gray-900 dark:text-white">{currentPage}</strong> of <strong class="font-semibold text-gray-900 dark:text-white">{totalPages}</strong>
				</p>
				<div class="flex items-center gap-1.5">
					<button
						type="button"
						onclick={() => {
							if (currentPage > 1) {
								currentPage--;
								loadTransactions();
							}
						}}
						disabled={currentPage <= 1}
						class="rounded-md border border-gray-200 px-2.5 py-1 text-xs font-semibold text-gray-700 hover:bg-gray-50 disabled:opacity-40 dark:border-gray-700 dark:text-gray-300 dark:hover:bg-gray-800"
					>
						Previous
					</button>
					<button
						type="button"
						onclick={() => {
							if (currentPage < totalPages) {
								currentPage++;
								loadTransactions();
							}
						}}
						disabled={currentPage >= totalPages}
						class="rounded-md border border-gray-200 px-2.5 py-1 text-xs font-semibold text-gray-700 hover:bg-gray-50 disabled:opacity-40 dark:border-gray-700 dark:text-gray-300 dark:hover:bg-gray-800"
					>
						Next
					</button>
				</div>
			</div>
		{/if}
	</div>
</div>

<!-- Modal: Transaction Details -->
{#if selectedTx}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-gray-900/60 backdrop-blur-xs p-4">
		<div class="w-full max-w-md rounded-2xl border border-gray-200 bg-white p-6 shadow-xl dark:border-gray-800 dark:bg-gray-900">
			<div class="flex items-center justify-between border-b border-gray-100 pb-3 dark:border-gray-800">
				<div>
					<h3 class="text-base font-bold text-gray-900 dark:text-white">Transaction #{selectedTx.id}</h3>
					<p class="text-xs text-gray-500">Ledger entry details</p>
				</div>
				<button
					type="button"
					onclick={() => (selectedTx = null)}
					class="rounded-lg p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-gray-800"
				>
					<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>

			<div class="mt-4 space-y-3 text-xs">
				<div class="flex items-center justify-between border-b border-gray-100 pb-2 dark:border-gray-800">
					<span class="text-gray-400">Customer</span>
					<span class="font-bold text-gray-900 dark:text-white">{selectedTx.customer_name}</span>
				</div>
				<div class="flex items-center justify-between border-b border-gray-100 pb-2 dark:border-gray-800">
					<span class="text-gray-400">Direction</span>
					<span class="font-bold uppercase {selectedTx.direction === 'credit' ? 'text-emerald-600' : 'text-rose-600'}">
						{selectedTx.direction}
					</span>
				</div>
				<div class="flex items-center justify-between border-b border-gray-100 pb-2 dark:border-gray-800">
					<span class="text-gray-400">Amount</span>
					<span class="font-mono font-bold text-sm {selectedTx.direction === 'credit' ? 'text-emerald-600' : 'text-rose-600'}">
						{selectedTx.direction === 'credit' ? '+' : '-'}{formatCurrency(selectedTx.amount)}
					</span>
				</div>
				<div class="flex items-center justify-between border-b border-gray-100 pb-2 dark:border-gray-800">
					<span class="text-gray-400">Balance Before</span>
					<span class="font-mono text-gray-700 dark:text-gray-300">{formatCurrency(selectedTx.balance_before)}</span>
				</div>
				<div class="flex items-center justify-between border-b border-gray-100 pb-2 dark:border-gray-800">
					<span class="text-gray-400">Balance After</span>
					<span class="font-mono font-semibold text-gray-900 dark:text-white">{formatCurrency(selectedTx.balance_after)}</span>
				</div>
				<div class="flex items-center justify-between border-b border-gray-100 pb-2 dark:border-gray-800">
					<span class="text-gray-400">Reference Type</span>
					<span class="font-mono text-gray-700 dark:text-gray-300">{selectedTx.reference}</span>
				</div>
				<div class="flex items-start justify-between border-b border-gray-100 pb-2 dark:border-gray-800">
					<span class="text-gray-400">Description</span>
					<span class="font-medium text-gray-800 dark:text-gray-200 text-right max-w-[240px]">
						{selectedTx.description || '—'}
					</span>
				</div>
				<div class="flex items-center justify-between">
					<span class="text-gray-400">Timestamp</span>
					<span class="font-mono text-gray-700 dark:text-gray-300">{selectedTx.created_at}</span>
				</div>
			</div>

			<div class="mt-6 flex justify-end">
				<button
					type="button"
					onclick={() => (selectedTx = null)}
					class="rounded-lg bg-gray-900 px-4 py-2 text-xs font-semibold text-white hover:bg-gray-800 dark:bg-white dark:text-gray-900"
				>
					Close
				</button>
			</div>
		</div>
	</div>
{/if}
