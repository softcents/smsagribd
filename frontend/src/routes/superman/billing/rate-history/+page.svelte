<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface RateHistoryRecord {
		id: number;
		plan_name: string;
		operator_name: string;
		event: string;
		old_masking_price: number | null;
		new_masking_price: number | null;
		old_non_masking_price: number | null;
		new_non_masking_price: number | null;
		changed_by: string | null;
		changed_at: string;
	}

	interface PlanOption {
		id: number;
		name: string;
	}

	let records = $state<RateHistoryRecord[]>([]);
	let plans = $state<PlanOption[]>([]);
	let loading = $state(true);

	// Filters
	let selectedPlanId = $state('all');
	let selectedEvent = $state('all');

	// Pagination
	let currentPage = $state(1);
	let perPage = $state(25);
	let totalRecords = $state(0);
	let totalPages = $derived(Math.ceil(totalRecords / perPage) || 1);

	async function loadHistory() {
		loading = true;
		try {
			const query = new URLSearchParams({
				page: String(currentPage),
				per_page: String(perPage)
			});
			if (selectedPlanId !== 'all') query.set('plan_id', selectedPlanId);
			if (selectedEvent !== 'all') query.set('event', selectedEvent);

			const res = await supermanApi.get<{
				data: RateHistoryRecord[];
				meta: { page: number; per_page: number; total: number };
			}>(`/api/v1/superman/billing/rate-history?${query.toString()}`);

			records = res.data || [];
			totalRecords = res.meta?.total || 0;
		} catch (err) {
			console.error('Failed to load rate history:', err);
		} finally {
			loading = false;
		}
	}

	async function loadPlans() {
		try {
			const res = await supermanApi.get<{ data: PlanOption[] }>('/api/v1/superman/billing/tariff-plans');
			plans = res.data || [];
		} catch (err) {
			console.error(err);
		}
	}

	onMount(() => {
		loadPlans();
		loadHistory();
	});

	function formatRate(val: number | null): string {
		if (val === null || val === undefined) return '—';
		return '৳' + val.toFixed(4);
	}
	const exportColumns: ExportColumn<RateHistoryRecord>[] = [
		{ header: 'ID', getValue: (r) => r.id },
		{ header: 'Tariff Plan', getValue: (r) => r.plan_name },
		{ header: 'Operator', getValue: (r) => r.operator_name },
		{ header: 'Event', getValue: (r) => r.event },
		{ header: 'Old Masking Rate', getValue: (r) => (r.old_masking_price !== null ? `BDT ${r.old_masking_price.toFixed(4)}` : '—') },
		{ header: 'New Masking Rate', getValue: (r) => (r.new_masking_price !== null ? `BDT ${r.new_masking_price.toFixed(4)}` : '—') },
		{ header: 'Old Non-Masking Rate', getValue: (r) => (r.old_non_masking_price !== null ? `BDT ${r.old_non_masking_price.toFixed(4)}` : '—') },
		{ header: 'New Non-Masking Rate', getValue: (r) => (r.new_non_masking_price !== null ? `BDT ${r.new_non_masking_price.toFixed(4)}` : '—') },
		{ header: 'Changed By', getValue: (r) => r.changed_by || 'System' },
		{ header: 'Timestamp', getValue: (r) => r.changed_at }
	];
</script>

<svelte:head>
	<title>Rate Change History — SMS Whiteinfo Administrator</title>
</svelte:head>

