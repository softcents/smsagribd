<script lang="ts">
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface SignupGatewayItem {
		id: number;
		name: string;
		type: string;
		status: string;
		body_template: string;
		created_at: string;
	}

	let gateways = $state<SignupGatewayItem[]>([
		{ id: 1, name: 'Default OTP Sender Route', type: 'SMS Direct OTP', status: 'Active', body_template: 'Your WhiteInfo verification code is: {otp}. Valid for 5 minutes.', created_at: new Date().toISOString() }
	]);

	const exportColumns: ExportColumn<SignupGatewayItem>[] = [
		{ header: 'Gateway Channel', getValue: (g) => g.name },
		{ header: 'Channel Type', getValue: (g) => g.type },
		{ header: 'OTP Message Template', getValue: (g) => g.body_template },
		{ header: 'Status', getValue: (g) => g.status }
	];
</script>

<svelte:head>
	<title>Signup Register Gateways — Superman Administrator</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">Signup & Registration Gateways</h1>
			<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
				Manage automated OTP dispatch channels and verification SMS templates used during user registration.
			</p>
		</div>
		<div class="flex items-center gap-2">
			<ExportDropdown
				title="Signup Gateways"
				columns={exportColumns}
				getData={() => gateways}
				disabled={gateways.length === 0}
			/>
		</div>
	</div>

	<div class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-xs dark:border-gray-800 dark:bg-gray-900">
		<table class="w-full text-left text-xs">
			<thead>
				<tr class="border-b border-gray-100 bg-gray-50/50 text-[10px] font-semibold uppercase tracking-wider text-gray-500 dark:border-gray-800 dark:bg-gray-800/50 dark:text-gray-400">
					<th class="px-5 py-3.5">Gateway Channel</th>
					<th class="px-5 py-3.5">Channel Type</th>
					<th class="px-5 py-3.5">OTP Message Template</th>
					<th class="px-5 py-3.5 text-right">Status</th>
				</tr>
			</thead>
			<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
				{#each gateways as g}
					<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30">
						<td class="px-5 py-4 font-semibold text-gray-950 dark:text-white">{g.name}</td>
						<td class="px-5 py-4 font-mono text-blue-600">{g.type}</td>
						<td class="px-5 py-4 text-gray-600 dark:text-gray-400 font-mono text-[11px]">{g.body_template}</td>
						<td class="px-5 py-4 text-right">
							<span class="inline-flex rounded-full bg-emerald-50 px-2.5 py-0.5 text-[10px] font-bold uppercase text-emerald-700">
								{g.status}
							</span>
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
</div>

