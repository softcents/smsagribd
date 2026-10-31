<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface RatePlan {
		id: number;
		name: string;
		description: string;
		is_active: boolean;
		created_at: string;
	}

	let plans = $state<RatePlan[]>([]);
	let loading = $state(true);

	const exportColumns: ExportColumn<RatePlan>[] = [
		{ header: 'ID', getValue: (p) => p.id },
		{ header: 'Plan Name', getValue: (p) => p.name },
		{ header: 'Description', getValue: (p) => p.description },
		{ header: 'Status', getValue: (p) => (p.is_active ? 'Active' : 'Inactive') },
		{ header: 'Created At', getValue: (p) => p.created_at }
	];

	onMount(async () => {
		try {
			const res = await supermanApi.get<{ data: RatePlan[] }>('/api/v1/superman/sps/rates');
			plans = res.data ?? [];
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
			<h1 class="text-xl font-bold text-gray-900 dark:text-white">Reseller Wholesale & Retail Tariff Plans</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
				Wholesale tariff rate cards and margin matrices configured for Service Provider resellers.
			</p>
		</div>
		<div class="flex items-center gap-3">
			<ExportDropdown
				title="Reseller Tariff Plans"
				columns={exportColumns}
				getData={() => plans}
				disabled={plans.length === 0}
			/>
			<a
				href="/superman/billing/tariff-plans"
				class="inline-flex items-center gap-2 rounded-lg bg-emerald-600 px-3 py-1.5 text-xs font-semibold text-white shadow-xs hover:bg-emerald-700 transition"
			>
				+ Create Tariff Plan
			</a>
		</div>
	</div>

	<!-- Tariff Plans Table -->
	<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl shadow-xs overflow-hidden">
		<div class="overflow-x-auto">
			<table class="w-full text-left text-xs">
				<thead class="bg-gray-50/75 dark:bg-gray-800/50 text-gray-500 uppercase tracking-wider font-semibold border-b border-gray-200 dark:border-gray-800">
					<tr>
						<th class="px-5 py-3">Plan Name</th>
						<th class="px-5 py-3">Description</th>
						<th class="px-5 py-3">Status</th>
						<th class="px-5 py-3">Created</th>
						<th class="px-5 py-3 text-right">Actions</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800 text-gray-700 dark:text-gray-300">
					{#if loading}
						<tr>
							<td colspan="5" class="px-5 py-8 text-center text-gray-400">Loading tariff plans...</td>
						</tr>
					{:else if plans.length === 0}
						<tr>
							<td colspan="5" class="px-5 py-8 text-center text-gray-400">No reseller plans defined.</td>
						</tr>
					{:else}
						{#each plans as p}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30 transition">
								<td class="px-5 py-3 font-bold text-gray-900 dark:text-white">{p.name}</td>
								<td class="px-5 py-3 text-gray-500">{p.description || 'Standard wholesale rates'}</td>
								<td class="px-5 py-3">
									<span class="inline-flex items-center px-2 py-0.5 rounded text-[11px] font-medium {p.is_active ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300' : 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300'}">
										{p.is_active ? 'Active' : 'Inactive'}
									</span>
								</td>
								<td class="px-5 py-3 text-gray-400">{new Date(p.created_at).toLocaleDateString()}</td>
								<td class="px-5 py-3 text-right">
									<a href="/superman/billing/tariff-plans" class="text-xs text-emerald-600 hover:text-emerald-700 font-semibold">
										Configure Rates
									</a>
								</td>
							</tr>
						{/each}
					{/if}
				</tbody>
			</table>
		</div>
	</div>
</div>

