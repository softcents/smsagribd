<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface CustomerItem {
		id: number;
		name: string;
		email: string;
		phone: string;
		balance: number;
		kyc_status: string;
		status: string;
		created_at: string;
	}

	let customers = $state<CustomerItem[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let search = $state('');

	async function loadCustomers() {
		loading = true;
		error = null;
		try {
			const res = await supermanApi.get<{ data: CustomerItem[] }>('/api/v1/superman/customers?per_page=100');
			customers = res.data ?? [];
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Failed to load customer reports.';
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		void loadCustomers();
	});

	let filtered = $derived(
		customers.filter((c) => search === '' || c.name.toLowerCase().includes(search.toLowerCase()) || c.email.toLowerCase().includes(search.toLowerCase()) || c.phone.includes(search))
	);

	const exportColumns: ExportColumn<CustomerItem>[] = [
		{ header: 'Customer', getValue: (c) => c.name },
		{ header: 'Contact Email', getValue: (c) => c.email },
		{ header: 'Phone', getValue: (c) => c.phone },
		{ header: 'KYC Status', getValue: (c) => (c.kyc_status || 'Unverified').toUpperCase() },
		{ header: 'Account State', getValue: (c) => c.status.toUpperCase() },
		{ header: 'Current Balance', getValue: (c) => `BDT ${c.balance.toFixed(2)}` },
		{ header: 'Joined', getValue: (c) => (c.created_at ? new Date(c.created_at).toLocaleDateString() : '—') }
	];
</script>

<svelte:head>
	<title>Customer Reports — Superman Administrator</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">Customer Account Reports</h1>
			<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
				Client account balances, registration dates, KYC verification levels, and active status.
			</p>
		</div>
		<div class="flex items-center gap-2">
			<ExportDropdown
				title="Customer Account Reports"
				columns={exportColumns}
				getData={() => filtered}
				disabled={filtered.length === 0}
			/>
			<button onclick={loadCustomers} class="rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs font-medium text-gray-700 shadow-xs hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300">
				Refresh
			</button>
		</div>
	</div>

	<div class="flex items-center gap-3 rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
		<input
			type="search"
			bind:value={search}
			placeholder="Search by customer name, email, or mobile..."
			class="w-full max-w-sm rounded-lg border border-gray-200 bg-gray-50/70 px-3 py-1.5 text-xs outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
		/>
	</div>

	<div class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-xs dark:border-gray-800 dark:bg-gray-900">
		{#if loading}
			<div class="py-16 text-center text-xs text-gray-500">Loading customer reports...</div>
		{:else if error}
			<div class="p-6 text-center text-xs text-rose-600">{error}</div>
		{:else if filtered.length === 0}
			<div class="py-16 text-center text-xs text-gray-500">No customer records found.</div>
		{:else}
			<div class="overflow-x-auto">
				<table class="w-full text-left text-xs">
					<thead>
						<tr class="border-b border-gray-100 bg-gray-50/50 text-[10px] font-semibold uppercase tracking-wider text-gray-500 dark:border-gray-800 dark:bg-gray-800/50 dark:text-gray-400">
							<th class="px-4 py-3">Customer</th>
							<th class="px-4 py-3">Contact Email</th>
							<th class="px-4 py-3">Phone</th>
							<th class="px-4 py-3">KYC Status</th>
							<th class="px-4 py-3">Account State</th>
							<th class="px-4 py-3 text-right">Current Balance</th>
							<th class="px-4 py-3 text-right">Joined</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
						{#each filtered as c}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30">
								<td class="px-4 py-3 font-semibold text-gray-950 dark:text-white">{c.name}</td>
								<td class="px-4 py-3 text-gray-600 dark:text-gray-400">{c.email}</td>
								<td class="px-4 py-3 font-mono">{c.phone}</td>
								<td class="px-4 py-3">
									<span class="inline-flex rounded-full px-2 py-0.5 text-[10px] font-bold uppercase {c.kyc_status === 'verified' ? 'bg-emerald-50 text-emerald-700' : 'bg-amber-50 text-amber-700'}">
										{c.kyc_status || 'Unverified'}
									</span>
								</td>
								<td class="px-4 py-3 font-semibold text-emerald-600 uppercase text-[11px]">{c.status}</td>
								<td class="px-4 py-3 text-right font-bold text-gray-900 dark:text-white">৳{Number(c.balance || 0).toFixed(2)}</td>
								<td class="px-4 py-3 text-right text-gray-400">{new Date(c.created_at).toLocaleDateString()}</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</div>
</div>

