<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface StaffMember {
		id: number;
		name: string;
		email: string;
		phone: string;
		last_login_at: string | null;
		status: string;
		created_at: string;
	}

	let staff = $state<StaffMember[]>([]);
	let loading = $state(true);

	const exportColumns: ExportColumn<StaffMember>[] = [
		{ header: 'ID', getValue: (s) => s.id },
		{ header: 'Staff Member', getValue: (s) => s.name },
		{ header: 'Staff Email', getValue: (s) => s.email },
		{ header: 'Emergency Phone', getValue: (s) => s.phone || '—' },
		{ header: 'Last Portal Login', getValue: (s) => s.last_login_at ? new Date(s.last_login_at).toLocaleString() : 'Never logged in' },
		{ header: 'Account Status', getValue: (s) => s.status.toUpperCase() },
		{ header: 'Member Since', getValue: (s) => new Date(s.created_at).toLocaleDateString() }
	];

	async function loadStaff() {
		loading = true;
		try {
			const res = await supermanApi.get<{ data: StaffMember[] }>('/api/v1/superman/audit-logs/employees');
			staff = res.data ?? [];
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		loadStaff();
	});

	function getStatusBadge(status: string) {
		const s = status.toLowerCase();
		if (s === 'approved' || s === 'active') {
			return 'bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-950/40 dark:text-emerald-400 dark:border-emerald-800';
		}
		if (s === 'banned' || s === 'suspended') {
			return 'bg-rose-50 text-rose-700 border-rose-200 dark:bg-rose-950/40 dark:text-rose-400 dark:border-rose-800';
		}
		return 'bg-amber-50 text-amber-700 border-amber-200 dark:bg-amber-950/40 dark:text-amber-400 dark:border-amber-800';
	}
</script>

<div class="space-y-6">
	<!-- Page Header -->
	<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
		<div>
			<h1 class="text-xl font-bold text-gray-900 dark:text-white">Administrative Staff & Staff Login Activity</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
				Real-time monitoring of Superman administrative operators, secure portal session timestamps, and staff accounts.
			</p>
		</div>
		<div class="flex items-center gap-3">
			<button
				onclick={loadStaff}
				class="inline-flex items-center gap-1.5 px-3 py-2 text-xs font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 dark:bg-gray-800 dark:text-gray-300 dark:border-gray-700 dark:hover:bg-gray-700/60 shadow-xs transition"
			>
				<svg class="w-3.5 h-3.5 {loading ? 'animate-spin' : ''}" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
				</svg>
				Refresh
			</button>
			<ExportDropdown
				title="Staff Member & Login Activity"
				columns={exportColumns}
				getData={() => staff}
				disabled={staff.length === 0}
			/>
		</div>
	</div>

	<!-- Table -->
	<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl shadow-xs overflow-hidden">
		<div class="overflow-x-auto">
			<table class="w-full text-left text-xs">
				<thead class="bg-gray-50/75 dark:bg-gray-800/50 text-gray-500 uppercase tracking-wider font-semibold border-b border-gray-200 dark:border-gray-800">
					<tr>
						<th class="px-5 py-3">Staff Member</th>
						<th class="px-5 py-3">Staff Email</th>
						<th class="px-5 py-3">Emergency Phone</th>
						<th class="px-5 py-3">Last Portal Login</th>
						<th class="px-5 py-3">Status</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800 text-gray-700 dark:text-gray-300">
					{#if loading}
						<tr>
							<td colspan="5" class="px-5 py-10 text-center text-gray-400">Loading staff activity...</td>
						</tr>
					{:else if staff.length === 0}
						<tr>
							<td colspan="5" class="px-5 py-10 text-center text-gray-400">No staff members found.</td>
						</tr>
					{:else}
						{#each staff as s}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30 transition">
								<td class="px-5 py-4">
									<div class="font-bold text-gray-900 dark:text-white flex items-center gap-2">
										<div class="w-7 h-7 rounded-full bg-emerald-100 dark:bg-emerald-950/60 text-emerald-700 dark:text-emerald-400 flex items-center justify-center font-bold text-xs uppercase">
											{s.name.substring(0, 2)}
										</div>
										<span>{s.name}</span>
									</div>
								</td>
								<td class="px-5 py-4 font-mono text-gray-700 dark:text-gray-300">
									{s.email}
								</td>
								<td class="px-5 py-4 font-mono text-gray-600 dark:text-gray-400">
									{s.phone || '—'}
								</td>
								<td class="px-5 py-4 whitespace-nowrap font-mono">
									{#if s.last_login_at}
										<span class="text-gray-700 dark:text-gray-300">
											{new Date(s.last_login_at).toLocaleString('en-GB', { day: '2-digit', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit', hour12: true })}
										</span>
									{:else}
										<span class="text-gray-400 italic">Never logged in</span>
									{/if}
								</td>
								<td class="px-5 py-4">
									<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-[11px] font-semibold border {getStatusBadge(s.status)}">
										{s.status.toUpperCase()}
									</span>
								</td>
							</tr>
						{/each}
					{/if}
				</tbody>
			</table>
		</div>
	</div>
</div>
