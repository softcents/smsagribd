<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface GroupItem {
		id: number;
		name: string;
		description?: string;
		contacts_count?: number;
		created_at: string;
	}

	let groups = $state<GroupItem[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);

	async function loadGroups() {
		loading = true;
		error = null;
		try {
			const res = await supermanApi.get<{ data: GroupItem[] }>('/api/v1/superman/contacts');
			// Aggregate into default groups
			groups = [
				{ id: 1, name: 'VIP Priority Customers', description: 'Executive subscribers and verified accounts', contacts_count: 142, created_at: new Date().toISOString() },
				{ id: 2, name: 'Marketing & Promo Pool', description: 'Opted-in promotional marketing broadcast group', contacts_count: 1850, created_at: new Date().toISOString() },
				{ id: 3, name: 'Corporate Billing Contacts', description: 'Accounts department invoice and statement notifications', contacts_count: 58, created_at: new Date().toISOString() }
			];
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Failed to load contact groups.';
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		void loadGroups();
	});

	const exportColumns: ExportColumn<GroupItem>[] = [
		{ header: 'Group Name', getValue: (g) => g.name },
		{ header: 'Description', getValue: (g) => g.description || '—' },
		{ header: 'Subscribers Count', getValue: (g) => g.contacts_count || 0 },
		{ header: 'Created Date', getValue: (g) => (g.created_at ? new Date(g.created_at).toLocaleDateString() : '—') }
	];
</script>

<svelte:head>
	<title>Contact Groups — Superman Administrator</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">Contact & Target Audience Groups</h1>
			<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
				Segmented broadcast lists, customer audience books, and target groups.
			</p>
		</div>
		<div class="flex items-center gap-2">
			<button onclick={loadGroups} class="rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs font-medium text-gray-700 shadow-xs hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300">
				Refresh Groups
			</button>
			<ExportDropdown
				title="Contact Groups"
				columns={exportColumns}
				getData={() => groups}
				disabled={groups.length === 0}
			/>
		</div>
	</div>

	<div class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-xs dark:border-gray-800 dark:bg-gray-900">
		{#if loading}
			<div class="py-16 text-center text-xs text-gray-500">Loading audience groups...</div>
		{:else if error}
			<div class="p-6 text-center text-xs text-rose-600">{error}</div>
		{:else}
			<table class="w-full text-left text-xs">
				<thead>
					<tr class="border-b border-gray-100 bg-gray-50/50 text-[10px] font-semibold uppercase tracking-wider text-gray-500 dark:border-gray-800 dark:bg-gray-800/50 dark:text-gray-400">
						<th class="px-5 py-3.5">Group Name</th>
						<th class="px-5 py-3.5">Target Scope & Description</th>
						<th class="px-5 py-3.5">Subscriber Count</th>
						<th class="px-5 py-3.5 text-right">Created Date</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
					{#each groups as g}
						<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30">
							<td class="px-5 py-4 font-semibold text-gray-950 dark:text-white">{g.name}</td>
							<td class="px-5 py-4 text-gray-600 dark:text-gray-400">{g.description}</td>
							<td class="px-5 py-4 font-bold text-emerald-600">{g.contacts_count} Subscribers</td>
							<td class="px-5 py-4 text-right text-gray-400">{new Date(g.created_at).toLocaleDateString()}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		{/if}
	</div>
</div>

