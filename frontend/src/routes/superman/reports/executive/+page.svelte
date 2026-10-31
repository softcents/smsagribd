<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';

	interface ExecutiveSummary {
		total_dispatches: number;
		delivered_count: number;
		delivery_rate_percent: number;
		failed_count: number;
		gross_revenue: number;
		total_operator_cost: number;
		net_gross_profit: number;
		profit_margin_percent: number;
		active_customers: number;
		active_senders: number;
	}

	let data = $state<ExecutiveSummary | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);

	async function loadExecutiveReport() {
		loading = true;
		error = null;
		try {
			const res = await supermanApi.get<{ data: ExecutiveSummary }>('/api/v1/superman/reports/executive');
			data = res.data;
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Failed to load executive report.';
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		void loadExecutiveReport();
	});
</script>

<svelte:head>
	<title>Executive Report — Superman Administrator</title>
</svelte:head>

<div class="space-y-6">
	<!-- Top Bar -->
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">Executive Platform Overview</h1>
			<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
				High-level business performance metrics, total gross revenue, gateway operator expense, and net profit margins.
			</p>
		</div>
		<button
			onclick={loadExecutiveReport}
			class="inline-flex items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs font-medium text-gray-700 shadow-xs hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300 dark:hover:bg-gray-800"
		>
			Refresh Metrics
		</button>
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-20 text-xs text-gray-500">
			<svg class="mr-2 h-4 w-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path></svg>
			Compiling executive telemetry...
		</div>
	{:else if error}
		<div class="rounded-xl border border-rose-200 bg-rose-50 p-6 text-center text-xs text-rose-700 dark:border-rose-900 dark:bg-rose-950/30 dark:text-rose-300">
			{error}
		</div>
	{:else if data}
		<!-- Primary Financial KPIs -->
		<div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
			<div class="rounded-xl border border-gray-200 bg-white p-5 shadow-xs dark:border-gray-800 dark:bg-gray-900">
				<span class="text-[11px] font-semibold text-gray-500 uppercase tracking-wider">Gross SMS Revenue</span>
				<p class="mt-2 text-2xl font-black text-gray-950 dark:text-white">৳{data.gross_revenue.toFixed(2)}</p>
				<p class="mt-1 text-[11px] text-emerald-600 font-medium">Billed across all dispatches</p>
			</div>

			<div class="rounded-xl border border-gray-200 bg-white p-5 shadow-xs dark:border-gray-800 dark:bg-gray-900">
				<span class="text-[11px] font-semibold text-gray-500 uppercase tracking-wider">Operator Carrier Cost</span>
				<p class="mt-2 text-2xl font-black text-rose-600">৳{data.total_operator_cost.toFixed(2)}</p>
				<p class="mt-1 text-[11px] text-gray-500">Upstream gateway settlement</p>
			</div>

			<div class="rounded-xl border border-gray-200 bg-white p-5 shadow-xs dark:border-gray-800 dark:bg-gray-900">
				<span class="text-[11px] font-semibold text-gray-500 uppercase tracking-wider">Net Gross Profit</span>
				<p class="mt-2 text-2xl font-black text-emerald-600">৳{data.net_gross_profit.toFixed(2)}</p>
				<p class="mt-1 text-[11px] text-emerald-600 font-medium">{data.profit_margin_percent.toFixed(1)}% Gross Margin</p>
			</div>

			<div class="rounded-xl border border-gray-200 bg-white p-5 shadow-xs dark:border-gray-800 dark:bg-gray-900">
				<span class="text-[11px] font-semibold text-gray-500 uppercase tracking-wider">Delivery Success Rate</span>
				<p class="mt-2 text-2xl font-black text-blue-600">{data.delivery_rate_percent.toFixed(1)}%</p>
				<p class="mt-1 text-[11px] text-gray-500">{data.delivered_count} delivered / {data.total_dispatches} total</p>
			</div>
		</div>

		<!-- Operational Metrics Breakdown -->
		<div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
			<!-- Traffic Distribution Card -->
			<div class="rounded-xl border border-gray-200 bg-white p-6 shadow-xs dark:border-gray-800 dark:bg-gray-900 space-y-4">
				<h3 class="text-sm font-bold text-gray-950 dark:text-white">Traffic & Dispatch Breakdown</h3>
				<div class="space-y-3">
					<div class="flex items-center justify-between text-xs">
						<span class="text-gray-600 dark:text-gray-400">Total SMS Dispatched</span>
						<span class="font-bold text-gray-900 dark:text-white">{data.total_dispatches}</span>
					</div>
					<div class="flex items-center justify-between text-xs">
						<span class="text-gray-600 dark:text-gray-400">Delivered Handset Receipts</span>
						<span class="font-semibold text-emerald-600">{data.delivered_count}</span>
					</div>
					<div class="flex items-center justify-between text-xs">
						<span class="text-gray-600 dark:text-gray-400">Failed / Expired / Blocked</span>
						<span class="font-semibold text-rose-600">{data.failed_count}</span>
					</div>
					<div class="flex items-center justify-between text-xs">
						<span class="text-gray-600 dark:text-gray-400">Active Customer Accounts</span>
						<span class="font-bold text-gray-900 dark:text-white">{data.active_customers}</span>
					</div>
					<div class="flex items-center justify-between text-xs">
						<span class="text-gray-600 dark:text-gray-400">Active Approved Sender IDs</span>
						<span class="font-bold text-gray-900 dark:text-white">{data.active_senders}</span>
					</div>
				</div>
			</div>

			<!-- Quick Navigation Link Card -->
			<div class="rounded-xl border border-gray-200 bg-white p-6 shadow-xs dark:border-gray-800 dark:bg-gray-900 space-y-4">
				<h3 class="text-sm font-bold text-gray-950 dark:text-white">Granular Report Drilldowns</h3>
				<p class="text-xs text-gray-500">Access targeted sub-reports for detailed reconciliation and ledger audits:</p>
				<div class="grid grid-cols-2 gap-3 pt-2">
					<a href="/superman/reports/sms" class="flex flex-col rounded-lg border border-gray-100 bg-gray-50 p-3 hover:border-emerald-200 hover:bg-emerald-50/50 transition dark:border-gray-800 dark:bg-gray-800/40">
						<span class="text-xs font-bold text-gray-900 dark:text-white">SMS Reports</span>
						<span class="text-[10px] text-gray-500">Full dispatch audit table</span>
					</a>
					<a href="/superman/reports/financial" class="flex flex-col rounded-lg border border-gray-100 bg-gray-50 p-3 hover:border-emerald-200 hover:bg-emerald-50/50 transition dark:border-gray-800 dark:bg-gray-800/40">
						<span class="text-xs font-bold text-gray-900 dark:text-white">Financial Reports</span>
						<span class="text-[10px] text-gray-500">Transactions & payments</span>
					</a>
					<a href="/superman/reports/customer" class="flex flex-col rounded-lg border border-gray-100 bg-gray-50 p-3 hover:border-emerald-200 hover:bg-emerald-50/50 transition dark:border-gray-800 dark:bg-gray-800/40">
						<span class="text-xs font-bold text-gray-900 dark:text-white">Customer Reports</span>
						<span class="text-[10px] text-gray-500">Client balances & volumes</span>
					</a>
					<a href="/superman/reports/provider" class="flex flex-col rounded-lg border border-gray-100 bg-gray-50 p-3 hover:border-emerald-200 hover:bg-emerald-50/50 transition dark:border-gray-800 dark:bg-gray-800/40">
						<span class="text-xs font-bold text-gray-900 dark:text-white">Provider Reports</span>
						<span class="text-[10px] text-gray-500">Gateways & carrier costs</span>
					</a>
				</div>
			</div>
		</div>
	{/if}
</div>

