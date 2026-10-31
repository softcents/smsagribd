<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface IpAuditLog {
		id: number;
		actor_name: string;
		actor_type: string;
		user_id: number;
		customer_name: string;
		customer_email: string;
		action: string;
		ip_address: string;
		note: string;
		source_ip: string;
		old_values: string;
		new_values: string;
		created_at: string;
	}

	let logs = $state<IpAuditLog[]>([]);
	let loading = $state(true);
	let selectedLog = $state<IpAuditLog | null>(null);

	const exportColumns: ExportColumn<IpAuditLog>[] = [
		{ header: 'ID', getValue: (l) => l.id },
		{ header: 'Timestamp', getValue: (l) => new Date(l.created_at).toLocaleString() },
		{ header: 'Actor', getValue: (l) => `${l.actor_name} (${l.actor_type})` },
		{ header: 'Customer', getValue: (l) => l.customer_name },
		{ header: 'Customer Email', getValue: (l) => l.customer_email },
		{ header: 'Action', getValue: (l) => l.action.toUpperCase() },
		{ header: 'IP / CIDR', getValue: (l) => l.ip_address },
		{ header: 'Note / Reason', getValue: (l) => l.note || '—' },
		{ header: 'Origin IP', getValue: (l) => l.source_ip || '—' }
	];

	async function loadLogs() {
		loading = true;
		try {
			const res = await supermanApi.get<{ data: IpAuditLog[] }>('/api/v1/superman/audit-logs/ip-whitelist');
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

	function getActionBadge(action: string) {
		const act = action.toLowerCase();
		if (act === 'approved' || act === 'enabled') {
			return 'bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-950/40 dark:text-emerald-400 dark:border-emerald-800';
		}
		if (act === 'rejected' || act === 'disabled') {
			return 'bg-rose-50 text-rose-700 border-rose-200 dark:bg-rose-950/40 dark:text-rose-400 dark:border-rose-800';
		}
		if (act === 'updated') {
			return 'bg-amber-50 text-amber-700 border-amber-200 dark:bg-amber-950/40 dark:text-amber-400 dark:border-amber-800';
		}
		if (act === 'created') {
			return 'bg-blue-50 text-blue-700 border-blue-200 dark:bg-blue-950/40 dark:text-blue-400 dark:border-blue-800';
		}
		return 'bg-gray-50 text-gray-700 border-gray-200 dark:bg-gray-800 dark:text-gray-300 dark:border-gray-700';
	}
</script>

<div class="space-y-6">
	<!-- Page Header -->
	<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
		<div>
			<h1 class="text-xl font-bold text-gray-900 dark:text-white">IP Whitelist Audit Log</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
				Real-time audit log of customer static IP requests, Superman approvals, policy state mutations, and origin telemetry.
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
				title="IP Whitelist Audit Log"
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
						<th class="px-5 py-3">Actor</th>
						<th class="px-5 py-3">Customer</th>
						<th class="px-5 py-3">Action</th>
						<th class="px-5 py-3">IP / CIDR</th>
						<th class="px-5 py-3">Note / Reason</th>
						<th class="px-5 py-3">Origin IP</th>
						<th class="px-5 py-3 text-right">Actions</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800 text-gray-700 dark:text-gray-300">
					{#if loading}
						<tr>
							<td colspan="8" class="px-5 py-10 text-center text-gray-400">Loading IP whitelist audit trail...</td>
						</tr>
					{:else if logs.length === 0}
						<tr>
							<td colspan="8" class="px-5 py-10 text-center text-gray-400">No IP audit records found in database.</td>
						</tr>
					{:else}
						{#each logs as l}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30 transition">
								<td class="px-5 py-3.5 whitespace-nowrap text-gray-500 dark:text-gray-400 font-mono">
									{new Date(l.created_at).toLocaleString('en-GB', { day: '2-digit', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit', second: '2-digit', hour12: true })}
								</td>
								<td class="px-5 py-3.5">
									<div class="font-bold text-gray-900 dark:text-white">{l.actor_name}</div>
									<div class="text-[11px] text-gray-400 capitalize">{l.actor_type}</div>
								</td>
								<td class="px-5 py-3.5">
									<a href="/superman/users/customers/{l.user_id}" class="font-semibold text-emerald-600 hover:text-emerald-700 dark:text-emerald-400 dark:hover:underline">
										{l.customer_name}
									</a>
									{#if l.customer_email}
										<div class="text-[11px] text-gray-400">{l.customer_email}</div>
									{/if}
								</td>
								<td class="px-5 py-3.5">
									<span class="inline-flex items-center px-2 py-0.5 rounded-full text-[11px] font-semibold border {getActionBadge(l.action)}">
										{l.action.toUpperCase()}
									</span>
								</td>
								<td class="px-5 py-3.5 font-mono font-bold text-gray-800 dark:text-gray-200">
									{l.ip_address}
								</td>
								<td class="px-5 py-3.5 max-w-xs truncate text-gray-600 dark:text-gray-400" title={l.note}>
									{l.note || '—'}
								</td>
								<td class="px-5 py-3.5 font-mono text-gray-500">
									{l.source_ip || '—'}
								</td>
								<td class="px-5 py-3.5 text-right">
									<button
										onclick={() => selectedLog = l}
										class="inline-flex items-center gap-1 px-2.5 py-1 text-xs font-medium text-emerald-700 bg-emerald-50 hover:bg-emerald-100 dark:bg-emerald-950/40 dark:text-emerald-400 rounded-md transition"
									>
										<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
										</svg>
										View
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

<!-- Details Modal -->
{#if selectedLog}
	<div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-xs">
		<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-2xl max-w-lg w-full p-6 shadow-xl space-y-4">
			<div class="flex items-center justify-between border-b border-gray-100 dark:border-gray-800 pb-3">
				<div>
					<h3 class="text-base font-bold text-gray-900 dark:text-white">Audit Details: {selectedLog.action.toUpperCase()} {selectedLog.ip_address}</h3>
					<p class="text-xs text-gray-400 mt-0.5">Audit Record #{selectedLog.id}</p>
				</div>
				<button onclick={() => selectedLog = null} class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300">
					<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
				</button>
			</div>

			<div class="space-y-3 text-xs">
				<div>
					<span class="text-gray-400 font-medium">Performed By:</span>
					<div class="font-bold text-gray-900 dark:text-white mt-0.5">{selectedLog.actor_name} ({selectedLog.actor_type})</div>
				</div>
				<div>
					<span class="text-gray-400 font-medium">Customer Affected:</span>
					<div class="font-semibold text-emerald-600 dark:text-emerald-400 mt-0.5">{selectedLog.customer_name} ({selectedLog.customer_email || 'No email'})</div>
				</div>
				<div>
					<span class="text-gray-400 font-medium">Originating IP Address:</span>
					<div class="font-mono text-gray-800 dark:text-gray-200 mt-0.5">{selectedLog.source_ip || 'Unknown'}</div>
				</div>
				<div>
					<span class="text-gray-400 font-medium">Note / Reason:</span>
					<div class="text-gray-700 dark:text-gray-300 bg-gray-50 dark:bg-gray-800/60 p-2.5 rounded-lg border border-gray-200 dark:border-gray-700/60 mt-0.5">
						{selectedLog.note || 'No remarks recorded.'}
					</div>
				</div>
				{#if selectedLog.old_values}
					<div>
						<span class="text-gray-400 font-medium">Previous Values:</span>
						<pre class="font-mono text-[11px] bg-gray-50 dark:bg-gray-800/60 p-2.5 rounded-lg border border-gray-200 dark:border-gray-700/60 overflow-x-auto mt-0.5">{selectedLog.old_values}</pre>
					</div>
				{/if}
				{#if selectedLog.new_values}
					<div>
						<span class="text-gray-400 font-medium">Applied Values:</span>
						<pre class="font-mono text-[11px] bg-emerald-50 dark:bg-emerald-950/30 text-emerald-700 dark:text-emerald-300 p-2.5 rounded-lg border border-emerald-200 dark:border-emerald-800/60 overflow-x-auto mt-0.5">{selectedLog.new_values}</pre>
					</div>
				{/if}
			</div>

			<div class="pt-2 flex justify-end">
				<button
					onclick={() => selectedLog = null}
					class="px-4 py-2 text-xs font-medium text-gray-700 bg-gray-100 hover:bg-gray-200 dark:bg-gray-800 dark:text-gray-300 dark:hover:bg-gray-700 rounded-lg transition"
				>
					Close
				</button>
			</div>
		</div>
	</div>
{/if}
