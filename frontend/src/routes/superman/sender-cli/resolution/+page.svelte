<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface ResolutionOutput {
		sender_id: string;
		is_registered: boolean;
		sender_type: string;
		masking_category: string;
		is_masking: boolean;
		approval_status: string;
		gateway_name: string;
		gateway_driver: string;
		recipient_phone: string;
		operator_name: string;
		operator_base_cost: number;
		gateway_charged_rate: number;
		route_health: string;
		status: string;
	}

	interface SenderRow {
		id: number;
		sender_id: string;
		type: string;
		gateway_name?: string | null;
		users: { id: number; name: string }[];
		messages_count: number;
		approval_status: string;
	}

	let simSenderInput = $state('WHITEINFO');
	let simRecipientPhone = $state('01712345678');
	let resolution = $state<ResolutionOutput | null>(null);
	let resolving = $state(false);

	let senders = $state<SenderRow[]>([]);
	let search = $state('');
	let currentPage = $state(1);
	let perPage = $state(10);
	let total = $state(0);
	let totalPages = $state(1);
	let loadingSenders = $state(false);

	async function runResolution() {
		const sender = simSenderInput.trim();
		const phone = simRecipientPhone.trim();
		if (!sender || !phone) {
			resolution = null;
			return;
		}

		resolving = true;
		try {
			const res = await supermanApi.get<{ data: ResolutionOutput }>(
				`/api/v1/superman/sender-cli/resolution?sender=${encodeURIComponent(sender)}&phone=${encodeURIComponent(phone)}`
			);
			resolution = res.data;
		} catch (err: unknown) {
			console.error('Resolution failed:', err);
		} finally {
			resolving = false;
		}
	}

	async function loadSenders(newPage = currentPage) {
		loadingSenders = true;
		currentPage = newPage;
		try {
			let url = `/api/v1/superman/senders?page=${currentPage}&per_page=${perPage}`;
			if (search.trim()) url += `&search=${encodeURIComponent(search.trim())}`;
			const res = await supermanApi.get<{
				data: SenderRow[];
				meta: { page: number; per_page: number; total: number; total_pages: number };
			}>(url);
			senders = res.data || [];
			if (res.meta) {
				total = res.meta.total;
				totalPages = res.meta.total_pages;
			}
		} catch (err: unknown) {
			console.error('Failed to load senders for resolution matrix:', err);
		} finally {
			loadingSenders = false;
		}
	}

	function selectSenderForTest(senderId: string) {
		simSenderInput = senderId;
		void runResolution();
		window.scrollTo({ top: 0, behavior: 'smooth' });
	}

	onMount(() => {
		const paramSender = $page.url.searchParams.get('sender');
		if (paramSender) {
			simSenderInput = paramSender;
		}
		void runResolution();
		void loadSenders();
	});

	const exportColumns: ExportColumn<SenderRow>[] = [
		{ header: 'Sender ID', getValue: (s) => s.sender_id },
		{ header: 'Type', getValue: (s) => s.type },
		{ header: 'Assigned Gateway Channel', getValue: (s) => s.gateway_name || 'Unassigned' },
		{ header: 'Assigned Users', getValue: (s) => (s.users && s.users.length > 0 ? s.users.map((u) => u.name).join(', ') : 'None') },
		{ header: 'Historical Traffic', getValue: (s) => s.messages_count },
		{ header: 'Approval State', getValue: (s) => s.approval_status.toUpperCase() }
	];
</script>

<svelte:head>
	<title>Sender Resolution — SMS Whiteinfo Administrator</title>
</svelte:head>