<div class="space-y-6">
	<!-- Page Header -->
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<div class="flex items-center gap-2">
				<a href="/superman/billing" class="text-xs font-semibold text-gray-500 hover:text-emerald-600 dark:hover:text-emerald-400">
					Billing
				</a>
				<span class="text-xs text-gray-400">/</span>
				<span class="text-xs font-semibold text-gray-900 dark:text-white">Rate History</span>
			</div>
			<h1 class="mt-1 text-xl font-bold tracking-tight text-gray-900 dark:text-white sm:text-2xl">
				Rate Change History & Audit
			</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400">
				Audit trail of all operator price revisions, new rate setups, and plan adjustments over time.
			</p>
		</div>

		<div class="flex items-center gap-2">
			<ExportDropdown
				title="Rate Change History Audit"
				columns={exportColumns}
				getData={() => records}
				disabled={records.length === 0}
			/>
			<a
				href="/superman/billing/tariff-plans"
				class="inline-flex items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs font-semibold text-gray-700 transition hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300 dark:hover:bg-gray-800"
			>
				<svg class="h-4 w-4 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z" />
				</svg>
				Tariff Plans
			</a>
		</div>
	</div>

	<!-- Filters Bar -->
	<div class="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-gray-200 bg-white p-3 shadow-xs dark:border-gray-800 dark:bg-gray-900">
		<div class="flex flex-wrap items-center gap-2">
			<!-- Plan Filter -->
			<div class="flex items-center gap-1.5">
				<span class="text-xs text-gray-500">Plan:</span>
				<select
					bind:value={selectedPlanId}
					onchange={() => {
						currentPage = 1;
						loadHistory();
					}}
					class="rounded-lg border border-gray-200 bg-white px-2.5 py-1.5 text-xs text-gray-700 focus:border-emerald-500 focus:outline-hidden dark:border-gray-800 dark:bg-gray-800 dark:text-gray-300"
				>
					<option value="all">All Tariff Plans</option>
					{#each plans as p}
						<option value={String(p.id)}>{p.name}</option>
					{/each}
				</select>
			</div>

			<!-- Event Filter -->
			<div class="flex items-center gap-1.5">
				<span class="text-xs text-gray-500">Event:</span>
				<select
					bind:value={selectedEvent}
					onchange={() => {
						currentPage = 1;
						loadHistory();
					}}
					class="rounded-lg border border-gray-200 bg-white px-2.5 py-1.5 text-xs text-gray-700 focus:border-emerald-500 focus:outline-hidden dark:border-gray-800 dark:bg-gray-800 dark:text-gray-300"
				>
					<option value="all">All Events</option>
					<option value="created">Created</option>
					<option value="updated">Updated</option>
					<option value="deleted">Deleted</option>
				</select>
			</div>
		</div>

		<div class="flex items-center gap-3 text-xs text-gray-400">
			<span>Total revisions: <strong class="font-mono text-gray-700 dark:text-gray-200">{totalRecords}</strong></span>
			<button
				type="button"
				onclick={loadHistory}
				class="rounded-md border border-gray-200 px-2 py-1 text-gray-600 hover:bg-gray-50 dark:border-gray-700 dark:text-gray-300 dark:hover:bg-gray-800"
			>
				Refresh
			</button>
		</div>
	</div>

	<!-- Rate History Table -->
	<div class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-xs dark:border-gray-800 dark:bg-gray-900">
		<div class="overflow-x-auto">
			<table class="w-full text-left text-xs">
				<thead>
					<tr class="border-b border-gray-100 bg-gray-50/50 text-[10px] font-semibold uppercase tracking-wider text-gray-400 dark:border-gray-800 dark:bg-gray-800/40">
						<th class="py-3 px-4">Plan Name</th>
						<th class="py-3 px-3">Operator</th>
						<th class="py-3 px-3">Event</th>
						<th class="py-3 px-3 text-right">Old Masking</th>
						<th class="py-3 px-3 text-right">New Masking</th>
						<th class="py-3 px-3 text-right">Old Non-Masking</th>
						<th class="py-3 px-3 text-right">New Non-Masking</th>
						<th class="py-3 px-3">Changed By</th>
						<th class="py-3 px-4 text-right">Date & Time</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
					{#each records as r}
						<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30 transition-colors">
							<td class="py-3 px-4 font-bold text-gray-900 dark:text-white">
								{r.plan_name}
							</td>
							<td class="py-3 px-3 font-semibold text-gray-800 dark:text-gray-200">
								{r.operator_name}
							</td>
							<td class="py-3 px-3">
								<span class="inline-flex rounded-full px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wider {r.event === 'created' ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300' : r.event === 'updated' ? 'bg-sky-100 text-sky-800 dark:bg-sky-950 dark:text-sky-300' : 'bg-rose-100 text-rose-800 dark:bg-rose-950 dark:text-rose-300'}">
									{r.event}
								</span>
							</td>
							<td class="py-3 px-3 text-right font-mono text-gray-400 line-through">
								{formatRate(r.old_masking_price)}
							</td>
							<td class="py-3 px-3 text-right font-mono font-bold text-emerald-600 dark:text-emerald-400">
								{formatRate(r.new_masking_price)}
							</td>
							<td class="py-3 px-3 text-right font-mono text-gray-400 line-through">
								{formatRate(r.old_non_masking_price)}
							</td>
							<td class="py-3 px-3 text-right font-mono font-bold text-sky-600 dark:text-sky-400">
								{formatRate(r.new_non_masking_price)}
							</td>
							<td class="py-3 px-3 text-gray-600 dark:text-gray-400">
								{r.changed_by || 'System Automatic'}
							</td>
							<td class="py-3 px-4 text-right font-mono text-[11px] text-gray-400">
								{r.changed_at}
							</td>
						</tr>
					{:else}
						<tr>
							<td colspan="9" class="py-8 text-center text-xs text-gray-400">
								{loading ? 'Loading rate change history…' : 'No rate revisions found matching criteria.'}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>

		<!-- Pagination -->
		{#if totalPages > 1}
			<div class="flex items-center justify-between border-t border-gray-100 px-4 py-3 dark:border-gray-800">
				<p class="text-xs text-gray-500">
					Page <strong class="font-semibold text-gray-900 dark:text-white">{currentPage}</strong> of <strong class="font-semibold text-gray-900 dark:text-white">{totalPages}</strong>
				</p>
				<div class="flex items-center gap-1.5">
					<button
						type="button"
						onclick={() => {
							if (currentPage > 1) {
								currentPage--;
								loadHistory();
							}
						}}
						disabled={currentPage <= 1}
						class="rounded-md border border-gray-200 px-2.5 py-1 text-xs font-semibold text-gray-700 hover:bg-gray-50 disabled:opacity-40 dark:border-gray-700 dark:text-gray-300 dark:hover:bg-gray-800"
					>
						Previous
					</button>
					<button
						type="button"
						onclick={() => {
							if (currentPage < totalPages) {
								currentPage++;
								loadHistory();
							}
						}}
						disabled={currentPage >= totalPages}
						class="rounded-md border border-gray-200 px-2.5 py-1 text-xs font-semibold text-gray-700 hover:bg-gray-50 disabled:opacity-40 dark:border-gray-700 dark:text-gray-300 dark:hover:bg-gray-800"
					>
						Next
					</button>
				</div>
			</div>
		{/if}
	</div>
</div>
