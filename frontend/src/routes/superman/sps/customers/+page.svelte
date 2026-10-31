<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface SubCustomer {
		id: number;
		name: string;
		phone: string;
		email: string;
		balance: number;
		kyc_status: string;
		account_status: string;
		reseller_name: string;
		sp_id: number;
		created_at: string;
	}

	let customers = $state<SubCustomer[]>([]);
	let search = $state('');
	let loading = $state(true);
	let page = $state(1);
	let perPage = $state(25);
	let total = $state(0);

	const exportColumns: ExportColumn<SubCustomer>[] = [
		{ header: 'ID', getValue: (c) => c.id },
		{ header: 'Customer', getValue: (c) => c.name },
		{ header: 'Phone', getValue: (c) => c.phone },
		{ header: 'Email', getValue: (c) => c.email },
		{ header: 'Parent Reseller', getValue: (c) => c.reseller_name },
		{ header: 'Balance (BDT)', getValue: (c) => `BDT ${c.balance.toFixed(2)}` },
		{ header: 'KYC Status', getValue: (c) => c.kyc_status },
		{ header: 'Account Status', getValue: (c) => c.account_status },
		{ header: 'Created At', getValue: (c) => c.created_at }
	];

	async function loadCustomers() {
		loading = true;
		try {
			const res = await supermanApi.get<{ data: SubCustomer[]; meta: { total: number } }>(
				`/api/v1/superman/sps/customers?page=${page}&per_page=${perPage}&search=${encodeURIComponent(search)}`
			);
			customers = res.data ?? [];
			total = res.meta?.total ?? 0;
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	}

	onMount(loadCustomers);

	function handleSearch() {
		page = 1;
		loadCustomers();
	}
</script>

<div class="space-y-6">
	<!-- Page Header -->
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-xl font-bold text-gray-900 dark:text-white">Reseller-Affiliated Sub-Customers</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
				Customers created by or mapped under Service Provider (SP) Reseller accounts.
			</p>
		</div>
		<div class="flex items-center gap-2">
			<ExportDropdown
				title="Reseller Affiliated Sub-Customers"
				columns={exportColumns}
				getData={() => customers}
				disabled={customers.length === 0}
			/>
			<form onsubmit={(e) => { e.preventDefault(); handleSearch(); }} class="flex items-center gap-2">
				<input
					type="text"
					placeholder="Search sub-customers..."
					bind:value={search}
					class="rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 px-3 py-1.5 text-xs text-gray-900 dark:text-white focus:outline-none focus:ring-1 focus:ring-emerald-500 w-64 shadow-xs"
				/>
				<button
					type="submit"
					class="rounded-lg bg-emerald-600 px-3 py-1.5 text-xs font-semibold text-white shadow-xs hover:bg-emerald-700 transition"
				>
					Search
				</button>
			</form>
		</div>
	</div>

	<!-- Sub-customers Table -->
	<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl shadow-xs overflow-hidden">
		<div class="overflow-x-auto">
			<table class="w-full text-left text-xs">
				<thead class="bg-gray-50/75 dark:bg-gray-800/50 text-gray-500 uppercase tracking-wider font-semibold border-b border-gray-200 dark:border-gray-800">
					<tr>
						<th class="px-5 py-3">Customer</th>
						<th class="px-5 py-3">Parent Reseller</th>
						<th class="px-5 py-3">Phone</th>
						<th class="px-5 py-3">Balance (BDT)</th>
						<th class="px-5 py-3">KYC Status</th>
						<th class="px-5 py-3">Account Status</th>
						<th class="px-5 py-3 text-right">Actions</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800 text-gray-700 dark:text-gray-300">
					{#if loading}
						<tr>
							<td colspan="7" class="px-5 py-8 text-center text-gray-400">Loading sub-customers...</td>
						</tr>
					{:else if customers.length === 0}
						<tr>
							<td colspan="7" class="px-5 py-8 text-center text-gray-400">No reseller sub-customers found.</td>
						</tr>
					{:else}
						{#each customers as c}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30 transition">
								<td class="px-5 py-3">
									<a href="/superman/users/customers/{c.id}" class="font-bold text-gray-900 dark:text-white hover:text-emerald-600 transition">
										{c.name}
									</a>
									<div class="text-[11px] text-gray-400">{c.email}</div>
								</td>
								<td class="px-5 py-3">
									<a href="/superman/users/sps/{c.sp_id}" class="font-semibold text-emerald-600 dark:text-emerald-400 hover:underline">
										{c.reseller_name}
									</a>
								</td>
								<td class="px-5 py-3 font-mono">{c.phone}</td>
								<td class="px-5 py-3 font-semibold text-gray-900 dark:text-white">৳{c.balance.toFixed(2)}</td>
								<td class="px-5 py-3">
									<span class="inline-flex items-center px-2 py-0.5 rounded text-[11px] font-medium {c.kyc_status === 'verified' ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300' : 'bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-300'}">
										{c.kyc_status}
									</span>
								</td>
								<td class="px-5 py-3">
									<span class="inline-flex items-center px-2 py-0.5 rounded text-[11px] font-medium {c.account_status === 'approved' ? 'bg-blue-100 text-blue-800 dark:bg-blue-950 dark:text-blue-300' : 'bg-red-100 text-red-800 dark:bg-red-950 dark:text-red-300'}">
										{c.account_status}
									</span>
								</td>
								<td class="px-5 py-3 text-right">
									<a href="/superman/users/customers/{c.id}" class="text-xs text-emerald-600 hover:text-emerald-700 font-semibold">
										Inspect
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

