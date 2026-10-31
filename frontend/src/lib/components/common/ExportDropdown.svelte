<script lang="ts">
	import { exportToCSV, exportToPDF, type ExportColumn } from '$lib/utils/export.js';

	interface Props {
		title: string;
		columns: ExportColumn<any>[];
		getData: () => Promise<any[]> | any[];
		disabled?: boolean;
		filterSummary?: string;
		class?: string;
	}

	let {
		title,
		columns,
		getData,
		disabled = false,
		filterSummary = '',
		class: customClass = ''
	}: Props = $props();

	let open = $state(false);
	let exporting = $state(false);
	let toastMessage = $state<string | null>(null);
	let toastType = $state<'success' | 'error'>('success');
	let toastTimeout: any;

	function showToast(msg: string, type: 'success' | 'error' = 'success') {
		toastMessage = msg;
		toastType = type;
		if (toastTimeout) clearTimeout(toastTimeout);
		toastTimeout = setTimeout(() => {
			toastMessage = null;
		}, 3500);
	}

	function toggleDropdown(e: Event) {
		e.stopPropagation();
		if (disabled || exporting) return;
		open = !open;
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape' && open) {
			open = false;
		}
	}

	async function handleExport(type: 'csv' | 'pdf') {
		open = false;
		exporting = true;
		try {
			const data = await getData();
			if (!data || data.length === 0) {
				showToast('No data available to export.', 'error');
				return;
			}

			if (type === 'csv') {
				exportToCSV(columns, data, { title, filterSummary });
				showToast(`Exported ${data.length} rows to CSV!`, 'success');
			} else {
				exportToPDF(columns, data, { title, filterSummary });
				showToast(`Exported ${data.length} rows to PDF!`, 'success');
			}
		} catch (err: unknown) {
			const errMsg = err instanceof Error ? err.message : 'Export failed.';
			showToast(errMsg, 'error');
		} finally {
			exporting = false;
		}
	}
</script>

<svelte:window onclick={() => (open = false)} onkeydown={handleKeydown} />

<div class={`relative inline-block ${customClass}`}>
	<button
		type="button"
		disabled={disabled || exporting}
		onclick={toggleDropdown}
		class="inline-flex items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs font-medium text-gray-700 shadow-xs hover:bg-gray-50 focus:outline-none disabled:cursor-not-allowed disabled:opacity-50 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300 dark:hover:bg-gray-800"
		aria-haspopup="true"
		aria-expanded={open}
	>
		{#if exporting}
			<svg class="h-4 w-4 animate-spin text-emerald-600 dark:text-emerald-400" fill="none" viewBox="0 0 24 24">
				<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
				<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
			</svg>
			<span>Exporting...</span>
		{:else}
			<svg class="h-4 w-4 text-gray-500 dark:text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 16.5v2.25A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75V16.5M16.5 12L12 16.5m0 0L7.5 12m4.5 4.5V3" />
			</svg>
			<span>Export</span>
			<svg class="h-3.5 w-3.5 text-gray-400" viewBox="0 0 20 20" fill="currentColor">
				<path fill-rule="evenodd" d="M5.23 7.21a.75.75 0 011.06.02L10 11.168l3.71-3.938a.75.75 0 111.08 1.04l-4.25 4.5a.75.75 0 01-1.08 0l-4.25-4.5a.75.75 0 01.02-1.06z" clip-rule="evenodd" />
			</svg>
		{/if}
	</button>

	{#if open}
		<div
			class="absolute right-0 top-full mt-1.5 z-50 w-44 rounded-xl border border-gray-100 bg-white py-1.5 shadow-xl ring-1 ring-black/5 dark:border-gray-800 dark:bg-gray-900"
		>
			<button
				type="button"
				onclick={() => void handleExport('pdf')}
				class="flex w-full items-center gap-2.5 px-3.5 py-2 text-left text-xs font-medium text-gray-700 hover:bg-gray-50 dark:text-gray-200 dark:hover:bg-gray-800/60"
			>
				<svg class="h-4 w-4 text-rose-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m2.25 0H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z" />
				</svg>
				<span>Export as PDF</span>
			</button>

			<button
				type="button"
				onclick={() => void handleExport('csv')}
				class="flex w-full items-center gap-2.5 px-3.5 py-2 text-left text-xs font-medium text-gray-700 hover:bg-gray-50 dark:text-gray-200 dark:hover:bg-gray-800/60"
			>
				<svg class="h-4 w-4 text-emerald-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3.375 19.5h17.25m-17.25 0a1.125 1.125 0 01-1.125-1.125M3.375 19.5h7.5c.621 0 1.125-.504 1.125-1.125m-9.75 0V5.625m0 12.75v-1.5c0-.621.504-1.125 1.125-1.125m18.375 2.625V5.625m0 12.75c0 .621-.504 1.125-1.125 1.125m1.125-1.125v-1.5c0-.621-.504-1.125-1.125-1.125m0 3.75h-7.5A1.125 1.125 0 0112 18.375m9.75-12.75c0-.621-.504-1.125-1.125-1.125H3.375c-.621 0-1.125.504-1.125 1.125m19.5 0v1.5c0 .621-.504 1.125-1.125 1.125M1.875 5.625v1.5c0 .621.504 1.125 1.125 1.125m0 0h17.25m-17.25 0h7.5c.621 0 1.125.504 1.125 1.125v1.5c0 .621-.504 1.125-1.125 1.125m-7.5-3.75v3.75m0 0h7.5m-7.5 0v3.75m0 0h7.5" />
				</svg>
				<span>Export as CSV</span>
			</button>
		</div>
	{/if}

	{#if toastMessage}
		<div
			class={`fixed bottom-5 right-5 z-50 flex items-center gap-2 rounded-xl px-4 py-2.5 text-xs font-semibold shadow-lg backdrop-blur-sm transition-all duration-300 ${
				toastType === 'success'
					? 'border border-emerald-200 bg-emerald-50 text-emerald-800 dark:border-emerald-800 dark:bg-emerald-950/80 dark:text-emerald-200'
					: 'border border-rose-200 bg-rose-50 text-rose-800 dark:border-rose-800 dark:bg-rose-950/80 dark:text-rose-200'
			}`}
		>
			{#if toastType === 'success'}
				<svg class="h-4 w-4 text-emerald-600 dark:text-emerald-400" viewBox="0 0 20 20" fill="currentColor">
					<path fill-rule="evenodd" d="M16.704 4.153a.75.75 0 0 1 .143 1.052l-8 10.5a.75.75 0 0 1-1.127.075l-4.5-4.5a.75.75 0 0 1 1.06-1.06l3.894 3.893 7.48-9.817a.75.75 0 0 1 1.05-.143Z" clip-rule="evenodd" />
				</svg>
			{:else}
				<svg class="h-4 w-4 text-rose-600 dark:text-rose-400" viewBox="0 0 20 20" fill="currentColor">
					<path fill-rule="evenodd" d="M18 10a8 8 0 1 1-16 0 8 8 0 0 1 16 0Zm-8-5a.75.75 0 0 1 .75.75v4.5a.75.75 0 0 1-1.5 0v-4.5A.75.75 0 0 1 10 5Zm0 10a1 1 0 1 0 0-2 1 1 0 0 0 0 2Z" clip-rule="evenodd" />
				</svg>
			{/if}
			<span>{toastMessage}</span>
		</div>
	{/if}
</div>

