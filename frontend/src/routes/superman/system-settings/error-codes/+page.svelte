<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface ErrorCode {
		code: number;
		name: string;
		category: string;
		http_status: number;
		description: string;
	}

	let codes = $state<ErrorCode[]>([]);
	let search = $state('');
	let loading = $state(true);

	const exportColumns: ExportColumn<ErrorCode>[] = [
		{ header: 'Error Code', getValue: (c) => c.code },
		{ header: 'Name / Identifier', getValue: (c) => c.name },
		{ header: 'Category', getValue: (c) => c.category },
		{ header: 'HTTP Mapping', getValue: (c) => c.http_status },
		{ header: 'Description', getValue: (c) => c.description }
	];

	onMount(async () => {
		try {
			const res = await supermanApi.get<{ data: ErrorCode[] }>('/api/v1/superman/system-settings/error-codes');
			if (res.data) codes = res.data;
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	});

	let filteredCodes = $derived(
		codes.filter(
			(c) =>
				c.code.toString().includes(search) ||
				c.name.toLowerCase().includes(search.toLowerCase()) ||
				c.description.toLowerCase().includes(search.toLowerCase()) ||
				c.category.toLowerCase().includes(search.toLowerCase())
		)
	);
</script>

<div class="space-y-6">
	<!-- Page Header -->
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-xl font-bold text-gray-900 dark:text-white">Standard API Response & Error Code Directory</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
				Authoritative standard GSM, HTTP, and internal routing response codes returned by the SMS gateway engine.
			</p>
		</div>
		<div class="flex items-center gap-3">
			<ExportDropdown
				title="Standard API Response Error Codes"
				columns={exportColumns}
				getData={() => filteredCodes}
				disabled={filteredCodes.length === 0}
			/>
			<input
				type="text"
				placeholder="Search error codes or names..."
				bind:value={search}
				class="rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 px-3 py-1.5 text-xs text-gray-900 dark:text-white focus:outline-none focus:ring-1 focus:ring-emerald-500 w-64 shadow-xs"
			/>
		</div>
	</div>

	<!-- Error Codes Table -->
	<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl shadow-xs overflow-hidden">
		<div class="overflow-x-auto">
			<table class="w-full text-left text-xs">
				<thead class="bg-gray-50/75 dark:bg-gray-800/50 text-gray-500 uppercase tracking-wider font-semibold border-b border-gray-200 dark:border-gray-800">
					<tr>
						<th class="px-5 py-3">Code</th>
						<th class="px-5 py-3">Constant / Identifier</th>
						<th class="px-5 py-3">Category Range</th>
						<th class="px-5 py-3">HTTP Status</th>
						<th class="px-5 py-3">Meaning / Client Message</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800 text-gray-700 dark:text-gray-300">
					{#if loading}
						<tr>
							<td colspan="5" class="px-5 py-8 text-center text-gray-400">Loading error code directory...</td>
						</tr>
					{:else}
						{#each filteredCodes as c}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30 transition">
								<td class="px-5 py-3 font-mono font-bold text-gray-900 dark:text-white">
									<span class="inline-flex items-center px-2 py-0.5 rounded text-xs {c.code < 2000 ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300' : c.code < 4000 ? 'bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-300' : 'bg-red-100 text-red-800 dark:bg-red-950 dark:text-red-300'}">
										{c.code}
									</span>
								</td>
								<td class="px-5 py-3 font-mono font-semibold text-gray-800 dark:text-gray-200">{c.name}</td>
								<td class="px-5 py-3 text-gray-500">{c.category}</td>
								<td class="px-5 py-3 font-mono font-medium text-gray-700 dark:text-gray-300">{c.http_status}</td>
								<td class="px-5 py-3 text-gray-600 dark:text-gray-400">{c.description}</td>
							</tr>
						{/each}
					{/if}
				</tbody>
			</table>
		</div>
	</div>
</div>

