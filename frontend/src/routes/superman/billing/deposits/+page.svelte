<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface Deposit {
		id: number;
		customer: string;
		amount: number;
		gateway: string | null;
		transaction_id: string | null;
		status: string;
		created_at: string;
	}

	let deposits = $state<Deposit[]>([]);
	let loading = $state(true);
	let search = $state('');
	let statusFilter = $state('all');
	let methodFilter = $state('all');

	// Selected deposit for detail modal
	let selectedDeposit = $state<Deposit | null>(null);
	let actionLoading = $state<number | null>(null);

	// Pagination
	let currentPage = $state(1);
	let perPage = $state(25);
	let totalRecords = $state(0);
	let totalPages = $derived(Math.ceil(totalRecords / perPage) || 1);

	// Summary stats
	let completedTotal = $derived(
		deposits.filter((d) => d.status === 'completed').reduce((acc, d) => acc + d.amount, 0)
	);
	let pendingCount = $derived(
		deposits.filter((d) => d.status === 'pending').length
	);
	let pendingTotal = $derived(
		deposits.filter((d) => d.status === 'pending').reduce((acc, d) => acc + d.amount, 0)
	);

	async function loadDeposits() {
		loading = true;
		try {
			const query = new URLSearchParams({
				page: String(currentPage),
				per_page: String(perPage)
			});
			if (search) query.set('search', search);

			const res = await supermanApi.get<{
				data: Deposit[];
				meta: { page: number; per_page: number; total: number };
			}>(`/api/v1/superman/deposits?${query.toString()}`);

			deposits = res.data || [];
			totalRecords = res.meta?.total || 0;
		} catch (err) {
			console.error('Failed to load deposits:', err);
		} finally {
			loading = false;
		}
	}

	async function approveDeposit(deposit: Deposit) {
		if (!window.confirm(`Approve deposit of ৳${deposit.amount.toFixed(2)} for ${deposit.customer}? This will immediately credit their balance.`)) {
			return;
		}

		actionLoading = deposit.id;
		try {
			await supermanApi.post(`/api/v1/superman/deposits/${deposit.id}/approve`, {});
			await loadDeposits();
		} catch (err: unknown) {
			window.alert(err instanceof Error ? err.message : 'Failed to approve deposit.');
		} finally {
			actionLoading = null;
		}
	}

	async function rejectDeposit(deposit: Deposit) {
		if (!window.confirm(`Reject deposit #${deposit.id} of ৳${deposit.amount.toFixed(2)} for ${deposit.customer}?`)) {
			return;
		}

		actionLoading = deposit.id;
		try {
			await supermanApi.post(`/api/v1/superman/deposits/${deposit.id}/reject`, {});
			await loadDeposits();
		} catch (err: unknown) {
			window.alert(err instanceof Error ? err.message : 'Failed to reject deposit.');
		} finally {
			actionLoading = null;
		}
	}

	onMount(() => {
		// Read initial status filter from URL if present (e.g. ?status=pending)
		const paramStatus = $page.url.searchParams.get('status');
		if (paramStatus) {
			statusFilter = paramStatus;
		}
		loadDeposits();
	});

	let filteredDeposits = $derived(
		deposits.filter((d) => {
			if (statusFilter !== 'all' && d.status !== statusFilter) return false;
			if (methodFilter !== 'all') {
				const gw = (d.gateway || 'manual').toLowerCase();
				if (!gw.includes(methodFilter.toLowerCase())) return false;
			}
			return true;
		})
	);

	function formatCurrency(val: number): string {
		return '৳' + (val || 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
	}

	const exportColumns: ExportColumn<Deposit>[] = [
		{ header: 'Deposit ID', getValue: (d) => `#${d.id}` },
		{ header: 'Customer', getValue: (d) => d.customer },
		{ header: 'Amount (BDT)', getValue: (d) => `BDT ${d.amount.toFixed(2)}` },
		{ header: 'Method', getValue: (d) => d.gateway || 'Manual' },
		{ header: 'Trx ID / Ref', getValue: (d) => d.transaction_id || '—' },
		{ header: 'Status', getValue: (d) => d.status.toUpperCase() },
		{ header: 'Date', getValue: (d) => new Date(d.created_at).toLocaleString() }
	];
</script>

<svelte:head>
	<title>Customer Deposits — SMS Whiteinfo Administrator</title>
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
				<span class="text-xs font-semibold text-gray-900 dark:text-white">Deposits</span>
			</div>
			<h1 class="mt-1 text-xl font-bold tracking-tight text-gray-900 dark:text-white sm:text-2xl">
				Customer Deposits & Top-ups
			</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400">
				Manage inbound online payments, merchant gateways, manual bank approvals, and credit transactions.
			</p>
		</div>

		<div class="flex items-center gap-2">
			<button
				type="button"
				onclick={loadDeposits}
				class="inline-flex items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs font-semibold text-gray-700 transition hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300 dark:hover:bg-gray-800"
			>
				<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0 3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182m0-4.991v4.99" />
				</svg>
				Refresh
			</button>
			<ExportDropdown
				title="Customer Deposits"
				columns={exportColumns}
				getData={() => filteredDeposits}
				disabled={filteredDeposits.length === 0}
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
				href="/superman/billing/transactions"
				class="inline-flex items-center gap-1.5 rounded-lg bg-emerald-600 px-3.5 py-2 text-xs font-semibold text-white shadow-xs hover:bg-emerald-500"
			>
				<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
				</svg>
				Ledger Transactions
			</a>
		</div>
	</div>

	<!-- Summary Cards -->
	<div class="grid grid-cols-2 gap-3 sm:grid-cols-4">
		<div class="rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
			<span class="text-[11px] font-medium text-gray-400 uppercase tracking-wider">Completed Total</span>
			<p class="mt-1 text-lg font-bold text-emerald-600 dark:text-emerald-400">{formatCurrency(completedTotal)}</p>
			<p class="text-[10px] text-gray-400">Successfully credited</p>
		</div>

		<div class="rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
			<span class="text-[11px] font-medium text-gray-400 uppercase tracking-wider">Pending Approval</span>
			<p class="mt-1 text-lg font-bold text-amber-600 dark:text-amber-400">{formatCurrency(pendingTotal)}</p>
			<p class="text-[10px] text-amber-600 dark:text-amber-400">{pendingCount} request{pendingCount === 1 ? '' : 's'} awaiting review</p>
		</div>

		<div class="rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
			<span class="text-[11px] font-medium text-gray-400 uppercase tracking-wider">Total Records</span>
			<p class="mt-1 text-lg font-bold text-gray-900 dark:text-white">{totalRecords.toLocaleString()}</p>
			<p class="text-[10px] text-gray-400">All deposit history</p>
		</div>

		<div class="rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
			<span class="text-[11px] font-medium text-gray-400 uppercase tracking-wider">Default Currency</span>
			<p class="mt-1 text-lg font-bold text-gray-900 dark:text-white">BDT (৳)</p>
			<p class="text-[10px] text-gray-400">Bangladeshi Taka</p>
		</div>
	</div>

	<!-- Filter & Search Toolbar -->
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
						loadDeposits();
					}
				}}
				placeholder="Search customer, transaction ID… (Enter)"
				class="w-full rounded-lg border border-gray-200 bg-gray-50 py-1.5 pl-9 pr-3 text-xs text-gray-900 focus:border-emerald-500 focus:bg-white focus:outline-hidden dark:border-gray-800 dark:bg-gray-800 dark:text-white"
			/>
		</div>

		<div class="flex flex-wrap items-center gap-2">
			<!-- Status Filter -->
			<select
				bind:value={statusFilter}
				class="rounded-lg border border-gray-200 bg-white px-2.5 py-1.5 text-xs text-gray-700 focus:border-emerald-500 focus:outline-hidden dark:border-gray-800 dark:bg-gray-800 dark:text-gray-300"
			>
				<option value="all">All Statuses</option>
				<option value="completed">Completed</option>
				<option value="pending">Pending Review</option>
				<option value="failed">Failed / Rejected</option>
			</select>

			<!-- Method Filter -->
			<select
				bind:value={methodFilter}
				class="rounded-lg border border-gray-200 bg-white px-2.5 py-1.5 text-xs text-gray-700 focus:border-emerald-500 focus:outline-hidden dark:border-gray-800 dark:bg-gray-800 dark:text-gray-300"
			>
				<option value="all">All Channels</option>
				<option value="manual">Manual Bank</option>
				<option value="bkash">bKash</option>
				<option value="nagad">Nagad</option>
				<option value="rocket">Rocket</option>
				<option value="piprapay">PipraPay</option>
			</select>

			<button
				type="button"
				onclick={() => {
					currentPage = 1;
					loadDeposits();
				}}
				class="rounded-lg bg-gray-900 px-3 py-1.5 text-xs font-semibold text-white hover:bg-gray-800 dark:bg-white dark:text-gray-900 dark:hover:bg-gray-100"
			>
				Filter
			</button>
		</div>
	</div>

	<!-- Deposits Table -->
	<div class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-xs dark:border-gray-800 dark:bg-gray-900">
		<div class="overflow-x-auto">
			<table class="w-full text-left text-xs">
				<thead>
					<tr class="border-b border-gray-100 bg-gray-50/50 text-[10px] font-semibold uppercase tracking-wider text-gray-400 dark:border-gray-800 dark:bg-gray-800/40">
						<th class="py-3 px-4">Deposit ID</th>
						<th class="py-3 px-3">Customer</th>
						<th class="py-3 px-3 text-right">Amount</th>
						<th class="py-3 px-3">Gateway & Channel</th>
						<th class="py-3 px-3">Transaction Reference</th>
						<th class="py-3 px-3">Status</th>
						<th class="py-3 px-3 text-right">Date & Time</th>
						<th class="py-3 px-4 text-right">Actions</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
					{#each filteredDeposits as d}
						<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30 transition-colors">
							<td class="py-3 px-4 font-mono font-semibold text-gray-500">
								#{d.id}
							</td>
							<td class="py-3 px-3">
								<p class="font-bold text-gray-900 dark:text-white">{d.customer}</p>
							</td>
							<td class="py-3 px-3 text-right font-mono font-bold text-emerald-600 dark:text-emerald-400">
								{formatCurrency(d.amount)}
							</td>
							<td class="py-3 px-3 capitalize text-gray-700 dark:text-gray-300">
								{d.gateway || 'Manual'}
							</td>
							<td class="py-3 px-3 font-mono text-[11px] text-gray-500">
								{d.transaction_id || '—'}
							</td>
							<td class="py-3 px-3">
								<span class="inline-flex rounded-full px-2 py-0.5 text-[10px] font-semibold {d.status === 'completed' ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300' : d.status === 'pending' ? 'bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-300' : 'bg-rose-100 text-rose-800 dark:bg-rose-950 dark:text-rose-300'}">
									{d.status}
								</span>
							</td>
							<td class="py-3 px-3 text-right font-mono text-[11px] text-gray-400">
								{new Date(d.created_at).toLocaleString([], { dateStyle: 'medium', timeStyle: 'short' })}
							</td>
							<td class="py-3 px-4 text-right">
								<div class="flex items-center justify-end gap-1.5">
									{#if d.status === 'pending'}
										<button
											type="button"
											onclick={() => approveDeposit(d)}
											disabled={actionLoading === d.id}
											class="rounded-lg bg-emerald-600 px-2.5 py-1 text-xs font-semibold text-white shadow-2xs hover:bg-emerald-500 disabled:opacity-50"
										>
											{actionLoading === d.id ? '…' : 'Approve'}
										</button>
										<button
											type="button"
											onclick={() => rejectDeposit(d)}
											disabled={actionLoading === d.id}
											class="rounded-lg border border-rose-200 bg-white px-2 py-1 text-xs font-semibold text-rose-600 hover:bg-rose-50 disabled:opacity-50 dark:border-rose-900/40 dark:bg-gray-800 dark:text-rose-400"
										>
											Reject
										</button>
									{:else}
										<button
											type="button"
											onclick={() => (selectedDeposit = d)}
											class="rounded-lg border border-gray-200 bg-white px-2.5 py-1 text-xs font-medium text-gray-700 hover:bg-gray-50 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-300"
										>
											View
										</button>
									{/if}
								</div>
							</td>
						</tr>
					{:else}
						<tr>
							<td colspan="8" class="py-8 text-center text-xs text-gray-400">
								{loading ? 'Loading deposits…' : 'No deposits found matching criteria.'}
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
								loadDeposits();
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
								loadDeposits();
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

<!-- Modal: View Deposit Details -->
{#if selectedDeposit}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-gray-900/60 backdrop-blur-xs p-4">
		<div class="w-full max-w-md rounded-2xl border border-gray-200 bg-white p-6 shadow-xl dark:border-gray-800 dark:bg-gray-900">
			<div class="flex items-center justify-between border-b border-gray-100 pb-3 dark:border-gray-800">
				<div>
					<h3 class="text-base font-bold text-gray-900 dark:text-white">Deposit #{selectedDeposit.id}</h3>
					<p class="text-xs text-gray-500">Full payment receipt information</p>
				</div>
				<button
					type="button"
					onclick={() => (selectedDeposit = null)}
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
					<span class="font-bold text-gray-900 dark:text-white">{selectedDeposit.customer}</span>
				</div>
				<div class="flex items-center justify-between border-b border-gray-100 pb-2 dark:border-gray-800">
					<span class="text-gray-400">Amount</span>
					<span class="font-mono font-bold text-emerald-600 dark:text-emerald-400 text-sm">
						{formatCurrency(selectedDeposit.amount)}
					</span>
				</div>
				<div class="flex items-center justify-between border-b border-gray-100 pb-2 dark:border-gray-800">
					<span class="text-gray-400">Payment Channel</span>
					<span class="capitalize text-gray-800 dark:text-gray-200">{selectedDeposit.gateway || 'Manual'}</span>
				</div>
				<div class="flex items-center justify-between border-b border-gray-100 pb-2 dark:border-gray-800">
					<span class="text-gray-400">Transaction Reference</span>
					<span class="font-mono text-gray-700 dark:text-gray-300">{selectedDeposit.transaction_id || '—'}</span>
				</div>
				<div class="flex items-center justify-between border-b border-gray-100 pb-2 dark:border-gray-800">
					<span class="text-gray-400">Status</span>
					<span class="font-semibold capitalize text-emerald-600 dark:text-emerald-400">{selectedDeposit.status}</span>
				</div>
				<div class="flex items-center justify-between">
					<span class="text-gray-400">Created At</span>
					<span class="font-mono text-gray-700 dark:text-gray-300">
						{new Date(selectedDeposit.created_at).toLocaleString()}
					</span>
				</div>
			</div>

			<div class="mt-6 flex justify-end">
				<button
					type="button"
					onclick={() => (selectedDeposit = null)}
					class="rounded-lg bg-gray-900 px-4 py-2 text-xs font-semibold text-white hover:bg-gray-800 dark:bg-white dark:text-gray-900"
				>
					Close
				</button>
			</div>
		</div>
	</div>
{/if}
