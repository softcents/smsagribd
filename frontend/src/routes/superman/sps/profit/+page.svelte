<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';

	interface ProfitSummary {
		gross_revenue: number;
		wholesale_cost: number;
		reseller_margin: number;
		margin_percent: number;
	}

	let data = $state<ProfitSummary>({
		gross_revenue: 0,
		wholesale_cost: 0,
		reseller_margin: 0,
		margin_percent: 0
	});
	let loading = $state(true);

	onMount(async () => {
		try {
			const res = await supermanApi.get<{ data: ProfitSummary }>('/api/v1/superman/sps/profit');
			if (res.data) data = res.data;
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
			<h1 class="text-xl font-bold text-gray-900 dark:text-white">Reseller Profit & Wholesale Margins</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
				Real-time revenue, wholesale operator cost, reseller markup spread, and net platform yield.
			</p>
		</div>
	</div>

	<!-- KPI Cards -->
	<div class="grid grid-cols-1 md:grid-cols-4 gap-4">
		<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-5 shadow-xs">
			<span class="text-xs font-medium text-gray-500 uppercase tracking-wider">Gross Client Billed</span>
			<p class="text-3xl font-bold text-gray-900 dark:text-white mt-1">৳{data.gross_revenue.toFixed(2)}</p>
			<span class="text-[11px] text-gray-400 mt-1 block">Total end-user revenue</span>
		</div>
		<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-5 shadow-xs">
			<span class="text-xs font-medium text-amber-600 uppercase tracking-wider">Carrier Wholesale Cost</span>
			<p class="text-3xl font-bold text-amber-600 mt-1">৳{data.wholesale_cost.toFixed(2)}</p>
			<span class="text-[11px] text-gray-400 mt-1 block">Telco settlement outlays</span>
		</div>
		<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-5 shadow-xs">
			<span class="text-xs font-medium text-emerald-600 uppercase tracking-wider">Reseller Margin Spread</span>
			<p class="text-3xl font-bold text-emerald-600 mt-1">৳{data.reseller_margin.toFixed(2)}</p>
			<span class="text-[11px] text-emerald-600 mt-1 block">Yielded to SP partners</span>
		</div>
		<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-5 shadow-xs">
			<span class="text-xs font-medium text-purple-600 uppercase tracking-wider">Average Markup</span>
			<p class="text-3xl font-bold text-purple-600 mt-1">{data.margin_percent}%</p>
			<span class="text-[11px] text-gray-400 mt-1 block">Effective retail markup</span>
		</div>
	</div>
</div>

