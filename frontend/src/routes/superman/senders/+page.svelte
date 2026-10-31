<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface CustomerRef {
		id: number;
		name: string;
	}

	interface Sender {
		id: number;
		sender_id: string;
		type: string;
		masking_category: string;
		is_enabled: boolean;
		approval_status: string;
		rejection_reason: string | null;
		gateway_name: string | null;
		requested_by_name: string | null;
		reviewed_at: string | null;
		created_at: string;
		users: CustomerRef[];
		users_count: number;
		messages_count: number;
	}

	let searchQuery = $state('');
	let typeFilter = $state('');
	let statusFilter = $state('');

	let senders = $state<Sender[]>([]);
	let loading = $state(false);

	let page = $state(1);
	let perPage = $state(15);
	let total = $state(0);
	let totalPages = $state(1);

	async function loadSenders(newPage = page) {
		loading = true;
		page = newPage;
		try {
			let url = `/api/v1/superman/senders?page=${page}&per_page=${perPage}`;
			if (typeFilter) url += `&type=${encodeURIComponent(typeFilter)}`;
			if (statusFilter) url += `&status=${encodeURIComponent(statusFilter)}`;
			if (searchQuery.trim()) url += `&search=${encodeURIComponent(searchQuery.trim())}`;
			const response = await supermanApi.get<{
				data: Sender[];
				meta: { page: number; per_page: number; total: number; total_pages: number };
			}>(url);
			senders = response.data || [];
			if (response.meta) {
				total = response.meta.total;
				totalPages = response.meta.total_pages;
			}
		} catch (error: unknown) {
			window.alert(error instanceof Error ? error.message : 'Unable to load sender IDs.');
		} finally {
			loading = false;
		}
	}

	async function toggleEnabled(s: Sender) {
		const targetState = !s.is_enabled;
		try {
			await supermanApi.patch(`/api/v1/superman/senders/${s.id}/enabled`, {
				is_enabled: targetState
			});
			s.is_enabled = targetState;
		} catch (err: unknown) {
			window.alert(err instanceof Error ? err.message : 'Unable to update sender state.');
		}
	}

	async function approveSender(s: Sender) {
		if (!confirm(`Are you sure you want to approve Sender ID "${s.sender_id}"?`)) return;
		try {
			const res = await supermanApi.post<{ message: string }>(`/api/v1/superman/senders/${s.id}/approve`);
			s.approval_status = 'approved';
			s.is_enabled = true;
			window.alert(res.message);
		} catch (error: unknown) {
			window.alert(error instanceof Error ? error.message : 'Unable to approve sender ID.');
		}
	}

	async function rejectSender(s: Sender) {
		const reason = prompt(`Enter reason for rejecting Sender ID "${s.sender_id}":`);
		if (!reason) return;
		try {
			const res = await supermanApi.post<{ message: string }>(`/api/v1/superman/senders/${s.id}/reject`, { reason });
			s.approval_status = 'rejected';
			s.is_enabled = false;
			s.rejection_reason = reason;
			window.alert(res.message);
		} catch (error: unknown) {
			window.alert(error instanceof Error ? error.message : 'Unable to reject sender ID.');
		}
	}

	onMount(() => {
		void loadSenders();
	});

	const exportColumns: ExportColumn<Sender>[] = [
		{ header: 'Sender ID', getValue: (s) => s.sender_id },
		{ header: 'Type', getValue: (s) => s.type },
		{ header: 'Gateway', getValue: (s) => s.gateway_name || 'Unassigned' },
		{ header: 'Assigned Customers', getValue: (s) => (s.users && s.users.length > 0 ? s.users.map((u) => u.name).join(', ') : 'None') },
		{ header: 'Messages Sent', getValue: (s) => s.messages_count },
		{ header: 'Approval Status', getValue: (s) => s.approval_status.toUpperCase() },
		{ header: 'Operating State', getValue: (s) => (s.is_enabled ? 'Enabled' : 'Disabled') },
		{ header: 'Created Date', getValue: (s) => (s.created_at ? new Date(s.created_at).toLocaleDateString() : '—') }
	];
