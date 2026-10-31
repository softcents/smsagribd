<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface ResellerBrand {
		id: number;
		reseller: string;
		brand_name: string;
		support_email: string;
		phone: string;
		status: string;
		created_at: string;
	}

	interface WhiteLabelResponse {
		data: ResellerBrand[];
	}

	let brands = $state<ResellerBrand[]>([]);
	let loading = $state(true);

	const exportColumns: ExportColumn<ResellerBrand>[] = [
		{ header: 'ID', getValue: (b) => b.id },
		{ header: 'Reseller', getValue: (b) => b.reseller },
		{ header: 'Brand Name', getValue: (b) => b.brand_name || '—' },
		{ header: 'Support Email', getValue: (b) => b.support_email },
		{ header: 'Support Helpline', getValue: (b) => b.phone },
		{ header: 'Account Status', getValue: (b) => b.status.toUpperCase() },
		{ header: 'Onboarding Date', getValue: (b) => new Date(b.created_at).toLocaleDateString() }
	];

	async function loadBrands() {
		loading = true;
		try {
			const res = await supermanApi.get<WhiteLabelResponse>('/api/v1/superman/sps/white-label');
			brands = res.data ?? [];
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		loadBrands();
	});

	function getStatusBadge(status: string) {
		const s = status.toLowerCase();
		if (s === 'approved' || s === 'active') {
			return 'bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-950/40 dark:text-emerald-400 dark:border-emerald-800';
		}
		if (s === 'banned' || s === 'suspended') {
			return 'bg-rose-50 text-rose-700 border-rose-200 dark:bg-rose-950/40 dark:text-rose-400 dark:border-rose-800';
		}
		return 'bg-amber-50 text-amber-700 border-amber-200 dark:bg-amber-950/40 dark:text-amber-400 dark:border-amber-800';
	}
</script>

<div class="space-y-6">
	<!-- Page Header -->
	<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
		<div>
			<h1 class="text-xl font-bold text-gray-900 dark:text-white">Reseller White-Label & Custom Branding</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
				White-label identity mappings, company brands, and customer contact endpoints for Service Provider partners.
			</p>
		</div>
		<div class="flex items-center gap-3">
			<button
				onclick={loadBrands}
				class="inline-flex items-center gap-1.5 px-3 py-2 text-xs font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 dark:bg-gray-800 dark:text-gray-300 dark:border-gray-700 dark:hover:bg-gray-700/60 shadow-xs transition"
			>
				<svg class="w-3.5 h-3.5 {loading ? 'animate-spin' : ''}" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
				</svg>
				Refresh
			</button>
			<ExportDropdown
				title="Reseller White-Label Branding"
				columns={exportColumns}
				getData={() => brands}
				disabled={brands.length === 0}
			/>
		</div>
	</div>

	<!-- Table -->
	<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl shadow-xs overflow-hidden">
		<div class="overflow-x-auto">
			<table class="w-full text-left text-xs">
				<thead class="bg-gray-50/75 dark:bg-gray-800/50 text-gray-500 uppercase tracking-wider font-semibold border-b border-gray-200 dark:border-gray-800">
					<tr>
						<th class="px-5 py-3">Reseller</th>
						<th class="px-5 py-3">Brand Name</th>
						<th class="px-5 py-3">Support Email</th>
						<th class="px-5 py-3">Support Helpline</th>
						<th class="px-5 py-3 text-center">Status</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800 text-gray-700 dark:text-gray-300">
					{#if loading}
						<tr>
							<td colspan="5" class="px-5 py-10 text-center text-gray-400">Loading white-label branding profiles...</td>
						</tr>
					{:else if brands.length === 0}
						<tr>
							<td colspan="5" class="px-5 py-10 text-center text-gray-400">No reseller branding records found.</td>
						</tr>
					{:else}
						{#each brands as b}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30 transition">
								<td class="px-5 py-3.5 font-bold text-gray-900 dark:text-white">
									{b.reseller}
								</td>
								<td class="px-5 py-3.5 font-semibold text-emerald-600 dark:text-emerald-400">
									{b.brand_name || '—'}
								</td>
								<td class="px-5 py-3.5 font-mono text-gray-600 dark:text-gray-400">
									{b.support_email}
								</td>
								<td class="px-5 py-3.5 font-mono text-gray-600 dark:text-gray-400">
									{b.phone}
								</td>
								<td class="px-5 py-3.5 text-center">
									<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-[11px] font-semibold border {getStatusBadge(b.status)}">
										{b.status.toUpperCase()}
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
