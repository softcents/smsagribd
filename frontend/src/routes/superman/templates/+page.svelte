<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface TemplateItem {
		id: number;
		name: string;
		body: string;
		customer: string;
		approval: string;
		sender_id?: string;
		created_at: string;
	}

	let templates = $state<TemplateItem[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let search = $state('');

	async function loadTemplates() {
		loading = true;
		error = null;
		try {
			const res = await supermanApi.get<{ data: TemplateItem[] }>('/api/v1/superman/templates');
			templates = res.data ?? [];
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Failed to load templates.';
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		void loadTemplates();
	});

	let filtered = $derived(
		templates.filter(t => search === '' || t.name.toLowerCase().includes(search.toLowerCase()) || t.body.toLowerCase().includes(search.toLowerCase()))
	);

	const exportColumns: ExportColumn<TemplateItem>[] = [
		{ header: 'Template Name', getValue: (t) => t.name },
		{ header: 'Customer', getValue: (t) => t.customer },
		{ header: 'Sender ID', getValue: (t) => t.sender_id || '—' },
		{ header: 'Template Body', getValue: (t) => t.body },
		{ header: 'Approval Status', getValue: (t) => t.approval.toUpperCase() },
		{ header: 'Created Date', getValue: (t) => (t.created_at ? new Date(t.created_at).toLocaleDateString() : '—') }
	];
</script>

<svelte:head>
	<title>Bulk SMS Templates — Superman Administrator</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">Bulk SMS Content Templates</h1>
			<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
				Pre-approved templates, transactional formats, and compliance checks for customer messaging.
			</p>
		</div>
		<div class="flex items-center gap-2">
			<button onclick={loadTemplates} class="rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs font-medium text-gray-700 shadow-xs hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300">
				Refresh Templates
			</button>
			<ExportDropdown
				title="SMS Content Templates"
				columns={exportColumns}
				getData={() => filtered}
				disabled={filtered.length === 0}
			/>
		</div>
	</div>

	<div class="flex items-center gap-3 rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
		<input
			type="search"
			bind:value={search}
			placeholder="Search by template title or body..."
			class="w-full max-w-sm rounded-lg border border-gray-200 bg-gray-50/70 px-3 py-1.5 text-xs outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
		/>
	</div>

	<div class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-xs dark:border-gray-800 dark:bg-gray-900">
		{#if loading}
			<div class="py-16 text-center text-xs text-gray-500">Loading templates...</div>
		{:else if error}
			<div class="p-6 text-center text-xs text-rose-600">{error}</div>
		{:else if filtered.length === 0}
			<div class="py-16 text-center text-xs text-gray-500">No template records registered.</div>
		{:else}
			<div class="overflow-x-auto">
				<table class="w-full text-left text-xs">
					<thead>
						<tr class="border-b border-gray-100 bg-gray-50/50 text-[10px] font-semibold uppercase tracking-wider text-gray-500 dark:border-gray-800 dark:bg-gray-800/50 dark:text-gray-400">
							<th class="px-4 py-3">Template Name</th>
							<th class="px-4 py-3">Account</th>
							<th class="px-4 py-3">Sender Header</th>
							<th class="px-4 py-3">Content Body</th>
							<th class="px-4 py-3">Approval State</th>
							<th class="px-4 py-3 text-right">Created At</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
						{#each filtered as t}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30">
								<td class="px-4 py-3 font-semibold text-gray-950 dark:text-white">{t.name}</td>
								<td class="px-4 py-3 text-gray-600 dark:text-gray-400">{t.customer}</td>
								<td class="px-4 py-3 font-mono font-medium text-blue-600">{t.sender_id || 'Any'}</td>
								<td class="max-w-md truncate px-4 py-3 text-gray-700 dark:text-gray-300" title={t.body}>{t.body}</td>
								<td class="px-4 py-3">
									<span class="inline-flex rounded-full px-2 py-0.5 text-[10px] font-bold uppercase {t.approval === 'approved' ? 'bg-emerald-50 text-emerald-700' : 'bg-amber-50 text-amber-700'}">
										{t.approval}
									</span>
								</td>
								<td class="px-4 py-3 text-right text-gray-400">{new Date(t.created_at).toLocaleDateString()}</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</div>
</div>

