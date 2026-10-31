<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface RoutePlan {
		id: number;
		name: string;
		code: string;
		traffic_type: string;
		failover_strategy: string;
		is_default: boolean;
		is_enabled: boolean;
		description: string;
		priority_chain: string;
		gateways_summary: string;
		created_at: string;
	}

	let plans = $state<RoutePlan[]>([]);
	let loading = $state(true);

	const exportColumns: ExportColumn<RoutePlan>[] = [
		{ header: 'ID', getValue: (p) => p.id },
		{ header: 'Plan Name', getValue: (p) => p.name },
		{ header: 'Plan Code', getValue: (p) => p.code },
		{ header: 'Traffic Category', getValue: (p) => p.traffic_type.toUpperCase() },
		{ header: 'Failover Chain (1 -> 2 -> 3)', getValue: (p) => p.priority_chain },
		{ header: 'Default Plan', getValue: (p) => p.is_default ? 'YES' : 'NO' },
		{ header: 'Operating Status', getValue: (p) => p.is_enabled ? 'ACTIVE' : 'INACTIVE' },
		{ header: 'Description', getValue: (p) => p.description || '—' }
	];

	async function loadPlans() {
		loading = true;
		try {
			const res = await supermanApi.get<{ data: RoutePlan[] }>('/api/v1/superman/routing/rules');
			plans = res.data ?? [];
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		loadPlans();
	});

	function getTrafficBadge(type: string) {
		const t = type.toLowerCase();
		if (t === 'otp') {
			return 'bg-rose-50 text-rose-700 border-rose-200 dark:bg-rose-950/40 dark:text-rose-400 dark:border-rose-800';
		}
		if (t === 'transactional') {
			return 'bg-blue-50 text-blue-700 border-blue-200 dark:bg-blue-950/40 dark:text-blue-400 dark:border-blue-800';
		}
		if (t === 'promotional') {
			return 'bg-amber-50 text-amber-700 border-amber-200 dark:bg-amber-950/40 dark:text-amber-400 dark:border-amber-800';
		}
		return 'bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-950/40 dark:text-emerald-400 dark:border-emerald-800';
	}
</script>

<div class="space-y-6">
	<!-- Page Header -->
	<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
		<div>
			<h1 class="text-xl font-bold text-gray-900 dark:text-white">SMS Routing Plans & Rule Matrix</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
				Configure intelligent multi-carrier failover chains, primary/backup gateway priorities, and traffic steering profiles.
			</p>
		</div>
		<div class="flex items-center gap-3">
			<button
				onclick={loadPlans}
				class="inline-flex items-center gap-1.5 px-3 py-2 text-xs font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 dark:bg-gray-800 dark:text-gray-300 dark:border-gray-700 dark:hover:bg-gray-700/60 shadow-xs transition"
			>
				<svg class="w-3.5 h-3.5 {loading ? 'animate-spin' : ''}" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
				</svg>
				Refresh
			</button>
			<ExportDropdown
				title="SMS Routing Plans Matrix"
				columns={exportColumns}
				getData={() => plans}
				disabled={plans.length === 0}
			/>
		</div>
	</div>

	<!-- Rules Table -->
	<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl shadow-xs overflow-hidden">
		<div class="overflow-x-auto">
			<table class="w-full text-left text-xs">
				<thead class="bg-gray-50/75 dark:bg-gray-800/50 text-gray-500 uppercase tracking-wider font-semibold border-b border-gray-200 dark:border-gray-800">
					<tr>
						<th class="px-5 py-3">Plan Name</th>
						<th class="px-5 py-3">Code</th>
						<th class="px-5 py-3">Traffic</th>
						<th class="px-5 py-3">Priority Chain (1 → 2 → 3)</th>
						<th class="px-5 py-3 text-center">Default</th>
						<th class="px-5 py-3 text-center">Status</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800 text-gray-700 dark:text-gray-300">
					{#if loading}
						<tr>
							<td colspan="6" class="px-5 py-10 text-center text-gray-400">Loading routing plans matrix...</td>
						</tr>
					{:else if plans.length === 0}
						<tr>
							<td colspan="6" class="px-5 py-10 text-center text-gray-400">No routing profiles configured.</td>
						</tr>
					{:else}
						{#each plans as p}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30 transition">
								<td class="px-5 py-4">
									<div class="font-bold text-gray-900 dark:text-white">{p.name}</div>
									<div class="text-[11px] text-gray-400 mt-0.5">{p.description}</div>
								</td>
								<td class="px-5 py-4">
									<span class="inline-flex items-center px-2 py-0.5 rounded text-[11px] font-mono font-medium bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300">
										{p.code}
									</span>
								</td>
								<td class="px-5 py-4">
									<span class="inline-flex items-center px-2 py-0.5 rounded-full text-[11px] font-semibold border {getTrafficBadge(p.traffic_type)}">
										{p.traffic_type.toUpperCase()}
									</span>
								</td>
								<td class="px-5 py-4">
									<span class="inline-flex items-center px-2.5 py-1 rounded-md text-[11px] font-semibold bg-emerald-50 text-emerald-800 border border-emerald-200 dark:bg-emerald-950/40 dark:text-emerald-300 dark:border-emerald-800">
										{p.priority_chain}
									</span>
								</td>
								<td class="px-5 py-4 text-center">
									{#if p.is_default}
										<span class="inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-bold bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300">
											YES
										</span>
									{:else}
										<span class="inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-bold bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400">
											NO
										</span>
									{/if}
								</td>
								<td class="px-5 py-4 text-center">
									<span class="inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-bold {p.is_enabled ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300' : 'bg-rose-100 text-rose-800 dark:bg-rose-950 dark:text-rose-300'}">
										{p.is_enabled ? 'ACTIVE' : 'INACTIVE'}
									</span>
								</td>
							</tr>
						{/each}
					{/if}
				</tbody>
			</table>
		</div>
	</div>
</div>
