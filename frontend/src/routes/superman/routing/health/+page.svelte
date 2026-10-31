<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface RouteHealthItem {
		id: number;
		route_plan_name: string;
		gateway_name: string;
		priority: number;
		weight: number;
		is_active: boolean;
		created_at: string;
	}

	let items = $state<RouteHealthItem[]>([]);
	let loading = $state(true);

	const exportColumns: ExportColumn<RouteHealthItem>[] = [
		{ header: 'ID', getValue: (i) => i.id },
		{ header: 'Assigned Route Plan', getValue: (i) => i.route_plan_name },
		{ header: 'Gateway', getValue: (i) => i.gateway_name },
		{ header: 'Failover Order', getValue: (i) => `Priority ${i.priority}` },
		{ header: 'Traffic Weight', getValue: (i) => `${i.weight}%` },
		{ header: 'Link Active', getValue: (i) => i.is_active ? 'ACTIVE' : 'DISABLED' },
		{ header: 'Configured Date', getValue: (i) => new Date(i.created_at).toLocaleDateString() }
	];

	async function loadHealth() {
		loading = true;
		try {
			const res = await supermanApi.get<{ data: RouteHealthItem[] }>('/api/v1/superman/routing/health');
			items = res.data ?? [];
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		loadHealth();
	});
</script>

<div class="space-y-6">
	<!-- Page Header -->
	<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
		<div>
			<h1 class="text-xl font-bold text-gray-900 dark:text-white">Gateway Route Link Health</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
				Real-time availability and priority failover bindings between outbound route plans and interconnect telecom gateways.
			</p>
		</div>
		<div class="flex items-center gap-3">
			<button
				onclick={loadHealth}
				class="inline-flex items-center gap-1.5 px-3 py-2 text-xs font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 dark:bg-gray-800 dark:text-gray-300 dark:border-gray-700 dark:hover:bg-gray-700/60 shadow-xs transition"
			>
				<svg class="w-3.5 h-3.5 {loading ? 'animate-spin' : ''}" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
				</svg>
				Refresh
			</button>
			<ExportDropdown
				title="Gateway Route Link Health"
				columns={exportColumns}
				getData={() => items}
				disabled={items.length === 0}
			/>
		</div>
	</div>

	<!-- Table -->
	<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl shadow-xs overflow-hidden">
		<div class="overflow-x-auto">
			<table class="w-full text-left text-xs">
				<thead class="bg-gray-50/75 dark:bg-gray-800/50 text-gray-500 uppercase tracking-wider font-semibold border-b border-gray-200 dark:border-gray-800">
					<tr>
						<th class="px-5 py-3">Assigned Route Plan</th>
						<th class="px-5 py-3">Gateway</th>
						<th class="px-5 py-3 text-center">Failover Order</th>
						<th class="px-5 py-3 text-center">Traffic Weight</th>
						<th class="px-5 py-3 text-center">Link Active</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800 text-gray-700 dark:text-gray-300">
					{#if loading}
						<tr>
							<td colspan="5" class="px-5 py-10 text-center text-gray-400">Loading route link health diagnostics...</td>
						</tr>
					{:else if items.length === 0}
						<tr>
							<td colspan="5" class="px-5 py-10 text-center text-gray-400">No route plan gateway links configured.</td>
						</tr>
					{:else}
						{#each items as it}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30 transition">
								<td class="px-5 py-3.5 font-bold text-gray-900 dark:text-white">
									{it.route_plan_name}
								</td>
								<td class="px-5 py-3.5">
									<span class="inline-flex items-center px-2 py-0.5 rounded-full text-[11px] font-semibold bg-emerald-50 text-emerald-800 border border-emerald-200 dark:bg-emerald-950/40 dark:text-emerald-300 dark:border-emerald-800">
										{it.gateway_name}
									</span>
								</td>
								<td class="px-5 py-3.5 text-center font-mono font-bold text-gray-700 dark:text-gray-300">
									Priority {it.priority}
								</td>
								<td class="px-5 py-3.5 text-center font-mono text-gray-700 dark:text-gray-300">
									{it.weight}%
								</td>
								<td class="px-5 py-3.5 text-center">
									{#if it.is_active}
										<span class="inline-flex items-center text-emerald-600 dark:text-emerald-400">
											<svg class="w-5 h-5 mx-auto" fill="none" viewBox="0 0 24 24" stroke="currentColor">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7" />
											</svg>
										</span>
									{:else}
										<span class="inline-flex items-center text-gray-300 dark:text-gray-600">
											<svg class="w-5 h-5 mx-auto" fill="none" viewBox="0 0 24 24" stroke="currentColor">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
											</svg>
										</span>
									{/if}
								</td>
							</tr>
						{/each}
					{/if}
				</tbody>
			</table>
		</div>
	</div>
</div>
