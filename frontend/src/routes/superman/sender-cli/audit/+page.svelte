<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface AuditSummary {
		total_senders: number;
		total_messages: number;
		delivered_messages: number;
		total_revenue: number;
		success_rate: number;
	}

	interface LedgerItem {
		id: number;
		sender_id: string;
		type: string;
		masking_category: string;
		gateway_name?: string | null;
		is_enabled: boolean;
		approval_status: string;
		total_messages: number;
		delivered_messages: number;
		failed_messages: number;
		success_rate: number;
		total_spend: number;
		users: { id: number; name: string }[];
	}

	interface DispatchItem {
		id: number;
		created_at?: string;
		sender_id?: string;
		recipient: string;
		customer_name?: string;
		gateway_name?: string;
		status: string;
		cost: number;
	}

	interface GatewayRef {
		id: number;
		name: string;
	}

	let activeView = $state<'table' | 'activity'>('table');
	let search = $state('');
	let typeFilter = $state('');
	let gatewayFilter = $state('');

	let summary = $state<AuditSummary>({
		total_senders: 0,
		total_messages: 0,
		delivered_messages: 0,
		total_revenue: 0,
		success_rate: 100
	});

	let gateways = $state<GatewayRef[]>([]);
	let ledger = $state<LedgerItem[]>([]);
	let dispatches = $state<DispatchItem[]>([]);

	let loading = $state(false);
	let page = $state(1);
	let perPage = $state(15);
	let total = $state(0);
	let totalPages = $state(1);

	async function loadSummary() {
		try {
			const res = await supermanApi.get<{ data: AuditSummary }>('/api/v1/superman/sender-cli/summary');
			if (res.data) summary = res.data;
		} catch (err: unknown) {
			console.error('Failed to load audit summary:', err);
		}
	}

	async function loadOptions() {
		try {
			const res = await supermanApi.get<{ data: { gateways: GatewayRef[] } }>('/api/v1/superman/sender-cli/options');
			if (res.data?.gateways) gateways = res.data.gateways;
		} catch (err: unknown) {
			console.error('Failed to load gateways:', err);
		}
	}

	async function loadLedger(newPage = page) {
		loading = true;
		page = newPage;
		try {
			let url = `/api/v1/superman/sender-cli/ledger?page=${page}&per_page=${perPage}`;
			if (search.trim()) url += `&search=${encodeURIComponent(search.trim())}`;
			if (typeFilter) url += `&type=${encodeURIComponent(typeFilter)}`;
			if (gatewayFilter) url += `&gateway_id=${encodeURIComponent(gatewayFilter)}`;

			const res = await supermanApi.get<{
				data: LedgerItem[];
				meta: { page: number; per_page: number; total: number; total_pages: number };
			}>(url);

			ledger = res.data || [];
			if (res.meta) {
				total = res.meta.total;
				totalPages = res.meta.total_pages;
			}
		} catch (err: unknown) {
			console.error('Failed to load audit ledger:', err);
		} finally {
			loading = false;
		}
	}

	async function loadDispatches() {
		loading = true;
		try {
			let url = '/api/v1/superman/sender-cli/dispatches';
			if (search.trim()) url += `?search=${encodeURIComponent(search.trim())}`;

			const res = await supermanApi.get<{ data: DispatchItem[] }>(url);
			dispatches = res.data || [];
		} catch (err: unknown) {
			console.error('Failed to load dispatches:', err);
		} finally {
			loading = false;
		}
	}

	function handleViewSwitch(view: 'table' | 'activity') {
		activeView = view;
		if (view === 'table') {
			void loadLedger(1);
		} else {
			void loadDispatches();
		}
	}

	onMount(() => {
		void loadSummary();
		void loadOptions();
		void loadLedger();
	});

	const exportLedgerColumns: ExportColumn<LedgerItem>[] = [
		{ header: 'Sender / CLI', getValue: (l) => l.sender_id },
		{ header: 'Category', getValue: (l) => l.masking_category || l.type },
		{ header: 'Gateway', getValue: (l) => l.gateway_name || 'Unassigned' },
		{ header: 'Assigned Customers', getValue: (l) => (l.users && l.users.length > 0 ? l.users.map((u) => u.name).join(', ') : 'None') },
		{ header: 'Total SMS Sent', getValue: (l) => l.total_messages },
		{ header: 'Delivered', getValue: (l) => l.delivered_messages },
		{ header: 'Failed', getValue: (l) => l.failed_messages },
		{ header: 'Success Rate', getValue: (l) => `${l.success_rate.toFixed(1)}%` },
		{ header: 'Total Volume (BDT)', getValue: (l) => `BDT ${l.total_spend.toFixed(2)}` },
		{ header: 'Operating State', getValue: (l) => (l.is_enabled ? 'Active' : 'Disabled') }
	];

	const exportDispatchColumns: ExportColumn<DispatchItem>[] = [
		{ header: 'Timestamp', getValue: (d) => (d.created_at ? new Date(d.created_at).toLocaleString() : '—') },
		{ header: 'Sender ID', getValue: (d) => d.sender_id || '—' },
		{ header: 'Recipient', getValue: (d) => d.recipient },
		{ header: 'Customer', getValue: (d) => d.customer_name || '—' },
		{ header: 'Gateway', getValue: (d) => d.gateway_name || 'Auto' },
		{ header: 'Status', getValue: (d) => d.status.toUpperCase() },
		{ header: 'Cost (BDT)', getValue: (d) => `BDT ${Number(d.cost).toFixed(4)}` }
	];
