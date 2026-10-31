<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface SpSender {
		id: number;
		sender_id: string;
		type: string;
		approval_status: string;
		reseller_name: string;
		created_at: string;
	}

	let senders = $state<SpSender[]>([]);
	let loading = $state(true);

	const exportColumns: ExportColumn<SpSender>[] = [
		{ header: 'ID', getValue: (s) => s.id },
		{ header: 'Sender ID / Header', getValue: (s) => s.sender_id },
		{ header: 'Type', getValue: (s) => s.type },
		{ header: 'Parent Reseller', getValue: (s) => s.reseller_name },
		{ header: 'Status', getValue: (s) => s.approval_status },
		{ header: 'Created At', getValue: (s) => s.created_at }
	];

	onMount(async () => {
		try {
			const res = await supermanApi.get<{ data: SpSender[] }>('/api/v1/superman/sps/senders');
			senders = res.data ?? [];
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	});
</script>

<div class="space-y-6">
	<!-- Page Header -->
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-xl font-bold text-gray-900 dark:text-white">Reseller Senders & CLI Bindings</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
				Alphanumeric headers and numeric shortcodes allocated across Service Provider reseller accounts.
			</p>
		</div>
		<div class="flex items-center gap-3">
			<ExportDropdown
				title="Reseller Senders and CLI Bindings"
				columns={exportColumns}
				getData={() => senders}
				disabled={senders.length === 0}
			/>
			<a
				href="/superman/sender-cli/add"
				class="inline-flex items-center gap-2 rounded-lg bg-emerald-600 px-3 py-1.5 text-xs font-semibold text-white shadow-xs hover:bg-emerald-700 transition"
			>
				+ Assign Sender ID
			</a>
		</div>
	</div>

	<!-- Senders Table -->
	<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl shadow-xs overflow-hidden">
		<div class="overflow-x-auto">
			<table class="w-full text-left text-xs">
				<thead class="bg-gray-50/75 dark:bg-gray-800/50 text-gray-500 uppercase tracking-wider font-semibold border-b border-gray-200 dark:border-gray-800">
					<tr>
						<th class="px-5 py-3">Sender Header (CLI)</th>
						<th class="px-5 py-3">Owner Reseller</th>
						<th class="px-5 py-3">Header Type</th>
						<th class="px-5 py-3">Status</th>
						<th class="px-5 py-3">Assigned Date</th>
						<th class="px-5 py-3 text-right">Actions</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800 text-gray-700 dark:text-gray-300">
					{#if loading}
						<tr>
							<td colspan="6" class="px-5 py-8 text-center text-gray-400">Loading reseller senders...</td>
						</tr>
					{:else if senders.length === 0}
						<tr>
							<td colspan="6" class="px-5 py-8 text-center text-gray-400">No reseller senders found.</td>
						</tr>
					{:else}
						{#each senders as s}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30 transition">
								<td class="px-5 py-3 font-mono font-bold text-gray-900 dark:text-white">{s.sender_id}</td>
								<td class="px-5 py-3 font-semibold text-emerald-600 dark:text-emerald-400">{s.reseller_name}</td>
								<td class="px-5 py-3 font-medium uppercase text-gray-500">{s.type}</td>
								<td class="px-5 py-3">
									<span class="inline-flex items-center px-2 py-0.5 rounded text-[11px] font-medium {s.approval_status === 'approved' ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300' : 'bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-300'}">
										{s.approval_status}
									</span>
								</td>
								<td class="px-5 py-3 text-gray-400">{new Date(s.created_at).toLocaleDateString()}</td>
								<td class="px-5 py-3 text-right">
									<a href="/superman/senders" class="text-xs text-emerald-600 hover:text-emerald-700 font-semibold">
										Configure
									</a>
								</td>
							</tr>
						{/each}
					{/if}
				</tbody>
			</table>
		</div>
	</div>
</div>

