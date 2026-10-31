<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface ContactItem {
		id: number;
		name: string;
		mobile: string;
		email?: string;
		created_at: string;
	}

	let contacts = $state<ContactItem[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let search = $state('');

	async function loadContacts() {
		loading = true;
		error = null;
		try {
			const params = new URLSearchParams({ per_page: '100' });
			if (search.trim()) params.set('search', search.trim());
			const res = await supermanApi.get<{ data: ContactItem[] }>(`/api/v1/superman/contacts?${params.toString()}`);
			contacts = res.data ?? [];
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Failed to load contacts directory.';
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		void loadContacts();
	});

	const exportColumns: ExportColumn<ContactItem>[] = [
		{ header: 'Contact Name', getValue: (c) => c.name },
		{ header: 'Mobile Number', getValue: (c) => c.mobile },
		{ header: 'Email', getValue: (c) => c.email || '—' },
		{ header: 'Created Date', getValue: (c) => (c.created_at ? new Date(c.created_at).toLocaleDateString() : '—') }
	];
</script>

<svelte:head>
	<title>Contacts — Superman Administrator</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">Contacts & Directory</h1>
			<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
				Global audience registry, subscriber MSISDNs, and customer address book entries.
			</p>
		</div>
		<div class="flex items-center gap-2">
			<button onclick={loadContacts} class="rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs font-medium text-gray-700 shadow-xs hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300">
				Refresh Contacts
			</button>
			<ExportDropdown
				title="Contacts Directory"
				columns={exportColumns}
				getData={() => contacts}
				disabled={contacts.length === 0}
			/>
		</div>
	</div>

	<div class="flex items-center gap-3 rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
		<input
			type="search"
			bind:value={search}
			onkeydown={(e) => e.key === 'Enter' && loadContacts()}
			placeholder="Search by contact name or phone number..."
			class="w-full max-w-sm rounded-lg border border-gray-200 bg-gray-50/70 px-3 py-1.5 text-xs outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
		/>
	</div>

	<div class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-xs dark:border-gray-800 dark:bg-gray-900">
		{#if loading}
			<div class="py-16 text-center text-xs text-gray-500">Loading contacts...</div>
		{:else if error}
			<div class="p-6 text-center text-xs text-rose-600">{error}</div>
		{:else if contacts.length === 0}
			<div class="py-16 text-center text-xs text-gray-500">No contacts saved in address books.</div>
		{:else}
			<div class="overflow-x-auto">
				<table class="w-full text-left text-xs">
					<thead>
						<tr class="border-b border-gray-100 bg-gray-50/50 text-[10px] font-semibold uppercase tracking-wider text-gray-500 dark:border-gray-800 dark:bg-gray-800/50 dark:text-gray-400">
							<th class="px-5 py-3.5">Contact Name</th>
							<th class="px-5 py-3.5">Mobile MSISDN</th>
							<th class="px-5 py-3.5">Email Address</th>
							<th class="px-5 py-3.5 text-right">Created Date</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
						{#each contacts as c}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30">
								<td class="px-5 py-3.5 font-semibold text-gray-950 dark:text-white">{c.name}</td>
								<td class="px-5 py-3.5 font-mono text-gray-800 dark:text-gray-200">{c.mobile}</td>
								<td class="px-5 py-3.5 text-gray-500">{c.email || '—'}</td>
								<td class="px-5 py-3.5 text-right text-gray-400">{new Date(c.created_at).toLocaleDateString()}</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</div>
</div>

