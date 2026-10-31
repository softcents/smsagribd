<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	const gatewayId = page.params.id;

	interface GatewayDetail {
		id: number;
		name: string;
		driver: string;
		route_code: string | null;
		priority: number;
		is_enabled: boolean;
		is_otp_enabled: boolean;
		supports_bulk: boolean;
		requests_per_second: number;
		timeout_seconds: number;
		dlr_support: boolean;
		cost_mode: string;
		dipping_required: boolean;
		has_third_party_dipping: boolean;
		health_status: string;
		created_at: string;
		updated_at: string;
		credentials?: Record<string, unknown>;
		billing_summary: {
			revenue: number;
			operator_cost: number;
			dipping_cost: number;
			total_cost: number;
			net_profit: number;
			profit_margin: number;
			dlr_delivery_percent: number | null;
			total_messages: number;
		};
	}

	interface DashboardCard {
		label: string;
		value: string;
		description: string;
		color: string;
		icon: string;
	}

	interface OperatorRate {
		id: number;
		operator_id: number;
		operator_name: string;
		route_code: string | null;
		country_name: string;
		masking_price: number | null;
		non_masking_price: number | null;
		dipping_charge: number | null;
		operator_cost: number;
		dipping_cost: number;
		created_at: string;
		updated_at: string;
	}

	interface OperatorOption {
		id: number;
		name: string;
		country_name?: string;
	}

	let gateway = $state<GatewayDetail | null>(null);
	let dashboardCards = $state<DashboardCard[]>([]);
	let rates = $state<OperatorRate[]>([]);
	let totalRates = $state(0);
	let currentPage = $state(1);
	let perPage = $state(10);
	let lastPage = $state(1);
	let searchRate = $state('');

	let loading = $state(true);
	let loadingRates = $state(false);
	let refreshingBalances = $state(false);
	let error = $state<string | null>(null);

	// Edit Gateway Modal
	let showEditGatewayModal = $state(false);
	let savingGateway = $state(false);
	let editGatewayError = $state<string | null>(null);
	let editName = $state('');
	let editDriver = $state('');
	let editPriority = $state(1);
	let editRps = $state(300);
	let editTimeout = $state(10);
	let editCostMode = $state('manual');
	let editDlrSupport = $state(true);
	let editDippingRequired = $state(true);
	let editHasThirdParty = $state(false);
	let editIsEnabled = $state(true);
	let editIsOtpEnabled = $state(true);
	let editSupportsBulk = $state(true);

	// Add/Edit Operator Rate Modal
	let showRateModal = $state(false);
	let rateModalMode = $state<'create' | 'edit'>('create');
	let editingRateId = $state<number | null>(null);
	let rateOperatorId = $state<number>(1);
	let rateMaskingPrice = $state<string>('');
	let rateNonMaskingPrice = $state<string>('');
	let rateDippingCharge = $state<string>('');
	let savingRate = $state(false);
	let rateModalError = $state<string | null>(null);
	let operatorsList = $state<OperatorOption[]>([]);

	async function loadGatewayData() {
		loading = true;
		error = null;
		try {
			const res = await supermanApi.get<{ data: GatewayDetail }>(`/api/v1/superman/gateways/${gatewayId}`);
			gateway = res.data;
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Failed to load gateway details.';
		} finally {
			loading = false;
		}
	}

	async function loadDashboardCards() {
		try {
			const res = await supermanApi.get<{ cards: DashboardCard[] }>(`/api/v1/superman/gateways/${gatewayId}/dashboard`);
			dashboardCards = res.cards ?? [];
		} catch {
			// keep empty / default
		}
	}

	async function refreshDashboardBalances() {
		refreshingBalances = true;
		try {
			await loadDashboardCards();
		} finally {
			refreshingBalances = false;
		}
	}

	async function loadRates() {
		loadingRates = true;
		try {
			let url = `/api/v1/superman/gateways/${gatewayId}/operator-rates?page=${currentPage}&per_page=${perPage}`;
			if (searchRate.trim()) {
				url += `&search=${encodeURIComponent(searchRate.trim())}`;
			}
			const res = await supermanApi.get<{
				data: OperatorRate[];
				meta: { total: number; current_page: number; per_page: number; last_page: number };
			}>(url);
			rates = res.data ?? [];
			totalRates = res.meta?.total ?? 0;
			currentPage = res.meta?.current_page ?? 1;
			lastPage = res.meta?.last_page ?? 1;
		} catch {
			// ignore
		} finally {
			loadingRates = false;
		}
	}

	async function loadOperatorsList() {
		try {
			const res = await supermanApi.get<{ data: OperatorOption[] }>('/api/v1/superman/operators');
			operatorsList = res.data ?? [];
		} catch {
			// ignore
		}
	}

	function openEditGatewayModal() {
		if (!gateway) return;
		editName = gateway.name;
		editDriver = gateway.driver;
		editPriority = gateway.priority;
		editRps = gateway.requests_per_second;
		editTimeout = gateway.timeout_seconds;
		editCostMode = gateway.cost_mode;
		editDlrSupport = gateway.dlr_support;
		editDippingRequired = gateway.dipping_required;
		editHasThirdParty = gateway.has_third_party_dipping;
		editIsEnabled = gateway.is_enabled;
		editIsOtpEnabled = gateway.is_otp_enabled;
		editSupportsBulk = gateway.supports_bulk;
		editGatewayError = null;
		showEditGatewayModal = true;
	}

	async function saveGatewayEdits() {
		savingGateway = true;
		editGatewayError = null;
		try {
			await supermanApi.put(`/api/v1/superman/gateways/${gatewayId}`, {
				name: editName,
				driver: editDriver,
				priority: Number(editPriority),
				requests_per_second: Number(editRps),
				timeout_seconds: Number(editTimeout),
				cost_mode: editCostMode,
				dlr_support: editDlrSupport,
				dipping_required: editDippingRequired,
				has_third_party_dipping: editHasThirdParty,
				is_enabled: editIsEnabled,
				is_otp_enabled: editIsOtpEnabled,
				supports_bulk: editSupportsBulk
			});
			showEditGatewayModal = false;
			await loadGatewayData();
		} catch (err: unknown) {
			editGatewayError = err instanceof Error ? err.message : 'Failed to update gateway.';
		} finally {
			savingGateway = false;
		}
	}

	function openCreateRateModal() {
		rateModalMode = 'create';
		editingRateId = null;
		rateOperatorId = operatorsList[0]?.id ?? 1;
		rateMaskingPrice = '';
		rateNonMaskingPrice = '0.1600';
		rateDippingCharge = '0.0575';
		rateModalError = null;
		showRateModal = true;
	}

	function openEditRateModal(r: OperatorRate) {
		rateModalMode = 'edit';
		editingRateId = r.id;
		rateOperatorId = r.operator_id;
		rateMaskingPrice = r.masking_price !== null ? String(r.masking_price) : '';
		rateNonMaskingPrice = r.non_masking_price !== null ? String(r.non_masking_price) : '';
		rateDippingCharge = r.dipping_charge !== null ? String(r.dipping_charge) : '';
		rateModalError = null;
		showRateModal = true;
	}

	async function saveRate() {
		savingRate = true;
		rateModalError = null;
		try {
			const payload: Record<string, unknown> = {
				operator_id: Number(rateOperatorId),
				masking_price: rateMaskingPrice !== '' ? Number(rateMaskingPrice) : null,
				non_masking_price: rateNonMaskingPrice !== '' ? Number(rateNonMaskingPrice) : null,
				dipping_charge: rateDippingCharge !== '' ? Number(rateDippingCharge) : null
			};

			if (rateModalMode === 'create') {
				await supermanApi.post(`/api/v1/superman/gateways/${gatewayId}/operator-rates`, payload);
			} else if (editingRateId) {
				await supermanApi.put(`/api/v1/superman/gateways/${gatewayId}/operator-rates/${editingRateId}`, payload);
			}
			showRateModal = false;
			await loadRates();
		} catch (err: unknown) {
			rateModalError = err instanceof Error ? err.message : 'Failed to save rate.';
		} finally {
			savingRate = false;
		}
	}

	async function deleteRate(r: OperatorRate) {
		if (!confirm(`Delete operator rate for "${r.operator_name}"?`)) return;
		try {
			await supermanApi.delete(`/api/v1/superman/gateways/${gatewayId}/operator-rates/${r.id}`);
			await loadRates();
		} catch (err: unknown) {
			alert(err instanceof Error ? err.message : 'Failed to delete rate.');
		}
	}

	onMount(() => {
		void loadGatewayData();
		void loadDashboardCards();
		void loadRates();
		void loadOperatorsList();
	});

	function formatTaka(val: number | null | undefined, decimals = 2): string {
		if (val === null || val === undefined) return '0.00 ৳';
		return `${val.toFixed(decimals)} ৳`;
	}

	const exportRateColumns: ExportColumn<OperatorRate>[] = [
		{ header: 'Operator', getValue: (r) => r.operator_name },
		{ header: 'Route Code', getValue: (r) => r.route_code || '—' },
		{ header: 'Country', getValue: (r) => r.country_name },
		{ header: 'Masking Price', getValue: (r) => (r.masking_price !== null ? `BDT ${r.masking_price.toFixed(4)}` : '—') },
		{ header: 'MNO Operator Rate', getValue: (r) => (r.non_masking_price !== null ? `BDT ${r.non_masking_price.toFixed(4)}` : '—') },
		{ header: 'Dipping Charge', getValue: (r) => (r.dipping_charge !== null ? `BDT ${r.dipping_charge.toFixed(4)}` : '—') },
		{ header: 'Operator Cost', getValue: (r) => `BDT ${r.operator_cost.toFixed(4)}` },
		{ header: 'Dipping Cost', getValue: (r) => `BDT ${r.dipping_cost.toFixed(4)}` }
	];
