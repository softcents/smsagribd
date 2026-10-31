<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';

	interface ExecutiveSummary {
		gross_revenue: number;
		total_operator_cost: number;
		net_gross_profit: number;
		profit_margin_percent: number;
	}

	let data = $state<ExecutiveSummary | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);

	async function loadAnalytics() {
		loading = true;
		error = null;
		try {
			const res = await supermanApi.get<{ data: ExecutiveSummary }>('/api/v1/superman/reports/executive');
			data = res.data;
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Failed to load financial analytics.';
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		void loadAnalytics();
	});
</script>

<svelte:head>
	<title>Financial Analytics — Superman Administrator</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">Financial & Capital Analytics</h1>
			<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
				Platform revenue aggregation, gateway cost tracking, and net margin performance.
			</p>
		</div>
		<button onclick={loadAnalytics} class="rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs font-medium text-gray-700 shadow-xs hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300">
			Refresh Analytics
		</button>
	</div>

	{#if loading}
		<div class="py-20 text-center text-xs text-gray-500">Compiling financial analytics...</div>
	{:else if error}
		<div class="rounded-xl border border-rose-200 bg-rose-50 p-6 text-center text-xs text-rose-700">{error}</div>
	{:else if data}
		<div class="grid grid-cols-1 gap-4 sm:grid-cols-3">
			<div class="rounded-xl border border-gray-200 bg-white p-5 shadow-xs dark:border-gray-800 dark:bg-gray-900">
				<span class="text-[11px] font-semibold text-gray-500 uppercase tracking-wider">Gross Billed Revenue</span>
				<p class="mt-2 text-3xl font-black text-gray-950 dark:text-white">৳{data.gross_revenue.toFixed(2)}</p>
				<p class="mt-1 text-xs text-gray-400">Total customer deductions</p>
			</div>
			<div class="rounded-xl border border-gray-200 bg-white p-5 shadow-xs dark:border-gray-800 dark:bg-gray-900">
				<span class="text-[11px] font-semibold text-gray-500 uppercase tracking-wider">Gateway Operator Expense</span>
				<p class="mt-2 text-3xl font-black text-rose-600">৳{data.total_operator_cost.toFixed(2)}</p>
				<p class="mt-1 text-xs text-gray-400">Settled to mobile carriers</p>
			</div>
			<div class="rounded-xl border border-gray-200 bg-white p-5 shadow-xs dark:border-gray-800 dark:bg-gray-900">
				<span class="text-[11px] font-semibold text-gray-500 uppercase tracking-wider">Net Gross Profit</span>
				<p class="mt-2 text-3xl font-black text-emerald-600">৳{data.net_gross_profit.toFixed(2)}</p>
				<p class="mt-1 text-xs text-emerald-600 font-semibold">{data.profit_margin_percent.toFixed(1)}% Platform Margin</p>
			</div>
		</div>
	{/if}
</div>

