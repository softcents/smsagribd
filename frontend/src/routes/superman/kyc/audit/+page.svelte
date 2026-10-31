<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface KYCAuditRecord {
		id: number;
		admin: string;
		customer_id: number;
		customer_name: string;
		customer_email: string;
		action: string;
		field: string;
		old_value: string;
		new_value: string;
		note: string | null;
		ip_address: string;
		created_at: string;
	}

	interface AuditResponse {
		data: KYCAuditRecord[];
		meta: {
			page: number;
			per_page: number;
			total: number;
			total_pages: number;
		};
	}

	let auditLogs = $state<KYCAuditRecord[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);

	let search = $state('');
	let currentPage = $state(1);
	let perPage = $state(10);
	let totalRecords = $state(0);
	let totalPages = $state(1);

	let searchTimeout: ReturnType<typeof setTimeout> | null = null;

	function handleSearchInput() {
		if (searchTimeout) clearTimeout(searchTimeout);
		searchTimeout = setTimeout(() => {
			currentPage = 1;
			void loadAuditLogs();
		}, 300);
	}

	async function loadAuditLogs() {
		loading = true;
		error = null;
		try {
			const queryParams = new URLSearchParams({
				page: currentPage.toString(),
				per_page: perPage.toString()
			});
			if (search.trim()) {
				queryParams.set('search', search.trim());
			}

			const response = await supermanApi.get<AuditResponse>(`/api/v1/superman/kyc/audit?${queryParams.toString()}`);
			auditLogs = response.data;
			totalRecords = response.meta.total;
			totalPages = response.meta.total_pages;
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Unable to load KYC audit logs.';
		} finally {
			loading = false;
		}
	}

	function goToPage(page: number) {
		if (page < 1 || page > totalPages || page === currentPage) return;
		currentPage = page;
		void loadAuditLogs();
	}

	function handlePerPageChange(event: Event) {
		const target = event.target as HTMLSelectElement;
		perPage = Number(target.value);
		currentPage = 1;
		void loadAuditLogs();
	}

	function formatDateTime(str: string): string {
		const d = new Date(str);
		if (isNaN(d.getTime())) return str;
		return d.toLocaleDateString('en-US', {
			month: 'short',
			day: '2-digit',
			year: 'numeric'
		}) + ' ' + d.toTimeString().slice(0, 8);
	}

	onMount(() => {
		void loadAuditLogs();
	});

	const exportColumns: ExportColumn<KYCAuditRecord>[] = [
		{ header: 'Administrator', getValue: (l) => l.admin },
		{ header: 'Customer Name', getValue: (l) => l.customer_name },
		{ header: 'Customer Email', getValue: (l) => l.customer_email },
		{ header: 'Action', getValue: (l) => l.action.toUpperCase() },
		{ header: 'Field Modified', getValue: (l) => l.field },
		{ header: 'Old Value', getValue: (l) => l.old_value || '—' },
		{ header: 'New Value', getValue: (l) => l.new_value || '—' },
		{ header: 'Note', getValue: (l) => l.note || '—' },
		{ header: 'IP Address', getValue: (l) => l.ip_address },
		{ header: 'Date & Time', getValue: (l) => formatDateTime(l.created_at) }
	];

	async function getAllAuditLogsForExport(): Promise<KYCAuditRecord[]> {
		try {
			const queryParams = new URLSearchParams({
				page: '1',
				per_page: '10000'
			});
			if (search.trim()) {
				queryParams.set('search', search.trim());
			}
			const response = await supermanApi.get<AuditResponse>(`/api/v1/superman/kyc/audit?${queryParams.toString()}`);
			return response.data ?? auditLogs;
		} catch {
			return auditLogs;
		}
	}
</script>

<svelte:head>
	<title>KYC Audit Log — SMS Whiteinfo Administrator</title>
</svelte:head>

