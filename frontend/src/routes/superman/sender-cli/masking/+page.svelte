<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface CustomerRef {
		id: number;
		name: string;
		email?: string | null;
	}

	interface GatewayRef {
		id: number;
		name: string;
	}

	interface SenderItem {
		id: number;
		sender_id: string;
		type: string;
		masking_category: string;
		is_enabled: boolean;
		approval_status: string;
		rejection_reason?: string | null;
		gateway_id?: number | null;
		gateway_name?: string | null;
		requested_by_id?: number | null;
		requested_by_name?: string | null;
		reviewed_at?: string | null;
		created_at: string;
		users: CustomerRef[];
		users_count: number;
		messages_count: number;
	}

	let senders = $state<SenderItem[]>([]);
	let gateways = $state<GatewayRef[]>([]);
	let allCustomers = $state<CustomerRef[]>([]);
	let availableSendersList = $state<{ id: number; sender_id: string }[]>([]);

	let loading = $state(false);
	let error = $state<string | null>(null);

	// Filters
	let search = $state('');
	let gatewayFilter = $state('');
	let assignmentFilter = $state('');

	// Pagination
	let page = $state(1);
	let perPage = $state(10);
	let total = $state(0);
	let totalPages = $state(1);

	// Assignment Modal
	let showAssignModal = $state(false);
	let assignSenderId = $state<number | null>(null);
	let assignCustomerIds = $state<number[]>([]);
	let assignSaving = $state(false);
	let assignCustomerSearch = $state('');

	// Dropdown menu state per row
	let activeMenuId = $state<number | null>(null);

	async function loadOptions() {
		try {
			const res = await supermanApi.get<{
				data: {
					gateways: GatewayRef[];
					customers: CustomerRef[];
					senders: { id: number; sender_id: string; type: string }[];
				};
			}>('/api/v1/superman/sender-cli/options?type=masking');
			gateways = res.data.gateways || [];
			allCustomers = res.data.customers || [];
			availableSendersList = (res.data.senders || []).map((s) => ({ id: s.id, sender_id: s.sender_id }));
		} catch (err: unknown) {
			console.error('Failed to load sender-cli options:', err);
		}
	}

	async function loadSenders(newPage = page) {
		loading = true;
		error = null;
		page = newPage;
		try {
			let url = `/api/v1/superman/senders?type=masking&page=${page}&per_page=${perPage}`;
			if (search.trim()) url += `&search=${encodeURIComponent(search.trim())}`;
			if (gatewayFilter) url += `&gateway_id=${encodeURIComponent(gatewayFilter)}`;
			if (assignmentFilter) url += `&assignment=${encodeURIComponent(assignmentFilter)}`;

			const res = await supermanApi.get<{
				data: SenderItem[];
				meta: { page: number; per_page: number; total: number; total_pages: number };
			}>(url);

			senders = res.data || [];
			if (res.meta) {
				total = res.meta.total;
				totalPages = res.meta.total_pages;
			}
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Unable to load masking senders.';
		} finally {
			loading = false;
		}
	}

	function openAssignmentModal(sender?: SenderItem) {
		if (sender) {
			assignSenderId = sender.id;
			assignCustomerIds = (sender.users || []).map((u) => u.id);
		} else {
			assignSenderId = availableSendersList.length > 0 ? availableSendersList[0].id : null;
			assignCustomerIds = [];
		}
		assignCustomerSearch = '';
		showAssignModal = true;
	}

	function toggleCustomerSelection(id: number) {
		if (assignCustomerIds.includes(id)) {
			assignCustomerIds = assignCustomerIds.filter((cid) => cid !== id);
		} else {
			assignCustomerIds = [...assignCustomerIds, id];
		}
	}

	async function saveAssignment() {
		if (!assignSenderId) {
			window.alert('Please select a sender ID.');
			return;
		}
		if (assignCustomerIds.length === 0) {
			window.alert('Please select at least one customer.');
			return;
		}

		assignSaving = true;
		try {
			const res = await supermanApi.post<{ message: string }>(
				`/api/v1/superman/senders/${assignSenderId}/assignment`,
				{ user_ids: assignCustomerIds }
			);
			showAssignModal = false;
			window.alert(res.message || 'Sender ID assigned successfully.');
			void loadSenders();
		} catch (err: unknown) {
			window.alert(err instanceof Error ? err.message : 'Unable to assign sender ID.');
		} finally {
			assignSaving = false;
		}
	}

	async function toggleEnabled(sender: SenderItem) {
		const targetState = !sender.is_enabled;
		try {
			await supermanApi.patch(`/api/v1/superman/senders/${sender.id}/enabled`, {
				is_enabled: targetState
			});
			sender.is_enabled = targetState;
		} catch (err: unknown) {
			window.alert(err instanceof Error ? err.message : 'Unable to update sender state.');
		}
	}

	async function approveSender(sender: SenderItem) {
		if (!confirm(`Are you sure you want to approve "${sender.sender_id}"?`)) return;
		try {
			await supermanApi.post(`/api/v1/superman/senders/${sender.id}/approve`);
			sender.approval_status = 'approved';
			sender.is_enabled = true;
			window.alert(`Sender "${sender.sender_id}" approved.`);
		} catch (err: unknown) {
			window.alert(err instanceof Error ? err.message : 'Unable to approve sender.');
		}
	}

	async function rejectSender(sender: SenderItem) {
		const reason = prompt(`Enter rejection reason for "${sender.sender_id}":`);
		if (!reason) return;
		try {
			await supermanApi.post(`/api/v1/superman/senders/${sender.id}/reject`, { reason });
			sender.approval_status = 'rejected';
			sender.is_enabled = false;
			sender.rejection_reason = reason;
			window.alert(`Sender "${sender.sender_id}" rejected.`);
		} catch (err: unknown) {
			window.alert(err instanceof Error ? err.message : 'Unable to reject sender.');
		}
	}

	async function deleteSender(sender: SenderItem) {
		if (!confirm(`Delete sender ID "${sender.sender_id}"? This action can be undone by SuperAdmin.`)) return;
		try {
			await supermanApi.delete(`/api/v1/superman/senders/${sender.id}`);
			window.alert(`Sender "${sender.sender_id}" deleted.`);
			void loadSenders();
		} catch (err: unknown) {
			window.alert(err instanceof Error ? err.message : 'Unable to delete sender.');
		}
	}

	let filteredModalCustomers = $derived(
		allCustomers.filter((c) => {
			if (!assignCustomerSearch.trim()) return true;
			const q = assignCustomerSearch.toLowerCase();
			return c.name.toLowerCase().includes(q) || (c.email && c.email.toLowerCase().includes(q));
		})
	);

	onMount(() => {
		void loadOptions();
		void loadSenders();
	});

	const exportColumns: ExportColumn<SenderItem>[] = [
		{ header: 'Sender ID', getValue: (s) => s.sender_id },
		{ header: 'Type', getValue: (s) => s.type },
		{ header: 'Gateway', getValue: (s) => s.gateway_name || 'Unassigned' },
		{ header: 'Assigned Customers', getValue: (s) => (s.users && s.users.length > 0 ? s.users.map((u) => u.name).join(', ') : 'None') },
		{ header: 'Messages Sent', getValue: (s) => s.messages_count },
		{ header: 'Approval Status', getValue: (s) => s.approval_status.toUpperCase() },
		{ header: 'Operating State', getValue: (s) => (s.is_enabled ? 'Enabled' : 'Disabled') },
		{ header: 'Created Date', getValue: (s) => new Date(s.created_at).toLocaleDateString() }
	];
