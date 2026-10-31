<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface TransactionRecord {
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

	let records = $state<TransactionRecord[]>([]);
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

	async function loadReports() {
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
				data: TransactionRecord[];
				meta: { page: number; per_page: number; total: number };
				summary: ReportSummary;
			}>(`/api/v1/superman/billing/reports?${query.toString()}`);

			records = res.data || [];
			totalRecords = res.meta?.total || 0;
			if (res.summary) {
				summary = res.summary;
			}
		} catch (err) {
			console.error('Failed to load billing reports:', err);
		} finally {
			loading = false;
		}
	}

	const exportColumns: ExportColumn<TransactionRecord>[] = [
		{ header: 'TX ID', getValue: (r) => r.id },
		{ header: 'Date', getValue: (r) => r.created_at },
		{ header: 'Customer', getValue: (r) => r.customer_name },
		{ header: 'Phone', getValue: (r) => r.customer_phone },
		{ header: 'Type', getValue: (r) => r.direction },
		{ header: 'Amount', getValue: (r) => `BDT ${r.amount.toFixed(2)}` },
		{ header: 'Balance Before', getValue: (r) => `BDT ${r.balance_before.toFixed(2)}` },
		{ header: 'Balance After', getValue: (r) => `BDT ${r.balance_after.toFixed(2)}` },
		{ header: 'Description', getValue: (r) => r.description || '—' }
	];

	onMount(() => {
		loadReports();
	});

	function formatCurrency(val: number): string {
		return '৳' + (val || 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
	}
</script>

<svelte:head>
	<title>Financial & Billing Audit Reports — SMS Whiteinfo Administrator</title>
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
				<span class="text-xs font-semibold text-gray-900 dark:text-white">Reports</span>
			</div>
			<h1 class="mt-1 text-xl font-bold tracking-tight text-gray-900 dark:text-white sm:text-2xl">
				Financial & Billing Audit Reports
			</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400">
				Detailed ledger transactions audit trail, custom period balance movements, and downloadable financial reports.
			</p>
		</div>

		<div class="flex items-center gap-2">
			<ExportDropdown
				title="Financial and Billing Audit Reports"
				columns={exportColumns}
				getData={() => records}
				disabled={records.length === 0}
			/>
			<a
				href="/superman/billing"
				class="inline-flex items-center gap-1.5 rounded-lg bg-emerald-600 px-3.5 py-2 text-xs font-semibold text-white shadow-xs hover:bg-emerald-500"
			>
				<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
				</svg>
				Billing Dashboard
			</a>
		</div>
	</div>

	<!-- Summary Totals -->
	<div class="grid grid-cols-2 gap-3 sm:grid-cols-4">
		<div class="rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
			<span class="text-[11px] font-medium text-gray-400 uppercase tracking-wider">Total Credits</span>
			<p class="mt-1 text-lg font-bold text-emerald-600 dark:text-emerald-400">+{formatCurrency(summary.total_credits)}</p>
			<p class="text-[10px] text-gray-400">Top-ups and manual additions</p>
		</div>

		<div class="rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
			<span class="text-[11px] font-medium text-gray-400 uppercase tracking-wider">Total Debits</span>
			<p class="mt-1 text-lg font-bold text-rose-600 dark:text-rose-400">-{formatCurrency(summary.total_debits)}</p>
			<p class="text-[10px] text-gray-400">Outbound SMS consumption</p>
		</div>

		<div class="rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
			<span class="text-[11px] font-medium text-gray-400 uppercase tracking-wider">Net Movement</span>
			<p class="mt-1 text-lg font-bold {summary.net_flow >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-rose-600'}">
				{formatCurrency(summary.net_flow)}
			</p>
			<p class="text-[10px] text-gray-400">Credits minus debits</p>
		</div>

		<div class="rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
			<span class="text-[11px] font-medium text-gray-400 uppercase tracking-wider">Transactions Analyzed</span>
			<p class="mt-1 text-lg font-bold text-gray-900 dark:text-white">{summary.total_tx_count.toLocaleString()}</p>
			<p class="text-[10px] text-gray-400">Audit entries in query</p>
		</div>
	</div>

	<!-- Filter Controls -->
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
						loadReports();
					}
				}}
				placeholder="Search customer, note, phone… (Enter)"
				class="w-full rounded-lg border border-gray-200 bg-gray-50 py-1.5 pl-9 pr-3 text-xs text-gray-900 focus:border-emerald-500 focus:bg-white focus:outline-hidden dark:border-gray-800 dark:bg-gray-800 dark:text-white"
			/>
		</div>

		<div class="flex flex-wrap items-center gap-2">
			<!-- Direction Filter -->
			<select
				bind:value={directionFilter}
				class="rounded-lg border border-gray-200 bg-white px-2.5 py-1.5 text-xs text-gray-700 focus:border-emerald-500 focus:outline-hidden dark:border-gray-800 dark:bg-gray-800 dark:text-gray-300"
			>
				<option value="all">All Movements</option>
				<option value="credit">Credit Only (+)</option>
				<option value="debit">Debit Only (-)</option>
			</select>

			<!-- Date Range -->
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
					loadReports();
				}}
				class="rounded-lg bg-gray-900 px-3 py-1.5 text-xs font-semibold text-white hover:bg-gray-800 dark:bg-white dark:text-gray-900 dark:hover:bg-gray-100"
			>
				Generate Report
			</button>
		</div>
	</div>

	<!-- Report Table -->
	<div class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-xs dark:border-gray-800 dark:bg-gray-900">
		<div class="overflow-x-auto">
			<table class="w-full text-left text-xs">
				<thead>
					<tr class="border-b border-gray-100 bg-gray-50/50 text-[10px] font-semibold uppercase tracking-wider text-gray-400 dark:border-gray-800 dark:bg-gray-800/40">
						<th class="py-3 px-4">Timestamp</th>
						<th class="py-3 px-3">Customer</th>
						<th class="py-3 px-3">Type</th>
						<th class="py-3 px-3 text-right">Amount</th>
						<th class="py-3 px-3 text-right">Balance Before</th>
						<th class="py-3 px-3 text-right">Balance After</th>
						<th class="py-3 px-4">Reference / Audit Details</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
					{#each records as r}
						<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30 transition-colors">
							<td class="py-3 px-4 font-mono text-[11px] text-gray-500 whitespace-nowrap">
								{r.created_at}
							</td>
							<td class="py-3 px-3">
								<p class="font-bold text-gray-900 dark:text-white">{r.customer_name}</p>
								<p class="font-mono text-[10px] text-gray-400">{r.customer_phone}</p>
							</td>
							<td class="py-3 px-3">
								<span class="inline-flex rounded-md px-1.5 py-0.5 text-[9px] font-bold uppercase tracking-wider {r.direction === 'credit' ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300' : 'bg-rose-100 text-rose-800 dark:bg-rose-950 dark:text-rose-300'}">
									{r.direction}
								</span>
							</td>
							<td class="py-3 px-3 text-right font-mono font-bold {r.direction === 'credit' ? 'text-emerald-600 dark:text-emerald-400' : 'text-rose-600 dark:text-rose-400'}">
								{r.direction === 'credit' ? '+' : '-'}{formatCurrency(r.amount)}
							</td>
							<td class="py-3 px-3 text-right font-mono text-gray-500">
								{formatCurrency(r.balance_before)}
							</td>
							<td class="py-3 px-3 text-right font-mono font-semibold text-gray-900 dark:text-white">
								{formatCurrency(r.balance_after)}
							</td>
							<td class="py-3 px-4 max-w-sm truncate text-gray-600 dark:text-gray-300" title={r.description || ''}>
								{r.description || '—'}
							</td>
						</tr>
					{:else}
						<tr>
							<td colspan="7" class="py-8 text-center text-xs text-gray-400">
								{loading ? 'Compiling financial audit reports…' : 'No records found matching filters.'}
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
								loadReports();
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
								loadReports();
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
