<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';

	interface RoutingStats {
		total_plans: number;
		active_plans: number;
		total_rules: number;
		active_rules: number;
	}

	let stats = $state<RoutingStats>({
		total_plans: 0,
		active_plans: 0,
		total_rules: 0,
		active_rules: 0
	});
	let loading = $state(true);

	onMount(async () => {
		try {
			const res = await supermanApi.get<{ data: RoutingStats }>('/api/v1/superman/system-settings/routing');
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
			<h1 class="text-xl font-bold text-gray-900 dark:text-white">SMS Routing & Failover Overview</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
				Dynamic operator prefix matching, least-cost routing (LCR), and automated upstream carrier failover.
			</p>
		</div>
		<div class="flex items-center gap-3">
			<a
				href="/superman/routing/testing"
				class="inline-flex items-center gap-2 rounded-lg bg-emerald-600 px-3 py-1.5 text-xs font-semibold text-white shadow-xs hover:bg-emerald-700 transition"
			>
				Launch Route Simulator
			</a>
		</div>
	</div>

	<!-- KPI Stat Cards -->
	<div class="grid grid-cols-1 md:grid-cols-4 gap-4">
		<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-5 shadow-xs">
			<span class="text-xs font-medium text-gray-500 uppercase tracking-wider">Total Route Plans</span>
			<p class="text-3xl font-bold text-gray-900 dark:text-white mt-1">{loading ? '...' : stats.total_plans}</p>
			<span class="text-[11px] text-gray-400 mt-1 block">Registered route definitions</span>
		</div>
		<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-5 shadow-xs">
			<span class="text-xs font-medium text-emerald-600 dark:text-emerald-400 uppercase tracking-wider">Active Route Plans</span>
			<p class="text-3xl font-bold text-emerald-600 dark:text-emerald-400 mt-1">{loading ? '...' : stats.active_plans}</p>
			<span class="text-[11px] text-gray-400 mt-1 block">Live in dispatch engine</span>
		</div>
		<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-5 shadow-xs">
			<span class="text-xs font-medium text-blue-600 dark:text-blue-400 uppercase tracking-wider">Total Rules</span>
			<p class="text-3xl font-bold text-blue-600 dark:text-blue-400 mt-1">{loading ? '...' : stats.total_rules}</p>
			<span class="text-[11px] text-gray-400 mt-1 block">MNO prefix & traffic filters</span>
		</div>
		<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-5 shadow-xs">
			<span class="text-xs font-medium text-purple-600 dark:text-purple-400 uppercase tracking-wider">Active Rules</span>
			<p class="text-3xl font-bold text-purple-600 dark:text-purple-400 mt-1">{loading ? '...' : stats.active_rules}</p>
			<span class="text-[11px] text-gray-400 mt-1 block">Currently filtering traffic</span>
		</div>
	</div>

	<!-- Navigation Hub Cards -->
	<div class="grid grid-cols-1 md:grid-cols-3 gap-6">
		<a
			href="/superman/routing/dashboard"
			class="group block p-6 bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl hover:border-emerald-500 dark:hover:border-emerald-500 transition shadow-xs"
		>
			<div class="flex items-center gap-3">
				<div class="p-2.5 rounded-lg bg-emerald-50 dark:bg-emerald-950/50 text-emerald-600 dark:text-emerald-400">
					<svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.75">
						<path stroke-linecap="round" stroke-linejoin="round" d="M3.75 13.5 10.5-11.25L12 10.5h8.25L9.75 21.75 12 13.5H3.75Z" />
					</svg>
				</div>
				<div>
					<h3 class="font-semibold text-gray-900 dark:text-white group-hover:text-emerald-600 transition">Routing Control</h3>
					<span class="text-xs text-gray-500">Live Traffic Distribution</span>
				</div>
			</div>
			<p class="text-xs text-gray-600 dark:text-gray-400 mt-4 leading-relaxed">
				View primary and secondary gateway bindings, default route fallbacks, and real-time status.
			</p>
		</a>

		<a
			href="/superman/routing/rules"
			class="group block p-6 bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl hover:border-emerald-500 dark:hover:border-emerald-500 transition shadow-xs"
		>
			<div class="flex items-center gap-3">
				<div class="p-2.5 rounded-lg bg-blue-50 dark:bg-blue-950/50 text-blue-600 dark:text-blue-400">
					<svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.75">
						<path stroke-linecap="round" stroke-linejoin="round" d="M10.5 6h9.75M10.5 12h9.75M10.5 18h9.75M3.75 6H7.5M3.75 12H7.5M3.75 18H7.5" />
					</svg>
				</div>
				<div>
					<h3 class="font-semibold text-gray-900 dark:text-white group-hover:text-emerald-600 transition">Routing Rules</h3>
					<span class="text-xs text-gray-500">Operator Prefix Steering</span>
				</div>
			</div>
			<p class="text-xs text-gray-600 dark:text-gray-400 mt-4 leading-relaxed">
				Bind GP (017/013), BL (019/014), Robi/Airtel (018/016), and Teletalk (015) to dedicated carrier pipes.
			</p>
		</a>

		<a
			href="/superman/routing/load-balancing"
			class="group block p-6 bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl hover:border-emerald-500 dark:hover:border-emerald-500 transition shadow-xs"
		>
			<div class="flex items-center gap-3">
				<div class="p-2.5 rounded-lg bg-purple-50 dark:bg-purple-950/50 text-purple-600 dark:text-purple-400">
					<svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.75">
						<path stroke-linecap="round" stroke-linejoin="round" d="M12 3v17.25m0 0c-1.472 0-2.882.265-4.185.75M12 20.25c1.472 0 2.882.265 4.185.75M18.75 4.97A48.416 48.416 0 0 0 12 4.5c-2.291 0-4.545.16-6.75.47m13.5 0c1.01.143 2.01.317 3 .52m-16.5-.52c-1.01.143-2.01.317-3 .52m16.5 0a2.25 2.25 0 0 1 1.95 2.18 48.437 48.437 0 0 1-.99 8.24m-17.46-8.24A2.25 2.25 0 0 0 1.5 7.15a48.437 48.437 0 0 0-.99 8.24m17.46-8.24c-1.63 1.83-3.66 3.19-5.97 3.94M4.5 7.15c1.63 1.83 3.66 3.19 5.97 3.94" />
					</svg>
				</div>
				<div>
					<h3 class="font-semibold text-gray-900 dark:text-white group-hover:text-emerald-600 transition">Load Balancing</h3>
					<span class="text-xs text-gray-500">Weighted Round-Robin</span>
				</div>
			</div>
			<p class="text-xs text-gray-600 dark:text-gray-400 mt-4 leading-relaxed">
				Configure traffic ratio weights across multiple carrier interconnects to balance throughput and cost.
			</p>
		</a>
	</div>
</div>

