<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	let isPaused = $state(false);
	let viewMode = $state<'stream' | 'table'>('stream');
	let queueFilter = $state<'all' | 'pending' | 'inflight' | 'dispatched'>('all');
	let typeFilter = $state<'all' | 'quick' | 'bulk' | 'api' | 'web'>('all');
	let searchFilter = $state('');

	interface DashboardResponse {
		kpis: {
			total_queue: number;
			pending_queue: number;
			processing: number;
			recent_sent: number;
			recent_failed: number;
			dlr_success: number;
			current_tps: number;
			db_latency_ms: number;
			quick_queue: number;
			bulk_queue: number;
		};
		gateways: Array<{ name: string; driver: string; status: string; tps: number; success_rate: number }>;
		live_messages: Array<{
			id: number;
			customer_name: string;
			source: string;
			recipient: string;
			body: string;
			cost: number;
			status: string;
			updated_at: string;
		}>;
	}

	let kpis = $state({
		total_queue: 0,
		pending_queue: 0,
		processing: 0,
		recent_sent: 0,
		recent_failed: 0,
		dlr_success: 100.0,
		current_tps: 0.0,
		db_latency_ms: 0,
		quick_queue: 0,
		bulk_queue: 0
	});

	let gateways = $state<DashboardResponse['gateways']>([]);
	let liveMessages = $state<DashboardResponse['live_messages']>([]);
	let dashboardError = $state<string | null>(null);

	async function loadDashboard() {
		try {
			const response = await supermanApi.get<{ data: DashboardResponse }>('/api/v1/superman/dashboard');
			kpis = response.data.kpis;
			gateways = response.data.gateways;
			liveMessages = response.data.live_messages;
			dashboardError = null;
		} catch (error: unknown) {
			dashboardError = error instanceof Error ? error.message : 'Unable to refresh dashboard data.';
		}
	}

	async function triggerAction(action: 'pause' | 'resume' | 'retry-failed' | 'requeue-stuck' | 'purge') {
		try {
			const response = await supermanApi.post<{ message: string }>(`/api/v1/superman/queue-actions/${action}`);
			await loadDashboard();
			window.alert(response.message);
		} catch (error: unknown) {
			window.alert(error instanceof Error ? error.message : 'Unable to perform this queue action.');
		}
	}

	function formatMessageTime(timestamp: string) {
		return new Date(timestamp).toLocaleString('en-US', {
			hour: '2-digit',
			minute: '2-digit',
			second: '2-digit',
			hour12: true
		});
	}

	let currentTime = $state('');

	function updateTime() {
		const now = new Date();
		currentTime = now.toLocaleString('en-US', {
			day: '2-digit',
			month: 'short',
			year: 'numeric',
			hour: '2-digit',
			minute: '2-digit',
			second: '2-digit',
			hour12: true
		});
	}

	// Filtered messages derived from filter state
	let filteredMessages = $derived(liveMessages.filter((msg) => {
		if (queueFilter === 'pending' && msg.status !== 'pending') return false;
		if (queueFilter === 'inflight' && !['sending', 'processing', 'in-flight'].includes(msg.status)) return false;
		if (queueFilter === 'dispatched' && !['delivered', 'sent', 'dispatched'].includes(msg.status)) return false;
		if (typeFilter !== 'all' && msg.source?.toLowerCase() !== typeFilter) return false;
		if (searchFilter) {
			const q = searchFilter.toLowerCase();
			return (
				msg.customer_name?.toLowerCase().includes(q) ||
				msg.recipient?.toLowerCase().includes(q) ||
				msg.body?.toLowerCase().includes(q)
			);
		}
		return true;
	}));

	function statusColor(status: string) {
		if (['delivered', 'sent', 'dispatched'].includes(status)) return 'text-emerald-600 dark:text-emerald-400';
		if (status === 'pending') return 'text-amber-600 dark:text-amber-400';
		if (['sending', 'processing', 'in-flight'].includes(status)) return 'text-blue-600 dark:text-blue-400';
		if (['failed', 'error'].includes(status)) return 'text-rose-600 dark:text-rose-400';
		return 'text-gray-600 dark:text-gray-400';
	}

	function statusDot(status: string) {
		if (['delivered', 'sent', 'dispatched'].includes(status)) return 'bg-emerald-500';
		if (status === 'pending') return 'bg-amber-500';
		if (['sending', 'processing', 'in-flight'].includes(status)) return 'bg-blue-500';
		if (['failed', 'error'].includes(status)) return 'bg-rose-500';
		return 'bg-gray-400';
	}

	function statusBorderLeft(status: string) {
		if (['delivered', 'sent', 'dispatched'].includes(status)) return 'border-l-emerald-400';
		if (status === 'pending') return 'border-l-amber-400';
		if (['sending', 'processing', 'in-flight'].includes(status)) return 'border-l-blue-400';
		if (['failed', 'error'].includes(status)) return 'border-l-rose-400';
		return 'border-l-gray-300';
	}

	onMount(() => {
		updateTime();
		void loadDashboard();
		const interval = setInterval(updateTime, 1000);
		const refresh = setInterval(() => {
			if (!isPaused) {
				void loadDashboard();
			}
		}, 5000);

		return () => {
			clearInterval(interval);
			clearInterval(refresh);
		};
	});

	const exportColumns: ExportColumn<DashboardResponse['live_messages'][number]>[] = [
		{ header: 'Message ID', getValue: (m) => `#${m.id}` },
		{ header: 'Customer', getValue: (m) => m.customer_name },
		{ header: 'Source', getValue: (m) => m.source.toUpperCase() },
		{ header: 'Recipient', getValue: (m) => m.recipient },
		{ header: 'Message Body', getValue: (m) => m.body },
		{ header: 'Status', getValue: (m) => m.status.toUpperCase() },
		{ header: 'Cost (BDT)', getValue: (m) => `BDT ${Number(m.cost).toFixed(4)}` },
		{ header: 'Timestamp', getValue: (m) => (m.updated_at ? new Date(m.updated_at).toLocaleString() : '—') }
	];
