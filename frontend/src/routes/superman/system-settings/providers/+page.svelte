<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';

	interface ProviderStats {
		total_gateways: number;
		active_gateways: number;
		healthy_gateways: number;
		otp_gateways: number;
	}

	let stats = $state<ProviderStats>({
		total_gateways: 0,
		active_gateways: 0,
		healthy_gateways: 0,
		otp_gateways: 0
	});
	let loading = $state(true);

	onMount(async () => {
		try {
			const res = await supermanApi.get<{ data: ProviderStats }>('/api/v1/superman/system-settings/providers');
			if (res.data) stats = res.data;
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	});
</script>

<div class="space-y-6">
	<!-- Page Header -->
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-xl font-bold text-gray-900 dark:text-white">Provider Settings Overview</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
				Carrier gateway configuration, live HTTP health monitors, and telecom interconnect telemetry.
			</p>
		</div>
		<div class="flex items-center gap-3">
			<a
				href="/superman/gateways"
				class="inline-flex items-center gap-2 rounded-lg bg-emerald-600 px-3 py-1.5 text-xs font-semibold text-white shadow-xs hover:bg-emerald-700 transition"
			>
				Manage Gateways
			</a>
		</div>
	</div>

	<!-- KPI Stat Cards -->
	<div class="grid grid-cols-1 md:grid-cols-4 gap-4">
		<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-5 shadow-xs">
			<span class="text-xs font-medium text-gray-500 uppercase tracking-wider">Total Gateways</span>
			<p class="text-3xl font-bold text-gray-900 dark:text-white mt-1">
				{loading ? '...' : stats.total_gateways}
			</p>
			<span class="text-[11px] text-gray-400 mt-1 block">Registered upstream routes</span>
		</div>
		<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-5 shadow-xs">
			<span class="text-xs font-medium text-emerald-600 dark:text-emerald-400 uppercase tracking-wider">Active Gateways</span>
			<p class="text-3xl font-bold text-emerald-600 dark:text-emerald-400 mt-1">
				{loading ? '...' : stats.active_gateways}
			</p>
			<span class="text-[11px] text-gray-400 mt-1 block">Enabled for live traffic</span>
		</div>
		<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-5 shadow-xs">
			<span class="text-xs font-medium text-blue-600 dark:text-blue-400 uppercase tracking-wider">Healthy & Online</span>
			<p class="text-3xl font-bold text-blue-600 dark:text-blue-400 mt-1">
				{loading ? '...' : stats.healthy_gateways}
			</p>
			<span class="text-[11px] text-gray-400 mt-1 block">Passing automated HTTP pings</span>
		</div>
		<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-5 shadow-xs">
			<span class="text-xs font-medium text-purple-600 dark:text-purple-400 uppercase tracking-wider">OTP Channels</span>
			<p class="text-3xl font-bold text-purple-600 dark:text-purple-400 mt-1">
				{loading ? '...' : stats.otp_gateways}
			</p>
			<span class="text-[11px] text-gray-400 mt-1 block">High-priority 2FA enabled</span>
		</div>
	</div>

	<!-- Navigation Hub Cards -->
	<div class="grid grid-cols-1 md:grid-cols-3 gap-6">
		<a
			href="/superman/gateways"
			class="group block p-6 bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl hover:border-emerald-500 dark:hover:border-emerald-500 transition shadow-xs"
		>
			<div class="flex items-center gap-3">
				<div class="p-2.5 rounded-lg bg-emerald-50 dark:bg-emerald-950/50 text-emerald-600 dark:text-emerald-400">
					<svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.75">
						<path stroke-linecap="round" stroke-linejoin="round" d="M21.75 17.25v-.228a4.5 4.5 0 0 0-.12-1.03l-2.268-9.64a3.375 3.375 0 0 0-3.285-2.602H7.923a3.375 3.375 0 0 0-3.285 2.602l-2.268 9.64a4.5 4.5 0 0 0-.12 1.03v.228m19.5 0a3 3 0 0 1-3 3H5.25a3 3 0 0 1-3-3m19.5 0a3 3 0 0 0-3-3H5.25a3 3 0 0 0-3 3m16.5 0h.008v.008h-.008v-.008Zm-3 0h.008v.008h-.008v-.008Z" />
					</svg>
				</div>
				<div>
					<h3 class="font-semibold text-gray-900 dark:text-white group-hover:text-emerald-600 transition">Manage Gateways</h3>
					<span class="text-xs text-gray-500">Infozillion, ReveSMS, QSMS, Anbernet</span>
				</div>
			</div>
			<p class="text-xs text-gray-600 dark:text-gray-400 mt-4 leading-relaxed">
				Configure endpoint URLs, API secrets, priority ranks, dipping tariffs, and maximum requests per second.
			</p>
		</a>

		<a
			href="/superman/providers/health"
			class="group block p-6 bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl hover:border-emerald-500 dark:hover:border-emerald-500 transition shadow-xs"
		>
			<div class="flex items-center gap-3">
				<div class="p-2.5 rounded-lg bg-emerald-50 dark:bg-emerald-950/50 text-emerald-600 dark:text-emerald-400">
					<svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.75">
						<path stroke-linecap="round" stroke-linejoin="round" d="M21 8.25c0-2.485-2.099-4.5-4.688-4.5-1.935 0-3.597 1.126-4.312 2.733-.715-1.607-2.377-2.733-4.313-2.733C5.1 3.75 3 5.765 3 8.25c0 7.22 9 12 9 12s9-4.78 9-12Z" />
					</svg>
				</div>
				<div>
					<h3 class="font-semibold text-gray-900 dark:text-white group-hover:text-emerald-600 transition">Provider Health & Latency</h3>
					<span class="text-xs text-gray-500">Live HTTP Round-Trip Ping</span>
				</div>
			</div>
			<p class="text-xs text-gray-600 dark:text-gray-400 mt-4 leading-relaxed">
				Real-time latency gauges, health states (Healthy, Degraded, Down), automated reachability triggers, and error ratios.
			</p>
		</a>

		<a
			href="/superman/routing/rules"
			class="group block p-6 bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl hover:border-emerald-500 dark:hover:border-emerald-500 transition shadow-xs"
		>
			<div class="flex items-center gap-3">
				<div class="p-2.5 rounded-lg bg-amber-50 dark:bg-amber-950/50 text-amber-600 dark:text-amber-400">
					<svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.75">
						<path stroke-linecap="round" stroke-linejoin="round" d="M7.5 21 3 16.5m0 0L7.5 12M3 16.5h13.5m0-13.5L21 7.5m0 0L16.5 12M21 7.5H7.5" />
					</svg>
				</div>
				<div>
					<h3 class="font-semibold text-gray-900 dark:text-white group-hover:text-emerald-600 transition">Route Plans & Failover</h3>
					<span class="text-xs text-gray-500">Intelligent Traffic Steering</span>
				</div>
			</div>
			<p class="text-xs text-gray-600 dark:text-gray-400 mt-4 leading-relaxed">
				Build priority 1 ➔ 2 ➔ 3 failover chains and traffic-type rules (OTP, Transactional, Promotional, Default).
			</p>
		</a>
	</div>
</div>

