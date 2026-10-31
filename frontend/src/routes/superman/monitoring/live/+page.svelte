<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface StreamItem {
		id: number;
		customer_name: string;
		customer_email: string;
		recipient: string;
		body: string;
		status: string;
		source: string;
		source_ip: string;
		gateway: string;
		gateway_name: string;
		sender_id: string;
		cost: number;
		parts: number;
		failed_reason: string;
		created_at: string;
		updated_at: string;
	}

	interface LiveData {
		total_messages: number;
		delivered: number;
		pending: number;
		failed: number;
		active_now: number;
		msgs_per_min: number;
		active_batches: number;
		live_success_rate: number;
		active_gateways: number;
		live_cost: number;
		stream: StreamItem[];
	}

	let data = $state<LiveData | null>(null);
	let loading = $state(true);
	let interval: any;

	async function fetchLive() {
		try {
			const res = await supermanApi.get<{ data: LiveData }>('/api/v1/superman/monitoring/live');
			data = res.data ?? null;
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		fetchLive();
		interval = setInterval(fetchLive, 5000);
	});

	onDestroy(() => {
		if (interval) clearInterval(interval);
	});

	const exportColumns: ExportColumn<StreamItem>[] = [
		{ header: 'Message ID', getValue: (m) => `#${m.id}` },
		{ header: 'Customer', getValue: (m) => m.customer_name },
		{ header: 'Recipient', getValue: (m) => m.recipient },
		{ header: 'CLI / Sender', getValue: (m) => m.sender_id },
		{ header: 'Carrier Gateway', getValue: (m) => m.gateway || m.gateway_name },
		{ header: 'Status', getValue: (m) => m.status.toUpperCase() },
		{ header: 'Segments', getValue: (m) => m.parts },
		{ header: 'Carrier Cost (BDT)', getValue: (m) => m.cost.toFixed(4) },
		{ header: 'Message Text', getValue: (m) => m.body || '—' },
		{ header: 'Timestamp', getValue: (m) => new Date(m.created_at || m.updated_at).toLocaleString() }
	];

	function getStatusBadge(status: string) {
		const s = status.toLowerCase();
		if (s === 'delivered') {
			return 'bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-950/40 dark:text-emerald-400 dark:border-emerald-800';
		}
		if (s === 'sent') {
			return 'bg-blue-50 text-blue-700 border-blue-200 dark:bg-blue-950/40 dark:text-blue-400 dark:border-blue-800';
		}
		if (s === 'pending' || s === 'sending') {
			return 'bg-amber-50 text-amber-700 border-amber-200 dark:bg-amber-950/40 dark:text-amber-400 dark:border-amber-800';
		}
		return 'bg-rose-50 text-rose-700 border-rose-200 dark:bg-rose-950/40 dark:text-rose-400 dark:border-rose-800';
	}
</script>

