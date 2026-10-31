<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';

	interface SmsSettings {
		sms_enabled: boolean;
		default_sms_type: string;
		default_encoding: string;
		auto_detect_unicode: boolean;
		force_unicode: boolean;
		allow_long_sms: boolean;
		max_sms_parts: number;
		duplicate_protection_enabled: boolean;
		duplicate_window_seconds: number;
		default_country_code: string;
		log_payloads: boolean;
		log_retention_days: number;
		message_id_prefix: string;
	}

	let settings = $state<SmsSettings>({
		sms_enabled: true,
		default_sms_type: 'plain_text',
		default_encoding: 'GSM-7',
		auto_detect_unicode: true,
		force_unicode: false,
		allow_long_sms: true,
		max_sms_parts: 5,
		duplicate_protection_enabled: true,
		duplicate_window_seconds: 60,
		default_country_code: '880',
		log_payloads: true,
		log_retention_days: 90,
		message_id_prefix: 'WI'
	});

	let loading = $state(true);
	let saving = $state(false);
	let saved = $state(false);
	let error = $state<string | null>(null);

	onMount(async () => {
		try {
			const res = await supermanApi.get<{ data: SmsSettings }>('/api/v1/superman/settings/sms');
			if (res.data) settings = { ...settings, ...res.data };
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Unable to load SMS settings.';
		} finally {
			loading = false;
		}
	});

	async function saveSettings(e: SubmitEvent) {
		e.preventDefault();
		saving = true;
		saved = false;
		error = null;
		try {
			await supermanApi.put('/api/v1/superman/settings/sms', settings);
			saved = true;
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Failed to save SMS settings.';
		} finally {
			saving = false;
		}
	}
</script>

<svelte:head>
	<title>SMS Technical Settings — Superman Administrator</title>
</svelte:head>

<div class="space-y-6">
	<div>
		<h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">SMS Technical Settings</h1>
		<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
			Configure global SMS dispatch, encoding standards, multi-part concat limits, and deduplication windows.
		</p>
	</div>

	<form onsubmit={saveSettings} class="max-w-3xl space-y-5 rounded-xl border border-gray-200 bg-white p-6 shadow-xs dark:border-gray-800 dark:bg-gray-900">
		{#if error}
			<div class="rounded-lg bg-rose-50 p-3 text-xs text-rose-700 dark:bg-rose-950/40 dark:text-rose-300">{error}</div>
		{/if}
		{#if saved}
			<div class="rounded-lg bg-emerald-50 p-3 text-xs text-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-300">SMS technical settings saved successfully.</div>
		{/if}

		<div class="grid gap-4 sm:grid-cols-2">
			<div>
				<label class="block text-xs font-semibold text-gray-700 dark:text-gray-300">Default SMS Type</label>
				<select
					bind:value={settings.default_sms_type}
					class="mt-1 block w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-xs text-gray-900 outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
				>
					<option value="plain_text">Plain Text</option>
					<option value="masking">Masking SMS</option>
					<option value="non_masking">Non-Masking SMS</option>
					<option value="transactional">Transactional</option>
					<option value="promotional">Promotional</option>
				</select>
			</div>

			<div>
				<label class="block text-xs font-semibold text-gray-700 dark:text-gray-300">Character Encoding</label>
				<select
					bind:value={settings.default_encoding}
					class="mt-1 block w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-xs text-gray-900 outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
				>
					<option value="GSM-7">GSM 7-bit (160 chars / SMS)</option>
					<option value="UCS-2">UCS-2 / Unicode (70 chars / SMS)</option>
					<option value="UTF-8">UTF-8 Automatic Detection</option>
				</select>
			</div>

			<div>
				<label class="block text-xs font-semibold text-gray-700 dark:text-gray-300">Default Country Dial Code</label>
				<input
					type="text"
					bind:value={settings.default_country_code}
					class="mt-1 block w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-xs text-gray-900 outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
				/>
			</div>

			<div>
				<label class="block text-xs font-semibold text-gray-700 dark:text-gray-300">Max SMS Parts (Concatenation)</label>
				<input
					type="number"
					bind:value={settings.max_sms_parts}
					min="1"
					max="10"
					class="mt-1 block w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-xs text-gray-900 outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
				/>
			</div>

			<div>
				<label class="block text-xs font-semibold text-gray-700 dark:text-gray-300">Deduplication Window (Seconds)</label>
				<input
					type="number"
					bind:value={settings.duplicate_window_seconds}
					min="0"
					max="3600"
					class="mt-1 block w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-xs text-gray-900 outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
				/>
			</div>

			<div>
				<label class="block text-xs font-semibold text-gray-700 dark:text-gray-300">Message ID Prefix</label>
				<input
					type="text"
					bind:value={settings.message_id_prefix}
					class="mt-1 block w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-xs text-gray-900 outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
				/>
			</div>
		</div>

		<div class="space-y-2 border-t border-gray-100 pt-3 dark:border-gray-800">
			<label class="flex items-center gap-2 cursor-pointer">
				<input type="checkbox" bind:checked={settings.sms_enabled} class="h-4 w-4 rounded border-gray-300 text-emerald-600 focus:ring-emerald-500" />
				<span class="text-xs font-medium text-gray-700 dark:text-gray-300">Global SMS Sending Enabled (Master Switch)</span>
			</label>
			<label class="flex items-center gap-2 cursor-pointer">
				<input type="checkbox" bind:checked={settings.auto_detect_unicode} class="h-4 w-4 rounded border-gray-300 text-emerald-600 focus:ring-emerald-500" />
				<span class="text-xs font-medium text-gray-700 dark:text-gray-300">Auto-Detect Unicode Characters (Bangla / Arabic / Emojis)</span>
			</label>
			<label class="flex items-center gap-2 cursor-pointer">
				<input type="checkbox" bind:checked={settings.duplicate_protection_enabled} class="h-4 w-4 rounded border-gray-300 text-emerald-600 focus:ring-emerald-500" />
				<span class="text-xs font-medium text-gray-700 dark:text-gray-300">Block Duplicate SMS Submissions within Window</span>
			</label>
			<label class="flex items-center gap-2 cursor-pointer">
				<input type="checkbox" bind:checked={settings.log_payloads} class="h-4 w-4 rounded border-gray-300 text-emerald-600 focus:ring-emerald-500" />
				<span class="text-xs font-medium text-gray-700 dark:text-gray-300">Record Full Request and Response Payloads for Audit</span>
			</label>
		</div>

		<div class="border-t border-gray-100 pt-3 dark:border-gray-800">
			<button
				type="submit"
				disabled={saving}
				class="rounded-lg bg-emerald-600 px-4 py-2 text-xs font-semibold text-white shadow-xs transition hover:bg-emerald-700 disabled:opacity-60"
			>
				{saving ? 'Saving...' : 'Save SMS Settings'}
			</button>
		</div>
	</form>
</div>

