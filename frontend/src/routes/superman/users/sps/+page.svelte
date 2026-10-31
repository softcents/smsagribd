<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface SPReseller {
		id: number;
		name: string;
		email: string;
		phone: string;
		business_name: string | null;
		balance: number;
		status: string;
		wholesale_plan?: string | null;
		customer_count: number;
		created_at: string;
	}

	interface SPSummary {
		total_sps: number;
		total_balance: number;
		total_customers: number;
		active_sps_count: number;
		active_rate: number;
	}

	interface TariffPlan {
		id: number;
		name: string;
	}

	interface SPDetail {
		sp: SPReseller;
		api_access?: boolean;
		sms_auto_approved?: boolean;
		wholesale_plan_id?: number | null;
		customers?: Array<{ id: number; name: string; email: string; phone: string; balance: number; status: string; created_at: string }>;
		recent_transactions?: Array<{ id: number; amount: number; direction: string; balance_before: number; balance_after: number; description: string | null; created_at: string }>;
	}

	let statusFilter = $state('all');
	let searchQuery = $state('');

	// Pagination
	let currentPage = $state(1);
	let perPage = $state(10);
	let totalRecords = $state(0);

	let sps = $state<SPReseller[]>([]);
	let tariffPlans = $state<TariffPlan[]>([]);
	let summary = $state<SPSummary>({
		total_sps: 0,
		total_balance: 0,
		total_customers: 0,
		active_sps_count: 0,
		active_rate: 0
	});

	let loading = $state(false);
	let activeMenuId = $state<number | null>(null);

	// Modals
	let depositModalOpen = $state(false);
	let selectedSP = $state<SPReseller | null>(null);
	let depositAmount = $state<number>(1000);
	let depositDescription = $state('');
	let depositSubmitting = $state(false);
	let depositError = $state<string | null>(null);

	let viewModalOpen = $state(false);
	let detailLoading = $state(false);
	let currentDetail = $state<SPDetail | null>(null);

	let editModalOpen = $state(false);
	let editSPData = $state({
		id: 0,
		name: '',
		email: '',
		phone: '',
		business_name: '',
		status: 'approved',
		api_access: false,
		sms_auto_approved: false,
		tariff_plan_id: 0
	});
	let editSubmitting = $state(false);

	function resetFilters() {
		statusFilter = 'all';
		searchQuery = '';
		currentPage = 1;
		void loadSPs();
	}

	async function loadTariffPlans() {
		try {
			const res = await supermanApi.get<{ data: TariffPlan[] }>('/api/v1/superman/tariff-plans');
			tariffPlans = res.data ?? [];
		} catch {
			tariffPlans = [];
		}
	}

	async function loadSPs() {
		loading = true;
		try {
			const params = new URLSearchParams({
				page: String(currentPage),
				per_page: String(perPage)
			});
			if (searchQuery.trim()) params.set('search', searchQuery.trim());
			if (statusFilter !== 'all') params.set('status', statusFilter);

			const response = await supermanApi.get<{
				data: SPReseller[];
				meta: { page: number; per_page: number; total: number };
				summary?: SPSummary;
			}>(`/api/v1/superman/sps?${params.toString()}`);

			sps = response.data ?? [];
			totalRecords = response.meta?.total ?? sps.length;

			if (response.summary) {
				summary = response.summary;
			} else {
				computeClientSummary(sps);
			}
		} catch (error: unknown) {
			window.alert(error instanceof Error ? error.message : 'Unable to load resellers.');
		} finally {
			loading = false;
		}
	}

	function computeClientSummary(list: SPReseller[]) {
		const total = list.length;
		const totalBal = list.reduce((acc, s) => acc + (s.balance || 0), 0);
		const totalCust = list.reduce((acc, s) => acc + (s.customer_count || 0), 0);
		const active = list.filter((s) => s.status === 'approved').length;

		summary = {
			total_sps: total,
			total_balance: totalBal,
			total_customers: totalCust,
			active_sps_count: active,
			active_rate: total > 0 ? Math.round((active / total) * 1000) / 10 : 0
		};
	}

	function handleRowClick(sp: SPReseller, e: MouseEvent) {
		const target = e.target as HTMLElement;
		if (target.closest('button') || target.closest('a') || target.closest('select') || target.closest('.no-row-nav')) {
			return;
		}
		goto(`/superman/users/sps/${sp.id}`);
	}

	function openDepositModal(sp: SPReseller) {
		selectedSP = sp;
		depositAmount = 1000;
		depositDescription = '';
		depositError = null;
		depositModalOpen = true;
		activeMenuId = null;
	}

	async function submitDeposit() {
		if (!selectedSP || depositAmount === 0) return;
		depositSubmitting = true;
		depositError = null;
		try {
			const res = await supermanApi.post<{ message: string; data: { balance: number } }>(
				`/api/v1/superman/sps/${selectedSP.id}/deposit`,
				{
					amount: depositAmount,
					description: depositDescription
				}
			);
			selectedSP.balance = res.data.balance;
			depositModalOpen = false;
			window.alert(res.message);
			void loadSPs();
		} catch (error: unknown) {
			depositError = error instanceof Error ? error.message : 'Unable to adjust reseller balance.';
		} finally {
			depositSubmitting = false;
		}
	}

	async function openViewModal(sp: SPReseller) {
		selectedSP = sp;
		viewModalOpen = true;
		detailLoading = true;
		currentDetail = null;
		activeMenuId = null;
		try {
			const res = await supermanApi.get<{ data: SPDetail }>(`/api/v1/superman/sps/${sp.id}`);
			currentDetail = res.data;
		} catch (error: unknown) {
			window.alert(error instanceof Error ? error.message : 'Unable to load reseller details.');
			viewModalOpen = false;
		} finally {
			detailLoading = false;
		}
	}

	async function openEditModal(sp: SPReseller) {
		selectedSP = sp;
		activeMenuId = null;
		editModalOpen = true;
		editSPData = {
			id: sp.id,
			name: sp.name,
			email: sp.email,
			phone: sp.phone,
			business_name: sp.business_name ?? '',
			status: sp.status,
			api_access: false,
			sms_auto_approved: false,
			tariff_plan_id: 0
		};

		try {
			const res = await supermanApi.get<{ data: SPDetail }>(`/api/v1/superman/sps/${sp.id}`);
			if (res.data) {
				editSPData.api_access = Boolean(res.data.api_access);
				editSPData.sms_auto_approved = Boolean(res.data.sms_auto_approved);
				editSPData.tariff_plan_id = res.data.wholesale_plan_id ?? 0;
			}
		} catch {
			// keep defaults
		}
	}

	async function submitEdit() {
		if (!editSPData.id) return;
		editSubmitting = true;
		try {
			await supermanApi.put(`/api/v1/superman/sps/${editSPData.id}`, {
				name: editSPData.name,
				email: editSPData.email,
				phone: editSPData.phone,
				business_name: editSPData.business_name,
				status: editSPData.status,
				api_access: editSPData.api_access,
				sms_auto_approved: editSPData.sms_auto_approved,
				tariff_plan_id: editSPData.tariff_plan_id > 0 ? editSPData.tariff_plan_id : null
			});
			editModalOpen = false;
			window.alert('Reseller updated successfully.');
			void loadSPs();
		} catch (error: unknown) {
			window.alert(error instanceof Error ? error.message : 'Unable to update reseller.');
		} finally {
			editSubmitting = false;
		}
	}

	function handleImpersonate(sp: SPReseller) {
		activeMenuId = null;
		window.alert(`Simulating Login As Reseller ${sp.name} (${sp.email})...`);
	}

	function closeMenu() {
		activeMenuId = null;
	}

	onMount(() => {
		void loadTariffPlans();
		void loadSPs();
	});

	let totalPages = $derived(Math.max(1, Math.ceil(totalRecords / perPage)));
	let startRecord = $derived(totalRecords === 0 ? 0 : (currentPage - 1) * perPage + 1);
	let endRecord = $derived(Math.min(currentPage * perPage, totalRecords));

	const exportColumns: ExportColumn<SPReseller>[] = [
		{ header: 'Reseller Name', getValue: (s) => s.name },
		{ header: 'Email', getValue: (s) => s.email },
		{ header: 'Phone', getValue: (s) => s.phone },
		{ header: 'Business / Company', getValue: (s) => s.business_name || '—' },
		{ header: 'Balance (BDT)', getValue: (s) => `BDT ${s.balance.toFixed(2)}` },
		{ header: 'Customers Count', getValue: (s) => s.customer_count },
		{ header: 'Account Status', getValue: (s) => s.status.toUpperCase() },
		{ header: 'Wholesale Plan', getValue: (s) => s.wholesale_plan || '—' },
		{ header: 'Created Date', getValue: (s) => new Date(s.created_at).toLocaleDateString() }
	];
