<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface FinancialStats {
		completed_deposits: number;
		customer_spend: number;
		gateway_cost: number;
		pending_deposits: number;
		pending_deposits_count: number;
		transactions_count: number;
		system_balance: number;
	}

	interface ProfitStats {
		period_revenue: number;
		period_cost: number;
		period_profit: number;
		period_margin: number;
		total_revenue: number;
		total_cost: number;
		total_profit: number;
		total_margin: number;
	}

	interface TrendPoint {
		date: string;
		total: number;
	}

	interface SpendingPoint {
		date: string;
		revenue: number;
		cost: number;
	}

	interface MethodPoint {
		method: string;
		count: number;
		total: number;
	}

	interface VolumePoint {
		date: string;
		credits: number;
		debits: number;
	}

	interface DepositItem {
		id: number;
		user_id: number;
		customer_name: string;
		customer_phone: string;
		amount: number;
		currency: string;
		payment_method: string;
		transaction_id: string | null;
		status: string;
		deposited_by: string | null;
		paid_at: string | null;
		created_at: string;
	}

	interface TransactionItem {
		id: number;
		user_id: number;
		customer_name: string;
		customer_phone: string;
		reference: string;
		direction: string;
		amount: number;
		balance_before: number;
		balance_after: number;
		description: string | null;
		created_at: string;
	}

	interface CustomerOption {
		id: number;
		name: string;
		phone: string;
		balance: number;
	}

	let loading = $state(true);
	let period = $state('30days');
	let financialStats = $state<FinancialStats>({
		completed_deposits: 0,
		customer_spend: 0,
		gateway_cost: 0,
		pending_deposits: 0,
		pending_deposits_count: 0,
		transactions_count: 0,
		system_balance: 0
	});
	let profitStats = $state<ProfitStats>({
		period_revenue: 0,
		period_cost: 0,
		period_profit: 0,
		period_margin: 0,
		total_revenue: 0,
		total_cost: 0,
		total_profit: 0,
		total_margin: 0
	});

	let depositTrend = $state<TrendPoint[]>([]);
	let spendingTrend = $state<SpendingPoint[]>([]);
	let methodsDist = $state<MethodPoint[]>([]);
	let txVolume = $state<VolumePoint[]>([]);
	let recentDeposits = $state<DepositItem[]>([]);
	let recentTransactions = $state<TransactionItem[]>([]);

	// Adjust Balance Modal
	let showAdjustModal = $state(false);
	let customersList = $state<CustomerOption[]>([]);
	let adjustUserId = $state<number | ''>('');
	let adjustDirection = $state<'credit' | 'debit'>('credit');
	let adjustAmount = $state<number | ''>('');
	let adjustDescription = $state('');
	let adjustSubmitting = $state(false);
	let adjustMessage = $state<string | null>(null);
	let adjustError = $state<string | null>(null);

	async function loadDashboard() {
		loading = true;
		try {
			const res = await supermanApi.get<{
				data: {
					financial_stats: FinancialStats;
					profit_stats: ProfitStats;
					deposit_trend: TrendPoint[];
					spending_trend: SpendingPoint[];
					methods_distribution: MethodPoint[];
					transaction_volume: VolumePoint[];
					recent_deposits: DepositItem[];
					recent_transactions: TransactionItem[];
				};
			}>(`/api/v1/superman/billing/dashboard-stats?period=${period}`);

			if (res.data) {
				financialStats = res.data.financial_stats;
				profitStats = res.data.profit_stats;
				depositTrend = res.data.deposit_trend;
				spendingTrend = res.data.spending_trend;
				methodsDist = res.data.methods_distribution;
				txVolume = res.data.transaction_volume;
				recentDeposits = res.data.recent_deposits;
				recentTransactions = res.data.recent_transactions;
			}
		} catch (err: unknown) {
			console.error('Failed to load billing dashboard:', err);
		} finally {
			loading = false;
		}
	}

	async function openAdjustModal() {
		showAdjustModal = true;
		adjustMessage = null;
		adjustError = null;
		adjustAmount = '';
		adjustDescription = '';
		if (customersList.length === 0) {
			try {
				const res = await supermanApi.get<{ data: any[] }>('/api/v1/superman/customers?per_page=100');
				customersList = res.data.map((c) => ({
					id: c.id,
					name: c.name,
					phone: c.phone || '',
					balance: c.balance || 0
				}));
			} catch (err) {
				console.error(err);
			}
		}
	}

	async function handleAdjustBalance(e: SubmitEvent) {
		e.preventDefault();
		if (!adjustUserId || !adjustAmount || Number(adjustAmount) <= 0 || !adjustDescription.trim()) {
			adjustError = 'Please fill out all required fields with a valid amount.';
			return;
		}

		adjustSubmitting = true;
		adjustError = null;
		adjustMessage = null;
		try {
			const res = await supermanApi.post<{ message: string }>('/api/v1/superman/billing/adjust-balance', {
				user_id: Number(adjustUserId),
				direction: adjustDirection,
				amount: Number(adjustAmount),
				description: adjustDescription.trim()
			});
			adjustMessage = res.message;
			setTimeout(() => {
				showAdjustModal = false;
				loadDashboard();
			}, 1500);
		} catch (err: unknown) {
			adjustError = err instanceof Error ? err.message : 'Adjustment failed.';
		} finally {
			adjustSubmitting = false;
		}
	}

	async function handleApproveDeposit(id: number) {
		if (!confirm('Are you sure you want to approve this deposit? Balance will be credited to the customer account.')) {
			return;
		}
		try {
			await supermanApi.post(`/api/v1/superman/deposits/${id}/approve`, {});
			loadDashboard();
		} catch (err: unknown) {
			alert(err instanceof Error ? err.message : 'Failed to approve deposit.');
		}
	}

	onMount(() => {
		loadDashboard();
	});

	function formatCurrency(val: number): string {
		return '৳' + (val || 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
	}

	function formatDateLabel(val: string): string {
		if (!val) return '';
		const clean = val.includes('T') ? val.split('T')[0] : val;
		const parts = clean.split('-');
		if (parts.length === 3) {
			const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
			const m = parseInt(parts[1], 10) - 1;
			const d = parts[2];
			if (m >= 0 && m < 12) {
				return `${months[m]} ${d}`;
			}
		}
		const d = new Date(val);
		if (!isNaN(d.getTime())) {
			return d.toLocaleDateString('en-US', { month: 'short', day: '2-digit' });
		}
		return val;
	}

	// Chart max calculations for clean relative SVG bars
	let maxDeposit = $derived(
		depositTrend.length > 0 ? Math.max(...depositTrend.map((d) => d.total), 100) : 100
	);
	let maxSpending = $derived(
		spendingTrend.length > 0
			? Math.max(...spendingTrend.map((s) => Math.max(s.revenue, s.cost)), 100)
			: 100
	);

	const exportDepositColumns: ExportColumn<DepositItem>[] = [
		{ header: 'Customer', getValue: (d) => d.customer_name },
		{ header: 'Phone', getValue: (d) => d.customer_phone || '—' },
		{ header: 'Amount (BDT)', getValue: (d) => `BDT ${d.amount.toFixed(2)}` },
		{ header: 'Method', getValue: (d) => d.payment_method },
		{ header: 'Trx ID / Ref', getValue: (d) => d.transaction_id || '—' },
		{ header: 'Status', getValue: (d) => d.status.toUpperCase() },
		{ header: 'Date', getValue: (d) => (d.created_at ? new Date(d.created_at).toLocaleString() : '—') }
	];

	const exportTransactionColumns: ExportColumn<TransactionItem>[] = [
		{ header: 'Date & Time', getValue: (t) => (t.created_at ? new Date(t.created_at).toLocaleString() : '—') },
		{ header: 'Customer', getValue: (t) => t.customer_name },
		{ header: 'Phone', getValue: (t) => t.customer_phone || '—' },
		{ header: 'Type', getValue: (t) => t.direction.toUpperCase() },
		{ header: 'Amount (BDT)', getValue: (t) => `BDT ${t.amount.toFixed(2)}` },
		{ header: 'Balance Before', getValue: (t) => `BDT ${t.balance_before.toFixed(2)}` },
		{ header: 'Balance After', getValue: (t) => `BDT ${t.balance_after.toFixed(2)}` },
		{ header: 'Description', getValue: (t) => t.description || '—' }
	];
</script>

<svelte:head>
	<title>Billing Dashboard — SMS Whiteinfo Administrator</title>
</svelte:head>

<div class="space-y-6">
	<!-- Page Header -->
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<div class="flex items-center gap-2.5">
				<div class="flex h-10 w-10 items-center justify-center rounded-xl bg-emerald-500/10 text-emerald-600 dark:bg-emerald-500/20 dark:text-emerald-400">
					<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
					</svg>
				</div>
				<div>
					<h1 class="text-xl font-bold tracking-tight text-gray-900 dark:text-white sm:text-2xl">
						Billing & Financial Control Center
					</h1>
					<p class="text-xs text-gray-500 dark:text-gray-400">
						Overview of deposits, customer spending, operator carrier costs, and system profit margins.
					</p>
				</div>
			</div>
		</div>

		<!-- Action Buttons -->
		<div class="flex flex-wrap items-center gap-2">
			<button
				type="button"
				onclick={openAdjustModal}
				class="inline-flex items-center gap-1.5 rounded-lg bg-emerald-600 px-3.5 py-2 text-xs font-semibold text-white shadow-xs transition hover:bg-emerald-500 focus:outline-hidden focus:ring-2 focus:ring-emerald-500/40"
			>
				<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
				</svg>
				Adjust Balance
			</button>

			<a
				href="/superman/billing/deposits?status=pending"
				class="inline-flex items-center gap-1.5 rounded-lg border border-amber-300 bg-amber-50 px-3 py-2 text-xs font-semibold text-amber-800 transition hover:bg-amber-100 dark:border-amber-800/60 dark:bg-amber-950/40 dark:text-amber-300"
			>
				<svg class="h-4 w-4 text-amber-600 dark:text-amber-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
				</svg>
				Pending Deposits
				{#if financialStats.pending_deposits_count > 0}
					<span class="inline-flex items-center rounded-full bg-amber-200 px-1.5 py-0.5 text-[10px] font-bold text-amber-900 dark:bg-amber-800 dark:text-amber-100">
						{financialStats.pending_deposits_count}
					</span>
				{/if}
			</a>

			<a
				href="/superman/billing/transactions"
				class="inline-flex items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs font-medium text-gray-700 transition hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300 dark:hover:bg-gray-800"
			>
				<svg class="h-4 w-4 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
				</svg>
				All Transactions
			</a>

			<a
				href="/superman/billing/tariff-plans"
				class="inline-flex items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs font-medium text-gray-700 transition hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300 dark:hover:bg-gray-800"
			>
				<svg class="h-4 w-4 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z" />
				</svg>
				Tariff Plans
			</a>
		</div>
	</div>

	<!-- Period Tabs Filter -->
	<div class="flex items-center justify-between border-b border-gray-200 pb-3 dark:border-gray-800">
		<div class="flex items-center gap-1 overflow-x-auto text-xs">
			{#each [
				{ id: 'today', label: 'Today' },
				{ id: 'yesterday', label: 'Yesterday' },
				{ id: '7days', label: 'Last 7 Days' },
				{ id: '30days', label: 'Last 30 Days' },
				{ id: 'this_month', label: 'This Month' },
				{ id: 'all', label: 'All Time' }
			] as p}
				<button
					type="button"
					onclick={() => {
						period = p.id;
						loadDashboard();
					}}
					class="rounded-lg px-3 py-1.5 font-medium transition-colors {period === p.id ? 'bg-gray-900 text-white dark:bg-white dark:text-gray-950 shadow-xs' : 'text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-800'}"
				>
					{p.label}
				</button>
			{/each}
		</div>

		{#if loading}
			<div class="flex items-center gap-1.5 text-xs text-gray-400">
				<div class="h-3.5 w-3.5 animate-spin rounded-full border-2 border-emerald-500 border-t-transparent"></div>
				<span>Updating metrics…</span>
			</div>
		{/if}
	</div>

	<!-- Financial Overview Stats (6 Cards) -->
	<div>
		<h2 class="mb-3 text-xs font-bold uppercase tracking-wider text-gray-400 dark:text-gray-500">
			Financial Flow & Deposits
		</h2>
		<div class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-6">
			<!-- Deposits Completed -->
			<div class="rounded-xl border border-gray-200 bg-white p-3.5 shadow-xs dark:border-gray-800 dark:bg-gray-900">
				<div class="flex items-center justify-between text-gray-500">
					<span class="text-[11px] font-medium uppercase tracking-wider">Deposits</span>
					<span class="rounded-md bg-emerald-50 p-1 text-emerald-600 dark:bg-emerald-950/60 dark:text-emerald-400">
						<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
						</svg>
					</span>
				</div>
				<p class="mt-2 text-base font-bold text-gray-900 dark:text-white">
					{formatCurrency(financialStats.completed_deposits)}
				</p>
				<p class="mt-1 text-[10px] text-emerald-600 dark:text-emerald-400">Completed in period</p>
			</div>

			<!-- Customer Spend -->
			<div class="rounded-xl border border-gray-200 bg-white p-3.5 shadow-xs dark:border-gray-800 dark:bg-gray-900">
				<div class="flex items-center justify-between text-gray-500">
					<span class="text-[11px] font-medium uppercase tracking-wider">Customer Spend</span>
					<span class="rounded-md bg-sky-50 p-1 text-sky-600 dark:bg-sky-950/60 dark:text-sky-400">
						<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 3h2l.4 2M7 13h10l4-8H5.4M7 13L5.4 5M7 13l-2.293 2.293c-.63.63-.184 1.707.707 1.707H17m0 0a2 2 0 100 4 2 2 0 000-4zm-8 2a2 2 0 11-4 0 2 2 0 014 0z" />
						</svg>
					</span>
				</div>
				<p class="mt-2 text-base font-bold text-gray-900 dark:text-white">
					{formatCurrency(financialStats.customer_spend)}
				</p>
				<p class="mt-1 text-[10px] text-sky-600 dark:text-sky-400">Message revenue</p>
			</div>

			<!-- Gateway Cost -->
			<div class="rounded-xl border border-gray-200 bg-white p-3.5 shadow-xs dark:border-gray-800 dark:bg-gray-900">
				<div class="flex items-center justify-between text-gray-500">
					<span class="text-[11px] font-medium uppercase tracking-wider">Gateway Cost</span>
					<span class="rounded-md bg-rose-50 p-1 text-rose-600 dark:bg-rose-950/60 dark:text-rose-400">
						<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 14l-7 7m0 0l-7-7m7 7V3" />
						</svg>
					</span>
				</div>
				<p class="mt-2 text-base font-bold text-gray-900 dark:text-white">
					{formatCurrency(financialStats.gateway_cost)}
				</p>
				<p class="mt-1 text-[10px] text-rose-500 dark:text-rose-400">Carrier / platform costs</p>
			</div>

			<!-- Pending Deposits -->
			<div class="rounded-xl border border-gray-200 bg-white p-3.5 shadow-xs dark:border-gray-800 dark:bg-gray-900">
				<div class="flex items-center justify-between text-gray-500">
					<span class="text-[11px] font-medium uppercase tracking-wider">Pending Deposits</span>
					<span class="rounded-md bg-amber-50 p-1 text-amber-600 dark:bg-amber-950/60 dark:text-amber-400">
						<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
					</span>
				</div>
				<p class="mt-2 text-base font-bold text-gray-900 dark:text-white">
					{formatCurrency(financialStats.pending_deposits)}
				</p>
				<p class="mt-1 text-[10px] text-amber-600 dark:text-amber-400">{financialStats.pending_deposits_count} awaiting approval</p>
			</div>

			<!-- Transactions Count -->
			<div class="rounded-xl border border-gray-200 bg-white p-3.5 shadow-xs dark:border-gray-800 dark:bg-gray-900">
				<div class="flex items-center justify-between text-gray-500">
					<span class="text-[11px] font-medium uppercase tracking-wider">Transactions</span>
					<span class="rounded-md bg-indigo-50 p-1 text-indigo-600 dark:bg-indigo-950/60 dark:text-indigo-400">
						<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
						</svg>
					</span>
				</div>
				<p class="mt-2 text-base font-bold text-gray-900 dark:text-white">
					{financialStats.transactions_count.toLocaleString()}
				</p>
				<p class="mt-1 text-[10px] text-gray-500 dark:text-gray-400">Ledger entries in period</p>
			</div>

			<!-- System Balance -->
			<div class="rounded-xl border border-gray-200 bg-white p-3.5 shadow-xs dark:border-gray-800 dark:bg-gray-900">
				<div class="flex items-center justify-between text-gray-500">
					<span class="text-[11px] font-medium uppercase tracking-wider">System Balance</span>
					<span class="rounded-md bg-emerald-50 p-1 text-emerald-600 dark:bg-emerald-950/60 dark:text-emerald-400">
						<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z" />
						</svg>
					</span>
				</div>
				<p class="mt-2 text-base font-bold text-emerald-600 dark:text-emerald-400">
					{formatCurrency(financialStats.system_balance)}
				</p>
				<p class="mt-1 text-[10px] text-gray-500 dark:text-gray-400">Total customer balance</p>
			</div>
		</div>
	</div>

	<!-- Profit & Margin Analysis (6 Cards) -->
	<div>
		<h2 class="mb-3 text-xs font-bold uppercase tracking-wider text-gray-400 dark:text-gray-500">
			Profitability & Margins
		</h2>
		<div class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-6">
			<!-- Period Revenue -->
			<div class="rounded-xl border border-gray-200 bg-white p-3.5 shadow-xs dark:border-gray-800 dark:bg-gray-900">
				<span class="text-[10px] font-medium text-gray-400">Period Revenue</span>
				<p class="mt-1 text-base font-bold text-gray-900 dark:text-white">
					{formatCurrency(profitStats.period_revenue)}
				</p>
				<p class="text-[10px] text-gray-400">Selected period</p>
			</div>

			<!-- Period Cost -->
			<div class="rounded-xl border border-gray-200 bg-white p-3.5 shadow-xs dark:border-gray-800 dark:bg-gray-900">
				<span class="text-[10px] font-medium text-gray-400">Period Cost</span>
				<p class="mt-1 text-base font-bold text-rose-600 dark:text-rose-400">
					{formatCurrency(profitStats.period_cost)}
				</p>
				<p class="text-[10px] text-rose-400">Gateway deductions</p>
			</div>

			<!-- Period Profit & Margin -->
			<div class="rounded-xl border border-gray-200 bg-white p-3.5 shadow-xs dark:border-gray-800 dark:bg-gray-900">
				<span class="text-[10px] font-medium text-gray-400">Period Profit</span>
				<p class="mt-1 text-base font-bold {profitStats.period_profit >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-rose-600'}">
					{formatCurrency(profitStats.period_profit)}
				</p>
				<p class="text-[10px] font-semibold text-emerald-600 dark:text-emerald-400">
					{profitStats.period_margin.toFixed(1)}% margin
				</p>
			</div>

			<!-- Total Revenue All Time -->
			<div class="rounded-xl border border-gray-200 bg-white p-3.5 shadow-xs dark:border-gray-800 dark:bg-gray-900">
				<span class="text-[10px] font-medium text-gray-400">Total Revenue</span>
				<p class="mt-1 text-base font-bold text-gray-900 dark:text-white">
					{formatCurrency(profitStats.total_revenue)}
				</p>
				<p class="text-[10px] text-gray-400">All-time lifetime</p>
			</div>

			<!-- Total Cost All Time -->
			<div class="rounded-xl border border-gray-200 bg-white p-3.5 shadow-xs dark:border-gray-800 dark:bg-gray-900">
				<span class="text-[10px] font-medium text-gray-400">Total Cost</span>
				<p class="mt-1 text-base font-bold text-rose-600 dark:text-rose-400">
					{formatCurrency(profitStats.total_cost)}
				</p>
				<p class="text-[10px] text-rose-400">All-time gateway</p>
			</div>

			<!-- Total Profit All Time -->
			<div class="rounded-xl border border-gray-200 bg-white p-3.5 shadow-xs dark:border-gray-800 dark:bg-gray-900">
				<span class="text-[10px] font-medium text-gray-400">Total Profit</span>
				<p class="mt-1 text-base font-bold text-emerald-600 dark:text-emerald-400">
					{formatCurrency(profitStats.total_profit)}
				</p>
				<p class="text-[10px] font-semibold text-emerald-600 dark:text-emerald-400">
					{profitStats.total_margin.toFixed(1)}% margin
				</p>
			</div>
		</div>
	</div>

	<!-- Visual Trends (2x2 Grid) -->
	<div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
		<!-- Deposit Trend Bar Chart -->
		<div class="rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
			<div class="mb-3 flex items-center justify-between">
				<div>
					<h3 class="text-xs font-bold uppercase tracking-wider text-gray-700 dark:text-gray-300">
						Deposit Trend
					</h3>
					<p class="text-[11px] text-gray-400">Daily customer deposit volume</p>
				</div>
				<span class="rounded-md bg-emerald-50 px-2 py-0.5 text-[10px] font-semibold text-emerald-600 dark:bg-emerald-950 dark:text-emerald-400">
					Deposits (BDT)
				</span>
			</div>

			{#if depositTrend.length === 0}
				<div class="flex h-44 items-center justify-center text-xs text-gray-400">
					No completed deposits in this time frame.
				</div>
			{:else}
				<div class="flex h-44 items-end gap-2 pt-4">
					{#each depositTrend as point}
						{@const heightPct = Math.max(6, Math.round((point.total / maxDeposit) * 100))}
						<div class="group relative flex h-full flex-1 flex-col items-center justify-end">
							<!-- Tooltip -->
							<div class="pointer-events-none absolute -top-8 hidden rounded-md bg-gray-900 px-2 py-1 text-[10px] font-medium text-white shadow-md group-hover:block dark:bg-gray-100 dark:text-gray-900 z-10 whitespace-nowrap">
								{formatDateLabel(point.date)}: {formatCurrency(point.total)}
							</div>
							<div class="flex h-32 w-full items-end justify-center">
								<div
									style="height: {heightPct}%; min-height: 4px;"
									class="w-full max-w-[28px] rounded-t-sm bg-emerald-500/80 transition-all hover:bg-emerald-500 dark:bg-emerald-600"
								></div>
							</div>
							<span class="mt-1 h-5 text-[10px] font-medium text-gray-500 dark:text-gray-400 truncate max-w-full">
								{formatDateLabel(point.date)}
							</span>
						</div>
					{/each}
				</div>
			{/if}
		</div>

		<!-- Spending Trend (Revenue vs Cost) -->
		<div class="rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
			<div class="mb-3 flex items-center justify-between">
				<div>
					<h3 class="text-xs font-bold uppercase tracking-wider text-gray-700 dark:text-gray-300">
						Spending Trend
					</h3>
					<p class="text-[11px] text-gray-400">Customer message revenue vs operator cost</p>
				</div>
				<div class="flex items-center gap-2 text-[10px]">
					<span class="flex items-center gap-1 font-medium text-sky-600 dark:text-sky-400">
						<span class="h-2 w-2 rounded-full bg-sky-500"></span> Revenue
					</span>
					<span class="flex items-center gap-1 font-medium text-rose-500">
						<span class="h-2 w-2 rounded-full bg-rose-500"></span> Carrier Cost
					</span>
				</div>
			</div>

			{#if spendingTrend.length === 0}
				<div class="flex h-44 items-center justify-center text-xs text-gray-400">
					No message spending in this period.
				</div>
			{:else}
				<div class="flex h-44 items-end gap-2 pt-4">
					{#each spendingTrend as point}
						{@const revPct = Math.max(6, Math.round((point.revenue / maxSpending) * 100))}
						{@const costPct = Math.max(6, Math.round((point.cost / maxSpending) * 100))}
						<div class="group relative flex h-full flex-1 flex-col items-center justify-end">
							<div class="pointer-events-none absolute -top-10 hidden rounded-md bg-gray-900 px-2 py-1 text-[10px] font-medium text-white shadow-md group-hover:block dark:bg-gray-100 dark:text-gray-900 z-10 whitespace-nowrap">
								{formatDateLabel(point.date)}<br/>Rev: {formatCurrency(point.revenue)} | Cost: {formatCurrency(point.cost)}
							</div>
							<div class="flex h-32 w-full items-end justify-center gap-1">
								<div
									style="height: {revPct}%; min-height: 4px;"
									class="w-1/2 max-w-[14px] rounded-t-xs bg-sky-500 transition-all hover:bg-sky-400"
									title="Revenue: {formatCurrency(point.revenue)}"
								></div>
								<div
									style="height: {costPct}%; min-height: 4px;"
									class="w-1/2 max-w-[14px] rounded-t-xs bg-rose-500 transition-all hover:bg-rose-400"
									title="Cost: {formatCurrency(point.cost)}"
								></div>
							</div>
							<span class="mt-1 h-5 text-[10px] font-medium text-gray-500 dark:text-gray-400 truncate max-w-full">
								{formatDateLabel(point.date)}
							</span>
						</div>
					{/each}
				</div>
			{/if}
		</div>

		<!-- Payment Methods Distribution -->
		<div class="rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
			<div class="mb-3">
				<h3 class="text-xs font-bold uppercase tracking-wider text-gray-700 dark:text-gray-300">
					Deposits by Payment Method
				</h3>
				<p class="text-[11px] text-gray-400">Distribution of successful payment channels</p>
			</div>

			{#if methodsDist.length === 0}
				<div class="flex h-36 items-center justify-center text-xs text-gray-400">
					No deposit method data available.
				</div>
			{:else}
				<div class="space-y-3 pt-2">
					{#each methodsDist as m}
						{@const totalAll = methodsDist.reduce((acc, x) => acc + x.total, 0) || 1}
						{@const pct = Math.round((m.total / totalAll) * 100)}
						<div>
							<div class="flex items-center justify-between text-xs">
								<span class="font-semibold capitalize text-gray-800 dark:text-gray-200">{m.method}</span>
								<span class="font-mono text-gray-500">{formatCurrency(m.total)} ({m.count} txns, {pct}%)</span>
							</div>
							<div class="mt-1 h-2 w-full overflow-hidden rounded-full bg-gray-100 dark:bg-gray-800">
								<div
									style="width: {pct}%"
									class="h-full rounded-full bg-emerald-500 transition-all"
								></div>
							</div>
						</div>
					{/each}
				</div>
			{/if}
		</div>

		<!-- Transaction Volume -->
		<div class="rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
			<div class="mb-3 flex items-center justify-between">
				<div>
					<h3 class="text-xs font-bold uppercase tracking-wider text-gray-700 dark:text-gray-300">
						Transaction Volume
					</h3>
					<p class="text-[11px] text-gray-400">Credit vs Debit ledger operations</p>
				</div>
				<div class="flex items-center gap-2 text-[10px]">
					<span class="flex items-center gap-1 font-medium text-emerald-600 dark:text-emerald-400">
						<span class="h-2 w-2 rounded-full bg-emerald-500"></span> Credits
					</span>
					<span class="flex items-center gap-1 font-medium text-rose-500">
						<span class="h-2 w-2 rounded-full bg-rose-500"></span> Debits
					</span>
				</div>
			</div>

			{#if txVolume.length === 0}
				<div class="flex h-36 items-center justify-center text-xs text-gray-400">
					No transactions recorded in this period.
				</div>
			{:else}
				<div class="space-y-2 pt-2">
					{#each txVolume.slice(-6) as v}
						<div class="flex items-center justify-between border-b border-gray-100 py-1 text-xs dark:border-gray-800/60">
							<span class="font-mono text-gray-500">{formatDateLabel(v.date)}</span>
							<div class="flex items-center gap-4">
								<span class="font-mono font-semibold text-emerald-600 dark:text-emerald-400">
									+{v.credits} credits
								</span>
								<span class="font-mono font-semibold text-rose-600 dark:text-rose-400">
									-{v.debits} debits
								</span>
							</div>
						</div>
					{/each}
				</div>
			{/if}
		</div>
	</div>

	<!-- Recent Activity Tables (Stacked Full-Width for Clean Layout) -->
	<div class="space-y-6">
		<!-- Recent Deposits -->
		<div class="rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
			<div class="mb-3 flex items-center justify-between">
				<div>
					<h3 class="text-xs font-bold uppercase tracking-wider text-gray-700 dark:text-gray-300">
						Recent Deposits
					</h3>
					<p class="text-[11px] text-gray-400">Latest customer balance top-ups</p>
				</div>
				<div class="flex items-center gap-2">
					<ExportDropdown
						title="Recent Deposits"
						columns={exportDepositColumns}
						getData={() => recentDeposits}
						disabled={recentDeposits.length === 0}
					/>
					<a href="/superman/billing/deposits" class="text-xs font-semibold text-emerald-600 hover:underline dark:text-emerald-400">
						View All →
					</a>
				</div>
			</div>

			<div class="overflow-x-auto">
				<table class="w-full text-left text-xs">
					<thead>
						<tr class="border-b border-gray-100 text-[10px] font-semibold text-gray-400 uppercase dark:border-gray-800">
							<th class="py-2">Customer</th>
							<th class="py-2 text-right">Amount</th>
							<th class="py-2">Method</th>
							<th class="py-2">Trx ID / Ref</th>
							<th class="py-2">Status</th>
							<th class="py-2 text-right">Actions</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
						{#each recentDeposits as d}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/40">
								<td class="py-2.5">
									<a href="/superman/customers/{d.user_id}" class="font-semibold text-gray-900 hover:text-emerald-600 dark:text-white dark:hover:text-emerald-400">
										{d.customer_name}
									</a>
									<p class="font-mono text-[10px] text-gray-400">{d.customer_phone}</p>
								</td>
								<td class="py-2.5 text-right font-mono font-bold text-emerald-600 dark:text-emerald-400">
									{formatCurrency(d.amount)}
								</td>
								<td class="py-2.5">
									<span class="inline-flex items-center rounded-md px-1.5 py-0.5 text-[10px] font-medium capitalize bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200">
										{d.payment_method}
									</span>
								</td>
								<td class="py-2.5 font-mono text-[10px] text-gray-500 dark:text-gray-400 max-w-[130px] truncate" title={d.transaction_id || '—'}>
									{d.transaction_id || '—'}
								</td>
								<td class="py-2.5">
									<span class="inline-flex rounded-full px-2 py-0.5 text-[9px] font-semibold {d.status === 'completed' ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300' : 'bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-300'}">
										{d.status}
									</span>
								</td>
								<td class="py-2.5 text-right whitespace-nowrap">
									<div class="flex items-center justify-end gap-1.5">
										{#if d.status === 'pending'}
											<button
												type="button"
												onclick={() => handleApproveDeposit(d.id)}
												class="inline-flex items-center gap-1 rounded-md bg-emerald-50 px-2 py-1 text-[10px] font-semibold text-emerald-700 transition hover:bg-emerald-100 dark:bg-emerald-950/60 dark:text-emerald-300 dark:hover:bg-emerald-900/60"
												title="Approve Deposit"
											>
												<svg class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7" />
												</svg>
												Approve
											</button>
										{/if}
										<a
											href="/superman/billing/deposits"
											class="inline-flex items-center gap-1 rounded-md border border-gray-200 px-2 py-1 text-[10px] font-medium text-gray-600 transition hover:bg-gray-50 dark:border-gray-800 dark:text-gray-300 dark:hover:bg-gray-800"
											title="View All Deposits"
										>
											<svg class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
											</svg>
											View
										</a>
									</div>
								</td>
							</tr>
						{:else}
							<tr>
								<td colspan="6" class="py-4 text-center text-xs text-gray-400">No recent deposits found.</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>

		<!-- Recent Ledger Transactions -->
		<div class="rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
			<div class="mb-3 flex items-center justify-between">
				<div>
					<h3 class="text-xs font-bold uppercase tracking-wider text-gray-700 dark:text-gray-300">
						Recent Ledger Transactions
					</h3>
					<p class="text-[11px] text-gray-400">Latest credit & debit audit ledger</p>
				</div>
				<div class="flex items-center gap-2">
					<ExportDropdown
						title="Recent Transactions Ledger"
						columns={exportTransactionColumns}
						getData={() => recentTransactions}
						disabled={recentTransactions.length === 0}
					/>
					<a href="/superman/billing/transactions" class="text-xs font-semibold text-emerald-600 hover:underline dark:text-emerald-400">
						View All →
					</a>
				</div>
			</div>

			<div class="overflow-x-auto">
				<table class="w-full text-left text-xs">
					<thead>
						<tr class="border-b border-gray-100 text-[10px] font-semibold text-gray-400 uppercase dark:border-gray-800">
							<th class="py-2">Date & Time</th>
							<th class="py-2">Customer</th>
							<th class="py-2">Type</th>
							<th class="py-2 text-right">Amount</th>
							<th class="py-2 text-right">Before</th>
							<th class="py-2 text-right">After</th>
							<th class="py-2">Description</th>
							<th class="py-2 text-right">Action</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
						{#each recentTransactions as t}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/40">
								<td class="py-2.5 font-mono text-[10px] text-gray-400 whitespace-nowrap">
									{new Date(t.created_at).toLocaleDateString([], { day: '2-digit', month: 'short' })}, {new Date(t.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
								</td>
								<td class="py-2.5">
									<a href="/superman/customers/{t.user_id}" class="font-semibold text-gray-900 hover:text-emerald-600 dark:text-white dark:hover:text-emerald-400">
										{t.customer_name}
									</a>
									<p class="font-mono text-[10px] text-gray-400">{t.customer_phone}</p>
								</td>
								<td class="py-2.5">
									<span class="inline-flex rounded-md px-1.5 py-0.5 text-[9px] font-bold uppercase tracking-wider {t.direction === 'credit' ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300' : 'bg-rose-100 text-rose-800 dark:bg-rose-950 dark:text-rose-300'}">
										{t.direction}
									</span>
								</td>
								<td class="py-2.5 text-right font-mono font-bold {t.direction === 'credit' ? 'text-emerald-600 dark:text-emerald-400' : 'text-rose-600 dark:text-rose-400'}">
									{t.direction === 'credit' ? '+' : '-'}{formatCurrency(t.amount)}
								</td>
								<td class="py-2.5 text-right font-mono text-[11px] text-gray-400">
									{formatCurrency(t.balance_before)}
								</td>
								<td class="py-2.5 text-right font-mono font-semibold text-gray-900 dark:text-white">
									{formatCurrency(t.balance_after)}
								</td>
								<td class="py-2.5 max-w-[140px] truncate text-gray-600 dark:text-gray-300" title={t.description || ''}>
									{t.description || '—'}
								</td>
								<td class="py-2.5 text-right whitespace-nowrap">
									<a
										href="/superman/billing/transactions"
										class="inline-flex items-center gap-1 rounded-md border border-gray-200 px-2 py-1 text-[10px] font-medium text-gray-600 transition hover:bg-gray-50 dark:border-gray-800 dark:text-gray-300 dark:hover:bg-gray-800"
										title="View Full Ledger"
									>
										<svg class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
										</svg>
										View
									</a>
								</td>
							</tr>
						{:else}
							<tr>
								<td colspan="8" class="py-4 text-center text-xs text-gray-400">No recent ledger transactions found.</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>
	</div>
</div>

<!-- Modal: Adjust Balance -->
{#if showAdjustModal}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-gray-900/60 backdrop-blur-xs p-4">
		<div class="w-full max-w-lg rounded-2xl border border-gray-200 bg-white p-6 shadow-xl dark:border-gray-800 dark:bg-gray-900 animate-in fade-in zoom-in-95 duration-150">
			<div class="flex items-center justify-between border-b border-gray-100 pb-3 dark:border-gray-800">
				<div>
					<h3 class="text-base font-bold text-gray-900 dark:text-white">Manual Customer Balance Adjustment</h3>
					<p class="text-xs text-gray-500">Credit or deduct customer balance with a synchronous ledger entry.</p>
				</div>
				<button
					type="button"
					onclick={() => (showAdjustModal = false)}
					class="rounded-lg p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-gray-800"
				>
					<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>

			<form onsubmit={handleAdjustBalance} class="mt-4 space-y-4">
				{#if adjustMessage}
					<div class="rounded-lg bg-emerald-50 p-3 text-xs font-semibold text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300">
						{adjustMessage}
					</div>
				{/if}
				{#if adjustError}
					<div class="rounded-lg bg-rose-50 p-3 text-xs font-semibold text-rose-800 dark:bg-rose-950 dark:text-rose-300">
						{adjustError}
					</div>
				{/if}

				<!-- Customer Select -->
				<div>
					<label for="adjust-user" class="block text-xs font-semibold text-gray-700 dark:text-gray-300 mb-1">
						Customer <span class="text-rose-500">*</span>
					</label>
					<select
						id="adjust-user"
						bind:value={adjustUserId}
						required
						class="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-xs font-medium text-gray-900 focus:border-emerald-500 focus:outline-hidden dark:border-gray-700 dark:bg-gray-800 dark:text-white"
					>
						<option value="">— Select a customer —</option>
						{#each customersList as c}
							<option value={c.id}>
								{c.name} ({c.phone || 'No phone'}) — Current: {formatCurrency(c.balance)}
							</option>
						{/each}
					</select>
				</div>

				<!-- Adjustment Direction -->
				<div>
					<span class="block text-xs font-semibold text-gray-700 dark:text-gray-300 mb-1">
						Adjustment Type <span class="text-rose-500">*</span>
					</span>
					<div class="grid grid-cols-2 gap-3">
						<label class="flex items-center gap-2 rounded-lg border p-2.5 text-xs font-semibold cursor-pointer transition {adjustDirection === 'credit' ? 'border-emerald-500 bg-emerald-50/50 text-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-400' : 'border-gray-200 text-gray-700 dark:border-gray-800 dark:text-gray-300'}">
							<input
								type="radio"
								name="adjustDirection"
								value="credit"
								bind:group={adjustDirection}
								class="text-emerald-600"
							/>
							<span>Credit (Add Balance)</span>
						</label>
						<label class="flex items-center gap-2 rounded-lg border p-2.5 text-xs font-semibold cursor-pointer transition {adjustDirection === 'debit' ? 'border-rose-500 bg-rose-50/50 text-rose-700 dark:bg-rose-950/40 dark:text-rose-400' : 'border-gray-200 text-gray-700 dark:border-gray-800 dark:text-gray-300'}">
							<input
								type="radio"
								name="adjustDirection"
								value="debit"
								bind:group={adjustDirection}
								class="text-rose-600"
							/>
							<span>Debit (Deduct Balance)</span>
						</label>
					</div>
				</div>

				<!-- Amount -->
				<div>
					<label for="adjust-amt" class="block text-xs font-semibold text-gray-700 dark:text-gray-300 mb-1">
						Amount (BDT) <span class="text-rose-500">*</span>
					</label>
					<div class="relative">
						<span class="absolute inset-y-0 left-0 flex items-center pl-3 text-xs font-mono text-gray-400">৳</span>
						<input
							id="adjust-amt"
							type="number"
							step="0.01"
							min="0.01"
							bind:value={adjustAmount}
							required
							placeholder="0.00"
							class="w-full rounded-lg border border-gray-300 bg-white py-2 pl-8 pr-3 text-xs font-mono text-gray-900 focus:border-emerald-500 focus:outline-hidden dark:border-gray-700 dark:bg-gray-800 dark:text-white"
						/>
					</div>
				</div>

				<!-- Description / Reason -->
				<div>
					<label for="adjust-desc" class="block text-xs font-semibold text-gray-700 dark:text-gray-300 mb-1">
						Reason / Reference Note <span class="text-rose-500">*</span>
					</label>
					<textarea
						id="adjust-desc"
						bind:value={adjustDescription}
						required
						rows="2"
						placeholder="e.g. Manual bank deposit, special promotional credit, correction"
						class="w-full rounded-lg border border-gray-300 bg-white p-2.5 text-xs text-gray-900 focus:border-emerald-500 focus:outline-hidden dark:border-gray-700 dark:bg-gray-800 dark:text-white"
					></textarea>
				</div>

				<!-- Buttons -->
				<div class="flex items-center justify-end gap-2 pt-2 border-t border-gray-100 dark:border-gray-800">
					<button
						type="button"
						onclick={() => (showAdjustModal = false)}
						class="rounded-lg border border-gray-200 px-3.5 py-2 text-xs font-medium text-gray-700 hover:bg-gray-50 dark:border-gray-800 dark:text-gray-300 dark:hover:bg-gray-800"
					>
						Cancel
					</button>
					<button
						type="submit"
						disabled={adjustSubmitting}
						class="rounded-lg bg-emerald-600 px-4 py-2 text-xs font-semibold text-white shadow-xs hover:bg-emerald-500 disabled:opacity-50"
					>
						{adjustSubmitting ? 'Adjusting…' : 'Apply Adjustment'}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
