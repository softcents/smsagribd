<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface BackupItem {
		id: number;
		filename: string;
		size_mb: number;
		driver: string;
		status: string;
		created_at: string;
	}

	let backups = $state<BackupItem[]>([]);
	let loading = $state(true);
	let backingUp = $state(false);
	let message = $state<string | null>(null);

	const exportColumns: ExportColumn<BackupItem>[] = [
		{ header: 'ID', getValue: (b) => b.id },
		{ header: 'Archive Filename', getValue: (b) => b.filename },
		{ header: 'Size (MB)', getValue: (b) => `${b.size_mb.toFixed(2)} MB` },
		{ header: 'Storage Driver', getValue: (b) => b.driver },
		{ header: 'Status', getValue: (b) => b.status },
		{ header: 'Snapshot Timestamp', getValue: (b) => b.created_at }
	];

	onMount(async () => {
		try {
			const res = await supermanApi.get<{ data: BackupItem[] }>('/api/v1/superman/system-settings/backups');
			if (res.data) backups = res.data;
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	});

	async function takeBackup() {
		backingUp = true;
		message = null;
		try {
			const res = await supermanApi.post<{ message: string; backup: BackupItem }>('/api/v1/superman/system-settings/backups', {});
			message = res.message || 'Backup snapshot created successfully.';
			if (res.backup) {
				backups = [res.backup, ...backups];
			}
		} catch (e: any) {
			alert(e?.message || 'Failed to create backup snapshot.');
		} finally {
			backingUp = false;
		}
	}
</script>

<div class="space-y-6">
	<!-- Page Header -->
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-xl font-bold text-gray-900 dark:text-white">Database Backup & Snapshots</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
				Automated MySQL database snapshots, cloud offsite storage, and disaster recovery point-in-time archives.
			</p>
		</div>
		<div class="flex items-center gap-3">
			<ExportDropdown
				title="Database Backup Snapshots"
				columns={exportColumns}
				getData={() => backups}
				disabled={backups.length === 0}
			/>
			<button
				type="button"
				onclick={takeBackup}
				disabled={backingUp}
				class="inline-flex items-center gap-2 rounded-lg bg-emerald-600 px-3 py-1.5 text-xs font-semibold text-white shadow-xs hover:bg-emerald-700 transition disabled:opacity-50"
			>
				{backingUp ? 'Creating Snapshot...' : '+ Create Manual Backup Now'}
			</button>
		</div>
	</div>

	{#if message}
		<div class="rounded-lg bg-emerald-50 dark:bg-emerald-950/40 border border-emerald-200 dark:border-emerald-800 p-4 text-xs font-medium text-emerald-800 dark:text-emerald-300">
			{message}
		</div>
	{/if}

	<!-- Backup Archives Table -->
	<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl shadow-xs overflow-hidden">
		<div class="px-5 py-4 border-b border-gray-200 dark:border-gray-800">
			<h2 class="text-sm font-bold text-gray-900 dark:text-white">Existing Database Snapshots</h2>
		</div>
		<div class="overflow-x-auto">
			<table class="w-full text-left text-xs">
				<thead class="bg-gray-50/75 dark:bg-gray-800/50 text-gray-500 uppercase tracking-wider font-semibold border-b border-gray-200 dark:border-gray-800">
					<tr>
						<th class="px-5 py-3">Archive File Name</th>
						<th class="px-5 py-3">Size (MB)</th>
						<th class="px-5 py-3">Storage Driver</th>
						<th class="px-5 py-3">Integrity State</th>
						<th class="px-5 py-3">Created Date</th>
						<th class="px-5 py-3 text-right">Actions</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800 text-gray-700 dark:text-gray-300">
					{#if loading}
						<tr>
							<td colspan="6" class="px-5 py-8 text-center text-gray-400">Loading snapshot history...</td>
						</tr>
					{:else}
						{#each backups as b}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30 transition">
								<td class="px-5 py-3 font-mono font-semibold text-gray-900 dark:text-white">{b.filename}</td>
								<td class="px-5 py-3">{b.size_mb} MB</td>
								<td class="px-5 py-3 text-gray-500">{b.driver}</td>
								<td class="px-5 py-3">
									<span class="inline-flex items-center gap-1.5 text-xs text-emerald-600 font-medium">
										<span class="h-1.5 w-1.5 rounded-full bg-emerald-600"></span>
										{b.status}
									</span>
								</td>
								<td class="px-5 py-3 text-gray-500">{new Date(b.created_at).toLocaleString()}</td>
								<td class="px-5 py-3 text-right">
									<button
										type="button"
										class="text-xs text-emerald-600 hover:text-emerald-700 font-medium mr-3"
										onclick={() => alert(`Downloading archive ${b.filename}`)}
									>
										Download
									</button>
								</td>
							</tr>
						{/each}
					{/if}
				</tbody>
			</table>
		</div>
	</div>
</div>

