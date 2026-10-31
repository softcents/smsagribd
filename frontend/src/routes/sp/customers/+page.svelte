<script lang="ts">
	import { spAuth } from '$lib/stores/spAuth.svelte.js';
	import { spApi, SpApiError } from '$lib/sp-api.js';
	import { onMount } from 'svelte';

	interface Customer {
		id: number;
		name: string;
		email: string;
		phone: string;
		balance: number;
		status: string;
		kyc_status?: string;
		created_at?: string;
	}

	let searchQuery = $state('');
	let distributeModalOpen = $state(false);
	let selectedCustomer = $state<Customer | null>(null);
	let transferAmount = $state('');
	let transferNotes = $state('');
	let successMsg = $state<string | null>(null);
	let modalError = $state<string | null>(null);
	let loading = $state(true);
	let transferring = $state(false);

	let customers = $state<Customer[]>([]);

	async function loadCustomers() {
		loading = true;
		try {
			const query = searchQuery ? `?search=${encodeURIComponent(searchQuery)}` : '';
			const res = await spApi.get<{ data: Customer[] }>(`/api/v1/sp/customers${query}`);
			customers = res.data ?? [];
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		loadCustomers();
	});

	function openDistributeModal(customer: Customer) {
		selectedCustomer = customer;
		transferAmount = '';
		transferNotes = '';
		modalError = null;
		distributeModalOpen = true;
	}

	async function handleTransfer(e: SubmitEvent) {
		e.preventDefault();
		if (!selectedCustomer) return;

		const amount = parseFloat(transferAmount);
		if (isNaN(amount) || amount <= 0) {
			modalError = 'Please enter a valid positive amount.';
			return;
		}

		if (amount > spAuth.balance) {
			modalError = 'Insufficient S P Reseller balance.';
			return;
		}

		transferring = true;
		modalError = null;
		try {
			const res = await spApi.post<{ message: string; data: { sp_balance: number; customer_balance: number } }>(
				`/api/v1/sp/customers/${selectedCustomer.id}/fund`,
				{
					amount,
					description: transferNotes || undefined
				}
			);

			if (res?.data) {
				spAuth.updateBalance(res.data.sp_balance);
				selectedCustomer.balance = res.data.customer_balance;
			}

			successMsg = res.message ?? `Successfully transferred ৳${amount.toFixed(2)} to ${selectedCustomer.name}.`;
			distributeModalOpen = false;
			await loadCustomers();
			setTimeout(() => (successMsg = null), 5000);
		} catch (err) {
			if (err instanceof SpApiError) {
				modalError = (err.data.message as string) ?? 'Transfer failed.';
			} else {
				modalError = 'An unexpected error occurred during transfer.';
			}
		} finally {
			transferring = false;
		}
	}
</script>

<svelte:head>
	<title>Customers — SMS Whiteinfo S P Reseller</title>
</svelte:head>

