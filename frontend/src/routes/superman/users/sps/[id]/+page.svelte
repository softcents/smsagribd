<script lang="ts">
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	const spId = page.params.id;

	interface SPReseller {
		id: number;
		name: string;
		email: string;
		phone: string;
		business_name: string | null;
		balance: number;
		status: string;
		wholesale_plan?: string | null;
		customer_count: number;
		created_at: string;
	}

	interface SPDetail {
		sp: SPReseller;
		api_access?: boolean;
		sms_auto_approved?: boolean;
		wholesale_plan_id?: number | null;
		customers?: Array<{ id: number; name: string; email: string; phone: string; balance: number; status: string; created_at: string }>;
		recent_transactions?: Array<{ id: number; amount: number; direction: string; balance_before: number; balance_after: number; description: string | null; created_at: string }>;
	}

	let detail = $state<SPDetail | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);

	type SubCust = NonNullable<SPDetail['customers']>[number];
	const exportCustomerColumns: ExportColumn<SubCust>[] = [
		{ header: 'ID', getValue: (c) => c.id },
		{ header: 'Customer', getValue: (c) => c.name },
		{ header: 'Phone', getValue: (c) => c.phone || '—' },
		{ header: 'Email', getValue: (c) => c.email },
		{ header: 'Wallet Balance (BDT)', getValue: (c) => `BDT ${c.balance.toFixed(2)}` },
		{ header: 'Status', getValue: (c) => c.status },
		{ header: 'Created At', getValue: (c) => c.created_at }
	];

	type SpTx = NonNullable<SPDetail['recent_transactions']>[number];
	const exportTransactionColumns: ExportColumn<SpTx>[] = [
		{ header: 'ID', getValue: (t) => t.id },
		{ header: 'Direction', getValue: (t) => t.direction },
		{ header: 'Amount (BDT)', getValue: (t) => `BDT ${t.amount.toFixed(2)}` },
		{ header: 'Balance Before (BDT)', getValue: (t) => `BDT ${t.balance_before.toFixed(2)}` },
		{ header: 'Balance After (BDT)', getValue: (t) => `BDT ${t.balance_after.toFixed(2)}` },
		{ header: 'Description', getValue: (t) => t.description || '—' },
		{ header: 'Date', getValue: (t) => t.created_at }
	];

	// Modals
	let depositModalOpen = $state(false);
	let depositAmount = $state(1000);
	let depositDescription = $state('');
	let depositSubmitting = $state(false);

	let editModalOpen = $state(false);
	let editData = $state({
		name: '',
		email: '',
		phone: '',
		business_name: '',
		status: 'approved',
		api_access: false,
		sms_auto_approved: false
	});
	let editSubmitting = $state(false);

	async function loadSP() {
		loading = true;
		error = null;
		try {
			const res = await supermanApi.get<{ data: SPDetail }>(`/api/v1/superman/sps/${spId}`);
			detail = res.data;
			editData = {
				name: detail.sp.name,
				email: detail.sp.email,
				phone: detail.sp.phone,
				business_name: detail.sp.business_name ?? '',
				status: detail.sp.status,
				api_access: Boolean(detail.api_access),
				sms_auto_approved: Boolean(detail.sms_auto_approved)
			};
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Unable to load reseller details.';
		} finally {
			loading = false;
		}
	}

	async function submitDeposit() {
		if (depositAmount === 0) return;
		depositSubmitting = true;
		try {
			const res = await supermanApi.post<{ message: string; data: { balance: number } }>(
				`/api/v1/superman/sps/${spId}/deposit`,
				{ amount: depositAmount, description: depositDescription }
			);
			if (detail) {
				detail.sp.balance = res.data.balance;
			}
			depositModalOpen = false;
			window.alert(res.message);
			void loadSP();
		} catch (err: unknown) {
			window.alert(err instanceof Error ? err.message : 'Unable to adjust balance.');
		} finally {
			depositSubmitting = false;
		}
	}

	async function submitEdit() {
		editSubmitting = true;
		try {
			await supermanApi.put(`/api/v1/superman/sps/${spId}`, editData);
			editModalOpen = false;
			window.alert('Reseller updated successfully.');
			void loadSP();
		} catch (err: unknown) {
			window.alert(err instanceof Error ? err.message : 'Unable to update reseller.');
		} finally {
			editSubmitting = false;
		}
	}

	onMount(() => {
		void loadSP();
	});
