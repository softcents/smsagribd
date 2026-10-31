<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface KYCPermissionItem {
		id: number;
		name: string;
		email: string;
		type: string;
		permissions: string[];
	}

	const PERMISSION_KEYS = [
		{ key: 'kyc.view', label: 'View KYC' },
		{ key: 'kyc.review', label: 'Review Queue' },
		{ key: 'kyc.approve', label: 'Approve' },
		{ key: 'kyc.reject', label: 'Reject' },
		{ key: 'kyc.resubmit', label: 'Resubmit' },
		{ key: 'kyc.audit', label: 'Audit Logs' }
	];

	let adminList = $state<KYCPermissionItem[]>([]);
	let loading = $state(true);
	let savingId = $state<number | null>(null);
	let savedId = $state<number | null>(null);
	let error = $state<string | null>(null);

	async function loadPermissions() {
		loading = true;
		error = null;
		try {
			const res = await supermanApi.get<{ data: KYCPermissionItem[] }>('/api/v1/superman/kyc/permissions');
			adminList = res.data;
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Unable to load KYC permissions.';
		} finally {
			loading = false;
		}
	}

	function hasPermission(admin: KYCPermissionItem, permKey: string): boolean {
		return admin.permissions.includes(permKey);
	}

	function togglePermission(admin: KYCPermissionItem, permKey: string) {
		if (admin.permissions.includes(permKey)) {
			admin.permissions = admin.permissions.filter((p) => p !== permKey);
		} else {
			admin.permissions = [...admin.permissions, permKey];
		}
	}

	function grantAll(admin: KYCPermissionItem) {
		admin.permissions = PERMISSION_KEYS.map((p) => p.key);
	}

	function revokeAll(admin: KYCPermissionItem) {
		admin.permissions = [];
	}

	async function saveAdminPermissions(admin: KYCPermissionItem) {
		savingId = admin.id;
		savedId = null;
		error = null;
		try {
			await supermanApi.put(`/api/v1/superman/kyc/permissions/${admin.id}`, {
				permissions: admin.permissions
			});
			savedId = admin.id;
			setTimeout(() => {
				if (savedId === admin.id) savedId = null;
			}, 3000);
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Failed to save permissions.';
		} finally {
			savingId = null;
		}
	}

	onMount(() => {
		void loadPermissions();
	});

	const exportColumns: ExportColumn<KYCPermissionItem>[] = [
		{ header: 'Administrator Name', getValue: (a) => a.name },
		{ header: 'Email', getValue: (a) => a.email },
		{ header: 'Role Type', getValue: (a) => a.type.toUpperCase() },
		{ header: 'Assigned Permissions', getValue: (a) => a.permissions.join(', ') || 'None' }
	];
</script>

<svelte:head>
	<title>KYC Permissions — SMS Whiteinfo Administrator</title>
</svelte:head>

<div class="space-y-5">
	<!-- Page Header -->
	<div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">KYC Permissions</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400">Configure role-based access permissions for reviewing, approving, and auditing KYC verification requests.</p>
		</div>
		<div class="flex items-center gap-2">
			<button
				type="button"
				onclick={() => void loadPermissions()}
				class="inline-flex items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-3 py-1.5 text-xs font-semibold text-gray-700 shadow-2xs hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300 transition"
			>
				<svg class="h-3.5 w-3.5 {loading ? 'animate-spin' : ''}" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0 3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182m0-4.991v4.99" />
				</svg>
				Refresh
			</button>
			<ExportDropdown
				title="KYC Permissions"
				columns={exportColumns}
				getData={() => adminList}
				disabled={adminList.length === 0}
			/>
		</div>
	</div>

	<!-- Permissions Matrix Container -->
	<div class="rounded-xl border border-gray-200 bg-white shadow-xs dark:border-gray-800 dark:bg-gray-900">
		{#if error}
			<div class="p-4 text-xs text-rose-600 border-b border-gray-100 dark:border-gray-800">{error}</div>
		{/if}

		{#if loading}
			<div class="py-16 text-center text-xs text-gray-400">Loading administrator permissions…</div>
		{:else if adminList.length === 0}
			<div class="py-16 text-center text-xs text-gray-400">No administrator accounts found.</div>
		{:else}
			<div class="overflow-x-auto">
				<table class="w-full text-left text-xs">
					<thead>
						<tr class="border-b border-gray-100 text-[11px] font-bold text-gray-600 uppercase dark:border-gray-800 dark:text-gray-400">
							<th class="py-3 px-4 min-w-[200px]">ADMINISTRATOR</th>
							{#each PERMISSION_KEYS as perm}
								<th class="py-3 px-3 text-center whitespace-nowrap">{perm.label}</th>
							{/each}
							<th class="py-3 px-4 text-right">ACTIONS</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
						{#each adminList as admin}
							<tr class="hover:bg-gray-50/70 dark:hover:bg-gray-800/40 transition-colors">
								<td class="py-4 px-4">
									<div class="font-bold text-gray-900 dark:text-white flex items-center gap-2">
										{admin.name}
										<span class="rounded bg-gray-100 px-1.5 py-0.5 text-[10px] font-mono text-gray-600 dark:bg-gray-800 dark:text-gray-300 uppercase">
											{admin.type}
										</span>
									</div>
									<div class="text-[11px] text-gray-500 dark:text-gray-400">{admin.email}</div>
									<div class="mt-1 flex items-center gap-2 text-[10px] text-gray-400">
										<button type="button" onclick={() => grantAll(admin)} class="text-emerald-600 hover:underline">Select All</button>
										<span>•</span>
										<button type="button" onclick={() => revokeAll(admin)} class="text-rose-500 hover:underline">Clear All</button>
									</div>
								</td>

								{#each PERMISSION_KEYS as perm}
									<td class="py-4 px-3 text-center">
										<label class="inline-flex items-center justify-center cursor-pointer p-1">
											<input
												type="checkbox"
												checked={hasPermission(admin, perm.key)}
												onchange={() => togglePermission(admin, perm.key)}
												class="h-4 w-4 rounded border-gray-300 text-emerald-600 focus:ring-emerald-500 transition cursor-pointer"
											/>
										</label>
									</td>
								{/each}

								<td class="py-4 px-4 text-right">
									<div class="inline-flex items-center gap-2">
										{#if savedId === admin.id}
											<span class="text-[11px] font-semibold text-emerald-600 animate-pulse">Saved ✓</span>
										{/if}
										<button
											type="button"
											onclick={() => void saveAdminPermissions(admin)}
											disabled={savingId === admin.id}
											class="rounded-lg bg-emerald-600 px-3 py-1.5 text-xs font-semibold text-white shadow-2xs hover:bg-emerald-700 disabled:opacity-50 transition"
										>
											{savingId === admin.id ? 'Saving…' : 'Save'}
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