</script>

<svelte:head>
	<title>{gateway ? `View ${gateway.name}` : 'View Gateway'} — Superman Administrator</title>
</svelte:head>

<div class="space-y-6">
	<!-- Breadcrumb Navigation -->
	<nav class="flex items-center gap-2 text-xs text-gray-500 dark:text-gray-400">
		<a href="/superman/gateways" class="hover:text-gray-900 hover:underline dark:hover:text-white">Gateways</a>
		<svg class="h-3.5 w-3.5 text-gray-400" viewBox="0 0 20 20" fill="currentColor">
			<path fill-rule="evenodd" d="M7.21 14.77a.75.75 0 01.02-1.06L11.168 10 7.23 6.29a.75.75 0 111.04-1.08l4.5 4.25a.75.75 0 010 1.08l-4.5 4.25a.75.75 0 01-1.06-.02z" clip-rule="evenodd" />
		</svg>
		<span class="font-medium text-gray-700 dark:text-gray-300">{gateway?.name ?? `Gateway #${gatewayId}`}</span>
		<svg class="h-3.5 w-3.5 text-gray-400" viewBox="0 0 20 20" fill="currentColor">
			<path fill-rule="evenodd" d="M7.21 14.77a.75.75 0 01.02-1.06L11.168 10 7.23 6.29a.75.75 0 111.04-1.08l4.5 4.25a.75.75 0 010 1.08l-4.5 4.25a.75.75 0 01-1.06-.02z" clip-rule="evenodd" />
		</svg>
		<span class="text-gray-400">View</span>
	</nav>

	<!-- Page Heading & Actions -->
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">
				{gateway ? `View ${gateway.name}` : `View Gateway #${gatewayId}`}
			</h1>
		</div>
		<div class="flex items-center gap-2">
			<button
				onclick={openEditGatewayModal}
				class="inline-flex items-center gap-1.5 rounded-lg border border-amber-300 bg-white px-3.5 py-2 text-xs font-semibold text-amber-700 shadow-xs hover:bg-amber-50 dark:border-amber-700/50 dark:bg-amber-950/20 dark:text-amber-400 dark:hover:bg-amber-950/40"
			>
				<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m16.862 4.487 1.687-1.688a1.875 1.875 0 1 1 2.652 2.652L10.582 16.07a4.5 4.5 0 0 1-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 0 1 1.13-1.897l8.932-8.931Zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0 1 15.75 21H5.25A2.25 2.25 0 0 1 3 18.75V8.25A2.25 2.25 0 0 1 5.25 6H10" />
				</svg>
				Edit
			</button>
		</div>
	</div>

	{#if loading}
		<div class="flex items-center justify-center rounded-2xl border border-gray-200 bg-white py-24 text-xs text-gray-500 shadow-xs dark:border-gray-800 dark:bg-gray-900">
			<svg class="mr-2 h-5 w-5 animate-spin" fill="none" viewBox="0 0 24 24">
				<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
				<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
			</svg>
			Loading gateway details...
		</div>
	{:else if error || !gateway}
		<div class="rounded-2xl border border-rose-200 bg-rose-50/50 p-8 text-center text-xs text-rose-700 dark:border-rose-900/50 dark:bg-rose-950/20 dark:text-rose-400">
			{error || 'Gateway not found.'}
			<div class="mt-4">
				<a href="/superman/gateways" class="font-semibold underline">Back to Gateways</a>
			</div>
		</div>
	{:else}
		<!-- Section 1: Gateway Dashboard -->
		<div class="rounded-2xl border border-gray-200 bg-white p-5 shadow-xs dark:border-gray-800 dark:bg-gray-900">
			<div class="flex items-center justify-between border-b border-gray-100 pb-3 dark:border-gray-800">
				<h2 class="text-sm font-bold text-gray-950 dark:text-white">Gateway Dashboard</h2>
				<button
					onclick={refreshDashboardBalances}
					disabled={refreshingBalances}
					class="inline-flex items-center gap-1.5 text-xs font-medium text-gray-500 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white"
				>
					<svg class={`h-3.5 w-3.5 ${refreshingBalances ? 'animate-spin' : ''}`} fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
					</svg>
					Refresh
				</button>
			</div>

			<div class="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
				{#if dashboardCards.length > 0}
					{#each dashboardCards as card}
						<div class="flex items-center gap-3.5 rounded-xl border border-gray-100 bg-gray-50/60 p-4 dark:border-gray-800/80 dark:bg-gray-800/40">
							<div class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-emerald-50 text-emerald-600 dark:bg-emerald-950/50 dark:text-emerald-400">
								{#if card.icon.includes('banknotes')}
									<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M2.25 18.75a60.07 60.07 0 0 1 15.797 2.101c.727.198 1.453-.342 1.453-1.096V18.75M3.75 4.5v.75A.75.75 0 0 1 3 6H2.25m0 0v8.25m0 0h.75a.75.75 0 0 1 .75.75v.75m0 0H3m0 0h.75m0 0h16.5m0 0h.75m0 0v-.75a.75.75 0 0 1 .75-.75h.75M21.75 15V6.75m0 0h-.75a.75.75 0 0 1-.75-.75V5.25m0 0H21m0 0h-.75M3 12h18" />
									</svg>
								{:else}
									<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M10.5 1.5H8.25A2.25 2.25 0 0 0 6 3.75v16.5a2.25 2.25 0 0 0 2.25 2.25h7.5A2.25 2.25 0 0 0 18 20.25V3.75a2.25 2.25 0 0 0-2.25-2.25H13.5m-3 0V3h3V1.5m-3 0h3m-3 18.75h3" />
									</svg>
								{/if}
							</div>
							<div>
								<div class="text-[11px] font-medium text-gray-500 uppercase tracking-wider dark:text-gray-400">{card.label}</div>
								<div class="text-xl font-bold text-gray-900 dark:text-white">{card.value}</div>
								<div class="text-[11px] text-gray-400">{card.description}</div>
							</div>
						</div>
					{/each}
				{:else}
					<div class="flex items-center gap-3.5 rounded-xl border border-gray-100 bg-gray-50/60 p-4 dark:border-gray-800/80 dark:bg-gray-800/40">
						<div class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-emerald-50 text-emerald-600 dark:bg-emerald-950/50 dark:text-emerald-400">
							<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M2.25 18.75a60.07 60.07 0 0 1 15.797 2.101c.727.198 1.453-.342 1.453-1.096V18.75M3.75 4.5v.75A.75.75 0 0 1 3 6H2.25m0 0v8.25m0 0h.75a.75.75 0 0 1 .75.75v.75m0 0H3m0 0h.75m0 0h16.5m0 0h.75m0 0v-.75a.75.75 0 0 1 .75-.75h.75M21.75 15V6.75m0 0h-.75a.75.75 0 0 1-.75-.75V5.25m0 0H21m0 0h-.75M3 12h18" />
							</svg>
						</div>
						<div>
							<div class="text-[11px] font-medium text-gray-500 uppercase tracking-wider dark:text-gray-400">Wallet Balance</div>
							<div class="text-xl font-bold text-gray-900 dark:text-white">—</div>
							<div class="text-[11px] text-gray-400">Available balance at CP end</div>
						</div>
					</div>
					<div class="flex items-center gap-3.5 rounded-xl border border-gray-100 bg-gray-50/60 p-4 dark:border-gray-800/80 dark:bg-gray-800/40">
						<div class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-emerald-50 text-emerald-600 dark:bg-emerald-950/50 dark:text-emerald-400">
							<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M10.5 1.5H8.25A2.25 2.25 0 0 0 6 3.75v16.5a2.25 2.25 0 0 0 2.25 2.25h7.5A2.25 2.25 0 0 0 18 20.25V3.75a2.25 2.25 0 0 0-2.25-2.25H13.5m-3 0V3h3V1.5m-3 0h3m-3 18.75h3" />
							</svg>
						</div>
						<div>
							<div class="text-[11px] font-medium text-gray-500 uppercase tracking-wider dark:text-gray-400">{gateway.route_code || 'Carrier'} Balance</div>
							<div class="text-xl font-bold text-gray-900 dark:text-white">—</div>
							<div class="text-[11px] text-gray-400">{gateway.route_code || 'Carrier'} credit balance</div>
						</div>
					</div>
				{/if}
			</div>
		</div>

		<!-- Section 2: Billing & Profit Summary (Current Month) -->
		<div class="rounded-2xl border border-gray-200 bg-white p-5 shadow-xs dark:border-gray-800 dark:bg-gray-900">
			<div class="border-b border-gray-100 pb-3 dark:border-gray-800">
				<h2 class="text-sm font-bold text-gray-950 dark:text-white">Billing & Profit Summary (Current Month)</h2>
				<p class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">
					Authoritative cost resolution and tariff-based revenue calculation
				</p>
			</div>

			<!-- Row 1: Revenue, Operator Cost, Dipping Cost, Total Cost -->
			<div class="mt-4 grid grid-cols-2 gap-4 lg:grid-cols-4">
				<div class="rounded-xl border border-gray-100 bg-gray-50/50 p-4 dark:border-gray-800/80 dark:bg-gray-800/30">
					<div class="text-[11px] font-medium text-gray-500 uppercase tracking-wider">Revenue (Tariff)</div>
					<div class="mt-1.5">
						<span class="inline-flex items-center rounded-md bg-emerald-50 px-2.5 py-1 text-xs font-bold text-emerald-700 dark:bg-emerald-950/50 dark:text-emerald-400">
							{formatTaka(gateway.billing_summary.revenue)}
						</span>
					</div>
				</div>

				<div class="rounded-xl border border-gray-100 bg-gray-50/50 p-4 dark:border-gray-800/80 dark:bg-gray-800/30">
					<div class="text-[11px] font-medium text-gray-500 uppercase tracking-wider">Operator Cost</div>
					<div class="mt-1.5">
						<span class="inline-flex items-center rounded-md bg-amber-50 px-2.5 py-1 text-xs font-bold text-amber-700 dark:bg-amber-950/50 dark:text-amber-400">
							{formatTaka(gateway.billing_summary.operator_cost)}
						</span>
					</div>
				</div>

				<div class="rounded-xl border border-gray-100 bg-gray-50/50 p-4 dark:border-gray-800/80 dark:bg-gray-800/30">
					<div class="text-[11px] font-medium text-gray-500 uppercase tracking-wider">Dipping Cost</div>
					<div class="mt-1.5">
						<span class="inline-flex items-center rounded-md bg-cyan-50 px-2.5 py-1 text-xs font-bold text-cyan-700 dark:bg-cyan-950/50 dark:text-cyan-400">
							{formatTaka(gateway.billing_summary.dipping_cost)}
						</span>
					</div>
				</div>

				<div class="rounded-xl border border-gray-100 bg-gray-50/50 p-4 dark:border-gray-800/80 dark:bg-gray-800/30">
					<div class="text-[11px] font-medium text-gray-500 uppercase tracking-wider">Total Cost</div>
					<div class="mt-1.5 text-sm font-bold text-gray-900 dark:text-white">
						{formatTaka(gateway.billing_summary.total_cost)}
					</div>
				</div>
			</div>

			<!-- Row 2: Net Profit, Profit Margin, DLR Delivery %, Total Messages -->
			<div class="mt-4 grid grid-cols-2 gap-4 lg:grid-cols-4">
				<div class="rounded-xl border border-gray-100 bg-gray-50/50 p-4 dark:border-gray-800/80 dark:bg-gray-800/30">
					<div class="text-[11px] font-medium text-gray-500 uppercase tracking-wider">Net Profit</div>
					<div class="mt-1.5">
						<span class={`inline-flex items-center rounded-md px-2.5 py-1 text-xs font-bold ${
							gateway.billing_summary.net_profit >= 0
								? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-950/50 dark:text-emerald-400'
								: 'bg-rose-50 text-rose-700 dark:bg-rose-950/50 dark:text-rose-400'
						}`}>
							{formatTaka(gateway.billing_summary.net_profit)}
						</span>
					</div>
				</div>

				<div class="rounded-xl border border-gray-100 bg-gray-50/50 p-4 dark:border-gray-800/80 dark:bg-gray-800/30">
					<div class="text-[11px] font-medium text-gray-500 uppercase tracking-wider">Profit Margin</div>
					<div class="mt-1.5 text-sm font-semibold text-gray-900 dark:text-white">
						{gateway.billing_summary.profit_margin.toFixed(1)}%
					</div>
				</div>

				<div class="rounded-xl border border-gray-100 bg-gray-50/50 p-4 dark:border-gray-800/80 dark:bg-gray-800/30">
					<div class="text-[11px] font-medium text-gray-500 uppercase tracking-wider">DLR Delivery %</div>
					<div class="mt-1.5">
						{#if !gateway.dlr_support}
							<span class="inline-flex items-center rounded-md bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-600 dark:bg-gray-800 dark:text-gray-400">
								N/A (Disabled)
							</span>
						{:else if gateway.billing_summary.dlr_delivery_percent !== null}
							<span class="text-sm font-semibold text-emerald-600 dark:text-emerald-400">
								{gateway.billing_summary.dlr_delivery_percent.toFixed(1)}%
							</span>
						{:else}
							<span class="text-gray-400">—</span>
						{/if}
					</div>
				</div>

				<div class="rounded-xl border border-gray-100 bg-gray-50/50 p-4 dark:border-gray-800/80 dark:bg-gray-800/30">
					<div class="text-[11px] font-medium text-gray-500 uppercase tracking-wider">Total Messages</div>
					<div class="mt-1.5 text-sm font-semibold text-gray-900 dark:text-white">
						{gateway.billing_summary.total_messages.toLocaleString()}
					</div>
				</div>
			</div>
		</div>

		<!-- Section 3: Two-Column Section (Details + Billing & Cost Configuration) -->
		<div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
			<!-- Details Card -->
			<div class="rounded-2xl border border-gray-200 bg-white p-5 shadow-xs dark:border-gray-800 dark:bg-gray-900">
				<h2 class="border-b border-gray-100 pb-3 text-sm font-bold text-gray-950 dark:text-white">Details</h2>

				<div class="mt-4 grid grid-cols-2 gap-y-4 text-xs">
					<div>
						<div class="text-[11px] font-medium text-gray-400 uppercase">Name</div>
						<div class="mt-1 text-sm font-bold text-gray-950 dark:text-white">{gateway.name}</div>
					</div>

					<div>
						<div class="text-[11px] font-medium text-gray-400 uppercase">Route Code</div>
						<div class="mt-1">
							{#if gateway.route_code}
								<span class="inline-flex items-center rounded-md bg-teal-50 px-2 py-0.5 text-xs font-semibold text-teal-700 dark:bg-teal-950/40 dark:text-teal-300">
									{gateway.route_code}
								</span>
							{:else}
								<span class="text-gray-400">—</span>
							{/if}
						</div>
					</div>

					<div>
						<div class="text-[11px] font-medium text-gray-400 uppercase">Driver</div>
						<div class="mt-1">
							<span class="inline-flex items-center gap-1 rounded-md bg-emerald-50 px-2 py-0.5 text-xs font-semibold text-emerald-700 capitalize dark:bg-emerald-950/40 dark:text-emerald-300">
								{gateway.driver}
							</span>
						</div>
					</div>

					<div>
						<div class="text-[11px] font-medium text-gray-400 uppercase">RPS</div>
						<div class="mt-1 font-semibold text-gray-900 dark:text-white">{gateway.requests_per_second}</div>
					</div>

					<div>
						<div class="text-[11px] font-medium text-gray-400 uppercase">Enabled</div>
						<div class="mt-1">
							{#if gateway.is_enabled}
								<span class="inline-flex h-5 w-5 items-center justify-center rounded-full bg-emerald-100 text-emerald-600 dark:bg-emerald-950/50 dark:text-emerald-400">
									<svg class="h-3.5 w-3.5" viewBox="0 0 20 20" fill="currentColor">
										<path fill-rule="evenodd" d="M16.704 4.153a.75.75 0 0 1 .143 1.052l-8 10.5a.75.75 0 0 1-1.127.075l-4.5-4.5a.75.75 0 0 1 1.06-1.06l3.894 3.893 7.48-9.817a.75.75 0 0 1 1.05-.143Z" clip-rule="evenodd" />
									</svg>
								</span>
							{:else}
								<span class="inline-flex h-5 w-5 items-center justify-center rounded-full bg-rose-100 text-rose-600 dark:bg-rose-950/50 dark:text-rose-400">
									<svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
									</svg>
								</span>
							{/if}
						</div>
					</div>

					<div>
						<div class="text-[11px] font-medium text-gray-400 uppercase">Supports Promotional</div>
						<div class="mt-1">
							{#if gateway.supports_bulk}
								<span class="inline-flex h-5 w-5 items-center justify-center rounded-full bg-emerald-100 text-emerald-600 dark:bg-emerald-950/50 dark:text-emerald-400">
									<svg class="h-3.5 w-3.5" viewBox="0 0 20 20" fill="currentColor">
										<path fill-rule="evenodd" d="M16.704 4.153a.75.75 0 0 1 .143 1.052l-8 10.5a.75.75 0 0 1-1.127.075l-4.5-4.5a.75.75 0 0 1 1.06-1.06l3.894 3.893 7.48-9.817a.75.75 0 0 1 1.05-.143Z" clip-rule="evenodd" />
									</svg>
								</span>
							{:else}
								<span class="inline-flex h-5 w-5 items-center justify-center rounded-full bg-rose-100 text-rose-600 dark:bg-rose-950/50 dark:text-rose-400">
									<svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
									</svg>
								</span>
							{/if}
						</div>
					</div>

					<div>
						<div class="text-[11px] font-medium text-gray-400 uppercase">Created At</div>
						<div class="mt-1 font-medium text-gray-700 dark:text-gray-300">{gateway.created_at || '—'}</div>
					</div>

					<div>
						<div class="text-[11px] font-medium text-gray-400 uppercase">Updated At</div>
						<div class="mt-1 font-medium text-gray-700 dark:text-gray-300">{gateway.updated_at || '—'}</div>
					</div>
				</div>
			</div>

			<!-- Billing & Cost Configuration Card -->
			<div class="rounded-2xl border border-gray-200 bg-white p-5 shadow-xs dark:border-gray-800 dark:bg-gray-900">
				<h2 class="border-b border-gray-100 pb-3 text-sm font-bold text-gray-950 dark:text-white">
					Billing & Cost Configuration
				</h2>

				<div class="mt-4 grid grid-cols-2 gap-4 sm:grid-cols-4">
					<div>
						<div class="text-[11px] font-medium text-gray-400 uppercase">DLR Support</div>
						<div class="mt-2">
							{#if gateway.dlr_support}
								<span class="inline-flex h-5 w-5 items-center justify-center rounded-full bg-emerald-100 text-emerald-600 dark:bg-emerald-950/50 dark:text-emerald-400">
									<svg class="h-3.5 w-3.5" viewBox="0 0 20 20" fill="currentColor">
										<path fill-rule="evenodd" d="M16.704 4.153a.75.75 0 0 1 .143 1.052l-8 10.5a.75.75 0 0 1-1.127.075l-4.5-4.5a.75.75 0 0 1 1.06-1.06l3.894 3.893 7.48-9.817a.75.75 0 0 1 1.05-.143Z" clip-rule="evenodd" />
									</svg>
								</span>
							{:else}
								<span class="inline-flex h-5 w-5 items-center justify-center rounded-full bg-rose-100 text-rose-600 dark:bg-rose-950/50 dark:text-rose-400">
									<svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
									</svg>
								</span>
							{/if}
						</div>
					</div>

					<div>
						<div class="text-[11px] font-medium text-gray-400 uppercase">Cost Mode</div>
						<div class="mt-2">
							<span class="inline-flex items-center rounded-md bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-700 capitalize dark:bg-gray-800 dark:text-gray-300">
								{gateway.cost_mode}
							</span>
						</div>
					</div>

					<div>
						<div class="text-[11px] font-medium text-gray-400 uppercase">Mandatory Dipping</div>
						<div class="mt-2">
							{#if gateway.dipping_required}
								<span class="inline-flex h-5 w-5 items-center justify-center rounded-full bg-emerald-100 text-emerald-600 dark:bg-emerald-950/50 dark:text-emerald-400">
									<svg class="h-3.5 w-3.5" viewBox="0 0 20 20" fill="currentColor">
										<path fill-rule="evenodd" d="M16.704 4.153a.75.75 0 0 1 .143 1.052l-8 10.5a.75.75 0 0 1-1.127.075l-4.5-4.5a.75.75 0 0 1 1.06-1.06l3.894 3.893 7.48-9.817a.75.75 0 0 1 1.05-.143Z" clip-rule="evenodd" />
									</svg>
								</span>
							{:else}
								<span class="inline-flex h-5 w-5 items-center justify-center rounded-full bg-rose-100 text-rose-600 dark:bg-rose-950/50 dark:text-rose-400">
									<svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
									</svg>
								</span>
							{/if}
						</div>
					</div>

					<div>
						<div class="text-[11px] font-medium text-gray-400 uppercase">3rd-Party Dipping</div>
						<div class="mt-2">
							{#if gateway.has_third_party_dipping}
								<span class="inline-flex h-5 w-5 items-center justify-center rounded-full bg-emerald-100 text-emerald-600 dark:bg-emerald-950/50 dark:text-emerald-400">
									<svg class="h-3.5 w-3.5" viewBox="0 0 20 20" fill="currentColor">
										<path fill-rule="evenodd" d="M16.704 4.153a.75.75 0 0 1 .143 1.052l-8 10.5a.75.75 0 0 1-1.127.075l-4.5-4.5a.75.75 0 0 1 1.06-1.06l3.894 3.893 7.48-9.817a.75.75 0 0 1 1.05-.143Z" clip-rule="evenodd" />
									</svg>
								</span>
							{:else}
								<span class="inline-flex h-5 w-5 items-center justify-center rounded-full bg-rose-100 text-rose-600 dark:bg-rose-950/50 dark:text-rose-400">
									<svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
									</svg>
								</span>
							{/if}
						</div>
					</div>
				</div>
			</div>
		</div>

		<!-- Section 4: Operator Rates Table -->
		<div class="rounded-2xl border border-gray-200 bg-white p-5 shadow-xs dark:border-gray-800 dark:bg-gray-900">
			<div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between border-b border-gray-100 pb-4 dark:border-gray-800">
				<div>
					<h2 class="text-sm font-bold text-gray-950 dark:text-white">Operator Rates</h2>
				</div>
				<div class="flex items-center gap-2">
					<div class="relative w-64">
						<input
							type="search"
							bind:value={searchRate}
							oninput={() => { currentPage = 1; void loadRates(); }}
							placeholder="Search..."
							class="w-full rounded-lg border border-gray-200 bg-gray-50/70 py-1.5 pl-3 pr-3 text-xs outline-none focus:border-emerald-500 focus:bg-white dark:border-gray-700 dark:bg-gray-800 dark:text-white"
						/>
					</div>
					<ExportDropdown
						title={`${gateway.name} - Operator Rates`}
						columns={exportRateColumns}
						getData={() => rates}
						disabled={rates.length === 0}
					/>
					<button
						onclick={openCreateRateModal}
						class="inline-flex items-center gap-1 rounded-lg bg-emerald-600 px-3 py-1.5 text-xs font-semibold text-white shadow-xs hover:bg-emerald-700"
					>
						<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
						</svg>
						Add Rate
					</button>
				</div>
			</div>

			<div class="mt-4 overflow-x-auto">
				<table class="w-full text-left text-xs">
					<thead>
						<tr class="border-b border-gray-100 bg-gray-50/50 text-[10px] font-semibold uppercase tracking-wider text-gray-500 dark:border-gray-800 dark:bg-gray-800/50 dark:text-gray-400">
							<th class="px-4 py-3">Operator</th>
							<th class="px-4 py-3">Route Code</th>
							<th class="px-4 py-3">Country</th>
							<th class="px-4 py-3">Masking Price</th>
							<th class="px-4 py-3">MNO Operator Rate</th>
							<th class="px-4 py-3">Dipping Charge</th>
							<th class="px-4 py-3">Operator Cost</th>
							<th class="px-4 py-3">Dipping Cost</th>
							<th class="px-4 py-3 text-right">Actions</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
						{#if loadingRates}
							<tr>
								<td colspan="9" class="py-12 text-center text-xs text-gray-400">Loading rates...</td>
							</tr>
						{:else if rates.length === 0}
							<tr>
								<td colspan="9" class="py-12 text-center text-xs text-gray-400">No operator rates configured for this gateway.</td>
							</tr>
						{:else}
							{#each rates as r}
								<tr class="transition-colors hover:bg-gray-50/50 dark:hover:bg-gray-800/30">
									<td class="px-4 py-3.5 font-semibold text-gray-950 dark:text-white">
										{r.operator_name}
									</td>
									<td class="px-4 py-3.5">
										{#if r.route_code}
											<span class="inline-flex items-center rounded-md bg-teal-50 px-2 py-0.5 text-[11px] font-semibold text-teal-700 dark:bg-teal-950/40 dark:text-teal-300">
												{r.route_code}
											</span>
										{:else}
											<span class="text-gray-400">—</span>
										{/if}
									</td>
									<td class="px-4 py-3.5 text-gray-600 dark:text-gray-300">
										{r.country_name}
									</td>
									<td class="px-4 py-3.5 text-gray-700 dark:text-gray-300">
										{r.masking_price !== null ? `৳${r.masking_price.toFixed(4)}` : '—'}
									</td>
									<td class="px-4 py-3.5 font-medium text-gray-900 dark:text-white">
										{r.non_masking_price !== null ? `৳${r.non_masking_price.toFixed(4)}` : '—'}
									</td>
									<td class="px-4 py-3.5 font-medium text-gray-900 dark:text-white">
										{r.dipping_charge !== null ? `৳${r.dipping_charge.toFixed(4)}` : '—'}
									</td>
									<td class="px-4 py-3.5">
										<span class="inline-flex items-center rounded-md bg-amber-50 px-2 py-0.5 text-[11px] font-semibold text-amber-700 dark:bg-amber-950/40 dark:text-amber-400">
											{r.operator_cost.toFixed(4)} ৳
										</span>
									</td>
									<td class="px-4 py-3.5">
										<span class="inline-flex items-center rounded-md bg-cyan-50 px-2 py-0.5 text-[11px] font-semibold text-cyan-700 dark:bg-cyan-950/40 dark:text-cyan-400">
											{r.dipping_cost.toFixed(4)} ৳
										</span>
									</td>
									<td class="px-4 py-3.5 text-right">
										<div class="flex items-center justify-end gap-2">
											<button
												onclick={() => openEditRateModal(r)}
												class="rounded p-1 text-gray-500 hover:bg-gray-100 hover:text-gray-800 dark:text-gray-400 dark:hover:bg-gray-800 dark:hover:text-white"
												title="Edit Rate"
											>
												<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="m16.862 4.487 1.687-1.688a1.875 1.875 0 1 1 2.652 2.652L10.582 16.07a4.5 4.5 0 0 1-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 0 1 1.13-1.897l8.932-8.931Zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0 1 15.75 21H5.25A2.25 2.25 0 0 1 3 18.75V8.25A2.25 2.25 0 0 1 5.25 6H10" />
												</svg>
											</button>
											<button
												onclick={() => void deleteRate(r)}
												class="rounded p-1 text-rose-500 hover:bg-rose-50 dark:hover:bg-rose-950/30"
												title="Delete Rate"
											>
												<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="m14.74 9-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 0 0-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 0 0-7.5 0" />
												</svg>
											</button>
										</div>
									</td>
								</tr>
							{/each}
						{/if}
					</tbody>
				</table>
			</div>

			<!-- Pagination Footer -->
			<div class="mt-4 flex flex-col items-center justify-between gap-3 border-t border-gray-100 pt-3 text-xs text-gray-500 sm:flex-row dark:border-gray-800 dark:text-gray-400">
				<div>
					Showing {totalRates === 0 ? 0 : (currentPage - 1) * perPage + 1} to {Math.min(currentPage * perPage, totalRates)} of {totalRates} results
				</div>
				<div class="flex items-center gap-2">
					<span class="text-xs">Per page:</span>
					<select
						bind:value={perPage}
						onchange={() => { currentPage = 1; void loadRates(); }}
						class="rounded-lg border border-gray-200 bg-white px-2 py-1 text-xs text-gray-700 outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-300"
					>
						<option value={5}>5</option>
						<option value={10}>10</option>
						<option value={25}>25</option>
						<option value={50}>50</option>
					</select>
					<button
						disabled={currentPage <= 1}
						onclick={() => { currentPage--; void loadRates(); }}
						class="rounded-lg border border-gray-200 px-2.5 py-1 text-xs font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-40 dark:border-gray-700 dark:text-gray-300 dark:hover:bg-gray-800"
					>
						Previous
					</button>
					<span class="text-xs font-medium text-gray-700 dark:text-gray-300">{currentPage} / {lastPage}</span>
					<button
						disabled={currentPage >= lastPage}
						onclick={() => { currentPage++; void loadRates(); }}
						class="rounded-lg border border-gray-200 px-2.5 py-1 text-xs font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-40 dark:border-gray-700 dark:text-gray-300 dark:hover:bg-gray-800"
					>
						Next
					</button>
				</div>
			</div>
		</div>
	{/if}
</div>

<!-- Edit Gateway Modal -->
{#if showEditGatewayModal}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-gray-900/50 p-4 backdrop-blur-xs">
		<div class="w-full max-w-lg rounded-2xl border border-gray-200 bg-white p-6 shadow-xl dark:border-gray-800 dark:bg-gray-900">
			<div class="flex items-center justify-between border-b border-gray-100 pb-3 dark:border-gray-800">
				<h3 class="text-base font-bold text-gray-900 dark:text-white">Edit Gateway: {editName}</h3>
				<button onclick={() => (showEditGatewayModal = false)} class="rounded-lg p-1 text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800">
					<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
				</button>
			</div>

			{#if editGatewayError}
				<div class="mt-3 rounded-lg bg-rose-50 p-3 text-xs text-rose-700 dark:bg-rose-950/40 dark:text-rose-300">{editGatewayError}</div>
			{/if}

			<form onsubmit={(e) => { e.preventDefault(); void saveGatewayEdits(); }} class="mt-4 space-y-4">
				<div>
					<label class="block text-xs font-semibold text-gray-700 dark:text-gray-300">Name</label>
					<input
						type="text"
						bind:value={editName}
						required
						class="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2 text-xs outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
					/>
				</div>

				<div class="grid grid-cols-2 gap-3">
					<div>
						<label class="block text-xs font-semibold text-gray-700 dark:text-gray-300">Driver</label>
						<select
							bind:value={editDriver}
							class="mt-1 w-full rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
						>
							<option value="infozillion">Infozillion</option>
							<option value="smpp">SMPP Protocol</option>
							<option value="http">Generic HTTP / REST</option>
							<option value="infobip">Infobip</option>
							<option value="twilio">Twilio</option>
						</select>
					</div>
					<div>
						<label class="block text-xs font-semibold text-gray-700 dark:text-gray-300">Routing Priority</label>
						<input
							type="number"
							bind:value={editPriority}
							min="1"
							max="99"
							class="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2 text-xs outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
						/>
					</div>
				</div>

				<div class="grid grid-cols-2 gap-3">
					<div>
						<label class="block text-xs font-semibold text-gray-700 dark:text-gray-300">Max Throughput (RPS)</label>
						<input
							type="number"
							bind:value={editRps}
							min="1"
							max="1000"
							class="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2 text-xs outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
						/>
					</div>
					<div>
						<label class="block text-xs font-semibold text-gray-700 dark:text-gray-300">Timeout (Seconds)</label>
						<input
							type="number"
							bind:value={editTimeout}
							min="1"
							max="60"
							class="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2 text-xs outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
						/>
					</div>
				</div>

				<div class="space-y-2 border-t border-gray-100 pt-3 dark:border-gray-800">
					<label class="flex items-center gap-2 cursor-pointer">
						<input type="checkbox" bind:checked={editIsEnabled} class="h-4 w-4 rounded border-gray-300 text-emerald-600 focus:ring-emerald-500" />
						<span class="text-xs text-gray-700 dark:text-gray-300">Enable this gateway for active dispatches</span>
					</label>
					<label class="flex items-center gap-2 cursor-pointer">
						<input type="checkbox" bind:checked={editIsOtpEnabled} class="h-4 w-4 rounded border-gray-300 text-emerald-600 focus:ring-emerald-500" />
						<span class="text-xs text-gray-700 dark:text-gray-300">Include in high-priority OTP pool</span>
					</label>
					<label class="flex items-center gap-2 cursor-pointer">
						<input type="checkbox" bind:checked={editSupportsBulk} class="h-4 w-4 rounded border-gray-300 text-emerald-600 focus:ring-emerald-500" />
						<span class="text-xs text-gray-700 dark:text-gray-300">Supports Promotional Messages</span>
					</label>
					<label class="flex items-center gap-2 cursor-pointer">
						<input type="checkbox" bind:checked={editDlrSupport} class="h-4 w-4 rounded border-gray-300 text-emerald-600 focus:ring-emerald-500" />
						<span class="text-xs text-gray-700 dark:text-gray-300">Support upstream DLR receipts</span>
					</label>
					<label class="flex items-center gap-2 cursor-pointer">
						<input type="checkbox" bind:checked={editDippingRequired} class="h-4 w-4 rounded border-gray-300 text-emerald-600 focus:ring-emerald-500" />
						<span class="text-xs text-gray-700 dark:text-gray-300">Mandatory Number Dipping</span>
					</label>
				</div>

				<div class="flex justify-end gap-2 border-t border-gray-100 pt-3 dark:border-gray-800">
					<button
						type="button"
						onclick={() => (showEditGatewayModal = false)}
						class="rounded-lg border border-gray-200 px-4 py-2 text-xs font-medium text-gray-700 hover:bg-gray-50 dark:border-gray-700 dark:text-gray-300 dark:hover:bg-gray-800"
					>
						Cancel
					</button>
					<button
						type="submit"
						disabled={savingGateway}
						class="rounded-lg bg-emerald-600 px-4 py-2 text-xs font-semibold text-white shadow-xs hover:bg-emerald-700 disabled:opacity-60"
					>
						{savingGateway ? 'Saving...' : 'Save Changes'}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}

<!-- Add / Edit Operator Rate Modal -->
{#if showRateModal}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-gray-900/50 p-4 backdrop-blur-xs">
		<div class="w-full max-w-md rounded-2xl border border-gray-200 bg-white p-6 shadow-xl dark:border-gray-800 dark:bg-gray-900">
			<div class="flex items-center justify-between border-b border-gray-100 pb-3 dark:border-gray-800">
				<h3 class="text-base font-bold text-gray-900 dark:text-white">
					{rateModalMode === 'create' ? 'Add Operator Rate' : 'Edit Operator Rate'}
				</h3>
				<button onclick={() => (showRateModal = false)} class="rounded-lg p-1 text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800">
					<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
				</button>
			</div>

			{#if rateModalError}
				<div class="mt-3 rounded-lg bg-rose-50 p-3 text-xs text-rose-700 dark:bg-rose-950/40 dark:text-rose-300">{rateModalError}</div>
			{/if}

			<form onsubmit={(e) => { e.preventDefault(); void saveRate(); }} class="mt-4 space-y-4">
				{#if rateModalMode === 'create'}
					<div>
						<label class="block text-xs font-semibold text-gray-700 dark:text-gray-300">Operator</label>
						<select
							bind:value={rateOperatorId}
							class="mt-1 w-full rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
						>
							{#each operatorsList as op}
								<option value={op.id}>{op.name} {op.country_name ? `(${op.country_name})` : ''}</option>
							{/each}
						</select>
					</div>
				{/if}

				<div>
					<label class="block text-xs font-semibold text-gray-700 dark:text-gray-300">Masking Price (৳)</label>
					<input
						type="number"
						step="0.0001"
						bind:value={rateMaskingPrice}
						placeholder="Optional, e.g. 0.4500"
						class="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2 text-xs outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
					/>
				</div>

				<div>
					<label class="block text-xs font-semibold text-gray-700 dark:text-gray-300">MNO Operator Rate (Non-Masking) (৳)</label>
					<input
						type="number"
						step="0.0001"
						bind:value={rateNonMaskingPrice}
						placeholder="e.g. 0.1600"
						class="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2 text-xs outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
					/>
				</div>

				<div>
					<label class="block text-xs font-semibold text-gray-700 dark:text-gray-300">Dipping Charge (Per Number) (৳)</label>
					<input
						type="number"
						step="0.0001"
						bind:value={rateDippingCharge}
						placeholder="e.g. 0.0575"
						class="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2 text-xs outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
					/>
				</div>

				<div class="flex justify-end gap-2 border-t border-gray-100 pt-3 dark:border-gray-800">
					<button
						type="button"
						onclick={() => (showRateModal = false)}
						class="rounded-lg border border-gray-200 px-4 py-2 text-xs font-medium text-gray-700 hover:bg-gray-50 dark:border-gray-700 dark:text-gray-300 dark:hover:bg-gray-800"
					>
						Cancel
					</button>
					<button
						type="submit"
						disabled={savingRate}
						class="rounded-lg bg-emerald-600 px-4 py-2 text-xs font-semibold text-white shadow-xs hover:bg-emerald-700 disabled:opacity-60"
					>
						{savingRate ? 'Saving...' : 'Save Rate'}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}

