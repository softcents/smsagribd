<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface Invoice {
		id: string;
		deposit_id: number;
		customer_name: string;
		customer_email: string;
		customer_phone: string;
		amount: number;
		currency: string;
		payment_method: string;
		status: string;
		issued_date: string;
		paid_at: string | null;
	}

	let invoices = $state<Invoice[]>([]);
	let loading = $state(true);
	let search = $state('');
	let statusFilter = $state('all');

	// Pagination
	let currentPage = $state(1);
	let perPage = $state(25);
	let totalRecords = $state(0);
	let totalPages = $derived(Math.ceil(totalRecords / perPage) || 1);

	// Preview Invoice modal
	let previewInvoice = $state<Invoice | null>(null);

	async function loadInvoices() {
		loading = true;
		try {
			const query = new URLSearchParams({
				page: String(currentPage),
				per_page: String(perPage)
			});
			if (search) query.set('search', search);
			if (statusFilter !== 'all') query.set('status', statusFilter);

			const res = await supermanApi.get<{
				data: Invoice[];
				meta: { page: number; per_page: number; total: number };
			}>(`/api/v1/superman/billing/invoices?${query.toString()}`);

			invoices = res.data || [];
			totalRecords = res.meta?.total || 0;
		} catch (err) {
			console.error('Failed to load invoices:', err);
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		loadInvoices();
	});

	function formatCurrency(val: number): string {
		return '৳' + (val || 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
	}

	function printInvoice() {
		window.print();
	}
	const exportColumns: ExportColumn<Invoice>[] = [
		{ header: 'Invoice ID', getValue: (i) => i.id },
		{ header: 'Customer', getValue: (i) => i.customer_name },
		{ header: 'Phone', getValue: (i) => i.customer_phone },
		{ header: 'Amount', getValue: (i) => `BDT ${i.amount.toFixed(2)}` },
		{ header: 'Payment Method', getValue: (i) => i.payment_method },
		{ header: 'Status', getValue: (i) => i.status },
		{ header: 'Issued Date', getValue: (i) => i.issued_date }
	];
</script>

<svelte:head>
	<title>Customer Invoices — SMS Whiteinfo Administrator</title>
</svelte:head>

<div class="space-y-6">
	<!-- Page Header -->
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<div class="flex items-center gap-2">
				<a href="/superman/billing" class="text-xs font-semibold text-gray-500 hover:text-emerald-600 dark:hover:text-emerald-400">
					Billing
				</a>
				<span class="text-xs text-gray-400">/</span>
				<span class="text-xs font-semibold text-gray-900 dark:text-white">Invoices</span>
			</div>
			<h1 class="mt-1 text-xl font-bold tracking-tight text-gray-900 dark:text-white sm:text-2xl">
				Invoices & Billing Receipts
			</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400">
				Official tax receipts, top-up invoices, customer statements, and printable payment documents.
			</p>
		</div>

		<div class="flex items-center gap-2">
			<ExportDropdown
				title="Customer Invoices"
				columns={exportColumns}
				getData={() => invoices}
				disabled={invoices.length === 0}
			/>
			<a
				href="/superman/billing"
				class="inline-flex items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs font-semibold text-gray-700 transition hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300 dark:hover:bg-gray-800"
			>
				<svg class="h-4 w-4 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 17l-5-5m0 0l5-5m-5 5h12" />
				</svg>
				Billing Dashboard
			</a>
			<a
				href="/superman/billing/payments"
				class="inline-flex items-center gap-1.5 rounded-lg bg-emerald-600 px-3.5 py-2 text-xs font-semibold text-white shadow-xs hover:bg-emerald-500"
			>
				<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z" />
				</svg>
				Payment Gateways
			</a>
		</div>
	</div>

	<!-- Toolbar & Filters -->
	<div class="flex flex-col gap-3 rounded-xl border border-gray-200 bg-white p-3 shadow-xs dark:border-gray-800 dark:bg-gray-900 sm:flex-row sm:items-center sm:justify-between">
		<div class="relative flex-1 max-w-sm">
			<svg class="pointer-events-none absolute left-3 top-2.5 h-4 w-4 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
			</svg>
			<input
				type="text"
				bind:value={search}
				onkeydown={(e) => {
					if (e.key === 'Enter') {
						currentPage = 1;
						loadInvoices();
					}
				}}
				placeholder="Search customer, invoice ID… (Enter)"
				class="w-full rounded-lg border border-gray-200 bg-gray-50 py-1.5 pl-9 pr-3 text-xs text-gray-900 focus:border-emerald-500 focus:bg-white focus:outline-hidden dark:border-gray-800 dark:bg-gray-800 dark:text-white"
			/>
		</div>

		<div class="flex items-center gap-2">
			<select
				bind:value={statusFilter}
				class="rounded-lg border border-gray-200 bg-white px-2.5 py-1.5 text-xs text-gray-700 focus:border-emerald-500 focus:outline-hidden dark:border-gray-800 dark:bg-gray-800 dark:text-gray-300"
			>
				<option value="all">All Invoices</option>
				<option value="completed">Paid / Completed</option>
				<option value="pending">Pending</option>
				<option value="failed">Failed / Cancelled</option>
			</select>

			<button
				type="button"
				onclick={() => {
					currentPage = 1;
					loadInvoices();
				}}
				class="rounded-lg bg-gray-900 px-3 py-1.5 text-xs font-semibold text-white hover:bg-gray-800 dark:bg-white dark:text-gray-900 dark:hover:bg-gray-100"
			>
				Filter
			</button>
		</div>
	</div>

	<!-- Invoices Table -->
	<div class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-xs dark:border-gray-800 dark:bg-gray-900">
		<div class="overflow-x-auto">
			<table class="w-full text-left text-xs">
				<thead>
					<tr class="border-b border-gray-100 bg-gray-50/50 text-[10px] font-semibold uppercase tracking-wider text-gray-400 dark:border-gray-800 dark:bg-gray-800/40">
						<th class="py-3 px-4">Invoice #</th>
						<th class="py-3 px-3">Issued Date</th>
						<th class="py-3 px-3">Customer</th>
						<th class="py-3 px-3 text-right">Total Amount</th>
						<th class="py-3 px-3">Payment Channel</th>
						<th class="py-3 px-3">Status</th>
						<th class="py-3 px-3">Paid Date</th>
						<th class="py-3 px-4 text-right">Action</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
					{#each invoices as inv}
						<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30 transition-colors">
							<td class="py-3 px-4 font-mono font-bold text-gray-900 dark:text-white">
								{inv.id}
							</td>
							<td class="py-3 px-3 font-mono text-[11px] text-gray-500">
								{inv.issued_date}
							</td>
							<td class="py-3 px-3">
								<p class="font-bold text-gray-900 dark:text-white">{inv.customer_name}</p>
								<p class="text-[10px] text-gray-400">{inv.customer_email}</p>
							</td>
							<td class="py-3 px-3 text-right font-mono font-bold text-emerald-600 dark:text-emerald-400">
								{formatCurrency(inv.amount)}
							</td>
							<td class="py-3 px-3 capitalize text-gray-700 dark:text-gray-300">
								{inv.payment_method}
							</td>
							<td class="py-3 px-3">
								<span class="inline-flex rounded-full px-2 py-0.5 text-[10px] font-semibold {inv.status === 'completed' ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300' : inv.status === 'pending' ? 'bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-300' : 'bg-rose-100 text-rose-800 dark:bg-rose-950 dark:text-rose-300'}">
									{inv.status === 'completed' ? 'Paid' : inv.status}
								</span>
							</td>
							<td class="py-3 px-3 font-mono text-[11px] text-gray-400">
								{inv.paid_at || '—'}
							</td>
							<td class="py-3 px-4 text-right">
								<button
									type="button"
									onclick={() => (previewInvoice = inv)}
									class="inline-flex items-center gap-1 rounded-lg border border-gray-200 bg-white px-2.5 py-1 text-xs font-semibold text-gray-700 shadow-2xs hover:bg-gray-50 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-200"
								>
									<svg class="h-3.5 w-3.5 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
									</svg>
									View Invoice
								</button>
							</td>
						</tr>
					{:else}
						<tr>
							<td colspan="8" class="py-8 text-center text-xs text-gray-400">
								{loading ? 'Loading invoices…' : 'No invoices found.'}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>

		<!-- Pagination -->
		{#if totalPages > 1}
			<div class="flex items-center justify-between border-t border-gray-100 px-4 py-3 dark:border-gray-800">
				<p class="text-xs text-gray-500">
					Page <strong class="font-semibold text-gray-900 dark:text-white">{currentPage}</strong> of <strong class="font-semibold text-gray-900 dark:text-white">{totalPages}</strong>
				</p>
				<div class="flex items-center gap-1.5">
					<button
						type="button"
						onclick={() => {
							if (currentPage > 1) {
								currentPage--;
								loadInvoices();
							}
						}}
						disabled={currentPage <= 1}
						class="rounded-md border border-gray-200 px-2.5 py-1 text-xs font-semibold text-gray-700 hover:bg-gray-50 disabled:opacity-40 dark:border-gray-700 dark:text-gray-300 dark:hover:bg-gray-800"
					>
						Previous
					</button>
					<button
						type="button"
						onclick={() => {
							if (currentPage < totalPages) {
								currentPage++;
								loadInvoices();
							}
						}}
						disabled={currentPage >= totalPages}
						class="rounded-md border border-gray-200 px-2.5 py-1 text-xs font-semibold text-gray-700 hover:bg-gray-50 disabled:opacity-40 dark:border-gray-700 dark:text-gray-300 dark:hover:bg-gray-800"
					>
						Next
					</button>
				</div>
			</div>
		{/if}
	</div>
</div>

<!-- Modal: Printable Tax Invoice -->
{#if previewInvoice}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-gray-900/60 backdrop-blur-xs p-4">
		<div class="w-full max-w-2xl max-h-[90vh] flex flex-col rounded-2xl border border-gray-200 bg-white shadow-2xl dark:border-gray-800 dark:bg-gray-900 animate-in fade-in zoom-in-95 duration-150">
			<!-- Header -->
			<div class="flex items-center justify-between border-b border-gray-100 p-4 dark:border-gray-800 print:hidden">
				<div class="flex items-center gap-2">
					<span class="rounded-md bg-emerald-50 px-2 py-0.5 text-xs font-bold text-emerald-700 dark:bg-emerald-950 dark:text-emerald-300">
						Tax Receipt
					</span>
					<span class="font-mono text-xs font-semibold text-gray-600 dark:text-gray-300">{previewInvoice.id}</span>
				</div>
				<div class="flex items-center gap-2">
					<button
						type="button"
						onclick={printInvoice}
						class="inline-flex items-center gap-1.5 rounded-lg bg-gray-900 px-3 py-1.5 text-xs font-semibold text-white hover:bg-gray-800 dark:bg-white dark:text-gray-900"
					>
						<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z" />
						</svg>
						Print
					</button>
					<button
						type="button"
						onclick={() => (previewInvoice = null)}
						class="rounded-lg p-1.5 text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800"
					>
						<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
						</svg>
					</button>
				</div>
			</div>

			<!-- Printable Body -->
			<div class="overflow-y-auto p-6 space-y-6 text-gray-900 dark:text-white" id="invoice-printable">
				<!-- Invoice Top Branding -->
				<div class="flex items-start justify-between border-b border-gray-100 pb-5 dark:border-gray-800">
					<div>
						<img src="/logo.png?v=aggregator_v2" alt="SMS Whiteinfo" class="h-8 w-auto object-contain" />
						<p class="mt-2 text-xs font-semibold text-gray-700 dark:text-gray-300">Whiteinfo Technologies BD</p>
						<p class="text-[11px] text-gray-400">Enterprise Bulk SMS & Telecommunication Services</p>
						<p class="text-[11px] text-gray-400">Dhaka, Bangladesh · support@whiteinfobd.com</p>
					</div>
					<div class="text-right">
						<h2 class="text-lg font-black tracking-tight text-gray-900 dark:text-white uppercase">Invoice</h2>
						<p class="font-mono text-xs font-bold text-emerald-600 dark:text-emerald-400">{previewInvoice.id}</p>
						<p class="mt-1 text-xs text-gray-400">Date: {previewInvoice.issued_date}</p>
						<p class="text-xs text-gray-400">Status: <strong class="text-emerald-600 capitalize">{previewInvoice.status}</strong></p>
					</div>
				</div>

				<!-- Bill To -->
				<div class="grid grid-cols-2 gap-4 text-xs">
					<div>
						<p class="text-[10px] font-bold uppercase tracking-wider text-gray-400">Bill To:</p>
						<p class="mt-1 font-bold text-gray-900 dark:text-white">{previewInvoice.customer_name}</p>
						<p class="text-gray-500">{previewInvoice.customer_email}</p>
						<p class="font-mono text-gray-500">{previewInvoice.customer_phone || 'N/A'}</p>
					</div>
					<div class="text-right">
						<p class="text-[10px] font-bold uppercase tracking-wider text-gray-400">Payment Channel:</p>
						<p class="mt-1 font-semibold capitalize text-gray-800 dark:text-gray-200">{previewInvoice.payment_method}</p>
						{#if previewInvoice.paid_at}
							<p class="text-gray-400 text-[11px]">Paid At: {previewInvoice.paid_at}</p>
						{/if}
					</div>
				</div>

				<!-- Items Table -->
				<div class="rounded-lg border border-gray-200 overflow-hidden dark:border-gray-800">
					<table class="w-full text-left text-xs">
						<thead>
							<tr class="border-b border-gray-200 bg-gray-50 text-[10px] font-semibold uppercase text-gray-500 dark:border-gray-800 dark:bg-gray-800">
								<th class="py-2.5 px-3">Description</th>
								<th class="py-2.5 px-3 text-center">Qty</th>
								<th class="py-2.5 px-3 text-right">Unit Price</th>
								<th class="py-2.5 px-3 text-right">Total</th>
							</tr>
						</thead>
						<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
							<tr>
								<td class="py-3 px-3">
									<p class="font-semibold text-gray-900 dark:text-white">Customer Account Balance Top-up</p>
									<p class="text-[10px] text-gray-400">Deposit reference #{previewInvoice.deposit_id} ({previewInvoice.payment_method})</p>
								</td>
								<td class="py-3 px-3 text-center font-mono">1</td>
								<td class="py-3 px-3 text-right font-mono">{formatCurrency(previewInvoice.amount)}</td>
								<td class="py-3 px-3 text-right font-mono font-bold text-gray-900 dark:text-white">
									{formatCurrency(previewInvoice.amount)}
								</td>
							</tr>
						</tbody>
					</table>
				</div>

				<!-- Totals Calculation -->
				<div class="flex justify-end text-xs">
					<div class="w-64 space-y-1.5">
						<div class="flex justify-between py-1 text-gray-500">
							<span>Subtotal:</span>
							<span class="font-mono text-gray-800 dark:text-gray-200">{formatCurrency(previewInvoice.amount)}</span>
						</div>
						<div class="flex justify-between py-1 text-gray-500">
							<span>VAT / Tax (0%):</span>
							<span class="font-mono text-gray-800 dark:text-gray-200">৳0.00</span>
						</div>
						<div class="flex justify-between border-t border-gray-200 pt-2 font-bold text-gray-900 dark:border-gray-800 dark:text-white text-sm">
							<span>Total Paid:</span>
							<span class="font-mono text-emerald-600 dark:text-emerald-400">{formatCurrency(previewInvoice.amount)}</span>
						</div>
					</div>
				</div>

				<!-- Note footer -->
				<div class="border-t border-gray-100 pt-4 text-[11px] text-gray-400 text-center dark:border-gray-800">
					This is a computer-generated tax invoice for electronic SMS service delivery. No signature required.
				</div>
			</div>
		</div>
	</div>
{/if}
