<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface KYCRequest {
		id: number;
		customer_name: string;
		email: string;
		phone: string;
		nid_number: string | null;
		business_name: string | null;
		trade_license: string | null;
		status: string;
		submitted_at: string | null;
	}

	interface KYCDetail {
		request: KYCRequest;
		business_address: string | null;
		front_nid: string | null;
		back_nid: string | null;
		live_photo: string | null;
		trade_license_file: string | null;
		rejection_reason: string | null;
		resubmit_fields: string | null;
		kyc_verified_at: string | null;
	}

	let requests = $state<KYCRequest[]>([]);
	let loading = $state(false);
	let error = $state<string | null>(null);

	// Tabs: 'pending' (default queue), 'under_review', 'resubmit_required', 'all'
	let activeTab = $state<'pending' | 'under_review' | 'resubmit_required' | 'all'>('pending');
	let search = $state('');

	// Inspection Modal state
	let detailModalOpen = $state(false);
	let detailLoading = $state(false);
	let activeDetail = $state<KYCDetail | null>(null);
	let previewImage = $state<string | null>(null);

	let filteredRequests = $derived(
		requests.filter((r) => {
			if (!search.trim()) return true;
			const q = search.toLowerCase();
			return (
				r.customer_name.toLowerCase().includes(q) ||
				r.email.toLowerCase().includes(q) ||
				r.phone.toLowerCase().includes(q) ||
				(r.nid_number && r.nid_number.toLowerCase().includes(q)) ||
				(r.business_name && r.business_name.toLowerCase().includes(q))
			);
		})
	);

	async function loadRequests() {
		loading = true;
		error = null;
		try {
			const queryParam = activeTab === 'all' ? '?status=all' : activeTab === 'pending' ? '' : `?status=${activeTab}`;
			const response = await supermanApi.get<{ data: KYCRequest[] }>(`/api/v1/superman/kyc/requests${queryParam}`);
			requests = response.data;
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Unable to load KYC requests.';
		} finally {
			loading = false;
		}
	}

	function switchTab(tab: 'pending' | 'under_review' | 'resubmit_required' | 'all') {
		activeTab = tab;
		void loadRequests();
	}

	async function inspectRequest(r: KYCRequest) {
		detailModalOpen = true;
		detailLoading = true;
		activeDetail = null;
		previewImage = null;
		try {
			const res = await supermanApi.get<{ data: KYCDetail }>(`/api/v1/superman/kyc/requests/${r.id}`);
			activeDetail = res.data;
		} catch {
			// Fallback: try customer endpoint
			try {
				const resCust = await supermanApi.get<{ data: any }>(`/api/v1/superman/customers/${r.id}`);
				activeDetail = {
					request: r,
					business_address: resCust.data.address ?? null,
					front_nid: resCust.data.front_nid ?? null,
					back_nid: resCust.data.back_nid ?? null,
					live_photo: resCust.data.live_photo ?? null,
					trade_license_file: resCust.data.trade_license ?? null,
					rejection_reason: null,
					resubmit_fields: null,
					kyc_verified_at: null
				};
			} catch (inspectError: unknown) {
				alert(inspectError instanceof Error ? inspectError.message : 'Unable to inspect KYC details.');
				detailModalOpen = false;
			}
		} finally {
			detailLoading = false;
		}
	}

	async function handleApprove(id: number, name: string) {
		if (!confirm(`Are you sure you want to approve KYC for ${name}?`)) return;
		try {
			const response = await supermanApi.post<{ message: string }>(`/api/v1/superman/kyc/requests/${id}/approve`);
			detailModalOpen = false;
			await loadRequests();
			alert(response.message || `KYC for ${name} has been approved.`);
		} catch (err: unknown) {
			alert(err instanceof Error ? err.message : 'Unable to approve this KYC request.');
		}
	}

	async function handleReject(id: number, name: string) {
		const reason = prompt('Please provide reason for KYC rejection:');
		if (reason) {
			try {
				const response = await supermanApi.post<{ message: string }>(`/api/v1/superman/kyc/requests/${id}/reject`, { reason });
				detailModalOpen = false;
				await loadRequests();
				alert(response.message || `KYC for ${name} rejected.`);
			} catch (err: unknown) {
				alert(err instanceof Error ? err.message : 'Unable to reject this KYC request.');
			}
		}
	}

	async function handleResubmit(id: number) {
		const reason = prompt('Specify what corrections are needed:');
		if (!reason) return;
		const fields = prompt('Which documents need re-uploading? (e.g. Front NID, Live Photo):') ?? 'All';
		try {
			const response = await supermanApi.post<{ message: string }>(`/api/v1/superman/kyc/requests/${id}/resubmit`, {
				reason,
				fields
			});
			detailModalOpen = false;
			await loadRequests();
			alert(response.message || 'Resubmission request sent to customer.');
		} catch (err: unknown) {
			alert(err instanceof Error ? err.message : 'Unable to request resubmission.');
		}
	}

	function formatDateTime(str: string | null): string {
		if (!str) return '—';
		const d = new Date(str);
		if (isNaN(d.getTime())) return str;
		return d.toLocaleDateString('en-US', {
			month: 'short',
			day: '2-digit',
			year: 'numeric'
		}) + ' ' + d.toTimeString().slice(0, 8);
	}

	onMount(() => {
		void loadRequests();
	});

	const exportColumns: ExportColumn<KYCRequest>[] = [
		{ header: 'Customer Name', getValue: (r) => r.customer_name },
		{ header: 'Email', getValue: (r) => r.email },
		{ header: 'Phone', getValue: (r) => r.phone },
		{ header: 'NID Number', getValue: (r) => r.nid_number || '—' },
		{ header: 'Business Name', getValue: (r) => r.business_name || '—' },
		{ header: 'Trade License', getValue: (r) => r.trade_license || '—' },
		{ header: 'Status', getValue: (r) => r.status.toUpperCase() },
		{ header: 'Submitted At', getValue: (r) => (r.submitted_at ? new Date(r.submitted_at).toLocaleString() : '—') }
	];
