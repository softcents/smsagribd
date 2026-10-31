<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';

	interface StatCardData {
		revenue: number;
		cost: number;
		profit: number;
		margin: number;
		messages: number;
	}

	interface TrendItem {
		date: string;
		label: string;
		revenue: number;
		cost: number;
		profit: number;
	}

	interface GatewayBreakdownItem {
		gateway_id: number;
		gateway: string;
		driver: string;
		messages: number;
		revenue: number;
		operator_cost: number;
		primary_dipping_cost: number;
		third_party_dipping_cost: number | null;
		gateway_cost: number;
		cost: number;
		profit: number;
		margin: number;
		status: string;
		dlr_delivery_rate: number;
		dlr_support: boolean;
	}

	interface OperatorBreakdownItem {
		operator: string;
		country: string;
		messages: number;
		revenue: number;
		operator_cost: number;
		gateway_cost: number;
		cost: number;
		profit: number;
	}

	interface PerformanceItem {
		gateway: string;
		count: number;
	}

	interface TrafficItem {
		operator: string;
		count: number;
	}

	interface TopOperatorItem {
		operator: string;
		messages: number;
		revenue: number;
	}

	interface AnalyticsPayload {
		period_stats: StatCardData;
		overall_stats: StatCardData;
		profit_trend: TrendItem[];
		gateway_breakdown: GatewayBreakdownItem[];
		operator_breakdown: OperatorBreakdownItem[];
		gateway_performance: PerformanceItem[];
		operator_traffic: TrafficItem[];
		top_operators: TopOperatorItem[];
	}

	let period = $state('last_30_days');
	let loading = $state(true);
	let error = $state<string | null>(null);
	let data = $state<AnalyticsPayload | null>(null);

	// Filter & search states
	let gwSearch = $state('');
	let opSearch = $state('');
	let gwPage = $state(1);
	let opPage = $state(1);
	const perPage = 10;

	// Tooltip state for trend chart
	let hoveredIndex = $state<number | null>(null);

	const dateRangeOptions = [
		{ value: 'today', label: 'Today' },
		{ value: 'yesterday', label: 'Yesterday' },
		{ value: 'this_week', label: 'This Week' },
		{ value: 'last_7_days', label: 'Last 7 Days' },
		{ value: 'this_month', label: 'This Month' },
		{ value: 'last_30_days', label: 'Last 30 Days' },
		{ value: 'this_year', label: 'This Year' },
		{ value: 'all', label: 'All Time' }
	];

	async function loadAnalytics() {
		loading = true;
		error = null;
		try {
			const res = await supermanApi.get<AnalyticsPayload>(
				`/api/v1/superman/analytics/cost-profit?period=${encodeURIComponent(period)}`
			);
			data = res;
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Failed to load analytics data';
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		void loadAnalytics();
	});

	function handlePeriodChange(newPeriod: string) {
		period = newPeriod;
		void loadAnalytics();
	}

	// Filtered gateways
	const filteredGateways = $derived(
		(data?.gateway_breakdown || []).filter((g) =>
			g.gateway.toLowerCase().includes(gwSearch.toLowerCase()) ||
			g.driver.toLowerCase().includes(gwSearch.toLowerCase())
		)
	);

	const paginatedGateways = $derived(
		filteredGateways.slice((gwPage - 1) * perPage, gwPage * perPage)
	);

	const totalGwPages = $derived(Math.max(1, Math.ceil(filteredGateways.length / perPage)));

	// Filtered operators
	const filteredOperators = $derived(
		(data?.operator_breakdown || []).filter((o) =>
			o.operator.toLowerCase().includes(opSearch.toLowerCase()) ||
			o.country.toLowerCase().includes(opSearch.toLowerCase())
		)
	);

	const paginatedOperators = $derived(
		filteredOperators.slice((opPage - 1) * perPage, opPage * perPage)
	);

	const totalOpPages = $derived(Math.max(1, Math.ceil(filteredOperators.length / perPage)));

	// Export Columns
	const gatewayExportCols = [
		{ key: 'gateway', header: 'GATEWAY' },
		{ key: 'driver', header: 'DRIVER' },
		{ key: 'messages', header: 'MESSAGES' },
		{ key: 'revenue', header: 'REVENUE (BDT)' },
		{ key: 'operator_cost', header: 'OPERATOR COST (BDT)' },
		{ key: 'primary_dipping_cost', header: 'DIPPING COST (BDT)' },
		{ key: 'third_party_dipping_cost', header: '3RD-PARTY DIPPING' },
		{ key: 'cost', header: 'TOTAL COST (BDT)' },
		{ key: 'profit', header: 'NET PROFIT (BDT)' },
		{ key: 'margin', header: 'MARGIN (%)' },
		{ key: 'dlr_delivery_rate', header: 'DELIVERY (%)' }
	];

	const operatorExportCols = [
		{ key: 'operator', header: 'OPERATOR' },
		{ key: 'country', header: 'COUNTRY' },
		{ key: 'messages', header: 'MESSAGES' },
		{ key: 'revenue', header: 'REVENUE (BDT)' },
		{ key: 'operator_cost', header: 'OPERATOR COST (BDT)' },
		{ key: 'gateway_cost', header: 'GATEWAY COST (BDT)' },
		{ key: 'profit', header: 'PROFIT (BDT)' }
	];

	// Color palette for charts
	const palette = ['#10b981', '#0ea5e9', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899', '#6366f1', '#14b8a6'];

	// Chart dimension helpers
	const chartWidth = 900;
	const chartHeight = 280;
	const padding = { top: 20, right: 30, bottom: 40, left: 60 };
	const innerWidth = chartWidth - padding.left - padding.right;
	const innerHeight = chartHeight - padding.top - padding.bottom;

	const trendMaxVal = $derived.by(() => {
		if (!data?.profit_trend || data.profit_trend.length === 0) return 100;
		const max = Math.max(
			...data.profit_trend.map((d) => Math.max(d.revenue, d.cost, d.profit, 0))
		);
		return Math.max(max * 1.15, 10);
	});

	function getX(idx: number, total: number) {
		if (total <= 1) return padding.left + innerWidth / 2;
		return padding.left + (idx / (total - 1)) * innerWidth;
	}

	function getY(val: number) {
		const ratio = Math.max(0, val) / trendMaxVal;
		return padding.top + innerHeight - ratio * innerHeight;
	}

	function generatePath(items: TrendItem[], key: 'revenue' | 'cost' | 'profit'): string {
		if (!items || items.length === 0) return '';
		const points = items.map((item, idx) => ({
			x: getX(idx, items.length),
			y: getY(item[key])
		}));

		if (points.length === 1) {
			return `M ${points[0].x} ${points[0].y}`;
		}

		let path = `M ${points[0].x} ${points[0].y}`;
		for (let i = 0; i < points.length - 1; i++) {
			const p0 = points[i];
			const p1 = points[i + 1];
			const cp1x = p0.x + (p1.x - p0.x) / 3;
			const cp1y = p0.y;
			const cp2x = p0.x + (2 * (p1.x - p0.x)) / 3;
			const cp2y = p1.y;
			path += ` C ${cp1x} ${cp1y}, ${cp2x} ${cp2y}, ${p1.x} ${p1.y}`;
		}
		return path;
	}

	function generateAreaPath(items: TrendItem[], key: 'revenue' | 'cost' | 'profit'): string {
		if (!items || items.length === 0) return '';
		const linePath = generatePath(items, key);
		const baselineY = padding.top + innerHeight;
		const lastX = getX(items.length - 1, items.length);
		const firstX = getX(0, items.length);
		return `${linePath} L ${lastX} ${baselineY} L ${firstX} ${baselineY} Z`;
	}

	// Donut helpers
	function getDonutArcs(items: { label: string; count: number }[], total: number) {
		if (total === 0) return [];
		let currentAngle = 0;
		return items.map((item, idx) => {
			const slice = (item.count / total) * 360;
			const start = currentAngle;
			const end = currentAngle + slice;
			currentAngle = end;

			const startRad = ((start - 90) * Math.PI) / 180;
			const endRad = ((end - 90) * Math.PI) / 180;

			const rOuter = 70;
			const rInner = 44;
			const cx = 90;
			const cy = 90;

			const x1 = cx + rOuter * Math.cos(startRad);
			const y1 = cy + rOuter * Math.sin(startRad);
			const x2 = cx + rOuter * Math.cos(endRad);
			const y2 = cy + rOuter * Math.sin(endRad);

			const x3 = cx + rInner * Math.cos(endRad);
			const y3 = cy + rInner * Math.sin(endRad);
			const x4 = cx + rInner * Math.cos(startRad);
			const y4 = cy + rInner * Math.sin(startRad);

			const largeArc = slice > 180 ? 1 : 0;
			const d = `M ${x1} ${y1} A ${rOuter} ${rOuter} 0 ${largeArc} 1 ${x2} ${y2} L ${x3} ${y3} A ${rInner} ${rInner} 0 ${largeArc} 0 ${x4} ${y4} Z`;

			return {
				d,
				color: palette[idx % palette.length],
				label: item.label,
				count: item.count,
				percentage: ((item.count / total) * 100).toFixed(1)
			};
		});
	}

	const gwTotal = $derived(
		(data?.gateway_performance || []).reduce((sum, g) => sum + g.count, 0)
	);

	const gwArcs = $derived(
		getDonutArcs(
			(data?.gateway_performance || []).map((g) => ({ label: g.gateway, count: g.count })),
			gwTotal
		)
	);

	const opTotal = $derived(
		(data?.operator_traffic || []).reduce((sum, o) => sum + o.count, 0)
	);

	const opArcs = $derived(
		getDonutArcs(
			(data?.operator_traffic || []).map((o) => ({ label: o.operator, count: o.count })),
			opTotal
		)
	);
</script>

<svelte:head>
	<title>Cost & Profit Analytics — White Info Superman</title>
</svelte:head>

<div class="space-y-6">
	<!-- Top Bar -->
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h1 class="text-2xl font-bold tracking-tight text-gray-900 dark:text-white">Cost & Profit Analytics</h1>
			<p class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">
				Authoritative cost resolution and tariff-based revenue calculation.
			</p>
		</div>

		<!-- Date Range Filter -->
		<div class="flex items-center gap-3">
			<div class="flex items-center gap-2">
				<label for="date-range-select" class="text-xs font-semibold text-gray-600 dark:text-gray-300">Date Range</label>
				<select
					id="date-range-select"
					value={period}
					onchange={(e) => handlePeriodChange(e.currentTarget.value)}
					class="rounded-lg border border-gray-200 bg-white px-3 py-1.5 text-xs font-medium text-gray-700 shadow-xs focus:border-emerald-500 focus:outline-none focus:ring-1 focus:ring-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-200"
				>
					{#each dateRangeOptions as opt}
						<option value={opt.value}>{opt.label}</option>
					{/each}
				</select>
			</div>

			<button
				onclick={loadAnalytics}
				disabled={loading}
				class="inline-flex items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-3 py-1.5 text-xs font-medium text-gray-700 shadow-xs hover:bg-gray-50 disabled:opacity-50 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-200"
			>
				<svg class="h-3.5 w-3.5 {loading ? 'animate-spin' : ''}" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
				</svg>
				Refresh
			</button>
		</div>
	</div>

	{#if loading && !data}
		<div class="flex flex-col items-center justify-center py-24 text-gray-500">
			<svg class="h-8 w-8 animate-spin text-emerald-600" fill="none" viewBox="0 0 24 24">
				<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
				<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
			</svg>
			<p class="mt-3 text-xs font-medium">Resolving tariff revenue and gateway costs...</p>
		</div>
	{:else if error}
		<div class="rounded-xl border border-rose-200 bg-rose-50 p-6 text-center text-xs text-rose-700 dark:border-rose-900/50 dark:bg-rose-950/20 dark:text-rose-400">
			<p class="font-semibold">{error}</p>
			<button onclick={loadAnalytics} class="mt-3 rounded-md bg-rose-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-rose-500">Try Again</button>
		</div>
	{:else if data}
		<!-- 6 KPI Stats Overview Cards -->
		<div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
			<!-- Period Revenue -->
			<div class="rounded-xl border border-gray-100 bg-white p-5 shadow-xs transition hover:shadow-md dark:border-gray-800 dark:bg-gray-900">
				<div class="flex items-center justify-between">
					<span class="text-xs font-medium text-gray-500 dark:text-gray-400">Period Revenue</span>
					<div class="flex h-8 w-8 items-center justify-center rounded-lg bg-emerald-50 text-emerald-600 dark:bg-emerald-950/40">
						<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 9V7a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2m2 4h10a2 2 0 002-2v-6a2 2 0 00-2-2H9a2 2 0 00-2 2v6a2 2 0 002 2zm7-5a2 2 0 11-4 0 2 2 0 014 0z" />
						</svg>
					</div>
				</div>
				<p class="mt-3 text-2xl font-bold tracking-tight text-gray-950 dark:text-white">
					{data.period_stats.revenue.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })} BDT
				</p>
				<p class="mt-1.5 text-xs text-emerald-600 dark:text-emerald-400 font-medium">
					{data.period_stats.messages.toLocaleString()} messages
				</p>
			</div>

			<!-- Period Cost -->
			<div class="rounded-xl border border-gray-100 bg-white p-5 shadow-xs transition hover:shadow-md dark:border-gray-800 dark:bg-gray-900">
				<div class="flex items-center justify-between">
					<span class="text-xs font-medium text-gray-500 dark:text-gray-400">Period Cost</span>
					<div class="flex h-8 w-8 items-center justify-center rounded-lg bg-rose-50 text-rose-600 dark:bg-rose-950/40">
						<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 17h8m0 0V9m0 8l-8-8-4 4-6-6" />
						</svg>
					</div>
				</div>
				<p class="mt-3 text-2xl font-bold tracking-tight text-gray-950 dark:text-white">
					{data.period_stats.cost.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })} BDT
				</p>
				<p class="mt-1.5 text-xs text-rose-600 dark:text-rose-400 font-medium">
					Gateway cost
				</p>
			</div>

			<!-- Period Profit -->
			<div class="rounded-xl border border-gray-100 bg-white p-5 shadow-xs transition hover:shadow-md dark:border-gray-800 dark:bg-gray-900">
				<div class="flex items-center justify-between">
					<span class="text-xs font-medium text-gray-500 dark:text-gray-400">Period Profit</span>
					<div class="flex h-8 w-8 items-center justify-center rounded-lg {data.period_stats.profit >= 0 ? 'bg-emerald-50 text-emerald-600 dark:bg-emerald-950/40' : 'bg-rose-50 text-rose-600 dark:bg-rose-950/40'}">
						<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
						</svg>
					</div>
				</div>
				<p class="mt-3 text-2xl font-bold tracking-tight text-gray-950 dark:text-white">
					{data.period_stats.profit.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })} BDT
				</p>
				<p class="mt-1.5 text-xs font-medium {data.period_stats.profit >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-rose-600 dark:text-rose-400'}">
					{data.period_stats.margin.toFixed(1)}% margin
				</p>
			</div>

			<!-- Total Revenue -->
			<div class="rounded-xl border border-gray-100 bg-white p-5 shadow-xs transition hover:shadow-md dark:border-gray-800 dark:bg-gray-900">
				<div class="flex items-center justify-between">
					<span class="text-xs font-medium text-gray-500 dark:text-gray-400">Total Revenue</span>
					<div class="flex h-8 w-8 items-center justify-center rounded-lg bg-emerald-50 text-emerald-600 dark:bg-emerald-950/40">
						<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
					</div>
				</div>
				<p class="mt-3 text-2xl font-bold tracking-tight text-gray-950 dark:text-white">
					{data.overall_stats.revenue.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })} BDT
				</p>
				<p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
					All time
				</p>
			</div>

			<!-- Total Cost -->
			<div class="rounded-xl border border-gray-100 bg-white p-5 shadow-xs transition hover:shadow-md dark:border-gray-800 dark:bg-gray-900">
				<div class="flex items-center justify-between">
					<span class="text-xs font-medium text-gray-500 dark:text-gray-400">Total Cost</span>
					<div class="flex h-8 w-8 items-center justify-center rounded-lg bg-rose-50 text-rose-600 dark:bg-rose-950/40">
						<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12H9m12 0a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
					</div>
				</div>
				<p class="mt-3 text-2xl font-bold tracking-tight text-gray-950 dark:text-white">
					{data.overall_stats.cost.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })} BDT
				</p>
				<p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
					All time
				</p>
			</div>

			<!-- Total Profit -->
			<div class="rounded-xl border border-gray-100 bg-white p-5 shadow-xs transition hover:shadow-md dark:border-gray-800 dark:bg-gray-900">
				<div class="flex items-center justify-between">
					<span class="text-xs font-medium text-gray-500 dark:text-gray-400">Total Profit</span>
					<div class="flex h-8 w-8 items-center justify-center rounded-lg {data.overall_stats.profit >= 0 ? 'bg-emerald-50 text-emerald-600 dark:bg-emerald-950/40' : 'bg-rose-50 text-rose-600 dark:bg-rose-950/40'}">
						<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 3v4M3 5h4M6 17v4m-2-2h4m5-16l2.286 6.857L21 12l-5.714 2.143L13 21l-2.286-6.857L5 12l5.714-2.143L13 3z" />
						</svg>
					</div>
				</div>
				<p class="mt-3 text-2xl font-bold tracking-tight text-gray-950 dark:text-white">
					{data.overall_stats.profit.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })} BDT
				</p>
				<p class="mt-1.5 text-xs font-medium {data.overall_stats.profit >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-rose-600 dark:text-rose-400'}">
					{data.overall_stats.margin.toFixed(1)}% margin
				</p>
			</div>
		</div>

		<!-- Profit Trend Chart Section -->
		<div class="rounded-xl border border-gray-100 bg-white p-6 shadow-xs dark:border-gray-800 dark:bg-gray-900">
			<div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
				<div>
					<h2 class="text-base font-bold text-gray-900 dark:text-white">Profit Trend</h2>
					<p class="text-xs text-gray-500 dark:text-gray-400">Daily revenue, cost, and profit</p>
				</div>
				<!-- Chart Legend -->
				<div class="flex items-center gap-4 text-xs font-medium">
					<div class="flex items-center gap-1.5">
						<span class="h-2.5 w-2.5 rounded-full bg-emerald-500"></span>
						<span class="text-gray-700 dark:text-gray-300">Revenue</span>
					</div>
					<div class="flex items-center gap-1.5">
						<span class="h-2.5 w-2.5 rounded-full bg-rose-500"></span>
						<span class="text-gray-700 dark:text-gray-300">Cost</span>
					</div>
					<div class="flex items-center gap-1.5">
						<span class="h-2.5 w-2.5 rounded-full bg-sky-500"></span>
						<span class="text-gray-700 dark:text-gray-300">Profit</span>
					</div>
				</div>
			</div>

			<!-- SVG Chart Container -->
			<div class="relative mt-6 overflow-x-auto">
				<svg
					viewBox="0 0 {chartWidth} {chartHeight}"
					class="w-full h-auto select-none overflow-visible"
					onmouseleave={() => (hoveredIndex = null)}
				>
					<defs>
						<linearGradient id="revGrad" x1="0" y1="0" x2="0" y2="1">
							<stop offset="0%" stop-color="#10b981" stop-opacity="0.25" />
							<stop offset="100%" stop-color="#10b981" stop-opacity="0.0" />
						</linearGradient>
						<linearGradient id="costGrad" x1="0" y1="0" x2="0" y2="1">
							<stop offset="0%" stop-color="#ef4444" stop-opacity="0.25" />
							<stop offset="100%" stop-color="#ef4444" stop-opacity="0.0" />
						</linearGradient>
						<linearGradient id="profGrad" x1="0" y1="0" x2="0" y2="1">
							<stop offset="0%" stop-color="#0ea5e9" stop-opacity="0.25" />
							<stop offset="100%" stop-color="#0ea5e9" stop-opacity="0.0" />
						</linearGradient>
					</defs>

					<!-- Y Grid Lines & Labels -->
					{#each [0, 0.25, 0.5, 0.75, 1] as tickRatio}
						{@const yVal = padding.top + innerHeight * (1 - tickRatio)}
						{@const labelVal = Math.round(trendMaxVal * tickRatio)}
						<line
							x1={padding.left}
							y1={yVal}
							x2={chartWidth - padding.right}
							y2={yVal}
							stroke="currentColor"
							stroke-dasharray="3 3"
							class="text-gray-100 dark:text-gray-800"
							stroke-width="1"
						/>
						<text
							x={padding.left - 10}
							y={yVal + 4}
							text-anchor="end"
							class="text-[10px] fill-gray-400 dark:fill-gray-500 font-mono"
						>
							{labelVal}
						</text>
					{/each}

					<!-- X Axis Labels -->
					{#if data.profit_trend && data.profit_trend.length > 0}
						{#each data.profit_trend as item, idx}
							{#if idx % Math.max(1, Math.floor(data.profit_trend.length / 8)) === 0 || idx === data.profit_trend.length - 1}
								<text
									x={getX(idx, data.profit_trend.length)}
									y={chartHeight - 12}
									text-anchor="middle"
									class="text-[10px] fill-gray-400 dark:fill-gray-500 font-mono"
								>
									{item.label}
								</text>
							{/if}
						{/each}
					{/if}

					<!-- Area fills -->
					<path d={generateAreaPath(data.profit_trend, 'revenue')} fill="url(#revGrad)" />
					<path d={generateAreaPath(data.profit_trend, 'cost')} fill="url(#costGrad)" />
					<path d={generateAreaPath(data.profit_trend, 'profit')} fill="url(#profGrad)" />

					<!-- Lines -->
					<path d={generatePath(data.profit_trend, 'revenue')} fill="none" stroke="#10b981" stroke-width="2.5" stroke-linecap="round" />
					<path d={generatePath(data.profit_trend, 'cost')} fill="none" stroke="#ef4444" stroke-width="2.5" stroke-linecap="round" />
					<path d={generatePath(data.profit_trend, 'profit')} fill="none" stroke="#0ea5e9" stroke-width="2.5" stroke-linecap="round" />

					<!-- Interactive hover overlay columns -->
					{#if data.profit_trend}
						{#each data.profit_trend as item, idx}
							{@const x = getX(idx, data.profit_trend.length)}
							<rect
								x={x - (innerWidth / data.profit_trend.length) / 2}
								y={padding.top}
								width={innerWidth / data.profit_trend.length}
								height={innerHeight}
								fill="transparent"
								class="cursor-pointer"
								onmouseenter={() => (hoveredIndex = idx)}
							/>

							{#if hoveredIndex === idx}
								<line
									x1={x}
									y1={padding.top}
									x2={x}
									y2={padding.top + innerHeight}
									stroke="currentColor"
									class="text-gray-400 dark:text-gray-600"
									stroke-width="1.5"
									stroke-dasharray="2 2"
								/>
								<circle cx={x} cy={getY(item.revenue)} r="4" fill="#10b981" stroke="#fff" stroke-width="2" />
								<circle cx={x} cy={getY(item.cost)} r="4" fill="#ef4444" stroke="#fff" stroke-width="2" />
								<circle cx={x} cy={getY(item.profit)} r="4" fill="#0ea5e9" stroke="#fff" stroke-width="2" />
							{/if}
						{/each}
					{/if}
				</svg>

				<!-- Floating Tooltip -->
				{#if hoveredIndex !== null && data.profit_trend && data.profit_trend[hoveredIndex]}
					{@const hItem = data.profit_trend[hoveredIndex]}
					<div
						class="pointer-events-none absolute top-4 z-20 rounded-lg border border-gray-200 bg-white/95 p-2.5 text-xs shadow-lg backdrop-blur-xs dark:border-gray-700 dark:bg-gray-800/95"
						style="left: {Math.min(Math.max(10, (getX(hoveredIndex, data.profit_trend.length) / chartWidth) * 100 - 10), 80)}%"
					>
						<p class="font-bold text-gray-900 dark:text-white">{hItem.date}</p>
						<div class="mt-1 space-y-0.5 font-mono text-[11px]">
							<div class="flex items-center justify-between gap-3 text-emerald-600 dark:text-emerald-400">
								<span>Revenue:</span>
								<span class="font-bold">৳{hItem.revenue.toFixed(2)}</span>
							</div>
							<div class="flex items-center justify-between gap-3 text-rose-600 dark:text-rose-400">
								<span>Cost:</span>
								<span class="font-bold">৳{hItem.cost.toFixed(2)}</span>
							</div>
							<div class="flex items-center justify-between gap-3 text-sky-600 dark:text-sky-400">
								<span>Profit:</span>
								<span class="font-bold">৳{hItem.profit.toFixed(2)}</span>
							</div>
						</div>
					</div>
				{/if}
			</div>
		</div>

		<!-- Gateway Cost & Profit Breakdown Table Section -->
		<div class="rounded-xl border border-gray-100 bg-white p-6 shadow-xs dark:border-gray-800 dark:bg-gray-900 space-y-4">
			<div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
				<div>
					<h2 class="text-base font-bold text-gray-900 dark:text-white">Gateway Cost & Profit Breakdown</h2>
					<p class="text-xs text-gray-500 dark:text-gray-400">Auditable Revenue, Cost, Dipping, and Profit breakdown per gateway</p>
				</div>
				<div class="flex items-center gap-2">
					<div class="relative w-64">
						<input
							type="text"
							placeholder="Search by gateway..."
							bind:value={gwSearch}
							class="w-full rounded-lg border border-gray-200 bg-gray-50/50 py-1.5 pl-8 pr-3 text-xs text-gray-800 focus:border-emerald-500 focus:bg-white focus:outline-none focus:ring-1 focus:ring-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-200"
						/>
						<svg class="absolute left-2.5 top-2 h-3.5 w-3.5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
						</svg>
					</div>

					<ExportDropdown
						title="Gateway Cost & Profit Breakdown"
						columns={gatewayExportCols}
						getData={async () => filteredGateways}
					/>
				</div>
			</div>

			<!-- Gateways Table -->
			<div class="overflow-x-auto rounded-lg border border-gray-100 dark:border-gray-800">
				<table class="w-full text-left text-xs">
					<thead class="bg-gray-50/75 text-[11px] font-semibold uppercase tracking-wider text-gray-600 dark:bg-gray-800/50 dark:text-gray-400 border-b border-gray-200 dark:border-gray-800">
						<tr>
							<th class="py-3 px-4">Gateway</th>
							<th class="py-3 px-3">Driver</th>
							<th class="py-3 px-3 text-right">Messages</th>
							<th class="py-3 px-3 text-right">Revenue (Tariff)</th>
							<th class="py-3 px-3 text-right">Operator Cost</th>
							<th class="py-3 px-3 text-right">Dipping Cost</th>
							<th class="py-3 px-3 text-right">3rd-Party Dipping</th>
							<th class="py-3 px-3 text-right">Total Cost</th>
							<th class="py-3 px-3 text-right">Net Profit</th>
							<th class="py-3 px-3 text-right">Margin</th>
							<th class="py-3 px-4 text-right">Delivery %</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
						{#if paginatedGateways.length === 0}
							<tr>
								<td colspan="11" class="py-8 text-center text-gray-400">
									No gateway records found matching the period or search filter.
								</td>
							</tr>
						{:else}
							{#each paginatedGateways as gw}
								<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30 transition">
									<td class="py-3 px-4 font-bold text-gray-900 dark:text-white">
										{gw.gateway}
									</td>
									<td class="py-3 px-3">
										<span class="inline-flex items-center rounded-md bg-blue-50 px-2 py-0.5 text-[10px] font-medium text-blue-700 dark:bg-blue-950/40 dark:text-blue-300">
											{gw.driver}
										</span>
									</td>
									<td class="py-3 px-3 text-right font-medium text-gray-800 dark:text-gray-200">
										{gw.messages.toLocaleString()}
									</td>
									<td class="py-3 px-3 text-right">
										<span class="inline-flex rounded px-2 py-0.5 text-[11px] font-medium bg-emerald-50 text-emerald-700 dark:bg-emerald-950/30 dark:text-emerald-400">
											{gw.revenue.toFixed(2)} ৳
										</span>
									</td>
									<td class="py-3 px-3 text-right">
										<span class="inline-flex rounded px-2 py-0.5 text-[11px] font-medium bg-amber-50 text-amber-700 dark:bg-amber-950/30 dark:text-amber-400">
											{gw.operator_cost.toFixed(2)} ৳
										</span>
									</td>
									<td class="py-3 px-3 text-right">
										<span class="inline-flex rounded px-2 py-0.5 text-[11px] font-medium bg-sky-50 text-sky-700 dark:bg-sky-950/30 dark:text-sky-400">
											{gw.primary_dipping_cost.toFixed(2)} ৳
										</span>
									</td>
									<td class="py-3 px-3 text-right text-gray-400">
										{gw.third_party_dipping_cost !== null ? `${gw.third_party_dipping_cost.toFixed(2)} ৳` : '—'}
									</td>
									<td class="py-3 px-3 text-right font-bold text-gray-900 dark:text-white">
										{gw.cost.toFixed(2)} ৳
									</td>
									<td class="py-3 px-3 text-right">
										<span class="inline-flex rounded px-2 py-0.5 text-[11px] font-medium {gw.profit >= 0 ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-950/30 dark:text-emerald-400' : 'bg-rose-50 text-rose-700 dark:bg-rose-950/30 dark:text-rose-400'}">
											{gw.profit.toFixed(2)} ৳
										</span>
									</td>
									<td class="py-3 px-3 text-right font-medium text-gray-800 dark:text-gray-200">
										{gw.margin.toFixed(1)}%
									</td>
									<td class="py-3 px-4 text-right">
										{#if gw.dlr_support}
											<span class="font-medium text-gray-900 dark:text-white">{gw.dlr_delivery_rate.toFixed(1)}%</span>
										{:else}
											<span class="text-gray-400 text-[10px]">N/A</span>
										{/if}
									</td>
								</tr>
							{/each}
						{/if}
					</tbody>
				</table>
			</div>

			<!-- Pagination -->
			{#if totalGwPages > 1}
				<div class="flex items-center justify-between text-xs text-gray-500 pt-2">
					<span>Showing {(gwPage - 1) * perPage + 1} to {Math.min(gwPage * perPage, filteredGateways.length)} of {filteredGateways.length}</span>
					<div class="flex items-center gap-1">
						<button
							disabled={gwPage <= 1}
							onclick={() => gwPage--}
							class="rounded border border-gray-200 px-2.5 py-1 text-xs hover:bg-gray-50 disabled:opacity-40 dark:border-gray-700"
						>
							Prev
						</button>
						<span class="px-2 font-medium">{gwPage} / {totalGwPages}</span>
						<button
							disabled={gwPage >= totalGwPages}
							onclick={() => gwPage++}
							class="rounded border border-gray-200 px-2.5 py-1 text-xs hover:bg-gray-50 disabled:opacity-40 dark:border-gray-700"
						>
							Next
						</button>
					</div>
				</div>
			{/if}
		</div>

		<!-- Operator Cost Breakdown Table Section -->
		<div class="rounded-xl border border-gray-100 bg-white p-6 shadow-xs dark:border-gray-800 dark:bg-gray-900 space-y-4">
			<div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
				<div>
					<h2 class="text-base font-bold text-gray-900 dark:text-white">Operator Cost Breakdown</h2>
					<p class="text-xs text-gray-500 dark:text-gray-400">Revenue and cost per operator</p>
				</div>
				<div class="flex items-center gap-2">
					<div class="relative w-64">
						<input
							type="text"
							placeholder="Search by operator..."
							bind:value={opSearch}
							class="w-full rounded-lg border border-gray-200 bg-gray-50/50 py-1.5 pl-8 pr-3 text-xs text-gray-800 focus:border-emerald-500 focus:bg-white focus:outline-none focus:ring-1 focus:ring-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-200"
						/>
						<svg class="absolute left-2.5 top-2 h-3.5 w-3.5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
						</svg>
					</div>

					<ExportDropdown
						title="Operator Cost Breakdown"
						columns={operatorExportCols}
						getData={async () => filteredOperators}
					/>
				</div>
			</div>

			<!-- Operators Table -->
			<div class="overflow-x-auto rounded-lg border border-gray-100 dark:border-gray-800">
				<table class="w-full text-left text-xs">
					<thead class="bg-gray-50/75 text-[11px] font-semibold uppercase tracking-wider text-gray-600 dark:bg-gray-800/50 dark:text-gray-400 border-b border-gray-200 dark:border-gray-800">
						<tr>
							<th class="py-3 px-4">Operator</th>
							<th class="py-3 px-3">Country</th>
							<th class="py-3 px-3 text-right">Messages</th>
							<th class="py-3 px-3 text-right">Revenue</th>
							<th class="py-3 px-3 text-right">Operator Cost</th>
							<th class="py-3 px-3 text-right">Gateway Cost</th>
							<th class="py-3 px-4 text-right">Profit</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
						{#if paginatedOperators.length === 0}
							<tr>
								<td colspan="7" class="py-8 text-center text-gray-400">
									No operator records found matching the period or search filter.
								</td>
							</tr>
						{:else}
							{#each paginatedOperators as op}
								<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30 transition">
									<td class="py-3 px-4 font-bold text-gray-900 dark:text-white">
										{op.operator}
									</td>
									<td class="py-3 px-3 text-gray-600 dark:text-gray-300">
										{op.country}
									</td>
									<td class="py-3 px-3 text-right font-medium text-gray-800 dark:text-gray-200">
										{op.messages.toLocaleString()}
									</td>
									<td class="py-3 px-3 text-right font-medium text-gray-900 dark:text-white">
										{op.revenue.toFixed(2)} BDT
									</td>
									<td class="py-3 px-3 text-right font-medium text-rose-600 dark:text-rose-400">
										{op.operator_cost.toFixed(2)} BDT
									</td>
									<td class="py-3 px-3 text-right font-medium text-amber-600 dark:text-amber-400">
										{op.gateway_cost.toFixed(2)} BDT
									</td>
									<td class="py-3 px-4 text-right font-bold {op.profit >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-rose-600 dark:text-rose-400'}">
										{op.profit.toFixed(2)} BDT
									</td>
								</tr>
							{/each}
						{/if}
					</tbody>
				</table>
			</div>

			<!-- Pagination -->
			{#if totalOpPages > 1}
				<div class="flex items-center justify-between text-xs text-gray-500 pt-2">
					<span>Showing {(opPage - 1) * perPage + 1} to {Math.min(opPage * perPage, filteredOperators.length)} of {filteredOperators.length}</span>
					<div class="flex items-center gap-1">
						<button
							disabled={opPage <= 1}
							onclick={() => opPage--}
							class="rounded border border-gray-200 px-2.5 py-1 text-xs hover:bg-gray-50 disabled:opacity-40 dark:border-gray-700"
						>
							Prev
						</button>
						<span class="px-2 font-medium">{opPage} / {totalOpPages}</span>
						<button
							disabled={opPage >= totalOpPages}
							onclick={() => opPage++}
							class="rounded border border-gray-200 px-2.5 py-1 text-xs hover:bg-gray-50 disabled:opacity-40 dark:border-gray-700"
						>
							Next
						</button>
					</div>
				</div>
			{/if}
		</div>

		<!-- Bottom 3-Card Row: Gateway Performance, Operator Traffic, Top Operators -->
		<div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
			<!-- Gateway Performance Donut -->
			<div class="rounded-xl border border-gray-100 bg-white p-6 shadow-xs dark:border-gray-800 dark:bg-gray-900 flex flex-col">
				<h2 class="text-sm font-bold text-gray-900 dark:text-white">Gateway Performance</h2>
				<p class="text-xs text-gray-500 dark:text-gray-400">Messages per gateway</p>

				<div class="mt-4 flex flex-1 flex-col items-center justify-center">
					<div class="relative flex items-center justify-center">
						<svg width="180" height="180" viewBox="0 0 180 180">
							{#if gwArcs.length === 0}
								<circle cx="90" cy="90" r="57" fill="none" stroke="#e5e7eb" stroke-width="26" class="dark:stroke-gray-800" />
							{:else}
								{#each gwArcs as arc}
									<path d={arc.d} fill={arc.color} />
								{/each}
							{/if}
						</svg>
						<div class="pointer-events-none absolute flex flex-col items-center justify-center text-center">
							<span class="text-lg font-black text-gray-900 dark:text-white">{gwTotal.toLocaleString()}</span>
							<span class="text-[10px] text-gray-400 uppercase">Messages</span>
						</div>
					</div>

					<!-- Legend -->
					<div class="mt-4 w-full space-y-1.5 text-xs">
						{#each gwArcs as arc}
							<div class="flex items-center justify-between">
								<div class="flex items-center gap-2">
									<span class="h-2.5 w-2.5 rounded-full" style="background-color: {arc.color}"></span>
									<span class="text-gray-700 dark:text-gray-300 font-medium">{arc.label}</span>
								</div>
								<span class="font-mono text-gray-500 dark:text-gray-400">{arc.count} ({arc.percentage}%)</span>
							</div>
						{/each}
					</div>
				</div>
			</div>

			<!-- Operator Traffic Donut -->
			<div class="rounded-xl border border-gray-100 bg-white p-6 shadow-xs dark:border-gray-800 dark:bg-gray-900 flex flex-col">
				<h2 class="text-sm font-bold text-gray-900 dark:text-white">Operator Traffic</h2>
				<p class="text-xs text-gray-500 dark:text-gray-400">Messages per operator</p>

				<div class="mt-4 flex flex-1 flex-col items-center justify-center">
					<div class="relative flex items-center justify-center">
						<svg width="180" height="180" viewBox="0 0 180 180">
							{#if opArcs.length === 0}
								<circle cx="90" cy="90" r="57" fill="none" stroke="#e5e7eb" stroke-width="26" class="dark:stroke-gray-800" />
							{:else}
								{#each opArcs as arc}
									<path d={arc.d} fill={arc.color} />
								{/each}
							{/if}
						</svg>
						<div class="pointer-events-none absolute flex flex-col items-center justify-center text-center">
							<span class="text-lg font-black text-gray-900 dark:text-white">{opTotal.toLocaleString()}</span>
							<span class="text-[10px] text-gray-400 uppercase">Traffic</span>
						</div>
					</div>

					<!-- Legend -->
					<div class="mt-4 w-full space-y-1.5 text-xs">
						{#each opArcs as arc}
							<div class="flex items-center justify-between">
								<div class="flex items-center gap-2">
									<span class="h-2.5 w-2.5 rounded-full" style="background-color: {arc.color}"></span>
									<span class="text-gray-700 dark:text-gray-300 font-medium">{arc.label}</span>
								</div>
								<span class="font-mono text-gray-500 dark:text-gray-400">{arc.count} ({arc.percentage}%)</span>
							</div>
						{/each}
					</div>
				</div>
			</div>

			<!-- Top Operators Table -->
			<div class="rounded-xl border border-gray-100 bg-white p-6 shadow-xs dark:border-gray-800 dark:bg-gray-900 flex flex-col">
				<h2 class="text-sm font-bold text-gray-900 dark:text-white">Top Operators</h2>
				<p class="text-xs text-gray-500 dark:text-gray-400">Ranked by traffic & tariff volume</p>

				<div class="mt-4 flex-1 overflow-x-auto">
					<table class="w-full text-left text-xs">
						<thead class="bg-gray-50/75 text-[11px] font-semibold uppercase tracking-wider text-gray-600 dark:bg-gray-800/50 dark:text-gray-400 border-b border-gray-200 dark:border-gray-800">
							<tr>
								<th class="py-2.5 px-3">Operator</th>
								<th class="py-2.5 px-3 text-right">Messages</th>
								<th class="py-2.5 px-3 text-right">Revenue</th>
							</tr>
						</thead>
						<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
							{#if !data.top_operators || data.top_operators.length === 0}
								<tr>
									<td colspan="3" class="py-6 text-center text-gray-400">No operator data available</td>
								</tr>
							{:else}
								{#each data.top_operators as topOp}
									<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30">
										<td class="py-2.5 px-3 font-semibold text-gray-900 dark:text-white">{topOp.operator}</td>
										<td class="py-2.5 px-3 text-right text-gray-800 dark:text-gray-200">{topOp.messages.toLocaleString()}</td>
										<td class="py-2.5 px-3 text-right font-medium text-emerald-600 dark:text-emerald-400">{topOp.revenue.toFixed(2)} BDT</td>
									</tr>
								{/each}
							{/if}
						</tbody>
					</table>
				</div>
			</div>
		</div>
	{/if}
</div>
