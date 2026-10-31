<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface RoutePlan {
		id: number;
		name: string;
		code: string;
		description: string;
		gateways: string;
		is_default: boolean;
	}

	let routes = $state<RoutePlan[]>([]);
	let loading = $state(true);

	const exportColumns: ExportColumn<RoutePlan>[] = [
		{ header: 'ID', getValue: (r) => r.id },
		{ header: 'Route Name', getValue: (r) => r.name },
		{ header: 'Code', getValue: (r) => r.code },
		{ header: 'Description', getValue: (r) => r.description },
		{ header: 'Gateway Sequence', getValue: (r) => r.gateways },
		{ header: 'Default Route', getValue: (r) => (r.is_default ? 'Yes' : 'No') }
	];

	onMount(async () => {
		try {
			const res = await supermanApi.get<{ data: RoutePlan[] }>('/api/v1/superman/sps/routing');
			routes = res.data ?? [];
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
			<h1 class="text-xl font-bold text-gray-900 dark:text-white">Reseller Custom Routing Bindings</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
				Dedicated carrier gateway bindings and failover sequences assigned exclusively to Service Provider networks.
			</p>
		</div>
		<div class="flex items-center gap-3">
			<ExportDropdown
				title="Reseller Custom Routing Bindings"
				columns={exportColumns}
				getData={() => routes}
				disabled={routes.length === 0}
			/>
			<a
				href="/superman/routing/dashboard"
				class="inline-flex items-center gap-2 rounded-lg bg-emerald-600 px-3 py-1.5 text-xs font-semibold text-white shadow-xs hover:bg-emerald-700 transition"
			>
				Global Routing Control
			</a>
		</div>
	</div>

	<!-- Routing Table -->
	<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl shadow-xs overflow-hidden">
		<div class="overflow-x-auto">
			<table class="w-full text-left text-xs">
				<thead class="bg-gray-50/75 dark:bg-gray-800/50 text-gray-500 uppercase tracking-wider font-semibold border-b border-gray-200 dark:border-gray-800">
					<tr>
						<th class="px-5 py-3">Route Plan</th>
						<th class="px-5 py-3">Identifier Code</th>
						<th class="px-5 py-3">Bound Gateway Priority Chain</th>
						<th class="px-5 py-3">Type</th>
						<th class="px-5 py-3 text-right">Actions</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800 text-gray-700 dark:text-gray-300">
					{#if loading}
						<tr>
							<td colspan="5" class="px-5 py-8 text-center text-gray-400">Loading reseller routing...</td>
						</tr>
					{:else if routes.length === 0}
						<tr>
							<td colspan="5" class="px-5 py-8 text-center text-gray-400">No custom routes found.</td>
						</tr>
					{:else}
						{#each routes as r}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30 transition">
								<td class="px-5 py-3 font-bold text-gray-900 dark:text-white">{r.name}</td>
								<td class="px-5 py-3 font-mono text-gray-500">{r.code}</td>
								<td class="px-5 py-3 font-medium text-emerald-600 dark:text-emerald-400">{r.gateways}</td>
								<td class="px-5 py-3">
									<span class="inline-flex items-center px-2 py-0.5 rounded text-[11px] font-medium {r.is_default ? 'bg-purple-100 text-purple-800 dark:bg-purple-950 dark:text-purple-300' : 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300'}">
										{r.is_default ? 'Default Global' : 'Custom SP'}
									</span>
								</td>
								<td class="px-5 py-3 text-right">
									<a href="/superman/routing/dashboard" class="text-xs text-emerald-600 hover:text-emerald-700 font-semibold">
										Modify Chain
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

