<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface UsageCdrItem {
		id: number;
		customer_name: string;
		customer_phone: string;
		sender_name: string;
		recipient: string;
		operator_name: string;
		cost: number;
		gateway_cost: number;
		margin: number;
		status: string;
		created_at: string;
	}

	let cdrList = $state<UsageCdrItem[]>([]);
	let loading = $state(true);
	let search = $state('');
	let statusFilter = $state('all');

	// Pagination
	let currentPage = $state(1);
	let perPage = $state(25);
	let totalRecords = $state(0);
	let totalPages = $derived(Math.ceil(totalRecords / perPage) || 1);

	// Calculated totals
	let totalBilled = $derived(cdrList.reduce((acc, c) => acc + c.cost, 0));
	let totalCost = $derived(cdrList.reduce((acc, c) => acc + c.gateway_cost, 0));
	let totalMargin = $derived(totalBilled - totalCost);
	let marginPercent = $derived(totalBilled > 0 ? (totalMargin / totalBilled) * 100 : 0);

	async function loadCDR() {
		loading = true;
		try {
			const query = new URLSearchParams({
				page: String(currentPage),
				per_page: String(perPage)
			});
			if (search) query.set('search', search);
			if (statusFilter !== 'all') query.set('status', statusFilter);

			const res = await supermanApi.get<{
				data: UsageCdrItem[];
				meta: { page: number; per_page: number; total: number };
			}>(`/api/v1/superman/billing/usage-cdr?${query.toString()}`);

			cdrList = res.data || [];
			totalRecords = res.meta?.total || 0;
		} catch (err) {
			console.error('Failed to load Usage / CDR records:', err);
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		loadCDR();
	});

	function formatRate(val: number): string {
		return '৳' + (val || 0).toFixed(4);
	}
	function formatCurrency(val: number): string {
		return '৳' + (val || 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
	}
	const exportColumns: ExportColumn<UsageCdrItem>[] = [
		{ header: 'ID', getValue: (c) => c.id },
		{ header: 'Customer', getValue: (c) => c.customer_name },
		{ header: 'Phone', getValue: (c) => c.customer_phone },
		{ header: 'Sender ID', getValue: (c) => c.sender_name },
		{ header: 'Recipient', getValue: (c) => c.recipient },
		{ header: 'Operator', getValue: (c) => c.operator_name },
		{ header: 'Billed (BDT)', getValue: (c) => `BDT ${c.cost.toFixed(4)}` },
		{ header: 'Cost (BDT)', getValue: (c) => `BDT ${c.gateway_cost.toFixed(4)}` },
		{ header: 'Margin (BDT)', getValue: (c) => `BDT ${c.margin.toFixed(4)}` },
		{ header: 'Status', getValue: (c) => c.status },
		{ header: 'Timestamp', getValue: (c) => c.created_at }
	];
</script>

<svelte:head>
	<title>Usage & CDR — SMS Whiteinfo Administrator</title>
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
				<span class="text-xs font-semibold text-gray-900 dark:text-white">Usage & CDR</span>
			</div>
			<h1 class="mt-1 text-xl font-bold tracking-tight text-gray-900 dark:text-white sm:text-2xl">
				Usage & Call Detail Records (CDR)
			</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400">
				Real-time message transmission audit, retail billing vs wholesale operator cost margins, and delivery logs.
			</p>
		</div>

		<div class="flex items-center gap-2">
			<ExportDropdown
				title="Usage and Call Detail Records CDR"
				columns={exportColumns}
				getData={() => cdrList}
				disabled={cdrList.length === 0}
			/>
			<a
				href="/superman/billing"
				class="inline-flex items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs font-semibold text-gray-700 transition hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300 dark:hover:bg-gray-800"
			>
				<svg class="h-4 w-4 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 17l-5-5m0 0l5-5m-5 5h12" />
				</svg>
				Billing Dashboard
			</a>
			<a
				href="/superman/billing/reports"
				class="inline-flex items-center gap-1.5 rounded-lg bg-emerald-600 px-3.5 py-2 text-xs font-semibold text-white shadow-xs hover:bg-emerald-500"
			>
				<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
				</svg>
				Financial Reports
			</a>
		</div>
	</div>

	<!-- Metric Summary Cards -->
	<div class="grid grid-cols-2 gap-3 sm:grid-cols-4">
		<div class="rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
			<span class="text-[11px] font-medium text-gray-400 uppercase tracking-wider">Page Revenue</span>
			<p class="mt-1 text-lg font-bold text-gray-900 dark:text-white">{formatCurrency(totalBilled)}</p>
			<p class="text-[10px] text-gray-400">Charged to customers</p>
		</div>

		<div class="rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
			<span class="text-[11px] font-medium text-gray-400 uppercase tracking-wider">Page Operator Cost</span>
			<p class="mt-1 text-lg font-bold text-rose-600 dark:text-rose-400">{formatCurrency(totalCost)}</p>
			<p class="text-[10px] text-rose-400">Carrier interconnect fees</p>
		</div>

		<div class="rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
			<span class="text-[11px] font-medium text-gray-400 uppercase tracking-wider">Gross Profit Margin</span>
			<p class="mt-1 text-lg font-bold text-emerald-600 dark:text-emerald-400">{formatCurrency(totalMargin)}</p>
			<p class="text-[10px] font-semibold text-emerald-600 dark:text-emerald-400">{marginPercent.toFixed(1)}% profit</p>
		</div>

		<div class="rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
			<span class="text-[11px] font-medium text-gray-400 uppercase tracking-wider">Total CDR Logged</span>
			<p class="mt-1 text-lg font-bold text-gray-900 dark:text-white">{totalRecords.toLocaleString()}</p>
			<p class="text-[10px] text-gray-400">Total SMS transactions</p>
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
						loadCDR();
					}
				}}
				placeholder="Search customer, recipient, sender… (Enter)"
				class="w-full rounded-lg border border-gray-200 bg-gray-50 py-1.5 pl-9 pr-3 text-xs text-gray-900 focus:border-emerald-500 focus:bg-white focus:outline-hidden dark:border-gray-800 dark:bg-gray-800 dark:text-white"
			/>
		</div>

		<div class="flex items-center gap-2">
			<select
				bind:value={statusFilter}
				class="rounded-lg border border-gray-200 bg-white px-2.5 py-1.5 text-xs text-gray-700 focus:border-emerald-500 focus:outline-hidden dark:border-gray-800 dark:bg-gray-800 dark:text-gray-300"
			>
				<option value="all">All Delivery Statuses</option>
				<option value="delivered">Delivered</option>
				<option value="sent">Sent (Pending DLR)</option>
				<option value="pending">Queued</option>
				<option value="failed">Failed / Undelivered</option>
			</select>

			<button
				type="button"
				onclick={() => {
					currentPage = 1;
					loadCDR();
				}}
				class="rounded-lg bg-gray-900 px-3 py-1.5 text-xs font-semibold text-white hover:bg-gray-800 dark:bg-white dark:text-gray-900 dark:hover:bg-gray-100"
			>
				Filter
			</button>
		</div>
	</div>

	<!-- Usage / CDR Table -->
	<div class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-xs dark:border-gray-800 dark:bg-gray-900">
		<div class="overflow-x-auto">
			<table class="w-full text-left text-xs">
				<thead>
					<tr class="border-b border-gray-100 bg-gray-50/50 text-[10px] font-semibold uppercase tracking-wider text-gray-400 dark:border-gray-800 dark:bg-gray-800/40">
						<th class="py-3 px-4">Msg ID</th>
						<th class="py-3 px-3">Customer</th>
						<th class="py-3 px-3">Sender ID</th>
						<th class="py-3 px-3">Recipient</th>
						<th class="py-3 px-3">Operator</th>
						<th class="py-3 px-3 text-right">Retail Rate</th>
						<th class="py-3 px-3 text-right">Gateway Cost</th>
						<th class="py-3 px-3 text-right">Profit Margin</th>
						<th class="py-3 px-3">Status</th>
						<th class="py-3 px-4 text-right">Time</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
					{#each cdrList as cdr}
						<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30 transition-colors">
							<td class="py-3 px-4 font-mono font-medium text-gray-500">
								#{cdr.id}
							</td>
							<td class="py-3 px-3">
								<p class="font-bold text-gray-900 dark:text-white">{cdr.customer_name}</p>
								<p class="font-mono text-[10px] text-gray-400">{cdr.customer_phone}</p>
							</td>
							<td class="py-3 px-3 font-semibold text-gray-800 dark:text-gray-200">
								{cdr.sender_name}
							</td>
							<td class="py-3 px-3 font-mono font-semibold text-gray-800 dark:text-gray-200">
								{cdr.recipient}
							</td>
							<td class="py-3 px-3">
								<span class="inline-flex rounded-md bg-gray-100 px-2 py-0.5 text-[10px] font-semibold text-gray-800 dark:bg-gray-800 dark:text-gray-200">
									{cdr.operator_name}
								</span>
							</td>
							<td class="py-3 px-3 text-right font-mono font-bold text-gray-900 dark:text-white">
								{formatRate(cdr.cost)}
							</td>
							<td class="py-3 px-3 text-right font-mono text-rose-600 dark:text-rose-400">
								{formatRate(cdr.gateway_cost)}
							</td>
							<td class="py-3 px-3 text-right font-mono font-bold {cdr.margin >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-rose-600 dark:text-rose-400'}">
								{formatRate(cdr.margin)}
							</td>
							<td class="py-3 px-3">
								<span class="inline-flex rounded-full px-2 py-0.5 text-[10px] font-semibold {cdr.status === 'delivered' ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300' : cdr.status === 'sent' ? 'bg-sky-100 text-sky-800 dark:bg-sky-950 dark:text-sky-300' : cdr.status === 'pending' ? 'bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-300' : 'bg-rose-100 text-rose-800 dark:bg-rose-950 dark:text-rose-300'}">
									{cdr.status}
								</span>
							</td>
							<td class="py-3 px-4 text-right font-mono text-[11px] text-gray-400">
								{cdr.created_at}
							</td>
						</tr>
					{:else}
						<tr>
							<td colspan="10" class="py-8 text-center text-xs text-gray-400">
								{loading ? 'Loading Usage / CDR records…' : 'No records found.'}
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
								loadCDR();
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
								loadCDR();
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
