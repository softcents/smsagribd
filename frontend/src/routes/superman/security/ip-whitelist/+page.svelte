<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface WhitelistRequest {
		id: number;
		customer_id?: number;
		user_id?: number;
		customer?: string;
		customer_name?: string;
		ip_address: string;
		status: string;
		reason?: string;
		notes?: string;
		created_at: string;
	}

	let requests = $state<WhitelistRequest[]>([]);
	let loading = $state(true);
	let actionMsg = $state<string | null>(null);

	async function loadRequests() {
		loading = true;
		try {
			const res = await supermanApi.get<{ data: WhitelistRequest[] }>('/api/v1/superman/security/ip-whitelist');
			requests = res.data ?? [];
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	}

	onMount(loadRequests);

	async function updateStatus(id: number, action: 'approve' | 'reject') {
		try {
			await supermanApi.post(`/api/v1/superman/approval-center/ip-whitelists/${id}/${action}`, {});
			actionMsg = `IP request #${id} has been ${action}d.`;
			await loadRequests();
		} catch (e: any) {
			alert(e?.message || `Failed to ${action} IP request.`);
		}
	}

	const exportColumns: ExportColumn<WhitelistRequest>[] = [
		{ header: 'Customer', getValue: (r) => r.customer_name || r.customer || '—' },
		{ header: 'IP Address', getValue: (r) => r.ip_address },
		{ header: 'Status', getValue: (r) => r.status.toUpperCase() },
		{ header: 'Reason', getValue: (r) => r.reason || r.notes || 'Production Server IP' },
		{ header: 'Requested At', getValue: (r) => (r.created_at ? new Date(r.created_at).toLocaleString() : '—') }
	];
</script>

<div class="space-y-6">
	<!-- Page Header -->
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-xl font-bold text-gray-900 dark:text-white">IP Whitelist Access Management</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
				Authorize static IP addresses for automated HTTP REST dispatch and developer API integration.
			</p>
		</div>
		<div class="flex items-center gap-2">
			<button
				onclick={loadRequests}
				class="inline-flex items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-3 py-1.5 text-xs font-semibold text-gray-700 shadow-2xs hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300 transition"
			>
				Refresh
			</button>
			<ExportDropdown
				title="IP Whitelist Requests"
				columns={exportColumns}
				getData={() => requests}
				disabled={requests.length === 0}
			/>
		</div>
	</div>

	{#if actionMsg}
		<div class="rounded-lg bg-emerald-50 dark:bg-emerald-950/40 border border-emerald-200 dark:border-emerald-800 p-4 text-xs font-medium text-emerald-800 dark:text-emerald-300">
			{actionMsg}
		</div>
	{/if}

	<!-- IP Whitelist Requests Table -->
	<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl shadow-xs overflow-hidden">
		<div class="overflow-x-auto">
			<table class="w-full text-left text-xs">
				<thead class="bg-gray-50/75 dark:bg-gray-800/50 text-gray-500 uppercase tracking-wider font-semibold border-b border-gray-200 dark:border-gray-800">
					<tr>
						<th class="px-5 py-3">Customer Account</th>
						<th class="px-5 py-3">Requested IP Address</th>
						<th class="px-5 py-3">Reason / Scope</th>
						<th class="px-5 py-3">Status</th>
						<th class="px-5 py-3">Requested Date</th>
						<th class="px-5 py-3 text-right">Actions</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800 text-gray-700 dark:text-gray-300">
					{#if loading}
						<tr>
							<td colspan="6" class="px-5 py-8 text-center text-gray-400">Loading IP whitelist requests...</td>
						</tr>
					{:else if requests.length === 0}
						<tr>
							<td colspan="6" class="px-5 py-8 text-center text-gray-400">No IP whitelist records found.</td>
						</tr>
					{:else}
						{#each requests as r}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30 transition">
								<td class="px-5 py-3 font-bold text-gray-900 dark:text-white">
									<a href="/superman/users/customers/{r.customer_id || r.user_id || ''}" class="hover:text-emerald-600 transition">
										{r.customer_name || r.customer || '—'}
									</a>
								</td>
								<td class="px-5 py-3 font-mono font-bold text-emerald-600 dark:text-emerald-400">{r.ip_address}</td>
								<td class="px-5 py-3 text-gray-500">{r.reason || r.notes || 'Production Server IP'}</td>
								<td class="px-5 py-3">
									<span class="inline-flex items-center px-2 py-0.5 rounded text-[11px] font-medium {r.status === 'approved' ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300' : r.status === 'rejected' ? 'bg-red-100 text-red-800 dark:bg-red-950 dark:text-red-300' : 'bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-300'}">
										{r.status}
									</span>
								</td>
								<td class="px-5 py-3 text-gray-400">{new Date(r.created_at).toLocaleDateString()}</td>
								<td class="px-5 py-3 text-right">
									{#if r.status === 'pending'}
										<div class="flex items-center justify-end gap-2">
											<button
												type="button"
												onclick={() => updateStatus(r.id, 'approve')}
												class="rounded bg-emerald-600 px-2 py-1 text-[11px] font-semibold text-white hover:bg-emerald-700"
											>
												Approve
											</button>
											<button
												type="button"
												onclick={() => updateStatus(r.id, 'reject')}
												class="rounded bg-red-600 px-2 py-1 text-[11px] font-semibold text-white hover:bg-red-700"
											>
												Reject
											</button>
										</div>
									{:else}
										<span class="text-xs text-gray-400">Processed</span>
									{/if}
								</td>
							</tr>
						{/each}
					{/if}
				</tbody>
			</table>
		</div>
	</div>
</div>

