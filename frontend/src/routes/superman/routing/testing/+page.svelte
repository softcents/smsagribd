<script lang="ts">
	import { supermanApi } from '$lib/superman-api.js';

	interface SimulationResult {
		phone: string;
		detected_mno: string;
		traffic_type: string;
		matched_plan: string;
		matched_code: string;
		primary_gateway: string;
		fallback_gw_1: string;
		fallback_gw_2: string;
		protocol: string;
		tps_limit: string;
		status: string;
	}

	let phone = $state('01712345678');
	let trafficType = $state('otp');
	let senderId = $state('WhiteInfo');
	let simulating = $state(false);
	let result = $state<SimulationResult | null>(null);

	async function runSimulation() {
		simulating = true;
		result = null;
		try {
			const res = await supermanApi.post<{ data: SimulationResult }>('/api/v1/superman/routing/testing', {
				phone,
				traffic_type: trafficType,
				sender_id: senderId
			});
			result = res.data ?? null;
		} catch (e: any) {
			alert(e?.message || 'Route simulation failed.');
		} finally {
			simulating = false;
		}
	}
</script>

<div class="max-w-4xl space-y-6">
	<!-- Page Header -->
	<div>
		<h1 class="text-xl font-bold text-gray-900 dark:text-white">Intelligent Outbound Route Simulator</h1>
		<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
			Simulate live routing resolution without sending SMS. Test carrier MNO detection, fallback priority chains, and dipping overheads.
		</p>
	</div>

	<!-- Simulator Input Form -->
	<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-6 shadow-xs space-y-6">
		<h2 class="text-sm font-bold text-gray-900 dark:text-white">Simulation Parameters</h2>
		<div class="grid grid-cols-1 md:grid-cols-3 gap-6">
			<div>
				<label for="sim_phone" class="block text-xs font-semibold text-gray-700 dark:text-gray-300 mb-1">
					Destination MSISDN (Phone #)
				</label>
				<input
					id="sim_phone"
					type="text"
					bind:value={phone}
					placeholder="01712345678"
					class="w-full rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 px-3 py-2 text-xs font-mono text-gray-900 dark:text-white focus:outline-none focus:ring-1 focus:ring-emerald-500"
				/>
				<span class="text-[11px] text-gray-400 mt-1 block">Supports 017/013, 019/014, 018/016, 015.</span>
			</div>

			<div>
				<label for="sim_traffic" class="block text-xs font-semibold text-gray-700 dark:text-gray-300 mb-1">
					Traffic Sensitivity Class
				</label>
				<select
					id="sim_traffic"
					bind:value={trafficType}
					class="w-full rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 px-3 py-2 text-xs text-gray-900 dark:text-white focus:outline-none focus:ring-1 focus:ring-emerald-500"
				>
					<option value="otp">OTP / 2FA High-Priority</option>
					<option value="transactional">Transactional Notification</option>
					<option value="promotional">Promotional / Marketing Bulk</option>
				</select>
			</div>

			<div>
				<label for="sim_sender" class="block text-xs font-semibold text-gray-700 dark:text-gray-300 mb-1">
					Sender ID / CLI Header
				</label>
				<input
					id="sim_sender"
					type="text"
					bind:value={senderId}
					placeholder="WhiteInfo"
					class="w-full rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 px-3 py-2 text-xs text-gray-900 dark:text-white focus:outline-none focus:ring-1 focus:ring-emerald-500"
				/>
			</div>
		</div>

		<div class="flex justify-end pt-2">
			<button
				type="button"
				onclick={runSimulation}
				disabled={simulating}
				class="rounded-lg bg-emerald-600 px-4 py-2 text-xs font-semibold text-white shadow-xs hover:bg-emerald-700 transition disabled:opacity-50"
			>
				{simulating ? 'Analyzing Route Engine...' : 'Run Route Simulation'}
			</button>
		</div>
	</div>

	<!-- Simulation Outcome Card -->
	{#if result}
		<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-6 shadow-xs space-y-6">
			<div class="flex items-center justify-between pb-4 border-b border-gray-100 dark:border-gray-800">
				<div>
					<span class="inline-flex items-center gap-1.5 text-xs text-emerald-600 font-bold uppercase tracking-wider">
						<span class="h-2 w-2 rounded-full bg-emerald-600"></span>
						{result.status}
					</span>
					<h3 class="text-base font-bold text-gray-900 dark:text-white mt-1">
						Target MNO: <span class="text-emerald-600">{result.detected_mno}</span>
					</h3>
				</div>
				<span class="px-2.5 py-1 rounded bg-gray-100 dark:bg-gray-800 font-mono text-xs font-bold text-gray-700 dark:text-gray-300">
					{result.traffic_type}
				</span>
			</div>

			<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
				<div class="p-4 rounded-xl bg-gray-50 dark:bg-gray-800/40 space-y-2 text-xs">
					<div class="text-gray-400 font-medium">Matched Route Plan</div>
					<div class="font-bold text-gray-900 dark:text-white text-sm">{result.matched_plan}</div>
					<div class="font-mono text-gray-500 text-[11px]">Rule Identifier: {result.matched_code}</div>
				</div>

				<div class="p-4 rounded-xl bg-emerald-50 dark:bg-emerald-950/40 border border-emerald-200 dark:border-emerald-800 space-y-2 text-xs">
					<div class="text-emerald-700 dark:text-emerald-300 font-medium">Selected Primary Carrier Pipe</div>
					<div class="font-bold text-emerald-900 dark:text-emerald-200 text-sm">{result.primary_gateway}</div>
					<div class="font-mono text-emerald-600 dark:text-emerald-400 text-[11px]">{result.protocol} • Rate Limit: {result.tps_limit}</div>
				</div>
			</div>

			<div>
				<h4 class="text-xs font-bold text-gray-700 dark:text-gray-300 mb-2">Automated Failover Backup Chain</h4>
				<div class="flex items-center gap-2 text-xs font-mono">
					<span class="px-3 py-1.5 rounded-lg bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300 font-bold">
						1. {result.primary_gateway}
					</span>
					<span class="text-gray-400">➔</span>
					<span class="px-3 py-1.5 rounded-lg bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300">
						2. {result.fallback_gw_1}
					</span>
					<span class="text-gray-400">➔</span>
					<span class="px-3 py-1.5 rounded-lg bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300">
						3. {result.fallback_gw_2}
					</span>
				</div>
			</div>
		</div>
	{/if}
</div>