<div class="space-y-6">
	<!-- Header Banner -->
	<div class="rounded-2xl bg-gradient-to-r from-blue-700 via-indigo-700 to-purple-700 p-6 text-white shadow-lg">
		<div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
			<div>
				<div class="inline-flex items-center gap-1.5 rounded-full bg-white/20 px-3 py-1 text-xs font-semibold text-white backdrop-blur-md">
					<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" d="M7.5 21 3 16.5m0 0L7.5 12M3 16.5h13.5m0-13.5L21 7.5m0 0L16.5 12M21 7.5H7.5" />
					</svg>
					Intelligent CLI Dispatch Engine
				</div>
				<h1 class="mt-2 text-2xl font-bold tracking-tight text-white">Sender & Route Resolution Engine</h1>
				<p class="mt-1 text-xs text-blue-100 max-w-2xl">
					Real-time carrier route diagnosis, gateway channel resolution, operator prefix matching, and dynamic tariff calculation for outgoing SMS traffic.
				</p>
			</div>
			<div class="flex items-center gap-3">
				<span class="inline-flex items-center gap-1.5 rounded-xl bg-white/10 px-4 py-2 text-xs font-semibold text-white backdrop-blur-md ring-1 ring-white/20">
					<span class="h-2 w-2 rounded-full bg-emerald-400 animate-pulse"></span>
					Resolution Engine Online
				</span>
			</div>
		</div>
	</div>

	<!-- Interactive Live Resolution Simulator -->
	<div class="overflow-hidden rounded-2xl bg-white p-6 shadow-xs ring-1 ring-gray-950/5 dark:bg-gray-900 dark:ring-white/10 space-y-6">
		<div class="flex items-center gap-3 border-b border-gray-100 pb-4 dark:border-gray-800">
			<div class="rounded-lg bg-blue-500/10 p-2 text-blue-600 dark:text-blue-400">
				<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z" />
				</svg>
			</div>
			<div>
				<h3 class="text-sm font-bold text-gray-950 dark:text-white">Live Sender & Prefix Resolution Diagnostic</h3>
				<p class="text-xs text-gray-500 dark:text-gray-400">Enter a sender ID and recipient phone to view the step-by-step resolution chain and calculated routing cost.</p>
			</div>
		</div>

		<div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
			<!-- Simulator Inputs -->
			<div class="space-y-4 rounded-xl bg-gray-50 p-5 dark:bg-gray-800/50">
				<div>
					<label class="block text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">Sender ID / CLI Input</label>
					<input
						type="text"
						bind:value={simSenderInput}
						oninput={() => void runResolution()}
						placeholder="e.g. WHITEINFO or 01712345678"
						class="mt-1 block w-full rounded-lg border border-gray-200 bg-white py-2 px-3 font-mono text-xs shadow-xs focus:border-blue-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
					/>
				</div>

				<div>
					<label class="block text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">Destination MSISDN</label>
					<input
						type="text"
						bind:value={simRecipientPhone}
						oninput={() => void runResolution()}
						placeholder="e.g. 01712345678 or +8801812345678"
						class="mt-1 block w-full rounded-lg border border-gray-200 bg-white py-2 px-3 font-mono text-xs shadow-xs focus:border-blue-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
					/>
				</div>

				<div class="pt-2 text-xs text-gray-500">
					<p class="font-medium text-gray-700 dark:text-gray-300">Quick Test Senders:</p>
					<div class="mt-2 flex flex-wrap gap-1.5">
						{#each senders.slice(0, 5) as s}
							<button
								type="button"
								onclick={() => selectSenderForTest(s.sender_id)}
								class="rounded bg-white px-2 py-1 text-xs font-mono font-medium text-gray-700 shadow-2xs hover:bg-blue-50 hover:text-blue-600 dark:bg-gray-700 dark:text-gray-300 dark:hover:bg-gray-600 transition"
							>
								{s.sender_id}
							</button>
						{/each}
					</div>
				</div>
			</div>

			<!-- Pipeline Visualization Output -->
			<div class="lg:col-span-2 space-y-3">
				{#if resolution}
					<div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
						<!-- Step 1: Sender Classification -->
						<div class="rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
							<div class="flex items-center gap-2 text-xs font-semibold text-gray-500">
								<span class="flex h-5 w-5 items-center justify-center rounded-full bg-blue-100 text-[10px] font-bold text-blue-600 dark:bg-blue-900/50 dark:text-blue-300">1</span>
								SENDER CLASSIFICATION
							</div>
							<div class="mt-2">
								<p class="font-mono text-base font-bold text-gray-950 dark:text-white">{resolution.sender_id}</p>
								<div class="mt-1 flex items-center gap-2">
									<span class="inline-flex rounded px-2 py-0.5 text-xs font-medium {resolution.is_masking ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-300' : 'bg-blue-50 text-blue-700 dark:bg-blue-950/40 dark:text-blue-300'}">
										{resolution.sender_type}
									</span>
									<span class="text-xs text-gray-400">
										{resolution.is_registered ? '✓ Registered' : '⚡ Dynamic'}
									</span>
								</div>
							</div>
						</div>

						<!-- Step 2: Operator Identification -->
						<div class="rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
							<div class="flex items-center gap-2 text-xs font-semibold text-gray-500">
								<span class="flex h-5 w-5 items-center justify-center rounded-full bg-indigo-100 text-[10px] font-bold text-indigo-600 dark:bg-indigo-900/50 dark:text-indigo-300">2</span>
								OPERATOR IDENTIFICATION
							</div>
							<div class="mt-2">
								<p class="text-base font-bold text-gray-950 dark:text-white">{resolution.operator_name}</p>
								<p class="mt-1 text-xs text-gray-500 font-mono">
									Target: <span class="text-gray-900 dark:text-white">{resolution.recipient_phone}</span> (Bangladesh Network)
								</p>
							</div>
						</div>

						<!-- Step 3: Gateway Channel -->
						<div class="rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
							<div class="flex items-center gap-2 text-xs font-semibold text-gray-500">
								<span class="flex h-5 w-5 items-center justify-center rounded-full bg-purple-100 text-[10px] font-bold text-purple-600 dark:bg-purple-900/50 dark:text-purple-300">3</span>
								GATEWAY DISPATCH CHANNEL
							</div>
							<div class="mt-2">
								<p class="text-base font-bold text-gray-950 dark:text-white">{resolution.gateway_name}</p>
								<p class="mt-1 text-xs text-purple-600 dark:text-purple-400 font-mono">
									Driver: {resolution.gateway_driver.toUpperCase()} • Primary Bound
								</p>
							</div>
						</div>

						<!-- Step 4: Resolved Tariff & Cost -->
						<div class="rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
							<div class="flex items-center gap-2 text-xs font-semibold text-gray-500">
								<span class="flex h-5 w-5 items-center justify-center rounded-full bg-amber-100 text-[10px] font-bold text-amber-600 dark:bg-amber-900/50 dark:text-amber-300">4</span>
								RESOLVED TARIFF & COST
							</div>
							<div class="mt-2 flex items-baseline gap-4">
								<div>
									<span class="text-xs text-gray-500">Gateway Cost:</span>
									<p class="font-mono text-base font-bold text-emerald-600 dark:text-emerald-400">
										৳{resolution.gateway_charged_rate.toFixed(4)}
									</p>
								</div>
								<div class="border-l border-gray-200 pl-4 dark:border-gray-700">
									<span class="text-xs text-gray-500">Operator Base:</span>
									<p class="font-mono text-sm font-semibold text-gray-700 dark:text-gray-300">
										৳{resolution.operator_base_cost.toFixed(4)}
									</p>
								</div>
							</div>
						</div>
					</div>

					<!-- Step 5: Route Verdict Banner -->
					<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between rounded-xl bg-emerald-50 p-4 text-xs text-emerald-800 ring-1 ring-emerald-600/20 dark:bg-emerald-950/30 dark:text-emerald-300 gap-2">
						<div class="flex items-center gap-2 font-medium">
							<svg class="h-5 w-5 text-emerald-600 shrink-0" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
							</svg>
							<span>Resolution Path Verified: <strong>{resolution.sender_id}</strong> &rarr; <strong>{resolution.operator_name}</strong> via <strong>{resolution.gateway_name}</strong></span>
						</div>
						<span class="font-mono font-bold">{resolution.route_health}</span>
					</div>
				{:else}
					<div class="flex h-full min-h-[180px] items-center justify-center rounded-xl border border-dashed border-gray-300 p-8 text-center text-gray-400 dark:border-gray-700">
						<p class="text-xs">Enter a Sender ID and MSISDN to resolve the dispatch route.</p>
					</div>
				{/if}
			</div>
		</div>
	</div>

	<!-- Active Sender Resolution Matrix Table -->
	<div class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-xs dark:border-gray-800 dark:bg-gray-900">
		<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between border-b border-gray-100 p-5 dark:border-gray-800 gap-3">
			<div>
				<h3 class="text-sm font-bold text-gray-950 dark:text-white">Active Sender Resolution Matrix</h3>
				<p class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">All registered senders and their bound gateway dispatch channels.</p>
			</div>
			<div class="flex items-center gap-2 w-full sm:w-auto">
				<div class="w-full sm:w-64">
					<input
						type="text"
						bind:value={search}
						onkeydown={(e) => e.key === 'Enter' && void loadSenders(1)}
						placeholder="Filter senders..."
						class="w-full rounded-lg border border-gray-200 bg-gray-50/70 py-1.5 px-3 text-xs shadow-xs focus:border-blue-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
					/>
				</div>
				<ExportDropdown
					title="Sender Resolution Matrix"
					columns={exportColumns}
					getData={() => senders}
					disabled={senders.length === 0}
				/>
			</div>
		</div>

		<div class="overflow-x-auto">
			<table class="w-full text-left text-xs text-gray-600 dark:text-gray-400">
				<thead class="bg-gray-50/70 text-[10px] font-semibold uppercase text-gray-700 dark:bg-gray-800/40 dark:text-gray-300">
					<tr>
						<th class="px-5 py-3.5">Sender ID</th>
						<th class="px-5 py-3.5">Type</th>
						<th class="px-5 py-3.5">Assigned Gateway Channel</th>
						<th class="px-5 py-3.5">Assigned Users</th>
						<th class="px-5 py-3.5">Historical Traffic</th>
						<th class="px-5 py-3.5">Approval State</th>
						<th class="px-5 py-3.5 text-right">Action</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
					{#if loadingSenders && senders.length === 0}
						<tr>
							<td colspan="7" class="px-5 py-10 text-center text-gray-400">Loading matrix...</td>
						</tr>
					{:else if senders.length === 0}
						<tr>
							<td colspan="7" class="px-5 py-10 text-center text-gray-400">No senders found.</td>
						</tr>
					{:else}
						{#each senders as s}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/40 transition">
								<td class="px-5 py-3.5 font-mono font-bold text-gray-950 dark:text-white">
									{s.sender_id}
								</td>
								<td class="px-5 py-3.5">
									<span class="inline-flex rounded px-2 py-0.5 text-xs font-medium {s.type !== 'numeric' ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-300' : 'bg-blue-50 text-blue-700 dark:bg-blue-950/40 dark:text-blue-300'}">
										{s.type === 'numeric' ? 'Non-Masking' : (s.type === 'shortcode' ? 'Shortcode' : 'Masking')}
									</span>
								</td>
								<td class="px-5 py-3.5 font-medium text-gray-900 dark:text-white">
									<span class="inline-flex items-center gap-1.5">
										<span class="h-2 w-2 rounded-full bg-emerald-500"></span>
										{s.gateway_name || 'Default Gateway'}
									</span>
								</td>
								<td class="px-5 py-3.5 text-xs">
									{#if s.users && s.users.length > 0}
										{s.users.map((u) => u.name).slice(0, 2).join(', ')}{s.users.length > 2 ? ` +${s.users.length - 2}` : ''}
									{:else}
										<span class="text-gray-400">Global / All Users</span>
									{/if}
								</td>
								<td class="px-5 py-3.5 font-mono text-gray-700 dark:text-gray-300">
									{s.messages_count.toLocaleString()} msgs
								</td>
								<td class="px-5 py-3.5">
									<span class="inline-flex items-center rounded-md bg-emerald-50 px-2 py-0.5 text-xs font-medium text-emerald-700 ring-1 ring-inset ring-emerald-600/20 dark:bg-emerald-950/40 dark:text-emerald-300">
										{s.approval_status || 'approved'}
									</span>
								</td>
								<td class="px-5 py-3.5 text-right">
									<button
										type="button"
										onclick={() => selectSenderForTest(s.sender_id)}
										class="inline-flex items-center gap-1 text-xs font-semibold text-blue-600 hover:text-blue-700 dark:text-blue-400"
									>
										Test Route &rarr;
									</button>
								</td>
							</tr>
						{/each}
					{/if}
				</tbody>
			</table>
		</div>

		<!-- Pagination Footer -->
		{#if totalPages > 1}
			<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between border-t border-gray-100 px-5 py-3 dark:border-gray-800 gap-3">
				<div class="text-xs text-gray-500 dark:text-gray-400">
					Showing {(currentPage - 1) * perPage + 1} to {Math.min(currentPage * perPage, total)} of {total} senders
				</div>
				<div class="flex items-center gap-1.5">
					<button
						type="button"
						disabled={currentPage <= 1}
						onclick={() => void loadSenders(currentPage - 1)}
						class="rounded-lg border border-gray-200 px-3 py-1 text-xs font-semibold text-gray-700 disabled:opacity-40 hover:bg-gray-50 dark:border-gray-700 dark:text-gray-300"
					>
						Previous
					</button>
					<span class="px-2 text-xs text-gray-600 dark:text-gray-400">Page {currentPage} of {totalPages}</span>
					<button
						type="button"
						disabled={currentPage >= totalPages}
						onclick={() => void loadSenders(currentPage + 1)}
						class="rounded-lg border border-gray-200 px-3 py-1 text-xs font-semibold text-gray-700 disabled:opacity-40 hover:bg-gray-50 dark:border-gray-700 dark:text-gray-300"
					>
						Next
					</button>
				</div>
			</div>
		{/if}
	</div>
</div>

