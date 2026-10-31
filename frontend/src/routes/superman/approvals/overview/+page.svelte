<script lang="ts">
	import { onMount } from 'svelte';
	import { supermanApi } from '$lib/superman-api.js';

	interface ApprovalsOverview {
		pending_customers: number;
		pending_resellers: number;
		pending_senders: number;
		pending_ip_whitelists: number;
	}

	let overview = $state<ApprovalsOverview>({
		pending_customers: 0,
		pending_resellers: 0,
		pending_senders: 0,
		pending_ip_whitelists: 0
	});
	let loading = $state(true);
	let error = $state<string | null>(null);

	async function loadOverview() {
		loading = true;
		error = null;
		try {
			const res = await supermanApi.get<{ data: ApprovalsOverview }>('/api/v1/superman/approvals/overview');
			if (res.data) overview = res.data;
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Failed to load approval center counts.';
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		void loadOverview();
	});
</script>

<svelte:head>
	<title>Approval Center — Superman Administrator</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">Central Approval Hub</h1>
			<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
				Administrative review queue for client onboarding, reseller privileges, sender IDs, and IP access requests.
			</p>
		</div>
		<button onclick={loadOverview} class="rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs font-medium text-gray-700 shadow-xs hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300">
			Refresh Queue
		</button>
	</div>

	<!-- 4 Approval Action Cards -->
	<div class="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4">
		<!-- Pending Customers -->
		<a href="/superman/approvals/customers" class="group flex flex-col justify-between rounded-2xl border border-gray-200 bg-white p-5 shadow-xs transition hover:border-emerald-500 hover:shadow-md dark:border-gray-800 dark:bg-gray-900">
			<div>
				<div class="flex items-center justify-between">
					<span class="text-xs font-semibold text-gray-500 uppercase tracking-wider">Pending Clients</span>
					<span class="rounded-full bg-amber-50 px-2 py-0.5 text-xs font-bold text-amber-700 dark:bg-amber-950/40 dark:text-amber-400">
						Action Required
					</span>
				</div>
				<p class="mt-3 text-3xl font-black text-gray-950 dark:text-white">{overview.pending_customers}</p>
				<p class="mt-1 text-xs text-gray-500">New customer signups requiring account activation</p>
			</div>
			<div class="mt-4 flex items-center gap-1 text-xs font-bold text-emerald-600 group-hover:underline">
				Review Customers →
			</div>
		</a>

		<!-- Pending Resellers -->
		<a href="/superman/approvals/resellers" class="group flex flex-col justify-between rounded-2xl border border-gray-200 bg-white p-5 shadow-xs transition hover:border-emerald-500 hover:shadow-md dark:border-gray-800 dark:bg-gray-900">
			<div>
				<div class="flex items-center justify-between">
					<span class="text-xs font-semibold text-gray-500 uppercase tracking-wider">Pending Resellers</span>
					<span class="rounded-full bg-blue-50 px-2 py-0.5 text-xs font-bold text-blue-700 dark:bg-blue-950/40 dark:text-blue-400">
						S P Onboarding
					</span>
				</div>
				<p class="mt-3 text-3xl font-black text-gray-950 dark:text-white">{overview.pending_resellers}</p>
				<p class="mt-1 text-xs text-gray-500">Service provider reseller applications</p>
			</div>
			<div class="mt-4 flex items-center gap-1 text-xs font-bold text-emerald-600 group-hover:underline">
				Review Resellers →
			</div>
		</a>

		<!-- Pending Senders -->
		<a href="/superman/approvals/senders" class="group flex flex-col justify-between rounded-2xl border border-gray-200 bg-white p-5 shadow-xs transition hover:border-emerald-500 hover:shadow-md dark:border-gray-800 dark:bg-gray-900">
			<div>
				<div class="flex items-center justify-between">
					<span class="text-xs font-semibold text-gray-500 uppercase tracking-wider">Pending Sender IDs</span>
					<span class="rounded-full bg-purple-50 px-2 py-0.5 text-xs font-bold text-purple-700 dark:bg-purple-950/40 dark:text-purple-400">
						Telco Regulatory
					</span>
				</div>
				<p class="mt-3 text-3xl font-black text-gray-950 dark:text-white">{overview.pending_senders}</p>
				<p class="mt-1 text-xs text-gray-500">Alphanumeric brand masking header requests</p>
			</div>
			<div class="mt-4 flex items-center gap-1 text-xs font-bold text-emerald-600 group-hover:underline">
				Review Sender IDs →
			</div>
		</a>

		<!-- IP Whitelists -->
		<a href="/superman/approval-center/ip-whitelists" class="group flex flex-col justify-between rounded-2xl border border-gray-200 bg-white p-5 shadow-xs transition hover:border-emerald-500 hover:shadow-md dark:border-gray-800 dark:bg-gray-900">
			<div>
				<div class="flex items-center justify-between">
					<span class="text-xs font-semibold text-gray-500 uppercase tracking-wider">IP Whitelists</span>
					<span class="rounded-full bg-rose-50 px-2 py-0.5 text-xs font-bold text-rose-700 dark:bg-rose-950/40 dark:text-rose-400">
						API Firewall
					</span>
				</div>
				<p class="mt-3 text-3xl font-black text-gray-950 dark:text-white">{overview.pending_ip_whitelists}</p>
				<p class="mt-1 text-xs text-gray-500">Direct HTTP API server IP access petitions</p>
			</div>
			<div class="mt-4 flex items-center gap-1 text-xs font-bold text-emerald-600 group-hover:underline">
				Review IP Requests →
			</div>
		</a>
	</div>
</div>