<div class="space-y-5">
	<!-- Page Header -->
	<div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">KYC Audit Log</h1>
		</div>
		<div class="flex items-center gap-2">
			<ExportDropdown
				title="KYC Audit Log"
				columns={exportColumns}
				getData={getAllAuditLogsForExport}
				disabled={totalRecords === 0 && auditLogs.length === 0}
			/>
			<button
				type="button"
				onclick={() => void loadAuditLogs()}
				class="inline-flex items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-3 py-1.5 text-xs font-semibold text-gray-700 shadow-2xs hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300 transition"
			>
				<svg class="h-3.5 w-3.5 {loading ? 'animate-spin' : ''}" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0 3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182m0-4.991v4.99" />
				</svg>
				Refresh
			</button>
		</div>
	</div>

	<!-- Main Table Container -->
	<div class="rounded-xl border border-gray-200 bg-white shadow-xs dark:border-gray-800 dark:bg-gray-900">
		<!-- Toolbar with Right-Aligned Search -->
		<div class="flex flex-col gap-3 p-4 sm:flex-row sm:items-center sm:justify-between border-b border-gray-100 dark:border-gray-800">
			<div></div>
			<div class="flex items-center justify-end gap-2.5">
				<div class="relative w-full max-w-xs">
					<svg class="absolute left-3 top-2.5 h-3.5 w-3.5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z" />
					</svg>
					<input
						type="search"
						bind:value={search}
						oninput={handleSearchInput}
						placeholder="Search"
						class="w-full rounded-lg border border-gray-200 bg-gray-50/50 py-1.5 pl-8 pr-3 text-xs outline-none focus:border-emerald-500 focus:bg-white dark:border-gray-700 dark:bg-gray-800 dark:text-white transition"
					/>
				</div>
				<!-- Filter icon button -->
				<button
					type="button"
					onclick={() => void loadAuditLogs()}
					class="rounded-lg border border-gray-200 bg-white p-2 text-gray-500 hover:bg-gray-50 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-400 transition"
					title="Filter"
				>
					<svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" d="M12 3c2.755 0 5.455.232 8.083.678.533.09.917.556.917 1.096v1.044a2.25 2.25 0 0 1-.659 1.591l-5.432 5.432a2.25 2.25 0 0 0-.659 1.591v2.927a2.25 2.25 0 0 1-1.244 2.013L9.75 21v-6.568a2.25 2.25 0 0 0-.659-1.591L3.659 7.409A2.25 2.25 0 0 1 3 5.818V4.774c0-.54.384-1.006.917-1.096A48.32 48.32 0 0 1 12 3Z" />
					</svg>
				</button>
			</div>
		</div>

		<!-- Table matching Screenshot 5 -->
		{#if error}
			<div class="p-6 text-center text-xs text-rose-600">{error}</div>
		{:else}
			<div class="overflow-x-auto">
				<table class="w-full text-left text-xs">
					<thead>
						<tr class="border-b border-gray-100 text-[11px] font-bold text-gray-600 uppercase dark:border-gray-800 dark:text-gray-400">
							<th class="py-3 px-4">ADMIN</th>
							<th class="py-3 px-4">CUSTOMER</th>
							<th class="py-3 px-4">ACTION</th>
							<th class="py-3 px-4">FIELD</th>
							<th class="py-3 px-4">OLD VALUE</th>
							<th class="py-3 px-4">NEW VALUE</th>
							<th class="py-3 px-4 text-right">DATE &amp; TIME</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
						{#if auditLogs.length === 0}
							<tr>
								<td colspan="7" class="py-12 text-center text-gray-400 text-xs">
									No KYC audit log records found.
								</td>
							</tr>
						{:else}
							{#each auditLogs as log}
								<tr class="hover:bg-gray-50/70 dark:hover:bg-gray-800/40 transition-colors">
									<td class="py-3.5 px-4 font-bold text-gray-900 dark:text-white">
										{log.admin}
									</td>
									<td class="py-3.5 px-4">
										<div class="font-bold text-gray-900 dark:text-white">{log.customer_name}</div>
										<div class="text-[11px] text-gray-500 dark:text-gray-400">{log.customer_email}</div>
									</td>
									<td class="py-3.5 px-4">
										{#if log.action === 'Approved'}
											<span class="inline-flex items-center rounded-full border border-emerald-200 bg-emerald-50 px-2.5 py-0.5 text-[11px] font-semibold text-emerald-700 dark:border-emerald-800 dark:bg-emerald-950/40 dark:text-emerald-300">
												Approved
											</span>
										{:else if log.action === 'Resubmission Requested'}
											<span class="inline-flex items-center rounded-full border border-amber-200 bg-amber-50 px-2.5 py-0.5 text-[11px] font-semibold text-amber-700 dark:border-amber-800 dark:bg-amber-950/40 dark:text-amber-300">
												Resubmission Requested
											</span>
										{:else if log.action === 'Rejected'}
											<span class="inline-flex items-center rounded-full border border-rose-200 bg-rose-50 px-2.5 py-0.5 text-[11px] font-semibold text-rose-700 dark:border-rose-800 dark:bg-rose-950/40 dark:text-rose-300">
												Rejected
											</span>
										{:else}
											<span class="inline-flex items-center rounded-full border border-gray-200 bg-gray-50 px-2.5 py-0.5 text-[11px] font-semibold text-gray-700 dark:border-gray-700 dark:bg-gray-800/40 dark:text-gray-300">
												{log.action}
											</span>
										{/if}
									</td>
									<td class="py-3.5 px-4 text-gray-700 dark:text-gray-300">
										{log.field}
									</td>
									<td class="py-3.5 px-4 font-mono text-[11px] text-gray-600 dark:text-gray-400">
										{log.old_value}
									</td>
									<td class="py-3.5 px-4 font-mono text-[11px] text-gray-900 dark:text-white font-medium">
										{log.new_value}
									</td>
									<td class="py-3.5 px-4 text-right font-mono text-[11px] text-gray-500 dark:text-gray-400 whitespace-nowrap">
										{formatDateTime(log.created_at)}
									</td>
								</tr>
							{/each}
						{/if}
					</tbody>
				</table>
			</div>

			<!-- Pagination Footer -->
			<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 p-4 border-t border-gray-100 dark:border-gray-800 text-xs text-gray-500">
				<div>
					Showing {totalRecords === 0 ? 0 : (currentPage - 1) * perPage + 1} to {Math.min(currentPage * perPage, totalRecords)} of {totalRecords} results
				</div>
				<div class="flex items-center gap-4">
					<div class="flex items-center gap-2">
						<span>Per page</span>
						<select
							value={perPage}
							onchange={handlePerPageChange}
							class="rounded-lg border border-gray-200 bg-white py-1 px-2 text-xs text-gray-700 outline-none dark:border-gray-700 dark:bg-gray-800 dark:text-gray-300"
						>
							<option value={10}>10</option>
							<option value={25}>25</option>
							<option value={50}>50</option>
						</select>
					</div>

					<div class="flex items-center gap-1">
						{#each Array.from({ length: totalPages }, (_, i) => i + 1) as p}
							<button
								type="button"
								onclick={() => goToPage(p)}
								class="h-7 w-7 rounded-lg border text-xs font-semibold transition {p === currentPage ? 'border-emerald-600 bg-emerald-600 text-white' : 'border-gray-200 bg-white text-gray-700 hover:bg-gray-50 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-300'}"
							>
								{p}
							</button>
						{/each}
						{#if totalPages > currentPage}
							<button
								type="button"
								onclick={() => goToPage(currentPage + 1)}
								class="h-7 w-7 rounded-lg border border-gray-200 bg-white text-xs font-semibold text-gray-700 hover:bg-gray-50 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-300 transition"
							>
								&gt;
							</button>
						{/if}
					</div>
				</div>
			</div>
		{/if}
	</div>
</div>
