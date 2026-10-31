<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface OperatorItem {
		id: number;
		name: string;
		country_name: string;
		codes: string;
		masking_price: number;
		non_masking_price: number;
		is_active: boolean;
	}

	let operators = $state<OperatorItem[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let search = $state('');

	async function loadOperators() {
		loading = true;
		error = null;
		try {
			const res = await supermanApi.get<{ data: OperatorItem[] }>('/api/v1/superman/operators');
			operators = res.data ?? [];
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Failed to load operators.';
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		void loadOperators();
	});

	let filtered = $derived(
		operators.filter(o => search === '' || o.name.toLowerCase().includes(search.toLowerCase()) || (o.codes && o.codes.includes(search)))
	);

	const exportColumns: ExportColumn<OperatorItem>[] = [
		{ header: 'Operator Name', getValue: (o) => o.name },
		{ header: 'Country', getValue: (o) => o.country_name },
		{ header: 'Prefix Codes', getValue: (o) => o.codes },
		{ header: 'Masking Base Rate (BDT)', getValue: (o) => `BDT ${Number(o.masking_price).toFixed(4)}` },
		{ header: 'Non-Masking Base Rate (BDT)', getValue: (o) => `BDT ${Number(o.non_masking_price).toFixed(4)}` },
		{ header: 'Status', getValue: (o) => (o.is_active ? 'Active' : 'Inactive') }
	];
</script>

<svelte:head>
	<title>Operators — Superman Administrator</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">Mobile Network Operators (MNOs)</h1>
			<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
				Configured mobile carriers, dialing prefix routes, and baseline carrier tariffs.
			</p>
		</div>
		<div class="flex items-center gap-2">
			<button onclick={loadOperators} class="rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs font-medium text-gray-700 shadow-xs hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300">
				Refresh Operators
			</button>
			<ExportDropdown
				title="Mobile Network Operators"
				columns={exportColumns}
				getData={() => filtered}
				disabled={filtered.length === 0}
			/>
		</div>
	</div>

	<div class="flex items-center gap-3 rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
		<input
			type="search"
			bind:value={search}
			placeholder="Search operator name or prefix code..."
			class="w-full max-w-sm rounded-lg border border-gray-200 bg-gray-50/70 px-3 py-1.5 text-xs outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
		/>
	</div>

	<div class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-xs dark:border-gray-800 dark:bg-gray-900">
		{#if loading}
			<div class="py-16 text-center text-xs text-gray-500">Loading mobile operators...</div>
		{:else if error}
			<div class="p-6 text-center text-xs text-rose-600">{error}</div>
		{:else if filtered.length === 0}
			<div class="py-16 text-center text-xs text-gray-500">No mobile operators found.</div>
		{:else}
			<table class="w-full text-left text-xs">
				<thead>
					<tr class="border-b border-gray-100 bg-gray-50/50 text-[10px] font-semibold uppercase tracking-wider text-gray-500 dark:border-gray-800 dark:bg-gray-800/50 dark:text-gray-400">
						<th class="px-5 py-3.5">Operator Name</th>
						<th class="px-5 py-3.5">Country</th>
						<th class="px-5 py-3.5">Mobile Prefixes</th>
						<th class="px-5 py-3.5 text-right">Base Masking</th>
						<th class="px-5 py-3.5 text-right">Base Non-Masking</th>
						<th class="px-5 py-3.5 text-right">Status</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
					{#each filtered as o}
						<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30">
							<td class="px-5 py-4 font-semibold text-gray-950 dark:text-white">{o.name}</td>
							<td class="px-5 py-4 text-gray-600 dark:text-gray-400">{o.country_name || 'Bangladesh'}</td>
							<td class="px-5 py-4 font-mono font-medium text-blue-600">{o.codes || '—'}</td>
							<td class="px-5 py-4 text-right font-medium">৳{Number(o.masking_price || 0).toFixed(4)}</td>
							<td class="px-5 py-4 text-right font-medium">৳{Number(o.non_masking_price || 0).toFixed(4)}</td>
							<td class="px-5 py-4 text-right">
								<span class="inline-flex rounded-full px-2 py-0.5 text-[10px] font-bold uppercase {o.is_active ? 'bg-emerald-50 text-emerald-700' : 'bg-gray-100 text-gray-500'}">
									{o.is_active ? 'Active' : 'Disabled'}
								</span>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		{/if}
	</div>
</div>

