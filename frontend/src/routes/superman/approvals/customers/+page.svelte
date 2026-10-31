<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface PendingCustomer {
		id: number;
		name: string;
		email: string;
		phone: string;
		business_name?: string;
		kyc_status?: string;
		created_at: string;
	}

	let customers = $state<PendingCustomer[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let processingId = $state<number | null>(null);

	async function loadPending() {
		loading = true;
		error = null;
		try {
			const res = await supermanApi.get<{ data: PendingCustomer[] }>('/api/v1/superman/customers?status=pending');
			customers = res.data ?? [];
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Failed to load pending customer approvals.';
		} finally {
			loading = false;
		}
	}

	async function approveCustomer(id: number) {
		processingId = id;
		try {
			await supermanApi.put(`/api/v1/superman/customers/${id}/status`, { status: 'active' });
			await loadPending();
		} catch (err: unknown) {
			alert(err instanceof Error ? err.message : 'Approval failed.');
		} finally {
			processingId = null;
		}
	}

	async function rejectCustomer(id: number) {
		processingId = id;
		try {
			await supermanApi.put(`/api/v1/superman/customers/${id}/status`, { status: 'rejected' });
			await loadPending();
		} catch (err: unknown) {
			alert(err instanceof Error ? err.message : 'Rejection failed.');
		} finally {
			processingId = null;
		}
	}

	onMount(() => {
		void loadPending();
	});

	const exportColumns: ExportColumn<PendingCustomer>[] = [
		{ header: 'Customer Name', getValue: (c) => c.name },
		{ header: 'Email', getValue: (c) => c.email },
		{ header: 'Phone', getValue: (c) => c.phone },
		{ header: 'Business / Company', getValue: (c) => c.business_name || '—' },
		{ header: 'KYC Status', getValue: (c) => (c.kyc_status ? c.kyc_status.toUpperCase() : 'UNSUBMITTED') },
		{ header: 'Registered Date', getValue: (c) => (c.created_at ? new Date(c.created_at).toLocaleString() : '—') }
	];
</script>

<svelte:head>
	<title>Pending Customer Approvals — Superman Administrator</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">Pending Customer Registrations</h1>
			<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
				Review and authorize new client accounts before granting SMS dispatch and deposit capabilities.
			</p>
		</div>
		<div class="flex items-center gap-2">
			<button onclick={loadPending} class="rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs font-medium text-gray-700 shadow-xs hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300">
				Refresh List
			</button>
			<ExportDropdown
				title="Pending Customer Approvals"
				columns={exportColumns}
				getData={() => customers}
				disabled={customers.length === 0}
			/>
		</div>
	</div>

	<div class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-xs dark:border-gray-800 dark:bg-gray-900">
		{#if loading}
			<div class="py-16 text-center text-xs text-gray-500">Scanning pending registrations...</div>
		{:else if error}
			<div class="p-6 text-center text-xs text-rose-600">{error}</div>
		{:else if customers.length === 0}
			<div class="py-16 text-center text-xs text-gray-500">
				No customers currently pending activation. All accounts processed.
			</div>
		{:else}
			<table class="w-full text-left text-xs">
				<thead>
					<tr class="border-b border-gray-100 bg-gray-50/50 text-[10px] font-semibold uppercase tracking-wider text-gray-500 dark:border-gray-800 dark:bg-gray-800/50 dark:text-gray-400">
						<th class="px-5 py-3.5">Customer Name</th>
						<th class="px-5 py-3.5">Email & Phone</th>
						<th class="px-5 py-3.5">Company / Business</th>
						<th class="px-5 py-3.5">KYC Status</th>
						<th class="px-5 py-3.5">Registered Date</th>
						<th class="px-5 py-3.5 text-right">Actions</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
					{#each customers as c}
						<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30">
							<td class="px-5 py-4 font-semibold text-gray-950 dark:text-white">{c.name}</td>
							<td class="px-5 py-4 text-gray-600 dark:text-gray-400">
								<div>{c.email}</div>
								<div class="text-[11px] text-gray-400">{c.phone}</div>
							</td>
							<td class="px-5 py-4 text-gray-700 dark:text-gray-300">{c.business_name || 'Individual'}</td>
							<td class="px-5 py-4">
								<span class="inline-flex rounded-full bg-amber-50 px-2 py-0.5 text-[10px] font-bold uppercase text-amber-700">
									{c.kyc_status || 'Pending'}
								</span>
							</td>
							<td class="px-5 py-4 text-gray-400">{new Date(c.created_at).toLocaleDateString()}</td>
							<td class="px-5 py-4 text-right">
								<div class="flex items-center justify-end gap-2">
									<button
										disabled={processingId === c.id}
										onclick={() => approveCustomer(c.id)}
										class="rounded bg-emerald-600 px-2.5 py-1 text-xs font-semibold text-white shadow-xs hover:bg-emerald-700 disabled:opacity-50"
									>
										Approve
									</button>
									<button
										disabled={processingId === c.id}
										onclick={() => rejectCustomer(c.id)}
										class="rounded bg-rose-50 px-2.5 py-1 text-xs font-semibold text-rose-700 hover:bg-rose-100 dark:bg-rose-950/40 dark:text-rose-300 disabled:opacity-50"
									>
										Reject
									</button>
								</div>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		{/if}
	</div>
</div>