</script>

<svelte:head>
	<title>{detail ? `${detail.sp.name} — S P Reseller` : 'Reseller Profile'} — SMS Whiteinfo</title>
</svelte:head>

<div class="space-y-6 text-gray-900 dark:text-white antialiased font-sans">
	<!-- Breadcrumb -->
	<nav class="flex items-center gap-1.5 text-xs text-gray-500 dark:text-gray-400">
		<a href="/superman/users/sps" class="hover:text-emerald-600 transition-colors">S P</a>
		<span>›</span>
		<span class="font-medium text-gray-900 dark:text-white">{detail?.sp.name ?? `Reseller #${spId}`}</span>
	</nav>

	{#if loading}
		<div class="rounded-2xl bg-white p-16 text-center dark:bg-gray-900 ring-1 ring-gray-950/5 dark:ring-white/10">
			<div class="inline-block h-8 w-8 animate-spin rounded-full border-2 border-emerald-600 border-t-transparent"></div>
			<p class="mt-3 text-xs text-gray-500">Loading reseller account and customer networks...</p>
		</div>
	{:else if error || !detail}
		<div class="rounded-2xl bg-white p-8 text-center dark:bg-gray-900 ring-1 ring-gray-950/5">
			<p class="text-rose-600 text-sm">{error || 'Reseller not found.'}</p>
			<a href="/superman/users/sps" class="mt-3 inline-block text-xs font-semibold text-emerald-600 hover:underline">
				← Back to S P list
			</a>
		</div>
	{:else}
		<!-- Reseller Hero Header -->
		<div class="rounded-2xl bg-white p-6 shadow-xs ring-1 ring-gray-950/5 dark:bg-gray-900 dark:ring-white/10">
			<div class="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-4">
				<div class="flex items-start gap-4">
					<div class="flex h-14 w-14 shrink-0 items-center justify-center rounded-2xl bg-sky-100 text-sky-800 dark:bg-sky-950/60 dark:text-sky-300 text-xl font-bold">
						{detail.sp.name.slice(0, 1).toUpperCase()}
					</div>
					<div class="space-y-1">
						<div class="flex flex-wrap items-center gap-2">
							<h1 class="text-xl font-bold tracking-tight text-gray-950 dark:text-white">{detail.sp.name}</h1>
							<span class="rounded-md bg-gray-100 px-2 py-0.5 text-[11px] font-mono font-bold text-gray-600 dark:bg-gray-800 dark:text-gray-300">
								ID #{detail.sp.id}
							</span>
							{#if detail.sp.status === 'approved'}
								<span class="inline-flex items-center gap-1 rounded-full bg-emerald-50 px-2.5 py-0.5 text-[11px] font-semibold text-emerald-700 border border-emerald-200 dark:bg-emerald-950/40 dark:border-emerald-800 dark:text-emerald-400">
									<span class="h-1.5 w-1.5 rounded-full bg-emerald-500"></span>
									Active Reseller
								</span>
							{:else}
								<span class="inline-flex items-center gap-1 rounded-full bg-amber-50 px-2.5 py-0.5 text-[11px] font-semibold text-amber-700 border border-amber-200 dark:bg-amber-950/40 dark:border-amber-800 dark:text-amber-400">
									<span class="h-1.5 w-1.5 rounded-full bg-amber-500"></span>
									{detail.sp.status}
								</span>
							{/if}
						</div>
						<div class="flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-gray-500 dark:text-gray-400">
							<span>✉ {detail.sp.email}</span>
							<span>📞 {detail.sp.phone || 'No phone'}</span>
							<span>👥 {detail.sp.customer_count} Customers Managed</span>
						</div>
					</div>
				</div>

				<div class="flex items-center gap-2">
					<ExportDropdown
						title={`${detail.sp.name} - Reseller Customers`}
						columns={exportCustomerColumns}
						getData={() => detail?.customers ?? []}
						disabled={!detail?.customers || detail.customers.length === 0}
					/>
					<button
						type="button"
						onclick={() => (depositModalOpen = true)}
						class="inline-flex items-center gap-1.5 rounded-lg bg-emerald-600 px-3.5 py-2 text-xs font-bold text-white shadow-xs hover:bg-emerald-700 transition"
					>
						+ Make Deposit
					</button>
					<button
						type="button"
						onclick={() => (editModalOpen = true)}
						class="inline-flex items-center gap-1.5 rounded-lg border border-gray-300 bg-white px-3.5 py-2 text-xs font-semibold text-gray-700 shadow-2xs hover:bg-gray-50 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-200 transition"
					>
						✏ Edit Reseller
					</button>
				</div>
			</div>
		</div>

		<!-- Financials & Wholesale Plan -->
		<div class="grid grid-cols-1 md:grid-cols-3 gap-6">
			<!-- Wallet Balance -->
			<div class="rounded-xl bg-white p-5 shadow-xs ring-1 ring-gray-950/5 dark:bg-gray-900 dark:ring-white/10 flex flex-col justify-between">
				<div>
					<span class="text-xs font-semibold text-gray-500 uppercase tracking-wider">Reseller Wallet Reserve</span>
					<div class="mt-2 text-3xl font-extrabold text-emerald-600 dark:text-emerald-400 font-mono tracking-tight">
						৳{detail.sp.balance.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
					</div>
					<p class="mt-1 text-xs text-gray-400">Funds available for retail distribution and SMS balance transfers</p>
				</div>
				<button
					type="button"
					onclick={() => (depositModalOpen = true)}
					class="mt-4 w-full rounded-lg bg-emerald-50 py-2 text-xs font-bold text-emerald-700 hover:bg-emerald-100 dark:bg-emerald-950/40 dark:text-emerald-300 text-center transition"
				>
					+ Adjust Wallet
				</button>
			</div>

			<!-- Wholesale Plan & Permissions -->
			<div class="md:col-span-2 rounded-xl bg-white p-5 shadow-xs ring-1 ring-gray-950/5 dark:bg-gray-900 dark:ring-white/10 space-y-4">
				<h3 class="text-xs font-bold uppercase tracking-wider text-gray-700 dark:text-gray-300">Wholesale Plan &amp; Permissions</h3>
				<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
					<div class="p-3.5 rounded-lg bg-gray-50 dark:bg-gray-800/40 space-y-1">
						<span class="text-xs font-medium text-gray-500">Wholesale Tariff Plan</span>
						<p class="text-sm font-bold text-sky-700 dark:text-sky-300">{detail.sp.wholesale_plan || 'No Wholesale Plan Assigned'}</p>
					</div>

					<div class="p-3.5 rounded-lg bg-gray-50 dark:bg-gray-800/40 space-y-1">
						<span class="text-xs font-medium text-gray-500">Total Customer Accounts</span>
						<p class="text-sm font-bold text-gray-900 dark:text-white font-mono">{detail.sp.customer_count} Customers</p>
					</div>

					<div class="flex items-center justify-between p-3.5 rounded-lg bg-gray-50 dark:bg-gray-800/40">
						<span class="text-xs font-semibold text-gray-900 dark:text-white">API Access</span>
						<span class="rounded px-2 py-0.5 text-xs font-bold {detail.api_access ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300' : 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400'}">
							{detail.api_access ? 'Enabled' : 'Disabled'}
						</span>
					</div>

					<div class="flex items-center justify-between p-3.5 rounded-lg bg-gray-50 dark:bg-gray-800/40">
						<span class="text-xs font-semibold text-gray-900 dark:text-white">Bulk Auto-Approve</span>
						<span class="rounded px-2 py-0.5 text-xs font-bold {detail.sms_auto_approved ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300' : 'bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-300'}">
							{detail.sms_auto_approved ? 'Enabled' : 'Disabled'}
						</span>
					</div>
				</div>
			</div>
		</div>

		<!-- Customers Managed Table -->
		<div class="rounded-xl bg-white shadow-xs ring-1 ring-gray-950/5 dark:bg-gray-900 dark:ring-white/10 overflow-hidden">
			<div class="px-5 py-3.5 border-b border-gray-100 dark:border-gray-800 flex items-center justify-between">
				<h3 class="text-xs font-bold uppercase tracking-wider text-gray-700 dark:text-gray-300">
					Customers Registered Under {detail.sp.name}
				</h3>
				<div class="flex items-center gap-2">
					<ExportDropdown
						title={`${detail.sp.name} - Managed Sub-Customers`}
						columns={exportCustomerColumns}
						getData={() => detail?.customers ?? []}
						disabled={!detail?.customers || detail.customers.length === 0}
					/>
					<span class="text-xs text-gray-400 font-mono">{detail.customers?.length ?? 0} Listed</span>
				</div>
			</div>
			<div class="overflow-x-auto">
				<table class="w-full text-xs text-left">
					<thead class="bg-gray-50 dark:bg-gray-800/50 text-[10px] uppercase font-bold text-gray-400">
						<tr>
							<th class="px-4 py-2.5">Customer</th>
							<th class="px-4 py-2.5">Phone</th>
							<th class="px-4 py-2.5">Email</th>
							<th class="px-4 py-2.5">Wallet</th>
							<th class="px-4 py-2.5">Status</th>
							<th class="px-4 py-2.5 text-right">Created</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
						{#if !detail.customers || detail.customers.length === 0}
							<tr>
								<td colspan="6" class="px-4 py-8 text-center text-xs text-gray-400">No customers registered under this reseller yet.</td>
							</tr>
						{:else}
							{#each detail.customers as c}
								<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30">
									<td class="px-4 py-2.5 font-semibold text-gray-900 dark:text-white">
										<a href={`/superman/users/customers/${c.id}`} class="hover:text-emerald-600 transition-colors">
											{c.name}
										</a>
									</td>
									<td class="px-4 py-2.5 font-mono text-gray-600 dark:text-gray-400">{c.phone || '—'}</td>
									<td class="px-4 py-2.5 text-gray-600 dark:text-gray-400">{c.email}</td>
									<td class="px-4 py-2.5 font-mono font-bold text-emerald-600 dark:text-emerald-400">৳{c.balance.toFixed(2)}</td>
									<td class="px-4 py-2.5">
										<span class="inline-block rounded px-2 py-0.5 text-[10px] font-semibold {c.status === 'approved' ? 'bg-emerald-50 text-emerald-700' : 'bg-gray-100 text-gray-600'}">
											{c.status}
										</span>
									</td>
									<td class="px-4 py-2.5 text-right font-mono text-gray-400">{new Date(c.created_at).toLocaleDateString()}</td>
								</tr>
							{/each}
						{/if}
					</tbody>
				</table>
			</div>
		</div>

		<!-- Recent Transactions Table -->
		<div class="rounded-xl bg-white shadow-xs ring-1 ring-gray-950/5 dark:bg-gray-900 dark:ring-white/10 overflow-hidden">
			<div class="px-5 py-3.5 border-b border-gray-100 dark:border-gray-800 flex items-center justify-between">
				<h3 class="text-xs font-bold uppercase tracking-wider text-gray-700 dark:text-gray-300">Recent Wallet Transactions</h3>
				<ExportDropdown
					title={`${detail.sp.name} - Wallet Transactions`}
					columns={exportTransactionColumns}
					getData={() => detail?.recent_transactions ?? []}
					disabled={!detail?.recent_transactions || detail.recent_transactions.length === 0}
				/>
			</div>
			<div class="overflow-x-auto">
				<table class="w-full text-xs text-left">
					<thead class="bg-gray-50 dark:bg-gray-800/50 text-[10px] uppercase font-bold text-gray-400">
						<tr>
							<th class="px-4 py-2.5">ID</th>
							<th class="px-4 py-2.5">Amount</th>
							<th class="px-4 py-2.5">Direction</th>
							<th class="px-4 py-2.5">Balance After</th>
							<th class="px-4 py-2.5">Description</th>
							<th class="px-4 py-2.5 text-right">Date</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
						{#if !detail.recent_transactions || detail.recent_transactions.length === 0}
							<tr>
								<td colspan="6" class="px-4 py-6 text-center text-xs text-gray-400">No transaction records found.</td>
							</tr>
						{:else}
							{#each detail.recent_transactions as tx}
								<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30">
									<td class="px-4 py-2 font-mono text-gray-400">#{tx.id}</td>
									<td class="px-4 py-2 font-mono font-bold {tx.direction === 'credit' ? 'text-emerald-600' : 'text-rose-600'}">
										{tx.direction === 'credit' ? '+' : '-'}৳{tx.amount.toFixed(2)}
									</td>
									<td class="px-4 py-2 uppercase text-[10px] font-semibold">{tx.direction}</td>
									<td class="px-4 py-2 font-mono">৳{tx.balance_after.toFixed(2)}</td>
									<td class="px-4 py-2 text-gray-600 dark:text-gray-400 truncate max-w-xs">{tx.description || '—'}</td>
									<td class="px-4 py-2 text-right font-mono text-gray-400">{new Date(tx.created_at).toLocaleString()}</td>
								</tr>
							{/each}
						{/if}
					</tbody>
				</table>
			</div>
		</div>
	{/if}
</div>

<!-- ============================================================
     DEPOSIT MODAL
     ============================================================ -->
{#if depositModalOpen && detail}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-gray-950/60 p-4 backdrop-blur-xs">
		<div class="w-full max-w-md rounded-2xl bg-white p-6 shadow-2xl ring-1 ring-gray-900/10 dark:bg-gray-900 dark:ring-white/10">
			<h3 class="text-base font-bold text-gray-900 dark:text-white">Adjust Reseller Wallet</h3>
			<p class="text-xs text-gray-500 mt-0.5">{detail.sp.name} (Current: ৳{detail.sp.balance.toFixed(2)})</p>

			<div class="mt-4 space-y-3">
				<div>
					<label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">
						Amount (BDT) <span class="text-gray-400 font-normal">— enter negative value to deduct</span>
					</label>
					<input
						type="number"
						step="0.01"
						bind:value={depositAmount}
						class="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm text-gray-900 shadow-2xs outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
					/>
				</div>

				<div>
					<label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Remarks / Note</label>
					<input
						type="text"
						bind:value={depositDescription}
						placeholder="e.g. Bank wire deposit ref #98765"
						class="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-xs text-gray-900 shadow-2xs outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
					/>
				</div>
			</div>

			<div class="mt-6 flex items-center justify-end gap-2">
				<button
					type="button"
					onclick={() => (depositModalOpen = false)}
					class="rounded-lg border border-gray-300 px-3.5 py-2 text-xs font-semibold text-gray-700 hover:bg-gray-50 dark:border-gray-700 dark:text-gray-300"
				>
					Cancel
				</button>
				<button
					type="button"
					disabled={depositSubmitting}
					onclick={submitDeposit}
					class="rounded-lg bg-emerald-600 px-4 py-2 text-xs font-bold text-white shadow-xs hover:bg-emerald-700 disabled:opacity-50"
				>
					{depositSubmitting ? 'Adjusting...' : 'Confirm Balance Adjustment'}
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- ============================================================
     EDIT SP MODAL
     ============================================================ -->
{#if editModalOpen && detail}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-gray-950/60 p-4 backdrop-blur-xs">
		<div class="w-full max-w-lg rounded-2xl bg-white p-6 shadow-2xl ring-1 ring-gray-900/10 dark:bg-gray-900 dark:ring-white/10">
			<h3 class="text-base font-bold text-gray-900 dark:text-white">Edit S P Reseller</h3>

			<div class="mt-4 space-y-3">
				<div>
					<label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Company / S P Name</label>
					<input
						type="text"
						bind:value={editData.name}
						class="w-full rounded-lg border border-gray-300 px-3 py-1.5 text-xs dark:border-gray-700 dark:bg-gray-800 dark:text-white"
					/>
				</div>

				<div class="grid grid-cols-2 gap-3">
					<div>
						<label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Email</label>
						<input
							type="email"
							bind:value={editData.email}
							class="w-full rounded-lg border border-gray-300 px-3 py-1.5 text-xs dark:border-gray-700 dark:bg-gray-800 dark:text-white"
						/>
					</div>
					<div>
						<label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Phone</label>
						<input
							type="text"
							bind:value={editData.phone}
							class="w-full rounded-lg border border-gray-300 px-3 py-1.5 text-xs dark:border-gray-700 dark:bg-gray-800 dark:text-white"
						/>
					</div>
				</div>

				<div class="grid grid-cols-2 gap-3">
					<div>
						<label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Business Name</label>
						<input
							type="text"
							bind:value={editData.business_name}
							class="w-full rounded-lg border border-gray-300 px-3 py-1.5 text-xs dark:border-gray-700 dark:bg-gray-800 dark:text-white"
						/>
					</div>
					<div>
						<label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Status</label>
						<select
							bind:value={editData.status}
							class="w-full rounded-lg border border-gray-300 px-3 py-1.5 text-xs dark:border-gray-700 dark:bg-gray-800 dark:text-white"
						>
							<option value="approved">Approved</option>
							<option value="pending">Pending</option>
							<option value="rejected">Rejected</option>
							<option value="banned">Banned</option>
						</select>
					</div>
				</div>

				<div class="pt-2 border-t border-gray-100 dark:border-gray-800 space-y-2">
					<label class="flex items-center gap-2 cursor-pointer">
						<input type="checkbox" bind:checked={editData.api_access} class="rounded text-emerald-600" />
						<span class="text-xs font-semibold text-gray-800 dark:text-gray-200">API Access Enabled</span>
					</label>
					<label class="flex items-center gap-2 cursor-pointer">
						<input type="checkbox" bind:checked={editData.sms_auto_approved} class="rounded text-emerald-600" />
						<span class="text-xs font-semibold text-gray-800 dark:text-gray-200">Bulk SMS Auto-Approve</span>
					</label>
				</div>
			</div>

			<div class="mt-6 flex items-center justify-end gap-2">
				<button
					type="button"
					onclick={() => (editModalOpen = false)}
					class="rounded-lg border border-gray-300 px-3.5 py-2 text-xs font-semibold text-gray-700 hover:bg-gray-50 dark:border-gray-700 dark:text-gray-300"
				>
					Cancel
				</button>
				<button
					type="button"
					disabled={editSubmitting}
					onclick={submitEdit}
					class="rounded-lg bg-emerald-600 px-4 py-2 text-xs font-bold text-white shadow-xs hover:bg-emerald-700 disabled:opacity-50"
				>
					{editSubmitting ? 'Saving...' : 'Save Reseller'}
				</button>
			</div>
		</div>
	</div>
{/if}
