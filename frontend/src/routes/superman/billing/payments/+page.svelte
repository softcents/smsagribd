<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface PaymentItem {
		id: number;
		transaction_id: string | null;
		customer_name: string;
		customer_phone: string;
		amount: number;
		currency: string;
		method: string;
		status: string;
		paid_at: string | null;
		created_at: string;
	}

	let payments = $state<PaymentItem[]>([]);
	let loading = $state(true);
	let search = $state('');
	let methodFilter = $state('all');
	let statusFilter = $state('all');

	// Pagination
	let currentPage = $state(1);
	let perPage = $state(25);
	let totalRecords = $state(0);
	let totalPages = $derived(Math.ceil(totalRecords / perPage) || 1);

	// Stats
	let totalCollected = $derived(
		payments.filter((p) => p.status === 'completed').reduce((acc, p) => acc + p.amount, 0)
	);
	let completedCount = $derived(
		payments.filter((p) => p.status === 'completed').length
	);
	let successRate = $derived(
		payments.length > 0 ? Math.round((completedCount / payments.length) * 100) : 100
	);

	async function loadPayments() {
		loading = true;
		try {
			const query = new URLSearchParams({
				page: String(currentPage),
				per_page: String(perPage)
			});
			if (search) query.set('search', search);
			if (methodFilter !== 'all') query.set('method', methodFilter);
			if (statusFilter !== 'all') query.set('status', statusFilter);

			const res = await supermanApi.get<{
				data: PaymentItem[];
				meta: { page: number; per_page: number; total: number };
			}>(`/api/v1/superman/billing/payments?${query.toString()}`);

			payments = res.data || [];
			totalRecords = res.meta?.total || 0;
		} catch (err) {
			console.error('Failed to load payments:', err);
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		loadPayments();
	});

	function formatCurrency(val: number): string {
		return '৳' + (val || 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
	}
	const exportColumns: ExportColumn<PaymentItem>[] = [
		{ header: 'ID', getValue: (p) => p.id },
		{ header: 'Customer', getValue: (p) => p.customer_name },
		{ header: 'Phone', getValue: (p) => p.customer_phone },
		{ header: 'Amount', getValue: (p) => `BDT ${p.amount.toFixed(2)}` },
		{ header: 'Method', getValue: (p) => p.method },
		{ header: 'Status', getValue: (p) => p.status },
		{ header: 'Transaction ID', getValue: (p) => p.transaction_id || '—' },
		{ header: 'Created At', getValue: (p) => p.created_at }
	];
</script>

<svelte:head>
	<title>Payments & Gateways — SMS Whiteinfo Administrator</title>
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
				<span class="text-xs font-semibold text-gray-900 dark:text-white">Payments</span>
			</div>
			<h1 class="mt-1 text-xl font-bold tracking-tight text-gray-900 dark:text-white sm:text-2xl">
				Payment Gateways & Channel Processing
			</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400">
				Monitor payment gateways, automated merchant reconciliations, and manual payment settlement records.
			</p>
		</div>

		<div class="flex items-center gap-2">
			<ExportDropdown
				title="Payments and Gateways"
				columns={exportColumns}
				getData={() => payments}
				disabled={payments.length === 0}
			/>
			<a
				href="/superman/billing/deposits"
				class="inline-flex items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs font-semibold text-gray-700 transition hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300 dark:hover:bg-gray-800"
			>
				<svg class="h-4 w-4 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
				</svg>
				Customer Deposits
			</a>
			<a
				href="/superman/billing/invoices"
				class="inline-flex items-center gap-1.5 rounded-lg bg-emerald-600 px-3.5 py-2 text-xs font-semibold text-white shadow-xs hover:bg-emerald-500"
			>
				<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
				</svg>
				Invoices
			</a>
		</div>
	</div>

	<!-- Payment Metric Summary Cards -->
	<div class="grid grid-cols-2 gap-3 sm:grid-cols-4">
		<div class="rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
			<span class="text-[11px] font-medium text-gray-400 uppercase tracking-wider">Total Collected</span>
			<p class="mt-1 text-lg font-bold text-emerald-600 dark:text-emerald-400">{formatCurrency(totalCollected)}</p>
			<p class="text-[10px] text-gray-400">Successful collections</p>
		</div>

		<div class="rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
			<span class="text-[11px] font-medium text-gray-400 uppercase tracking-wider">Gateway Success Rate</span>
			<p class="mt-1 text-lg font-bold text-sky-600 dark:text-sky-400">{successRate}%</p>
			<p class="text-[10px] text-gray-400">{completedCount} of {payments.length} successful</p>
		</div>

		<div class="rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
			<span class="text-[11px] font-medium text-gray-400 uppercase tracking-wider">Average Ticket</span>
			<p class="mt-1 text-lg font-bold text-gray-900 dark:text-white">
				{formatCurrency(completedCount > 0 ? totalCollected / completedCount : 0)}
			</p>
			<p class="text-[10px] text-gray-400">Per payment avg</p>
		</div>

		<div class="rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
			<span class="text-[11px] font-medium text-gray-400 uppercase tracking-wider">Total Transactions</span>
			<p class="mt-1 text-lg font-bold text-gray-900 dark:text-white">{totalRecords.toLocaleString()}</p>
			<p class="text-[10px] text-gray-400">Processed across all gateways</p>
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
						loadPayments();
					}
				}}
				placeholder="Search reference, customer… (Enter)"
				class="w-full rounded-lg border border-gray-200 bg-gray-50 py-1.5 pl-9 pr-3 text-xs text-gray-900 focus:border-emerald-500 focus:bg-white focus:outline-hidden dark:border-gray-800 dark:bg-gray-800 dark:text-white"
			/>
		</div>

		<div class="flex items-center gap-2">
			<!-- Method Filter -->
			<select
				bind:value={methodFilter}
				class="rounded-lg border border-gray-200 bg-white px-2.5 py-1.5 text-xs text-gray-700 focus:border-emerald-500 focus:outline-hidden dark:border-gray-800 dark:bg-gray-800 dark:text-gray-300"
			>
				<option value="all">All Gateways</option>
				<option value="manual">Manual Bank</option>
				<option value="bkash">bKash</option>
				<option value="nagad">Nagad</option>
				<option value="rocket">Rocket</option>
				<option value="piprapay">PipraPay</option>
			</select>

			<!-- Status Filter -->
			<select
				bind:value={statusFilter}
				class="rounded-lg border border-gray-200 bg-white px-2.5 py-1.5 text-xs text-gray-700 focus:border-emerald-500 focus:outline-hidden dark:border-gray-800 dark:bg-gray-800 dark:text-gray-300"
			>
				<option value="all">All Statuses</option>
				<option value="completed">Completed</option>
				<option value="pending">Pending</option>
				<option value="failed">Failed</option>
			</select>

			<button
				type="button"
				onclick={() => {
					currentPage = 1;
					loadPayments();
				}}
				class="rounded-lg bg-gray-900 px-3 py-1.5 text-xs font-semibold text-white hover:bg-gray-800 dark:bg-white dark:text-gray-900 dark:hover:bg-gray-100"
			>
				Filter
			</button>
		</div>
	</div>

	<!-- Payments Table -->
	<div class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-xs dark:border-gray-800 dark:bg-gray-900">
		<div class="overflow-x-auto">
			<table class="w-full text-left text-xs">
				<thead>
					<tr class="border-b border-gray-100 bg-gray-50/50 text-[10px] font-semibold uppercase tracking-wider text-gray-400 dark:border-gray-800 dark:bg-gray-800/40">
						<th class="py-3 px-4">Gateway Reference</th>
						<th class="py-3 px-3">Customer</th>
						<th class="py-3 px-3">Gateway Method</th>
						<th class="py-3 px-3 text-right">Amount</th>
						<th class="py-3 px-3">Status</th>
						<th class="py-3 px-3">Settlement Date</th>
						<th class="py-3 px-4 text-right">Created</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
					{#each payments as p}
						<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30 transition-colors">
							<td class="py-3 px-4 font-mono font-semibold text-gray-900 dark:text-white">
								{p.transaction_id || `DEP-${p.id}`}
							</td>
							<td class="py-3 px-3">
								<p class="font-bold text-gray-900 dark:text-white">{p.customer_name}</p>
								<p class="font-mono text-[10px] text-gray-400">{p.customer_phone}</p>
							</td>
							<td class="py-3 px-3">
								<span class="inline-flex rounded-md bg-gray-100 px-2 py-0.5 text-[10px] font-semibold capitalize text-gray-800 dark:bg-gray-800 dark:text-gray-200">
									{p.method}
								</span>
							</td>
							<td class="py-3 px-3 text-right font-mono font-bold text-emerald-600 dark:text-emerald-400">
								{formatCurrency(p.amount)}
							</td>
							<td class="py-3 px-3">
								<span class="inline-flex rounded-full px-2 py-0.5 text-[10px] font-semibold {p.status === 'completed' ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300' : p.status === 'pending' ? 'bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-300' : 'bg-rose-100 text-rose-800 dark:bg-rose-950 dark:text-rose-300'}">
									{p.status}
								</span>
							</td>
							<td class="py-3 px-3 font-mono text-[11px] text-gray-500">
								{p.paid_at || '—'}
							</td>
							<td class="py-3 px-4 text-right font-mono text-[11px] text-gray-400">
								{p.created_at}
							</td>
						</tr>
					{:else}
						<tr>
							<td colspan="7" class="py-8 text-center text-xs text-gray-400">
								{loading ? 'Loading payments…' : 'No payments found matching filters.'}
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
								loadPayments();
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
								loadPayments();
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
