<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface TariffPlan {
		id: number;
		name: string;
		type: string;
		status: string;
	}

	interface CustomerItem {
		user_id: number;
		customer_name: string;
		customer_email: string;
		customer_phone: string;
		plan_id: number | null;
		plan_name: string | null;
		plan_type: string | null;
		effective_from: string | null;
		assigned_by: string | null;
		status: string | null;
	}

	interface AssignmentHistory {
		id: number;
		plan_name: string;
		plan_type: string;
		status: string;
		effective_from: string;
		assigned_by: string | null;
		created_at: string;
		notes: string | null;
	}

	let customers = $state<CustomerItem[]>([]);
	let activePlans = $state<TariffPlan[]>([]);
	let loading = $state(true);
	let search = $state('');

	// Currently selected customer for assignment / details
	let selectedCustomerId = $state<number | null>(null);
	let selectedCustomer = $derived(
		customers.find((c) => c.user_id === selectedCustomerId) || null
	);

	let assignmentHistory = $state<AssignmentHistory[]>([]);
	let historyLoading = $state(false);

	// Form inputs for new assignment
	let newPlanId = $state<number | ''>('');
	let effectiveFrom = $state(new Date().toISOString().slice(0, 10));
	let assignmentNotes = $state('');
	let submitting = $state(false);
	let successMsg = $state<string | null>(null);
	let errorMsg = $state<string | null>(null);

	async function loadInitial() {
		loading = true;
		try {
			const [custRes, plansRes] = await Promise.all([
				supermanApi.get<{ data: CustomerItem[] }>('/api/v1/superman/billing/assignments'),
				supermanApi.get<{ data: TariffPlan[] }>('/api/v1/superman/billing/tariff-plans')
			]);
			customers = custRes.data || [];
			activePlans = (plansRes.data || []).filter((p) => p.status === 'active');
		} catch (err) {
			console.error('Failed to load assignments data:', err);
		} finally {
			loading = false;
		}
	}

	async function selectCustomer(userId: number) {
		selectedCustomerId = userId;
		newPlanId = '';
		effectiveFrom = new Date().toISOString().slice(0, 10);
		assignmentNotes = '';
		successMsg = null;
		errorMsg = null;

		historyLoading = true;
		try {
			const res = await supermanApi.get<{ data: AssignmentHistory[] }>(
				`/api/v1/superman/billing/assignments/${userId}`
			);
			assignmentHistory = res.data || [];
		} catch (err) {
			console.error('Failed to load assignment history:', err);
		} finally {
			historyLoading = false;
		}
	}

	async function handleAssignPlan(e: SubmitEvent) {
		e.preventDefault();
		if (!selectedCustomerId || !newPlanId) {
			errorMsg = 'Please select a tariff plan to assign.';
			return;
		}

		submitting = true;
		errorMsg = null;
		successMsg = null;
		try {
			const res = await supermanApi.post<{ message: string }>('/api/v1/superman/billing/assignments', {
				user_id: selectedCustomerId,
				tariff_plan_id: Number(newPlanId),
				effective_from: effectiveFrom,
				notes: assignmentNotes.trim() || null
			});

			successMsg = res.message;
			assignmentNotes = '';
			newPlanId = '';

			// Refresh history and customer list
			await Promise.all([
				selectCustomer(selectedCustomerId),
				supermanApi.get<{ data: CustomerItem[] }>('/api/v1/superman/billing/assignments').then((r) => {
					customers = r.data || [];
				})
			]);
		} catch (err: unknown) {
			errorMsg = err instanceof Error ? err.message : 'Failed to assign tariff plan.';
		} finally {
			submitting = false;
		}
	}

	onMount(() => {
		loadInitial();
	});

	let filteredCustomers = $derived(
		customers.filter((c) => {
			if (!search) return true;
			const term = search.toLowerCase();
			return (
				c.customer_name.toLowerCase().includes(term) ||
				c.customer_email.toLowerCase().includes(term) ||
				c.customer_phone.toLowerCase().includes(term) ||
				(c.plan_name && c.plan_name.toLowerCase().includes(term))
			);
		})
	);
	const exportColumns: ExportColumn<CustomerItem>[] = [
		{ header: 'Customer Name', getValue: (c) => c.customer_name },
		{ header: 'Email', getValue: (c) => c.customer_email },
		{ header: 'Phone', getValue: (c) => c.customer_phone },
		{ header: 'Current Plan', getValue: (c) => c.plan_name || 'No Plan' },
		{ header: 'Plan Type', getValue: (c) => c.plan_type || '—' },
		{ header: 'Effective From', getValue: (c) => c.effective_from || '—' },
		{ header: 'Status', getValue: (c) => c.status || 'Active' }
	];
</script>

