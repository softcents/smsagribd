<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { supermanApi } from '$lib/superman-api.js';

	let d = $state<any>(null);
	let countryName = $state('');
	let errorMsg = $state('');

	onMount(async () => {
		try {
			const [op, co]: any[] = await Promise.all([
				supermanApi.get(`/api/v1/superman/operators/${$page.params.id}`),
				supermanApi.get('/api/v1/superman/countries')
			]);
			d = op.data;
			countryName = (co.data ?? []).find((c: any) => c.id === d.country_id)?.name ?? '';
		} catch (e: unknown) {
			errorMsg = e instanceof Error ? e.message : 'Unable to load operator.';
		}
	});
</script>

<svelte:head><title>Operator — Superman Administrator</title></svelte:head>

<div class="space-y-6">
	<div class="flex items-end justify-between">
		<div>
			<nav class="mb-2 text-xs text-gray-500">
				<a href="/superman/operators" class="hover:text-emerald-600">Operators</a><span class="mx-1">›</span><span class="text-gray-800 dark:text-gray-200">View</span>
			</nav>
			<h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">{d?.name ?? 'Operator'}</h1>
		</div>
		{#if d}<a href={`/superman/operators/${d.id}/edit`} class="rounded-lg border border-amber-500 px-4 py-2 text-sm font-semibold text-amber-700 hover:bg-amber-50">Edit</a>{/if}
	</div>

	{#if errorMsg}
		<div class="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">{errorMsg}</div>
	{:else if !d}
		<div class="text-sm text-gray-500">Loading…</div>
	{:else}
		<div class="max-w-3xl rounded-xl border border-gray-200 bg-white p-6 shadow-xs dark:border-gray-800 dark:bg-gray-900">
			<dl class="grid grid-cols-1 gap-5 text-sm sm:grid-cols-2">
				<div><dt class="text-gray-500">Country</dt><dd class="mt-1 font-semibold text-gray-950 dark:text-white">{countryName}</dd></div>
				<div><dt class="text-gray-500">Name</dt><dd class="mt-1 font-semibold text-gray-950 dark:text-white">{d.name}</dd></div>
				<div><dt class="text-gray-500">Codes</dt><dd class="mt-1 font-semibold text-gray-950 dark:text-white">{d.codes}</dd></div>
				<div><dt class="text-gray-500">Active</dt><dd class="mt-1 font-semibold">{d.is_active ? '✓ Yes' : '✕ No'}</dd></div>
				<div><dt class="text-gray-500">Masking price (per part)</dt><dd class="mt-1 font-semibold text-gray-950 dark:text-white">৳{Number(d.masking_price).toFixed(4)}</dd></div>
				<div><dt class="text-gray-500">Non-masking price (per part)</dt><dd class="mt-1 font-semibold text-gray-950 dark:text-white">৳{Number(d.non_masking_price).toFixed(4)}</dd></div>
			</dl>
		</div>
	{/if}
</div>
