<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface InvoiceItem {
		id: number;
		invoice_number: string;
		customer_name: string;
		amount: number;
		status: string;
		created_at: string;
	}

	let invoices = $state<InvoiceItem[]>([]);
	let loading = $state(true);

	const exportColumns: ExportColumn<InvoiceItem>[] = [
		{ header: 'ID', getValue: (i) => i.id },
		{ header: 'Invoice Number', getValue: (i) => i.invoice_number },
		{ header: 'Reseller', getValue: (i) => i.customer_name },
		{ header: 'Amount (BDT)', getValue: (i) => `BDT ${i.amount.toFixed(2)}` },
		{ header: 'Status', getValue: (i) => i.status },
		{ header: 'Date', getValue: (i) => i.created_at }
	];

	onMount(async () => {
		try {
			const res = await supermanApi.get<{ data: InvoiceItem[] }>('/api/v1/superman/sps/billing');
			invoices = res.data ?? [];
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	});
</script>

<div class="space-y-6">
	<!-- Page Header -->
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-xl font-bold text-gray-900 dark:text-white">Reseller Billing & Wholesale Invoices</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
				Monthly billing statements, wholesale balance settlements, and tax invoice generation for resellers.
			</p>
		</div>
		<div class="flex items-center gap-3">
			<ExportDropdown
				title="Reseller Wholesale Invoices"
				columns={exportColumns}
				getData={() => invoices}
				disabled={invoices.length === 0}
			/>
			<a
				href="/superman/billing/invoices"
				class="inline-flex items-center gap-2 rounded-lg bg-emerald-600 px-3 py-1.5 text-xs font-semibold text-white shadow-xs hover:bg-emerald-700 transition"
			>
				Global Invoice Directory
			</a>
		</div>
	</div>

	<!-- Invoices Table -->
	<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl shadow-xs overflow-hidden">
		<div class="overflow-x-auto">
			<table class="w-full text-left text-xs">
				<thead class="bg-gray-50/75 dark:bg-gray-800/50 text-gray-500 uppercase tracking-wider font-semibold border-b border-gray-200 dark:border-gray-800">
					<tr>
						<th class="px-5 py-3">Invoice #</th>
						<th class="px-5 py-3">Partner / Reseller</th>
						<th class="px-5 py-3">Amount (BDT)</th>
						<th class="px-5 py-3">Status</th>
						<th class="px-5 py-3">Issue Date</th>
						<th class="px-5 py-3 text-right">Actions</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800 text-gray-700 dark:text-gray-300">
					{#if loading}
						<tr>
							<td colspan="6" class="px-5 py-8 text-center text-gray-400">Loading invoices...</td>
						</tr>
					{:else if invoices.length === 0}
						<tr>
							<td colspan="6" class="px-5 py-8 text-center text-gray-400">No reseller invoices found.</td>
						</tr>
					{:else}
						{#each invoices as inv}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30 transition">
								<td class="px-5 py-3 font-mono font-bold text-gray-900 dark:text-white">{inv.invoice_number}</td>
								<td class="px-5 py-3 font-semibold text-gray-800 dark:text-gray-200">{inv.customer_name}</td>
								<td class="px-5 py-3 font-bold text-gray-900 dark:text-white">৳{inv.amount.toFixed(2)}</td>
								<td class="px-5 py-3">
									<span class="inline-flex items-center px-2 py-0.5 rounded text-[11px] font-medium {inv.status === 'paid' ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300' : 'bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-300'}">
										{inv.status}
									</span>
								</td>
								<td class="px-5 py-3 text-gray-400">{new Date(inv.created_at).toLocaleDateString()}</td>
								<td class="px-5 py-3 text-right">
									<button
										type="button"
										class="text-xs text-emerald-600 hover:text-emerald-700 font-semibold"
										onclick={() => alert(`Viewing PDF invoice for ${inv.invoice_number}`)}
									>
										View PDF
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

