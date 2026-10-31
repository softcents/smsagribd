<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface SecurityActivityItem {
		id: number;
		event: string;
		user: string;
		severity: string;
		ip_address: string;
		source_ip: string;
		details: string;
		created_at: string;
	}

	let logs = $state<SecurityActivityItem[]>([]);
	let loading = $state(true);

	const exportColumns: ExportColumn<SecurityActivityItem>[] = [
		{ header: 'ID', getValue: (l) => l.id },
		{ header: 'Timestamp', getValue: (l) => new Date(l.created_at).toLocaleString() },
		{ header: 'User / Operator', getValue: (l) => l.user },
		{ header: 'Security Event', getValue: (l) => l.event.toUpperCase() },
		{ header: 'Target IP', getValue: (l) => l.ip_address },
		{ header: 'Remote IP', getValue: (l) => l.source_ip || '—' },
		{ header: 'Details', getValue: (l) => l.details || '—' }
	];

	async function loadLogs() {
		loading = true;
		try {
			const res = await supermanApi.get<{ data: SecurityActivityItem[] }>('/api/v1/superman/audit-logs/security');
			logs = res.data ?? [];
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		loadLogs();
	});

	function getSeverityBadge(severity: string) {
		const s = severity.toUpperCase();
		if (s === 'SUCCESS') {
			return 'bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-950/40 dark:text-emerald-400 dark:border-emerald-800';
		}
		if (s === 'WARNING') {
			return 'bg-amber-50 text-amber-700 border-amber-200 dark:bg-amber-950/40 dark:text-amber-400 dark:border-amber-800';
		}
		if (s === 'DANGER') {
			return 'bg-rose-50 text-rose-700 border-rose-200 dark:bg-rose-950/40 dark:text-rose-400 dark:border-rose-800';
		}
		return 'bg-blue-50 text-blue-700 border-blue-200 dark:bg-blue-950/40 dark:text-blue-400 dark:border-blue-800';
	}
</script>

<div class="space-y-6">
	<!-- Page Header -->
	<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
		<div>
			<h1 class="text-xl font-bold text-gray-900 dark:text-white">Security & Whitelist Access Activity</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
				Real-time audit log of authenticated IP requests, policy enforcement, origin hosts, and security events.
			</p>
		</div>
		<div class="flex items-center gap-3">
			<button
				onclick={loadLogs}
				class="inline-flex items-center gap-1.5 px-3 py-2 text-xs font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 dark:bg-gray-800 dark:text-gray-300 dark:border-gray-700 dark:hover:bg-gray-700/60 shadow-xs transition"
			>
				<svg class="w-3.5 h-3.5 {loading ? 'animate-spin' : ''}" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
				</svg>
				Refresh
			</button>
			<ExportDropdown
				title="Security and Whitelist Access Activity"
				columns={exportColumns}
				getData={() => logs}
				disabled={logs.length === 0}
			/>
		</div>
	</div>

	<!-- Table -->
	<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl shadow-xs overflow-hidden">
		<div class="overflow-x-auto">
			<table class="w-full text-left text-xs">
				<thead class="bg-gray-50/75 dark:bg-gray-800/50 text-gray-500 uppercase tracking-wider font-semibold border-b border-gray-200 dark:border-gray-800">
					<tr>
						<th class="px-5 py-3">Timestamp</th>
						<th class="px-5 py-3">User / Operator</th>
						<th class="px-5 py-3">Security Event</th>
						<th class="px-5 py-3">Target IP</th>
						<th class="px-5 py-3">Remote IP</th>
						<th class="px-5 py-3">Details</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800 text-gray-700 dark:text-gray-300">
					{#if loading}
						<tr>
							<td colspan="6" class="px-5 py-10 text-center text-gray-400">Loading security events...</td>
						</tr>
					{:else if logs.length === 0}
						<tr>
							<td colspan="6" class="px-5 py-10 text-center text-gray-400">No security incidents recorded.</td>
						</tr>
					{:else}
						{#each logs as l}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30 transition">
								<td class="px-5 py-3.5 whitespace-nowrap text-gray-500 dark:text-gray-400 font-mono">
									{new Date(l.created_at).toLocaleString('en-GB', { day: '2-digit', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit', second: '2-digit', hour12: true })}
								</td>
								<td class="px-5 py-3.5 font-bold text-gray-900 dark:text-white">{l.user}</td>
								<td class="px-5 py-3.5">
									<span class="inline-flex items-center px-2 py-0.5 rounded-full text-[11px] font-semibold border {getSeverityBadge(l.severity)}">
										{l.event.toUpperCase()}
									</span>
								</td>
								<td class="px-5 py-3.5 font-mono font-bold text-emerald-600 dark:text-emerald-400">{l.ip_address}</td>
								<td class="px-5 py-3.5 font-mono text-gray-500">{l.source_ip || '—'}</td>
								<td class="px-5 py-3.5 text-gray-600 dark:text-gray-400 max-w-sm truncate" title={l.details}>{l.details || '—'}</td>
							</tr>
						{/each}
					{/if}
				</tbody>
			</table>
		</div>
	</div>
</div>
