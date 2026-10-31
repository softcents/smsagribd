<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface ErrorMapping {
		provider: string;
		upstream_code: string;
		standard_code: number;
		meaning: string;
		action: string;
	}

	let mappings = $state<ErrorMapping[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);

	async function loadMappings() {
		loading = true;
		error = null;
		try {
			const res = await supermanApi.get<{ data: ErrorMapping[] }>('/api/v1/superman/providers/error-mapping');
			mappings = res.data ?? [];
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Failed to load error code mappings.';
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		void loadMappings();
	});

	const exportColumns: ExportColumn<ErrorMapping>[] = [
		{ header: 'Provider / Gateway', getValue: (m) => m.provider },
		{ header: 'Vendor Upstream Error', getValue: (m) => m.upstream_code },
		{ header: 'Standard Platform Code', getValue: (m) => m.standard_code },
		{ header: 'Mapped Technical Meaning', getValue: (m) => m.meaning },
		{ header: 'Routing & Failover Action', getValue: (m) => m.action }
	];
</script>

<svelte:head>
	<title>Error Code Mapping — Superman Administrator</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">Gateway Error Code Normalization</h1>
			<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
				Automated translation rules for telco vendor rejections, SMPP error codes, and failover triggers.
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
				title="Gateway Error Code Mappings"
				columns={exportColumns}
				getData={() => mappings}
				disabled={mappings.length === 0}
			/>
		</div>
	</div>

	<!-- Information Banner -->
	<div class="rounded-xl border border-gray-200 bg-white p-5 shadow-xs dark:border-gray-800 dark:bg-gray-900">
		<h3 class="text-sm font-bold text-gray-900 dark:text-white">Intelligent Error Resolution Engine</h3>
		<p class="mt-1 text-xs text-gray-600 dark:text-gray-400">
			When a transmission fails upstream, our dispatch layer parses vendor error packets against these normalized error signatures.
			Transient network errors automatically route to secondary fallback carriers, while validation errors generate instant customer refunds.
		</p>
	</div>

	<div class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-xs dark:border-gray-800 dark:bg-gray-900">
		{#if loading}
			<div class="flex items-center justify-center py-16 text-xs text-gray-500">
				<svg class="mr-2 h-4 w-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path></svg>
				Loading error code mappings...
			</div>
		{:else if error}
			<div class="p-6 text-center text-xs text-rose-600">{error}</div>
		{:else}
			<table class="w-full text-left text-xs">
				<thead>
					<tr class="border-b border-gray-100 bg-gray-50/50 text-[10px] font-semibold uppercase tracking-wider text-gray-500 dark:border-gray-800 dark:bg-gray-800/50 dark:text-gray-400">
						<th class="px-5 py-3.5">Provider / Gateway</th>
						<th class="px-5 py-3.5">Vendor Upstream Error</th>
						<th class="px-5 py-3.5">Standard Platform Code</th>
						<th class="px-5 py-3.5">Error Meaning</th>
						<th class="px-5 py-3.5">Automated Platform Action</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
					{#each mappings as m}
						<tr class="transition-colors hover:bg-gray-50/50 dark:hover:bg-gray-800/30">
							<td class="px-5 py-4 font-semibold text-gray-950 dark:text-white">
								{m.provider}
							</td>
							<td class="px-5 py-4 font-mono text-rose-600 font-semibold">
								{m.upstream_code}
							</td>
							<td class="px-5 py-4 font-mono font-bold text-gray-700 dark:text-gray-300">
								#{m.standard_code}
							</td>
							<td class="px-5 py-4 text-gray-600 dark:text-gray-400">
								{m.meaning}
							</td>
							<td class="px-5 py-4">
								<span class="inline-flex rounded-md bg-emerald-50 px-2 py-0.5 text-[11px] font-medium text-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-400">
									{m.action}
								</span>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		{/if}
	</div>
</div>

