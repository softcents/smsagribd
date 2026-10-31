<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	interface GatewayItem {
		id: number;
		name: string;
		driver: string;
		priority: number;
		is_enabled: boolean;
		is_otp_enabled: boolean;
		requests_per_second: number;
		health_status: string;
		route_code?: string | null;
		cost_mode?: string;
		dlr_support?: boolean;
		created_at?: string;
	}

	let gateways = $state<GatewayItem[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let search = $state('');
	let driverFilter = $state('all');
	let statusFilter = $state('all');
	let activeMenuId = $state<number | null>(null);

	// Create / Edit modal state
	let showModal = $state(false);
	let modalMode = $state<'create' | 'edit'>('create');
	let editingId = $state<number | null>(null);
	let saving = $state(false);
	let modalError = $state<string | null>(null);

	let formName = $state('');
	let formDriver = $state('smpp');
	let formPriority = $state(1);
	let formIsEnabled = $state(true);
	let formIsOtpEnabled = $state(true);
	let formRps = $state(60);
	let formTimeout = $state(10);
	let formCostMode = $state('manual');
	let formDlrSupport = $state(true);

	async function loadGateways() {
		loading = true;
		error = null;
		try {
			const res = await supermanApi.get<{ data: GatewayItem[] }>('/api/v1/superman/gateways');
			gateways = res.data ?? [];
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Failed to load gateways.';
		} finally {
			loading = false;
		}
	}

	function openCreateModal() {
		modalMode = 'create';
		editingId = null;
		formName = '';
		formDriver = 'smpp';
		formPriority = 1;
		formIsEnabled = true;
		formIsOtpEnabled = true;
		formRps = 60;
		formTimeout = 10;
		formCostMode = 'manual';
		formDlrSupport = true;
		modalError = null;
		showModal = true;
		activeMenuId = null;
	}

	function openEditModal(g: GatewayItem, e?: Event) {
		if (e) e.stopPropagation();
		activeMenuId = null;
		modalMode = 'edit';
		editingId = g.id;
		formName = g.name;
		formDriver = g.driver;
		formPriority = g.priority;
		formIsEnabled = g.is_enabled;
		formIsOtpEnabled = g.is_otp_enabled;
		formRps = g.requests_per_second;
		formTimeout = 10;
		formCostMode = g.cost_mode || 'manual';
		formDlrSupport = g.dlr_support ?? true;
		modalError = null;
		showModal = true;
	}

	async function handleDelete(g: GatewayItem, e?: Event) {
		if (e) e.stopPropagation();
		activeMenuId = null;
		if (!confirm(`Are you sure you want to delete "${g.name}" (ID #${g.id})?`)) return;

		try {
			await supermanApi.delete(`/api/v1/superman/gateways/${g.id}`);
			await loadGateways();
		} catch (err: unknown) {
			alert(err instanceof Error ? err.message : 'Failed to delete gateway.');
		}
	}

	async function saveGateway() {
		saving = true;
		modalError = null;
		try {
			const payload = {
				name: formName,
				driver: formDriver,
				priority: Number(formPriority),
				is_enabled: formIsEnabled,
				is_otp_enabled: formIsOtpEnabled,
				requests_per_second: Number(formRps),
				timeout_seconds: Number(formTimeout),
				cost_mode: formCostMode,
				dlr_support: formDlrSupport
			};

			if (modalMode === 'create') {
				await supermanApi.post('/api/v1/superman/gateways', payload);
			} else if (editingId) {
				await supermanApi.put(`/api/v1/superman/gateways/${editingId}`, payload);
			}
			showModal = false;
			await loadGateways();
		} catch (err: unknown) {
			modalError = err instanceof Error ? err.message : 'Failed to save gateway.';
		} finally {
			saving = false;
		}
	}

	function toggleMenu(id: number, e: Event) {
		e.stopPropagation();
		activeMenuId = activeMenuId === id ? null : id;
	}

	onMount(() => {
		void loadGateways();
	});

	let filteredGateways = $derived(
		gateways.filter((g) => {
			const matchesSearch =
				search === '' ||
				g.name.toLowerCase().includes(search.toLowerCase()) ||
				g.driver.toLowerCase().includes(search.toLowerCase()) ||
				(g.route_code && g.route_code.toLowerCase().includes(search.toLowerCase()));
			const matchesDriver = driverFilter === 'all' || g.driver === driverFilter;
			const matchesStatus =
				statusFilter === 'all' ||
				(statusFilter === 'enabled' && g.is_enabled) ||
				(statusFilter === 'disabled' && !g.is_enabled);
			return matchesSearch && matchesDriver && matchesStatus;
		})
	);

	let totalRps = $derived(gateways.reduce((acc, g) => acc + (g.is_enabled ? g.requests_per_second : 0), 0));
	let activeCount = $derived(gateways.filter((g) => g.is_enabled).length);

	const exportColumns: ExportColumn<GatewayItem>[] = [
		{ header: 'Provider / Gateway', getValue: (g) => g.name },
		{ header: 'Gateway ID', getValue: (g) => `#${g.id}` },
		{ header: 'Route Code', getValue: (g) => g.route_code || '—' },
		{ header: 'Driver', getValue: (g) => g.driver.toUpperCase() },
		{ header: 'Priority', getValue: (g) => g.priority },
		{ header: 'Capacity (RPS)', getValue: (g) => `${g.requests_per_second} req/sec` },
		{ header: 'OTP Routing', getValue: (g) => (g.is_otp_enabled ? 'OTP Priority' : 'Standard Only') },
		{ header: 'Operating State', getValue: (g) => (g.is_enabled ? 'Enabled' : 'Disabled') }
	];
</script>

<svelte:window onclick={() => (activeMenuId = null)} />

<svelte:head>
	<title>Gateways / Providers — Superman Administrator</title>
</svelte:head>

<div class="space-y-6">
	<!-- Top Bar -->
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">Gateways / Providers</h1>
			<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
				Manage SMS upstream gateways, connectivity parameters, traffic distribution, and live routing priority.
			</p>
		</div>
		<div class="flex items-center gap-2">
			<button
				onclick={loadGateways}
				class="inline-flex items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs font-medium text-gray-700 shadow-xs hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300 dark:hover:bg-gray-800"
			>
				<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0 3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182m0-4.991v4.99" />
				</svg>
				Refresh
			</button>
			<ExportDropdown
				title="Gateways and Providers"
				columns={exportColumns}
				getData={() => filteredGateways}
				disabled={filteredGateways.length === 0}
			/>
			<button
				onclick={openCreateModal}
				class="inline-flex items-center gap-1.5 rounded-lg bg-emerald-600 px-3 py-2 text-xs font-semibold text-white shadow-xs hover:bg-emerald-700"
			>
				<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
				</svg>
				Add Gateway
			</button>
		</div>
	</div>

	<!-- KPI Summary Bar -->
	<div class="grid grid-cols-2 gap-4 sm:grid-cols-4">
		<div class="rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
			<span class="text-[11px] font-medium text-gray-500 uppercase tracking-wider">Total Gateways</span>
			<p class="mt-1 text-2xl font-bold text-gray-900 dark:text-white">{gateways.length}</p>
		</div>
		<div class="rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
			<span class="text-[11px] font-medium text-gray-500 uppercase tracking-wider">Active Providers</span>
			<p class="mt-1 text-2xl font-bold text-emerald-600">{activeCount}</p>
		</div>
		<div class="rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
			<span class="text-[11px] font-medium text-gray-500 uppercase tracking-wider">Aggregate Capacity</span>
			<p class="mt-1 text-2xl font-bold text-blue-600">{totalRps} <span class="text-xs font-normal text-gray-500">TPS</span></p>
		</div>
		<div class="rounded-xl border border-gray-200 bg-white p-4 shadow-xs dark:border-gray-800 dark:bg-gray-900">
			<span class="text-[11px] font-medium text-gray-500 uppercase tracking-wider">Health Status</span>
			<p class="mt-1 text-2xl font-bold text-emerald-600">100% <span class="text-xs font-normal text-gray-500">Online</span></p>
		</div>
	</div>

	<!-- Filter / Search Bar -->
	<div class="flex flex-col gap-3 rounded-xl border border-gray-200 bg-white p-4 shadow-xs sm:flex-row sm:items-center sm:justify-between dark:border-gray-800 dark:bg-gray-900">
		<div class="relative w-full max-w-sm">
			<input
				type="search"
				bind:value={search}
				placeholder="Search by provider name or driver..."
				class="w-full rounded-lg border border-gray-200 bg-gray-50/70 py-2 pl-3 pr-3 text-xs outline-none focus:border-emerald-500 focus:bg-white dark:border-gray-700 dark:bg-gray-800 dark:text-white"
			/>
		</div>
		<div class="flex items-center gap-2">
			<select
				bind:value={driverFilter}
				class="rounded-lg border border-gray-200 bg-white px-2.5 py-1.5 text-xs text-gray-700 outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-300"
			>
				<option value="all">All Drivers</option>
				<option value="infozillion">Infozillion</option>
				<option value="smpp">SMPP</option>
				<option value="http">HTTP / REST API</option>
				<option value="infobip">Infobip</option>
				<option value="twilio">Twilio</option>
			</select>
			<select
				bind:value={statusFilter}
				class="rounded-lg border border-gray-200 bg-white px-2.5 py-1.5 text-xs text-gray-700 outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-300"
			>
				<option value="all">All Statuses</option>
				<option value="enabled">Enabled Only</option>
				<option value="disabled">Disabled Only</option>
			</select>
		</div>
	</div>

	<!-- Data Table -->
	<div class="overflow-visible rounded-xl border border-gray-200 bg-white shadow-xs dark:border-gray-800 dark:bg-gray-900">
		{#if loading}
			<div class="flex items-center justify-center py-16 text-xs text-gray-500">
				<svg class="mr-2 h-4 w-4 animate-spin" fill="none" viewBox="0 0 24 24">
					<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
					<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
				</svg>
				Loading provider gateways...
			</div>
		{:else if error}
			<div class="p-6 text-center text-xs text-rose-600">
				{error}
			</div>
		{:else if filteredGateways.length === 0}
			<div class="py-12 text-center text-xs text-gray-500 dark:text-gray-400">
				No gateways match your search criteria.
			</div>
		{:else}
			<div class="overflow-x-auto overflow-y-visible">
				<table class="w-full text-left text-xs">
					<thead>
						<tr class="border-b border-gray-100 bg-gray-50/50 text-[10px] font-semibold uppercase tracking-wider text-gray-500 dark:border-gray-800 dark:bg-gray-800/50 dark:text-gray-400">
							<th class="px-4 py-3">Provider / Gateway</th>
							<th class="px-4 py-3">Route Code</th>
							<th class="px-4 py-3">Driver</th>
							<th class="px-4 py-3">Priority</th>
							<th class="px-4 py-3">Capacity (RPS)</th>
							<th class="px-4 py-3">OTP Routing</th>
							<th class="px-4 py-3">Operating State</th>
							<th class="px-4 py-3 text-right">Actions</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
						{#each filteredGateways as g}
							<tr
								onclick={() => goto(`/superman/gateways/${g.id}`)}
								class="cursor-pointer transition-colors hover:bg-gray-50/70 dark:hover:bg-gray-800/40"
							>
								<td class="px-4 py-3.5">
									<div class="font-semibold text-gray-950 dark:text-white">{g.name}</div>
									<div class="text-[11px] text-gray-400">Gateway ID: #{g.id}</div>
								</td>
								<td class="px-4 py-3.5">
									{#if g.route_code}
										<span class="inline-flex items-center rounded-md bg-teal-50 px-2 py-0.5 text-[11px] font-semibold text-teal-700 dark:bg-teal-950/40 dark:text-teal-300">
											{g.route_code}
										</span>
									{:else}
										<span class="text-gray-400">—</span>
									{/if}
								</td>
								<td class="px-4 py-3.5">
									<span class="inline-flex items-center rounded-md bg-blue-50 px-2 py-0.5 text-[11px] font-medium text-blue-700 uppercase dark:bg-blue-950/40 dark:text-blue-300">
										{g.driver}
									</span>
								</td>
								<td class="px-4 py-3.5">
									<span class="inline-flex h-6 w-6 items-center justify-center rounded-full bg-gray-100 text-xs font-bold text-gray-700 dark:bg-gray-800 dark:text-gray-300">
										{g.priority}
									</span>
								</td>
								<td class="px-4 py-3.5 font-medium text-gray-700 dark:text-gray-300">
									{g.requests_per_second} req/sec
								</td>
								<td class="px-4 py-3.5">
									{#if g.is_otp_enabled}
										<span class="inline-flex items-center gap-1 text-emerald-600 font-medium">
											<svg class="h-3.5 w-3.5" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M16.704 4.153a.75.75 0 0 1 .143 1.052l-8 10.5a.75.75 0 0 1-1.127.075l-4.5-4.5a.75.75 0 0 1 1.06-1.06l3.894 3.893 7.48-9.817a.75.75 0 0 1 1.05-.143Z" clip-rule="evenodd" /></svg>
											OTP Priority
										</span>
									{:else}
										<span class="text-gray-400">Standard Only</span>
									{/if}
								</td>
								<td class="px-4 py-3.5">
									{#if g.is_enabled}
										<span class="inline-flex items-center gap-1.5 rounded-full bg-emerald-50 px-2.5 py-0.5 text-[11px] font-medium text-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-400">
											<span class="h-1.5 w-1.5 rounded-full bg-emerald-500"></span>
											Enabled
										</span>
									{:else}
										<span class="inline-flex items-center gap-1.5 rounded-full bg-gray-100 px-2.5 py-0.5 text-[11px] font-medium text-gray-600 dark:bg-gray-800 dark:text-gray-400">
											<span class="h-1.5 w-1.5 rounded-full bg-gray-400"></span>
											Disabled
										</span>
									{/if}
								</td>
								<td class="relative px-4 py-3.5 text-right">
									<!-- Three dots button -->
									<button
										type="button"
										onclick={(e) => toggleMenu(g.id, e)}
										class="inline-flex h-8 w-8 items-center justify-center rounded-lg border border-transparent text-gray-500 hover:border-gray-200 hover:bg-gray-100 focus:outline-none dark:text-gray-400 dark:hover:border-gray-700 dark:hover:bg-gray-800"
										aria-label="Actions"
									>
										<svg class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
											<path d="M10 6a2 2 0 110-4 2 2 0 010 4zM10 12a2 2 0 110-4 2 2 0 010 4zM10 18a2 2 0 110-4 2 2 0 010 4z" />
										</svg>
									</button>

									<!-- Three dots dropdown menu -->
									{#if activeMenuId === g.id}
										<div
											class="absolute right-4 top-12 z-50 w-40 rounded-xl border border-gray-100 bg-white py-1.5 shadow-xl ring-1 ring-black/5 dark:border-gray-800 dark:bg-gray-900"
										>
											<button
												type="button"
												onclick={(e) => {
													e.stopPropagation();
													activeMenuId = null;
													void goto(`/superman/gateways/${g.id}`);
												}}
												class="flex w-full items-center gap-2 px-3.5 py-2 text-left text-xs font-medium text-gray-700 hover:bg-gray-50 dark:text-gray-200 dark:hover:bg-gray-800/60"
											>
												<svg class="h-4 w-4 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M2.036 12.322a1.012 1.012 0 010-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178z" />
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
												</svg>
												View
											</button>
											<button
												type="button"
												onclick={(e) => openEditModal(g, e)}
												class="flex w-full items-center gap-2 px-3.5 py-2 text-left text-xs font-medium text-gray-700 hover:bg-gray-50 dark:text-gray-200 dark:hover:bg-gray-800/60"
											>
												<svg class="h-4 w-4 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="m16.862 4.487 1.687-1.688a1.875 1.875 0 1 1 2.652 2.652L10.582 16.07a4.5 4.5 0 0 1-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 0 1 1.13-1.897l8.932-8.931Zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0 1 15.75 21H5.25A2.25 2.25 0 0 1 3 18.75V8.25A2.25 2.25 0 0 1 5.25 6H10" />
												</svg>
												Edit
											</button>
											<div class="my-1 border-t border-gray-100 dark:border-gray-800"></div>
											<button
												type="button"
												onclick={(e) => void handleDelete(g, e)}
												class="flex w-full items-center gap-2 px-3.5 py-2 text-left text-xs font-medium text-rose-600 hover:bg-rose-50/50 dark:text-rose-400 dark:hover:bg-rose-950/20"
											>
												<svg class="h-4 w-4 text-rose-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="m14.74 9-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 0 0-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 0 0-7.5 0" />
												</svg>
												Delete
											</button>
										</div>
									{/if}
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</div>
</div>

<!-- Add / Edit Modal -->
{#if showModal}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-gray-900/50 p-4 backdrop-blur-xs">
		<div class="w-full max-w-lg rounded-2xl border border-gray-200 bg-white p-6 shadow-xl dark:border-gray-800 dark:bg-gray-900">
			<div class="flex items-center justify-between border-b border-gray-100 pb-3 dark:border-gray-800">
				<h3 class="text-base font-bold text-gray-900 dark:text-white">
					{modalMode === 'create' ? 'Add New SMS Gateway' : `Edit Gateway: ${formName}`}
				</h3>
				<button onclick={() => (showModal = false)} class="rounded-lg p-1 text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800">
					<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
				</button>
			</div>

			{#if modalError}
				<div class="mt-3 rounded-lg bg-rose-50 p-3 text-xs text-rose-700 dark:bg-rose-950/40 dark:text-rose-300">{modalError}</div>
			{/if}

			<form onsubmit={(e) => { e.preventDefault(); void saveGateway(); }} class="mt-4 space-y-4">
				<div>
					<label class="block text-xs font-semibold text-gray-700 dark:text-gray-300">Provider Name</label>
					<input
						type="text"
						bind:value={formName}
						required
						placeholder="e.g. Banglalink SMPP Direct"
						class="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2 text-xs outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
					/>
				</div>

				<div class="grid grid-cols-2 gap-3">
					<div>
						<label class="block text-xs font-semibold text-gray-700 dark:text-gray-300">Driver</label>
						<select
							bind:value={formDriver}
							class="mt-1 w-full rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
						>
							<option value="infozillion">Infozillion</option>
							<option value="smpp">SMPP Protocol</option>
							<option value="http">Generic HTTP / REST</option>
							<option value="infobip">Infobip</option>
							<option value="twilio">Twilio</option>
						</select>
					</div>
					<div>
						<label class="block text-xs font-semibold text-gray-700 dark:text-gray-300">Routing Priority</label>
						<input
							type="number"
							bind:value={formPriority}
							min="1"
							max="99"
							class="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2 text-xs outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
						/>
					</div>
				</div>

				<div class="grid grid-cols-2 gap-3">
					<div>
						<label class="block text-xs font-semibold text-gray-700 dark:text-gray-300">Max Throughput (RPS)</label>
						<input
							type="number"
							bind:value={formRps}
							min="1"
							max="1000"
							class="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2 text-xs outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
						/>
					</div>
					<div>
						<label class="block text-xs font-semibold text-gray-700 dark:text-gray-300">Timeout (Seconds)</label>
						<input
							type="number"
							bind:value={formTimeout}
							min="1"
							max="60"
							class="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2 text-xs outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
						/>
					</div>
				</div>

				<div class="space-y-2 border-t border-gray-100 pt-3 dark:border-gray-800">
					<label class="flex items-center gap-2 cursor-pointer">
						<input type="checkbox" bind:checked={formIsEnabled} class="h-4 w-4 rounded border-gray-300 text-emerald-600 focus:ring-emerald-500" />
						<span class="text-xs text-gray-700 dark:text-gray-300">Enable this gateway for active message dispatches</span>
					</label>
					<label class="flex items-center gap-2 cursor-pointer">
						<input type="checkbox" bind:checked={formIsOtpEnabled} class="h-4 w-4 rounded border-gray-300 text-emerald-600 focus:ring-emerald-500" />
						<span class="text-xs text-gray-700 dark:text-gray-300">Include in high-priority OTP delivery pool</span>
					</label>
					<label class="flex items-center gap-2 cursor-pointer">
						<input type="checkbox" bind:checked={formDlrSupport} class="h-4 w-4 rounded border-gray-300 text-emerald-600 focus:ring-emerald-500" />
						<span class="text-xs text-gray-700 dark:text-gray-300">Support upstream Delivery Receipts (DLR callbacks)</span>
					</label>
				</div>

				<div class="flex justify-end gap-2 border-t border-gray-100 pt-3 dark:border-gray-800">
					<button
						type="button"
						onclick={() => (showModal = false)}
						class="rounded-lg border border-gray-200 px-4 py-2 text-xs font-medium text-gray-700 hover:bg-gray-50 dark:border-gray-700 dark:text-gray-300 dark:hover:bg-gray-800"
					>
						Cancel
					</button>
					<button
						type="submit"
						disabled={saving}
						class="rounded-lg bg-emerald-600 px-4 py-2 text-xs font-semibold text-white shadow-xs hover:bg-emerald-700 disabled:opacity-60"
					>
						{saving ? 'Saving...' : 'Save Gateway'}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
