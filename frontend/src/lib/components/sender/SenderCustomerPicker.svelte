<script lang="ts">
	import type { SenderUserRef } from '$lib/sender-cli.js';

	interface Props {
		customers: SenderUserRef[];
		selected: number[];
		onSelectedChange?: (selected: number[]) => void;
		emptyText?: string;
	}

	let { customers, selected = $bindable([]), onSelectedChange, emptyText = 'No customer accounts available.' }: Props = $props();

	let term = $state('');

	let filtered = $derived(
		customers.filter((customer) => {
			if (!term.trim()) return true;
			const q = term.trim().toLowerCase();

			return (
				customer.name.toLowerCase().includes(q) ||
				(customer.email ?? '').toLowerCase().includes(q)
			);
		})
	);

	let selectedIds = $derived(new Set(selected));

	function toggle(id: number) {
		const next = selected.includes(id)
			? selected.filter((value) => value !== id)
			: [...selected, id];
		selected = next;
		onSelectedChange?.(next);
	}

	function clearSelection() {
		selected = [];
		onSelectedChange?.([]);
	}
</script>

<div class="space-y-2">
	<div class="flex items-center gap-2">
		<div class="relative flex-1">
			<svg
				class="absolute left-2.5 top-2 h-3.5 w-3.5 text-gray-400"
				fill="none"
				viewBox="0 0 24 24"
				stroke-width="2"
				stroke="currentColor"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z"
				/>
			</svg>
			<input
				type="search"
				bind:value={term}
				placeholder="Search customers..."
				class="w-full rounded-lg border border-gray-200 bg-gray-50/60 py-1.5 pl-8 pr-3 text-xs outline-none focus:border-emerald-500 focus:bg-white dark:border-gray-700 dark:bg-gray-800 dark:text-white transition"
			/>
		</div>
		<span
			class="shrink-0 rounded-full border border-gray-200 bg-gray-50 px-2 py-1 text-[10px] font-bold text-gray-600 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-300"
		>
			{selected.length} selected
		</span>
	</div>

	<div
		class="max-h-52 overflow-y-auto rounded-lg border border-gray-200 divide-y divide-gray-100 dark:border-gray-700 dark:divide-gray-800"
	>
		{#if filtered.length === 0}
			<div class="px-3 py-6 text-center text-[11px] text-gray-400">
				{customers.length === 0 ? emptyText : 'No customer matches this search.'}
			</div>
		{:else}
			{#each filtered as customer}
				<label
					class="flex cursor-pointer items-center gap-2.5 px-3 py-2 text-xs transition-colors hover:bg-gray-50 dark:hover:bg-gray-800/50 {selectedIds.has(customer.id)
						? 'bg-emerald-50/60 dark:bg-emerald-950/20'
						: ''}"
				>
					<input
						type="checkbox"
						checked={selectedIds.has(customer.id)}
						onchange={() => toggle(customer.id)}
						class="h-3.5 w-3.5 shrink-0 rounded border-gray-300 text-emerald-600 focus:ring-emerald-500"
					/>
					<span class="min-w-0 flex-1">
						<span class="block truncate font-semibold text-gray-800 dark:text-gray-200">{customer.name}</span>
						{#if customer.email}
							<span class="block truncate text-[10px] text-gray-400">{customer.email}</span>
						{/if}
					</span>
				</label>
			{/each}
		{/if}
	</div>

	{#if selected.length > 0}
		<div class="flex flex-wrap items-center gap-1.5">
			{#each selected as id}
				{@const customer = customers.find((value) => value.id === id)}
				<button
					type="button"
					onclick={() => toggle(id)}
					class="inline-flex items-center gap-1 rounded-full border border-emerald-200 bg-emerald-50 px-2 py-0.5 text-[10px] font-semibold text-emerald-700 transition hover:bg-emerald-100 dark:border-emerald-800 dark:bg-emerald-950/40 dark:text-emerald-300"
				>
					{customer?.name ?? `#${id}`}
					<span aria-hidden="true">×</span>
				</button>
			{/each}
			<button
				type="button"
				onclick={clearSelection}
				class="text-[10px] font-semibold text-gray-400 hover:text-rose-600 transition"
			>
				Clear all
			</button>
		</div>
	{/if}
</div>
