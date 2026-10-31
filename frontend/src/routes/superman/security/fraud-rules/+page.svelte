<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface FraudRule {
		id: number;
		name: string;
		type: string;
		threshold: string;
		action: string;
		is_active: boolean;
	}

	let rules = $state<FraudRule[]>([]);
	let loading = $state(true);

	onMount(async () => {
		try {
			const res = await supermanApi.get<{ data: FraudRule[] }>('/api/v1/superman/security/fraud-rules');
			rules = res.data ?? [];
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	});

	const exportColumns: ExportColumn<FraudRule>[] = [
		{ header: 'Rule Name', getValue: (r) => r.name },
		{ header: 'Inspection Type', getValue: (r) => r.type },
		{ header: 'Threshold / Pattern', getValue: (r) => r.threshold },
		{ header: 'Automated Action', getValue: (r) => r.action },
		{ header: 'State', getValue: (r) => (r.is_active ? 'Active' : 'Disabled') }
	];
</script>

<div class="space-y-6">
	<!-- Page Header -->
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-xl font-bold text-gray-900 dark:text-white">Fraud Detection & Velocity Rules</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
				Real-time message payload inspection, velocity throttling per MSISDN, and spam keyword triggers.
			</p>
		</div>
		<div class="flex items-center gap-2">
			<ExportDropdown
				title="Fraud Detection Rules"
				columns={exportColumns}
				getData={() => rules}
				disabled={rules.length === 0}
			/>
			<button
				type="button"
				class="inline-flex items-center gap-2 rounded-lg bg-emerald-600 px-3 py-1.5 text-xs font-semibold text-white shadow-xs hover:bg-emerald-700 transition"
				onclick={() => alert('New fraud rule wizard opened.')}
			>
				+ Create Fraud Rule
			</button>
		</div>
	</div>

	<!-- Rules Table -->
	<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl shadow-xs overflow-hidden">
		<div class="overflow-x-auto">
			<table class="w-full text-left text-xs">
				<thead class="bg-gray-50/75 dark:bg-gray-800/50 text-gray-500 uppercase tracking-wider font-semibold border-b border-gray-200 dark:border-gray-800">
					<tr>
						<th class="px-5 py-3">Rule Name</th>
						<th class="px-5 py-3">Inspection Type</th>
						<th class="px-5 py-3">Threshold / Pattern</th>
						<th class="px-5 py-3">Automated Action</th>
						<th class="px-5 py-3">State</th>
						<th class="px-5 py-3 text-right">Actions</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800 text-gray-700 dark:text-gray-300">
					{#if loading}
						<tr>
							<td colspan="6" class="px-5 py-8 text-center text-gray-400">Loading fraud rules...</td>
						</tr>
					{:else if rules.length === 0}
						<tr>
							<td colspan="6" class="px-5 py-8 text-center text-gray-400">No fraud rules configured.</td>
						</tr>
					{:else}
						{#each rules as r}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30 transition">
								<td class="px-5 py-3 font-bold text-gray-900 dark:text-white">{r.name}</td>
								<td class="px-5 py-3 font-mono uppercase text-gray-500">{r.type}</td>
								<td class="px-5 py-3 font-mono text-gray-800 dark:text-gray-200">{r.threshold}</td>
								<td class="px-5 py-3 font-semibold text-red-600 dark:text-red-400">{r.action}</td>
								<td class="px-5 py-3">
									<span class="inline-flex items-center px-2 py-0.5 rounded text-[11px] font-medium {r.is_active ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300' : 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300'}">
										{r.is_active ? 'Active' : 'Disabled'}
									</span>
								</td>
								<td class="px-5 py-3 text-right">
									<button
										type="button"
										class="text-xs text-emerald-600 hover:text-emerald-700 font-semibold"
										onclick={() => alert(`Editing rule: ${r.name}`)}
									>
										Edit Rule
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

