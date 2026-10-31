<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface DlrMapping {
		gateway_code: string;
		system_status: string;
		description: string;
		badge_color: string;
	}

	let mappings = $state<DlrMapping[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);

	async function loadMappings() {
		loading = true;
		error = null;
		try {
			const res = await supermanApi.get<{ data: DlrMapping[] }>('/api/v1/superman/providers/dlr-mapping');
			mappings = res.data ?? [];
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Failed to load DLR status mappings.';
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		void loadMappings();
	});

	const exportColumns: ExportColumn<DlrMapping>[] = [
		{ header: 'Upstream Code / Pattern', getValue: (m) => m.gateway_code },
		{ header: 'Normalized Status', getValue: (m) => m.system_status.toUpperCase() },
		{ header: 'Description', getValue: (m) => m.description }
	];
</script>

<svelte:head>
	<title>DLR Status Mapping — Superman Administrator</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">Upstream Gateway DLR Status Mappings</h1>
			<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
				Mapping matrix between diverse telco vendor callback codes and standard platform delivery states.
			</p>
		</div>
		<div class="flex items-center gap-2">
			<button
				onclick={loadMappings}
				class="inline-flex items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs font-medium text-gray-700 shadow-xs hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300 dark:hover:bg-gray-800"
			>
				Refresh
			</button>
			<ExportDropdown
				title="DLR Status Mappings"
				columns={exportColumns}
				getData={() => mappings}
				disabled={mappings.length === 0}
			/>
		</div>
	</div>

	<!-- Information Banner -->
	<div class="rounded-xl border border-gray-200 bg-white p-5 shadow-xs dark:border-gray-800 dark:bg-gray-900">
		<h3 class="text-sm font-bold text-gray-900 dark:text-white">Delivery Receipt Normalization</h3>
		<p class="mt-1 text-xs text-gray-600 dark:text-gray-400">
			Upstream telcos and SMSCs communicate delivery outcomes via varying formats (SMPP TLV optional parameters, HTTP callback parameters, or short text tokens).
			Our normalization layer harmonizes all incoming DLR signals into unified customer webhooks and billing ledger records.
		</p>
	</div>

	<div class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-xs dark:border-gray-800 dark:bg-gray-900">
		{#if loading}
			<div class="flex items-center justify-center py-16 text-xs text-gray-500">
				<svg class="mr-2 h-4 w-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path></svg>
				Loading mappings...
			</div>
		{:else if error}
			<div class="p-6 text-center text-xs text-rose-600">{error}</div>
		{:else}
			<table class="w-full text-left text-xs">
				<thead>
					<tr class="border-b border-gray-100 bg-gray-50/50 text-[10px] font-semibold uppercase tracking-wider text-gray-500 dark:border-gray-800 dark:bg-gray-800/50 dark:text-gray-400">
						<th class="px-5 py-3.5">Upstream Provider Code / Pattern</th>
						<th class="px-5 py-3.5">Normalized Platform Status</th>
						<th class="px-5 py-3.5">Description & Handset Interpretation</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
					{#each mappings as m}
						<tr class="transition-colors hover:bg-gray-50/50 dark:hover:bg-gray-800/30">
							<td class="px-5 py-4 font-mono font-semibold text-gray-900 dark:text-white">
								{m.gateway_code}
							</td>
							<td class="px-5 py-4">
								<span class="inline-flex rounded-full px-2.5 py-1 text-[11px] font-semibold {m.badge_color === 'emerald' ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-400' : m.badge_color === 'blue' ? 'bg-blue-50 text-blue-700 dark:bg-blue-950/40 dark:text-blue-400' : m.badge_color === 'amber' ? 'bg-amber-50 text-amber-700 dark:bg-amber-950/40 dark:text-amber-400' : 'bg-rose-50 text-rose-700 dark:bg-rose-950/40 dark:text-rose-400'}">
									{m.system_status}
								</span>
							</td>
							<td class="px-5 py-4 text-gray-600 dark:text-gray-400">
								{m.description}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		{/if}
	</div>
</div>

