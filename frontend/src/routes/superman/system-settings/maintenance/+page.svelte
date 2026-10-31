<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';

	interface MaintenanceState {
		is_down: boolean;
		secret: string;
		notice_message: string;
		exempt_ips: string[];
	}

	let maintState = $state<MaintenanceState>({
		is_down: false,
		secret: 'whiteinfo-admin-bypass',
		notice_message: 'System is currently undergoing scheduled platform upgrades.',
		exempt_ips: ['127.0.0.1']
	});
	let loading = $state(true);
	let updating = $state(false);
	let message = $state<string | null>(null);

	onMount(async () => {
		try {
			const res = await supermanApi.get<{ data: MaintenanceState }>('/api/v1/superman/system-settings/maintenance');
			if (res.data) maintState = res.data;
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	});

	async function toggleMaintenance() {
		updating = true;
		message = null;
		try {
			const newState = !maintState.is_down;
			await supermanApi.post('/api/v1/superman/system-settings/maintenance', {
				is_down: newState,
				secret: maintState.secret,
				notice: maintState.notice_message
			});
			maintState.is_down = newState;
			message = newState
				? 'Maintenance mode has been ACTIVATED. End users will see maintenance notice.'
				: 'Maintenance mode DEACTIVATED. System is fully live and operational.';
		} catch (e: any) {
			alert(e?.message || 'Failed to toggle maintenance mode.');
		} finally {
			updating = false;
		}
	}
</script>

<div class="max-w-3xl space-y-6">
	<!-- Page Header -->
	<div>
		<h1 class="text-xl font-bold text-gray-900 dark:text-white">System Maintenance Mode</h1>
		<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
			Safely pause customer web portals, API dispatches, and incoming requests during database schema migrations or telecom upgrades.
		</p>
	</div>

	{#if message}
		<div class="rounded-lg p-4 text-xs font-medium {maintState.is_down ? 'bg-amber-50 dark:bg-amber-950/40 text-amber-800 dark:text-amber-300 border border-amber-200 dark:border-amber-800' : 'bg-emerald-50 dark:bg-emerald-950/40 text-emerald-800 dark:text-emerald-300 border border-emerald-200 dark:border-emerald-800'}">
			{message}
		</div>
	{/if}

	<!-- Maintenance Control Card -->
	<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-6 shadow-xs space-y-6">
		<div class="flex items-center justify-between pb-6 border-b border-gray-100 dark:border-gray-800">
			<div>
				<div class="flex items-center gap-2">
					<span class="h-3 w-3 rounded-full {maintState.is_down ? 'bg-amber-500 animate-pulse' : 'bg-emerald-500'}"></span>
					<h3 class="text-sm font-bold text-gray-900 dark:text-white">
						Current Status: {maintState.is_down ? 'MAINTENANCE MODE ACTIVE' : 'LIVE & OPERATIONAL'}
					</h3>
				</div>
				<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
					{maintState.is_down ? 'All non-admin API requests receive HTTP 503 Service Unavailable.' : 'All customer, reseller, and API gateways are fully reachable.'}
				</p>
			</div>

			<button
				type="button"
				onclick={toggleMaintenance}
				disabled={updating}
				class="rounded-lg px-4 py-2 text-xs font-semibold shadow-xs transition disabled:opacity-50 {maintState.is_down ? 'bg-emerald-600 text-white hover:bg-emerald-700' : 'bg-red-600 text-white hover:bg-red-700'}"
			>
				{updating ? 'Processing...' : maintState.is_down ? 'Bring System Online' : 'Enable Maintenance Mode'}
			</button>
		</div>

		<div class="space-y-4">
			<div>
				<label for="maintenance_notice" class="block text-xs font-semibold text-gray-700 dark:text-gray-300 mb-1">
					Customer Notice Message
				</label>
				<textarea
					id="maintenance_notice"
					bind:value={maintState.notice_message}
					rows="3"
					class="w-full rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 p-3 text-xs text-gray-900 dark:text-white focus:outline-none focus:ring-1 focus:ring-emerald-500"
				></textarea>
			</div>

			<div>
				<label for="maintenance_secret" class="block text-xs font-semibold text-gray-700 dark:text-gray-300 mb-1">
					Superman Admin Bypass Secret Token
				</label>
				<input
					id="maintenance_secret"
					type="text"
					bind:value={maintState.secret}
					class="w-full rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 px-3 py-2 text-xs font-mono text-gray-900 dark:text-white focus:outline-none focus:ring-1 focus:ring-emerald-500"
				/>
				<span class="text-[11px] text-gray-400 mt-1 block">Staff can append ?secret={maintState.secret} to bypass the maintenance curtain.</span>
			</div>
		</div>
	</div>
</div>
