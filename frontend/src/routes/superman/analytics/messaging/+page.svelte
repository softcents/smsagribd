<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';

	interface MessagingStats {
		total: number;
		successful: number;
		delivered: number;
		failed: number;
		success_rate: number;
		delivered_rate: number;
		failed_rate: number;
		total_cost: number;
		avg_parts: number;
		unicode: number;
		unicode_percent: number;
	}

	interface MessageTrendItem {
		date: string;
		label: string;
		count: number;
	}

	interface DeliveryRateTrendItem {
		date: string;
		label: string;
		rate: number;
	}

	interface SourceDistributionItem {
		source: string;
		count: number;
		percentage: number;
	}

	interface DlrSummaryItem {
		status: string;
		count: number;
		color: string;
	}

	interface FailedReasonItem {
		reason: string;
		count: number;
	}

	interface MessagingPayload {
		stats: MessagingStats;
		message_trends: MessageTrendItem[];
		delivery_rate_trends: DeliveryRateTrendItem[];
		source_distribution: SourceDistributionItem[];
		dlr_summary: DlrSummaryItem[];
		failed_reasons: FailedReasonItem[];
	}

	let period = $state('last_30_days');
	let loading = $state(true);
	let error = $state<string | null>(null);
	let data = $state<MessagingPayload | null>(null);

	// Tooltip states for trend charts
	let hoveredMsgIdx = $state<number | null>(null);
	let hoveredDlrIdx = $state<number | null>(null);

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
			const res = await supermanApi.get<MessagingPayload>(
				`/api/v1/superman/analytics/messaging?period=${encodeURIComponent(period)}`
			);
			data = res;
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Failed to load messaging analytics';
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

	const palette = ['#10b981', '#0ea5e9', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899'];

	// Chart dimensions
	const chartWidth = 500;
	const chartHeight = 220;
	const pad = { top: 20, right: 20, bottom: 35, left: 45 };
	const innerW = chartWidth - pad.left - pad.right;
	const innerH = chartHeight - pad.top - pad.bottom;

	// Message Trends Max
	const msgMax = $derived.by(() => {
		if (!data?.message_trends || data.message_trends.length === 0) return 10;
		const m = Math.max(...data.message_trends.map((d) => d.count));
		return Math.max(Math.ceil(m * 1.2), 5);
	});

	function getMsgX(idx: number, len: number) {
		if (len <= 1) return pad.left + innerW / 2;
		return pad.left + (idx / (len - 1)) * innerW;
	}

	function getMsgY(val: number) {
		return pad.top + innerH - (Math.max(0, val) / msgMax) * innerH;
	}

	function getDlrRateY(rate: number) {
		return pad.top + innerH - (Math.max(0, Math.min(100, rate)) / 100) * innerH;
	}

	function generateSinglePath(points: { x: number; y: number }[]): string {
		if (points.length === 0) return '';
		if (points.length === 1) return `M ${points[0].x} ${points[0].y}`;
		let p = `M ${points[0].x} ${points[0].y}`;
		for (let i = 0; i < points.length - 1; i++) {
			const p0 = points[i];
			const p1 = points[i + 1];
			const cp1x = p0.x + (p1.x - p0.x) / 3;
			const cp1y = p0.y;
			const cp2x = p0.x + (2 * (p1.x - p0.x)) / 3;
			const cp2y = p1.y;
			p += ` C ${cp1x} ${cp1y}, ${cp2x} ${cp2y}, ${p1.x} ${p1.y}`;
		}
		return p;
	}

	function generateSingleAreaPath(points: { x: number; y: number }[]): string {
		if (points.length === 0) return '';
		const line = generateSinglePath(points);
		const baselineY = pad.top + innerH;
		const lastX = points[points.length - 1].x;
		const firstX = points[0].x;
		return `${line} L ${lastX} ${baselineY} L ${firstX} ${baselineY} Z`;
	}

	// Donut helpers
	function getDonutArcs(items: { label: string; count: number; color?: string }[], total: number) {
		if (total === 0) return [];
		let currentAngle = 0;
		return items.map((item, idx) => {
			const slice = (item.count / total) * 360;
			const start = currentAngle;
			const end = currentAngle + slice;
			currentAngle = end;

			const startRad = ((start - 90) * Math.PI) / 180;
			const endRad = ((end - 90) * Math.PI) / 180;

			const rOuter = 65;
			const rInner = 40;
			const cx = 80;
			const cy = 80;

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
				color: item.color || palette[idx % palette.length],
				label: item.label,
				count: item.count,
				percentage: ((item.count / total) * 100).toFixed(1)
			};
		});
	}

	const sourceTotal = $derived(
		(data?.source_distribution || []).reduce((sum, s) => sum + s.count, 0)
	);

	const sourceArcs = $derived(
		getDonutArcs(
			(data?.source_distribution || []).map((s, idx) => ({
				label: s.source,
				count: s.count,
				color: palette[idx % palette.length]
			})),
			sourceTotal
		)
	);

	const dlrTotal = $derived(
		(data?.dlr_summary || []).reduce((sum, d) => sum + d.count, 0)
	);

	const dlrArcs = $derived(
		getDonutArcs(
			(data?.dlr_summary || []).map((d) => ({
				label: d.status,
				count: d.count,
				color: d.color
			})),
			dlrTotal
		)
	);
