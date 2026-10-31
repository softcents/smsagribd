<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface Operator {
		id: number;
		name: string;
		codes: string;
		masking_price: number;
		non_masking_price: number;
	}

	interface TariffRate {
		id: number;
		operator_id: number;
		operator_name: string;
		masking_price: number;
		non_masking_price: number;
		effective_from: string;
		notes: string | null;
	}

	interface TariffPlan {
		id: number;
		name: string;
		slug: string;
		type: string;
		status: string;
		currency: string;
		description: string | null;
		rates_count: number;
		assigned_users_count: number;
		rates?: TariffRate[];
		created_at: string;
	}

	let plans = $state<TariffPlan[]>([]);
	let operators = $state<Operator[]>([]);
	let loading = $state(true);
	let search = $state('');
	let typeFilter = $state('all');
	let statusFilter = $state('all');

	// Create / Edit Modal state
	let showModal = $state(false);
	let isEditing = $state(false);
	let editPlanId = $state<number | null>(null);
	let formName = $state('');
	let formSlug = $state('');
	let formType = $state('standard');
	let formStatus = $state('active');
	let formDescription = $state('');
	let formRates = $state<
		{ operator_id: number; operator_name: string; masking_price: number; non_masking_price: number; notes: string }[]
	>([]);
	let formSubmitting = $state(false);
	let formError = $state<string | null>(null);

	async function loadData() {
		loading = true;
		try {
			const [plansRes, opRes] = await Promise.all([
				supermanApi.get<{ data: TariffPlan[] }>('/api/v1/superman/billing/tariff-plans'),
				supermanApi.get<{ data: Operator[] }>('/api/v1/superman/operators')
			]);
			plans = plansRes.data || [];
			operators = opRes.data || [];
		} catch (err) {
			console.error('Failed to load tariff plans:', err);
		} finally {
			loading = false;
		}
	}

	function openCreateModal() {
		isEditing = false;
		editPlanId = null;
		formName = '';
		formSlug = '';
		formType = 'standard';
		formStatus = 'active';
		formDescription = '';
		formError = null;

		// Initialize default operator rates from active operators
		formRates = operators.map((op) => ({
			operator_id: op.id,
			operator_name: op.name,
			masking_price: op.masking_price || 0.45,
			non_masking_price: op.non_masking_price || 0.25,
			notes: ''
		}));

		showModal = true;
	}

	async function openEditModal(plan: TariffPlan) {
		isEditing = true;
		editPlanId = plan.id;
		formName = plan.name;
		formSlug = plan.slug;
		formType = plan.type;
		formStatus = plan.status;
		formDescription = plan.description || '';
		formError = null;

		try {
			const detailRes = await supermanApi.get<{ data: TariffPlan }>(`/api/v1/superman/billing/tariff-plans/${plan.id}`);
			const loadedRates = detailRes.data.rates || [];

			formRates = operators.map((op) => {
				const existing = loadedRates.find((r) => r.operator_id === op.id);
				return {
					operator_id: op.id,
					operator_name: op.name,
					masking_price: existing ? existing.masking_price : op.masking_price || 0.45,
					non_masking_price: existing ? existing.non_masking_price : op.non_masking_price || 0.25,
					notes: existing?.notes || ''
				};
			});
		} catch (err) {
			console.error(err);
		}

		showModal = true;
	}

	async function savePlan(e: SubmitEvent) {
		e.preventDefault();
		if (!formName.trim()) {
			formError = 'Plan name is required.';
			return;
		}

		formSubmitting = true;
		formError = null;
		try {
			const payload = {
				name: formName.trim(),
				slug: formSlug.trim(),
				type: formType,
				status: formStatus,
				currency: 'BDT',
				description: formDescription.trim() || null,
				rates: formRates.map((r) => ({
					operator_id: r.operator_id,
					masking_price: Number(r.masking_price),
					non_masking_price: Number(r.non_masking_price),
					notes: r.notes.trim() || null
				}))
			};

			if (isEditing && editPlanId) {
				await supermanApi.put(`/api/v1/superman/billing/tariff-plans/${editPlanId}`, payload);
			} else {
				await supermanApi.post('/api/v1/superman/billing/tariff-plans', payload);
			}

			showModal = false;
			await loadData();
		} catch (err: unknown) {
			formError = err instanceof Error ? err.message : 'Failed to save tariff plan.';
		} finally {
			formSubmitting = false;
		}
	}

	async function deletePlan(plan: TariffPlan) {
		if (plan.assigned_users_count > 0) {
			window.alert(`Cannot delete '${plan.name}'. It is currently assigned to ${plan.assigned_users_count} customer(s).`);
			return;
		}

		if (!window.confirm(`Are you sure you want to permanently delete '${plan.name}'?`)) {
			return;
		}

		try {
			await supermanApi.delete(`/api/v1/superman/billing/tariff-plans/${plan.id}`);
			await loadData();
		} catch (err: unknown) {
			window.alert(err instanceof Error ? err.message : 'Failed to delete tariff plan.');
		}
	}

	onMount(() => {
		loadData();
	});

	let filteredPlans = $derived(
		plans.filter((p) => {
			if (search) {
				const term = search.toLowerCase();
				if (!p.name.toLowerCase().includes(term) && !p.slug.toLowerCase().includes(term)) {
					return false;
				}
			}
			if (typeFilter !== 'all' && p.type !== typeFilter) return false;
			if (statusFilter !== 'all' && p.status !== statusFilter) return false;
			return true;
		})
	);
	const exportColumns: ExportColumn<TariffPlan>[] = [
		{ header: 'ID', getValue: (p) => p.id },
		{ header: 'Plan Name', getValue: (p) => p.name },
		{ header: 'Slug', getValue: (p) => p.slug },
		{ header: 'Plan Type', getValue: (p) => p.type },
		{ header: 'Status', getValue: (p) => p.status },
		{ header: 'Assigned Users', getValue: (p) => p.assigned_users_count },
		{ header: 'Operator Rates Count', getValue: (p) => p.rates_count },
		{ header: 'Created At', getValue: (p) => p.created_at }
	];