<svelte:head>
	<title>User-wise Tariff Assignment — SMS Whiteinfo Administrator</title>
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
				<span class="text-xs font-semibold text-gray-900 dark:text-white">User-wise Assignment</span>
			</div>
			<h1 class="mt-1 text-xl font-bold tracking-tight text-gray-900 dark:text-white sm:text-2xl">
				User-wise Tariff Plan Assignment
			</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400">
				Assign specific tariff rate schedules to individual customers and inspect full assignment revision history.
			</p>
		</div>

		<div class="flex items-center gap-2">
			<ExportDropdown
				title="User Tariff Plan Assignments"
				columns={exportColumns}
				getData={() => filteredCustomers}
				disabled={filteredCustomers.length === 0}
			/>
			<a
				href="/superman/billing/tariff-plans"
				class="inline-flex items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs font-semibold text-gray-700 transition hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300 dark:hover:bg-gray-800"
			>
				<svg class="h-4 w-4 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z" />
				</svg>
				Manage Plans
			</a>
			<a
				href="/superman/billing/rate-history"
				class="inline-flex items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs font-semibold text-gray-700 transition hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300 dark:hover:bg-gray-800"
			>
				<svg class="h-4 w-4 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
				</svg>
				Rate Change History
			</a>
		</div>
	</div>

	<!-- Top Customer Selection Section -->
	<div class="rounded-xl border border-gray-200 bg-white p-5 shadow-xs dark:border-gray-800 dark:bg-gray-900">
		<h2 class="text-sm font-bold text-gray-900 dark:text-white">Select Customer</h2>
		<p class="text-xs text-gray-500 dark:text-gray-400">
			Choose a customer to review active tariff schedules or switch them to a different tier.
		</p>

		<div class="mt-4 flex flex-col gap-3 sm:flex-row sm:items-center">
			<div class="flex-1 max-w-lg">
				<select
					value={selectedCustomerId ?? ''}
					onchange={(e) => {
						const val = (e.currentTarget as HTMLSelectElement).value;
						if (val) selectCustomer(Number(val));
						else selectedCustomerId = null;
					}}
					class="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-xs font-medium text-gray-900 focus:border-emerald-500 focus:outline-hidden dark:border-gray-700 dark:bg-gray-800 dark:text-white"
				>
					<option value="">— Select a customer to inspect —</option>
					{#each customers as c}
						<option value={c.user_id}>
							{c.customer_name} ({c.customer_phone || c.customer_email}) — Plan: {c.plan_name || 'Default Legacy'}
						</option>
					{/each}
				</select>
			</div>

			{#if selectedCustomer}
				<div class="flex items-center gap-2">
					<span class="rounded-md bg-emerald-50 px-2.5 py-1 text-xs font-semibold text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300">
						Active: {selectedCustomer.customer_name}
					</span>
					<button
						type="button"
						onclick={() => (selectedCustomerId = null)}
						class="text-xs text-gray-400 hover:text-gray-600 dark:hover:text-gray-200"
					>
						Clear
					</button>
				</div>
			{/if}
		</div>
	</div>

	<!-- Detail Section when customer is selected -->
	{#if selectedCustomer}
		<div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
			<!-- Current Plan Card -->
			<div class="rounded-xl border border-gray-200 bg-white p-5 shadow-xs dark:border-gray-800 dark:bg-gray-900 lg:col-span-1">
				<h3 class="text-xs font-bold uppercase tracking-wider text-gray-400 dark:text-gray-500">
					Current Tariff Plan
				</h3>

				<div class="mt-4 space-y-4">
					<div>
						<p class="text-xs text-gray-500 dark:text-gray-400">Plan Name</p>
						<p class="mt-0.5 text-base font-bold text-gray-900 dark:text-white">
							{selectedCustomer.plan_name || 'Standard Legacy Rates'}
						</p>
					</div>

					<div class="grid grid-cols-2 gap-4">
						<div>
							<p class="text-xs text-gray-500 dark:text-gray-400">Plan Type</p>
							<p class="mt-0.5 text-xs font-semibold uppercase text-gray-800 dark:text-gray-200">
								{selectedCustomer.plan_type || 'Default'}
							</p>
						</div>
						<div>
							<p class="text-xs text-gray-500 dark:text-gray-400">Effective Date</p>
							<p class="mt-0.5 font-mono text-xs text-gray-800 dark:text-gray-200">
								{selectedCustomer.effective_from || '—'}
							</p>
						</div>
					</div>

					<div>
						<p class="text-xs text-gray-500 dark:text-gray-400">Assigned By</p>
						<p class="mt-0.5 text-xs text-gray-800 dark:text-gray-200">
							{selectedCustomer.assigned_by || 'System Automatic'}
						</p>
					</div>

					<div class="rounded-lg bg-emerald-50/50 p-3 text-xs text-emerald-800 dark:bg-emerald-950/40 dark:text-emerald-300">
						Pricing resolves directly according to this plan for all customer outbound SMS routing.
					</div>
				</div>
			</div>

			<!-- Assign / Change Plan Form -->
			<div class="rounded-xl border border-gray-200 bg-white p-5 shadow-xs dark:border-gray-800 dark:bg-gray-900 lg:col-span-2">
				<h3 class="text-xs font-bold uppercase tracking-wider text-gray-400 dark:text-gray-500">
					Assign / Change Tariff Plan
				</h3>
				<p class="mt-1 text-xs text-gray-500">
					Selecting a new plan immediately supersedes the active one. The effective date dictates billing timestamp.
				</p>

				<form onsubmit={handleAssignPlan} class="mt-4 space-y-4">
					{#if successMsg}
						<div class="rounded-lg bg-emerald-50 p-3 text-xs font-semibold text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300">
							{successMsg}
						</div>
					{/if}
					{#if errorMsg}
						<div class="rounded-lg bg-rose-50 p-3 text-xs font-semibold text-rose-800 dark:bg-rose-950 dark:text-rose-300">
							{errorMsg}
						</div>
					{/if}

					<div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
						<!-- New Plan Select -->
						<div>
							<label for="new-plan" class="block text-xs font-semibold text-gray-700 dark:text-gray-300 mb-1">
								New Tariff Plan <span class="text-rose-500">*</span>
							</label>
							<select
								id="new-plan"
								bind:value={newPlanId}
								required
								class="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-xs font-medium text-gray-900 focus:border-emerald-500 focus:outline-hidden dark:border-gray-700 dark:bg-gray-800 dark:text-white"
							>
								<option value="">— Select an active plan —</option>
								{#each activePlans as p}
									<option value={p.id}>{p.name} ({p.type})</option>
								{/each}
							</select>
						</div>

						<!-- Effective From -->
						<div>
							<label for="effective-from" class="block text-xs font-semibold text-gray-700 dark:text-gray-300 mb-1">
								Effective Date <span class="text-rose-500">*</span>
							</label>
							<input
								id="effective-from"
								type="date"
								bind:value={effectiveFrom}
								required
								class="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-xs font-mono text-gray-900 focus:border-emerald-500 focus:outline-hidden dark:border-gray-700 dark:bg-gray-800 dark:text-white"
							/>
						</div>
					</div>

					<!-- Notes -->
					<div>
						<label for="assign-notes" class="block text-xs font-semibold text-gray-700 dark:text-gray-300 mb-1">
							Revision Notes / Reason (Optional)
						</label>
						<input
							id="assign-notes"
							type="text"
							bind:value={assignmentNotes}
							placeholder="e.g. Special custom volume deal requested by customer"
							class="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-xs text-gray-900 focus:border-emerald-500 focus:outline-hidden dark:border-gray-700 dark:bg-gray-800 dark:text-white"
						/>
					</div>

					<div class="flex items-center justify-end">
						<button
							type="submit"
							disabled={submitting || !newPlanId}
							class="inline-flex items-center gap-1.5 rounded-lg bg-emerald-600 px-4 py-2 text-xs font-semibold text-white shadow-xs hover:bg-emerald-500 disabled:opacity-50"
						>
							<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
							</svg>
							{submitting ? 'Assigning…' : 'Reassign Customer Plan'}
						</button>
					</div>
				</form>
			</div>
		</div>

		<!-- Assignment History Table for Selected Customer -->
		<div class="rounded-xl border border-gray-200 bg-white p-5 shadow-xs dark:border-gray-800 dark:bg-gray-900">
			<div class="mb-3 flex items-center justify-between">
				<div>
					<h3 class="text-xs font-bold uppercase tracking-wider text-gray-700 dark:text-gray-300">
						Assignment History for {selectedCustomer.customer_name}
					</h3>
					<p class="text-[11px] text-gray-400">All past and superseded plan assignments</p>
				</div>
			</div>

			<div class="overflow-x-auto">
				<table class="w-full text-left text-xs">
					<thead>
						<tr class="border-b border-gray-100 text-[10px] font-semibold uppercase text-gray-400 dark:border-gray-800">
							<th class="py-2.5">Plan</th>
							<th class="py-2.5">Status</th>
							<th class="py-2.5">Effective Date</th>
							<th class="py-2.5">Assigned By</th>
							<th class="py-2.5">Assigned At</th>
							<th class="py-2.5">Notes</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
						{#each assignmentHistory as h}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/40">
								<td class="py-2.5 font-bold text-gray-900 dark:text-white">
									{h.plan_name}
									<span class="ml-1 text-[10px] font-normal text-gray-400">({h.plan_type})</span>
								</td>
								<td class="py-2.5">
									<span class="inline-flex rounded-full px-2 py-0.5 text-[9px] font-semibold {h.status === 'active' ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300' : 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400'}">
										{h.status}
									</span>
								</td>
								<td class="py-2.5 font-mono text-gray-700 dark:text-gray-300">{h.effective_from}</td>
								<td class="py-2.5 text-gray-600 dark:text-gray-400">{h.assigned_by || 'System'}</td>
								<td class="py-2.5 font-mono text-[10px] text-gray-400">
									{new Date(h.created_at).toLocaleString([], { dateStyle: 'medium', timeStyle: 'short' })}
								</td>
								<td class="py-2.5 text-gray-500 dark:text-gray-400">{h.notes || '—'}</td>
							</tr>
						{:else}
							<tr>
								<td colspan="6" class="py-4 text-center text-xs text-gray-400">
									{historyLoading ? 'Loading assignment history…' : 'No previous plan assignments recorded for this customer.'}
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>
	{/if}

	<!-- Customer Assignments Master Table -->
	<div class="rounded-xl border border-gray-200 bg-white shadow-xs dark:border-gray-800 dark:bg-gray-900">
		<div class="flex flex-col gap-3 p-4 sm:flex-row sm:items-center sm:justify-between border-b border-gray-100 dark:border-gray-800">
			<div>
				<h3 class="text-xs font-bold uppercase tracking-wider text-gray-700 dark:text-gray-300">
					All Customer Plan Assignments
				</h3>
				<p class="text-[11px] text-gray-400">Search customer records and click any row to reassign</p>
			</div>

			<div class="relative w-full max-w-xs">
				<input
					type="text"
					bind:value={search}
					placeholder="Filter by customer or plan…"
					class="w-full rounded-lg border border-gray-200 bg-gray-50 py-1.5 pl-3 pr-8 text-xs text-gray-900 focus:border-emerald-500 focus:bg-white focus:outline-hidden dark:border-gray-800 dark:bg-gray-800 dark:text-white"
				/>
				<svg class="pointer-events-none absolute right-2.5 top-2 h-4 w-4 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
				</svg>
			</div>
		</div>

		<div class="overflow-x-auto">
			<table class="w-full text-left text-xs">
				<thead>
					<tr class="border-b border-gray-100 bg-gray-50/50 text-[10px] font-semibold uppercase text-gray-400 dark:border-gray-800 dark:bg-gray-800/40">
						<th class="py-3 px-4">Customer</th>
						<th class="py-3 px-3">Contact</th>
						<th class="py-3 px-3">Assigned Plan</th>
						<th class="py-3 px-3">Type</th>
						<th class="py-3 px-3">Effective Date</th>
						<th class="py-3 px-4 text-right">Action</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
					{#each filteredCustomers as c}
						<tr
							onclick={() => selectCustomer(c.user_id)}
							class="cursor-pointer transition hover:bg-emerald-50/40 dark:hover:bg-emerald-950/20 {selectedCustomerId === c.user_id ? 'bg-emerald-50/60 dark:bg-emerald-950/40' : ''}"
						>
							<td class="py-3 px-4">
								<p class="font-bold text-gray-900 dark:text-white">{c.customer_name}</p>
								<p class="text-[10px] text-gray-400">ID #{c.user_id}</p>
							</td>
							<td class="py-3 px-3">
								<p class="font-mono text-gray-700 dark:text-gray-300">{c.customer_phone || 'No phone'}</p>
								<p class="text-[10px] text-gray-400">{c.customer_email}</p>
							</td>
							<td class="py-3 px-3">
								{#if c.plan_name}
									<span class="font-semibold text-emerald-700 dark:text-emerald-400">
										{c.plan_name}
									</span>
								{:else}
									<span class="text-gray-400 italic">Default Legacy</span>
								{/if}
							</td>
							<td class="py-3 px-3">
								<span class="inline-flex rounded-md bg-gray-100 px-2 py-0.5 text-[10px] font-semibold capitalize text-gray-700 dark:bg-gray-800 dark:text-gray-300">
									{c.plan_type || 'standard'}
								</span>
							</td>
							<td class="py-3 px-3 font-mono text-gray-500">
								{c.effective_from || '—'}
							</td>
							<td class="py-3 px-4 text-right">
								<button
									type="button"
									onclick={(e) => {
										e.stopPropagation();
										selectCustomer(c.user_id);
									}}
									class="rounded-lg border border-gray-200 bg-white px-2.5 py-1 text-xs font-semibold text-gray-700 shadow-2xs hover:bg-gray-50 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-200"
								>
									{selectedCustomerId === c.user_id ? 'Selected' : 'Assign'}
								</button>
							</td>
						</tr>
					{:else}
						<tr>
							<td colspan="6" class="py-8 text-center text-xs text-gray-400">
								{loading ? 'Loading customer assignments…' : 'No customers matching search.'}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</div>
</div>