<div class="space-y-4">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">Assigned Customers</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400">Manage client accounts and distribute balance directly from your wallet.</p>
		</div>
	</div>

	{#if successMsg}
		<div class="rounded-xl border border-emerald-200 bg-emerald-50/80 p-3 text-xs text-emerald-800 dark:border-emerald-800 dark:bg-emerald-950/40 dark:text-emerald-300">
			{successMsg}
		</div>
	{/if}

	<div class="rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900 space-y-4">
		<div class="overflow-x-auto">
			<table class="w-full text-left text-xs">
				<thead>
					<tr class="border-b border-gray-100 dark:border-gray-800 text-gray-400 uppercase text-[10px] font-semibold">
						<th class="py-2.5">Customer</th>
						<th class="py-2.5">Phone</th>
						<th class="py-2.5">Current Balance</th>
						<th class="py-2.5">Status</th>
						<th class="py-2.5 text-right">Actions</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
					{#if loading}
						<tr>
							<td colspan="5" class="py-8 text-center text-gray-400">
								<div class="inline-block h-5 w-5 animate-spin rounded-full border-2 border-amber-500 border-t-transparent"></div>
								<div class="mt-1 text-xs">Loading assigned customers…</div>
							</td>
						</tr>
					{:else if customers.length === 0}
						<tr>
							<td colspan="5" class="py-8 text-center text-gray-400">
								No assigned customers found.
							</td>
						</tr>
					{:else}
						{#each customers as c}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30 transition-colors">
								<td class="py-3 font-semibold text-gray-900 dark:text-white">
									<div>{c.name}</div>
									<div class="text-[11px] text-gray-400 font-normal">{c.email}</div>
								</td>
								<td class="py-3 font-mono text-gray-700 dark:text-gray-300">{c.phone || '—'}</td>
								<td class="py-3 font-mono font-bold text-amber-600 dark:text-amber-400">৳{c.balance.toFixed(2)}</td>
								<td class="py-3">
									<span class="inline-block rounded-full bg-emerald-100 px-2 py-0.5 text-[10px] font-semibold text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300">
										{c.status}
									</span>
								</td>
								<td class="py-3 text-right">
									<button
										type="button"
										onclick={() => openDistributeModal(c)}
										class="rounded-lg bg-amber-500 px-3 py-1 text-xs font-semibold text-white shadow-2xs hover:bg-amber-600 transition cursor-pointer"
									>
										Fund Balance
									</button>
								</td>
							</tr>
						{/each}
					{/if}
				</tbody>
			</table>
		</div>
	</div>
</div>

<!-- Fund Balance Modal -->
{#if distributeModalOpen && selectedCustomer}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-gray-950/50 backdrop-blur-xs p-4">
		<div class="w-full max-w-md rounded-2xl border border-gray-200 bg-white p-6 shadow-2xl dark:border-gray-800 dark:bg-gray-900 space-y-4">
			<div class="flex items-center justify-between border-b border-gray-100 pb-3 dark:border-gray-800">
				<h3 class="text-sm font-bold text-gray-900 dark:text-white">Fund Customer Balance</h3>
				<button onclick={() => (distributeModalOpen = false)} class="text-gray-400 hover:text-gray-600 cursor-pointer">✕</button>
			</div>

			{#if modalError}
				<div class="rounded-xl border border-rose-200 bg-rose-50/80 p-3 text-xs text-rose-700 dark:border-rose-900/50 dark:bg-rose-950/40 dark:text-rose-300">
					{modalError}
				</div>
			{/if}

			<form onsubmit={handleTransfer} class="space-y-4">
				<div>
					<span class="text-xs text-gray-500">Customer:</span>
					<div class="text-sm font-bold text-gray-900 dark:text-white">{selectedCustomer.name} ({selectedCustomer.email})</div>
				</div>

				<div>
					<label for="amount" class="block text-xs font-medium text-gray-700 dark:text-gray-300">Amount (BDT)</label>
					<input
						id="amount"
						type="number"
						bind:value={transferAmount}
						step="0.01"
						required
						placeholder="e.g. 500"
						class="mt-1 block w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-xs text-gray-900 outline-none focus:border-amber-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
					/>
				</div>

				<div>
					<label for="notes" class="block text-xs font-medium text-gray-700 dark:text-gray-300">Notes (optional)</label>
					<input
						id="notes"
						type="text"
						bind:value={transferNotes}
						placeholder="Reference or receipt number"
						class="mt-1 block w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-xs text-gray-900 outline-none focus:border-amber-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
					/>
				</div>

				<div class="pt-2 flex items-center justify-end gap-2">
					<button
						type="button"
						onclick={() => (distributeModalOpen = false)}
						class="rounded-lg border border-gray-200 px-3 py-1.5 text-xs font-semibold text-gray-600 hover:bg-gray-50 dark:border-gray-700 dark:text-gray-300 cursor-pointer"
					>
						Cancel
					</button>
					<button
						type="submit"
						disabled={transferring}
						class="rounded-lg bg-amber-500 px-4 py-1.5 text-xs font-semibold text-white shadow-2xs hover:bg-amber-600 disabled:opacity-50 cursor-pointer flex items-center gap-1.5"
					>
						{#if transferring}
							<div class="h-3 w-3 animate-spin rounded-full border-2 border-white border-t-transparent"></div>
							<span>Transferring…</span>
						{:else}
							<span>Confirm Transfer</span>
						{/if}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
