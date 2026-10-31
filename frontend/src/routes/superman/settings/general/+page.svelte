<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';

	interface GeneralSettings {
		appName: string;
		logoUrl: string;
	}

	let appName = $state('');
	let logoUrl = $state('');
	let saving = $state(false);
	let saved = $state(false);
	let error = $state<string | null>(null);

	onMount(async () => {
		try {
			const response = await supermanApi.get<{ data: GeneralSettings }>('/api/v1/superman/settings/general');
			appName = response.data.appName;
			logoUrl = response.data.logoUrl;
		} catch (loadError: unknown) {
			error = loadError instanceof Error ? loadError.message : 'Unable to load company settings.';
		}
	});

	async function handleSave(event: SubmitEvent) {
		event.preventDefault();
		saving = true;
		saved = false;
		error = null;
		try {
			const response = await supermanApi.put<{ data: GeneralSettings }>('/api/v1/superman/settings/general', {
				appName,
				logoUrl
			});
			appName = response.data.appName;
			logoUrl = response.data.logoUrl;
			saved = true;
		} catch (saveError: unknown) {
			error = saveError instanceof Error ? saveError.message : 'Unable to save company settings.';
		} finally {
			saving = false;
		}
	}
</script>

<svelte:head>
	<title>Company Info — Superman Administrator</title>
</svelte:head>

<div class="space-y-6">
	<div>
		<h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">Company & Branding Information</h1>
		<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">Manage the application title, brand name, and public logo URL asset.</p>
	</div>

	<form onsubmit={handleSave} class="max-w-xl space-y-4 rounded-xl border border-gray-200 bg-white p-6 shadow-xs dark:border-gray-800 dark:bg-gray-900">
		{#if error}
			<div class="rounded-lg border border-rose-200 bg-rose-50 p-3 text-xs text-rose-700 dark:border-rose-900 dark:bg-rose-950/30 dark:text-rose-300">
				{error}
			</div>
		{/if}

		{#if saved}
			<div class="rounded-lg border border-emerald-200 bg-emerald-50 p-3 text-xs text-emerald-700 dark:border-emerald-900 dark:bg-emerald-950/30 dark:text-emerald-300">
				Company information saved successfully.
			</div>
		{/if}

		<div>
			<label for="appName" class="block text-xs font-semibold text-gray-700 dark:text-gray-300">Application Name</label>
			<input
				id="appName"
				type="text"
				bind:value={appName}
				required
				class="mt-1 block w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-xs text-gray-900 outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
			/>
		</div>

		<div>
			<label for="logoUrl" class="block text-xs font-semibold text-gray-700 dark:text-gray-300">Brand Logo Asset URL / Path</label>
			<input
				id="logoUrl"
				type="text"
				bind:value={logoUrl}
				required
				class="mt-1 block w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-xs text-gray-900 outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
			/>
		</div>

		<div class="border-t border-gray-100 pt-3 dark:border-gray-800">
			<button
				type="submit"
				disabled={saving}
				class="rounded-lg bg-emerald-600 px-4 py-2 text-xs font-semibold text-white shadow-xs transition hover:bg-emerald-700 disabled:opacity-60"
			>
				{saving ? 'Saving Changes...' : 'Save Company Info'}
			</button>
		</div>
	</form>
</div>