</script>

<svelte:head>
	<title>Sender Audit — SMS Whiteinfo Administrator</title>
</svelte:head>

<div class="space-y-6">
	<!-- Header KPIs Grid -->
	<div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
		<!-- Total Senders -->
		<div class="flex flex-col gap-y-2 rounded-xl border border-gray-200 bg-white p-5 shadow-xs dark:border-gray-800 dark:bg-gray-900">
			<div class="flex items-center gap-3">
				<div class="rounded-lg bg-blue-500/10 p-2 text-blue-600 dark:text-blue-400">
					<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" d="M15 9h3.75M15 12h3.75M15 15h3.75M4.5 19.5h15a2.25 2.25 0 0 0 2.25-2.25V6.75A2.25 2.25 0 0 0 19.5 4.5h-15a2.25 2.25 0 0 0-2.25 2.25v10.5A2.25 2.25 0 0 0 4.5 19.5Zm6-10.125a1.875 1.875 0 1 1-3.75 0 1.875 1.875 0 0 1 3.75 0Zm1.294 6.336a6.721 6.721 0 0 1-3.17.789 6.721 6.721 0 0 1-3.168-.789 3.376 3.376 0 0 1 6.338 0Z" />
					</svg>
				</div>
				<div>
					<p class="text-xs font-medium text-gray-500 dark:text-gray-400">Total Senders</p>
					<p class="text-2xl font-bold text-gray-950 dark:text-white">{summary.total_senders.toLocaleString()}</p>
				</div>
			</div>
			<p class="text-[11px] text-gray-400">Masking & Non-masking active senders</p>
		</div>

		<!-- Total Dispatches -->
		<div class="flex flex-col gap-y-2 rounded-xl border border-gray-200 bg-white p-5 shadow-xs dark:border-gray-800 dark:bg-gray-900">
			<div class="flex items-center gap-3">
				<div class="rounded-lg bg-purple-500/10 p-2 text-purple-600 dark:text-purple-400">
					<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" d="M6 12 3.269 3.125A59.769 59.769 0 0 1 21.485 12 59.768 59.768 0 0 1 3.27 20.875L5.999 12Zm0 0h7.5" />
					</svg>
				</div>
				<div>
					<p class="text-xs font-medium text-gray-500 dark:text-gray-400">Total Dispatches</p>
					<p class="text-2xl font-bold text-gray-950 dark:text-white">{summary.total_messages.toLocaleString()}</p>
				</div>
			</div>
			<p class="text-[11px] text-gray-400">Cumulative SMS routed by senders</p>
		</div>

		<!-- Delivery Success Rate -->
		<div class="flex flex-col gap-y-2 rounded-xl border border-gray-200 bg-white p-5 shadow-xs dark:border-gray-800 dark:bg-gray-900">
			<div class="flex items-center gap-3">
				<div class="rounded-lg bg-emerald-500/10 p-2 text-emerald-600 dark:text-emerald-400">
					<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
					</svg>
				</div>
				<div>
					<p class="text-xs font-medium text-gray-500 dark:text-gray-400">Delivery Success Rate</p>
					<p class="text-2xl font-bold text-emerald-600 dark:text-emerald-400">{summary.success_rate}%</p>
				</div>
			</div>
			<p class="text-[11px] text-gray-400">{summary.delivered_messages.toLocaleString()} delivered successfully</p>
		</div>

		<!-- Total Revenue Generated -->
		<div class="flex flex-col gap-y-2 rounded-xl border border-gray-200 bg-white p-5 shadow-xs dark:border-gray-800 dark:bg-gray-900">
			<div class="flex items-center gap-3">
				<div class="rounded-lg bg-amber-500/10 p-2 text-amber-600 dark:text-amber-400">
					<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" d="M2.25 18.75a60.07 60.07 0 0 1 15.797 2.101c.727.198 1.453-.342 1.453-1.096V18.75M3.75 4.5v.75A.75.75 0 0 1 3 6H2.25m0 0v10.5m0 0H3a.75.75 0 0 1 .75.75v.75m0 0a60.07 60.07 0 0 1 15.797 2.101c.727.198 1.453-.342 1.453-1.096V18.75M3.75 4.5v10.5m0 0a60.07 60.07 0 0 0 15.797 2.101c.727.198 1.453-.342 1.453-1.096V4.5M3.75 4.5H20.25m0 0a.75.75 0 0 1 .75.75v.75m0 0a60.07 60.07 0 0 1-15.797 2.101" />
					</svg>
				</div>
				<div>
					<p class="text-xs font-medium text-gray-500 dark:text-gray-400">Total Volume Generated</p>
					<p class="text-2xl font-bold text-amber-600 dark:text-amber-400">৳{summary.total_revenue.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}</p>
				</div>
			</div>
			<p class="text-[11px] text-gray-400">Sender generated billing volume</p>
		</div>
	</div>

	<!-- Controls & View Switcher -->
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<div class="inline-flex rounded-lg bg-gray-100 p-1 dark:bg-gray-800">
			<button
				type="button"
				onclick={() => handleViewSwitch('table')}
				class="inline-flex items-center gap-2 rounded-md px-4 py-2 text-xs font-medium transition-all {activeView === 'table' ? 'bg-white text-gray-950 shadow-2xs dark:bg-gray-900 dark:text-white' : 'text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white'}"
			>
				<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" d="M3.75 6A2.25 2.25 0 0 1 6 3.75h2.25A2.25 2.25 0 0 1 10.5 6v2.25a2.25 2.25 0 0 1-2.25 2.25H6a2.25 2.25 0 0 1-2.25-2.25V6ZM3.75 15.75A2.25 2.25 0 0 1 6 13.5h2.25a2.25 2.25 0 0 1 2.25 2.25V18a2.25 2.25 0 0 1-2.25 2.25H6A2.25 2.25 0 0 1 3.75 18v-2.25ZM13.5 6a2.25 2.25 0 0 1 2.25-2.25H18A2.25 2.25 0 0 1 20.25 6v2.25A2.25 2.25 0 0 1 18 10.5h-2.25a2.25 2.25 0 0 1-2.25-2.25V6ZM13.5 15.75a2.25 2.25 0 0 1 2.25-2.25H18a2.25 2.25 0 0 1 2.25 2.25V18A2.25 2.25 0 0 1 18 20.25h-2.25A2.25 2.25 0 0 1 13.5 18v-2.25Z" />
				</svg>
				Sender Audit Ledger
			</button>
			<button
				type="button"
				onclick={() => handleViewSwitch('activity')}
				class="inline-flex items-center gap-2 rounded-md px-4 py-2 text-xs font-medium transition-all {activeView === 'activity' ? 'bg-white text-gray-950 shadow-2xs dark:bg-gray-900 dark:text-white' : 'text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white'}"
			>
				<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
				</svg>
				Live Dispatch Stream (Recent 50)
			</button>
		</div>

		<div class="flex flex-wrap items-center gap-3">
			<div class="relative min-w-[200px]">
				<input
					type="text"
					bind:value={search}
					onkeydown={(e) => e.key === 'Enter' && (activeView === 'table' ? void loadLedger(1) : void loadDispatches())}
					placeholder="Search sender, user or gateway..."
					class="w-full rounded-lg border border-gray-200 bg-white py-1.5 pl-8 pr-3 text-xs shadow-xs focus:border-blue-500 dark:border-gray-700 dark:bg-gray-900 dark:text-white"
				/>
				<svg class="absolute left-2.5 top-2 h-4 w-4 text-gray-400" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z" />
				</svg>
			</div>

			{#if activeView === 'table'}
				<select
					bind:value={typeFilter}
					onchange={() => void loadLedger(1)}
					class="rounded-lg border border-gray-200 bg-white py-1.5 pl-3 pr-8 text-xs shadow-xs focus:border-blue-500 dark:border-gray-700 dark:bg-gray-900 dark:text-white"
				>
					<option value="">All Categories</option>
					<option value="masking">Masking Only</option>
					<option value="non_masking">Non-Masking (Numeric)</option>
				</select>

				<select
					bind:value={gatewayFilter}
					onchange={() => void loadLedger(1)}
					class="rounded-lg border border-gray-200 bg-white py-1.5 pl-3 pr-8 text-xs shadow-xs focus:border-blue-500 dark:border-gray-700 dark:bg-gray-900 dark:text-white"
				>
					<option value="">All Gateways</option>
					{#each gateways as gw}
						<option value={gw.id}>{gw.name}</option>
					{/each}
				</select>
			{/if}

			{#if activeView === 'table'}
				<ExportDropdown
					title="Sender Audit Ledger"
					columns={exportLedgerColumns}
					getData={() => ledger}
					disabled={ledger.length === 0}
				/>
			{:else}
				<ExportDropdown
					title="Live Dispatch Stream"
					columns={exportDispatchColumns}
					getData={() => dispatches}
					disabled={dispatches.length === 0}
				/>
			{/if}
		</div>
	</div>

	<!-- VIEW 1: SENDER AUDIT LEDGER TABLE -->
	{#if activeView === 'table'}
		<div class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-xs dark:border-gray-800 dark:bg-gray-900">
			<div class="overflow-x-auto">
				<table class="w-full min-w-[980px] text-left text-xs text-gray-600 dark:text-gray-400">
					<thead class="bg-gray-50/70 text-[10px] font-semibold uppercase text-gray-700 dark:bg-gray-800/40 dark:text-gray-300">
						<tr>
							<th class="px-5 py-3.5">Sender / CLI</th>
							<th class="px-5 py-3.5">Category</th>
							<th class="px-5 py-3.5">Gateway</th>
							<th class="px-5 py-3.5">Assigned Customers</th>
							<th class="px-5 py-3.5">Total SMS Sent</th>
							<th class="px-5 py-3.5">Delivered</th>
							<th class="px-5 py-3.5">Failed</th>
							<th class="px-5 py-3.5">Success Rate</th>
							<th class="px-5 py-3.5">Total Volume</th>
							<th class="px-5 py-3.5">Status</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
						{#if loading && ledger.length === 0}
							<tr>
								<td colspan="10" class="px-5 py-10 text-center text-gray-400">Loading audit ledger...</td>
							</tr>
						{:else if ledger.length === 0}
							<tr>
								<td colspan="10" class="px-5 py-10 text-center text-gray-400">No senders found.</td>
							</tr>
						{:else}
							{#each ledger as s}
								<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/40 transition">
									<td class="px-5 py-3.5 font-mono font-bold text-gray-950 dark:text-white">
										<span class="inline-flex items-center gap-1.5 rounded-lg bg-gray-100 px-2.5 py-1 font-mono text-xs font-semibold text-gray-900 dark:bg-gray-800 dark:text-white">
											{s.sender_id}
										</span>
									</td>
									<td class="px-5 py-3.5">
										<span class="inline-flex rounded px-2 py-0.5 text-xs font-medium {s.type !== 'numeric' ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-300' : 'bg-blue-50 text-blue-700 dark:bg-blue-950/40 dark:text-blue-300'}">
											{s.type === 'numeric' ? 'Non-Masking' : (s.type === 'shortcode' ? 'Shortcode' : 'Masking')}
										</span>
									</td>
									<td class="px-5 py-3.5 font-medium text-gray-700 dark:text-gray-300">
										{s.gateway_name || 'Default Gateway'}
									</td>
									<td class="px-5 py-3.5 text-xs">
										<div class="flex flex-wrap gap-1">
											{#if s.users && s.users.length > 0}
												{#each s.users.slice(0, 2) as u}
													<span class="inline-flex rounded bg-gray-100 px-1.5 py-0.5 text-gray-700 dark:bg-gray-800 dark:text-gray-300">
														{u.name}
													</span>
												{/each}
												{#if s.users.length > 2}
													<span class="text-gray-400">+{s.users.length - 2}</span>
												{/if}
											{:else}
												<span class="text-gray-400">Global</span>
											{/if}
										</div>
									</td>
									<td class="px-5 py-3.5 font-mono font-semibold text-gray-950 dark:text-white">
										{s.total_messages.toLocaleString()}
									</td>
									<td class="px-5 py-3.5 font-mono text-emerald-600 dark:text-emerald-400">
										{s.delivered_messages.toLocaleString()}
									</td>
									<td class="px-5 py-3.5 font-mono text-rose-600 dark:text-rose-400">
										{s.failed_messages.toLocaleString()}
									</td>
									<td class="px-5 py-3.5">
										<div class="flex items-center gap-2">
											<div class="h-2 w-16 overflow-hidden rounded-full bg-gray-200 dark:bg-gray-700">
												<div class="h-full bg-emerald-500 rounded-full" style="width: {s.success_rate}%"></div>
											</div>
											<span class="font-mono text-xs font-semibold {s.success_rate >= 90 ? 'text-emerald-600 dark:text-emerald-400' : (s.success_rate >= 70 ? 'text-amber-600 dark:text-amber-400' : 'text-rose-600 dark:text-rose-400')}">
												{s.success_rate.toFixed(1)}%
											</span>
										</div>
									</td>
									<td class="px-5 py-3.5 font-mono font-medium text-gray-950 dark:text-white">
										৳{s.total_spend.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
									</td>
									<td class="px-5 py-3.5">
										<span class="inline-flex items-center gap-1.5 text-xs {s.is_enabled ? 'text-emerald-600 dark:text-emerald-400' : 'text-gray-400'}">
											<span class="h-2 w-2 rounded-full {s.is_enabled ? 'bg-emerald-500' : 'bg-gray-400'}"></span>
											{s.is_enabled ? 'Active' : 'Inactive'}
										</span>
									</td>
								</tr>
							{/each}
						{/if}
					</tbody>
				</table>
			</div>

			<!-- Pagination footer -->
			{#if totalPages > 1}
				<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between border-t border-gray-100 px-5 py-3 dark:border-gray-800 gap-3">
					<div class="text-xs text-gray-500 dark:text-gray-400">
						Showing {(page - 1) * perPage + 1} to {Math.min(page * perPage, total)} of {total} senders
					</div>
					<div class="flex items-center gap-1.5">
						<button
							type="button"
							disabled={page <= 1}
							onclick={() => void loadLedger(page - 1)}
							class="rounded-lg border border-gray-200 px-3 py-1 text-xs font-semibold text-gray-700 disabled:opacity-40 hover:bg-gray-50 dark:border-gray-700 dark:text-gray-300"
						>
							Previous
						</button>
						<span class="px-2 text-xs text-gray-600 dark:text-gray-400">Page {page} of {totalPages}</span>
						<button
							type="button"
							disabled={page >= totalPages}
							onclick={() => void loadLedger(page + 1)}
							class="rounded-lg border border-gray-200 px-3 py-1 text-xs font-semibold text-gray-700 disabled:opacity-40 hover:bg-gray-50 dark:border-gray-700 dark:text-gray-300"
						>
							Next
						</button>
					</div>
				</div>
			{/if}
		</div>
	{/if}

	<!-- VIEW 2: LIVE DISPATCH STREAM -->
	{#if activeView === 'activity'}
		<div class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-xs dark:border-gray-800 dark:bg-gray-900">
			<div class="border-b border-gray-100 p-4 dark:border-gray-800 flex items-center justify-between">
				<p class="text-xs font-semibold uppercase tracking-wider text-gray-500">Recent 50 SMS Dispatches by Senders</p>
				<button
					type="button"
					onclick={() => void loadDispatches()}
					class="inline-flex items-center gap-1 text-xs text-emerald-600 hover:text-emerald-700 font-semibold"
				>
					<svg class="h-3.5 w-3.5 {loading ? 'animate-spin' : ''}" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0 3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182m0-4.991v4.99" />
					</svg>
					Refresh Stream
				</button>
			</div>

			<div class="overflow-x-auto">
				<table class="w-full min-w-[800px] text-left text-xs text-gray-600 dark:text-gray-400">
					<thead class="bg-gray-50/70 text-[10px] font-semibold uppercase text-gray-700 dark:bg-gray-800/40 dark:text-gray-300">
						<tr>
							<th class="px-5 py-3">Timestamp</th>
							<th class="px-5 py-3">Sender ID</th>
							<th class="px-5 py-3">Recipient</th>
							<th class="px-5 py-3">User / Customer</th>
							<th class="px-5 py-3">Gateway</th>
							<th class="px-5 py-3">Status</th>
							<th class="px-5 py-3">Cost</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100 font-mono text-xs dark:divide-gray-800">
						{#if loading && dispatches.length === 0}
							<tr>
								<td colspan="7" class="px-5 py-10 text-center font-sans text-gray-400">Loading stream...</td>
							</tr>
						{:else if dispatches.length === 0}
							<tr>
								<td colspan="7" class="px-5 py-10 text-center font-sans text-gray-400">No recent dispatches logged.</td>
							</tr>
						{:else}
							{#each dispatches as msg}
								<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/40 transition">
									<td class="px-5 py-3 text-gray-400">{msg.created_at || '—'}</td>
									<td class="px-5 py-3 font-bold text-gray-950 dark:text-white">{msg.sender_id || '—'}</td>
									<td class="px-5 py-3 text-gray-700 dark:text-gray-300">{msg.recipient}</td>
									<td class="px-5 py-3 font-sans">{msg.customer_name || '—'}</td>
									<td class="px-5 py-3 font-sans">{msg.gateway_name || '—'}</td>
									<td class="px-5 py-3">
										<span class="inline-flex rounded px-2 py-0.5 font-sans text-[11px] font-semibold {msg.status === 'delivered' ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-300' : (msg.status === 'sent' ? 'bg-blue-50 text-blue-700 dark:bg-blue-950/40 dark:text-blue-300' : 'bg-rose-50 text-rose-700 dark:bg-rose-950/40 dark:text-rose-300')}">
											{msg.status.toUpperCase()}
										</span>
									</td>
									<td class="px-5 py-3 font-bold text-gray-950 dark:text-white">৳{msg.cost.toFixed(4)}</td>
								</tr>
							{/each}
						{/if}
					</tbody>
				</table>
			</div>
		</div>
	{/if}
</div>

