<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface ResellerWallet {
		id: number;
		name: string;
		business_name: string;
		phone: string;
		email: string;
		balance: number;
		account_status: string;
		sub_clients_count: number;
	}

	let wallets = $state<ResellerWallet[]>([]);
	let loading = $state(true);

	const exportColumns: ExportColumn<ResellerWallet>[] = [
		{ header: 'ID', getValue: (w) => w.id },
		{ header: 'Reseller', getValue: (w) => w.name },
		{ header: 'Business Name', getValue: (w) => w.business_name },
		{ header: 'Phone', getValue: (w) => w.phone },
		{ header: 'Email', getValue: (w) => w.email },
		{ header: 'Wallet Balance (BDT)', getValue: (w) => `BDT ${w.balance.toFixed(2)}` },
		{ header: 'Account Status', getValue: (w) => w.account_status },
		{ header: 'Sub-Clients Count', getValue: (w) => w.sub_clients_count }
	];

	// Topup Modal State
	let selectedSp = $state<ResellerWallet | null>(null);
	let topupAmount = $state<number>(1000);
	let topupNote = $state<string>('');
	let submitting = $state(false);
	let successMsg = $state<string | null>(null);

	async function loadWallets() {
		loading = true;
		try {
			const res = await supermanApi.get<{ data: ResellerWallet[] }>('/api/v1/superman/sps/wallet');
			wallets = res.data ?? [];
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	}

	onMount(loadWallets);

	function openTopupModal(sp: ResellerWallet) {
		selectedSp = sp;
		topupAmount = 1000;
		topupNote = 'Admin wholesale credit';
	}

	async function submitTopup() {
		if (!selectedSp || topupAmount <= 0) return;
		submitting = true;
		successMsg = null;
		try {
			await supermanApi.post('/api/v1/superman/sps/wallet/topup', {
				sp_id: selectedSp.id,
				amount: topupAmount,
				note: topupNote
			});
			successMsg = `Successfully credited ৳${topupAmount.toFixed(2)} to ${selectedSp.name}.`;
			selectedSp = null;
			await loadWallets();
		} catch (e: any) {
			alert(e?.message || 'Failed to top-up wallet.');
		} finally {
			submitting = false;
		}
	}
</script>

<div class="space-y-6">
	<!-- Page Header -->
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-xl font-bold text-gray-900 dark:text-white">Reseller Balances & Wholesale Top-Ups</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
				Real-time reseller prepaid wallet balances, wholesale credit allocation, and ledger management.
			</p>
		</div>
		<ExportDropdown
			title="Reseller Wallets and Balances"
			columns={exportColumns}
			getData={() => wallets}
			disabled={wallets.length === 0}
		/>
	</div>

	{#if successMsg}
		<div class="rounded-lg bg-emerald-50 dark:bg-emerald-950/40 border border-emerald-200 dark:border-emerald-800 p-4 text-xs font-medium text-emerald-800 dark:text-emerald-300">
			{successMsg}
		</div>
	{/if}

	<!-- Wallets Table -->
	<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl shadow-xs overflow-hidden">
		<div class="overflow-x-auto">
			<table class="w-full text-left text-xs">
				<thead class="bg-gray-50/75 dark:bg-gray-800/50 text-gray-500 uppercase tracking-wider font-semibold border-b border-gray-200 dark:border-gray-800">
					<tr>
						<th class="px-5 py-3">Reseller</th>
						<th class="px-5 py-3">Organization</th>
						<th class="px-5 py-3">Phone</th>
						<th class="px-5 py-3">Sub-Clients</th>
						<th class="px-5 py-3">Balance (BDT)</th>
						<th class="px-5 py-3">Status</th>
						<th class="px-5 py-3 text-right">Actions</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800 text-gray-700 dark:text-gray-300">
					{#if loading}
						<tr>
							<td colspan="7" class="px-5 py-8 text-center text-gray-400">Loading reseller wallets...</td>
						</tr>
					{:else if wallets.length === 0}
						<tr>
							<td colspan="7" class="px-5 py-8 text-center text-gray-400">No active SP resellers found.</td>
						</tr>
					{:else}
						{#each wallets as w}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30 transition">
								<td class="px-5 py-3">
									<a href="/superman/users/sps/{w.id}" class="font-bold text-gray-900 dark:text-white hover:text-emerald-600 transition">
										{w.name}
									</a>
									<div class="text-[11px] text-gray-400">{w.email}</div>
								</td>
								<td class="px-5 py-3 text-gray-600 dark:text-gray-400">{w.business_name || '—'}</td>
								<td class="px-5 py-3 font-mono">{w.phone}</td>
								<td class="px-5 py-3">
									<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200">
										{w.sub_clients_count} clients
									</span>
								</td>
								<td class="px-5 py-3 font-bold text-emerald-600 dark:text-emerald-400">৳{w.balance.toFixed(2)}</td>
								<td class="px-5 py-3">
									<span class="inline-flex items-center px-2 py-0.5 rounded text-[11px] font-medium {w.account_status === 'approved' ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300' : 'bg-red-100 text-red-800 dark:bg-red-950 dark:text-red-300'}">
										{w.account_status}
									</span>
								</td>
								<td class="px-5 py-3 text-right">
									<button
										type="button"
										onclick={() => openTopupModal(w)}
										class="inline-flex items-center gap-1.5 rounded-md bg-emerald-600 px-2.5 py-1 text-xs font-semibold text-white shadow-xs hover:bg-emerald-700 transition"
									>
										+ Credit Wallet
									</button>
								</td>
							</tr>
						{/each}
					{/if}
				</tbody>
			</table>
		</div>
	</div>

	<!-- Top-up Modal -->
	{#if selectedSp}
		<div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4 backdrop-blur-xs">
			<div class="w-full max-w-md rounded-xl bg-white dark:bg-gray-900 p-6 shadow-xl border border-gray-200 dark:border-gray-800 space-y-4">
				<div class="flex items-center justify-between border-b border-gray-100 dark:border-gray-800 pb-3">
					<h3 class="text-sm font-bold text-gray-900 dark:text-white">Credit Reseller Wallet</h3>
					<button type="button" onclick={() => selectedSp = null} class="text-gray-400 hover:text-gray-600">✕</button>
				</div>

				<div class="space-y-3">
					<div class="p-3 rounded-lg bg-gray-50 dark:bg-gray-800/50 text-xs text-gray-600 dark:text-gray-400">
						Crediting wallet for <strong class="text-gray-900 dark:text-white">{selectedSp.name}</strong> ({selectedSp.phone}).
						Current Balance: <strong class="text-emerald-600">৳{selectedSp.balance.toFixed(2)}</strong>.
					</div>

					<div>
						<label for="topup_amt" class="block text-xs font-semibold text-gray-700 dark:text-gray-300 mb-1">Top-Up Amount (BDT)</label>
						<input
							id="topup_amt"
							type="number"
							bind:value={topupAmount}
							min="1"
							class="w-full rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 px-3 py-2 text-xs text-gray-900 dark:text-white focus:outline-none focus:ring-1 focus:ring-emerald-500 font-bold"
						/>
					</div>

					<div>
						<label for="topup_notes" class="block text-xs font-semibold text-gray-700 dark:text-gray-300 mb-1">Administrative Note / Reference</label>
						<input
							id="topup_notes"
							type="text"
							bind:value={topupNote}
							class="w-full rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 px-3 py-2 text-xs text-gray-900 dark:text-white focus:outline-none focus:ring-1 focus:ring-emerald-500"
						/>
					</div>
				</div>

				<div class="flex items-center justify-end gap-3 pt-3 border-t border-gray-100 dark:border-gray-800">
					<button
						type="button"
						onclick={() => selectedSp = null}
						class="rounded-lg border border-gray-300 dark:border-gray-700 px-3 py-1.5 text-xs font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-50 transition"
					>
						Cancel
					</button>
					<button
						type="button"
						onclick={submitTopup}
						disabled={submitting}
						class="rounded-lg bg-emerald-600 px-3 py-1.5 text-xs font-semibold text-white hover:bg-emerald-700 transition disabled:opacity-50"
					>
						{submitting ? 'Crediting...' : 'Confirm Credit'}
					</button>
				</div>
			</div>
		</div>
	{/if}
</div>

