<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface CustomDomain {
		id: number;
		reseller_name: string;
		company_name: string;
		domain: string;
		cname_target: string;
		ssl_status: string;
		status: string;
		created_at: string;
	}

	let domains = $state<CustomDomain[]>([]);
	let loading = $state(true);

	const exportColumns: ExportColumn<CustomDomain>[] = [
		{ header: 'ID', getValue: (d) => d.id },
		{ header: 'Reseller', getValue: (d) => d.reseller_name },
		{ header: 'Company', getValue: (d) => d.company_name || '—' },
		{ header: 'Domain CNAME Status', getValue: (d) => d.domain },
		{ header: 'CNAME Target', getValue: (d) => d.cname_target },
		{ header: 'SSL Status', getValue: (d) => d.ssl_status.toUpperCase() },
		{ header: 'Partner Status', getValue: (d) => d.status.toUpperCase() },
		{ header: 'Created Date', getValue: (d) => new Date(d.created_at).toLocaleDateString() }
	];

	async function loadDomains() {
		loading = true;
		try {
			const res = await supermanApi.get<{ data: CustomDomain[] }>('/api/v1/superman/sps/custom-domain');
			domains = res.data ?? [];
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		loadDomains();
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
			<h1 class="text-xl font-bold text-gray-900 dark:text-white">Reseller Custom Domain & CNAME Mapping</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
				Automated CNAME mapping, SSL provisioning, and domain ingress endpoints for wholesale partner portals.
			</p>
		</div>
		<div class="flex items-center gap-3">
			<button
				onclick={loadDomains}
				class="inline-flex items-center gap-1.5 px-3 py-2 text-xs font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 dark:bg-gray-800 dark:text-gray-300 dark:border-gray-700 dark:hover:bg-gray-700/60 shadow-xs transition"
			>
				<svg class="w-3.5 h-3.5 {loading ? 'animate-spin' : ''}" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
				</svg>
				Refresh
			</button>
			<ExportDropdown
				title="Custom Domain Mappings"
				columns={exportColumns}
				getData={() => domains}
				disabled={domains.length === 0}
			/>
		</div>
	</div>

	<!-- Custom Domains Table -->
	<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl shadow-xs overflow-hidden">
		<div class="overflow-x-auto">
			<table class="w-full text-left text-xs">
				<thead class="bg-gray-50/75 dark:bg-gray-800/50 text-gray-500 uppercase tracking-wider font-semibold border-b border-gray-200 dark:border-gray-800">
					<tr>
						<th class="px-5 py-3">Reseller</th>
						<th class="px-5 py-3">Company</th>
						<th class="px-5 py-3">Domain CNAME Status</th>
						<th class="px-5 py-3 text-center">Status</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800 text-gray-700 dark:text-gray-300">
					{#if loading}
						<tr>
							<td colspan="4" class="px-5 py-10 text-center text-gray-400">Loading custom domain mappings...</td>
						</tr>
					{:else if domains.length === 0}
						<tr>
							<td colspan="4" class="px-5 py-10 text-center text-gray-400">No reseller domains configured.</td>
						</tr>
					{:else}
						{#each domains as d}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30 transition">
								<td class="px-5 py-3.5 font-bold text-gray-900 dark:text-white">
									{d.reseller_name}
								</td>
								<td class="px-5 py-3.5 text-gray-500 dark:text-gray-400">
									{d.company_name || '—'}
								</td>
								<td class="px-5 py-3.5">
									<span class="inline-flex items-center px-2.5 py-1 rounded-md text-xs font-mono font-bold bg-emerald-50 text-emerald-800 border border-emerald-200 dark:bg-emerald-950/40 dark:text-emerald-300 dark:border-emerald-800">
										{d.domain}
									</span>
								</td>
								<td class="px-5 py-3.5 text-center">
									<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-[11px] font-semibold border {getStatusBadge(d.status)}">
										{d.status.toUpperCase()}
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