<div class="space-y-6">
	<!-- Page Header -->
	<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
		<div>
			<div class="flex items-center gap-2">
				<span class="h-2.5 w-2.5 rounded-full bg-emerald-500 animate-ping"></span>
				<h1 class="text-xl font-bold text-gray-900 dark:text-white">Network Operations Center (NOC)</h1>
			</div>
			<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
				Live Operations & Gateway Traffic Telemetry: real-time live message dispatch stream, queue drain, and carrier latency.
			</p>
		</div>
		<div class="flex items-center gap-2">
			<ExportDropdown
				title="NOC Live Dispatches"
				columns={exportColumns}
				getData={() => data?.stream || []}
				disabled={!data || data.stream.length === 0}
			/>
			<span class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200 dark:bg-emerald-950/40 dark:text-emerald-300 dark:border-emerald-800">
				<span class="h-1.5 w-1.5 rounded-full bg-emerald-500"></span>
				Auto-refreshing (5s)
			</span>
		</div>
	</div>

	{#if loading && !data}
		<div class="p-12 text-center text-xs text-gray-400">Connecting to NOC telemetry stream...</div>
	{:else if data}
		<!-- Live Queue Stats Overview Widget -->
		<div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-3">
			<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-4 shadow-xs">
				<div class="flex items-center justify-between text-xs text-gray-500 dark:text-gray-400">
					<span>⚡ In Queue Now</span>
				</div>
				<p class="text-2xl font-bold text-gray-900 dark:text-white mt-1">{(data.active_now || 0).toLocaleString()}</p>
				<span class="text-[11px] text-amber-600 mt-0.5 block">Pending & sending</span>
			</div>

			<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-4 shadow-xs">
				<div class="flex items-center justify-between text-xs text-gray-500 dark:text-gray-400">
					<span>📊 Msgs/Min</span>
				</div>
				<p class="text-2xl font-bold text-blue-600 dark:text-blue-400 mt-1">{data.msgs_per_min || 0}</p>
				<span class="text-[11px] text-gray-400 mt-0.5 block">5-min average rate</span>
			</div>

			<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-4 shadow-xs">
				<div class="flex items-center justify-between text-xs text-gray-500 dark:text-gray-400">
					<span>🔄 Active Batches</span>
				</div>
				<p class="text-2xl font-bold text-gray-900 dark:text-white mt-1">{data.active_batches || 0}</p>
				<span class="text-[11px] text-gray-400 mt-0.5 block">Currently processing</span>
			</div>

			<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-4 shadow-xs">
				<div class="flex items-center justify-between text-xs text-gray-500 dark:text-gray-400">
					<span>✅ Live Success</span>
				</div>
				<p class="text-2xl font-bold text-emerald-600 dark:text-emerald-400 mt-1">{data.live_success_rate || 0}%</p>
				<span class="text-[11px] text-gray-400 mt-0.5 block">5-min delivery window</span>
			</div>

			<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-4 shadow-xs col-span-2 sm:col-span-1">
				<div class="flex items-center justify-between text-xs text-gray-500 dark:text-gray-400">
					<span>🏗️ Active Gateways</span>
				</div>
				<p class="text-2xl font-bold text-gray-900 dark:text-white mt-1">{data.active_gateways || 0}</p>
				<span class="text-[11px] text-emerald-600 mt-0.5 block">{data.live_cost.toFixed(2)} BDT burned</span>
			</div>
		</div>

		<!-- Live Stream Table -->
		<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl shadow-xs overflow-hidden">
			<div class="px-5 py-4 border-b border-gray-200 dark:border-gray-800 flex items-center justify-between">
				<div>
					<h2 class="text-sm font-bold text-gray-900 dark:text-white">Live Dispatched Messages Stream</h2>
					<p class="text-xs text-gray-400 mt-0.5">Most recent carrier transmissions and deliverability status</p>
				</div>
				<span class="text-xs font-mono text-gray-400">Showing {data.stream.length} recent messages</span>
			</div>
			<div class="overflow-x-auto">
				<table class="w-full text-left text-xs">
					<thead class="bg-gray-50/75 dark:bg-gray-800/50 text-gray-500 uppercase tracking-wider font-semibold border-b border-gray-200 dark:border-gray-800">
						<tr>
							<th class="px-5 py-3">Message ID</th>
							<th class="px-5 py-3">Customer</th>
							<th class="px-5 py-3">Recipient</th>
							<th class="px-5 py-3">Sender CLI</th>
							<th class="px-5 py-3">Gateway</th>
							<th class="px-5 py-3">Body Preview</th>
							<th class="px-5 py-3">Status</th>
							<th class="px-5 py-3 text-right">Cost</th>
							<th class="px-5 py-3 whitespace-nowrap">Timestamp</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100 dark:divide-gray-800 text-gray-700 dark:text-gray-300">
						{#if data.stream.length === 0}
							<tr>
								<td colspan="9" class="px-5 py-10 text-center text-gray-400">No recent messages in stream.</td>
							</tr>
						{:else}
							{#each data.stream as m}
								<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30 transition">
									<td class="px-5 py-3.5 font-mono font-bold text-gray-900 dark:text-white">#{m.id}</td>
									<td class="px-5 py-3.5 font-medium text-gray-800 dark:text-gray-200">{m.customer_name}</td>
									<td class="px-5 py-3.5 font-mono font-semibold">{m.recipient}</td>
									<td class="px-5 py-3.5 font-mono text-gray-600 dark:text-gray-400">{m.sender_id}</td>
									<td class="px-5 py-3.5">
										<span class="inline-flex items-center px-2 py-0.5 rounded text-[11px] font-semibold bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300">
											{m.gateway || m.gateway_name}
										</span>
									</td>
									<td class="px-5 py-3.5 max-w-xs truncate text-gray-600 dark:text-gray-400" title={m.body}>{m.body || '—'}</td>
									<td class="px-5 py-3.5">
										<span class="inline-flex items-center px-2 py-0.5 rounded-full text-[11px] font-semibold border {getStatusBadge(m.status)}">
											{m.status.toUpperCase()}
										</span>
									</td>
									<td class="px-5 py-3.5 text-right font-mono font-semibold text-gray-700 dark:text-gray-300">
										{m.cost.toFixed(2)} ৳
									</td>
									<td class="px-5 py-3.5 whitespace-nowrap text-gray-400 font-mono">
										{new Date(m.created_at || m.updated_at).toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit', second: '2-digit', hour12: true })}
									</td>
								</tr>
							{/each}
						{/if}
					</tbody>
				</table>
			</div>
		</div>
	{/if}
</div>
