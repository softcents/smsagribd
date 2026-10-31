<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface ResellerAuditLog {
		id: number;
		admin_name: string;
		customer_name: string;
		action: string;
		field: string;
		old_value: string;
		new_value: string;
		created_at: string;
	}

	let logs = $state<ResellerAuditLog[]>([]);
	let loading = $state(true);

	const exportColumns: ExportColumn<ResellerAuditLog>[] = [
		{ header: 'ID', getValue: (l) => l.id },
		{ header: 'Admin', getValue: (l) => l.admin_name },
		{ header: 'Reseller Target', getValue: (l) => l.customer_name },
		{ header: 'Action Type', getValue: (l) => l.action },
		{ header: 'Field Modified', getValue: (l) => l.field },
		{ header: 'Old Value', getValue: (l) => l.old_value },
		{ header: 'New Value', getValue: (l) => l.new_value },
		{ header: 'Timestamp', getValue: (l) => l.created_at }
	];

	onMount(async () => {
		try {
			const res = await supermanApi.get<{ data: ResellerAuditLog[] }>('/api/v1/superman/sps/audit');
			logs = res.data ?? [];
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
			<h1 class="text-xl font-bold text-gray-900 dark:text-white">Reseller Audit Trail & Management History</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
				Audit trail of administrative actions, tariff alterations, wallet adjustments, and status changes on SP accounts.
			</p>
		</div>
		<ExportDropdown
			title="Reseller Audit Trail History"
			columns={exportColumns}
			getData={() => logs}
			disabled={logs.length === 0}
		/>
	</div>

	<!-- Audit Log Table -->
	<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl shadow-xs overflow-hidden">
		<div class="overflow-x-auto">
			<table class="w-full text-left text-xs">
				<thead class="bg-gray-50/75 dark:bg-gray-800/50 text-gray-500 uppercase tracking-wider font-semibold border-b border-gray-200 dark:border-gray-800">
					<tr>
						<th class="px-5 py-3">Admin</th>
						<th class="px-5 py-3">Reseller Target</th>
						<th class="px-5 py-3">Action Type</th>
						<th class="px-5 py-3">Field</th>
						<th class="px-5 py-3">Old Value</th>
						<th class="px-5 py-3">New Value</th>
						<th class="px-5 py-3">Timestamp</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800 text-gray-700 dark:text-gray-300">
					{#if loading}
						<tr>
							<td colspan="7" class="px-5 py-8 text-center text-gray-400">Loading reseller audit trail...</td>
						</tr>
					{:else if logs.length === 0}
						<tr>
							<td colspan="7" class="px-5 py-8 text-center text-gray-400">No reseller audit entries found.</td>
						</tr>
					{:else}
						{#each logs as l}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30 transition">
								<td class="px-5 py-3 font-bold text-gray-900 dark:text-white">{l.admin_name}</td>
								<td class="px-5 py-3 font-semibold text-emerald-600 dark:text-emerald-400">{l.customer_name}</td>
								<td class="px-5 py-3">
									<span class="inline-flex items-center px-2 py-0.5 rounded text-[11px] font-medium bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300">
										{l.action}
									</span>
								</td>
								<td class="px-5 py-3 font-mono text-gray-500">{l.field}</td>
								<td class="px-5 py-3 font-mono text-gray-400">{l.old_value || 'none'}</td>
								<td class="px-5 py-3 font-mono text-gray-900 dark:text-white font-semibold">{l.new_value || 'none'}</td>
								<td class="px-5 py-3 text-gray-400">{new Date(l.created_at).toLocaleString()}</td>
							</tr>
						{/each}
					{/if}
				</tbody>
			</table>
		</div>
	</div>
</div>