</script>

<svelte:head>
	<title>Messaging Analytics — White Info Superman</title>
</svelte:head>

<div class="space-y-6">
	<!-- Top Bar -->
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h1 class="text-2xl font-bold tracking-tight text-gray-900 dark:text-white">Messaging Analytics</h1>
			<p class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">
				High-frequency dispatch volume, delivery rates, and source telemetry.
			</p>
		</div>

		<!-- Date Range Filter -->
		<div class="flex items-center gap-3">
			<div class="flex items-center gap-2">
				<label for="msg-date-range-select" class="text-xs font-semibold text-gray-600 dark:text-gray-300">Date Range</label>
				<select
					id="msg-date-range-select"
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
			<p class="mt-3 text-xs font-medium">Compiling messaging analytics...</p>
		</div>
	{:else if error}
		<div class="rounded-xl border border-rose-200 bg-rose-50 p-6 text-center text-xs text-rose-700 dark:border-rose-900/50 dark:bg-rose-950/20 dark:text-rose-400">
			<p class="font-semibold">{error}</p>
			<button onclick={loadAnalytics} class="mt-3 rounded-md bg-rose-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-rose-500">Try Again</button>
		</div>
	{:else if data}
		<!-- 7 KPI Stats Cards -->
		<div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
			<!-- Total Messages -->
			<div class="rounded-xl border border-gray-100 bg-white p-5 shadow-xs transition hover:shadow-md dark:border-gray-800 dark:bg-gray-900">
				<div class="flex items-center justify-between">
					<span class="text-xs font-medium text-gray-500 dark:text-gray-400">Total Messages</span>
					<div class="flex h-8 w-8 items-center justify-center rounded-lg bg-blue-50 text-blue-600 dark:bg-blue-950/40">
						<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
						</svg>
					</div>
				</div>
				<p class="mt-3 text-2xl font-bold tracking-tight text-gray-950 dark:text-white">
					{data.stats.total.toLocaleString()}
				</p>
				<p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
					In selected period
				</p>
			</div>

			<!-- Success Rate -->
			<div class="rounded-xl border border-gray-100 bg-white p-5 shadow-xs transition hover:shadow-md dark:border-gray-800 dark:bg-gray-900">
				<div class="flex items-center justify-between">
					<span class="text-xs font-medium text-gray-500 dark:text-gray-400">Success Rate</span>
					<div class="flex h-8 w-8 items-center justify-center rounded-lg bg-emerald-50 text-emerald-600 dark:bg-emerald-950/40">
						<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4M7.835 4.697a3.42 3.42 0 001.946-.806 3.42 3.42 0 014.438 0 3.42 3.42 0 001.946.806 3.42 3.42 0 013.138 3.138 3.42 3.42 0 00.806 1.946 3.42 3.42 0 010 4.438 3.42 3.42 0 00-.806 1.946 3.42 3.42 0 01-3.138 3.138 3.42 3.42 0 00-1.946.806 3.42 3.42 0 01-4.438 0 3.42 3.42 0 00-1.946-.806 3.42 3.42 0 01-3.138-3.138 3.42 3.42 0 00-.806-1.946 3.42 3.42 0 010-4.438 3.42 3.42 0 00.806-1.946 3.42 3.42 0 013.138-3.138z" />
						</svg>
					</div>
				</div>
				<p class="mt-3 text-2xl font-bold tracking-tight text-emerald-600 dark:text-emerald-400">
					{data.stats.success_rate.toFixed(1)}%
				</p>
				<p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
					{data.stats.successful.toLocaleString()} sent or delivered
				</p>
			</div>

			<!-- Delivered -->
			<div class="rounded-xl border border-gray-100 bg-white p-5 shadow-xs transition hover:shadow-md dark:border-gray-800 dark:bg-gray-900">
				<div class="flex items-center justify-between">
					<span class="text-xs font-medium text-gray-500 dark:text-gray-400">Delivered</span>
					<div class="flex h-8 w-8 items-center justify-center rounded-lg bg-emerald-50 text-emerald-600 dark:bg-emerald-950/40">
						<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
						</svg>
					</div>
				</div>
				<p class="mt-3 text-2xl font-bold tracking-tight text-gray-950 dark:text-white">
					{data.stats.delivered.toLocaleString()}
				</p>
				<p class="mt-1.5 text-xs text-emerald-600 dark:text-emerald-400 font-medium">
					{data.stats.delivered_rate.toFixed(1)}% confirmed delivered
				</p>
			</div>

			<!-- Failed -->
			<div class="rounded-xl border border-gray-100 bg-white p-5 shadow-xs transition hover:shadow-md dark:border-gray-800 dark:bg-gray-900">
				<div class="flex items-center justify-between">
					<span class="text-xs font-medium text-gray-500 dark:text-gray-400">Failed</span>
					<div class="flex h-8 w-8 items-center justify-center rounded-lg bg-rose-50 text-rose-600 dark:bg-rose-950/40">
						<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
						</svg>
					</div>
				</div>
				<p class="mt-3 text-2xl font-bold tracking-tight text-rose-600 dark:text-rose-400">
					{data.stats.failed.toLocaleString()}
				</p>
				<p class="mt-1.5 text-xs text-rose-600 dark:text-rose-400 font-medium">
					{data.stats.failed_rate.toFixed(1)}% failure rate
				</p>
			</div>

			<!-- Total Cost -->
			<div class="rounded-xl border border-gray-100 bg-white p-5 shadow-xs transition hover:shadow-md dark:border-gray-800 dark:bg-gray-900">
				<div class="flex items-center justify-between">
					<span class="text-xs font-medium text-gray-500 dark:text-gray-400">Total Cost</span>
					<div class="flex h-8 w-8 items-center justify-center rounded-lg bg-amber-50 text-amber-600 dark:bg-amber-950/40">
						<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
					</div>
				</div>
				<p class="mt-3 text-2xl font-bold tracking-tight text-gray-950 dark:text-white">
					{data.stats.total_cost.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })} BDT
				</p>
				<p class="mt-1.5 text-xs text-amber-600 dark:text-amber-400 font-medium">
					Customer revenue
				</p>
			</div>

			<!-- Avg Parts -->
			<div class="rounded-xl border border-gray-100 bg-white p-5 shadow-xs transition hover:shadow-md dark:border-gray-800 dark:bg-gray-900">
				<div class="flex items-center justify-between">
					<span class="text-xs font-medium text-gray-500 dark:text-gray-400">Avg Parts</span>
					<div class="flex h-8 w-8 items-center justify-center rounded-lg bg-indigo-50 text-indigo-600 dark:bg-indigo-950/40">
						<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h7" />
						</svg>
					</div>
				</div>
				<p class="mt-3 text-2xl font-bold tracking-tight text-gray-950 dark:text-white">
					{data.stats.avg_parts.toFixed(1)}
				</p>
				<p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
					SMS segments per message
				</p>
			</div>

			<!-- Unicode -->
			<div class="rounded-xl border border-gray-100 bg-white p-5 shadow-xs transition hover:shadow-md dark:border-gray-800 dark:bg-gray-900 sm:col-span-2 lg:col-span-2">
				<div class="flex items-center justify-between">
					<span class="text-xs font-medium text-gray-500 dark:text-gray-400">Unicode</span>
					<div class="flex h-8 w-8 items-center justify-center rounded-lg bg-purple-50 text-purple-600 dark:bg-purple-950/40">
						<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 5h12M9 3v2m1.048 9.5A18.022 18.022 0 016.412 9m6.088 9h7M11 21l5-10 5 10M12.751 5C11.783 10.77 8.07 15.61 3 18.129" />
						</svg>
					</div>
				</div>
				<p class="mt-3 text-2xl font-bold tracking-tight text-gray-950 dark:text-white">
					{data.stats.unicode.toLocaleString()}
				</p>
				<p class="mt-1.5 text-xs text-purple-600 dark:text-purple-400 font-medium">
					{data.stats.unicode_percent.toFixed(1)}% of messages
				</p>
			</div>
		</div>

		<!-- Charts Row 1: Message Trends & Delivery Rate Trend -->
		<div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
			<!-- Message Trends -->
			<div class="rounded-xl border border-gray-100 bg-white p-6 shadow-xs dark:border-gray-800 dark:bg-gray-900">
				<div class="flex items-center justify-between">
					<div>
						<h2 class="text-sm font-bold text-gray-900 dark:text-white">Message Trends</h2>
						<p class="text-xs text-gray-500 dark:text-gray-400">Messages sent per day</p>
					</div>
					<div class="flex items-center gap-1.5 text-xs font-medium text-emerald-600 dark:text-emerald-400">
						<span class="h-2.5 w-2.5 rounded-full bg-emerald-500"></span>
						<span>Messages</span>
					</div>
				</div>

				<div class="relative mt-4">
					<svg
						viewBox="0 0 {chartWidth} {chartHeight}"
						class="w-full h-auto select-none"
						onmouseleave={() => (hoveredMsgIdx = null)}
					>
						<defs>
							<linearGradient id="msgTrendGrad" x1="0" y1="0" x2="0" y2="1">
								<stop offset="0%" stop-color="#10b981" stop-opacity="0.3" />
								<stop offset="100%" stop-color="#10b981" stop-opacity="0.0" />
							</linearGradient>
						</defs>

						<!-- Y Grid Lines -->
						{#each [0, 0.25, 0.5, 0.75, 1] as tickRatio}
							{@const yVal = pad.top + innerH * (1 - tickRatio)}
							{@const val = Math.round(msgMax * tickRatio)}
							<line x1={pad.left} y1={yVal} x2={chartWidth - pad.right} y2={yVal} stroke="currentColor" stroke-dasharray="3 3" class="text-gray-100 dark:text-gray-800" stroke-width="1" />
							<text x={pad.left - 8} y={yVal + 3} text-anchor="end" class="text-[10px] fill-gray-400 dark:fill-gray-500 font-mono">{val}</text>
						{/each}

						<!-- X Labels -->
						{#if data.message_trends}
							{#each data.message_trends as item, idx}
								{#if idx % Math.max(1, Math.floor(data.message_trends.length / 6)) === 0 || idx === data.message_trends.length - 1}
									<text x={getMsgX(idx, data.message_trends.length)} y={chartHeight - 8} text-anchor="middle" class="text-[10px] fill-gray-400 dark:fill-gray-500 font-mono">{item.label}</text>
								{/if}
							{/each}
						{/if}

						<!-- Area & Line -->
						{#if data?.message_trends && data.message_trends.length > 0}
							{@const trends = data.message_trends}
							{@const points = trends.map((item, idx) => ({ x: getMsgX(idx, trends.length), y: getMsgY(item.count) }))}
							<path d={generateSingleAreaPath(points)} fill="url(#msgTrendGrad)" />
							<path d={generateSinglePath(points)} fill="none" stroke="#10b981" stroke-width="2.5" stroke-linecap="round" />

							<!-- Hover Targets -->
							{#each points as pt, idx}
								<rect
									x={pt.x - (innerW / trends.length) / 2}
									y={pad.top}
									width={innerW / trends.length}
									height={innerH}
									fill="transparent"
									class="cursor-pointer"
									onmouseenter={() => (hoveredMsgIdx = idx)}
								/>
								{#if hoveredMsgIdx === idx}
									<line x1={pt.x} y1={pad.top} x2={pt.x} y2={pad.top + innerH} stroke="currentColor" class="text-gray-400" stroke-width="1.5" stroke-dasharray="2 2" />
									<circle cx={pt.x} cy={pt.y} r="4" fill="#10b981" stroke="#fff" stroke-width="2" />
								{/if}
							{/each}
						{/if}
					</svg>

					{#if hoveredMsgIdx !== null && data.message_trends && data.message_trends[hoveredMsgIdx]}
						{@const itm = data.message_trends[hoveredMsgIdx]}
						<div class="pointer-events-none absolute top-2 right-2 rounded-lg border border-gray-200 bg-white/95 p-2 text-xs shadow-md backdrop-blur-xs dark:border-gray-700 dark:bg-gray-800/95">
							<span class="font-bold text-gray-900 dark:text-white">{itm.date}:</span>
							<span class="ml-1 font-mono font-bold text-emerald-600">{itm.count.toLocaleString()} messages</span>
						</div>
					{/if}
				</div>
			</div>

			<!-- Delivery Rate Trend -->
			<div class="rounded-xl border border-gray-100 bg-white p-6 shadow-xs dark:border-gray-800 dark:bg-gray-900">
				<div class="flex items-center justify-between">
					<div>
						<h2 class="text-sm font-bold text-gray-900 dark:text-white">Delivery Rate Trend</h2>
						<p class="text-xs text-gray-500 dark:text-gray-400">Daily delivery rate %</p>
					</div>
					<div class="flex items-center gap-1.5 text-xs font-medium text-emerald-600 dark:text-emerald-400">
						<span class="h-2.5 w-2.5 rounded-full bg-emerald-500"></span>
						<span>Success Rate %</span>
					</div>
				</div>

				<div class="relative mt-4">
					<svg
						viewBox="0 0 {chartWidth} {chartHeight}"
						class="w-full h-auto select-none"
						onmouseleave={() => (hoveredDlrIdx = null)}
					>
						<defs>
							<linearGradient id="dlrTrendGrad" x1="0" y1="0" x2="0" y2="1">
								<stop offset="0%" stop-color="#10b981" stop-opacity="0.25" />
								<stop offset="100%" stop-color="#10b981" stop-opacity="0.0" />
							</linearGradient>
						</defs>

						<!-- Y Grid Lines (0%, 25%, 50%, 75%, 100%) -->
						{#each [0, 0.25, 0.5, 0.75, 1] as tickRatio}
							{@const yVal = pad.top + innerH * (1 - tickRatio)}
							<line x1={pad.left} y1={yVal} x2={chartWidth - pad.right} y2={yVal} stroke="currentColor" stroke-dasharray="3 3" class="text-gray-100 dark:text-gray-800" stroke-width="1" />
							<text x={pad.left - 8} y={yVal + 3} text-anchor="end" class="text-[10px] fill-gray-400 dark:fill-gray-500 font-mono">{tickRatio * 100}%</text>
						{/each}

						<!-- X Labels -->
						{#if data.delivery_rate_trends}
							{#each data.delivery_rate_trends as item, idx}
								{#if idx % Math.max(1, Math.floor(data.delivery_rate_trends.length / 6)) === 0 || idx === data.delivery_rate_trends.length - 1}
									<text x={getMsgX(idx, data.delivery_rate_trends.length)} y={chartHeight - 8} text-anchor="middle" class="text-[10px] fill-gray-400 dark:fill-gray-500 font-mono">{item.label}</text>
								{/if}
							{/each}
						{/if}

						<!-- Area & Line -->
						{#if data?.delivery_rate_trends && data.delivery_rate_trends.length > 0}
							{@const dlrTrends = data.delivery_rate_trends}
							{@const points = dlrTrends.map((item, idx) => ({ x: getMsgX(idx, dlrTrends.length), y: getDlrRateY(item.rate) }))}
							<path d={generateSingleAreaPath(points)} fill="url(#dlrTrendGrad)" />
							<path d={generateSinglePath(points)} fill="none" stroke="#10b981" stroke-width="2.5" stroke-linecap="round" />

							<!-- Hover Targets -->
							{#each points as pt, idx}
								<rect
									x={pt.x - (innerW / dlrTrends.length) / 2}
									y={pad.top}
									width={innerW / dlrTrends.length}
									height={innerH}
									fill="transparent"
									class="cursor-pointer"
									onmouseenter={() => (hoveredDlrIdx = idx)}
								/>
								{#if hoveredDlrIdx === idx}
									<line x1={pt.x} y1={pad.top} x2={pt.x} y2={pad.top + innerH} stroke="currentColor" class="text-gray-400" stroke-width="1.5" stroke-dasharray="2 2" />
									<circle cx={pt.x} cy={pt.y} r="4" fill="#10b981" stroke="#fff" stroke-width="2" />
								{/if}
							{/each}
						{/if}
					</svg>

					{#if hoveredDlrIdx !== null && data.delivery_rate_trends && data.delivery_rate_trends[hoveredDlrIdx]}
						{@const itm = data.delivery_rate_trends[hoveredDlrIdx]}
						<div class="pointer-events-none absolute top-2 right-2 rounded-lg border border-gray-200 bg-white/95 p-2 text-xs shadow-md backdrop-blur-xs dark:border-gray-700 dark:bg-gray-800/95">
							<span class="font-bold text-gray-900 dark:text-white">{itm.date}:</span>
							<span class="ml-1 font-mono font-bold text-emerald-600">{itm.rate.toFixed(1)}% success</span>
						</div>
					{/if}
				</div>
			</div>
		</div>

		<!-- Charts Row 2: Source Distribution, DLR Summary, Failed Reason Summary -->
		<div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
			<!-- Source Distribution Donut -->
			<div class="rounded-xl border border-gray-100 bg-white p-6 shadow-xs dark:border-gray-800 dark:bg-gray-900 flex flex-col">
				<h2 class="text-sm font-bold text-gray-900 dark:text-white">Source Distribution</h2>
				<p class="text-xs text-gray-500 dark:text-gray-400">Messages by source (web vs API)</p>

				<div class="mt-4 flex flex-1 flex-col items-center justify-center">
					<div class="relative flex items-center justify-center">
						<svg width="160" height="160" viewBox="0 0 160 160">
							{#if sourceArcs.length === 0}
								<circle cx="80" cy="80" r="52.5" fill="none" stroke="#e5e7eb" stroke-width="25" class="dark:stroke-gray-800" />
							{:else}
								{#each sourceArcs as arc}
									<path d={arc.d} fill={arc.color} />
								{/each}
							{/if}
						</svg>
						<div class="pointer-events-none absolute flex flex-col items-center justify-center text-center">
							<span class="text-base font-black text-gray-900 dark:text-white">{sourceTotal.toLocaleString()}</span>
							<span class="text-[9px] text-gray-400 uppercase">Total</span>
						</div>
					</div>

					<div class="mt-4 w-full space-y-1.5 text-xs">
						{#each sourceArcs as arc}
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

			<!-- DLR Summary Donut -->
			<div class="rounded-xl border border-gray-100 bg-white p-6 shadow-xs dark:border-gray-800 dark:bg-gray-900 flex flex-col">
				<h2 class="text-sm font-bold text-gray-900 dark:text-white">DLR Summary</h2>
				<p class="text-xs text-gray-500 dark:text-gray-400">Delivery status breakdown</p>

				<div class="mt-4 flex flex-1 flex-col items-center justify-center">
					<div class="relative flex items-center justify-center">
						<svg width="160" height="160" viewBox="0 0 160 160">
							{#if dlrArcs.length === 0}
								<circle cx="80" cy="80" r="52.5" fill="none" stroke="#e5e7eb" stroke-width="25" class="dark:stroke-gray-800" />
							{:else}
								{#each dlrArcs as arc}
									<path d={arc.d} fill={arc.color} />
								{/each}
							{/if}
						</svg>
						<div class="pointer-events-none absolute flex flex-col items-center justify-center text-center">
							<span class="text-base font-black text-gray-900 dark:text-white">{dlrTotal.toLocaleString()}</span>
							<span class="text-[9px] text-gray-400 uppercase">DLR</span>
						</div>
					</div>

					<div class="mt-4 w-full space-y-1.5 text-xs">
						{#each dlrArcs as arc}
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

			<!-- Failed Reason Summary -->
			<div class="rounded-xl border border-gray-100 bg-white p-6 shadow-xs dark:border-gray-800 dark:bg-gray-900 flex flex-col">
				<h2 class="text-sm font-bold text-gray-900 dark:text-white">Failed Reason Summary</h2>
				<p class="text-xs text-gray-500 dark:text-gray-400">Top failure reasons</p>

				<div class="mt-4 flex-1">
					{#if !data.failed_reasons || data.failed_reasons.length === 0}
						<div class="flex h-full flex-col items-center justify-center py-10 text-center text-xs text-gray-400">
							<svg class="h-8 w-8 text-emerald-500 mb-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
							</svg>
							No failures recorded in this period.
						</div>
					{:else}
						<div class="space-y-3">
							{#each data.failed_reasons as fr, idx}
								<div>
									<div class="flex items-center justify-between text-xs mb-1">
										<span class="font-medium text-gray-800 dark:text-gray-200 truncate max-w-[200px]" title={fr.reason}>{fr.reason}</span>
										<span class="font-mono text-gray-500 dark:text-gray-400">{fr.count}</span>
									</div>
									<div class="h-1.5 w-full rounded-full bg-gray-100 dark:bg-gray-800 overflow-hidden">
										<div
											class="h-full rounded-full"
											style="width: {Math.min(100, (fr.count / (data.stats.failed || 1)) * 100)}%; background-color: {palette[idx % palette.length]}"
										></div>
									</div>
								</div>
							{/each}
						</div>
					{/if}
				</div>
			</div>
		</div>
	{/if}
</div>
