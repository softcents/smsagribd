<script lang="ts">
	import { spApi } from '$lib/sp-api.js';
	import { onMount } from 'svelte';

	interface SmsLog {
		id: number;
		customer: string;
		recipient: string;
		body: string;
		cost: number;
		status: string;
		created_at: string;
	}

	let logs = $state<SmsLog[]>([]);
	let loading = $state(true);

	onMount(async () => {
		try {
			const res = await spApi.get<{ data: SmsLog[] }>('/api/v1/sp/messages');
			logs = res.data ?? [];
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>Customer SMS Logs — SMS Whiteinfo S P Reseller</title>
</svelte:head>

<div class="space-y-4">
	<div>
		<h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">Customer SMS Traffic</h1>
		<p class="text-xs text-gray-500 dark:text-gray-400">Real-time delivery status for messages sent by your registered clients.</p>
	</div>

	<div class="rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900 space-y-4">
		<div class="overflow-x-auto">
			<table class="w-full text-left text-xs">
				<thead>
					<tr class="border-b border-gray-100 dark:border-gray-800 text-gray-400 uppercase text-[10px] font-semibold">
						<th class="py-2.5">Message ID</th>
						<th class="py-2.5">Customer</th>
						<th class="py-2.5">Recipient</th>
						<th class="py-2.5">Content</th>
						<th class="py-2.5">Deduction</th>
						<th class="py-2.5">Status</th>
						<th class="py-2.5">Time</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
					{#if loading}
						<tr>
							<td colspan="7" class="py-8 text-center text-gray-400">
								<div class="inline-block h-5 w-5 animate-spin rounded-full border-2 border-amber-500 border-t-transparent"></div>
								<div class="mt-1 text-xs">Loading customer SMS traffic logs…</div>
							</td>
						</tr>
					{:else if logs.length === 0}
						<tr>
							<td colspan="7" class="py-8 text-center text-gray-400">
								No message delivery logs found for assigned clients.
							</td>
						</tr>
					{:else}
						{#each logs as log}
							<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30 transition-colors">
								<td class="py-3 font-mono text-gray-500">#{log.id}</td>
								<td class="py-3 font-semibold text-gray-900 dark:text-white">{log.customer}</td>
								<td class="py-3 font-mono text-gray-700 dark:text-gray-300">{log.recipient}</td>
								<td class="py-3 text-gray-600 dark:text-gray-300 max-w-sm truncate">{log.body}</td>
								<td class="py-3 font-mono text-emerald-600 dark:text-emerald-400 font-semibold">৳{log.cost.toFixed(2)}</td>
								<td class="py-3">
									<span class="inline-flex items-center gap-1 text-[11px] font-semibold text-emerald-600 dark:text-emerald-400">
										<span class="w-1.5 h-1.5 rounded-full bg-emerald-500"></span>
										{log.status}
									</span>
								</td>
								<td class="py-3 text-[11px] font-mono text-gray-400">{log.created_at}</td>
							</tr>
						{/each}
					{/if}
				</tbody>
			</table>
		</div>
	</div>
</div>
