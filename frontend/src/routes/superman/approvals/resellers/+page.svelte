<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface PendingReseller {
		id: number;
		name: string;
		email: string;
		phone: string;
		business_name?: string;
		status: string;
		created_at: string;
	}

	let resellers = $state<PendingReseller[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let processingId = $state<number | null>(null);

	async function loadPending() {
		loading = true;
		error = null;
		try {
			const res = await supermanApi.get<{ data: PendingReseller[] }>('/api/v1/superman/sps?status=pending');
			resellers = res.data ?? [];
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Failed to load pending reseller applications.';
		} finally {
			loading = false;
		}
	}

	async function approveReseller(id: number) {
		processingId = id;
		try {
			await supermanApi.put(`/api/v1/superman/sps/${id}/status`, { status: 'active' });
			await loadPending();
		} catch (err: unknown) {
			alert(err instanceof Error ? err.message : 'Approval failed.');
		} finally {
			processingId = null;
		}
	}

	async function rejectReseller(id: number) {
		processingId = id;
		try {
			await supermanApi.put(`/api/v1/superman/sps/${id}/status`, { status: 'rejected' });
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

	const exportColumns: ExportColumn<PendingReseller>[] = [
		{ header: 'Reseller Name', getValue: (r) => r.name },
		{ header: 'Email', getValue: (r) => r.email },
		{ header: 'Phone', getValue: (r) => r.phone },
		{ header: 'Business / Company', getValue: (r) => r.business_name || '—' },
		{ header: 'Status', getValue: (r) => r.status.toUpperCase() },
		{ header: 'Application Date', getValue: (r) => (r.created_at ? new Date(r.created_at).toLocaleString() : '—') }
	];
</script>

<svelte:head>
	<title>Pending Reseller Approvals — Superman Administrator</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">Pending Reseller (S P) Applications</h1>
			<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
				Review and authorize new Service Provider partners applying for multi-tenant white-label accounts.
			</p>
		</div>
		<div class="flex items-center gap-2">
			<button onclick={loadPending} class="rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs font-medium text-gray-700 shadow-xs hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300">
				Refresh List
			</button>
			<ExportDropdown
				title="Pending Reseller Applications"
				columns={exportColumns}
				getData={() => resellers}
				disabled={resellers.length === 0}
			/>
		</div>
	</div>

	<div class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-xs dark:border-gray-800 dark:bg-gray-900">
		{#if loading}
			<div class="py-16 text-center text-xs text-gray-500">Scanning pending applications...</div>
		{:else if error}
			<div class="p-6 text-center text-xs text-rose-600">{error}</div>
		{:else if resellers.length === 0}
			<div class="py-16 text-center text-xs text-gray-500">
				No reseller applications pending review.
			</div>
		{:else}
			<table class="w-full text-left text-xs">
				<thead>
					<tr class="border-b border-gray-100 bg-gray-50/50 text-[10px] font-semibold uppercase tracking-wider text-gray-500 dark:border-gray-800 dark:bg-gray-800/50 dark:text-gray-400">
						<th class="px-5 py-3.5">Reseller Name</th>
						<th class="px-5 py-3.5">Contact Email</th>
						<th class="px-5 py-3.5">Phone</th>
						<th class="px-5 py-3.5">Business Name</th>
						<th class="px-5 py-3.5">Applied Date</th>
						<th class="px-5 py-3.5 text-right">Actions</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
					{#each resellers as r}
						<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30">
							<td class="px-5 py-4 font-semibold text-gray-950 dark:text-white">{r.name}</td>
							<td class="px-5 py-4 text-gray-600 dark:text-gray-400">{r.email}</td>
							<td class="px-5 py-4 font-mono">{r.phone}</td>
							<td class="px-5 py-4 font-medium text-gray-700 dark:text-gray-300">{r.business_name || 'Sole Trader'}</td>
							<td class="px-5 py-4 text-gray-400">{new Date(r.created_at).toLocaleDateString()}</td>
							<td class="px-5 py-4 text-right">
								<div class="flex items-center justify-end gap-2">
									<button
										disabled={processingId === r.id}
										onclick={() => approveReseller(r.id)}
										class="rounded bg-emerald-600 px-2.5 py-1 text-xs font-semibold text-white shadow-xs hover:bg-emerald-700 disabled:opacity-50"
									>
										Approve
									</button>
									<button
										disabled={processingId === r.id}
										onclick={() => rejectReseller(r.id)}
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

