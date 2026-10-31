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
		created_at: string;
	}

	let plans = $state<RoutePlan[]>([]);
	let loading = $state(true);

	const exportColumns: ExportColumn<RoutePlan>[] = [
		{ header: 'ID', getValue: (p) => p.id },
		{ header: 'Route Name', getValue: (p) => p.name },
		{ header: 'Identifier Code', getValue: (p) => p.code },
		{ header: 'Description', getValue: (p) => p.description },
		{ header: 'Carrier Sequence', getValue: (p) => p.gateways },
		{ header: 'Default Route', getValue: (p) => (p.is_default ? 'Yes' : 'No') },
		{ header: 'Created At', getValue: (p) => p.created_at }
	];

	onMount(async () => {
		try {
			const res = await supermanApi.get<{ data: RoutePlan[] }>('/api/v1/superman/routing/routes');
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
			<h1 class="text-xl font-bold text-gray-900 dark:text-white">Route Plan Definitions</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
				Create and maintain priority ordered gateway routes that can be bound to customer profiles and tariff tiers.
			</p>
		</div>
		<div class="flex items-center gap-3">
			<ExportDropdown
				title="Route Plan Definitions"
				columns={exportColumns}
				getData={() => plans}
				disabled={plans.length === 0}
			/>
			<button
				type="button"
				class="inline-flex items-center gap-2 rounded-lg bg-emerald-600 px-3 py-1.5 text-xs font-semibold text-white shadow-xs hover:bg-emerald-700 transition"
				onclick={() => alert('New Route Plan Wizard opened.')}
			>
				+ Create Route Plan
			</button>
		</div>
	</div>

	<!-- Plans Table -->
	<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl shadow-xs overflow-hidden">
		<div class="overflow-x-auto">
			<table class="w-full text-left text-xs">
				<thead class="bg-gray-50/75 dark:bg-gray-800/50 text-gray-500 uppercase tracking-wider font-semibold border-b border-gray-200 dark:border-gray-800">
					<tr>
						<th class="px-5 py-3">Route Name</th>
						<th class="px-5 py-3">Identifier Code</th>
						<th class="px-5 py-3">Carrier Sequence</th>
						<th class="px-5 py-3">Default Route</th>
						<th class="px-5 py-3 text-right">Actions</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800 text-gray-700 dark:text-gray-300">
					{#if loading}
						<tr>
							<td colspan="5" class="px-5 py-8 text-center text-gray-400">Loading route plans...</td>
						</tr>
					{:else if plans.length === 0}
						<tr>
							<td colspan="5" class="px-5 py-8 text-center text-gray-400">No route plans found.</td>
						</tr>
					{:else}
						{#each plans as p}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30 transition">
								<td class="px-5 py-3 font-bold text-gray-900 dark:text-white">{p.name}</td>
								<td class="px-5 py-3 font-mono text-gray-500">{p.code}</td>
								<td class="px-5 py-3 font-medium text-emerald-600 dark:text-emerald-400">{p.gateways}</td>
								<td class="px-5 py-3">
									<span class="inline-flex items-center px-2 py-0.5 rounded text-[11px] font-medium {p.is_default ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300' : 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300'}">
										{p.is_default ? 'Yes' : 'No'}
									</span>
								</td>
								<td class="px-5 py-3 text-right">
									<button
										type="button"
										class="text-xs text-emerald-600 hover:text-emerald-700 font-semibold"
										onclick={() => alert(`Editing route plan: ${p.name}`)}
									>
										Edit
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