</script>

<svelte:window onclick={closeMenu} />

<svelte:head>
	<title>S P Resellers Management — SMS Whiteinfo Administrator</title>
</svelte:head>

<div class="space-y-6 text-gray-900 dark:text-white antialiased font-sans">
	<!-- Top Breadcrumb & Page Header -->
	<div class="space-y-1">
		<nav class="flex items-center gap-1.5 text-xs text-gray-500 dark:text-gray-400">
			<span>S P</span>
			<span>›</span>
			<span class="font-medium text-gray-900 dark:text-white">List</span>
		</nav>
		<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 pt-1">
			<h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">S P</h1>
			<div class="flex items-center gap-2">
				<ExportDropdown
					title="SP Resellers List"
					columns={exportColumns}
					getData={() => sps}
					disabled={sps.length === 0}
				/>
				<button
					type="button"
					onclick={() => {
						editSPData = { id: 0, name: '', email: '', phone: '', business_name: '', status: 'approved', api_access: false, sms_auto_approved: false, tariff_plan_id: 0 };
						editModalOpen = true;
					}}
					class="inline-flex items-center gap-1.5 rounded-lg bg-emerald-600 px-3.5 py-1.5 text-xs font-bold text-white shadow-2xs hover:bg-emerald-700 transition"
				>
					<svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke-width="2.5" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
					</svg>
					New S P
				</button>
			</div>
		</div>
	</div>

	<!-- Filters Panel -->
	<div class="rounded-xl bg-white p-5 shadow-xs ring-1 ring-gray-950/5 dark:bg-gray-900 dark:ring-white/10">
		<div class="flex items-center justify-between pb-3 border-b border-gray-100 dark:border-gray-800">
			<h2 class="text-xs font-bold uppercase tracking-wider text-gray-700 dark:text-gray-300">Filters</h2>
			<button
				type="button"
				onclick={resetFilters}
				class="text-xs font-semibold text-rose-600 hover:text-rose-700 dark:text-rose-400 hover:underline"
			>
				Reset
			</button>
		</div>

		<div class="mt-4 grid grid-cols-1 sm:grid-cols-3 gap-3">
			<div>
				<label class="block text-[11px] font-medium text-gray-700 dark:text-gray-300 mb-1">Status</label>
				<select
					bind:value={statusFilter}
					onchange={() => { currentPage = 1; void loadSPs(); }}
					class="w-full rounded-lg border border-gray-300 bg-white px-2.5 py-1.5 text-xs text-gray-900 shadow-2xs outline-none focus:border-emerald-500 focus:ring-1 focus:ring-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
				>
					<option value="all">All</option>
					<option value="approved">Approved</option>
					<option value="pending">Pending</option>
					<option value="rejected">Rejected</option>
					<option value="banned">Banned</option>
				</select>
			</div>
		</div>
	</div>

	<!-- SP Resellers Table Card -->
	<div class="rounded-xl bg-white shadow-xs ring-1 ring-gray-950/5 dark:bg-gray-900 dark:ring-white/10 overflow-hidden">
		<!-- Search Toolbar -->
		<div class="flex items-center justify-end gap-3 px-4 py-3 border-b border-gray-100 dark:border-gray-800">
			<div class="relative w-72">
				<svg class="absolute left-3 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z" />
				</svg>
				<input
					type="text"
					bind:value={searchQuery}
					oninput={() => { currentPage = 1; void loadSPs(); }}
					placeholder="Search"
					class="w-full pl-9 pr-3 py-1.5 rounded-lg border border-gray-200 bg-white dark:border-gray-700 dark:bg-gray-800 dark:text-white text-xs outline-none focus:border-emerald-500 focus:ring-1 focus:ring-emerald-500 placeholder:text-gray-400"
				/>
			</div>
		</div>

		<!-- Table -->
		<div class="overflow-x-auto">
			<table class="w-full text-xs text-left">
				<thead class="bg-gray-50/75 dark:bg-gray-800/50 border-b border-gray-200/80 dark:border-gray-800 text-[10px] uppercase font-bold text-gray-500 dark:text-gray-400 tracking-wider">
					<tr>
						<th class="px-4 py-3">NAME</th>
						<th class="px-4 py-3">EMAIL</th>
						<th class="px-4 py-3">PHONE</th>
						<th class="px-4 py-3">WALLET</th>
						<th class="px-4 py-3 text-center">CUSTOMERS</th>
						<th class="px-4 py-3">WHOLESALE PLAN</th>
						<th class="px-4 py-3">STATUS</th>
						<th class="px-4 py-3 text-right">ACTIONS</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
					{#if loading && sps.length === 0}
						<tr>
							<td colspan="8" class="py-12 text-center text-xs text-gray-500">
								<div class="flex items-center justify-center gap-2">
									<div class="h-4 w-4 animate-spin rounded-full border-2 border-emerald-600 border-t-transparent"></div>
									<span>Loading S P resellers...</span>
								</div>
							</td>
						</tr>
					{:else if sps.length === 0}
						<tr>
							<td colspan="8" class="py-12 text-center text-xs text-gray-500">
								No resellers found matching the selected filters.
							</td>
						</tr>
					{:else}
						{#each sps as sp (sp.id)}
							<tr
								onclick={(e) => handleRowClick(sp, e)}
								class="hover:bg-gray-50/70 dark:hover:bg-gray-800/40 transition-colors cursor-pointer group"
							>
								<!-- Name -->
								<td class="px-4 py-3">
									<div class="font-semibold text-xs text-gray-900 dark:text-white group-hover:text-emerald-600 transition-colors">
										{sp.name}
									</div>
									{#if sp.business_name}
										<div class="text-[10px] text-gray-400">{sp.business_name}</div>
									{/if}
								</td>

								<!-- Email -->
								<td class="px-4 py-3 text-gray-600 dark:text-gray-300">{sp.email}</td>

								<!-- Phone -->
								<td class="px-4 py-3 font-mono text-gray-600 dark:text-gray-300">{sp.phone || '—'}</td>

								<!-- Wallet -->
								<td class="px-4 py-3 font-mono font-bold text-emerald-600 dark:text-emerald-400">
									৳{sp.balance.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
								</td>

								<!-- Customers Count -->
								<td class="px-4 py-3 text-center">
									<span class="inline-block rounded-full bg-gray-100 dark:bg-gray-800 px-2.5 py-0.5 text-xs font-bold text-gray-700 dark:text-gray-300">
										{sp.customer_count}
									</span>
								</td>

								<!-- Wholesale Plan -->
								<td class="px-4 py-3">
									{#if sp.wholesale_plan}
										<span class="rounded bg-sky-50 px-2 py-0.5 text-[11px] font-semibold text-sky-700 dark:bg-sky-950/40 dark:text-sky-300 border border-sky-200 dark:border-sky-800">
											{sp.wholesale_plan}
										</span>
									{:else}
										<span class="text-gray-400">No plan</span>
									{/if}
								</td>

								<!-- Status -->
								<td class="px-4 py-3">
									{#if sp.status === 'approved'}
										<span class="inline-flex items-center gap-1 rounded-full bg-emerald-50 px-2.5 py-0.5 text-[11px] font-semibold text-emerald-700 border border-emerald-200 dark:bg-emerald-950/40 dark:border-emerald-800 dark:text-emerald-400">
											<span class="h-1.5 w-1.5 rounded-full bg-emerald-500"></span>
											Approved
										</span>
									{:else}
										<span class="inline-flex items-center gap-1 rounded-full bg-amber-50 px-2.5 py-0.5 text-[11px] font-semibold text-amber-700 border border-amber-200 dark:bg-amber-950/40 dark:border-amber-800 dark:text-amber-400">
											<span class="h-1.5 w-1.5 rounded-full bg-amber-500"></span>
											{sp.status}
										</span>
									{/if}
								</td>

								<!-- Actions Column -->
								<td class="px-4 py-3 text-right no-row-nav">
									<div class="flex items-center justify-end gap-1.5 relative">
										<!-- Login As -->
										<button
											type="button"
											onclick={(e) => { e.stopPropagation(); handleImpersonate(sp); }}
											class="inline-flex items-center gap-1 px-2 py-1 rounded text-xs font-semibold text-gray-600 hover:text-emerald-600 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-800 transition"
										>
											<svg class="h-3.5 w-3.5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
												<path stroke-linecap="round" stroke-linejoin="round" d="M15.75 9V5.25A2.25 2.25 0 0 0 13.5 3h-6a2.25 2.25 0 0 0-2.25 2.25v13.5A2.25 2.25 0 0 0 7.5 21h6a2.25 2.25 0 0 0 2.25-2.25V15m3 0 3-3m0 0-3-3m3 3H9" />
											</svg>
											Login As
										</button>

										<!-- Three-Dot Button with green circular outline -->
										<div class="relative">
											<button
												type="button"
												onclick={(e) => {
													e.stopPropagation();
													activeMenuId = activeMenuId === sp.id ? null : sp.id;
												}}
												class="flex h-7 w-7 items-center justify-center rounded-full border border-emerald-500/80 text-emerald-600 hover:bg-emerald-50 dark:hover:bg-emerald-950/40 transition"
												title="Actions"
											>
												<svg class="h-4 w-4" fill="currentColor" viewBox="0 0 24 24">
													<path d="M12 6.75a1.5 1.5 0 1 1 0-3 1.5 1.5 0 0 1 0 3Zm0 6.75a1.5 1.5 0 1 1 0-3 1.5 1.5 0 0 1 0 3Zm0 6.75a1.5 1.5 0 1 1 0-3 1.5 1.5 0 0 1 0 3Z" />
												</svg>
											</button>

											{#if activeMenuId === sp.id}
												<div
													class="absolute right-0 top-full z-50 mt-1 w-48 rounded-xl border border-gray-200 bg-white py-1.5 shadow-xl dark:border-gray-700 dark:bg-gray-800 text-left"
												>
													<button
														type="button"
														onclick={() => openViewModal(sp)}
														class="flex w-full items-center gap-2 px-3 py-2 text-xs font-medium text-sky-600 hover:bg-sky-50 dark:text-sky-400 dark:hover:bg-sky-950/40"
													>
														<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke-width="1.75" stroke="currentColor">
															<path stroke-linecap="round" stroke-linejoin="round" d="M2.036 12.322a1.012 1.012 0 0 1 0-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178Z" />
														</svg>
														View S P
													</button>
													<button
														type="button"
														onclick={() => openDepositModal(sp)}
														class="flex w-full items-center gap-2 px-3 py-2 text-xs font-medium text-emerald-600 hover:bg-emerald-50 dark:text-emerald-400 dark:hover:bg-emerald-950/40"
													>
														<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke-width="1.75" stroke="currentColor">
															<path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
														</svg>
														Make Deposit
													</button>
													<button
														type="button"
														onclick={() => openEditModal(sp)}
														class="flex w-full items-center gap-2 px-3 py-2 text-xs font-medium text-amber-600 hover:bg-amber-50 dark:text-amber-400 dark:hover:bg-amber-950/40"
													>
														<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke-width="1.75" stroke="currentColor">
															<path stroke-linecap="round" stroke-linejoin="round" d="m16.862 4.487 1.687-1.688a1.875 1.875 0 1 1 2.652 2.652L6.832 19.82a4.5 4.5 0 0 1-1.897 1.13l-2.685.8.8-2.685a4.5 4.5 0 0 1 1.13-1.897L16.863 4.487Zm0 0L19.5 7.125" />
														</svg>
														Edit S P
													</button>
												</div>
											{/if}
										</div>
									</div>
								</td>
							</tr>
						{/each}
					{/if}
				</tbody>
			</table>
		</div>

		<!-- Pagination Footer -->
		<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 px-4 py-3 border-t border-gray-100 dark:border-gray-800 text-xs text-gray-600 dark:text-gray-400">
			<div>
				Showing {startRecord} to {endRecord} of {totalRecords} results
			</div>

			<div class="flex items-center gap-4">
				<div class="flex items-center gap-1.5">
					<span>Per page</span>
					<select
						bind:value={perPage}
						onchange={() => { currentPage = 1; void loadSPs(); }}
						class="rounded-lg border border-gray-200 bg-white px-2 py-1 text-xs dark:border-gray-700 dark:bg-gray-800 dark:text-white"
					>
						<option value={10}>10</option>
						<option value={25}>25</option>
						<option value={50}>50</option>
						<option value={100}>100</option>
					</select>
				</div>

				<div class="flex items-center gap-1">
					<button
						type="button"
						disabled={currentPage <= 1}
						onclick={() => { currentPage--; void loadSPs(); }}
						class="h-7 w-7 rounded-lg border border-gray-200 bg-white flex items-center justify-center hover:bg-gray-50 disabled:opacity-40 dark:border-gray-700 dark:bg-gray-800"
					>
						‹
					</button>
					{#each Array.from({ length: Math.min(5, totalPages) }, (_, i) => i + 1) as pageNum}
						<button
							type="button"
							onclick={() => { currentPage = pageNum; void loadSPs(); }}
							class="h-7 min-w-[28px] px-2 rounded-lg text-xs font-semibold {currentPage === pageNum ? 'bg-emerald-600 text-white' : 'border border-gray-200 bg-white hover:bg-gray-50 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-200'}"
						>
							{pageNum}
						</button>
					{/each}
					<button
						type="button"
						disabled={currentPage >= totalPages}
						onclick={() => { currentPage++; void loadSPs(); }}
						class="h-7 w-7 rounded-lg border border-gray-200 bg-white flex items-center justify-center hover:bg-gray-50 disabled:opacity-40 dark:border-gray-700 dark:bg-gray-800"
					>
						›
					</button>
				</div>
			</div>
		</div>
	</div>

	<!-- SP Summary & Total Calculation -->
	<div class="rounded-xl bg-white p-6 shadow-xs ring-1 ring-gray-950/5 dark:bg-gray-900 dark:ring-white/10">
		<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between pb-4 border-b border-gray-100 dark:border-gray-800 gap-2">
			<div>
				<h3 class="text-base font-semibold text-gray-950 dark:text-white">
					S P Reseller Summary &amp; Total Calculation
				</h3>
				<p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
					Aggregate financial and operational metrics for the selected reseller filters
				</p>
			</div>
			<div>
				<span class="inline-flex items-center px-2.5 py-1 rounded-md text-xs font-bold bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300">
					{summary.total_sps} Resellers
				</span>
			</div>
		</div>

		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mt-6">
			<!-- Total Wallet Reserve -->
			<div class="rounded-xl bg-gray-50/70 p-4 ring-1 ring-gray-950/5 dark:bg-gray-800/40 dark:ring-white/10">
				<div class="text-xs font-medium text-gray-500 dark:text-gray-400">Total Wallet Reserve</div>
				<div class="mt-2 text-2xl font-bold tracking-tight text-gray-950 dark:text-white font-mono">
					BDT {summary.total_balance.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
				</div>
			</div>

			<!-- Total Resellers -->
			<div class="rounded-xl bg-gray-50/70 p-4 ring-1 ring-gray-950/5 dark:bg-gray-800/40 dark:ring-white/10">
				<div class="text-xs font-medium text-gray-500 dark:text-gray-400">Total S P Accounts</div>
				<div class="mt-2 text-2xl font-bold tracking-tight text-gray-950 dark:text-white font-mono">
					{summary.total_sps}
				</div>
			</div>

			<!-- Managed Customers -->
			<div class="rounded-xl bg-gray-50/70 p-4 ring-1 ring-gray-950/5 dark:bg-gray-800/40 dark:ring-white/10">
				<div class="text-xs font-medium text-gray-500 dark:text-gray-400">Managed Customers</div>
				<div class="mt-2 text-2xl font-bold tracking-tight text-gray-950 dark:text-white font-mono">
					{summary.total_customers}
				</div>
			</div>

			<!-- Active Standing -->
			<div class="rounded-xl bg-gray-50/70 p-4 ring-1 ring-gray-950/5 dark:bg-gray-800/40 dark:ring-white/10">
				<div class="text-xs font-medium text-gray-500 dark:text-gray-400">Active Standing</div>
				<div class="mt-2 flex items-baseline gap-2">
					<span class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white font-mono">{summary.active_sps_count}</span>
					<span class="text-xs font-bold text-emerald-600 dark:text-emerald-400">({summary.active_rate}%)</span>
				</div>
			</div>
		</div>
	</div>
</div>

<!-- ============================================================
     VIEW SP MODAL
     ============================================================ -->
{#if viewModalOpen && selectedSP}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-gray-950/60 p-4 backdrop-blur-xs">
		<div class="w-full max-w-2xl rounded-2xl bg-white shadow-2xl ring-1 ring-gray-900/10 dark:bg-gray-900 dark:ring-white/10 max-h-[85vh] flex flex-col overflow-hidden">
			<div class="flex items-center justify-between border-b border-gray-100 px-6 py-4 dark:border-gray-800">
				<div>
					<h3 class="text-base font-bold text-gray-900 dark:text-white">Reseller Profile</h3>
					<p class="text-xs text-gray-500">ID #{selectedSP.id} • {selectedSP.name}</p>
				</div>
				<div class="flex items-center gap-2">
					<a href={`/superman/users/sps/${selectedSP.id}`} class="text-xs font-semibold text-emerald-600 hover:underline">
						Open Full Page →
					</a>
					<button type="button" onclick={() => (viewModalOpen = false)} class="rounded-lg p-1.5 text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800">✕</button>
				</div>
			</div>

			<div class="flex-1 overflow-y-auto p-6 space-y-4">
				{#if detailLoading}
					<div class="py-12 text-center text-xs text-gray-500">Loading reseller details...</div>
				{:else if currentDetail}
					<div class="grid grid-cols-2 gap-4 p-4 rounded-xl bg-gray-50 dark:bg-gray-800/40 text-xs">
						<div>
							<span class="text-gray-400 uppercase font-bold text-[10px]">Email</span>
							<p class="font-semibold text-gray-900 dark:text-white mt-0.5">{currentDetail.sp.email}</p>
						</div>
						<div>
							<span class="text-gray-400 uppercase font-bold text-[10px]">Phone</span>
							<p class="font-mono font-semibold text-gray-900 dark:text-white mt-0.5">{currentDetail.sp.phone || '—'}</p>
						</div>
						<div>
							<span class="text-gray-400 uppercase font-bold text-[10px]">Wallet Balance</span>
							<p class="font-mono font-bold text-emerald-600 dark:text-emerald-400 mt-0.5">৳{currentDetail.sp.balance.toFixed(2)}</p>
						</div>
						<div>
							<span class="text-gray-400 uppercase font-bold text-[10px]">Customers Count</span>
							<p class="font-bold text-gray-800 dark:text-gray-200 mt-0.5">{currentDetail.sp.customer_count}</p>
						</div>
					</div>

					{#if currentDetail.customers && currentDetail.customers.length > 0}
						<div>
							<h4 class="text-xs font-bold uppercase tracking-wider text-gray-700 dark:text-gray-300 mb-2">Customers managed by {currentDetail.sp.name}</h4>
							<div class="divide-y divide-gray-100 dark:divide-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 overflow-hidden">
								{#each currentDetail.customers as c}
									<div class="flex items-center justify-between p-2.5 text-xs hover:bg-gray-50 dark:hover:bg-gray-800/50">
										<div>
											<span class="font-semibold text-gray-900 dark:text-white">{c.name}</span>
											<span class="text-gray-400 font-mono ml-2">{c.phone}</span>
										</div>
										<span class="font-mono font-bold text-emerald-600">৳{c.balance.toFixed(2)}</span>
									</div>
								{/each}
							</div>
						</div>
					{/if}
				{/if}
			</div>

			<div class="border-t border-gray-100 dark:border-gray-800 px-6 py-3 flex items-center justify-between bg-gray-50/50 dark:bg-gray-850">
				<button
					type="button"
					onclick={() => { viewModalOpen = false; openDepositModal(selectedSP!); }}
					class="rounded-lg bg-emerald-600 px-3 py-1.5 text-xs font-semibold text-white hover:bg-emerald-700"
				>
					+ Deposit
				</button>
				<button type="button" onclick={() => (viewModalOpen = false)} class="rounded-lg bg-gray-200 dark:bg-gray-700 px-4 py-1.5 text-xs font-semibold text-gray-800 dark:text-gray-200">
					Close
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- ============================================================
     DEPOSIT MODAL
     ============================================================ -->
{#if depositModalOpen && selectedSP}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-gray-950/60 p-4 backdrop-blur-xs">
		<div class="w-full max-w-md rounded-2xl bg-white p-6 shadow-2xl ring-1 ring-gray-900/10 dark:bg-gray-900 dark:ring-white/10">
			<h3 class="text-base font-bold text-gray-900 dark:text-white">Adjust Reseller Wallet</h3>
			<p class="text-xs text-gray-500 mt-0.5">{selectedSP.name} (Current: ৳{selectedSP.balance.toFixed(2)})</p>

			{#if depositError}
				<div class="mt-3 rounded-lg border border-rose-200 bg-rose-50 p-2.5 text-xs text-rose-700 dark:bg-rose-950/50 dark:text-rose-300">
					{depositError}
				</div>
			{/if}

			<div class="mt-4 space-y-3">
				<div>
					<label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">
						Amount (BDT) <span class="text-gray-400 font-normal">— enter negative value to deduct</span>
					</label>
					<input
						type="number"
						step="0.01"
						bind:value={depositAmount}
						class="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm text-gray-900 shadow-2xs outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
					/>
				</div>

				<div>
					<label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Remarks / Note</label>
					<input
						type="text"
						bind:value={depositDescription}
						placeholder="e.g. Bank wire deposit ref #98765"
						class="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-xs text-gray-900 shadow-2xs outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
					/>
				</div>
			</div>

			<div class="mt-6 flex items-center justify-end gap-2">
				<button
					type="button"
					onclick={() => (depositModalOpen = false)}
					class="rounded-lg border border-gray-300 px-3.5 py-2 text-xs font-semibold text-gray-700 hover:bg-gray-50 dark:border-gray-700 dark:text-gray-300"
				>
					Cancel
				</button>
				<button
					type="button"
					disabled={depositSubmitting}
					onclick={submitDeposit}
					class="rounded-lg bg-emerald-600 px-4 py-2 text-xs font-bold text-white shadow-xs hover:bg-emerald-700 disabled:opacity-50"
				>
					{depositSubmitting ? 'Adjusting...' : 'Confirm Balance Adjustment'}
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- ============================================================
     EDIT SP MODAL
     ============================================================ -->
{#if editModalOpen}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-gray-950/60 p-4 backdrop-blur-xs">
		<div class="w-full max-w-lg rounded-2xl bg-white p-6 shadow-2xl ring-1 ring-gray-900/10 dark:bg-gray-900 dark:ring-white/10">
			<h3 class="text-base font-bold text-gray-900 dark:text-white">
				{editSPData.id ? 'Edit S P Reseller' : 'Create New S P'}
			</h3>

			<div class="mt-4 space-y-3">
				<div>
					<label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Company / S P Name</label>
					<input
						type="text"
						bind:value={editSPData.name}
						class="w-full rounded-lg border border-gray-300 px-3 py-1.5 text-xs dark:border-gray-700 dark:bg-gray-800 dark:text-white"
					/>
				</div>

				<div class="grid grid-cols-2 gap-3">
					<div>
						<label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Email</label>
						<input
							type="email"
							bind:value={editSPData.email}
							class="w-full rounded-lg border border-gray-300 px-3 py-1.5 text-xs dark:border-gray-700 dark:bg-gray-800 dark:text-white"
						/>
					</div>
					<div>
						<label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Phone</label>
						<input
							type="text"
							bind:value={editSPData.phone}
							class="w-full rounded-lg border border-gray-300 px-3 py-1.5 text-xs dark:border-gray-700 dark:bg-gray-800 dark:text-white"
						/>
					</div>
				</div>

				<div class="grid grid-cols-2 gap-3">
					<div>
						<label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Status</label>
						<select
							bind:value={editSPData.status}
							class="w-full rounded-lg border border-gray-300 px-3 py-1.5 text-xs dark:border-gray-700 dark:bg-gray-800 dark:text-white"
						>
							<option value="approved">Approved</option>
							<option value="pending">Pending</option>
							<option value="rejected">Rejected</option>
							<option value="banned">Banned</option>
						</select>
					</div>
					<div>
						<label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Wholesale Plan</label>
						<select
							bind:value={editSPData.tariff_plan_id}
							class="w-full rounded-lg border border-gray-300 px-3 py-1.5 text-xs dark:border-gray-700 dark:bg-gray-800 dark:text-white"
						>
							<option value={0}>No wholesale plan</option>
							{#each tariffPlans as tp}
								<option value={tp.id}>{tp.name}</option>
							{/each}
						</select>
					</div>
				</div>

				<div class="pt-2 border-t border-gray-100 dark:border-gray-800 space-y-2">
					<label class="flex items-center gap-2 cursor-pointer">
						<input type="checkbox" bind:checked={editSPData.api_access} class="rounded text-emerald-600" />
						<span class="text-xs font-semibold text-gray-800 dark:text-gray-200">API Access</span>
					</label>
					<label class="flex items-center gap-2 cursor-pointer">
						<input type="checkbox" bind:checked={editSPData.sms_auto_approved} class="rounded text-emerald-600" />
						<span class="text-xs font-semibold text-gray-800 dark:text-gray-200">Bulk SMS Auto-Approve</span>
					</label>
				</div>
			</div>

			<div class="mt-6 flex items-center justify-end gap-2">
				<button
					type="button"
					onclick={() => (editModalOpen = false)}
					class="rounded-lg border border-gray-300 px-3.5 py-2 text-xs font-semibold text-gray-700 hover:bg-gray-50 dark:border-gray-700 dark:text-gray-300"
				>
					Cancel
				</button>
				<button
					type="button"
					disabled={editSubmitting}
					onclick={submitEdit}
					class="rounded-lg bg-emerald-600 px-4 py-2 text-xs font-bold text-white shadow-xs hover:bg-emerald-700 disabled:opacity-50"
				>
					{editSubmitting ? 'Saving...' : 'Save Reseller'}
				</button>
			</div>
		</div>
	</div>
{/if}