</script>

<svelte:head>
	<title>Tariff Plans — SMS Whiteinfo Administrator</title>
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
				<span class="text-xs font-semibold text-gray-900 dark:text-white">Tariff Plans</span>
			</div>
			<h1 class="mt-1 text-xl font-bold tracking-tight text-gray-900 dark:text-white sm:text-2xl">
				Tariff Plans & Operator Pricing
			</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400">
				Manage pricing tiers, operator masking and non-masking rates, and wholesale customer packages.
			</p>
		</div>

		<div class="flex items-center gap-2">
			<ExportDropdown
				title="Tariff Plans and Operator Pricing"
				columns={exportColumns}
				getData={() => filteredPlans}
				disabled={filteredPlans.length === 0}
			/>
			<a
				href="/superman/billing/assignments"
				class="inline-flex items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs font-medium text-gray-700 transition hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300 dark:hover:bg-gray-800"
			>
				<svg class="h-4 w-4 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
				</svg>
				User Assignments
			</a>

			<button
				type="button"
				onclick={openCreateModal}
				class="inline-flex items-center gap-1.5 rounded-lg bg-emerald-600 px-3.5 py-2 text-xs font-semibold text-white shadow-xs transition hover:bg-emerald-500"
			>
				<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
				</svg>
				New Tariff Plan
			</button>
		</div>
	</div>

	<!-- Filter & Search Toolbar -->
	<div class="flex flex-col gap-3 rounded-xl border border-gray-200 bg-white p-3 shadow-xs dark:border-gray-800 dark:bg-gray-900 sm:flex-row sm:items-center sm:justify-between">
		<div class="relative flex-1 max-w-sm">
			<svg class="pointer-events-none absolute left-3 top-2.5 h-4 w-4 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
			</svg>
			<input
				type="text"
				bind:value={search}
				placeholder="Search plan name or slug…"
				class="w-full rounded-lg border border-gray-200 bg-gray-50 py-1.5 pl-9 pr-3 text-xs text-gray-900 focus:border-emerald-500 focus:bg-white focus:outline-hidden dark:border-gray-800 dark:bg-gray-800 dark:text-white"
			/>
		</div>

		<div class="flex items-center gap-2">
			<!-- Type Filter -->
			<select
				bind:value={typeFilter}
				class="rounded-lg border border-gray-200 bg-white px-2.5 py-1.5 text-xs text-gray-700 focus:border-emerald-500 focus:outline-hidden dark:border-gray-800 dark:bg-gray-800 dark:text-gray-300"
			>
				<option value="all">All Types</option>
				<option value="standard">Standard</option>
				<option value="volume">Volume Tier</option>
				<option value="promotional">Promotional</option>
				<option value="custom">Custom / Special</option>
			</select>

			<!-- Status Filter -->
			<select
				bind:value={statusFilter}
				class="rounded-lg border border-gray-200 bg-white px-2.5 py-1.5 text-xs text-gray-700 focus:border-emerald-500 focus:outline-hidden dark:border-gray-800 dark:bg-gray-800 dark:text-gray-300"
			>
				<option value="all">All Statuses</option>
				<option value="active">Active Only</option>
				<option value="inactive">Inactive Only</option>
			</select>

			<span class="text-xs text-gray-400">
				Showing {filteredPlans.length} plan{filteredPlans.length === 1 ? '' : 's'}
			</span>
		</div>
	</div>

	<!-- Tariff Plans Table -->
	<div class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-xs dark:border-gray-800 dark:bg-gray-900">
		<div class="overflow-x-auto">
			<table class="w-full text-left text-xs">
				<thead>
					<tr class="border-b border-gray-100 bg-gray-50/50 text-[10px] font-semibold text-gray-400 uppercase tracking-wider dark:border-gray-800 dark:bg-gray-800/40">
						<th class="py-3 px-4">Plan Name & Details</th>
						<th class="py-3 px-3">Type</th>
						<th class="py-3 px-3">Status</th>
						<th class="py-3 px-3 text-center">Configured Rates</th>
						<th class="py-3 px-3 text-center">Active Users</th>
						<th class="py-3 px-3 text-right">Created</th>
						<th class="py-3 px-4 text-right">Actions</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
					{#each filteredPlans as plan}
						<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30 transition-colors">
							<td class="py-3 px-4">
								<div class="flex items-center gap-2.5">
									<div class="flex h-8 w-8 items-center justify-center rounded-lg bg-emerald-50 font-bold text-emerald-700 dark:bg-emerald-950 dark:text-emerald-400">
										৳
									</div>
									<div>
										<p class="font-bold text-gray-900 dark:text-white">{plan.name}</p>
										<p class="font-mono text-[10px] text-gray-400">{plan.slug} · {plan.currency}</p>
									</div>
								</div>
								{#if plan.description}
									<p class="mt-1 text-[11px] text-gray-500 dark:text-gray-400 line-clamp-1">{plan.description}</p>
								{/if}
							</td>
							<td class="py-3 px-3">
								<span class="inline-flex rounded-md bg-gray-100 px-2 py-0.5 text-[10px] font-semibold capitalize text-gray-700 dark:bg-gray-800 dark:text-gray-300">
									{plan.type}
								</span>
							</td>
							<td class="py-3 px-3">
								<span class="inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[10px] font-semibold {plan.status === 'active' ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300' : 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400'}">
									<span class="h-1.5 w-1.5 rounded-full {plan.status === 'active' ? 'bg-emerald-500' : 'bg-gray-400'}"></span>
									{plan.status}
								</span>
							</td>
							<td class="py-3 px-3 text-center">
								<span class="font-mono font-semibold text-gray-700 dark:text-gray-300">
									{plan.rates_count} operator rates
								</span>
							</td>
							<td class="py-3 px-3 text-center">
								<a
									href="/superman/billing/assignments"
									class="inline-flex items-center gap-1 font-mono font-semibold text-emerald-600 hover:underline dark:text-emerald-400"
								>
									{plan.assigned_users_count} customer{plan.assigned_users_count === 1 ? '' : 's'}
								</a>
							</td>
							<td class="py-3 px-3 text-right font-mono text-[11px] text-gray-400">
								{new Date(plan.created_at).toLocaleDateString([], { month: 'short', day: 'numeric', year: 'numeric' })}
							</td>
							<td class="py-3 px-4 text-right">
								<div class="flex items-center justify-end gap-1.5">
									<button
										type="button"
										onclick={() => openEditModal(plan)}
										class="rounded-lg border border-gray-200 bg-white px-2.5 py-1 text-xs font-semibold text-gray-700 hover:bg-gray-50 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-200"
									>
										Edit Rates
									</button>
									<button
										type="button"
										onclick={() => deletePlan(plan)}
										disabled={plan.assigned_users_count > 0}
										class="rounded-lg border border-rose-200 bg-white px-2 py-1 text-xs font-semibold text-rose-600 hover:bg-rose-50 disabled:opacity-40 disabled:cursor-not-allowed dark:border-rose-900/50 dark:bg-gray-800 dark:text-rose-400"
										title={plan.assigned_users_count > 0 ? 'Cannot delete plan with active users' : 'Delete plan'}
									>
										Delete
									</button>
								</div>
							</td>
						</tr>
					{:else}
						<tr>
							<td colspan="7" class="py-8 text-center text-xs text-gray-400">
								{loading ? 'Loading tariff plans…' : 'No tariff plans matching filters.'}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</div>
</div>

<!-- Modal: Create / Edit Tariff Plan -->
{#if showModal}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-gray-900/60 backdrop-blur-xs p-4">
		<div class="w-full max-w-2xl max-h-[90vh] flex flex-col rounded-2xl border border-gray-200 bg-white shadow-xl dark:border-gray-800 dark:bg-gray-900 animate-in fade-in zoom-in-95 duration-150">
			<!-- Header -->
			<div class="flex items-center justify-between border-b border-gray-100 p-5 dark:border-gray-800">
				<div>
					<h3 class="text-base font-bold text-gray-900 dark:text-white">
						{isEditing ? 'Edit Tariff Plan & Operator Rates' : 'Create New Tariff Plan'}
					</h3>
					<p class="text-xs text-gray-500">
						Set customer pricing per operator for both SMS masking and non-masking routes.
					</p>
				</div>
				<button
					type="button"
					onclick={() => (showModal = false)}
					class="rounded-lg p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-gray-800"
				>
					<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>

			<!-- Body -->
			<form onsubmit={savePlan} class="overflow-y-auto p-5 space-y-4 flex-1">
				{#if formError}
					<div class="rounded-lg bg-rose-50 p-3 text-xs font-semibold text-rose-800 dark:bg-rose-950 dark:text-rose-300">
						{formError}
					</div>
				{/if}

				<div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
					<!-- Plan Name -->
					<div>
						<label for="plan-name" class="block text-xs font-semibold text-gray-700 dark:text-gray-300 mb-1">
							Plan Name <span class="text-rose-500">*</span>
						</label>
						<input
							id="plan-name"
							type="text"
							bind:value={formName}
							required
							placeholder="e.g. Enterprise Tier 1"
							class="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-xs text-gray-900 focus:border-emerald-500 focus:outline-hidden dark:border-gray-700 dark:bg-gray-800 dark:text-white"
						/>
					</div>

					<!-- Slug -->
					<div>
						<label for="plan-slug" class="block text-xs font-semibold text-gray-700 dark:text-gray-300 mb-1">
							Slug (Unique identifier)
						</label>
						<input
							id="plan-slug"
							type="text"
							bind:value={formSlug}
							placeholder="e.g. enterprise-tier-1"
							class="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-xs font-mono text-gray-900 focus:border-emerald-500 focus:outline-hidden dark:border-gray-700 dark:bg-gray-800 dark:text-white"
						/>
					</div>

					<!-- Plan Type -->
					<div>
						<label for="plan-type" class="block text-xs font-semibold text-gray-700 dark:text-gray-300 mb-1">
							Type
						</label>
						<select
							id="plan-type"
							bind:value={formType}
							class="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-xs text-gray-900 focus:border-emerald-500 focus:outline-hidden dark:border-gray-700 dark:bg-gray-800 dark:text-white"
						>
							<option value="standard">Standard</option>
							<option value="volume">Volume Tier</option>
							<option value="promotional">Promotional</option>
							<option value="custom">Custom / Special</option>
						</select>
					</div>

					<!-- Plan Status -->
					<div>
						<label for="plan-status" class="block text-xs font-semibold text-gray-700 dark:text-gray-300 mb-1">
							Status
						</label>
						<select
							id="plan-status"
							bind:value={formStatus}
							class="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-xs text-gray-900 focus:border-emerald-500 focus:outline-hidden dark:border-gray-700 dark:bg-gray-800 dark:text-white"
						>
							<option value="active">Active</option>
							<option value="inactive">Inactive</option>
						</select>
					</div>
				</div>

				<!-- Description -->
				<div>
					<label for="plan-desc" class="block text-xs font-semibold text-gray-700 dark:text-gray-300 mb-1">
						Description / Internal Notes
					</label>
					<textarea
						id="plan-desc"
						bind:value={formDescription}
						rows="2"
						placeholder="Optional notes regarding this pricing tier…"
						class="w-full rounded-lg border border-gray-300 bg-white p-2.5 text-xs text-gray-900 focus:border-emerald-500 focus:outline-hidden dark:border-gray-700 dark:bg-gray-800 dark:text-white"
					></textarea>
				</div>

				<!-- Operator Rates Table -->
				<div class="border-t border-gray-100 pt-3 dark:border-gray-800">
					<h4 class="text-xs font-bold uppercase tracking-wider text-gray-700 dark:text-gray-300 mb-2">
						Operator Rates (BDT per SMS)
					</h4>
					<div class="overflow-x-auto rounded-lg border border-gray-200 dark:border-gray-800">
						<table class="w-full text-left text-xs">
							<thead>
								<tr class="border-b border-gray-200 bg-gray-50 text-[10px] font-semibold uppercase text-gray-500 dark:border-gray-800 dark:bg-gray-800">
									<th class="py-2 px-3">Operator</th>
									<th class="py-2 px-3">Masking Price (BDT)</th>
									<th class="py-2 px-3">Non-Masking Price (BDT)</th>
									<th class="py-2 px-3">Remarks</th>
								</tr>
							</thead>
							<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
								{#each formRates as rate, idx}
									<tr>
										<td class="py-2 px-3 font-semibold text-gray-800 dark:text-gray-200">
											{rate.operator_name}
										</td>
										<td class="py-2 px-3">
											<div class="relative max-w-[120px]">
												<span class="absolute inset-y-0 left-0 flex items-center pl-2 font-mono text-[10px] text-gray-400">৳</span>
												<input
													type="number"
													step="0.0001"
													min="0.0000"
													bind:value={formRates[idx].masking_price}
													required
													class="w-full rounded-md border border-gray-300 bg-white py-1 pl-6 pr-2 font-mono text-xs text-gray-900 focus:border-emerald-500 focus:outline-hidden dark:border-gray-700 dark:bg-gray-800 dark:text-white"
												/>
											</div>
										</td>
										<td class="py-2 px-3">
											<div class="relative max-w-[120px]">
												<span class="absolute inset-y-0 left-0 flex items-center pl-2 font-mono text-[10px] text-gray-400">৳</span>
												<input
													type="number"
													step="0.0001"
													min="0.0000"
													bind:value={formRates[idx].non_masking_price}
													required
													class="w-full rounded-md border border-gray-300 bg-white py-1 pl-6 pr-2 font-mono text-xs text-gray-900 focus:border-emerald-500 focus:outline-hidden dark:border-gray-700 dark:bg-gray-800 dark:text-white"
												/>
											</div>
										</td>
										<td class="py-2 px-3">
											<input
												type="text"
												bind:value={formRates[idx].notes}
												placeholder="Standard rate"
												class="w-full rounded-md border border-gray-200 bg-white py-1 px-2 text-xs text-gray-700 focus:border-emerald-500 focus:outline-hidden dark:border-gray-700 dark:bg-gray-800 dark:text-white"
											/>
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				</div>

				<!-- Footer -->
				<div class="flex items-center justify-end gap-2 pt-3 border-t border-gray-100 dark:border-gray-800">
					<button
						type="button"
						onclick={() => (showModal = false)}
						class="rounded-lg border border-gray-200 px-3.5 py-2 text-xs font-medium text-gray-700 hover:bg-gray-50 dark:border-gray-800 dark:text-gray-300 dark:hover:bg-gray-800"
					>
						Cancel
					</button>
					<button
						type="submit"
						disabled={formSubmitting}
						class="rounded-lg bg-emerald-600 px-4 py-2 text-xs font-semibold text-white shadow-xs hover:bg-emerald-500 disabled:opacity-50"
					>
						{formSubmitting ? 'Saving…' : isEditing ? 'Update Plan & Rates' : 'Create Tariff Plan'}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
