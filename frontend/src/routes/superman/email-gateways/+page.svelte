<script lang="ts">
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface EmailGatewayItem {
		id: number;
		name: string;
		host: string;
		port: number;
		username: string;
		encryption: string;
		is_active: boolean;
	}

	let gateways = $state<EmailGatewayItem[]>([
		{ id: 1, name: 'Primary Transactional SMTP', host: 'smtp.mailgun.org', port: 587, username: 'postmaster@sms.whiteinfobd.com', encryption: 'TLS', is_active: true },
		{ id: 2, name: 'Backup Amazon SES', host: 'email-smtp.ap-south-1.amazonaws.com', port: 465, username: 'AKIAEXAMPLE...', encryption: 'SSL', is_active: true }
	]);

	const exportColumns: ExportColumn<EmailGatewayItem>[] = [
		{ header: 'Gateway Name', getValue: (g) => g.name },
		{ header: 'Host', getValue: (g) => g.host },
		{ header: 'Port', getValue: (g) => g.port },
		{ header: 'Username / Sender', getValue: (g) => g.username },
		{ header: 'Encryption', getValue: (g) => g.encryption },
		{ header: 'Status', getValue: (g) => (g.is_active ? 'Active' : 'Inactive') }
	];
</script>

<svelte:head>
	<title>Email Gateways — Superman Administrator</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">Email (SMTP) Gateways</h1>
			<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
				Outbound mailers for 2FA login verification codes, deposit receipts, and alert dispatches.
			</p>
		</div>
		<div class="flex items-center gap-2">
			<ExportDropdown
				title="Email Gateways"
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
					<th class="px-5 py-3.5">Gateway Name</th>
					<th class="px-5 py-3.5">SMTP Host & Port</th>
					<th class="px-5 py-3.5">Username / Sender</th>
					<th class="px-5 py-3.5">Encryption</th>
					<th class="px-5 py-3.5 text-right">Status</th>
				</tr>
			</thead>
			<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
				{#each gateways as g}
					<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30">
						<td class="px-5 py-4 font-semibold text-gray-950 dark:text-white">{g.name}</td>
						<td class="px-5 py-4 font-mono">{g.host}:{g.port}</td>
						<td class="px-5 py-4 text-gray-600 dark:text-gray-400">{g.username}</td>
						<td class="px-5 py-4 font-mono uppercase text-blue-600 font-bold">{g.encryption}</td>
						<td class="px-5 py-4 text-right">
							<span class="inline-flex rounded-full bg-emerald-50 px-2.5 py-0.5 text-[10px] font-bold uppercase text-emerald-700">
								Active
							</span>
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
</div>

