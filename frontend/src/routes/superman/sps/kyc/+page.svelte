<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface SpKycItem {
		id: number;
		name: string;
		phone: string;
		business_name: string;
		kyc_status: string;
		trade_license: string;
		created_at: string;
	}

	let items = $state<SpKycItem[]>([]);
	let loading = $state(true);

	const exportColumns: ExportColumn<SpKycItem>[] = [
		{ header: 'ID', getValue: (k) => k.id },
		{ header: 'Reseller', getValue: (k) => k.name },
		{ header: 'Phone', getValue: (k) => k.phone },
		{ header: 'Business / Company', getValue: (k) => k.business_name },
		{ header: 'KYC Status', getValue: (k) => k.kyc_status },
		{ header: 'Trade License / Doc', getValue: (k) => k.trade_license || '—' },
		{ header: 'Submitted Date', getValue: (k) => k.created_at }
	];

	onMount(async () => {
		try {
			const res = await supermanApi.get<{ data: SpKycItem[] }>('/api/v1/superman/sps/kyc');
			items = res.data ?? [];
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
			<h1 class="text-xl font-bold text-gray-900 dark:text-white">Reseller KYC & Business Verification</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
				Trade license, TIN/BIN, National ID, and organization credential verification for Service Provider accounts.
			</p>
		</div>
		<div class="flex items-center gap-3">
			<ExportDropdown
				title="Reseller KYC Verification"
				columns={exportColumns}
				getData={() => items}
				disabled={items.length === 0}
			/>
			<a
				href="/superman/kyc/requests"
				class="inline-flex items-center gap-2 rounded-lg bg-emerald-600 px-3 py-1.5 text-xs font-semibold text-white shadow-xs hover:bg-emerald-700 transition"
			>
				All KYC Verification Requests
			</a>
		</div>
	</div>

	<!-- KYC Table -->
	<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl shadow-xs overflow-hidden">
		<div class="overflow-x-auto">
			<table class="w-full text-left text-xs">
				<thead class="bg-gray-50/75 dark:bg-gray-800/50 text-gray-500 uppercase tracking-wider font-semibold border-b border-gray-200 dark:border-gray-800">
					<tr>
						<th class="px-5 py-3">Reseller Name</th>
						<th class="px-5 py-3">Business Organization</th>
						<th class="px-5 py-3">Phone</th>
						<th class="px-5 py-3">Trade License #</th>
						<th class="px-5 py-3">Verification State</th>
						<th class="px-5 py-3 text-right">Actions</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800 text-gray-700 dark:text-gray-300">
					{#if loading}
						<tr>
							<td colspan="6" class="px-5 py-8 text-center text-gray-400">Loading reseller verification status...</td>
						</tr>
					{:else if items.length === 0}
						<tr>
							<td colspan="6" class="px-5 py-8 text-center text-gray-400">No reseller KYC records found.</td>
						</tr>
					{:else}
						{#each items as it}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30 transition">
								<td class="px-5 py-3 font-bold text-gray-900 dark:text-white">
									<a href="/superman/users/sps/{it.id}" class="hover:text-emerald-600 transition">{it.name}</a>
								</td>
								<td class="px-5 py-3 text-gray-600 dark:text-gray-400">{it.business_name || '—'}</td>
								<td class="px-5 py-3 font-mono">{it.phone}</td>
								<td class="px-5 py-3 font-mono text-gray-500">{it.trade_license || 'Pending Submission'}</td>
								<td class="px-5 py-3">
									<span class="inline-flex items-center px-2 py-0.5 rounded text-[11px] font-medium {it.kyc_status === 'verified' ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300' : 'bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-300'}">
										{it.kyc_status}
									</span>
								</td>
								<td class="px-5 py-3 text-right">
									<a href="/superman/users/sps/{it.id}" class="text-xs text-emerald-600 hover:text-emerald-700 font-semibold">
										Inspect KYC
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