</script>

<svelte:head>
	<title>Live Dashboard — SMS Whiteinfo Administrator</title>
</svelte:head>

<div class="w-full text-gray-900 dark:text-white antialiased font-sans space-y-4">
	<!-- Page Heading -->
	<div class="flex items-center justify-between pb-1">
		<h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">Live Dashboard</h1>
		<div class="flex items-center gap-2">
			<ExportDropdown
				title="Live Queue Messages"
				columns={exportColumns}
				getData={() => liveMessages}
				disabled={liveMessages.length === 0}
			/>
		</div>
	</div>

	<!-- Main Mission Control Container Card -->
	<div class="rounded-xl bg-white shadow-xs ring-1 ring-gray-950/5 dark:bg-gray-900 dark:ring-white/10 p-4 lg:p-5 space-y-4">
		<!-- 1. Top Bar: System Time & Live Indicator -->
		<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2.5 pb-3 border-b border-gray-100 dark:border-gray-800">
			<div class="flex items-center gap-3">
				<div class="flex items-center gap-1.5">
					<span class="text-xs text-gray-500 dark:text-gray-400 font-medium">System Time:</span>
					<span class="text-xs font-mono font-bold text-gray-900 dark:text-white">{currentTime}</span>
				</div>
				<span class="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-[10px] font-extrabold bg-emerald-100 text-emerald-800 border border-emerald-200 dark:bg-emerald-950/50 dark:text-emerald-400 dark:border-emerald-800">
					<span class="relative flex h-2 w-2">
						<span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
						<span class="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"></span>
					</span>
					LIVE DASHBOARD
				</span>
			</div>

			<div class="flex items-center gap-3">
				<button
					type="button"
					onclick={() => (isPaused = !isPaused)}
					class="inline-flex items-center gap-1.5 px-3 py-1 rounded-lg bg-gray-50 hover:bg-gray-100 border border-gray-200 dark:bg-gray-800 dark:hover:bg-gray-700 dark:border-gray-700 text-xs font-medium text-gray-700 dark:text-gray-300 transition-colors cursor-pointer"
				>
					<svg class="w-3.5 h-3.5 text-blue-600 dark:text-blue-400 {isPaused ? '' : 'animate-spin'}" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0 3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182m0-4.991v4.99" />
					</svg>
					<span>Auto Refresh (5s)</span>
					<span class="text-[10px] text-gray-400">({isPaused ? 'Paused' : 'Active'})</span>
				</button>

				<div class="flex items-center gap-1.5 text-gray-500 dark:text-gray-400 text-xs font-medium">
					<span class="w-2 h-2 rounded-full bg-emerald-500"></span>
					<span>All Engines Stable</span>
				</div>
			</div>
		</div>

		<!-- 2. Top 8 KPI Metric Cards -->
		<div class="grid grid-cols-2 sm:grid-cols-4 lg:grid-cols-8 gap-2.5">
			<!-- In-Queue Now -->
			<div class="rounded-xl border border-gray-200 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-800/40 p-3 flex flex-col justify-between min-h-[88px] hover:border-blue-300 hover:shadow-xs transition-all">
				<div class="flex items-center justify-between">
					<span class="text-[9px] font-bold text-gray-500 dark:text-gray-400 uppercase tracking-wider">IN-QUEUE NOW</span>
					<div class="w-5 h-5 rounded-full bg-blue-500/10 text-blue-600 flex items-center justify-center">
						<svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M3.75 12h16.5m-16.5 3.75h16.5M3.75 19.5h16.5M5.625 4.5h12.75a1.875 1.875 0 0 1 0 3.75H5.625a1.875 1.875 0 0 1 0-3.75Z" /></svg>
					</div>
				</div>
				<div class="my-0.5">
					<span class="text-base lg:text-lg font-extrabold text-gray-900 dark:text-white font-mono tracking-tight">{kpis.total_queue}</span>
				</div>
				<span class="text-[9px] text-blue-600 dark:text-blue-400 font-medium">Live Pipeline</span>
			</div>

			<!-- Pending Queue -->
			<div class="rounded-xl border border-amber-200 dark:border-amber-900/40 bg-amber-50/30 dark:bg-amber-950/20 p-3 flex flex-col justify-between min-h-[88px] hover:border-amber-400 hover:shadow-xs transition-all">
				<div class="flex items-center justify-between">
					<span class="text-[9px] font-bold text-amber-700 dark:text-amber-400 uppercase tracking-wider">PENDING QUEUE</span>
					<div class="w-5 h-5 rounded-full bg-amber-500/15 text-amber-600 flex items-center justify-center">
						<svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>
					</div>
				</div>
				<div class="my-0.5">
					<span class="text-base lg:text-lg font-extrabold text-amber-900 dark:text-amber-300 font-mono tracking-tight">{kpis.pending_queue}</span>
				</div>
				<span class="text-[9px] text-amber-700 dark:text-amber-400 font-medium">On verge of send</span>
			</div>

			<!-- In-Flight -->
			<div class="rounded-xl border border-emerald-200 dark:border-emerald-900/40 bg-emerald-50/30 dark:bg-emerald-950/20 p-3 flex flex-col justify-between min-h-[88px] hover:border-emerald-400 hover:shadow-xs transition-all">
				<div class="flex items-center justify-between">
					<span class="text-[9px] font-bold text-emerald-700 dark:text-emerald-400 uppercase tracking-wider">IN-FLIGHT</span>
					<div class="w-5 h-5 rounded-full bg-emerald-500/15 text-emerald-600 flex items-center justify-center">
						<svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M9.594 3.94c.09-.542.56-.94 1.11-.94h2.593c.55 0 1.02.398 1.11.94l.213 1.281c.063.374.313.686.645.87.074.04.147.083.22.127.324.196.72.257 1.075.124l1.217-.456a1.125 1.125 0 0 1 1.37.49l1.296 2.247a1.125 1.125 0 0 1-.26 1.431l-1.003.827c-.293.24-.438.613-.431.992a6.759 6.759 0 0 1 0 .255c-.007.378.138.75.43.99l1.005.828c.424.35.534.954.26 1.43l-1.298 2.247a1.125 1.125 0 0 1-1.369.491l-1.217-.456c-.355-.133-.75-.072-1.076.124a6.57 6.57 0 0 1-.22.128c-.331.183-.581.495-.644.869l-.213 1.28c-.09.543-.56.941-1.11.941h-2.594c-.55 0-1.02-.398-1.11-.94l-.213-1.281c-.062-.374-.312-.686-.644-.87a6.52 6.52 0 0 1-.22-.127c-.325-.196-.72-.257-1.076-.124l-1.217.456a1.125 1.125 0 0 1-1.369-.49l-1.297-2.247a1.125 1.125 0 0 1 .26-1.431l1.004-.827c.292-.24.437-.613.43-.992a6.932 6.932 0 0 1 0-.255c.007-.378-.138-.75-.43-.99l-1.004-.828a1.125 1.125 0 0 1-.26-1.43l1.297-2.247a1.125 1.125 0 0 1 1.37-.491l1.216.456c.356.133.751.072 1.076-.124.072-.044.146-.087.22-.128.332-.183.582-.495.644-.869l.214-1.281Z" /><path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z" /></svg>
					</div>
				</div>
				<div class="my-0.5">
					<span class="text-base lg:text-lg font-extrabold text-emerald-900 dark:text-emerald-300 font-mono tracking-tight">{kpis.processing}</span>
				</div>
				<span class="text-[9px] text-emerald-700 dark:text-emerald-400 font-medium">Transmitting</span>
			</div>

			<!-- Sent (15m) -->
			<div class="rounded-xl border border-purple-200 dark:border-purple-900/40 bg-purple-50/30 dark:bg-purple-950/20 p-3 flex flex-col justify-between min-h-[88px] hover:shadow-xs transition-shadow">
				<div class="flex items-center justify-between">
					<span class="text-[9px] font-bold text-purple-700 dark:text-purple-400 uppercase tracking-wider">SENT (15M)</span>
					<div class="w-5 h-5 rounded-full bg-purple-500/15 text-purple-600 flex items-center justify-center">
						<svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 12 3.269 3.125A59.769 59.769 0 0 1 21.485 12 59.768 59.768 0 0 1 3.27 20.875L5.999 12Zm0 0h7.5" /></svg>
					</div>
				</div>
				<div class="my-0.5">
					<span class="text-base lg:text-lg font-extrabold text-purple-900 dark:text-purple-300 font-mono tracking-tight">{kpis.recent_sent}</span>
				</div>
				<span class="text-[9px] text-purple-700 dark:text-purple-400 font-medium">Recent Dispatched</span>
			</div>

			<!-- Failed (15m) -->
			<div class="rounded-xl border border-rose-200 dark:border-rose-900/40 bg-rose-50/30 dark:bg-rose-950/20 p-3 flex flex-col justify-between min-h-[88px] hover:border-rose-400 hover:shadow-xs transition-all">
				<div class="flex items-center justify-between">
					<span class="text-[9px] font-bold text-rose-700 dark:text-rose-400 uppercase tracking-wider">FAILED (15M)</span>
					<div class="w-5 h-5 rounded-full bg-rose-500/15 text-rose-600 flex items-center justify-center">
						<svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m9.75 9.75 4.5 4.5m0-4.5-4.5 4.5M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>
					</div>
				</div>
				<div class="my-0.5">
					<span class="text-base lg:text-lg font-extrabold text-rose-900 dark:text-rose-300 font-mono tracking-tight">{kpis.recent_failed}</span>
				</div>
				<span class="text-[9px] text-rose-700 dark:text-rose-400 font-medium">Click to Retry</span>
			</div>

			<!-- Success Rate -->
			<div class="rounded-xl border border-cyan-200 dark:border-cyan-900/40 bg-cyan-50/30 dark:bg-cyan-950/20 p-3 flex flex-col justify-between min-h-[88px] hover:shadow-xs transition-shadow">
				<div class="flex items-center justify-between">
					<span class="text-[9px] font-bold text-cyan-700 dark:text-cyan-400 uppercase tracking-wider">SUCCESS RATE</span>
					<div class="w-5 h-5 rounded-full bg-cyan-500/15 text-cyan-600 flex items-center justify-center">
						<svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>
					</div>
				</div>
				<div class="my-0.5">
					<span class="text-base lg:text-lg font-extrabold text-cyan-900 dark:text-cyan-300 font-mono tracking-tight">{kpis.dlr_success}%</span>
				</div>
				<span class="text-[9px] text-cyan-700 dark:text-cyan-400 font-medium">Live Health</span>
			</div>

			<!-- Current TPS -->
			<div class="rounded-xl border border-gray-200 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-800/40 p-3 flex flex-col justify-between min-h-[88px] hover:shadow-xs transition-shadow">
				<div class="flex items-center justify-between">
					<span class="text-[9px] font-bold text-amber-700 dark:text-amber-400 uppercase tracking-wider">CURRENT TPS</span>
					<div class="w-5 h-5 rounded-full bg-amber-500/15 text-amber-600 flex items-center justify-center">
						<svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m3.75 13.5 10.5-11.25L12 10.5h8.25L9.75 21.75 12 13.5H3.75Z" /></svg>
					</div>
				</div>
				<div class="my-0.5">
					<span class="text-base lg:text-lg font-extrabold text-gray-900 dark:text-white font-mono tracking-tight">{kpis.current_tps} <span class="text-[9px] text-gray-400 font-normal">SMS/s</span></span>
				</div>
				<span class="text-[9px] text-amber-700 dark:text-amber-400 font-medium">Real-time Speed</span>
			</div>

			<!-- DB Latency -->
			<div class="rounded-xl border border-gray-200 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-800/40 p-3 flex flex-col justify-between min-h-[88px] hover:shadow-xs transition-shadow">
				<div class="flex items-center justify-between">
					<span class="text-[9px] font-bold text-blue-700 dark:text-blue-400 uppercase tracking-wider">DB LATENCY</span>
					<div class="w-5 h-5 rounded-full bg-blue-500/15 text-blue-600 flex items-center justify-center">
						<svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M8.25 3v1.5M4.5 8.25H3m18 0h-1.5M4.5 12H3m18 0h-1.5m-15 3.75H3m18 0h-1.5M8.25 19.5V21M12 3v1.5m0 15V21m3.75-18v1.5m0 15V21m-9-1.5h10.5a2.25 2.25 0 0 0 2.25-2.25V5.25a2.25 2.25 0 0 0-2.25-2.25H6.75A2.25 2.25 0 0 0 4.5 5.25v10.5A2.25 2.25 0 0 0 6.75 18Z" /></svg>
					</div>
				</div>
				<div class="my-0.5">
					<span class="text-base lg:text-lg font-extrabold text-gray-900 dark:text-white font-mono tracking-tight">{kpis.db_latency_ms.toFixed(2)} <span class="text-[9px] text-gray-400 font-normal">ms</span></span>
				</div>
				<span class="text-[9px] text-blue-700 dark:text-blue-400 font-medium">Engine Response</span>
			</div>
		</div>

		<!-- 3. Middle Row 1: Charts & Gateways Table -->
		<div class="grid grid-cols-1 lg:grid-cols-12 gap-3">
			<!-- Real-Time Queue Overview Chart (5 cols) -->
			<div class="lg:col-span-5 rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-3.5 flex flex-col justify-between shadow-xs">
				<div class="flex items-center justify-between mb-2">
					<div>
						<h3 class="text-xs font-bold text-gray-900 dark:text-white uppercase tracking-wider">REAL-TIME QUEUE OVERVIEW</h3>
						<div class="flex items-center gap-2.5 mt-1 text-[9px]">
							<span class="flex items-center gap-1 text-gray-600 dark:text-gray-300"><span class="w-2 h-1 rounded-full bg-amber-500"></span> Pending</span>
							<span class="flex items-center gap-1 text-gray-600 dark:text-gray-300"><span class="w-2 h-1 rounded-full bg-blue-500"></span> Sending</span>
							<span class="flex items-center gap-1 text-gray-600 dark:text-gray-300"><span class="w-2 h-1 rounded-full bg-emerald-500"></span> Sent</span>
							<span class="flex items-center gap-1 text-gray-600 dark:text-gray-300"><span class="w-2 h-1 rounded-full bg-rose-500"></span> Failed</span>
						</div>
					</div>
					<span class="text-[9px] px-2 py-0.5 rounded bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-300 font-mono">Last 15 Minutes ↓</span>
				</div>

				<div class="w-full h-36 relative mt-1">
					<svg class="w-full h-full" viewBox="0 0 500 160" preserveAspectRatio="none">
						<line x1="30" y1="20" x2="500" y2="20" stroke="#e2e8f0" stroke-dasharray="3,3" class="dark:stroke-gray-800" />
						<line x1="30" y1="60" x2="500" y2="60" stroke="#e2e8f0" stroke-dasharray="3,3" class="dark:stroke-gray-800" />
						<line x1="30" y1="100" x2="500" y2="100" stroke="#e2e8f0" stroke-dasharray="3,3" class="dark:stroke-gray-800" />
						<line x1="30" y1="140" x2="500" y2="140" stroke="#cbd5e1" class="dark:stroke-gray-700" />
						<text x="5" y="24" fill="#94a3b8" font-size="9" font-family="monospace">50</text>
						<text x="5" y="64" fill="#94a3b8" font-size="9" font-family="monospace">25</text>
						<text x="5" y="104" fill="#94a3b8" font-size="9" font-family="monospace">10</text>
						<text x="15" y="144" fill="#94a3b8" font-size="9" font-family="monospace">0</text>
						<polyline points="40,140 80,140 120,135 160,130 200,140 240,140 280,138 320,140 360,140 400,140 440,140 480,140" fill="none" stroke="#10b981" stroke-width="2" />
						<polyline points="40,140 80,140 120,140 160,140 200,140 240,140 280,140 320,140 360,140 400,140 440,140 480,140" fill="none" stroke="#f59e0b" stroke-width="2" />
					</svg>
				</div>
			</div>

			<!-- TPS Spline Chart (4 cols) -->
			<div class="lg:col-span-4 rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-3.5 flex flex-col justify-between shadow-xs">
				<div class="flex items-center justify-between mb-2">
					<h3 class="text-xs font-bold text-gray-900 dark:text-white uppercase tracking-wider">TPS (SMS PER SECOND)</h3>
					<span class="text-[9px] px-2 py-0.5 rounded bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-300 font-mono">Last 15 Minutes ↓</span>
				</div>

				<div class="w-full h-36 relative mt-1">
					<svg class="w-full h-full" viewBox="0 0 500 160" preserveAspectRatio="none">
						<line x1="30" y1="20" x2="500" y2="20" stroke="#e2e8f0" stroke-dasharray="3,3" class="dark:stroke-gray-800" />
						<line x1="30" y1="60" x2="500" y2="60" stroke="#e2e8f0" stroke-dasharray="3,3" class="dark:stroke-gray-800" />
						<line x1="30" y1="100" x2="500" y2="100" stroke="#e2e8f0" stroke-dasharray="3,3" class="dark:stroke-gray-800" />
						<line x1="30" y1="140" x2="500" y2="140" stroke="#cbd5e1" class="dark:stroke-gray-700" />
						<text x="5" y="24" fill="#94a3b8" font-size="9" font-family="monospace">1.0</text>
						<text x="5" y="64" fill="#94a3b8" font-size="9" font-family="monospace">0.5</text>
						<text x="15" y="144" fill="#94a3b8" font-size="9" font-family="monospace">0.0</text>
						<polyline points="40,140 80,140 120,135 160,130 200,140 240,140 280,138 320,140 360,140 400,140 440,140 480,140" fill="none" stroke="#0284c7" stroke-width="2" />
					</svg>
				</div>
			</div>

			<!-- Gateway Status Table (3 cols) -->
			<div class="lg:col-span-3 rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-3.5 flex flex-col justify-between shadow-xs">
				<div class="flex items-center justify-between mb-1.5">
					<h3 class="text-xs font-bold text-gray-900 dark:text-white uppercase tracking-wider">GATEWAY STATUS</h3>
					<span class="text-[9px] text-emerald-600 dark:text-emerald-400 font-semibold cursor-pointer">View All</span>
				</div>

				<div class="overflow-x-auto">
					<table class="w-full text-left text-xs">
						<thead>
							<tr class="border-b border-gray-100 dark:border-gray-800 text-gray-400 text-[9px] uppercase">
								<th class="pb-1.5 font-semibold">Gateway</th>
								<th class="pb-1.5 font-semibold">Status</th>
								<th class="pb-1.5 font-semibold text-center">TPS</th>
								<th class="pb-1.5 font-semibold text-right">Success %</th>
							</tr>
						</thead>
						<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
							{#each gateways as gw}
								<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/40 transition-colors">
									<td class="py-1.5 font-medium text-gray-900 dark:text-white text-[11px] truncate max-w-[85px]">{gw.name}</td>
									<td class="py-1.5">
										<span class="inline-flex items-center gap-1 text-[10px] font-semibold text-emerald-600 dark:text-emerald-400">
											<span class="w-1.5 h-1.5 rounded-full bg-emerald-500"></span>
											{gw.status}
										</span>
									</td>
									<td class="py-1.5 text-center font-mono text-gray-700 dark:text-gray-300 text-[11px]">{gw.tps}</td>
									<td class="py-1.5 text-right font-mono font-semibold text-[11px] text-emerald-600 dark:text-emerald-400">{gw.success_rate}%</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			</div>
		</div>

		<!-- 4. Action Command Bar -->
		<div class="flex flex-wrap items-center justify-between gap-2.5 pt-3 border-t border-gray-100 dark:border-gray-800">
			<div class="flex flex-wrap items-center gap-2">
				<button
					type="button"
					onclick={() => triggerAction('pause')}
					class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-emerald-50 hover:bg-emerald-100 border border-emerald-300 text-emerald-700 dark:bg-emerald-950/40 dark:border-emerald-800 dark:text-emerald-300 text-xs font-bold transition-colors cursor-pointer"
				>
					Pause All Queues
				</button>
				<button
					type="button"
					onclick={() => triggerAction('resume')}
					class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-blue-600 hover:bg-blue-700 text-white text-xs font-bold transition-colors cursor-pointer shadow-xs"
				>
					Resume All Queues
				</button>
				<button
					type="button"
					onclick={() => triggerAction('retry-failed')}
					class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-amber-500 hover:bg-amber-600 text-white text-xs font-bold transition-colors cursor-pointer shadow-xs"
				>
					Retry Failed
				</button>
				<button
					type="button"
					onclick={() => triggerAction('requeue-stuck')}
					class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-purple-600 hover:bg-purple-700 text-white text-xs font-bold transition-colors cursor-pointer shadow-xs"
				>
					Requeue Stuck
				</button>
				<button
					type="button"
					onclick={() => triggerAction('purge')}
					class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-rose-600 hover:bg-rose-700 text-white text-xs font-bold transition-colors cursor-pointer shadow-xs"
				>
					Clear / Purge Queue
				</button>
			</div>
			<div>
				<a
					href="/superman/system-logs/sms-logs"
					class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-white hover:bg-gray-50 border border-gray-300 dark:bg-gray-800 dark:hover:bg-gray-700 dark:border-gray-600 text-gray-700 dark:text-gray-200 text-xs font-bold transition-colors cursor-pointer shadow-xs"
				>
					Export Logs
				</a>
			</div>
		</div>
	</div>

	<!-- ================================================================
	     LIVE QUEUE & ACTION STREAM SECTION
	     ================================================================ -->
	<div class="rounded-xl bg-white shadow-xs ring-1 ring-gray-950/5 dark:bg-gray-900 dark:ring-white/10 overflow-hidden">
		<!-- Section Header -->
		<div class="flex flex-col sm:flex-row sm:items-center gap-3 px-4 lg:px-5 py-3.5 border-b border-gray-100 dark:border-gray-800">
			<div class="flex items-center gap-3 flex-1 min-w-0">
				<div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-emerald-100 dark:bg-emerald-950/50">
					<svg class="h-5 w-5 text-emerald-600 dark:text-emerald-400" fill="none" viewBox="0 0 24 24" stroke-width="1.75" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" d="M9.348 14.652a3.75 3.75 0 0 1 0-5.304m5.304 0a3.75 3.75 0 0 1 0 5.304m-7.425 2.121a6.75 6.75 0 0 1 0-9.546m9.546 0a6.75 6.75 0 0 1 0 9.546M12 12h.008v.008H12V12Z" />
					</svg>
				</div>
				<div class="min-w-0">
					<div class="flex items-center gap-2">
						<h2 class="text-sm font-bold text-gray-900 dark:text-white truncate">Live Queue & Action Stream</h2>
						<span class="inline-flex items-center gap-1.5 rounded-full bg-emerald-100 px-2 py-0.5 text-[10px] font-bold text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300 whitespace-nowrap">
							<span class="relative flex h-1.5 w-1.5">
								<span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
								<span class="relative inline-flex rounded-full h-1.5 w-1.5 bg-emerald-500"></span>
							</span>
							LIVE STREAMING (5s)
						</span>
					</div>
					<p class="text-[10px] text-gray-500 dark:text-gray-400 mt-0.5">Real-time monitor of live SMS currently in queue, on the verge of being processed, and in-flight gateway dispatches</p>
				</div>
			</div>

			<!-- Controls -->
			<div class="flex items-center gap-2 shrink-0 flex-wrap">
				<div class="relative">
					<svg class="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z" />
					</svg>
					<input
						type="text"
						bind:value={searchFilter}
						placeholder="Filter live: customer, phone, text..."
						class="pl-8 pr-3 py-1.5 text-xs rounded-lg border border-gray-200 bg-gray-50 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-200 outline-none focus:ring-1 focus:ring-emerald-500 focus:border-emerald-500 w-48 placeholder:text-gray-400"
					/>
				</div>
				<button
					type="button"
					onclick={() => (viewMode = 'stream')}
					class="inline-flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg text-xs font-semibold transition-colors {viewMode === 'stream' ? 'bg-gray-900 text-white dark:bg-white dark:text-gray-900' : 'text-gray-600 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-800'}"
				>
					<svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M3.75 12h16.5m-16.5 3.75h16.5M3.75 19.5h16.5M5.625 4.5h12.75a1.875 1.875 0 0 1 0 3.75H5.625a1.875 1.875 0 0 1 0-3.75Z" /></svg>
					Stream
				</button>
				<button
					type="button"
					onclick={() => (viewMode = 'table')}
					class="inline-flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg text-xs font-semibold transition-colors {viewMode === 'table' ? 'bg-gray-900 text-white dark:bg-white dark:text-gray-900' : 'text-gray-600 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-800'}"
				>
					<svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M3.375 19.5h17.25m-17.25 0a1.125 1.125 0 0 1-1.125-1.125M3.375 19.5h7.5c.621 0 1.125-.504 1.125-1.125m-9.75 0V5.625m0 12.75v-1.5c0-.621.504-1.125 1.125-1.125m18.375 2.625V5.625m0 12.75c0 .621-.504 1.125-1.125 1.125m1.125-1.125v-1.5c0-.621-.504-1.125-1.125-1.125m0 3.75h-7.5A1.125 1.125 0 0 1 12 18.375m9.75-12.75c0-.621-.504-1.125-1.125-1.125H3.375c-.621 0-1.125.504-1.125 1.125m19.5 0v1.5c0 .621-.504 1.125-1.125 1.125M2.25 5.625v1.5c0 .621.504 1.125 1.125 1.125m0 0h17.25m-17.25 0h7.5c.621 0 1.125.504 1.125 1.125M3.375 8.25c-.621 0-1.125.504-1.125 1.125v1.5c0 .621.504 1.125 1.125 1.125m17.25-3.75h-7.5c-.621 0-1.125.504-1.125 1.125m8.625-1.125c.621 0 1.125.504 1.125 1.125v1.5c0 .621-.504 1.125-1.125 1.125m-17.25 0h7.5" /></svg>
					Table
				</button>
				<button
					type="button"
					onclick={() => triggerAction('purge')}
					class="inline-flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg bg-rose-50 hover:bg-rose-100 border border-rose-200 text-rose-700 dark:bg-rose-950/30 dark:border-rose-900 dark:text-rose-400 text-xs font-semibold transition-colors"
				>
					<svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m14.74 9-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 0 0-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 0 0-7.5 0" /></svg>
					Purge
				</button>
				<button
					type="button"
					onclick={() => (isPaused = !isPaused)}
					class="inline-flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg border border-gray-200 bg-white hover:bg-gray-50 dark:border-gray-700 dark:bg-gray-800 dark:hover:bg-gray-700 text-gray-700 dark:text-gray-200 text-xs font-semibold transition-colors"
				>
					{#if isPaused}
						<svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M5.25 5.653c0-.856.917-1.398 1.667-.986l11.54 6.347a1.125 1.125 0 0 1 0 1.972l-11.54 6.347a1.125 1.125 0 0 1-1.667-.986V5.653Z" /></svg>
						Resume
					{:else}
						<svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M15.75 5.25v13.5m-7.5-13.5v13.5" /></svg>
						Pause
					{/if}
				</button>
			</div>
		</div>

		<!-- Filter Tabs Row -->
		<div class="flex flex-wrap items-center justify-between gap-2 px-4 lg:px-5 py-2.5 bg-gray-50/60 dark:bg-gray-800/30 border-b border-gray-100 dark:border-gray-800">
			<!-- Queue status filters -->
			<div class="flex items-center gap-1 flex-wrap">
				<span class="text-[10px] font-semibold text-gray-500 dark:text-gray-400 mr-1">QUEUE:</span>
				<button type="button" onclick={() => (queueFilter = 'all')}
					class="rounded-md px-2.5 py-1 text-[11px] font-semibold transition-colors {queueFilter === 'all' ? 'bg-gray-900 text-white dark:bg-white dark:text-gray-900' : 'text-gray-600 hover:bg-gray-200 dark:text-gray-400 dark:hover:bg-gray-700'}">
					All Live
				</button>
				<button type="button" onclick={() => (queueFilter = 'pending')}
					class="inline-flex items-center gap-1 rounded-md px-2.5 py-1 text-[11px] font-semibold transition-colors {queueFilter === 'pending' ? 'bg-amber-600 text-white' : 'text-amber-700 hover:bg-amber-50 dark:text-amber-400 dark:hover:bg-amber-950/30'}">
					<svg class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>
					In Queue / Pending
				</button>
				<button type="button" onclick={() => (queueFilter = 'inflight')}
					class="inline-flex items-center gap-1 rounded-md px-2.5 py-1 text-[11px] font-semibold transition-colors {queueFilter === 'inflight' ? 'bg-blue-600 text-white' : 'text-blue-700 hover:bg-blue-50 dark:text-blue-400 dark:hover:bg-blue-950/30'}">
					<svg class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m3.75 13.5 10.5-11.25L12 10.5h8.25L9.75 21.75 12 13.5H3.75Z" /></svg>
					In Flight / Sending
				</button>
				<button type="button" onclick={() => (queueFilter = 'dispatched')}
					class="inline-flex items-center gap-1 rounded-md px-2.5 py-1 text-[11px] font-semibold transition-colors {queueFilter === 'dispatched' ? 'bg-emerald-600 text-white' : 'text-emerald-700 hover:bg-emerald-50 dark:text-emerald-400 dark:hover:bg-emerald-950/30'}">
					<svg class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>
					Dispatched (&lt; 10s)
				</button>
			</div>

			<!-- Type filters -->
			<div class="flex items-center gap-1">
				<span class="text-[10px] font-semibold text-gray-500 dark:text-gray-400 mr-1">TYPE:</span>
				{#each [['all','All'],['quick','Quick'],['bulk','Bulk'],['api','API'],['web','Web']] as [val, label]}
					<button type="button" onclick={() => (typeFilter = val as typeof typeFilter)}
						class="rounded-md px-2 py-1 text-[11px] font-semibold transition-colors {typeFilter === val ? 'bg-gray-800 text-white dark:bg-gray-200 dark:text-gray-900' : 'text-gray-600 hover:bg-gray-200 dark:text-gray-400 dark:hover:bg-gray-700'}">
						{label}
					</button>
				{/each}
			</div>
		</div>

		<!-- Content: Empty state / Stream / Table -->
		{#if filteredMessages.length === 0}
			<!-- Empty / Idle state -->
			<div class="flex flex-col items-center justify-center py-16 px-4">
				<div class="relative mb-5">
					<div class="flex h-16 w-16 items-center justify-center rounded-full bg-emerald-100 dark:bg-emerald-950/50">
						<svg class="h-8 w-8 text-emerald-500 dark:text-emerald-400" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" d="M9.348 14.652a3.75 3.75 0 0 1 0-5.304m5.304 0a3.75 3.75 0 0 1 0 5.304m-7.425 2.121a6.75 6.75 0 0 1 0-9.546m9.546 0a6.75 6.75 0 0 1 0 9.546M12 12h.008v.008H12V12Z" />
						</svg>
					</div>
					<span class="absolute -right-1 -top-1 flex h-4 w-4 items-center justify-center rounded-full bg-emerald-500">
						<span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
					</span>
				</div>
				<h3 class="text-sm font-bold text-gray-900 dark:text-white">Queue is Idle & Ready</h3>
				<p class="mt-1.5 text-center text-xs text-gray-500 dark:text-gray-400 max-w-xs leading-relaxed">
					No messages are currently waiting in queue or in-flight right now. System is actively listening for live incoming SMS traffic (polling every 5 seconds).
				</p>
				<div class="mt-4 inline-flex items-center gap-1.5 rounded-full border border-emerald-200 bg-emerald-50 px-3 py-1.5 dark:border-emerald-800 dark:bg-emerald-950/30">
					<span class="relative flex h-2 w-2">
						<span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
						<span class="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"></span>
					</span>
					<span class="text-[11px] font-semibold text-emerald-700 dark:text-emerald-400">Live dispatch pipeline active &amp; listening</span>
				</div>
			</div>
		{:else if viewMode === 'stream'}
			<!-- Stream View: animated card rows with status colored left borders -->
			<div class="divide-y divide-gray-100 dark:divide-gray-800">
				{#each filteredMessages as msg (msg.id)}
					<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 py-3 px-4 lg:px-5 hover:bg-gray-50/50 dark:hover:bg-gray-800/30 border-l-2 {statusBorderLeft(msg.status)} transition-colors">
						<div class="space-y-1 min-w-0 flex-1">
							<div class="flex flex-wrap items-center gap-2">
								<span class="font-semibold text-xs text-gray-900 dark:text-white">{msg.customer_name}</span>
								<span class="rounded border border-sky-200 bg-sky-50 px-1.5 py-0.5 text-[10px] font-medium text-sky-700 dark:border-sky-900 dark:bg-sky-950/50 dark:text-sky-300">{msg.source}</span>
								<span class="rounded bg-gray-100 dark:bg-gray-800 px-1.5 py-0.5 text-[10px] font-mono text-gray-600 dark:text-gray-300">{msg.recipient}</span>
								<span class="text-[10px] font-semibold text-emerald-600 dark:text-emerald-400">৳{msg.cost}</span>
							</div>
							<p class="text-xs text-gray-600 dark:text-gray-300 truncate max-w-2xl">{msg.body}</p>
						</div>
						<div class="flex items-center gap-2.5 shrink-0">
							<span class="inline-flex items-center gap-1 text-[11px] font-semibold {statusColor(msg.status)}">
								<span class="w-1.5 h-1.5 rounded-full {statusDot(msg.status)}"></span>
								{msg.status}
							</span>
							<span class="text-[11px] text-gray-400 font-mono">{formatMessageTime(msg.updated_at)}</span>
						</div>
					</div>
				{/each}
			</div>
		{:else}
			<!-- Table View: proper data table -->
			<div class="overflow-x-auto">
				<table class="w-full text-xs">
					<thead class="bg-gray-50 dark:bg-gray-800/50 border-b border-gray-200 dark:border-gray-700">
						<tr>
							<th class="px-4 py-2.5 text-left text-[10px] font-bold text-gray-500 dark:text-gray-400 uppercase tracking-wider">ID</th>
							<th class="px-4 py-2.5 text-left text-[10px] font-bold text-gray-500 dark:text-gray-400 uppercase tracking-wider">Customer</th>
							<th class="px-4 py-2.5 text-left text-[10px] font-bold text-gray-500 dark:text-gray-400 uppercase tracking-wider">Type</th>
							<th class="px-4 py-2.5 text-left text-[10px] font-bold text-gray-500 dark:text-gray-400 uppercase tracking-wider">Recipient</th>
							<th class="px-4 py-2.5 text-left text-[10px] font-bold text-gray-500 dark:text-gray-400 uppercase tracking-wider">Message</th>
							<th class="px-4 py-2.5 text-center text-[10px] font-bold text-gray-500 dark:text-gray-400 uppercase tracking-wider">Cost</th>
							<th class="px-4 py-2.5 text-center text-[10px] font-bold text-gray-500 dark:text-gray-400 uppercase tracking-wider">Status</th>
							<th class="px-4 py-2.5 text-right text-[10px] font-bold text-gray-500 dark:text-gray-400 uppercase tracking-wider">Time</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
						{#each filteredMessages as msg (msg.id)}
							<tr class="hover:bg-gray-50/60 dark:hover:bg-gray-800/30 transition-colors">
								<td class="px-4 py-2.5 font-mono text-[10px] text-gray-500 dark:text-gray-400">#{msg.id}</td>
								<td class="px-4 py-2.5 font-semibold text-gray-900 dark:text-white whitespace-nowrap">{msg.customer_name}</td>
								<td class="px-4 py-2.5">
									<span class="rounded border border-sky-200 bg-sky-50 px-1.5 py-0.5 text-[10px] font-medium text-sky-700 dark:border-sky-900 dark:bg-sky-950/50 dark:text-sky-300">{msg.source}</span>
								</td>
								<td class="px-4 py-2.5 font-mono text-gray-600 dark:text-gray-300">{msg.recipient}</td>
								<td class="px-4 py-2.5 text-gray-600 dark:text-gray-300 max-w-xs truncate">{msg.body}</td>
								<td class="px-4 py-2.5 text-center font-semibold text-emerald-600 dark:text-emerald-400">৳{msg.cost}</td>
								<td class="px-4 py-2.5 text-center">
									<span class="inline-flex items-center gap-1 text-[10px] font-semibold {statusColor(msg.status)}">
										<span class="w-1.5 h-1.5 rounded-full {statusDot(msg.status)}"></span>
										{msg.status}
									</span>
								</td>
								<td class="px-4 py-2.5 text-right font-mono text-gray-400 whitespace-nowrap">{formatMessageTime(msg.updated_at)}</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</div>
</div>
