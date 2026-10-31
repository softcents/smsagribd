<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface SecurityLog {
		id: number;
		admin_name: string;
		action: string;
		category: string;
		target: string;
		details: string;
		ip_address?: string;
		created_at: string;
	}

	let logs = $state<SecurityLog[]>([]);
	let loading = $state(true);

	onMount(async () => {
		try {
			const res = await supermanApi.get<{ data: SecurityLog[] }>('/api/v1/superman/security/logs');
			logs = res.data ?? [];
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	});

	const exportColumns: ExportColumn<SecurityLog>[] = [
		{ header: 'Actor / Admin', getValue: (l) => l.admin_name },
		{ header: 'Action', getValue: (l) => l.action },
		{ header: 'Category', getValue: (l) => l.category },
		{ header: 'Target Subject', getValue: (l) => l.target },
		{ header: 'Details / Delta', getValue: (l) => l.details },
		{ header: 'Source IP', getValue: (l) => l.ip_address || '—' },
		{ header: 'Timestamp', getValue: (l) => (l.created_at ? new Date(l.created_at).toLocaleString() : '—') }
	];
</script>

<div class="space-y-6">
	<!-- Page Header -->
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-xl font-bold text-gray-900 dark:text-white">Platform Security Audit Logs</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
				Real-time audit log of access events, security policies, token operations, and whitelist validations.
			</p>
		</div>
		<div class="flex items-center gap-2">
			<ExportDropdown
				title="Security Audit Logs"
				columns={exportColumns}
				getData={() => logs}
				disabled={logs.length === 0}
			/>
		</div>
	</div>

	<!-- Logs Table -->
	<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl shadow-xs overflow-hidden">
		<div class="overflow-x-auto">
			<table class="w-full text-left text-xs">
				<thead class="bg-gray-50/75 dark:bg-gray-800/50 text-gray-500 uppercase tracking-wider font-semibold border-b border-gray-200 dark:border-gray-800">
					<tr>
						<th class="px-5 py-3">Actor / Admin</th>
						<th class="px-5 py-3">Action</th>
						<th class="px-5 py-3">Target Subject</th>
						<th class="px-5 py-3">Details / Delta</th>
						<th class="px-5 py-3">Source IP</th>
						<th class="px-5 py-3">Timestamp</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800 text-gray-700 dark:text-gray-300">
					{#if loading}
						<tr>
							<td colspan="6" class="px-5 py-8 text-center text-gray-400">Loading security audit records...</td>
						</tr>
					{:else if logs.length === 0}
						<tr>
							<td colspan="6" class="px-5 py-8 text-center text-gray-400">No security audit logs found.</td>
						</tr>
					{:else}
						{#each logs as l}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30 transition">
								<td class="px-5 py-3 font-bold text-gray-900 dark:text-white">{l.admin_name}</td>
								<td class="px-5 py-3">
									<span class="inline-flex items-center px-2 py-0.5 rounded text-[11px] font-medium bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300 font-mono">
										{l.action}
									</span>
								</td>
								<td class="px-5 py-3 font-semibold text-gray-800 dark:text-gray-200">{l.target}</td>
								<td class="px-5 py-3 text-gray-600 dark:text-gray-400">{l.details}</td>
								<td class="px-5 py-3 font-mono text-gray-500">{l.ip_address || 'Internal Host'}</td>
								<td class="px-5 py-3 text-gray-400">{new Date(l.created_at).toLocaleString()}</td>
							</tr>
						{/each}
					{/if}
				</tbody>
			</table>
		</div>
	</div>
</div>

