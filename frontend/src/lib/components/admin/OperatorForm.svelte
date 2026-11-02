<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { supermanApi } from '$lib/superman-api.js';

	let { id = 0 }: { id?: number } = $props();
	let editing = $derived(id > 0);

	const blank = () => ({ country_id: 0, name: '', codes: '', masking_price: '', non_masking_price: '', is_active: true });
	let form = $state(blank());
	let countries = $state<{ id: number; name: string }[]>([]);
	let saving = $state(false);
	let loading = $state(true);
	let errorMsg = $state('');

	onMount(async () => {
		try {
			const co: any = await supermanApi.get('/api/v1/superman/countries');
			countries = co.data ?? [];
			if (id > 0) {
				const res: any = await supermanApi.get(`/api/v1/superman/operators/${id}`);
				const d = res.data;
				form = {
					country_id: d.country_id,
					name: d.name,
					codes: d.codes,
					masking_price: String(d.masking_price),
					non_masking_price: String(d.non_masking_price),
					is_active: d.is_active
				};
			}
		} catch (e: unknown) {
			errorMsg = e instanceof Error ? e.message : 'Unable to load data.';
		} finally {
			loading = false;
		}
	});

	async function save(another = false) {
		errorMsg = '';
		saving = true;
		try {
			const payload = {
				country_id: Number(form.country_id),
				name: form.name,
				codes: form.codes,
				masking_price: form.masking_price === '' ? null : Number(form.masking_price),
				non_masking_price: form.non_masking_price === '' ? null : Number(form.non_masking_price),
				is_active: form.is_active
			};
			if (editing) {
				await supermanApi.put(`/api/v1/superman/operators/${id}`, payload);
				await goto('/superman/operators');
			} else {
				await supermanApi.post('/api/v1/superman/operators', payload);
				if (another) {
					form = blank();
					window.scrollTo({ top: 0, behavior: 'smooth' });
				} else {
					await goto('/superman/operators');
				}
			}
		} catch (e: unknown) {
			errorMsg = e instanceof Error ? e.message : 'Unable to save operator.';
		} finally {
			saving = false;
		}
	}

	const input = 'w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm text-gray-900 focus:border-emerald-500 focus:outline-none focus:ring-1 focus:ring-emerald-500 dark:border-gray-700 dark:bg-gray-900 dark:text-white';
	const label = 'mb-1 block text-sm font-semibold text-gray-800 dark:text-gray-200';
</script>

<div class="space-y-6">
	<div>
		<nav class="mb-2 text-xs text-gray-500">
			<a href="/superman/operators" class="hover:text-emerald-600">Operators</a><span class="mx-1">›</span>
			<span class="text-gray-800 dark:text-gray-200">{editing ? 'Edit' : 'Create'}</span>
		</nav>
		<h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">{editing ? 'Edit Operator' : 'Create Operator'}</h1>
	</div>

	{#if errorMsg}
		<div class="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">{errorMsg}</div>
	{/if}

	{#if loading}
		<div class="text-sm text-gray-500">Loading…</div>
	{:else}
		<div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
			<div class="rounded-xl border border-gray-200 bg-white shadow-xs dark:border-gray-800 dark:bg-gray-900">
				<div class="border-b border-gray-100 px-6 py-4 dark:border-gray-800"><h2 class="text-base font-semibold text-gray-950 dark:text-white">Details</h2></div>
				<div class="grid grid-cols-1 gap-5 p-6 sm:grid-cols-2">
					<div>
						<label class={label} for="op-country">Country <span class="text-red-500">*</span></label>
						<select id="op-country" class={input} bind:value={form.country_id}>
							<option value={0} disabled>Select an option</option>
							{#each countries as c (c.id)}<option value={c.id}>{c.name}</option>{/each}
						</select>
					</div>
					<div>
						<label class={label} for="op-name">Name <span class="text-red-500">*</span></label>
						<input id="op-name" class={input} placeholder="Grameenphone" bind:value={form.name} />
					</div>
					<div class="sm:col-span-2">
						<label class={label} for="op-codes">Codes <span class="text-red-500">*</span></label>
						<input id="op-codes" class={input} placeholder="17, 13" bind:value={form.codes} />
						<p class="mt-1 text-xs text-gray-500">Operator codes used in phone numbers (digits only, e.g. 17 for GP).</p>
					</div>
					<div class="sm:col-span-2">
						<span class={label}>Active</span>
						<label class="inline-flex cursor-pointer items-center">
							<input type="checkbox" class="peer sr-only" bind:checked={form.is_active} />
							<span class="relative h-6 w-11 rounded-full bg-gray-300 transition peer-checked:bg-emerald-600 after:absolute after:left-0.5 after:top-0.5 after:h-5 after:w-5 after:rounded-full after:bg-white after:transition peer-checked:after:translate-x-5"></span>
						</label>
					</div>
				</div>
			</div>

			<div class="rounded-xl border border-gray-200 bg-white shadow-xs dark:border-gray-800 dark:bg-gray-900">
				<div class="border-b border-gray-100 px-6 py-4 dark:border-gray-800">
					<h2 class="text-base font-semibold text-gray-950 dark:text-white">Default rates</h2>
					<p class="mt-1 text-sm text-gray-500">Per-SMS-part rate. Leave blank to require a per-user override before pricing.</p>
				</div>
				<div class="grid grid-cols-1 gap-5 p-6 sm:grid-cols-2">
					<div>
						<label class={label} for="op-mask">Masking price (per part)</label>
						<div class="flex items-center rounded-lg border border-gray-300 dark:border-gray-700">
							<span class="border-r border-gray-300 px-3 text-sm text-gray-500 dark:border-gray-700">৳</span>
							<input id="op-mask" type="number" step="0.0001" min="0" class="w-full bg-transparent px-3 py-2 text-sm focus:outline-none dark:text-white" bind:value={form.masking_price} />
						</div>
					</div>
					<div>
						<label class={label} for="op-nonmask">Non-masking price (per part)</label>
						<div class="flex items-center rounded-lg border border-gray-300 dark:border-gray-700">
							<span class="border-r border-gray-300 px-3 text-sm text-gray-500 dark:border-gray-700">৳</span>
							<input id="op-nonmask" type="number" step="0.0001" min="0" class="w-full bg-transparent px-3 py-2 text-sm focus:outline-none dark:text-white" bind:value={form.non_masking_price} />
						</div>
					</div>
				</div>
			</div>
		</div>

		<div class="flex flex-wrap gap-3">
			<button type="button" disabled={saving} onclick={() => save(false)} class="rounded-lg bg-emerald-600 px-4 py-2 text-sm font-bold text-white hover:bg-emerald-700 disabled:opacity-50">
				{saving ? 'Saving…' : editing ? 'Save changes' : 'Create'}
			</button>
			{#if !editing}
				<button type="button" disabled={saving} onclick={() => save(true)} class="rounded-lg border border-gray-300 px-4 py-2 text-sm font-semibold text-gray-800 hover:bg-gray-50 disabled:opacity-50 dark:border-gray-700 dark:text-gray-200">Create &amp; create another</button>
			{/if}
			<a href="/superman/operators" class="rounded-lg border border-gray-300 px-4 py-2 text-sm font-semibold text-gray-800 hover:bg-gray-50 dark:border-gray-700 dark:text-gray-200">Cancel</a>
		</div>
	{/if}
</div>
