<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface AuditLogItem {
		id: number;
		admin_name: string;
		action: string;
		category: string;
		target: string;
		details: string;
		ip_address?: string;
		created_at: string;
	}

	let logs = $state<AuditLogItem[]>([]);
	let search = $state('');
	let category = $state('');
	let loading = $state(true);
	let page = $state(1);
	let perPage = $state(25);
	let total = $state(0);

	const exportColumns: ExportColumn<AuditLogItem>[] = [
		{ header: 'ID', getValue: (l) => l.id },
		{ header: 'Admin / Actor', getValue: (l) => l.admin_name },
		{ header: 'Action', getValue: (l) => l.action },
		{ header: 'Category', getValue: (l) => l.category },
		{ header: 'Target', getValue: (l) => l.target },
		{ header: 'Details', getValue: (l) => l.details },
		{ header: 'IP Address', getValue: (l) => l.ip_address || '—' },
		{ header: 'Timestamp', getValue: (l) => l.created_at }
	];

	async function loadLogs() {
		loading = true;
		try {
			const res = await supermanApi.get<{ data: AuditLogItem[]; meta: { total: number } }>(
				`/api/v1/superman/audit-logs/global?page=${page}&per_page=${perPage}&category=${encodeURIComponent(category)}&search=${encodeURIComponent(search)}`
			);
			logs = res.data ?? [];
			total = res.meta?.total ?? 0;
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	}

	onMount(loadLogs);

	function handleSearch() {
		page = 1;
		loadLogs();
	}
</script>

<div class="space-y-6">
	<!-- Page Header -->
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-xl font-bold text-gray-900 dark:text-white">Global Audit Log & Event Ledger</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
				Authoritative chronological trail of administrative modifications, KYC actions, security triggers, and wallet transactions.
			</p>
		</div>
		<div class="flex items-center gap-2">
			<ExportDropdown
				title="Global Audit Log Event Ledger"
				columns={exportColumns}
				getData={() => logs}
				disabled={logs.length === 0}
			/>
			<select
				bind:value={category}
				onchange={handleSearch}
				class="rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 px-3 py-1.5 text-xs text-gray-900 dark:text-white focus:outline-none focus:ring-1 focus:ring-emerald-500 shadow-xs"
			>
				<option value="">All Categories</option>
				<option value="KYC">KYC Verification</option>
				<option value="IP_WHITELIST">IP Whitelist</option>
			</select>
			<input
				type="text"
				placeholder="Search details, admin..."
				bind:value={search}
				onkeydown={(e) => { if (e.key === 'Enter') handleSearch(); }}
				class="rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 px-3 py-1.5 text-xs text-gray-900 dark:text-white focus:outline-none focus:ring-1 focus:ring-emerald-500 w-56 shadow-xs"
			/>
			<button
				type="button"
				onclick={handleSearch}
				class="rounded-lg bg-emerald-600 px-3 py-1.5 text-xs font-semibold text-white shadow-xs hover:bg-emerald-700 transition"
			>
				Filter
			</button>
		</div>
	</div>

	<!-- Audit Log Table -->
	<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl shadow-xs overflow-hidden">
		<div class="overflow-x-auto">
			<table class="w-full text-left text-xs">
				<thead class="bg-gray-50/75 dark:bg-gray-800/50 text-gray-500 uppercase tracking-wider font-semibold border-b border-gray-200 dark:border-gray-800">
					<tr>
						<th class="px-5 py-3">Admin</th>
						<th class="px-5 py-3">Category</th>
						<th class="px-5 py-3">Action</th>
						<th class="px-5 py-3">Target Subject</th>
						<th class="px-5 py-3">Details</th>
						<th class="px-5 py-3">IP Address</th>
						<th class="px-5 py-3">Timestamp</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800 text-gray-700 dark:text-gray-300">
					{#if loading}
						<tr>
							<td colspan="7" class="px-5 py-8 text-center text-gray-400">Loading audit events...</td>
						</tr>
					{:else if logs.length === 0}
						<tr>
							<td colspan="7" class="px-5 py-8 text-center text-gray-400">No audit events match filters.</td>
						</tr>
					{:else}
						{#each logs as l}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30 transition">
								<td class="px-5 py-3 font-bold text-gray-900 dark:text-white">{l.admin_name}</td>
								<td class="px-5 py-3">
									<span class="inline-flex items-center px-2 py-0.5 rounded text-[10px] font-semibold tracking-wider font-mono bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200 uppercase">
										{l.category}
									</span>
								</td>
								<td class="px-5 py-3 font-semibold text-emerald-600 dark:text-emerald-400">{l.action}</td>
								<td class="px-5 py-3 font-medium text-gray-800 dark:text-gray-200">{l.target}</td>
								<td class="px-5 py-3 text-gray-600 dark:text-gray-400 font-mono text-[11px]">{l.details}</td>
								<td class="px-5 py-3 font-mono text-gray-400">{l.ip_address || '—'}</td>
								<td class="px-5 py-3 text-gray-400">{new Date(l.created_at).toLocaleString()}</td>
							</tr>
						{/each}
					{/if}
				</tbody>
			</table>
		</div>
	</div>
</div>

