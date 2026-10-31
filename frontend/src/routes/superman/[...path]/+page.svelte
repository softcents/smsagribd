<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import { blankSupermanPages, settingsSupermanPages, tableSupermanPages } from '$lib/superman-pages.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';

	let path = $derived($page.params.path);
	let tablePage = $derived(tableSupermanPages[path]);
	let settingsPage = $derived(settingsSupermanPages[path]);
	let title = $derived(tablePage?.title ?? settingsPage?.title ?? blankSupermanPages[path] ?? 'Superman Administrator');
	let rows = $state<Record<string, unknown>[]>([]);
	let settings = $state<Record<string, unknown>>({});
	let search = $state('');
	let error = $state<string | null>(null);
	let saving = $state(false);
	let saved = $state(false);
	let filteredRows = $derived(
		rows.filter((row) => JSON.stringify(row).toLowerCase().includes(search.toLowerCase()))
	);
	let exportColumns = $derived(
		tablePage?.columns.map((c) => ({
			header: c.label,
			key: c.key,
			getValue: (row: Record<string, unknown>) => format(row[c.key], c.format)
		})) ?? []
	);

	function format(value: unknown, formatType?: 'currency' | 'date' | 'boolean') {
		if (value === null || value === undefined || value === '') {
			return '—';
		}
		if (formatType === 'currency') {
			return `৳${Number(value).toFixed(2)}`;
		}
		if (formatType === 'date') {
			return new Date(String(value)).toLocaleString();
		}
		if (formatType === 'boolean') {
			return value ? 'Enabled' : 'Disabled';
		}
		return String(value);
	}

	function setSetting(key: string, value: string | number | boolean) {
		settings = { ...settings, [key]: value };
	}

	function inputValue(event: Event) {
		return (event.currentTarget as HTMLInputElement).value;
	}

	function checkedValue(event: Event) {
		return (event.currentTarget as HTMLInputElement).checked;
	}

	async function saveSettings() {
		if (!settingsPage) {
			return;
		}
		saving = true;
		saved = false;
		error = null;
		try {
			const response = await supermanApi.put<{ data: Record<string, unknown> }>(`/api/v1/superman/settings/${settingsPage.group}`, settings);
			settings = response.data;
			saved = true;
		} catch (saveError: unknown) {
			error = saveError instanceof Error ? saveError.message : 'Unable to save settings.';
		} finally {
			saving = false;
		}
	}

	onMount(async () => {
		if (settingsPage) {
			try {
				const response = await supermanApi.get<{ data: Record<string, unknown> }>(`/api/v1/superman/settings/${settingsPage.group}`);
				settings = response.data;
			} catch (loadError: unknown) {
				error = loadError instanceof Error ? loadError.message : 'Unable to load settings.';
			}
			return;
		}

		if (!tablePage) {
			return;
		}

		try {
			const response = await supermanApi.get<{ data: Record<string, unknown>[] }>(tablePage.endpoint);
			rows = response.data;
		} catch (loadError: unknown) {
			error = loadError instanceof Error ? loadError.message : 'Unable to load this Superman resource.';
		}
	});
</script>

<svelte:head>
	<title>{title} — SMS Whiteinfo Administrator</title>
</svelte:head>

<div class="space-y-4">
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">{title}</h1>
			{#if tablePage}
				<p class="text-xs text-gray-500 dark:text-gray-400">{tablePage.description}</p>
			{:else if settingsPage}
				<p class="text-xs text-gray-500 dark:text-gray-400">{settingsPage.description}</p>
			{/if}
		</div>
		{#if tablePage}
			<div class="flex items-center gap-2">
				<ExportDropdown
					title={title}
					columns={exportColumns}
					getData={() => filteredRows}
					disabled={filteredRows.length === 0}
				/>
			</div>
		{/if}
	</div>

	{#if tablePage}
		<div class="rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900 space-y-4">
			<div class="relative w-full max-w-xs">
				<input
					type="search"
					bind:value={search}
					placeholder={`Search ${title.toLowerCase()}...`}
					class="w-full rounded-lg border border-gray-200 bg-gray-50/70 py-1.5 pl-3 pr-3 text-xs outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
				/>
			</div>

			{#if error}
				<div class="rounded-lg border border-rose-200 bg-rose-50 p-3 text-xs text-rose-700 dark:border-rose-900 dark:bg-rose-950/30 dark:text-rose-300">{error}</div>
			{:else}
				<div class="overflow-x-auto">
					<table class="w-full text-left text-xs">
						<thead>
							<tr class="border-b border-gray-100 text-[10px] font-semibold uppercase text-gray-400 dark:border-gray-800">
								{#each tablePage.columns as column}
									<th class="py-2.5">{column.label}</th>
								{/each}
							</tr>
						</thead>
						<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
							{#each filteredRows as row}
								<tr class="transition-colors hover:bg-gray-50/50 dark:hover:bg-gray-800/30">
									{#each tablePage.columns as column}
										<td class="max-w-xs truncate py-3 text-gray-700 dark:text-gray-300">{format(row[column.key], column.format)}</td>
									{/each}
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{/if}
		</div>
	{/if}

	{#if settingsPage}
		<form onsubmit={(event) => { event.preventDefault(); void saveSettings(); }} class="max-w-3xl space-y-4 rounded-xl border border-gray-200 bg-white p-6 shadow-xs dark:border-gray-800 dark:bg-gray-900">
			{#if error}
				<div class="rounded-lg border border-rose-200 bg-rose-50 p-3 text-xs text-rose-700 dark:border-rose-900 dark:bg-rose-950/30 dark:text-rose-300">{error}</div>
			{/if}
			{#if saved}
				<div class="rounded-lg border border-emerald-200 bg-emerald-50 p-3 text-xs text-emerald-700 dark:border-emerald-900 dark:bg-emerald-950/30 dark:text-emerald-300">Settings saved successfully.</div>
			{/if}
			<div class="grid gap-4 sm:grid-cols-2">
				{#each settingsPage.fields as field}
					<label class="space-y-1.5 {field.type === 'text' && (field.key === 'trusted_portal_ips' || field.key === 'blocked_message') ? 'sm:col-span-2' : ''}">
						<span class="block text-xs font-medium text-gray-700 dark:text-gray-300">{field.label}</span>
						{#if field.type === 'boolean'}
							<input type="checkbox" checked={Boolean(settings[field.key])} onchange={(event) => setSetting(field.key, checkedValue(event))} class="h-4 w-4 rounded border-gray-300 text-emerald-600 focus:ring-emerald-500" />
						{:else if field.type === 'select'}
							<select value={String(settings[field.key] ?? '')} onchange={(event) => setSetting(field.key, inputValue(event))} class="block w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-xs text-gray-900 outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white">
								{#each field.options ?? [] as option}
									<option value={option.value}>{option.label}</option>
								{/each}
							</select>
						{:else}
							<input type={field.type} value={String(settings[field.key] ?? '')} onchange={(event) => setSetting(field.key, field.type === 'number' ? Number(inputValue(event)) : inputValue(event))} class="block w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-xs text-gray-900 outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white" />
						{/if}
					</label>
				{/each}
			</div>
			<div class="border-t border-gray-100 pt-3 dark:border-gray-800">
				<button type="submit" disabled={saving} class="rounded-lg bg-emerald-600 px-4 py-2 text-xs font-semibold text-white shadow-sm transition hover:bg-emerald-700 disabled:cursor-not-allowed disabled:opacity-60">{saving ? 'Saving…' : 'Save Changes'}</button>
			</div>
		</form>
	{/if}
</div>
