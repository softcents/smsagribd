<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';

	interface ExecutiveReport {
		today_sms: number;
		month_sms: number;
		gross_revenue: number;
		net_profit: number;
		delivered_pct: number;
		failed_pct: number;
	}

	let data = $state<ExecutiveReport | null>(null);
	let loading = $state(true);

	onMount(async () => {
		try {
			const res = await supermanApi.get<{ data: ExecutiveReport }>('/api/v1/superman/sps/reports');
			data = res.data ?? null;
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	});
</script>

<div class="space-y-6">
	<!-- Page Header -->
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-xl font-bold text-gray-900 dark:text-white">Service Provider (SP) Performance Reports</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
				Aggregated SMS volumes, operator delivery ratios, and financial yield across all reseller tiers.
			</p>
		</div>
		<div>
			<a
				href="/superman/reports/reseller"
				class="inline-flex items-center gap-2 rounded-lg bg-emerald-600 px-3 py-1.5 text-xs font-semibold text-white shadow-xs hover:bg-emerald-700 transition"
			>
				Detailed Reseller Report
			</a>
		</div>
	</div>

	{#if loading}
		<div class="p-8 text-center text-xs text-gray-400">Loading performance report metrics...</div>
	{:else if data}
		<div class="grid grid-cols-1 md:grid-cols-4 gap-4">
			<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-5 shadow-xs">
				<span class="text-xs font-medium text-gray-500 uppercase tracking-wider">Today's Reseller Volume</span>
				<p class="text-3xl font-bold text-gray-900 dark:text-white mt-1">{data.today_sms.toLocaleString()}</p>
				<span class="text-[11px] text-gray-400 mt-1 block">Live dispatches</span>
			</div>
			<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-5 shadow-xs">
				<span class="text-xs font-medium text-blue-600 uppercase tracking-wider">30-Day Dispatches</span>
				<p class="text-3xl font-bold text-blue-600 mt-1">{data.month_sms.toLocaleString()}</p>
				<span class="text-[11px] text-gray-400 mt-1 block">Monthly wholesale run-rate</span>
			</div>
			<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-5 shadow-xs">
				<span class="text-xs font-medium text-emerald-600 uppercase tracking-wider">Gross Billed</span>
				<p class="text-3xl font-bold text-emerald-600 mt-1">৳{data.gross_revenue.toFixed(2)}</p>
				<span class="text-[11px] text-emerald-600 mt-1 block">Net Profit: ৳{data.net_profit.toFixed(2)}</span>
			</div>
			<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-5 shadow-xs">
				<span class="text-xs font-medium text-purple-600 uppercase tracking-wider">Delivery Success</span>
				<p class="text-3xl font-bold text-purple-600 mt-1">{data.delivered_pct}%</p>
				<span class="text-[11px] text-gray-400 mt-1 block">Failed/Rejected: {data.failed_pct}%</span>
			</div>
		</div>
	{/if}
</div>

