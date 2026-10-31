<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface Customer {
		id: number;
		name: string;
		email: string;
		phone: string;
		business_name: string | null;
		balance: number;
		kyc_status: string;
		status: string;
		nid_number?: string | null;
		tariff?: string | null;
		last_login_at?: string | null;
		created_by_admin?: boolean;
		created_at: string;
	}

	interface CustomerSummary {
		total_customers: number;
		total_balance: number;
		avg_balance: number;
		funded_customers_count: number;
		zero_balance_count: number;
		with_nid_count: number;
		verified_kyc_count: number;
		under_review_kyc_count: number;
		other_kyc_count: number;
		kyc_verified_rate: number;
		approved_status_count: number;
		pending_status_count: number;
		inactive_status_count: number;
		approved_rate: number;
		funded_percentage: number;
	}

	interface TariffPlan {
		id: number;
		name: string;
	}

	interface CustomerDetail {
		customer: Customer;
		company_name: string | null;
		nid_number: string | null;
		front_nid: string | null;
		back_nid: string | null;
		live_photo: string | null;
		business_address: string | null;
		trade_license_number: string | null;
		trade_license_file: string | null;
		kyc_submitted_at: string | null;
		kyc_verified_at: string | null;
		api_access?: boolean;
		sms_auto_approved?: boolean;
		tariff_plan_id?: number | null;
		senders?: Array<{ id: number; sender_id: string; type: string; gateway_name: string }>;
		ip_whitelists?: Array<{ id: number; ip_address: string; status: string }>;
		recent_transactions?: Array<{ id: number; amount: number; direction: string; balance_before: number; balance_after: number; description: string | null; created_at: string }>;
	}

	// Filter state
	let creationSource = $state('all');
	let kycFilter = $state('all');
	let statusFilter = $state('all');
	let tariffFilter = $state('all');
	let balanceRange = $state('all');
	let regFrom = $state('');
	let regUntil = $state('');
	let searchQuery = $state('');

	// Pagination state
	let currentPage = $state(1);
	let perPage = $state(10);
	let totalRecords = $state(0);

	let customers = $state<Customer[]>([]);
	let tariffPlans = $state<TariffPlan[]>([]);
	let summary = $state<CustomerSummary>({
		total_customers: 0,
		total_balance: 0,
		avg_balance: 0,
		funded_customers_count: 0,
		zero_balance_count: 0,
		with_nid_count: 0,
		verified_kyc_count: 0,
		under_review_kyc_count: 0,
		other_kyc_count: 0,
		kyc_verified_rate: 0,
		approved_status_count: 0,
		pending_status_count: 0,
		inactive_status_count: 0,
		approved_rate: 0,
		funded_percentage: 0
	});

	let loading = $state(false);
	let activeMenuId = $state<number | null>(null);

	// Modals
	let depositModalOpen = $state(false);
	let selectedCustomer = $state<Customer | null>(null);
	let depositAmount = $state<number>(500);
	let depositDescription = $state('');
	let depositSubmitting = $state(false);
	let depositError = $state<string | null>(null);

	let viewModalOpen = $state(false);
	let detailLoading = $state(false);
	let currentDetail = $state<CustomerDetail | null>(null);

	let editModalOpen = $state(false);
	let editCustomerData = $state({
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

	let resubmitModalOpen = $state(false);
	let resubmitReason = $state('');
	let resubmitFields = $state('front_nid, back_nid, live_photo');
	let resubmitSubmitting = $state(false);

	let rejectModalOpen = $state(false);
	let rejectReason = $state('');
	let rejectRemarks = $state('');
	let rejectSubmitting = $state(false);

	function resetFilters() {
		creationSource = 'all';
		kycFilter = 'all';
		statusFilter = 'all';
		tariffFilter = 'all';
		balanceRange = 'all';
		regFrom = '';
		regUntil = '';
		searchQuery = '';
		currentPage = 1;
		void loadCustomers();
	}

	async function loadTariffPlans() {
		try {
			const res = await supermanApi.get<{ data: TariffPlan[] }>('/api/v1/superman/tariff-plans');
			tariffPlans = res.data ?? [];
		} catch {
			tariffPlans = [];
		}
	}

	async function loadCustomers() {
		loading = true;
		try {
			const params = new URLSearchParams({
				page: String(currentPage),
				per_page: String(perPage)
			});

			if (searchQuery.trim()) params.set('search', searchQuery.trim());
			if (creationSource !== 'all') params.set('creation_source', creationSource);
			if (kycFilter !== 'all') params.set('kyc_status', kycFilter);
			if (statusFilter !== 'all') params.set('status', statusFilter);
			if (tariffFilter !== 'all') params.set('tariff_plan', tariffFilter);
			if (balanceRange !== 'all') params.set('balance_range', balanceRange);
			if (regFrom) params.set('registered_from', regFrom);
			if (regUntil) params.set('registered_until', regUntil);

			const response = await supermanApi.get<{
				data: Customer[];
				meta: { page: number; per_page: number; total: number };
				summary?: CustomerSummary;
			}>(`/api/v1/superman/customers?${params.toString()}`);

			customers = response.data ?? [];
			totalRecords = response.meta?.total ?? customers.length;

			if (response.summary) {
				summary = response.summary;
			} else {
				// Fallback client-side calculation if summary not populated
				computeClientSummary(customers);
			}
		} catch (error: unknown) {
			window.alert(error instanceof Error ? error.message : 'Unable to load customers.');
		} finally {
			loading = false;
		}
	}

	function computeClientSummary(list: Customer[]) {
		const total = list.length;
		const totalBal = list.reduce((acc, c) => acc + (c.balance || 0), 0);
		const funded = list.filter((c) => c.balance > 0).length;
		const verified = list.filter((c) => c.kyc_status === 'verified').length;
		const underReview = list.filter((c) => c.kyc_status === 'under_review').length;
		const approved = list.filter((c) => c.status === 'approved').length;
		const pending = list.filter((c) => c.status === 'pending').length;
		const inactive = list.filter((c) => ['rejected', 'banned'].includes(c.status)).length;
		const withNid = list.filter((c) => Boolean(c.nid_number)).length;

		summary = {
			total_customers: total,
			total_balance: totalBal,
			avg_balance: total > 0 ? totalBal / total : 0,
			funded_customers_count: funded,
			zero_balance_count: total - funded,
			with_nid_count: withNid,
			verified_kyc_count: verified,
			under_review_kyc_count: underReview,
			other_kyc_count: Math.max(0, total - (verified + underReview)),
			kyc_verified_rate: total > 0 ? Math.round((verified / total) * 1000) / 10 : 0,
			approved_status_count: approved,
			pending_status_count: pending,
			inactive_status_count: inactive,
			approved_rate: total > 0 ? Math.round((approved / total) * 1000) / 10 : 0,
			funded_percentage: total > 0 ? Math.round((funded / total) * 1000) / 10 : 0
		};
	}

	function handleRowClick(c: Customer, e: MouseEvent) {
		const target = e.target as HTMLElement;
		if (target.closest('button') || target.closest('a') || target.closest('select') || target.closest('.no-row-nav')) {
			return;
		}
		goto(`/superman/users/customers/${c.id}`);
	}

	function openDepositModal(c: Customer) {
		selectedCustomer = c;
		depositAmount = 500;
		depositDescription = '';
		depositError = null;
		depositModalOpen = true;
		activeMenuId = null;
	}

	async function submitDeposit() {
		if (!selectedCustomer || depositAmount === 0) return;
		depositSubmitting = true;
		depositError = null;
		try {
			const res = await supermanApi.post<{ message: string; data: { balance: number } }>(
				`/api/v1/superman/customers/${selectedCustomer.id}/deposit`,
				{
					amount: depositAmount,
					description: depositDescription
				}
			);
			selectedCustomer.balance = res.data.balance;
			depositModalOpen = false;
			window.alert(res.message);
			void loadCustomers();
		} catch (error: unknown) {
			depositError = error instanceof Error ? error.message : 'Unable to adjust balance.';
		} finally {
			depositSubmitting = false;
		}
	}

	async function openViewModal(c: Customer) {
		selectedCustomer = c;
		viewModalOpen = true;
		detailLoading = true;
		currentDetail = null;
		activeMenuId = null;
		try {
			const res = await supermanApi.get<{ data: CustomerDetail }>(`/api/v1/superman/customers/${c.id}`);
			currentDetail = res.data;
		} catch (error: unknown) {
			window.alert(error instanceof Error ? error.message : 'Unable to load customer details.');
			viewModalOpen = false;
		} finally {
			detailLoading = false;
		}
	}

	async function openEditModal(c: Customer) {
		selectedCustomer = c;
		activeMenuId = null;
		editModalOpen = true;
		editCustomerData = {
			id: c.id,
			name: c.name,
			email: c.email,
			phone: c.phone,
			business_name: c.business_name ?? '',
			status: c.status,
			api_access: false,
			sms_auto_approved: false,
			tariff_plan_id: 0
		};

		try {
			const res = await supermanApi.get<{ data: CustomerDetail }>(`/api/v1/superman/customers/${c.id}`);
			if (res.data) {
				editCustomerData.api_access = Boolean(res.data.api_access);
				editCustomerData.sms_auto_approved = Boolean(res.data.sms_auto_approved);
				editCustomerData.tariff_plan_id = res.data.tariff_plan_id ?? 0;
			}
		} catch {
			// keep defaults
		}
	}

	async function submitEdit() {
		if (!editCustomerData.id) return;
		editSubmitting = true;
		try {
			await supermanApi.put(`/api/v1/superman/customers/${editCustomerData.id}`, {
				name: editCustomerData.name,
				email: editCustomerData.email,
				phone: editCustomerData.phone,
				business_name: editCustomerData.business_name,
				status: editCustomerData.status,
				api_access: editCustomerData.api_access,
				sms_auto_approved: editCustomerData.sms_auto_approved,
				tariff_plan_id: editCustomerData.tariff_plan_id > 0 ? editCustomerData.tariff_plan_id : null
			});
			editModalOpen = false;
			window.alert('Customer updated successfully.');
			void loadCustomers();
		} catch (error: unknown) {
			window.alert(error instanceof Error ? error.message : 'Unable to update customer.');
		} finally {
			editSubmitting = false;
		}
	}

	async function quickApproveKYC(c: Customer) {
		activeMenuId = null;
		if (!confirm(`Are you sure you want to approve KYC for ${c.name}?`)) return;
		try {
			await supermanApi.post(`/api/v1/superman/customers/${c.id}/approve-kyc`);
			c.kyc_status = 'verified';
			c.status = 'approved';
			window.alert('KYC approved successfully.');
			void loadCustomers();
		} catch (error: unknown) {
			window.alert(error instanceof Error ? error.message : 'Unable to approve KYC.');
		}
	}

	function openResubmitModal(c: Customer) {
		selectedCustomer = c;
		activeMenuId = null;
		resubmitReason = '';
		resubmitFields = 'front_nid, back_nid, live_photo';
		resubmitModalOpen = true;
	}

	async function submitResubmit() {
		if (!selectedCustomer || !resubmitReason.trim()) return;
		resubmitSubmitting = true;
		try {
			await supermanApi.post(`/api/v1/superman/customers/${selectedCustomer.id}/resubmit-kyc`, {
				reason: resubmitReason.trim(),
				fields: resubmitFields.trim()
			});
			resubmitModalOpen = false;
			window.alert('Resubmission request sent to customer.');
			void loadCustomers();
		} catch (error: unknown) {
			window.alert(error instanceof Error ? error.message : 'Unable to request resubmission.');
		} finally {
			resubmitSubmitting = false;
		}
	}

	function openRejectModal(c: Customer) {
		selectedCustomer = c;
		activeMenuId = null;
		rejectReason = '';
		rejectRemarks = '';
		rejectModalOpen = true;
	}

	async function submitReject() {
		if (!selectedCustomer || !rejectReason.trim()) return;
		rejectSubmitting = true;
		try {
			await supermanApi.post(`/api/v1/superman/customers/${selectedCustomer.id}/reject-kyc`, {
				reason: rejectReason.trim(),
				admin_remarks: rejectRemarks.trim()
			});
			rejectModalOpen = false;
			window.alert('KYC rejected.');
			void loadCustomers();
		} catch (error: unknown) {
			window.alert(error instanceof Error ? error.message : 'Unable to reject KYC.');
		} finally {
			rejectSubmitting = false;
		}
	}

	function handleImpersonate(c: Customer) {
		activeMenuId = null;
		window.alert(`Simulating Login As ${c.name} (${c.email}). Redirecting to customer portal session...`);
	}

	function closeMenu() {
		activeMenuId = null;
	}

	onMount(() => {
		void loadTariffPlans();
		void loadCustomers();
	});

	let totalPages = $derived(Math.max(1, Math.ceil(totalRecords / perPage)));
	let startRecord = $derived(totalRecords === 0 ? 0 : (currentPage - 1) * perPage + 1);
	let endRecord = $derived(Math.min(currentPage * perPage, totalRecords));

	const exportColumns: ExportColumn<Customer>[] = [
		{ header: 'Customer Name', getValue: (c) => c.name },
		{ header: 'Email', getValue: (c) => c.email },
		{ header: 'Phone', getValue: (c) => c.phone },
		{ header: 'Business / Company', getValue: (c) => c.business_name || '—' },
		{ header: 'Balance (BDT)', getValue: (c) => `BDT ${c.balance.toFixed(2)}` },
		{ header: 'KYC Status', getValue: (c) => c.kyc_status.toUpperCase() },
		{ header: 'Account Status', getValue: (c) => c.status.toUpperCase() },
		{ header: 'NID Number', getValue: (c) => c.nid_number || '—' },
		{ header: 'Tariff Plan', getValue: (c) => c.tariff || '—' },
		{ header: 'Created Date', getValue: (c) => new Date(c.created_at).toLocaleDateString() }
	];
</script>

<svelte:window onclick={closeMenu} />

<svelte:head>
	<title>Customers & Verification Management — SMS Whiteinfo Administrator</title>
</svelte:head>

<div class="space-y-6 text-gray-900 dark:text-white antialiased font-sans">
	<!-- Top Breadcrumb & Page Header -->
	<div class="space-y-1">
		<nav class="flex items-center gap-1.5 text-xs text-gray-500 dark:text-gray-400">
			<span>Customers</span>
			<span>›</span>
			<span class="font-medium text-gray-900 dark:text-white">List</span>
		</nav>
		<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 pt-1">
			<h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">Customers</h1>
			<div class="flex items-center gap-2">
				<ExportDropdown
					title="Customers List"
					columns={exportColumns}
					getData={() => customers}
					disabled={customers.length === 0}
				/>
				<button
					type="button"
					onclick={() => {
						editCustomerData = { id: 0, name: '', email: '', phone: '', business_name: '', status: 'approved', api_access: false, sms_auto_approved: false, tariff_plan_id: 0 };
						editModalOpen = true;
					}}
					class="inline-flex items-center gap-1.5 rounded-lg bg-emerald-600 px-3.5 py-1.5 text-xs font-bold text-white shadow-2xs hover:bg-emerald-700 transition"
				>
					<svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke-width="2.5" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
					</svg>
					New Customer
				</button>
			</div>
		</div>
	</div>

	<!-- ============================================================
	     FILTERS PANEL (Exact Filament above-content layout)
	     ============================================================ -->
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

		<!-- Filter form fields grid -->
		<div class="mt-4 grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-3">
			<!-- 1. Creation Source -->
			<div>
				<label class="block text-[11px] font-medium text-gray-700 dark:text-gray-300 mb-1">Creation Source</label>
				<select
					bind:value={creationSource}
					onchange={() => { currentPage = 1; void loadCustomers(); }}
					class="w-full rounded-lg border border-gray-300 bg-white px-2.5 py-1.5 text-xs text-gray-900 shadow-2xs outline-none focus:border-emerald-500 focus:ring-1 focus:ring-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
				>
					<option value="all">All</option>
					<option value="1">Super Admin Created</option>
					<option value="0">Self-Registered</option>
				</select>
			</div>

			<!-- 2. KYC Status -->
			<div>
				<label class="block text-[11px] font-medium text-gray-700 dark:text-gray-300 mb-1">KYC Status</label>
				<select
					bind:value={kycFilter}
					onchange={() => { currentPage = 1; void loadCustomers(); }}
					class="w-full rounded-lg border border-gray-300 bg-white px-2.5 py-1.5 text-xs text-gray-900 shadow-2xs outline-none focus:border-emerald-500 focus:ring-1 focus:ring-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
				>
					<option value="all">All</option>
					<option value="verified">Verified</option>
					<option value="under_review">Under Review</option>
					<option value="pending">Pending</option>
					<option value="rejected">Rejected</option>
					<option value="resubmit_required">Resubmission Requested</option>
				</select>
			</div>

			<!-- 3. Account Status -->
			<div>
				<label class="block text-[11px] font-medium text-gray-700 dark:text-gray-300 mb-1">Account Status</label>
				<select
					bind:value={statusFilter}
					onchange={() => { currentPage = 1; void loadCustomers(); }}
					class="w-full rounded-lg border border-gray-300 bg-white px-2.5 py-1.5 text-xs text-gray-900 shadow-2xs outline-none focus:border-emerald-500 focus:ring-1 focus:ring-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
				>
					<option value="all">All</option>
					<option value="approved">Approved</option>
					<option value="pending">Pending</option>
					<option value="rejected">Rejected</option>
					<option value="banned">Banned</option>
				</select>
			</div>

			<!-- 4. Tariff Plan -->
			<div>
				<label class="block text-[11px] font-medium text-gray-700 dark:text-gray-300 mb-1">Tariff Plan</label>
				<select
					bind:value={tariffFilter}
					onchange={() => { currentPage = 1; void loadCustomers(); }}
					class="w-full rounded-lg border border-gray-300 bg-white px-2.5 py-1.5 text-xs text-gray-900 shadow-2xs outline-none focus:border-emerald-500 focus:ring-1 focus:ring-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
				>
					<option value="all">All</option>
					{#each tariffPlans as tp}
						<option value={String(tp.id)}>{tp.name}</option>
					{/each}
				</select>
			</div>

			<!-- 5. Balance Range -->
			<div>
				<label class="block text-[11px] font-medium text-gray-700 dark:text-gray-300 mb-1">Balance Range</label>
				<select
					bind:value={balanceRange}
					onchange={() => { currentPage = 1; void loadCustomers(); }}
					class="w-full rounded-lg border border-gray-300 bg-white px-2.5 py-1.5 text-xs text-gray-900 shadow-2xs outline-none focus:border-emerald-500 focus:ring-1 focus:ring-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
				>
					<option value="all">All</option>
					<option value="positive">Funded (&gt; 0 BDT)</option>
					<option value="zero">Zero Balance (0 BDT)</option>
					<option value="low">Low (&lt; 100 BDT)</option>
					<option value="high">High (≥ 1,000 BDT)</option>
					<option value="vip">VIP (≥ 10,000 BDT)</option>
				</select>
			</div>

			<!-- 6. Registered From -->
			<div>
				<label class="block text-[11px] font-medium text-gray-700 dark:text-gray-300 mb-1">Registered From</label>
				<input
					type="date"
					bind:value={regFrom}
					onchange={() => { currentPage = 1; void loadCustomers(); }}
					class="w-full rounded-lg border border-gray-300 bg-white px-2 py-1.5 text-xs text-gray-900 shadow-2xs outline-none focus:border-emerald-500 focus:ring-1 focus:ring-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
				/>
			</div>

			<!-- 7. Registered Until -->
			<div>
				<label class="block text-[11px] font-medium text-gray-700 dark:text-gray-300 mb-1">Registered Until</label>
				<input
					type="date"
					bind:value={regUntil}
					onchange={() => { currentPage = 1; void loadCustomers(); }}
					class="w-full rounded-lg border border-gray-300 bg-white px-2 py-1.5 text-xs text-gray-900 shadow-2xs outline-none focus:border-emerald-500 focus:ring-1 focus:ring-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
				/>
			</div>
		</div>
	</div>

	<!-- ============================================================
	     CUSTOMERS TABLE CARD
	     ============================================================ -->
	<div class="rounded-xl bg-white shadow-xs ring-1 ring-gray-950/5 dark:bg-gray-900 dark:ring-white/10 overflow-hidden">
		<!-- Search & Table Density Toolbar -->
		<div class="flex items-center justify-end gap-3 px-4 py-3 border-b border-gray-100 dark:border-gray-800">
			<div class="relative w-72">
				<svg class="absolute left-3 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z" />
				</svg>
				<input
					type="text"
					bind:value={searchQuery}
					oninput={() => { currentPage = 1; void loadCustomers(); }}
					placeholder="Search"
					class="w-full pl-9 pr-3 py-1.5 rounded-lg border border-gray-200 bg-white dark:border-gray-700 dark:bg-gray-800 dark:text-white text-xs outline-none focus:border-emerald-500 focus:ring-1 focus:ring-emerald-500 placeholder:text-gray-400"
				/>
			</div>
			<button type="button" class="flex h-8 w-8 items-center justify-center rounded-lg border border-gray-200 text-gray-400 hover:bg-gray-50 dark:border-gray-700 dark:hover:bg-gray-800" title="Columns">
				<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke-width="1.75" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" d="M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25h16.5" />
				</svg>
			</button>
		</div>

		<!-- Table -->
		<div class="overflow-x-auto">
			<table class="w-full text-xs text-left">
				<thead class="bg-gray-50/75 dark:bg-gray-800/50 border-b border-gray-200/80 dark:border-gray-800 text-[10px] uppercase font-bold text-gray-500 dark:text-gray-400 tracking-wider">
					<tr>
						<th class="px-4 py-3">
							<span class="inline-flex items-center gap-1 cursor-pointer hover:text-gray-700 dark:hover:text-gray-200">
								PHONE
								<svg class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M19.5 8.25l-7.5 7.5-7.5-7.5" /></svg>
							</span>
						</th>
						<th class="px-4 py-3">TARIFF</th>
						<th class="px-4 py-3">NID NUMBER</th>
						<th class="px-4 py-3">
							<span class="inline-flex items-center gap-1 cursor-pointer hover:text-gray-700 dark:hover:text-gray-200">
								KYC STATUS
								<svg class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M19.5 8.25l-7.5 7.5-7.5-7.5" /></svg>
							</span>
						</th>
						<th class="px-4 py-3">
							<span class="inline-flex items-center gap-1 cursor-pointer hover:text-gray-700 dark:hover:text-gray-200">
								ACCOUNT STATUS
								<svg class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M19.5 8.25l-7.5 7.5-7.5-7.5" /></svg>
							</span>
						</th>
						<th class="px-4 py-3">
							<span class="inline-flex items-center gap-1 cursor-pointer hover:text-gray-700 dark:hover:text-gray-200">
								LAST LOGIN
								<svg class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M19.5 8.25l-7.5 7.5-7.5-7.5" /></svg>
							</span>
						</th>
						<th class="px-4 py-3 text-right">ACTIONS</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
					{#if loading && customers.length === 0}
						<tr>
							<td colspan="7" class="py-12 text-center text-xs text-gray-500">
								<div class="flex items-center justify-center gap-2">
									<div class="h-4 w-4 animate-spin rounded-full border-2 border-emerald-600 border-t-transparent"></div>
									<span>Loading customers...</span>
								</div>
							</td>
						</tr>
					{:else if customers.length === 0}
						<tr>
							<td colspan="7" class="py-12 text-center text-xs text-gray-500">
								No customers found matching the selected filters.
							</td>
						</tr>
					{:else}
						{#each customers as c (c.id)}
							<tr
								onclick={(e) => handleRowClick(c, e)}
								class="hover:bg-gray-50/70 dark:hover:bg-gray-800/40 transition-colors cursor-pointer group"
							>
								<!-- Phone + Name + Email -->
								<td class="px-4 py-3">
									<div class="font-mono text-xs font-semibold text-gray-900 dark:text-white group-hover:text-emerald-600 transition-colors">
										{c.phone || '—'}
									</div>
									<div class="text-[11px] text-gray-500 dark:text-gray-400 truncate max-w-xs">
										<span class="font-medium text-gray-700 dark:text-gray-300">{c.name}</span>
										{#if c.email}
											<span class="text-gray-400"> • {c.email}</span>
										{/if}
									</div>
								</td>

								<!-- Tariff -->
								<td class="px-4 py-3 text-xs text-gray-700 dark:text-gray-300">
									{#if c.tariff}
										<span class="inline-block rounded bg-sky-50 px-2 py-0.5 text-[11px] font-semibold text-sky-700 dark:bg-sky-950/40 dark:text-sky-300 border border-sky-200 dark:border-sky-800">
											{c.tariff}
										</span>
									{:else}
										<span class="text-gray-400">—</span>
									{/if}
								</td>

								<!-- NID Number -->
								<td class="px-4 py-3 font-mono text-xs text-gray-600 dark:text-gray-400">
									{c.nid_number || '—'}
								</td>

								<!-- KYC Status -->
								<td class="px-4 py-3">
									{#if c.kyc_status === 'verified'}
										<span class="inline-flex items-center gap-1 rounded-full bg-emerald-50 px-2.5 py-0.5 text-[11px] font-semibold text-emerald-700 border border-emerald-200 dark:bg-emerald-950/40 dark:border-emerald-800 dark:text-emerald-400">
											<span class="h-1.5 w-1.5 rounded-full bg-emerald-500"></span>
											Verified
										</span>
									{:else if c.kyc_status === 'under_review'}
										<span class="inline-flex items-center gap-1 rounded-full bg-amber-50 px-2.5 py-0.5 text-[11px] font-semibold text-amber-700 border border-amber-200 dark:bg-amber-950/40 dark:border-amber-800 dark:text-amber-400">
											<span class="h-1.5 w-1.5 rounded-full bg-amber-500"></span>
											Under Review
										</span>
									{:else if c.kyc_status === 'resubmit_required'}
										<span class="inline-flex items-center gap-1 rounded-full bg-orange-50 px-2.5 py-0.5 text-[11px] font-semibold text-orange-700 border border-orange-200 dark:bg-orange-950/40 dark:border-orange-800 dark:text-orange-400">
											<span class="h-1.5 w-1.5 rounded-full bg-orange-500"></span>
											Resubmit
										</span>
									{:else if c.kyc_status === 'rejected'}
										<span class="inline-flex items-center gap-1 rounded-full bg-rose-50 px-2.5 py-0.5 text-[11px] font-semibold text-rose-700 border border-rose-200 dark:bg-rose-950/40 dark:border-rose-800 dark:text-rose-400">
											<span class="h-1.5 w-1.5 rounded-full bg-rose-500"></span>
											Rejected
										</span>
									{:else}
										<span class="inline-flex items-center gap-1 rounded-full bg-gray-100 px-2.5 py-0.5 text-[11px] font-semibold text-gray-600 dark:bg-gray-800 dark:text-gray-400">
											<span class="h-1.5 w-1.5 rounded-full bg-gray-400"></span>
											Pending
										</span>
									{/if}
								</td>

								<!-- Account Status -->
								<td class="px-4 py-3">
									{#if c.status === 'approved'}
										<span class="inline-flex items-center gap-1 rounded-full bg-emerald-50 px-2.5 py-0.5 text-[11px] font-semibold text-emerald-700 border border-emerald-200 dark:bg-emerald-950/40 dark:border-emerald-800 dark:text-emerald-400">
											<span class="h-1.5 w-1.5 rounded-full bg-emerald-500"></span>
											Approved
										</span>
									{:else if c.status === 'pending'}
										<span class="inline-flex items-center gap-1 rounded-full bg-amber-50 px-2.5 py-0.5 text-[11px] font-semibold text-amber-700 border border-amber-200 dark:bg-amber-950/40 dark:border-amber-800 dark:text-amber-400">
											<span class="h-1.5 w-1.5 rounded-full bg-amber-500"></span>
											Pending
										</span>
									{:else}
										<span class="inline-flex items-center gap-1 rounded-full bg-rose-50 px-2.5 py-0.5 text-[11px] font-semibold text-rose-700 border border-rose-200 dark:bg-rose-950/40 dark:border-rose-800 dark:text-rose-400">
											<span class="h-1.5 w-1.5 rounded-full bg-rose-500"></span>
											{c.status}
										</span>
									{/if}
								</td>

								<!-- Last Login -->
								<td class="px-4 py-3 text-xs text-gray-500 dark:text-gray-400">
									{c.last_login_at || 'Never'}
								</td>

								<!-- Actions Column -->
								<td class="px-4 py-3 text-right no-row-nav">
									<div class="flex items-center justify-end gap-1.5 relative">
										<!-- Login As -->
										<button
											type="button"
											onclick={(e) => { e.stopPropagation(); handleImpersonate(c); }}
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
													activeMenuId = activeMenuId === c.id ? null : c.id;
												}}
												class="flex h-7 w-7 items-center justify-center rounded-full border border-emerald-500/80 text-emerald-600 hover:bg-emerald-50 dark:hover:bg-emerald-950/40 transition"
												title="Actions"
											>
												<svg class="h-4 w-4" fill="currentColor" viewBox="0 0 24 24">
													<path d="M12 6.75a1.5 1.5 0 1 1 0-3 1.5 1.5 0 0 1 0 3Zm0 6.75a1.5 1.5 0 1 1 0-3 1.5 1.5 0 0 1 0 3Zm0 6.75a1.5 1.5 0 1 1 0-3 1.5 1.5 0 0 1 0 3Z" />
												</svg>
											</button>

											{#if activeMenuId === c.id}
												<div
													class="absolute right-0 top-full z-50 mt-1 w-56 rounded-xl border border-gray-200 bg-white py-1.5 shadow-xl dark:border-gray-700 dark:bg-gray-800 text-left"
												>
													<!-- 1. View Customer & KYC -->
													<button
														type="button"
														onclick={() => openViewModal(c)}
														class="flex w-full items-center gap-2.5 px-3.5 py-2 text-xs font-medium text-sky-600 hover:bg-sky-50 dark:text-sky-400 dark:hover:bg-sky-950/40"
													>
														<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke-width="1.75" stroke="currentColor">
															<path stroke-linecap="round" stroke-linejoin="round" d="M2.036 12.322a1.012 1.012 0 0 1 0-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178Z" /><path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z" />
														</svg>
														View Customer & KYC
													</button>

													<!-- 2. Make Deposit -->
													<button
														type="button"
														onclick={() => openDepositModal(c)}
														class="flex w-full items-center gap-2.5 px-3.5 py-2 text-xs font-medium text-emerald-600 hover:bg-emerald-50 dark:text-emerald-400 dark:hover:bg-emerald-950/40"
													>
														<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke-width="1.75" stroke="currentColor">
															<path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
														</svg>
														Make Deposit
													</button>

													<!-- 3. Edit Customer -->
													<button
														type="button"
														onclick={() => openEditModal(c)}
														class="flex w-full items-center gap-2.5 px-3.5 py-2 text-xs font-medium text-amber-600 hover:bg-amber-50 dark:text-amber-400 dark:hover:bg-amber-950/40"
													>
														<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke-width="1.75" stroke="currentColor">
															<path stroke-linecap="round" stroke-linejoin="round" d="m16.862 4.487 1.687-1.688a1.875 1.875 0 1 1 2.652 2.652L6.832 19.82a4.5 4.5 0 0 1-1.897 1.13l-2.685.8.8-2.685a4.5 4.5 0 0 1 1.13-1.897L16.863 4.487Zm0 0L19.5 7.125" />
														</svg>
														Edit Customer
													</button>

													<div class="my-1 border-t border-gray-100 dark:border-gray-700/60"></div>

													<!-- 4. Approve KYC -->
													<button
														type="button"
														onclick={() => quickApproveKYC(c)}
														class="flex w-full items-center gap-2.5 px-3.5 py-2 text-xs font-medium text-emerald-700 hover:bg-emerald-50 dark:text-emerald-300 dark:hover:bg-emerald-950/40"
													>
														<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke-width="1.75" stroke="currentColor">
															<path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
														</svg>
														Approve KYC
													</button>

													<!-- 5. Request Resubmission -->
													<button
														type="button"
														onclick={() => openResubmitModal(c)}
														class="flex w-full items-center gap-2.5 px-3.5 py-2 text-xs font-medium text-orange-600 hover:bg-orange-50 dark:text-orange-400 dark:hover:bg-orange-950/40"
													>
														<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke-width="1.75" stroke="currentColor">
															<path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0 3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182m0-4.991v4.99" />
														</svg>
														Request Resubmission
													</button>

													<!-- 6. Reject KYC -->
													<button
														type="button"
														onclick={() => openRejectModal(c)}
														class="flex w-full items-center gap-2.5 px-3.5 py-2 text-xs font-medium text-rose-600 hover:bg-rose-50 dark:text-rose-400 dark:hover:bg-rose-950/40"
													>
														<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke-width="1.75" stroke="currentColor">
															<path stroke-linecap="round" stroke-linejoin="round" d="m9.75 9.75 4.5 4.5m0-4.5-4.5 4.5M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
														</svg>
														Reject KYC
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

		<!-- Pagination Footer (Exact Filament style from Screenshot 1) -->
		<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 px-4 py-3 border-t border-gray-100 dark:border-gray-800 text-xs text-gray-600 dark:text-gray-400">
			<div>
				Showing {startRecord} to {endRecord} of {totalRecords} results
			</div>

			<div class="flex items-center gap-4">
				<div class="flex items-center gap-1.5">
					<span>Per page</span>
					<select
						bind:value={perPage}
						onchange={() => { currentPage = 1; void loadCustomers(); }}
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
						onclick={() => { currentPage--; void loadCustomers(); }}
						class="h-7 w-7 rounded-lg border border-gray-200 bg-white flex items-center justify-center hover:bg-gray-50 disabled:opacity-40 dark:border-gray-700 dark:bg-gray-800"
					>
						‹
					</button>
					{#each Array.from({ length: Math.min(5, totalPages) }, (_, i) => i + 1) as pageNum}
						<button
							type="button"
							onclick={() => { currentPage = pageNum; void loadCustomers(); }}
							class="h-7 min-w-[28px] px-2 rounded-lg text-xs font-semibold {currentPage === pageNum ? 'bg-emerald-600 text-white' : 'border border-gray-200 bg-white hover:bg-gray-50 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-200'}"
						>
							{pageNum}
						</button>
					{/each}
					<button
						type="button"
						disabled={currentPage >= totalPages}
						onclick={() => { currentPage++; void loadCustomers(); }}
						class="h-7 w-7 rounded-lg border border-gray-200 bg-white flex items-center justify-center hover:bg-gray-50 disabled:opacity-40 dark:border-gray-700 dark:bg-gray-800"
					>
						›
					</button>
				</div>
			</div>
		</div>
	</div>

	<!-- ============================================================
	     CUSTOMER SUMMARY & TOTAL CALCULATION (Screenshot 1 Replica)
	     ============================================================ -->
	<div class="rounded-xl bg-white p-6 shadow-xs ring-1 ring-gray-950/5 dark:bg-gray-900 dark:ring-white/10">
		<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between pb-4 border-b border-gray-100 dark:border-gray-800 gap-2">
			<div>
				<h3 class="text-base font-semibold text-gray-950 dark:text-white">
					Customer Summary &amp; Total Calculation
				</h3>
				<p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
					Aggregate financial and verification metrics for the selected customer filters
				</p>
			</div>
			<div>
				<span class="inline-flex items-center px-2.5 py-1 rounded-md text-xs font-bold bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300">
					{summary.total_customers} Customers
				</span>
			</div>
		</div>

		<!-- Stats Grid -->
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mt-6">
			<!-- Total Customer Balance -->
			<div class="rounded-xl bg-gray-50/70 p-4 ring-1 ring-gray-950/5 dark:bg-gray-800/40 dark:ring-white/10 flex flex-col justify-between">
				<div>
					<div class="text-xs font-medium text-gray-500 dark:text-gray-400">Total Customer Balance</div>
					<div class="mt-2 text-2xl font-bold tracking-tight text-gray-950 dark:text-white font-mono">
						BDT {summary.total_balance.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
					</div>
				</div>
				<div class="mt-4 flex items-center justify-between text-xs text-gray-500 dark:text-gray-400 pt-2.5 border-t border-gray-200/60 dark:border-gray-700/60">
					<span>Average: <strong class="text-gray-800 dark:text-gray-200 font-semibold font-mono">BDT {summary.avg_balance.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}</strong></span>
					<span>Funded: <strong class="text-emerald-600 dark:text-emerald-400 font-semibold">{summary.funded_customers_count}</strong></span>
				</div>
			</div>

			<!-- Total Customers Count -->
			<div class="rounded-xl bg-gray-50/70 p-4 ring-1 ring-gray-950/5 dark:bg-gray-800/40 dark:ring-white/10 flex flex-col justify-between">
				<div>
					<div class="text-xs font-medium text-gray-500 dark:text-gray-400">Total Customers</div>
					<div class="mt-2 text-2xl font-bold tracking-tight text-gray-950 dark:text-white font-mono">
						{summary.total_customers}
					</div>
				</div>
				<div class="mt-4 flex items-center justify-between text-xs text-gray-500 dark:text-gray-400 pt-2.5 border-t border-gray-200/60 dark:border-gray-700/60">
					<span>Zero Balance: <strong class="text-gray-800 dark:text-gray-200 font-semibold">{summary.zero_balance_count}</strong></span>
					<span>With NID: <strong class="text-gray-800 dark:text-gray-200 font-semibold">{summary.with_nid_count}</strong></span>
				</div>
			</div>

			<!-- KYC Verification -->
			<div class="rounded-xl bg-gray-50/70 p-4 ring-1 ring-gray-950/5 dark:bg-gray-800/40 dark:ring-white/10 flex flex-col justify-between">
				<div>
					<div class="text-xs font-medium text-gray-500 dark:text-gray-400">KYC Verification</div>
					<div class="mt-2 flex items-baseline gap-2">
						<span class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white font-mono">{summary.verified_kyc_count}</span>
						<span class="text-xs font-bold text-emerald-600 dark:text-emerald-400">({summary.kyc_verified_rate}%)</span>
					</div>
				</div>
				<div class="mt-4 flex items-center justify-between text-xs text-gray-500 dark:text-gray-400 pt-2.5 border-t border-gray-200/60 dark:border-gray-700/60">
					<span>Under Review: <strong class="text-amber-600 dark:text-amber-400 font-semibold">{summary.under_review_kyc_count}</strong></span>
					<span>Pending/Other: <strong class="text-gray-800 dark:text-gray-200 font-semibold">{summary.other_kyc_count}</strong></span>
				</div>
			</div>

			<!-- Account Standing -->
			<div class="rounded-xl bg-gray-50/70 p-4 ring-1 ring-gray-950/5 dark:bg-gray-800/40 dark:ring-white/10 flex flex-col justify-between">
				<div>
					<div class="text-xs font-medium text-gray-500 dark:text-gray-400">Account Standing</div>
					<div class="mt-2 flex items-baseline gap-2">
						<span class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white font-mono">{summary.approved_status_count}</span>
						<span class="text-xs font-bold text-emerald-600 dark:text-emerald-400">({summary.approved_rate}%)</span>
					</div>
				</div>
				<div class="mt-4 flex items-center justify-between text-xs text-gray-500 dark:text-gray-400 pt-2.5 border-t border-gray-200/60 dark:border-gray-700/60">
					<span>Pending: <strong class="text-amber-600 dark:text-amber-400 font-semibold">{summary.pending_status_count}</strong></span>
					<span>Inactive/Banned: <strong class="text-rose-600 dark:text-rose-400 font-semibold">{summary.inactive_status_count}</strong></span>
				</div>
			</div>
		</div>

		<!-- Aggregate Banner -->
		<div class="mt-4 rounded-xl p-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 bg-gray-50/80 dark:bg-gray-800/40 ring-1 ring-gray-950/5 dark:ring-white/10">
			<div>
				<div class="text-sm font-semibold text-gray-950 dark:text-white">
					Aggregate Customer Deposits &amp; Wallet Reserve
				</div>
				<div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
					Total funds currently available in active customer wallets for SMS consumption
				</div>
			</div>
			<div class="text-left sm:text-right">
				<div class="text-xl font-bold tracking-tight text-emerald-600 dark:text-emerald-400 font-mono">
					BDT {summary.total_balance.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
				</div>
				<div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
					{summary.funded_percentage}% of customers currently maintain a positive balance
				</div>
			</div>
		</div>
	</div>
</div>

<!-- ============================================================
     VIEW CUSTOMER MODAL (User requested to keep this beautiful modal)
     ============================================================ -->
{#if viewModalOpen}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-gray-950/60 p-4 backdrop-blur-xs">
		<div class="w-full max-w-3xl rounded-2xl bg-white shadow-2xl ring-1 ring-gray-900/10 dark:bg-gray-900 dark:ring-white/10 max-h-[90vh] flex flex-col overflow-hidden">
			<!-- Header -->
			<div class="flex items-center justify-between border-b border-gray-100 px-6 py-4 dark:border-gray-800">
				<div>
					<h3 class="text-base font-bold text-gray-900 dark:text-white">Customer &amp; KYC Profile</h3>
					<p class="text-xs text-gray-500">ID #{selectedCustomer?.id} • {selectedCustomer?.name}</p>
				</div>
				<div class="flex items-center gap-2">
					{#if selectedCustomer}
						<a
							href={`/superman/users/customers/${selectedCustomer.id}`}
							class="text-xs font-semibold text-emerald-600 hover:text-emerald-700 dark:text-emerald-400 inline-flex items-center gap-1"
						>
							Open Full Page →
						</a>
					{/if}
					<button type="button" onclick={() => (viewModalOpen = false)} class="rounded-lg p-1.5 text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800">✕</button>
				</div>
			</div>

			<!-- Body -->
			<div class="flex-1 overflow-y-auto p-6 space-y-6">
				{#if detailLoading}
					<div class="py-16 text-center text-xs text-gray-500">
						<div class="inline-block h-6 w-6 animate-spin rounded-full border-2 border-emerald-600 border-t-transparent"></div>
						<p class="mt-2">Loading customer documents &amp; credentials...</p>
					</div>
				{:else if currentDetail}
					<!-- Basic Info Grid -->
					<div class="grid grid-cols-2 sm:grid-cols-3 gap-4 p-4 rounded-xl bg-gray-50 dark:bg-gray-800/40">
						<div>
							<span class="text-[10px] text-gray-500 uppercase font-bold">Email</span>
							<p class="text-xs font-semibold text-gray-900 dark:text-white mt-0.5">{currentDetail.customer.email}</p>
						</div>
						<div>
							<span class="text-[10px] text-gray-500 uppercase font-bold">Phone</span>
							<p class="text-xs font-mono font-semibold text-gray-900 dark:text-white mt-0.5">{currentDetail.customer.phone || '—'}</p>
						</div>
						<div>
							<span class="text-[10px] text-gray-500 uppercase font-bold">Wallet Balance</span>
							<p class="text-xs font-bold text-emerald-600 dark:text-emerald-400 font-mono mt-0.5">৳{currentDetail.customer.balance.toFixed(2)}</p>
						</div>
						<div>
							<span class="text-[10px] text-gray-500 uppercase font-bold">Business Name</span>
							<p class="text-xs text-gray-800 dark:text-gray-200 mt-0.5">{currentDetail.customer.business_name || currentDetail.company_name || '—'}</p>
						</div>
						<div>
							<span class="text-[10px] text-gray-500 uppercase font-bold">NID Number</span>
							<p class="text-xs font-mono font-semibold text-gray-800 dark:text-gray-200 mt-0.5">{currentDetail.nid_number || '—'}</p>
						</div>
						<div>
							<span class="text-[10px] text-gray-500 uppercase font-bold">Trade License</span>
							<p class="text-xs font-mono text-gray-800 dark:text-gray-200 mt-0.5">{currentDetail.trade_license_number || 'Not Submitted'}</p>
						</div>
					</div>

					<!-- KYC Images Grid -->
					<div>
						<h4 class="text-xs font-bold uppercase tracking-wider text-gray-700 dark:text-gray-300 mb-3">Verification Documents</h4>
						<div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
							<div class="space-y-1.5">
								<span class="text-[11px] font-medium text-gray-600 dark:text-gray-400">Front Side of NID</span>
								<div class="h-44 rounded-xl border border-gray-200 bg-gray-50 dark:border-gray-700 dark:bg-gray-800 overflow-hidden flex items-center justify-center p-2">
									{#if currentDetail.front_nid}
										<img src={currentDetail.front_nid.startsWith('http') ? currentDetail.front_nid : `/storage/${currentDetail.front_nid}`} alt="Front NID" class="max-h-full max-w-full object-contain rounded" />
									{:else}
										<span class="text-xs text-gray-400 italic">No document uploaded</span>
									{/if}
								</div>
							</div>

							<div class="space-y-1.5">
								<span class="text-[11px] font-medium text-gray-600 dark:text-gray-400">Back Side of NID</span>
								<div class="h-44 rounded-xl border border-gray-200 bg-gray-50 dark:border-gray-700 dark:bg-gray-800 overflow-hidden flex items-center justify-center p-2">
									{#if currentDetail.back_nid}
										<img src={currentDetail.back_nid.startsWith('http') ? currentDetail.back_nid : `/storage/${currentDetail.back_nid}`} alt="Back NID" class="max-h-full max-w-full object-contain rounded" />
									{:else}
										<span class="text-xs text-gray-400 italic">No document uploaded</span>
									{/if}
								</div>
							</div>

							<div class="space-y-1.5">
								<span class="text-[11px] font-medium text-gray-600 dark:text-gray-400">Live Camera Selfie</span>
								<div class="h-44 rounded-xl border border-gray-200 bg-gray-50 dark:border-gray-700 dark:bg-gray-800 overflow-hidden flex items-center justify-center p-2">
									{#if currentDetail.live_photo}
										<img src={currentDetail.live_photo.startsWith('http') ? currentDetail.live_photo : `/storage/${currentDetail.live_photo}`} alt="Selfie" class="max-h-full max-w-full object-contain rounded" />
									{:else}
										<span class="text-xs text-gray-400 italic">No live photo captured</span>
									{/if}
								</div>
							</div>
						</div>
					</div>
				{/if}
			</div>

			<!-- Footer -->
			<div class="border-t border-gray-100 dark:border-gray-800 px-6 py-3 flex items-center justify-between bg-gray-50/50 dark:bg-gray-850">
				{#if selectedCustomer}
					<div class="flex items-center gap-2">
						<button
							type="button"
							onclick={() => { viewModalOpen = false; openDepositModal(selectedCustomer!); }}
							class="rounded-lg bg-emerald-600 px-3 py-1.5 text-xs font-semibold text-white hover:bg-emerald-700"
						>
							+ Deposit
						</button>
						<button
							type="button"
							onclick={() => { viewModalOpen = false; openEditModal(selectedCustomer!); }}
							class="rounded-lg border border-gray-300 px-3 py-1.5 text-xs font-semibold text-gray-700 hover:bg-gray-100 dark:border-gray-700 dark:text-gray-200"
						>
							Edit Customer
						</button>
					</div>
				{/if}
				<button type="button" onclick={() => (viewModalOpen = false)} class="rounded-lg bg-gray-200 dark:bg-gray-700 px-4 py-1.5 text-xs font-semibold text-gray-800 dark:text-gray-200">
					Close
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- ============================================================
     MAKE DEPOSIT MODAL
     ============================================================ -->
{#if depositModalOpen && selectedCustomer}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-gray-950/60 p-4 backdrop-blur-xs">
		<div class="w-full max-w-md rounded-2xl bg-white p-6 shadow-2xl ring-1 ring-gray-900/10 dark:bg-gray-900 dark:ring-white/10">
			<h3 class="text-base font-bold text-gray-900 dark:text-white">Adjust Customer Balance</h3>
			<p class="text-xs text-gray-500 mt-0.5">{selectedCustomer.name} (Current: ৳{selectedCustomer.balance.toFixed(2)})</p>

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
						class="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm text-gray-900 shadow-2xs outline-none focus:border-emerald-500 focus:ring-1 focus:ring-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
					/>
				</div>

				<div>
					<label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Remarks / Note</label>
					<input
						type="text"
						bind:value={depositDescription}
						placeholder="e.g. Bank deposit reference #12345"
						class="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-xs text-gray-900 shadow-2xs outline-none focus:border-emerald-500 focus:ring-1 focus:ring-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
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
     EDIT CUSTOMER MODAL
     ============================================================ -->
{#if editModalOpen}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-gray-950/60 p-4 backdrop-blur-xs">
		<div class="w-full max-w-lg rounded-2xl bg-white p-6 shadow-2xl ring-1 ring-gray-900/10 dark:bg-gray-900 dark:ring-white/10">
			<h3 class="text-base font-bold text-gray-900 dark:text-white">
				{editCustomerData.id ? 'Edit Customer' : 'Create New Customer'}
			</h3>
			<p class="text-xs text-gray-500 mt-0.5">Update profile, account status, and SMS permissions.</p>

			<div class="mt-4 space-y-3 max-h-[70vh] overflow-y-auto pr-1">
				<div>
					<label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Full Name</label>
					<input
						type="text"
						bind:value={editCustomerData.name}
						class="w-full rounded-lg border border-gray-300 px-3 py-1.5 text-xs dark:border-gray-700 dark:bg-gray-800 dark:text-white"
					/>
				</div>

				<div class="grid grid-cols-2 gap-3">
					<div>
						<label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Email</label>
						<input
							type="email"
							bind:value={editCustomerData.email}
							class="w-full rounded-lg border border-gray-300 px-3 py-1.5 text-xs dark:border-gray-700 dark:bg-gray-800 dark:text-white"
						/>
					</div>
					<div>
						<label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Phone</label>
						<input
							type="text"
							bind:value={editCustomerData.phone}
							class="w-full rounded-lg border border-gray-300 px-3 py-1.5 text-xs dark:border-gray-700 dark:bg-gray-800 dark:text-white"
						/>
					</div>
				</div>

				<div class="grid grid-cols-2 gap-3">
					<div>
						<label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Business Name</label>
						<input
							type="text"
							bind:value={editCustomerData.business_name}
							class="w-full rounded-lg border border-gray-300 px-3 py-1.5 text-xs dark:border-gray-700 dark:bg-gray-800 dark:text-white"
						/>
					</div>
					<div>
						<label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Account Status</label>
						<select
							bind:value={editCustomerData.status}
							class="w-full rounded-lg border border-gray-300 px-3 py-1.5 text-xs dark:border-gray-700 dark:bg-gray-800 dark:text-white"
						>
							<option value="approved">Approved</option>
							<option value="pending">Pending</option>
							<option value="rejected">Rejected</option>
							<option value="banned">Banned</option>
						</select>
					</div>
				</div>

				<div>
					<label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Tariff Plan</label>
					<select
						bind:value={editCustomerData.tariff_plan_id}
						class="w-full rounded-lg border border-gray-300 px-3 py-1.5 text-xs dark:border-gray-700 dark:bg-gray-800 dark:text-white"
					>
						<option value={0}>No plan — Default Operator Pricing</option>
						{#each tariffPlans as tp}
							<option value={tp.id}>{tp.name}</option>
						{/each}
					</select>
				</div>

				<div class="pt-2 border-t border-gray-100 dark:border-gray-800 space-y-2">
					<label class="flex items-center gap-2 cursor-pointer">
						<input type="checkbox" bind:checked={editCustomerData.api_access} class="rounded text-emerald-600" />
						<span class="text-xs font-semibold text-gray-800 dark:text-gray-200">API Access Enabled</span>
					</label>
					<label class="flex items-center gap-2 cursor-pointer">
						<input type="checkbox" bind:checked={editCustomerData.sms_auto_approved} class="rounded text-emerald-600" />
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
					{editSubmitting ? 'Saving...' : 'Save Customer'}
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- ============================================================
     REQUEST RESUBMISSION MODAL
     ============================================================ -->
{#if resubmitModalOpen && selectedCustomer}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-gray-950/60 p-4 backdrop-blur-xs">
		<div class="w-full max-w-md rounded-2xl bg-white p-6 shadow-2xl ring-1 ring-gray-900/10 dark:bg-gray-900 dark:ring-white/10">
			<h3 class="text-base font-bold text-gray-900 dark:text-white">Request KYC Resubmission</h3>
			<p class="text-xs text-gray-500 mt-0.5">{selectedCustomer.name}</p>

			<div class="mt-4 space-y-3">
				<div>
					<label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Reason for Resubmission</label>
					<textarea
						bind:value={resubmitReason}
						rows="3"
						placeholder="e.g. Back side of NID is blurry, please provide a clearer photo."
						class="w-full rounded-lg border border-gray-300 p-2.5 text-xs dark:border-gray-700 dark:bg-gray-800 dark:text-white outline-none focus:border-emerald-500"
					></textarea>
				</div>

				<div>
					<label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Fields to Resubmit</label>
					<input
						type="text"
						bind:value={resubmitFields}
						placeholder="front_nid, back_nid, live_photo"
						class="w-full rounded-lg border border-gray-300 p-2 text-xs dark:border-gray-700 dark:bg-gray-800 dark:text-white"
					/>
				</div>
			</div>

			<div class="mt-6 flex items-center justify-end gap-2">
				<button
					type="button"
					onclick={() => (resubmitModalOpen = false)}
					class="rounded-lg border border-gray-300 px-3.5 py-2 text-xs font-semibold text-gray-700 dark:border-gray-700 dark:text-gray-300"
				>
					Cancel
				</button>
				<button
					type="button"
					disabled={resubmitSubmitting || !resubmitReason.trim()}
					onclick={submitResubmit}
					class="rounded-lg bg-orange-600 px-4 py-2 text-xs font-bold text-white shadow-xs hover:bg-orange-700 disabled:opacity-50"
				>
					{resubmitSubmitting ? 'Sending...' : 'Request Resubmission'}
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- ============================================================
     REJECT KYC MODAL
     ============================================================ -->
{#if rejectModalOpen && selectedCustomer}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-gray-950/60 p-4 backdrop-blur-xs">
		<div class="w-full max-w-md rounded-2xl bg-white p-6 shadow-2xl ring-1 ring-gray-900/10 dark:bg-gray-900 dark:ring-white/10">
			<h3 class="text-base font-bold text-rose-600">Reject KYC Verification</h3>
			<p class="text-xs text-gray-500 mt-0.5">{selectedCustomer.name}</p>

			<div class="mt-4 space-y-3">
				<div>
					<label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Rejection Reason *</label>
					<textarea
						bind:value={rejectReason}
						rows="3"
						placeholder="Provide official reason for rejection..."
						class="w-full rounded-lg border border-gray-300 p-2.5 text-xs dark:border-gray-700 dark:bg-gray-800 dark:text-white outline-none focus:border-rose-500"
					></textarea>
				</div>

				<div>
					<label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Internal Admin Remarks (Optional)</label>
					<input
						type="text"
						bind:value={rejectRemarks}
						placeholder="Visible to administrators only"
						class="w-full rounded-lg border border-gray-300 p-2 text-xs dark:border-gray-700 dark:bg-gray-800 dark:text-white"
					/>
				</div>
			</div>

			<div class="mt-6 flex items-center justify-end gap-2">
				<button
					type="button"
					onclick={() => (rejectModalOpen = false)}
					class="rounded-lg border border-gray-300 px-3.5 py-2 text-xs font-semibold text-gray-700 dark:border-gray-700 dark:text-gray-300"
				>
					Cancel
				</button>
				<button
					type="button"
					disabled={rejectSubmitting || !rejectReason.trim()}
					onclick={submitReject}
					class="rounded-lg bg-rose-600 px-4 py-2 text-xs font-bold text-white shadow-xs hover:bg-rose-700 disabled:opacity-50"
				>
					{rejectSubmitting ? 'Rejecting...' : 'Reject KYC'}
				</button>
			</div>
		</div>
	</div>
{/if}
