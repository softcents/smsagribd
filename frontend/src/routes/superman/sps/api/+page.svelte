<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface SpApiCredential {
		id: number;
		name: string;
		phone: string;
		api_token: string;
		ip_whitelist: string;
		created_at: string;
	}

	let apis = $state<SpApiCredential[]>([]);
	let loading = $state(true);

	const exportColumns: ExportColumn<SpApiCredential>[] = [
		{ header: 'ID', getValue: (a) => a.id },
		{ header: 'Reseller', getValue: (a) => a.name },
		{ header: 'Phone', getValue: (a) => a.phone },
		{ header: 'API Token', getValue: (a) => a.api_token },
		{ header: 'IP Whitelist', getValue: (a) => a.ip_whitelist || 'Any IP' },
		{ header: 'Created At', getValue: (a) => a.created_at }
	];

	onMount(async () => {
		try {
			const res = await supermanApi.get<{ data: SpApiCredential[] }>('/api/v1/superman/sps/api');
			apis = res.data ?? [];
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
			<h1 class="text-xl font-bold text-gray-900 dark:text-white">Reseller Developer APIs & Credentials</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
				Bearer API tokens, IP restriction firewalls, and programmatic endpoint quotas assigned to resellers.
			</p>
		</div>
		<ExportDropdown
			title="Reseller Developer APIs and Credentials"
			columns={exportColumns}
			getData={() => apis}
			disabled={apis.length === 0}
		/>
	</div>

	<!-- Credentials Table -->
	<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl shadow-xs overflow-hidden">
		<div class="overflow-x-auto">
			<table class="w-full text-left text-xs">
				<thead class="bg-gray-50/75 dark:bg-gray-800/50 text-gray-500 uppercase tracking-wider font-semibold border-b border-gray-200 dark:border-gray-800">
					<tr>
						<th class="px-5 py-3">Reseller</th>
						<th class="px-5 py-3">Phone</th>
						<th class="px-5 py-3">API Bearer Token</th>
						<th class="px-5 py-3">IP Whitelist Binding</th>
						<th class="px-5 py-3 text-right">Actions</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800 text-gray-700 dark:text-gray-300">
					{#if loading}
						<tr>
							<td colspan="5" class="px-5 py-8 text-center text-gray-400">Loading reseller API credentials...</td>
						</tr>
					{:else if apis.length === 0}
						<tr>
							<td colspan="5" class="px-5 py-8 text-center text-gray-400">No reseller API tokens active.</td>
						</tr>
					{:else}
						{#each apis as a}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30 transition">
								<td class="px-5 py-3 font-bold text-gray-900 dark:text-white">{a.name}</td>
								<td class="px-5 py-3 font-mono">{a.phone}</td>
								<td class="px-5 py-3 font-mono text-gray-500">{a.api_token.slice(0, 16)}••••••••</td>
								<td class="px-5 py-3 font-mono text-xs">{a.ip_whitelist}</td>
								<td class="px-5 py-3 text-right">
									<button
										type="button"
										class="text-xs text-red-600 hover:text-red-700 font-semibold"
										onclick={() => alert(`Revoking API token for ${a.name}`)}
									>
										Revoke Key
									</button>
								</td>
							</tr>
						{/each}
					{/if}
				</tbody>
			</table>
		</div>
	</div>
</div>