</script>

<svelte:head>
	<title>KYC Requests — SMS Whiteinfo Administrator</title>
</svelte:head>

<div class="space-y-4">
	<!-- Page Header matching Screenshot 4 -->
	<div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">KYC Verification Requests</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400">Review National ID, trade license, and live facial verification files.</p>
		</div>
		<div class="flex items-center gap-2">
			<button
				type="button"
				onclick={() => void loadRequests()}
				class="inline-flex items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-3 py-1.5 text-xs font-semibold text-gray-700 shadow-2xs hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300 transition"
			>
				<svg class="h-3.5 w-3.5 {loading ? 'animate-spin' : ''}" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0 3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182m0-4.991v4.99" />
				</svg>
				Refresh
			</button>
			<ExportDropdown
				title="KYC Requests"
				columns={exportColumns}
				getData={() => filteredRequests}
				disabled={filteredRequests.length === 0}
			/>
			<span class="rounded-full bg-purple-100 px-3 py-1 text-xs font-bold text-purple-800 dark:bg-purple-950 dark:text-purple-300">
				{requests.length} {activeTab === 'pending' ? 'Pending Review' : 'Records'}
			</span>
		</div>
	</div>

	<!-- Queue Navigation Tabs -->
	<div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between border-b border-gray-200 pb-2 dark:border-gray-800">
		<div class="flex items-center gap-1.5">
			<button
				type="button"
				onclick={() => switchTab('pending')}
				class="rounded-lg px-3 py-1.5 text-xs font-semibold transition {activeTab === 'pending' ? 'bg-purple-600 text-white shadow-xs' : 'text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-800'}"
			>
				Review Queue
			</button>
			<button
				type="button"
				onclick={() => switchTab('under_review')}
				class="rounded-lg px-3 py-1.5 text-xs font-semibold transition {activeTab === 'under_review' ? 'bg-purple-600 text-white shadow-xs' : 'text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-800'}"
			>
				Under Review
			</button>
			<button
				type="button"
				onclick={() => switchTab('resubmit_required')}
				class="rounded-lg px-3 py-1.5 text-xs font-semibold transition {activeTab === 'resubmit_required' ? 'bg-purple-600 text-white shadow-xs' : 'text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-800'}"
			>
				Resubmit Required
			</button>
			<button
				type="button"
				onclick={() => switchTab('all')}
				class="rounded-lg px-3 py-1.5 text-xs font-semibold transition {activeTab === 'all' ? 'bg-purple-600 text-white shadow-xs' : 'text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-800'}"
			>
				All Submissions
			</button>
		</div>

		<!-- Search filter -->
		<div class="relative w-full sm:w-64">
			<svg class="absolute left-3 top-2.5 h-3.5 w-3.5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
				<path stroke-linecap="round" stroke-linejoin="round" d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z" />
			</svg>
			<input
				type="search"
				bind:value={search}
				placeholder="Search requests..."
				class="w-full rounded-lg border border-gray-200 bg-white py-1.5 pl-8 pr-3 text-xs outline-none focus:border-purple-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white transition"
			/>
		</div>
	</div>

	<!-- Requests Container -->
	<div class="rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900 space-y-4">
		{#if error}
			<div class="p-6 text-center text-xs text-rose-600">{error}</div>
		{:else if filteredRequests.length === 0}
			<div class="py-16 text-center text-gray-400 text-xs space-y-3">
				<div class="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-emerald-50 text-emerald-600 dark:bg-emerald-950/40 dark:text-emerald-400">
					<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" />
					</svg>
				</div>
				<div>
					<p class="font-bold text-gray-800 dark:text-gray-200 text-sm">All KYC requests have been processed. Queue is clear!</p>
					<p class="text-[11px] text-gray-500 mt-0.5">There are no pending submissions requiring administrator verification right now.</p>
				</div>
				<div class="pt-2">
					<button
						type="button"
						onclick={() => switchTab('all')}
						class="rounded-lg bg-gray-100 hover:bg-gray-200 px-3.5 py-1.5 text-xs font-semibold text-gray-700 dark:bg-gray-800 dark:text-gray-300 dark:hover:bg-gray-700 transition"
					>
						View All Customers &amp; Submissions →
					</button>
				</div>
			</div>
		{:else}
			<div class="overflow-x-auto">
				<table class="w-full text-left text-xs">
					<thead>
						<tr class="border-b border-gray-100 dark:border-gray-800 text-gray-400 uppercase text-[10px] font-semibold">
							<th class="py-2.5 px-3">Customer</th>
							<th class="py-2.5 px-3">Business &amp; Trade License</th>
							<th class="py-2.5 px-3">NID Number</th>
							<th class="py-2.5 px-3">Status</th>
							<th class="py-2.5 px-3">Documents</th>
							<th class="py-2.5 px-3">Submitted</th>
							<th class="py-2.5 px-3 text-right">Decision</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
						{#each filteredRequests as r}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30 transition-colors">
								<td class="py-3 px-3">
									<div class="font-semibold text-gray-900 dark:text-white">{r.customer_name}</div>
									<div class="text-[11px] text-gray-400">{r.email} • {r.phone}</div>
								</td>
								<td class="py-3 px-3">
									<div class="font-medium text-gray-800 dark:text-gray-200">{r.business_name ?? '—'}</div>
									<div class="text-[11px] font-mono text-gray-400">{r.trade_license ?? '—'}</div>
								</td>
								<td class="py-3 px-3 font-mono font-medium text-gray-700 dark:text-gray-300">
									{r.nid_number ?? '—'}
								</td>
								<td class="py-3 px-3">
									{#if r.status === 'verified'}
										<span class="inline-flex items-center gap-1 rounded-full border border-emerald-300 bg-emerald-50 px-2 py-0.5 text-[10px] font-bold text-emerald-700 dark:border-emerald-800 dark:bg-emerald-950/40 dark:text-emerald-300">
											Verified
										</span>
									{:else if r.status === 'under_review'}
										<span class="inline-flex items-center gap-1 rounded-full border border-amber-300 bg-amber-50 px-2 py-0.5 text-[10px] font-bold text-amber-700 dark:border-amber-800 dark:bg-amber-950/40 dark:text-amber-300">
											Under Review
										</span>
									{:else if r.status === 'resubmit_required'}
										<span class="inline-flex items-center gap-1 rounded-full border border-blue-300 bg-blue-50 px-2 py-0.5 text-[10px] font-bold text-blue-700 dark:border-blue-800 dark:bg-blue-950/40 dark:text-blue-300">
											Resubmit
										</span>
									{:else if r.status === 'rejected'}
										<span class="inline-flex items-center gap-1 rounded-full border border-rose-300 bg-rose-50 px-2 py-0.5 text-[10px] font-bold text-rose-700 dark:border-rose-800 dark:bg-rose-950/40 dark:text-rose-300">
											Rejected
										</span>
									{:else}
										<span class="inline-flex items-center gap-1 rounded-full border border-gray-300 bg-gray-50 px-2 py-0.5 text-[10px] font-bold text-gray-700 dark:border-gray-700 dark:bg-gray-800/40 dark:text-gray-300">
											Pending
										</span>
									{/if}
								</td>
								<td class="py-3 px-3">
									<button
										type="button"
										onclick={() => void inspectRequest(r)}
										class="flex items-center gap-1.5 text-[11px] cursor-pointer hover:opacity-80"
										title="Click to view uploaded documents"
									>
										<span class="rounded bg-purple-50 border border-purple-200 px-2 py-0.5 text-purple-700 dark:bg-purple-950/40 dark:border-purple-800 dark:text-purple-300 font-semibold">Inspect Docs →</span>
									</button>
								</td>
								<td class="py-3 px-3 text-[11px] text-gray-400 font-mono">
									{formatDateTime(r.submitted_at)}
								</td>
								<td class="py-3 px-3 text-right">
									<div class="inline-flex items-center gap-1.5">
										<button
											type="button"
											onclick={() => void inspectRequest(r)}
											class="rounded-lg bg-gray-100 px-2.5 py-1 text-xs font-semibold text-gray-700 hover:bg-gray-200 dark:bg-gray-800 dark:text-gray-300 transition"
										>
											Inspect
										</button>
										{#if r.status !== 'verified'}
											<button
												type="button"
												onclick={() => void handleApprove(r.id, r.customer_name)}
												class="rounded-lg bg-emerald-600 px-2.5 py-1 text-xs font-semibold text-white shadow-xs hover:bg-emerald-700 transition"
											>
												Approve
											</button>
										{/if}
										<button
											type="button"
											onclick={() => void handleReject(r.id, r.customer_name)}
											class="rounded-lg border border-rose-200 px-2.5 py-1 text-xs font-semibold text-rose-600 hover:bg-rose-50 dark:border-rose-800 dark:text-rose-400 transition"
										>
											Reject
										</button>
									</div>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</div>
</div>

<!-- Inspect KYC Modal -->
{#if detailModalOpen}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
		<div class="w-full max-w-2xl max-h-[90vh] overflow-y-auto rounded-2xl bg-white p-6 shadow-2xl dark:bg-gray-900 border border-gray-100 dark:border-gray-800 space-y-4">
			<div class="flex items-center justify-between border-b border-gray-100 pb-3 dark:border-gray-800">
				<div>
					<h3 class="text-base font-bold text-gray-900 dark:text-white">KYC Verification File Inspector</h3>
					<p class="text-xs text-gray-500">Examine uploaded credentials and make approval decision.</p>
				</div>
				<button onclick={() => (detailModalOpen = false)} class="text-gray-400 hover:text-gray-600 text-sm">✕</button>
			</div>

			{#if detailLoading}
				<div class="py-12 text-center text-xs text-gray-400">Loading document details…</div>
			{:else if activeDetail}
				<div class="grid grid-cols-1 sm:grid-cols-2 gap-4 text-xs">
					<div class="space-y-1">
						<span class="text-gray-400 block font-medium">Customer</span>
						<span class="font-semibold text-gray-900 dark:text-white">{activeDetail.request.customer_name}</span>
						<div class="text-[11px] text-gray-500">{activeDetail.request.email} • {activeDetail.request.phone}</div>
					</div>
					<div class="space-y-1">
						<span class="text-gray-400 block font-medium">NID / Smart Card</span>
						<span class="font-mono font-bold text-gray-800 dark:text-gray-200">{activeDetail.request.nid_number ?? '—'}</span>
					</div>
					<div class="space-y-1">
						<span class="text-gray-400 block font-medium">Business / Trade License</span>
						<span class="font-semibold text-gray-900 dark:text-white">{activeDetail.request.business_name ?? '—'}</span>
						<div class="font-mono text-[11px] text-gray-500">{activeDetail.request.trade_license ?? '—'}</div>
					</div>
					<div class="space-y-1">
						<span class="text-gray-400 block font-medium">Business Address</span>
						<span class="text-gray-700 dark:text-gray-300">{activeDetail.business_address ?? '—'}</span>
					</div>
				</div>

				<div class="border-t border-gray-100 pt-3 dark:border-gray-800">
					<h4 class="text-xs font-bold text-gray-900 dark:text-white mb-2 uppercase tracking-wider text-[10px] text-gray-400">Uploaded Verification Files</h4>
					<div class="grid grid-cols-1 sm:grid-cols-2 gap-3 text-xs">
						<!-- Front NID -->
						<div class="rounded-xl border border-gray-200 p-3 bg-gray-50/50 dark:border-gray-800 dark:bg-gray-800/40 space-y-1.5">
							<span class="font-semibold text-gray-900 dark:text-white text-[11px]">Front National ID</span>
							{#if activeDetail.front_nid}
								<img src={activeDetail.front_nid} alt="Front NID" class="h-32 w-full object-cover rounded-lg border border-gray-200 dark:border-gray-700 cursor-pointer" onclick={() => (previewImage = activeDetail!.front_nid)} />
								<span class="inline-block rounded bg-emerald-100 text-emerald-800 px-2 py-0.5 text-[10px] font-bold">Document Available</span>
							{:else}
								<span class="text-rose-500 font-medium text-[11px]">Not Uploaded</span>
							{/if}
						</div>

						<!-- Back NID -->
						<div class="rounded-xl border border-gray-200 p-3 bg-gray-50/50 dark:border-gray-800 dark:bg-gray-800/40 space-y-1.5">
							<span class="font-semibold text-gray-900 dark:text-white text-[11px]">Back National ID</span>
							{#if activeDetail.back_nid}
								<img src={activeDetail.back_nid} alt="Back NID" class="h-32 w-full object-cover rounded-lg border border-gray-200 dark:border-gray-700 cursor-pointer" onclick={() => (previewImage = activeDetail!.back_nid)} />
								<span class="inline-block rounded bg-emerald-100 text-emerald-800 px-2 py-0.5 text-[10px] font-bold">Document Available</span>
							{:else}
								<span class="text-rose-500 font-medium text-[11px]">Not Uploaded</span>
							{/if}
						</div>

						<!-- Live Facial Photo -->
						<div class="rounded-xl border border-gray-200 p-3 bg-gray-50/50 dark:border-gray-800 dark:bg-gray-800/40 space-y-1.5">
							<span class="font-semibold text-gray-900 dark:text-white text-[11px]">Live Facial Photo</span>
							{#if activeDetail.live_photo}
								<img src={activeDetail.live_photo} alt="Live Facial Photo" class="h-32 w-full object-cover rounded-lg border border-gray-200 dark:border-gray-700 cursor-pointer" onclick={() => (previewImage = activeDetail!.live_photo)} />
								<span class="inline-block rounded bg-emerald-100 text-emerald-800 px-2 py-0.5 text-[10px] font-bold">Photo Captured</span>
							{:else}
								<span class="text-rose-500 font-medium text-[11px]">Not Captured</span>
							{/if}
						</div>

						<!-- Trade License -->
						<div class="rounded-xl border border-gray-200 p-3 bg-gray-50/50 dark:border-gray-800 dark:bg-gray-800/40 space-y-1.5">
							<span class="font-semibold text-gray-900 dark:text-white text-[11px]">Trade License File</span>
							{#if activeDetail.trade_license_file}
								<img src={activeDetail.trade_license_file} alt="Trade License" class="h-32 w-full object-cover rounded-lg border border-gray-200 dark:border-gray-700 cursor-pointer" onclick={() => (previewImage = activeDetail!.trade_license_file)} />
								<span class="inline-block rounded bg-emerald-100 text-emerald-800 px-2 py-0.5 text-[10px] font-bold">Document Available</span>
							{:else}
								<span class="text-gray-400 font-medium text-[11px]">Optional / None</span>
							{/if}
						</div>
					</div>
				</div>

				<div class="flex flex-wrap items-center justify-between gap-2 border-t border-gray-100 pt-3 dark:border-gray-800">
					<button
						type="button"
						onclick={() => void handleResubmit(activeDetail!.request.id)}
						class="rounded-lg bg-amber-500 hover:bg-amber-600 px-3 py-1.5 text-xs font-bold text-white transition"
					>
						Request Resubmission
					</button>

					<div class="flex items-center gap-2">
						<button
							type="button"
							onclick={() => void handleReject(activeDetail!.request.id, activeDetail!.request.customer_name)}
							class="rounded-lg border border-rose-300 text-rose-700 hover:bg-rose-50 dark:border-rose-800 dark:text-rose-300 px-3 py-1.5 text-xs font-bold transition"
						>
							Reject
						</button>
						<button
							type="button"
							onclick={() => void handleApprove(activeDetail!.request.id, activeDetail!.request.customer_name)}
							class="rounded-lg bg-emerald-600 hover:bg-emerald-700 px-4 py-1.5 text-xs font-bold text-white shadow-xs transition"
						>
							Approve KYC
						</button>
					</div>
				</div>
			{/if}
		</div>
	</div>
{/if}

<!-- Zoom Image Lightbox Modal -->
{#if previewImage}
	<div class="fixed inset-0 z-60 flex items-center justify-center bg-black/80 p-4" onclick={() => (previewImage = null)}>
		<div class="relative max-h-[90vh] max-w-[90vw] overflow-hidden rounded-xl bg-white p-2 dark:bg-gray-900" onclick={(e) => e.stopPropagation()}>
			<img src={previewImage} alt="Document Preview" class="max-h-[85vh] max-w-[85vw] object-contain rounded" />
			<button onclick={() => (previewImage = null)} class="absolute top-4 right-4 rounded-full bg-black/60 p-2 text-white hover:bg-black">✕</button>
		</div>
	</div>
{/if}
