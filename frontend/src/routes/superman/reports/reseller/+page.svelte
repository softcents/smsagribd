<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface ResellerItem {
		id: number;
		name: string;
		email: string;
		phone: string;
		business_name?: string;
		balance: number;
		customer_count: number;
		status: string;
	}

	let resellers = $state<ResellerItem[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let search = $state('');

	async function loadResellers() {
		loading = true;
		error = null;
		try {
			const res = await supermanApi.get<{ data: ResellerItem[] }>('/api/v1/superman/sps');
			resellers = res.data ?? [];
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Failed to load reseller reports.';
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		void loadResellers();
	});

	let filtered = $derived(
		resellers.filter((r) => search === '' || r.name.toLowerCase().includes(search.toLowerCase()) || r.email.toLowerCase().includes(search.toLowerCase()))
	);

	const exportColumns: ExportColumn<ResellerItem>[] = [
		{ header: 'Reseller Name', getValue: (r) => r.name },
		{ header: 'Email', getValue: (r) => r.email },
		{ header: 'Phone', getValue: (r) => r.phone },
		{ header: 'Business Entity', getValue: (r) => r.business_name || 'Individual' },
		{ header: 'Sub-Clients', getValue: (r) => `${r.customer_count || 0} Clients` },
		{ header: 'Account Status', getValue: (r) => r.status.toUpperCase() },
		{ header: 'Available Balance', getValue: (r) => `BDT ${Number(r.balance || 0).toFixed(2)}` }
	];
</script>

<svelte:head>
	<title>Reseller Reports — Superman Administrator</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">Reseller Performance & Billing Reports</h1>
			<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
				Overview of all Service Provider reseller accounts, credit balances, and sub-customer counts.
			</p>
		</div>
		<div class="flex items-center gap-2">
			<ExportDropdown
				title="Reseller Performance Reports"
				columns={exportColumns}
				getData={() => filtered}
				disabled={filtered.length === 0}
			/>
			<button onclick={loadResellers} class="rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs font-medium text-gray-700 shadow-xs hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300">
				Refresh
			</button>
		</div>
	</div>

	<div class="flex items-center gap-3 rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
		<input
			type="search"
			bind:value={search}
			placeholder="Search by reseller name or email..."
			class="w-full max-w-sm rounded-lg border border-gray-200 bg-gray-50/70 px-3 py-1.5 text-xs outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
		/>
	</div>

	<div class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-xs dark:border-gray-800 dark:bg-gray-900">
		{#if loading}
			<div class="py-16 text-center text-xs text-gray-500">Loading reseller reports...</div>
		{:else if error}
			<div class="p-6 text-center text-xs text-rose-600">{error}</div>
		{:else if filtered.length === 0}
			<div class="py-16 text-center text-xs text-gray-500">No reseller records found.</div>
		{:else}
			<div class="overflow-x-auto">
				<table class="w-full text-left text-xs">
					<thead>
						<tr class="border-b border-gray-100 bg-gray-50/50 text-[10px] font-semibold uppercase tracking-wider text-gray-500 dark:border-gray-800 dark:bg-gray-800/50 dark:text-gray-400">
							<th class="px-4 py-3">Reseller Name</th>
							<th class="px-4 py-3">Email & Phone</th>
							<th class="px-4 py-3">Business Entity</th>
							<th class="px-4 py-3">Sub-Clients</th>
							<th class="px-4 py-3">Account Status</th>
							<th class="px-4 py-3 text-right">Available Balance</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
						{#each filtered as r}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30">
								<td class="px-4 py-3 font-semibold text-gray-950 dark:text-white">{r.name}</td>
								<td class="px-4 py-3 text-gray-600 dark:text-gray-400">
									<div>{r.email}</div>
									<div class="text-[11px] text-gray-400">{r.phone}</div>
								</td>
								<td class="px-4 py-3 font-medium text-gray-700 dark:text-gray-300">{r.business_name || 'Individual'}</td>
								<td class="px-4 py-3 font-bold text-blue-600">{r.customer_count || 0} Clients</td>
								<td class="px-4 py-3 font-semibold text-emerald-600 uppercase text-[11px]">{r.status}</td>
								<td class="px-4 py-3 text-right font-bold text-gray-900 dark:text-white">৳{Number(r.balance || 0).toFixed(2)}</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</div>
</div>

