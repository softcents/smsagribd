<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { supermanApi } from '$lib/superman-api.js';

	interface Op {
		id: number;
		name: string;
		codes: string;
		is_active: boolean;
		country_name: string;
	}

	let all = $state<Op[]>([]);
	let loading = $state(true);
	let errorMsg = $state('');
	let search = $state('');
	let perPage = $state(10);
	let page = $state(1);
	let sortKey = $state<'country_name' | 'name'>('name');
	let sortDir = $state<'asc' | 'desc'>('asc');
	let filterOpen = $state(false);
	let fCountry = $state('');
	let fStatus = $state('');
	let selected = $state<number[]>([]);

	function codesText(raw: string): string {
		try {
			const v = JSON.parse(raw);
			return Array.isArray(v) ? v.join(', ') : String(raw);
		} catch {
			return raw;
		}
	}

	let countries = $derived([...new Set(all.map((o) => o.country_name))].sort());
	let activeFilters = $derived((fCountry ? 1 : 0) + (fStatus ? 1 : 0));

	let filtered = $derived.by(() => {
		const q = search.trim().toLowerCase();
		let list = all.filter((o) => {
			if (fCountry && o.country_name !== fCountry) return false;
			if (fStatus === 'active' && !o.is_active) return false;
			if (fStatus === 'inactive' && o.is_active) return false;
			if (!q) return true;
			return o.name.toLowerCase().includes(q) || o.country_name.toLowerCase().includes(q) || codesText(o.codes).includes(q);
		});
		list = [...list].sort((a, b) => {
			const r = a[sortKey].localeCompare(b[sortKey]);
			return sortDir === 'asc' ? r : -r;
		});
		return list;
	});

	let total = $derived(filtered.length);
	let lastPage = $derived(Math.max(1, Math.ceil(total / perPage)));
	let rows = $derived(filtered.slice((page - 1) * perPage, page * perPage));
	let allOnPage = $derived(rows.length > 0 && rows.every((o) => selected.includes(o.id)));

	function sortBy(key: 'country_name' | 'name') {
		if (sortKey === key) sortDir = sortDir === 'asc' ? 'desc' : 'asc';
		else {
			sortKey = key;
			sortDir = 'asc';
		}
		page = 1;
	}

	function toggleAll() {
		if (allOnPage) selected = selected.filter((id) => !rows.some((o) => o.id === id));
		else selected = [...new Set([...selected, ...rows.map((o) => o.id)])];
	}

	function toggleOne(id: number) {
		selected = selected.includes(id) ? selected.filter((x) => x !== id) : [...selected, id];
	}

	$effect(() => {
		search;
		fCountry;
		fStatus;
		perPage;
		page = 1;
	});

	onMount(async () => {
		try {
			const res: any = await supermanApi.get('/api/v1/superman/operators');
			all = res.data ?? [];
		} catch (e: unknown) {
			errorMsg = e instanceof Error ? e.message : 'Unable to load operators.';
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head><title>Operators — Superman Administrator</title></svelte:head>
<svelte:window onclick={() => (filterOpen = false)} />

<div class="space-y-6">
	<div class="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
		<div>
			<nav class="mb-2 text-xs text-gray-500"><span>Operators</span><span class="mx-1">›</span><span class="text-gray-800 dark:text-gray-200">List</span></nav>
			<h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">Operators</h1>
		</div>
		<button type="button" onclick={() => goto('/superman/operators/create')} class="rounded-lg border border-emerald-600 px-4 py-2 text-sm font-semibold text-emerald-700 hover:bg-emerald-50 dark:text-emerald-400 dark:hover:bg-emerald-950">New operator</button>
	</div>

	<div class="rounded-xl border border-gray-200 bg-white shadow-xs dark:border-gray-800 dark:bg-gray-900">
		<div class="flex items-center justify-between gap-3 border-b border-gray-100 p-4 dark:border-gray-800">
			<span class="text-xs text-gray-500">{selected.length > 0 ? `${selected.length} selected` : ''}</span>
			<div class="relative flex items-center gap-3">
				<input type="search" placeholder="Search" bind:value={search} class="w-64 rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-emerald-500 focus:outline-none dark:border-gray-700 dark:bg-gray-900 dark:text-white" />
				<button type="button" aria-label="Filters" onclick={(e) => { e.stopPropagation(); filterOpen = !filterOpen; }} class="relative rounded-lg border border-gray-300 px-3 py-2 text-sm dark:border-gray-700">
					⚲ Filter
					{#if activeFilters > 0}<span class="absolute -right-2 -top-2 rounded-full bg-emerald-600 px-1.5 text-[10px] font-bold text-white">{activeFilters}</span>{/if}
				</button>
				{#if filterOpen}
					<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
					<div onclick={(e) => e.stopPropagation()} class="absolute right-0 top-12 z-20 w-64 space-y-3 rounded-lg border border-gray-200 bg-white p-4 text-sm shadow-lg dark:border-gray-700 dark:bg-gray-900">
						<div>
							<label class="mb-1 block font-semibold" for="f-country">Country</label>
							<select id="f-country" bind:value={fCountry} class="w-full rounded-lg border border-gray-300 px-2 py-1.5 dark:border-gray-700 dark:bg-gray-900">
								<option value="">All</option>
								{#each countries as c}<option value={c}>{c}</option>{/each}
							</select>
						</div>
						<div>
							<label class="mb-1 block font-semibold" for="f-status">Active</label>
							<select id="f-status" bind:value={fStatus} class="w-full rounded-lg border border-gray-300 px-2 py-1.5 dark:border-gray-700 dark:bg-gray-900">
								<option value="">All</option>
								<option value="active">Active</option>
								<option value="inactive">Inactive</option>
							</select>
						</div>
						<button type="button" onclick={() => { fCountry = ''; fStatus = ''; }} class="text-xs font-semibold text-red-600">Reset filters</button>
					</div>
				{/if}
			</div>
		</div>

		<div class="overflow-x-auto">
			<table class="w-full text-left text-sm">
				<thead>
					<tr class="border-b border-gray-100 text-xs font-semibold uppercase tracking-wider text-gray-600 dark:border-gray-800 dark:text-gray-400">
						<th class="w-10 px-5 py-3.5"><input type="checkbox" checked={allOnPage} onchange={toggleAll} aria-label="Select all" /></th>
						<th class="px-5 py-3.5"><button type="button" onclick={() => sortBy('country_name')} class="font-semibold uppercase">Country {sortKey === 'country_name' ? (sortDir === 'asc' ? '▲' : '▼') : '⌄'}</button></th>
						<th class="px-5 py-3.5"><button type="button" onclick={() => sortBy('name')} class="font-semibold uppercase">Name {sortKey === 'name' ? (sortDir === 'asc' ? '▲' : '▼') : '⌄'}</button></th>
						<th class="px-5 py-3.5">Codes</th>
						<th class="px-5 py-3.5">Active</th>
						<th class="px-5 py-3.5"></th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
					{#if loading}
						<tr><td colspan="6" class="px-5 py-10 text-center text-gray-500">Loading…</td></tr>
					{:else if errorMsg}
						<tr><td colspan="6" class="px-5 py-10 text-center text-red-600">{errorMsg}</td></tr>
					{:else if rows.length === 0}
						<tr><td colspan="6" class="px-5 py-10 text-center text-gray-500">No operators found.</td></tr>
					{:else}
						{#each rows as o (o.id)}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30">
								<td class="px-5 py-4"><input type="checkbox" checked={selected.includes(o.id)} onchange={() => toggleOne(o.id)} aria-label="Select row" /></td>
								<td class="px-5 py-4">{o.country_name}</td>
								<td class="px-5 py-4">{o.name}</td>
								<td class="px-5 py-4">{codesText(o.codes)}</td>
								<td class="px-5 py-4">
									{#if o.is_active}<span class="inline-flex h-6 w-6 items-center justify-center rounded-full border border-emerald-500 text-emerald-600">✓</span>{:else}<span class="inline-flex h-6 w-6 items-center justify-center rounded-full border border-gray-300 text-gray-400">✕</span>{/if}
								</td>
								<td class="px-5 py-4 text-right font-semibold">
									<a href={`/superman/operators/${o.id}`} class="mr-4 text-sky-600 hover:underline">◉ View</a>
									<a href={`/superman/operators/${o.id}/edit`} class="text-amber-600 hover:underline">✎ Edit</a>
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
				<label class="flex items-center gap-2">Per page
					<select bind:value={perPage} class="rounded-lg border border-gray-300 px-2 py-1 dark:border-gray-700 dark:bg-gray-900">
						<option value={10}>10</option><option value={25}>25</option><option value={50}>50</option>
					</select>
				</label>
				<button type="button" disabled={page <= 1} onclick={() => (page -= 1)} class="rounded-lg border border-gray-300 px-3 py-1 disabled:opacity-40 dark:border-gray-700">‹</button>
				<span class="font-semibold">{page} / {lastPage}</span>
				<button type="button" disabled={page >= lastPage} onclick={() => (page += 1)} class="rounded-lg border border-gray-300 px-3 py-1 disabled:opacity-40 dark:border-gray-700">›</button>
			</div>
		</div>
	</div>
</div>