</script>

<svelte:head>
	<title>Masking Sender — SMS Whiteinfo Administrator</title>
</svelte:head>

<div class="space-y-6">
	<!-- Page Header -->
	<div class="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
		<div>
			<div class="flex items-center gap-2">
				<h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">Masking Sender IDs</h1>
				<span class="inline-flex items-center rounded-full bg-emerald-50 px-2.5 py-0.5 text-xs font-semibold text-emerald-700 ring-1 ring-inset ring-emerald-600/20 dark:bg-emerald-950/40 dark:text-emerald-300">
					Alphanumeric & Brand
				</span>
			</div>
			<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
				Brand and shortcode sender IDs only. Route channels and customer access control.
			</p>
		</div>

		<div class="flex items-center gap-2.5">
			<button
				type="button"
				onclick={() => loadSenders()}
				class="inline-flex items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs font-medium text-gray-700 shadow-2xs hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300 transition"
			>
				<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0 3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182m0-4.991v4.99" />
				</svg>
				Refresh
			</button>

			<ExportDropdown
				title="Masking Sender IDs"
				columns={exportColumns}
				getData={() => senders}
				disabled={senders.length === 0}
			/>

			<a
				href="/superman/sender-cli/add"
				class="inline-flex items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-3.5 py-2 text-xs font-semibold text-gray-700 shadow-2xs hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300 transition"
			>
				<svg class="h-4 w-4 text-gray-500" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
				</svg>
				Add Sender
			</a>

			<button
				type="button"
				onclick={() => openAssignmentModal()}
				class="inline-flex items-center gap-1.5 rounded-lg bg-emerald-600 px-3.5 py-2 text-xs font-semibold text-white shadow-xs hover:bg-emerald-500 transition"
			>
				<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" d="M18 18.72a9.094 9.094 0 0 0 3.741-.479 3 3 0 0 0-4.682-2.72m.94 3.198.001.031c0 .225-.012.447-.037.666A11.944 11.944 0 0 1 12 21c-2.17 0-4.207-.576-5.963-1.584A6.062 6.062 0 0 1 6 18.719m12 0a5.971 5.971 0 0 0-.941-3.197m0 0A5.995 5.995 0 0 0 12 12.75a5.995 5.995 0 0 0-5.058 2.772m0 0a3 3 0 0 0-4.681 2.72 8.986 8.986 0 0 0 3.74.477m.94-3.197a5.971 5.971 0 0 0-.94 3.197M15 6.75a3 3 0 1 1-6 0 3 3 0 0 1 6 0Zm6 3a2.25 2.25 0 1 1-4.5 0 2.25 2.25 0 0 1 4.5 0Zm-13.5 0a2.25 2.25 0 1 1-4.5 0 2.25 2.25 0 0 1 4.5 0Z" />
				</svg>
				Assign Sender ID
			</button>
		</div>
	</div>

	<!-- Search & Filter Card -->
	<div class="rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
		<div class="grid grid-cols-1 gap-3 sm:grid-cols-4">
			<div class="relative sm:col-span-2">
				<input
					type="text"
					bind:value={search}
					onkeydown={(e) => e.key === 'Enter' && void loadSenders(1)}
					placeholder="Search sender ID or gateway..."
					class="w-full rounded-lg border border-gray-200 bg-gray-50/70 py-2 pl-9 pr-3 text-xs outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
				/>
				<svg class="absolute left-3 top-2.5 h-4 w-4 text-gray-400" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z" />
				</svg>
			</div>

			<div>
				<select
					bind:value={gatewayFilter}
					onchange={() => void loadSenders(1)}
					class="w-full rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs text-gray-700 outline-none dark:border-gray-700 dark:bg-gray-800 dark:text-gray-300"
				>
					<option value="">All gateways</option>
					{#each gateways as gw}
						<option value={gw.id}>{gw.name}</option>
					{/each}
				</select>
			</div>

			<div>
				<select
					bind:value={assignmentFilter}
					onchange={() => void loadSenders(1)}
					class="w-full rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs text-gray-700 outline-none dark:border-gray-700 dark:bg-gray-800 dark:text-gray-300"
				>
					<option value="">All assignment states</option>
					<option value="used">Used / assigned</option>
					<option value="unused">Unused / unassigned</option>
				</select>
			</div>
		</div>
	</div>

	<!-- Main Inventory Table Card -->
	<div class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-xs dark:border-gray-800 dark:bg-gray-900">
		<div class="flex items-center justify-between border-b border-gray-100 px-5 py-4 dark:border-gray-800">
			<div>
				<h2 class="text-sm font-semibold text-gray-950 dark:text-white">Sender inventory</h2>
				<p class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">{total} masking sender IDs found</p>
			</div>
			<div class="flex items-center gap-2">
				<button
					type="button"
					onclick={() => void loadSenders(page)}
					class="inline-flex items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-3 py-1.5 text-xs font-semibold text-gray-700 hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-800 dark:text-gray-300 transition"
				>
					<svg class="h-3.5 w-3.5 {loading ? 'animate-spin' : ''}" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0 3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182m0-4.991v4.99" />
					</svg>
					Refresh
				</button>
			</div>
		</div>

		{#if error}
			<div class="p-5">
				<div class="rounded-lg border border-rose-200 bg-rose-50 p-4 text-xs text-rose-700 dark:border-rose-900 dark:bg-rose-950/30 dark:text-rose-300">
					{error}
				</div>
			</div>
		{:else}
			<div class="overflow-x-auto">
				<table class="w-full min-w-[900px] text-left text-xs">
					<thead>
						<tr class="border-b border-gray-100 bg-gray-50/70 text-[10px] font-semibold uppercase tracking-wider text-gray-400 dark:border-gray-800 dark:bg-gray-800/40">
							<th class="px-5 py-3.5">Sender ID</th>
							<th class="px-5 py-3.5">Type</th>
							<th class="px-5 py-3.5">Gateway</th>
							<th class="px-5 py-3.5">Customer Access</th>
							<th class="px-5 py-3.5 text-center">Messages</th>
							<th class="px-5 py-3.5">Availability</th>
							<th class="px-5 py-3.5 text-right">Actions</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
						{#if loading && senders.length === 0}
							<tr>
								<td colspan="7" class="px-5 py-12 text-center text-gray-400">Loading masking senders...</td>
							</tr>
						{:else if senders.length === 0}
							<tr>
								<td colspan="7" class="px-5 py-12 text-center">
									<p class="font-medium text-gray-700 dark:text-gray-300">No masking sender IDs found.</p>
									<p class="mt-1 text-xs text-gray-400">Try a different search or filter criteria.</p>
								</td>
							</tr>
						{:else}
							{#each senders as sender}
								<tr class="transition-colors hover:bg-gray-50/50 dark:hover:bg-gray-800/30">
									<td class="px-5 py-3.5">
										<span class="inline-flex rounded-lg bg-gray-100 px-2.5 py-1 font-mono text-xs font-bold tracking-tight text-gray-900 dark:bg-gray-800 dark:text-white">
											{sender.sender_id}
										</span>
									</td>
									<td class="px-5 py-3.5">
										<span class="inline-flex rounded-full bg-blue-50 px-2.5 py-0.5 text-[11px] font-semibold text-blue-700 ring-1 ring-inset ring-blue-600/20 dark:bg-blue-950/40 dark:text-blue-300">
											{sender.type === 'shortcode' ? 'Shortcode' : 'Brand ID'}
										</span>
									</td>
									<td class="px-5 py-3.5">
										<span class="font-medium text-gray-700 dark:text-gray-200">{sender.gateway_name ?? '—'}</span>
									</td>
									<td class="px-5 py-3.5">
										{#if sender.users_count === 0}
											<span class="inline-flex items-center gap-1.5 rounded-full bg-amber-50 px-2.5 py-0.5 text-[11px] font-semibold text-amber-700 ring-1 ring-inset ring-amber-600/20 dark:bg-amber-950/40 dark:text-amber-300">
												<span class="h-1.5 w-1.5 rounded-full bg-amber-500"></span>
												Unused
											</span>
										{:else}
											<div class="flex flex-wrap items-center gap-1.5">
												{#each (sender.users || []).slice(0, 2) as u}
													<span class="inline-flex max-w-[140px] truncate rounded-full bg-emerald-50 px-2.5 py-0.5 text-[11px] font-medium text-emerald-700 ring-1 ring-inset ring-emerald-600/20 dark:bg-emerald-950/40 dark:text-emerald-300">
														{u.name}
													</span>
												{/each}
												{#if sender.users_count > 2}
													<span class="inline-flex rounded-full bg-gray-100 px-2 py-0.5 text-[10px] font-semibold text-gray-600 dark:bg-gray-800 dark:text-gray-400">
														+{sender.users_count - 2} more
													</span>
												{/if}
											</div>
										{/if}
									</td>
									<td class="px-5 py-3.5 text-center font-mono font-medium text-gray-700 dark:text-gray-300">
										{sender.messages_count.toLocaleString()}
									</td>
									<td class="px-5 py-3.5">
										<button
											type="button"
											onclick={() => void toggleEnabled(sender)}
											class="inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-[11px] font-semibold ring-1 ring-inset transition {sender.is_enabled ? 'bg-emerald-50 text-emerald-700 ring-emerald-600/20 dark:bg-emerald-950/40 dark:text-emerald-300' : 'bg-gray-100 text-gray-500 ring-gray-400/20 dark:bg-gray-800 dark:text-gray-400'}"
										>
											<span class="h-1.5 w-1.5 rounded-full {sender.is_enabled ? 'bg-emerald-500' : 'bg-gray-400'}"></span>
											{sender.is_enabled ? 'Enabled' : 'Disabled'}
										</button>
									</td>
									<td class="px-5 py-3.5 text-right relative">
										<div class="inline-flex items-center gap-1">
											<button
												type="button"
												onclick={() => openAssignmentModal(sender)}
												class="rounded-lg border border-gray-200 bg-white px-2.5 py-1 text-xs font-semibold text-gray-700 hover:bg-gray-50 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-300 transition"
											>
												Assign
											</button>

											<div class="relative">
												<button
													type="button"
													onclick={() => (activeMenuId = activeMenuId === sender.id ? null : sender.id)}
													class="rounded-lg p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-gray-800 dark:hover:text-gray-300"
												>
													<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
														<path stroke-linecap="round" stroke-linejoin="round" d="M12 6.75a.75.75 0 1 1 0-1.5.75.75 0 0 1 0 1.5ZM12 12.75a.75.75 0 1 1 0-1.5.75.75 0 0 1 0 1.5ZM12 18.75a.75.75 0 1 1 0-1.5.75.75 0 0 1 0 1.5Z" />
													</svg>
												</button>

												{#if activeMenuId === sender.id}
													<div class="absolute right-0 top-full z-20 mt-1 w-36 rounded-lg border border-gray-200 bg-white p-1 shadow-lg dark:border-gray-800 dark:bg-gray-900">
														{#if sender.approval_status === 'pending'}
															<button
																type="button"
																onclick={() => { activeMenuId = null; void approveSender(sender); }}
																class="flex w-full items-center gap-2 rounded px-2.5 py-1.5 text-xs text-emerald-600 hover:bg-emerald-50 dark:hover:bg-emerald-950/30"
															>
																Approve
															</button>
															<button
																type="button"
																onclick={() => { activeMenuId = null; void rejectSender(sender); }}
																class="flex w-full items-center gap-2 rounded px-2.5 py-1.5 text-xs text-rose-600 hover:bg-rose-50 dark:hover:bg-rose-950/30"
															>
																Reject
															</button>
														{/if}
														<a
															href="/superman/sender-cli/resolution?sender={encodeURIComponent(sender.sender_id)}"
															class="flex w-full items-center gap-2 rounded px-2.5 py-1.5 text-xs text-blue-600 hover:bg-blue-50 dark:hover:bg-blue-950/30"
														>
															Test Route
														</a>
														<button
															type="button"
															onclick={() => { activeMenuId = null; void deleteSender(sender); }}
															class="flex w-full items-center gap-2 rounded px-2.5 py-1.5 text-xs text-gray-600 hover:bg-gray-50 dark:text-gray-400 dark:hover:bg-gray-800"
														>
															Delete
														</button>
													</div>
												{/if}
											</div>
										</div>
									</td>
								</tr>
							{/each}
						{/if}
					</tbody>
				</table>
			</div>

			<!-- Pagination footer -->
			{#if totalPages > 1}
				<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between border-t border-gray-100 px-5 py-3 dark:border-gray-800 gap-3">
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
		{/if}
	</div>

	<!-- Modal: Assign Sender ID -->
	{#if showAssignModal}
		<div class="fixed inset-0 z-50 flex items-center justify-center bg-gray-950/60 p-4 backdrop-blur-xs">
			<div class="w-full max-w-lg rounded-2xl border border-gray-200 bg-white p-6 shadow-xl dark:border-gray-800 dark:bg-gray-900 space-y-4">
				<div class="flex items-center justify-between border-b border-gray-100 pb-3 dark:border-gray-800">
					<div>
						<h3 class="text-base font-bold text-gray-950 dark:text-white">Assign Sender ID</h3>
						<p class="text-xs text-gray-500 dark:text-gray-400">Choose a sender ID and grant access to customer accounts.</p>
					</div>
					<button
						type="button"
						onclick={() => (showAssignModal = false)}
						class="rounded-lg p-1 text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800"
					>
						<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
						</svg>
					</button>
				</div>

				<div class="space-y-4">
					<div>
						<label class="block text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">Sender ID</label>
						<select
							bind:value={assignSenderId}
							class="mt-1 w-full rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs font-mono font-medium text-gray-900 outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
						>
							{#each availableSendersList as s}
								<option value={s.id}>{s.sender_id}</option>
							{/each}
						</select>
					</div>

					<div>
						<div class="flex items-center justify-between">
							<label class="block text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">Customer Accounts</label>
							<span class="text-xs text-emerald-600 font-medium">{assignCustomerIds.length} selected</span>
						</div>
						<input
							type="text"
							bind:value={assignCustomerSearch}
							placeholder="Search customer by name or email..."
							class="mt-1 w-full rounded-lg border border-gray-200 bg-gray-50/70 px-3 py-1.5 text-xs outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
						/>
						<div class="mt-2 max-h-52 overflow-y-auto rounded-lg border border-gray-200 divide-y divide-gray-100 dark:border-gray-800 dark:divide-gray-800 p-1">
							{#each filteredModalCustomers as customer}
								<label class="flex items-center gap-2.5 p-2 rounded hover:bg-gray-50 dark:hover:bg-gray-800/50 cursor-pointer">
									<input
										type="checkbox"
										checked={assignCustomerIds.includes(customer.id)}
										onchange={() => toggleCustomerSelection(customer.id)}
										class="h-4 w-4 rounded border-gray-300 text-emerald-600 focus:ring-emerald-500"
									/>
									<div class="flex-1 min-w-0">
										<p class="text-xs font-medium text-gray-900 dark:text-white truncate">{customer.name}</p>
										{#if customer.email}
											<p class="text-[11px] text-gray-400 truncate">{customer.email}</p>
										{/if}
									</div>
								</label>
							{/each}
						</div>
					</div>
				</div>

				<div class="flex items-center justify-end gap-2 border-t border-gray-100 pt-3 dark:border-gray-800">
					<button
						type="button"
						onclick={() => (showAssignModal = false)}
						class="rounded-lg border border-gray-200 px-3.5 py-2 text-xs font-semibold text-gray-700 hover:bg-gray-50 dark:border-gray-700 dark:text-gray-300"
					>
						Cancel
					</button>
					<button
						type="button"
						disabled={assignSaving}
						onclick={() => void saveAssignment()}
						class="inline-flex items-center gap-1.5 rounded-lg bg-emerald-600 px-4 py-2 text-xs font-semibold text-white shadow-xs hover:bg-emerald-500 disabled:opacity-50"
					>
						{#if assignSaving}
							<svg class="h-3.5 w-3.5 animate-spin" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0 3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182m0-4.991v4.99" />
							</svg>
						{/if}
						Save Assignment
					</button>
				</div>
			</div>
		</div>
	{/if}
</div>

