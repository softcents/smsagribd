<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { supermanApi } from '$lib/superman-api.js';

	interface Item {
		id: number;
		name: string;
		priority: number;
		endpoint_url: string;
		http_method: string;
		sender_id: string;
		is_active: boolean;
	}

	let items = $state<Item[]>([]);
	let loading = $state(true);
	let errorMsg = $state('');
	let search = $state('');
	let perPage = $state(10);
	let page = $state(1);
	let total = $state(0);
	let menuId = $state<number | null>(null);
	let timer: ReturnType<typeof setTimeout> | undefined;

	let lastPage = $derived(Math.max(1, Math.ceil(total / perPage)));

	async function load() {
		loading = true;
		errorMsg = '';
		try {
			const q = new URLSearchParams({ page: String(page), per_page: String(perPage) });
			if (search.trim()) q.set('search', search.trim());
			const res: any = await supermanApi.get(`/api/v1/superman/signup-gateways?${q.toString()}`);
			items = res.data ?? [];
			total = res.meta?.total ?? items.length;
		} catch (e: unknown) {
			errorMsg = e instanceof Error ? e.message : 'Unable to load gateways.';
			items = [];
			total = 0;
		} finally {
			loading = false;
		}
	}

	function onSearch() {
		clearTimeout(timer);
		timer = setTimeout(() => {
			page = 1;
			void load();
		}, 300);
	}

	function go(p: number) {
		page = Math.min(Math.max(1, p), lastPage);
		void load();
	}

	async function toggle(g: Item) {
		menuId = null;
		try {
			await supermanApi.put(`/api/v1/superman/signup-gateways/${g.id}/toggle`, {});
			await load();
		} catch (e: unknown) {
			window.alert(e instanceof Error ? e.message : 'Unable to update status.');
		}
	}

	async function remove(g: Item) {
		menuId = null;
		if (!window.confirm(`Delete gateway "${g.name}"?`)) return;
		try {
			await supermanApi.delete(`/api/v1/superman/signup-gateways/${g.id}`);
			if (items.length === 1 && page > 1) page -= 1;
			await load();
		} catch (e: unknown) {
			window.alert(e instanceof Error ? e.message : 'Unable to delete gateway.');
		}
	}

	onMount(() => void load());
</script>

<svelte:head><title>Signup Register Gateways — Superman Administrator</title></svelte:head>
<svelte:window onclick={() => (menuId = null)} />

