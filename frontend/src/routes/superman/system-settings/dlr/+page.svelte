<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';

	interface DlrConfig {
		max_retries: number;
		polling_interval_minutes: number;
		expiry_hours: number;
		auto_refund_undelivered: boolean;
		webhook_timeout_seconds: number;
		retry_backoff: string;
	}

	let form = $state<DlrConfig>({
		max_retries: 3,
		polling_interval_minutes: 2,
		expiry_hours: 48,
		auto_refund_undelivered: false,
		webhook_timeout_seconds: 10,
		retry_backoff: 'exponential'
	});
	let loading = $state(true);
	let saving = $state(false);
	let successMsg = $state<string | null>(null);

	onMount(async () => {
		try {
			const res = await supermanApi.get<{ data: DlrConfig }>('/api/v1/superman/system-settings/dlr');
			if (res.data) form = res.data;
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	});

	async function handleSave() {
		saving = true;
		successMsg = null;
		try {
			await supermanApi.post('/api/v1/superman/system-settings/dlr', form);
			successMsg = 'Delivery report (DLR) policies updated successfully.';
		} catch (e: any) {
			alert(e?.message || 'Failed to save DLR settings.');
		} finally {
			saving = false;
		}
	}
</script>

<div class="max-w-4xl space-y-6">
	<!-- Page Header -->
	<div>
		<h1 class="text-xl font-bold text-gray-900 dark:text-white">Delivery Report (DLR) Timing & Webhook Policies</h1>
		<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
			Configure upstream telecommunications status synchronization, customer webhook callback timeouts, and refund rules.
		</p>
	</div>

	{#if successMsg}
		<div class="rounded-lg bg-emerald-50 dark:bg-emerald-950/40 border border-emerald-200 dark:border-emerald-800 p-4 text-xs font-medium text-emerald-800 dark:text-emerald-300">
			{successMsg}
		</div>
	{/if}

	<!-- Form Card -->
	<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-6 shadow-xs space-y-6">
		<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
			<div>
				<label for="dlr_retries" class="block text-xs font-semibold text-gray-700 dark:text-gray-300 mb-1">
					Maximum Webhook Retries
				</label>
				<input
					id="dlr_retries"
					type="number"
					bind:value={form.max_retries}
					min="1"
					max="10"
					class="w-full rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 px-3 py-2 text-xs text-gray-900 dark:text-white focus:outline-none focus:ring-1 focus:ring-emerald-500"
				/>
				<span class="text-[11px] text-gray-400 mt-1 block">Number of delivery webhook retry attempts before marking failed.</span>
			</div>

			<div>
				<label for="dlr_polling" class="block text-xs font-semibold text-gray-700 dark:text-gray-300 mb-1">
					Carrier DLR Polling Frequency (Minutes)
				</label>
				<input
					id="dlr_polling"
					type="number"
					bind:value={form.polling_interval_minutes}
					min="1"
					max="60"
					class="w-full rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 px-3 py-2 text-xs text-gray-900 dark:text-white focus:outline-none focus:ring-1 focus:ring-emerald-500"
				/>
				<span class="text-[11px] text-gray-400 mt-1 block">Scheduled sweep to query pending SMS statuses from aggregators.</span>
			</div>

			<div>
				<label for="dlr_expiry" class="block text-xs font-semibold text-gray-700 dark:text-gray-300 mb-1">
					Message Status Expiry Horizon (Hours)
				</label>
				<input
					id="dlr_expiry"
					type="number"
					bind:value={form.expiry_hours}
					min="1"
					max="168"
					class="w-full rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 px-3 py-2 text-xs text-gray-900 dark:text-white focus:outline-none focus:ring-1 focus:ring-emerald-500"
				/>
				<span class="text-[11px] text-gray-400 mt-1 block">Time after which unacknowledged messages transition to 'undelivered'.</span>
			</div>

			<div>
				<label for="dlr_timeout" class="block text-xs font-semibold text-gray-700 dark:text-gray-300 mb-1">
					Customer Webhook Timeout (Seconds)
				</label>
				<input
					id="dlr_timeout"
					type="number"
					bind:value={form.webhook_timeout_seconds}
					min="1"
					max="60"
					class="w-full rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 px-3 py-2 text-xs text-gray-900 dark:text-white focus:outline-none focus:ring-1 focus:ring-emerald-500"
				/>
				<span class="text-[11px] text-gray-400 mt-1 block">HTTP socket timeout for delivering callbacks to customer servers.</span>
			</div>
		</div>

		<!-- Auto Refund Switch -->
		<div class="pt-4 border-t border-gray-100 dark:border-gray-800 flex items-center justify-between">
			<div>
				<h4 class="text-xs font-bold text-gray-900 dark:text-white">Auto-Refund Undelivered Messages</h4>
				<p class="text-[11px] text-gray-500 dark:text-gray-400 mt-0.5">
					Automatically credit SMS rate back to customer wallet upon carrier 'UNDELIV' or 'EXPIRED' status.
				</p>
			</div>
			<input
				type="checkbox"
				bind:checked={form.auto_refund_undelivered}
				class="h-4 w-4 rounded border-gray-300 text-emerald-600 focus:ring-emerald-500"
			/>
		</div>

		<div class="pt-4 flex justify-end">
			<button
				type="button"
				onclick={handleSave}
				disabled={saving}
				class="rounded-lg bg-emerald-600 px-4 py-2 text-xs font-semibold text-white shadow-xs hover:bg-emerald-700 transition disabled:opacity-50"
			>
				{saving ? 'Saving Changes...' : 'Save DLR Settings'}
			</button>
		</div>
	</div>
</div>

