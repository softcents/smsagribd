<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface FailoverIncident {
		id: number;
		phone: string;
		gateway_name: string;
		event: string;
		error_message: string;
		created_at: string;
	}

	let incidents = $state<FailoverIncident[]>([]);
	let loading = $state(true);

	const exportColumns: ExportColumn<FailoverIncident>[] = [
		{ header: 'ID', getValue: (i) => i.id },
		{ header: 'Timestamp', getValue: (i) => new Date(i.created_at).toLocaleString() },
		{ header: 'Recipient', getValue: (i) => i.phone },
		{ header: 'Failing Provider', getValue: (i) => i.gateway_name },
		{ header: 'Failover Event', getValue: (i) => getEventLabel(i.event) },
		{ header: 'Error Detail / Reason', getValue: (i) => i.error_message }
	];

	function getEventLabel(event: string): string {
		switch (event) {
			case 'sms_gateway_failover':
				return 'TRIGGERED FAILOVER';
			case 'sms_gateway_failover_exception':
				return 'TIMEOUT / CONNECTION ERROR';
			case 'sms_send_failed_all_gateways':
				return 'ALL GATEWAYS EXHAUSTED';
			default:
				return event.replace(/_/g, ' ').toUpperCase();
		}
	}

	function getEventBadge(event: string): string {
		if (event === 'sms_send_failed_all_gateways') {
			return 'bg-rose-50 text-rose-700 border-rose-200 dark:bg-rose-950/40 dark:text-rose-400 dark:border-rose-800';
		}
		if (event === 'sms_gateway_failover_exception') {
			return 'bg-red-50 text-red-700 border-red-200 dark:bg-red-950/40 dark:text-red-400 dark:border-red-800';
		}
		return 'bg-amber-50 text-amber-700 border-amber-200 dark:bg-amber-950/40 dark:text-amber-400 dark:border-amber-800';
	}

	async function loadIncidents() {
		loading = true;
		try {
			const res = await supermanApi.get<{ data: FailoverIncident[] }>('/api/v1/superman/routing/failover');
			incidents = res.data ?? [];
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		loadIncidents();
	});
</script>

<div class="space-y-6">
	<!-- Page Header -->
	<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
		<div>
			<h1 class="text-xl font-bold text-gray-900 dark:text-white">Automatic Failover Monitoring & Logs</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
				Real-time monitoring and log audit of provider network timeouts, connection exceptions, and automated failover reroutes.
			</p>
		</div>
		<div class="flex items-center gap-3">
			<button
				onclick={loadIncidents}
				class="inline-flex items-center gap-1.5 px-3 py-2 text-xs font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 dark:bg-gray-800 dark:text-gray-300 dark:border-gray-700 dark:hover:bg-gray-700/60 shadow-xs transition"
			>
				<svg class="w-3.5 h-3.5 {loading ? 'animate-spin' : ''}" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
				</svg>
				Refresh
			</button>
			<ExportDropdown
				title="Failover Audit & Incidents"
				columns={exportColumns}
				getData={() => incidents}
				disabled={incidents.length === 0}
			/>
		</div>
	</div>

	<!-- Incidents Table -->
	<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl shadow-xs overflow-hidden">
		<div class="overflow-x-auto">
			<table class="w-full text-left text-xs">
				<thead class="bg-gray-50/75 dark:bg-gray-800/50 text-gray-500 uppercase tracking-wider font-semibold border-b border-gray-200 dark:border-gray-800">
					<tr>
						<th class="px-5 py-3">Timestamp</th>
						<th class="px-5 py-3">Recipient</th>
						<th class="px-5 py-3">Failing Provider</th>
						<th class="px-5 py-3">Failover Event</th>
						<th class="px-5 py-3">Error Detail / Reason</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800 text-gray-700 dark:text-gray-300">
					{#if loading}
						<tr>
							<td colspan="5" class="px-5 py-10 text-center text-gray-400">Loading failover incident records...</td>
						</tr>
					{:else if incidents.length === 0}
						<tr>
							<td colspan="5" class="px-5 py-10 text-center text-gray-400">No carrier failovers recorded.</td>
						</tr>
					{:else}
						{#each incidents as inc}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30 transition">
								<td class="px-5 py-3.5 whitespace-nowrap text-gray-500 dark:text-gray-400 font-mono">
									{new Date(inc.created_at).toLocaleString('en-GB', { day: '2-digit', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit', second: '2-digit', hour12: true })}
								</td>
								<td class="px-5 py-3.5 font-bold font-mono text-gray-900 dark:text-white">
									{inc.phone}
								</td>
								<td class="px-5 py-3.5">
									<span class="inline-flex items-center px-2 py-0.5 rounded-full text-[11px] font-semibold bg-amber-50 text-amber-800 border border-amber-200 dark:bg-amber-950/40 dark:text-amber-300 dark:border-amber-800">
										{inc.gateway_name}
									</span>
								</td>
								<td class="px-5 py-3.5">
									<span class="inline-flex items-center px-2 py-0.5 rounded-full text-[11px] font-semibold border {getEventBadge(inc.event)}">
										{getEventLabel(inc.event)}
									</span>
								</td>
								<td class="px-5 py-3.5 text-gray-600 dark:text-gray-400 max-w-md break-words" title={inc.error_message}>
									{inc.error_message}
								</td>
							</tr>
						{/each}
					{/if}
				</tbody>
			</table>
		</div>
	</div>
</div>