<div class="space-y-6">
	<div class="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
		<div>
			<nav class="mb-2 text-xs text-gray-500">
				<span>Signup Register Gateways</span><span class="mx-1">›</span><span class="text-gray-800 dark:text-gray-200">List</span>
			</nav>
			<h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">Signup Register Gateways</h1>
			<p class="mt-1 text-xs text-gray-500">OTP gateways for the register page, tried in priority order with automatic failover.</p>
		</div>
		<button type="button" onclick={() => goto('/superman/signup-gateways/create')} class="rounded-lg border border-emerald-600 px-4 py-2 text-sm font-semibold text-emerald-700 hover:bg-emerald-50 dark:text-emerald-400 dark:hover:bg-emerald-950">
			New signup register gateway
		</button>
	</div>

	<div class="rounded-xl border border-gray-200 bg-white shadow-xs dark:border-gray-800 dark:bg-gray-900">
		<div class="flex justify-end border-b border-gray-100 p-4 dark:border-gray-800">
			<input
				type="search"
				placeholder="Search"
				bind:value={search}
				oninput={onSearch}
				class="w-full max-w-xs rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-emerald-500 focus:outline-none dark:border-gray-700 dark:bg-gray-900 dark:text-white"
			/>
		</div>

		<div class="overflow-x-auto">
			<table class="w-full text-left text-sm">
				<thead>
					<tr class="border-b border-gray-100 text-xs font-semibold uppercase tracking-wider text-gray-600 dark:border-gray-800 dark:text-gray-400">
						<th class="px-5 py-3.5">Priority</th>
						<th class="px-5 py-3.5">Name</th>
						<th class="px-5 py-3.5">Endpoint</th>
						<th class="px-5 py-3.5">Method</th>
						<th class="px-5 py-3.5">Sender</th>
						<th class="px-5 py-3.5">Active</th>
						<th class="px-5 py-3.5"></th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
					{#if loading}
						<tr><td colspan="7" class="px-5 py-10 text-center text-gray-500">Loading…</td></tr>
					{:else if errorMsg}
						<tr><td colspan="7" class="px-5 py-10 text-center text-red-600">{errorMsg}</td></tr>
					{:else if items.length === 0}
						<tr><td colspan="7" class="px-5 py-10 text-center text-gray-500">No signup register gateways yet. Create one to enable failover OTP sending.</td></tr>
					{:else}
						{#each items as g (g.id)}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30">
								<td class="px-5 py-4"><span class="rounded-md border border-emerald-200 bg-emerald-50 px-2 py-0.5 text-xs font-semibold text-emerald-700">Priority {g.priority}</span></td>
								<td class="px-5 py-4 font-bold text-gray-950 dark:text-white">{g.name}</td>
								<td class="max-w-xs truncate px-5 py-4 text-gray-700 dark:text-gray-300" title={g.endpoint_url}>{g.endpoint_url}</td>
								<td class="px-5 py-4"><span class="rounded-md bg-emerald-50 px-2 py-0.5 text-xs font-semibold text-emerald-700">{g.http_method}</span></td>
								<td class="px-5 py-4 text-gray-700 dark:text-gray-300">{g.sender_id}</td>
								<td class="px-5 py-4">
									{#if g.is_active}
										<span class="inline-flex h-6 w-6 items-center justify-center rounded-full border border-emerald-500 text-emerald-600" title="Active">✓</span>
									{:else}
										<span class="inline-flex h-6 w-6 items-center justify-center rounded-full border border-gray-300 text-gray-400" title="Inactive">✕</span>
									{/if}
								</td>
								<td class="relative px-5 py-4 text-right">
									<button
										type="button"
										aria-label="Row actions"
										onclick={(e) => {
											e.stopPropagation();
											menuId = menuId === g.id ? null : g.id;
										}}
										class="rounded-lg border border-emerald-500 px-3 py-1.5 text-emerald-700 hover:bg-emerald-50"
									>⋮</button>
									{#if menuId === g.id}
										<div class="absolute right-5 top-14 z-20 w-44 rounded-lg border border-gray-200 bg-white py-1 text-left text-sm shadow-lg dark:border-gray-700 dark:bg-gray-900">
											<button type="button" class="block w-full px-4 py-2 text-left hover:bg-gray-50 dark:hover:bg-gray-800" onclick={() => goto(`/superman/signup-gateways/${g.id}/edit`)}>Edit</button>
											<button type="button" class="block w-full px-4 py-2 text-left hover:bg-gray-50 dark:hover:bg-gray-800" onclick={() => toggle(g)}>{g.is_active ? 'Deactivate' : 'Activate'}</button>
											<button type="button" class="block w-full px-4 py-2 text-left text-red-600 hover:bg-red-50 dark:hover:bg-gray-800" onclick={() => remove(g)}>Delete</button>
										</div>
									{/if}
								</td>
							</tr>
						{/each}
					{/if}
				</tbody>
			</table>
		</div>

		<div class="flex flex-wrap items-center justify-between gap-3 border-t border-gray-100 px-5 py-4 text-sm text-gray-600 dark:border-gray-800 dark:text-gray-400">
			<span>Showing {total === 0 ? 0 : (page - 1) * perPage + 1} to {Math.min(page * perPage, total)} of {total} results</span>
			<div class="flex items-center gap-3">
				<label class="flex items-center gap-2">
					Per page
					<select bind:value={perPage} onchange={() => { page = 1; void load(); }} class="rounded-lg border border-gray-300 px-2 py-1 dark:border-gray-700 dark:bg-gray-900">
						<option value={10}>10</option>
						<option value={25}>25</option>
						<option value={50}>50</option>
					</select>
				</label>
				<button type="button" disabled={page <= 1} onclick={() => go(page - 1)} class="rounded-lg border border-gray-300 px-3 py-1 disabled:opacity-40 dark:border-gray-700">‹</button>
				<span class="font-semibold">{page} / {lastPage}</span>
				<button type="button" disabled={page >= lastPage} onclick={() => go(page + 1)} class="rounded-lg border border-gray-300 px-3 py-1 disabled:opacity-40 dark:border-gray-700">›</button>
			</div>
		</div>
	</div>
</div>