</script>

<svelte:head>
	<title>Sender IDs — SMS Whiteinfo Administrator</title>
</svelte:head>

<div class="space-y-5">
	<!-- Page Header -->
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
		<div>
			<h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">Sender ID / CLI Management</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400">Review, assign, and approve customer and reseller alphanumeric sender headers.</p>
		</div>
		<div class="flex items-center gap-2">
			<ExportDropdown
				title="Sender IDs"
				columns={exportColumns}
				getData={() => senders}
				disabled={senders.length === 0}
			/>
			<a
				href="/superman/sender-cli/masking"
				class="inline-flex items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs font-semibold text-gray-700 hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300 transition"
			>
				Masking Senders
			</a>
			<a
				href="/superman/sender-cli/non-masking"
				class="inline-flex items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs font-semibold text-gray-700 hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300 transition"
			>
				Non-Masking Senders
			</a>
			<a
				href="/superman/sender-cli/add"
				class="inline-flex items-center gap-1.5 rounded-lg bg-emerald-600 px-3.5 py-2 text-xs font-semibold text-white shadow-xs hover:bg-emerald-500 transition"
			>
				<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
				</svg>
				Add Sender
			</a>
			<button
				type="button"
				onclick={() => void loadSenders(page)}
				class="inline-flex items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs font-semibold text-gray-700 shadow-2xs hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300 transition"
			>
				<svg class="h-3.5 w-3.5 {loading ? 'animate-spin' : ''}" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0 3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182m0-4.991v4.99" /></svg>
				Refresh
			</button>
		</div>
	</div>

	<!-- Table Container Card -->
	<div class="rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900 space-y-4">
		<!-- Search & Filter Bar -->
		<div class="flex flex-wrap items-center justify-between gap-3">
			<div class="flex flex-wrap items-center gap-2">
				<div class="relative w-full sm:w-64">
					<input
						type="text"
						bind:value={searchQuery}
						onkeydown={(e) => e.key === 'Enter' && void loadSenders(1)}
						placeholder="Search sender IDs..."
						class="w-full rounded-lg border border-gray-200 bg-gray-50/70 py-1.5 pl-8 pr-3 text-xs outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
					/>
					<svg class="absolute left-2.5 top-2 h-4 w-4 text-gray-400" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z" /></svg>
				</div>
				<select
					bind:value={typeFilter}
					onchange={() => void loadSenders(1)}
					class="rounded-lg border border-gray-200 bg-white px-2.5 py-1.5 text-xs text-gray-700 outline-none dark:border-gray-700 dark:bg-gray-800 dark:text-gray-300"
				>
					<option value="">All Types</option>
					<option value="masking">Masking</option>
					<option value="non_masking">Non-Masking</option>
				</select>
				<select
					bind:value={statusFilter}
					onchange={() => void loadSenders(1)}
					class="rounded-lg border border-gray-200 bg-white px-2.5 py-1.5 text-xs text-gray-700 outline-none dark:border-gray-700 dark:bg-gray-800 dark:text-gray-300"
				>
					<option value="">All Statuses</option>
					<option value="pending">Pending</option>
					<option value="approved">Approved</option>
					<option value="rejected">Rejected</option>
				</select>
			</div>
			<div class="text-xs text-gray-400">
				Total: <span class="font-bold text-gray-700 dark:text-gray-300">{total}</span> sender IDs
			</div>
		</div>

		<!-- Senders Data Table -->
		<div class="overflow-x-auto">
			<table class="w-full min-w-[900px] text-left text-xs">
				<thead>
					<tr class="border-b border-gray-100 dark:border-gray-800 text-gray-400 uppercase text-[10px] font-semibold">
						<th class="py-2.5">Sender ID</th>
						<th class="py-2.5">Type</th>
						<th class="py-2.5">Requested By</th>
						<th class="py-2.5">Gateway</th>
						<th class="py-2.5">Approval Status</th>
						<th class="py-2.5">Operating State</th>
						<th class="py-2.5">Created At</th>
						<th class="py-2.5 text-right">Actions</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
					{#if senders.length === 0}
						<tr>
							<td colspan="8" class="py-8 text-center text-gray-400 text-xs">No sender IDs match the current filters.</td>
						</tr>
					{/if}
					{#each senders as s}
						<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30 transition-colors">
							<td class="py-3 font-mono font-bold text-gray-900 dark:text-white">
								{s.sender_id}
							</td>
							<td class="py-3">
								<span class="inline-block rounded px-2 py-0.5 text-[10px] font-medium {s.type !== 'numeric' ? 'bg-indigo-50 text-indigo-700 border border-indigo-200 dark:bg-indigo-950/40 dark:border-indigo-800 dark:text-indigo-300' : 'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300'}">
									{s.type}
								</span>
							</td>
							<td class="py-3 text-gray-700 dark:text-gray-300">{s.requested_by_name ?? 'Direct'}</td>
							<td class="py-3 text-gray-500 font-mono text-[11px]">{s.gateway_name ?? 'Default'}</td>
							<td class="py-3">
								<span class="inline-block rounded-full px-2 py-0.5 text-[10px] font-semibold {s.approval_status === 'approved' ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300' : s.approval_status === 'pending' ? 'bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-300' : 'bg-rose-100 text-rose-800 dark:bg-rose-950 dark:text-rose-300'}">
									{s.approval_status}
								</span>
							</td>
							<td class="py-3">
								<button
									type="button"
									onclick={() => void toggleEnabled(s)}
									class="inline-flex items-center gap-1 text-[11px] font-medium transition {s.is_enabled ? 'text-emerald-600 dark:text-emerald-400' : 'text-gray-400'}"
								>
									<span class="h-1.5 w-1.5 rounded-full {s.is_enabled ? 'bg-emerald-500' : 'bg-gray-400'}"></span>
									{s.is_enabled ? 'Enabled' : 'Disabled'}
								</button>
							</td>
							<td class="py-3 text-[11px] font-mono text-gray-400">{new Date(s.created_at).toLocaleDateString()}</td>
							<td class="py-3 text-right">
								<div class="inline-flex items-center gap-1.5">
									{#if s.approval_status !== 'approved'}
										<button
											type="button"
											onclick={() => void approveSender(s)}
											class="rounded bg-emerald-50 px-2 py-1 text-[10px] font-bold text-emerald-700 hover:bg-emerald-100 dark:bg-emerald-950/40 dark:text-emerald-300 transition"
										>
											Approve
										</button>
									{/if}
									{#if s.approval_status !== 'rejected'}
										<button
											type="button"
											onclick={() => void rejectSender(s)}
											class="rounded bg-rose-50 px-2 py-1 text-[10px] font-bold text-rose-700 hover:bg-rose-100 dark:bg-rose-950/40 dark:text-rose-300 transition"
										>
											Reject
										</button>
									{/if}
								</div>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>

		<!-- Pagination Footer -->
		{#if totalPages > 1}
			<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between border-t border-gray-100 pt-3 dark:border-gray-800 gap-3">
				<div class="text-xs text-gray-500 dark:text-gray-400">
					Showing {(page - 1) * perPage + 1} to {Math.min(page * perPage, total)} of {total} sender IDs
				</div>
				<div class="flex items-center gap-1.5">
					<button
						type="button"
						disabled={page <= 1}
						onclick={() => void loadSenders(page - 1)}
						class="rounded-lg border border-gray-200 px-3 py-1 text-xs font-semibold text-gray-700 disabled:opacity-40 hover:bg-gray-50 dark:border-gray-700 dark:text-gray-300"
					>
						Previous
					</button>
					<span class="px-2 text-xs text-gray-600 dark:text-gray-400">Page {page} of {totalPages}</span>
					<button
						type="button"
						disabled={page >= totalPages}
						onclick={() => void loadSenders(page + 1)}
						class="rounded-lg border border-gray-200 px-3 py-1 text-xs font-semibold text-gray-700 disabled:opacity-40 hover:bg-gray-50 dark:border-gray-700 dark:text-gray-300"
					>
						Next
					</button>
				</div>
			</div>
		{/if}
	</div>
</div>
