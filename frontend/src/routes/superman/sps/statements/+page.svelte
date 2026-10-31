<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface StatementItem {
		date: string;
		sent: number;
		delivered: number;
		failed: number;
		cost: number;
		profit: number;
	}

	let rows = $state<StatementItem[]>([]);
	let loading = $state(true);

	const exportColumns: ExportColumn<StatementItem>[] = [
		{ header: 'Billing Cycle / Date', getValue: (s) => s.date },
		{ header: 'Total Dispatches', getValue: (s) => s.sent },
		{ header: 'Delivered', getValue: (s) => s.delivered },
		{ header: 'Failed', getValue: (s) => s.failed },
		{ header: 'Delivery Rate (%)', getValue: (s) => `${s.sent > 0 ? ((s.delivered / s.sent) * 100).toFixed(1) : 0}%` },
		{ header: 'Billed Volume (BDT)', getValue: (s) => `BDT ${s.cost.toFixed(2)}` },
		{ header: 'Net Margin (BDT)', getValue: (s) => `BDT ${s.profit.toFixed(2)}` }
	];

	onMount(async () => {
		try {
			const res = await supermanApi.get<{ data: StatementItem[] }>('/api/v1/superman/sps/statements');
			rows = res.data ?? [];
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
			<h1 class="text-xl font-bold text-gray-900 dark:text-white">Reseller Monthly Statements & Ledger</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
				Aggregated financial statements, transaction debit/credit journals, and carrier cost audit reports.
			</p>
		</div>
		<ExportDropdown
			title="Reseller Monthly Statements and Ledger"
			columns={exportColumns}
			getData={() => rows}
			disabled={rows.length === 0}
		/>
	</div>

	<!-- Statements Table -->
	<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl shadow-xs overflow-hidden">
		<div class="overflow-x-auto">
			<table class="w-full text-left text-xs">
				<thead class="bg-gray-50/75 dark:bg-gray-800/50 text-gray-500 uppercase tracking-wider font-semibold border-b border-gray-200 dark:border-gray-800">
					<tr>
						<th class="px-5 py-3">Billing Cycle / Date</th>
						<th class="px-5 py-3">Total Dispatches</th>
						<th class="px-5 py-3">Delivered</th>
						<th class="px-5 py-3">Delivery Rate</th>
						<th class="px-5 py-3">Billed Volume (BDT)</th>
						<th class="px-5 py-3">Reseller Profit</th>
						<th class="px-5 py-3 text-right">Action</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800 text-gray-700 dark:text-gray-300">
					{#if loading}
						<tr>
							<td colspan="7" class="px-5 py-8 text-center text-gray-400">Loading monthly ledger statements...</td>
						</tr>
					{:else if rows.length === 0}
						<tr>
							<td colspan="7" class="px-5 py-8 text-center text-gray-400">No statement history found.</td>
						</tr>
					{:else}
						{#each rows as r}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30 transition">
								<td class="px-5 py-3 font-semibold text-gray-900 dark:text-white font-mono">{r.date}</td>
								<td class="px-5 py-3 font-bold">{r.sent.toLocaleString()}</td>
								<td class="px-5 py-3 text-emerald-600 font-semibold">{r.delivered.toLocaleString()}</td>
								<td class="px-5 py-3">
									<span class="inline-flex items-center px-2 py-0.5 rounded text-[11px] font-medium bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300">
										{r.sent > 0 ? ((r.delivered / r.sent) * 100).toFixed(1) : '100'}%
									</span>
								</td>
								<td class="px-5 py-3 font-bold text-gray-900 dark:text-white">৳{r.cost.toFixed(2)}</td>
								<td class="px-5 py-3 font-bold text-emerald-600 dark:text-emerald-400">৳{r.profit.toFixed(2)}</td>
								<td class="px-5 py-3 text-right">
									<button
										type="button"
										class="text-xs text-emerald-600 hover:text-emerald-700 font-semibold"
										onclick={() => alert(`Exporting statement for ${r.date}`)}
									>
										Export CSV
									</button>
								</td>
							</tr>
						{/each}
					{/if}
				</tbody>
			</table>
		</div>
	</div>
</div>

