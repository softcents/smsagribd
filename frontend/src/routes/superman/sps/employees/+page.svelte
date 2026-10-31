<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface SpEmployee {
		id: number;
		name: string;
		email: string;
		phone: string;
		reseller_name: string;
		sp_name: string;
		status: string;
		created_at: string;
	}

	let employees = $state<SpEmployee[]>([]);
	let loading = $state(true);

	const exportColumns: ExportColumn<SpEmployee>[] = [
		{ header: 'ID', getValue: (e) => e.id },
		{ header: 'Employee Name', getValue: (e) => e.name },
		{ header: 'Email', getValue: (e) => e.email },
		{ header: 'Phone', getValue: (e) => e.phone },
		{ header: 'Owning S P', getValue: (e) => e.sp_name || e.reseller_name },
		{ header: 'Account Status', getValue: (e) => e.status.toUpperCase() },
		{ header: 'Created At', getValue: (e) => new Date(e.created_at).toLocaleString() }
	];

	async function loadEmployees() {
		loading = true;
		try {
			const res = await supermanApi.get<{ data: SpEmployee[] }>('/api/v1/superman/sps/employees');
			employees = res.data ?? [];
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		loadEmployees();
	});

	function getStatusBadge(status: string) {
		const s = status.toLowerCase();
		if (s === 'approved' || s === 'active') {
			return 'bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-950/40 dark:text-emerald-400 dark:border-emerald-800';
		}
		if (s === 'banned' || s === 'rejected') {
			return 'bg-rose-50 text-rose-700 border-rose-200 dark:bg-rose-950/40 dark:text-rose-400 dark:border-rose-800';
		}
		return 'bg-amber-50 text-amber-700 border-amber-200 dark:bg-amber-950/40 dark:text-amber-400 dark:border-amber-800';
	}
</script>

<div class="space-y-6">
	<!-- Page Header -->
	<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
		<div>
			<h1 class="text-xl font-bold text-gray-900 dark:text-white">S P Employees</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
				Directory of wholesale partner staff and delegated employee sub-accounts.
			</p>
		</div>
		<div class="flex items-center gap-3">
			<button
				onclick={loadEmployees}
				class="inline-flex items-center gap-1.5 px-3 py-2 text-xs font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 dark:bg-gray-800 dark:text-gray-300 dark:border-gray-700 dark:hover:bg-gray-700/60 shadow-xs transition"
			>
				<svg class="w-3.5 h-3.5 {loading ? 'animate-spin' : ''}" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
				</svg>
				Refresh
			</button>
			<ExportDropdown
				title="S P Employees Directory"
				columns={exportColumns}
				getData={() => employees}
				disabled={employees.length === 0}
			/>
		</div>
	</div>

	<!-- Employees Table -->
	<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl shadow-xs overflow-hidden">
		<div class="overflow-x-auto">
			<table class="w-full text-left text-xs">
				<thead class="bg-gray-50/75 dark:bg-gray-800/50 text-gray-500 uppercase tracking-wider font-semibold border-b border-gray-200 dark:border-gray-800">
					<tr>
						<th class="px-5 py-3">Staff Name</th>
						<th class="px-5 py-3">Email</th>
						<th class="px-5 py-3">Phone</th>
						<th class="px-5 py-3">Owning S P</th>
						<th class="px-5 py-3 text-center">Status</th>
						<th class="px-5 py-3">Created Date</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800 text-gray-700 dark:text-gray-300">
					{#if loading}
						<tr>
							<td colspan="6" class="px-5 py-10 text-center text-gray-400">Loading S P employees...</td>
						</tr>
					{:else if employees.length === 0}
						<tr>
							<td colspan="6" class="px-5 py-12 text-center">
								<div class="max-w-sm mx-auto text-center space-y-1">
									<div class="w-10 h-10 rounded-full bg-gray-100 dark:bg-gray-800 flex items-center justify-center mx-auto text-gray-400">
										<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
										</svg>
									</div>
									<div class="font-bold text-gray-700 dark:text-gray-300">No S P Employees Found</div>
									<p class="text-xs text-gray-400">Resellers have not currently created any sub-operator accounts.</p>
								</div>
							</td>
						</tr>
					{:else}
						{#each employees as e}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30 transition">
								<td class="px-5 py-3.5 font-bold text-gray-900 dark:text-white">
									{e.name}
								</td>
								<td class="px-5 py-3.5 font-mono text-gray-600 dark:text-gray-400">
									{e.email}
								</td>
								<td class="px-5 py-3.5 font-mono text-gray-600 dark:text-gray-400">
									{e.phone}
								</td>
								<td class="px-5 py-3.5 font-semibold text-emerald-600 dark:text-emerald-400">
									{e.sp_name || e.reseller_name}
								</td>
								<td class="px-5 py-3.5 text-center">
									<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-[11px] font-semibold border {getStatusBadge(e.status)}">
										{e.status.toUpperCase()}
									</span>
								</td>
								<td class="px-5 py-3.5 font-mono text-gray-400">
									{new Date(e.created_at).toLocaleDateString()}
								</td>
							</tr>
						{/each}
					{/if}
				</tbody>
			</table>
		</div>
	</div>
</div>
