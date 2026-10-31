<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { supermanApi } from '$lib/superman-api.js';

	interface CustomerRef {
		id: number;
		name: string;
		email?: string | null;
	}

	interface GatewayRef {
		id: number;
		name: string;
	}

	let creationMode = $state<'single' | 'bulk'>('single');
	let senderId = $state('');
	let bulkSenderText = $state('');
	let selectedGatewayId = $state<number | null>(null);
	let selectedType = $state('alphanumeric');
	let isEnabled = $state(true);
	let selectedCustomerIds = $state<number[]>([]);
	let customerSearch = $state('');

	let gateways = $state<GatewayRef[]>([]);
	let customers = $state<CustomerRef[]>([]);
	let loading = $state(false);
	let submitting = $state(false);
	let errorMessage = $state<string | null>(null);
	let successMessage = $state<string | null>(null);

	async function loadFormOptions() {
		loading = true;
		try {
			const res = await supermanApi.get<{
				data: {
					gateways: GatewayRef[];
					customers: CustomerRef[];
				};
			}>('/api/v1/superman/sender-cli/options');
			gateways = res.data.gateways || [];
			customers = res.data.customers || [];
			if (gateways.length > 0 && selectedGatewayId === null) {
				selectedGatewayId = gateways[0].id;
			}
		} catch (err: unknown) {
			errorMessage = err instanceof Error ? err.message : 'Unable to load gateways and customer options.';
		} finally {
			loading = false;
		}
	}

	function toggleCustomer(id: number) {
		if (selectedCustomerIds.includes(id)) {
			selectedCustomerIds = selectedCustomerIds.filter((cid) => cid !== id);
		} else {
			selectedCustomerIds = [...selectedCustomerIds, id];
		}
	}

	function handleFileUpload(event: Event) {
		const input = event.target as HTMLInputElement;
		if (!input.files || input.files.length === 0) return;
		const file = input.files[0];
		const reader = new FileReader();
		reader.onload = (e) => {
			const text = String(e.target?.result || '');
			bulkSenderText = text;
		};
		reader.readAsText(file);
	}

	async function handleSubmit() {
		errorMessage = null;
		successMessage = null;

		if (!selectedGatewayId) {
			errorMessage = 'Please choose a gateway provider.';
			return;
		}

		let sendersToCreate: string[] = [];
		if (creationMode === 'single') {
			if (!senderId.trim()) {
				errorMessage = 'Sender ID is required.';
				return;
			}
			sendersToCreate = [senderId.trim()];
		} else {
			// Extract senders from bulk text (comma, newline, or tab separated)
			const rawTokens = bulkSenderText
				.split(/[\r\n,;\t]+/)
				.map((t) => t.trim())
				.filter((t) => t.length > 0);
			// Remove headers if present (e.g., 'sender_id', 'cli')
			sendersToCreate = rawTokens.filter((t) => !['sender_id', 'cli', 'sender', 'senderid'].includes(t.toLowerCase()));
			if (sendersToCreate.length === 0) {
				errorMessage = 'Please enter or upload at least one sender ID for bulk creation.';
				return;
			}
		}

		submitting = true;
		try {
			const payload = {
				sender_id: sendersToCreate[0],
				sender_ids: sendersToCreate.length > 1 ? sendersToCreate.slice(1) : [],
				gateway_id: selectedGatewayId,
				type: selectedType,
				is_enabled: isEnabled,
				user_ids: selectedCustomerIds
			};

			const res = await supermanApi.post<{ message: string; data: { created: number; updated: number; id: number } }>(
				'/api/v1/superman/senders',
				payload
			);

			successMessage = res.message || `${sendersToCreate.length} sender(s) created successfully.`;
			setTimeout(() => {
				if (selectedType === 'numeric') {
					void goto('/superman/sender-cli/non-masking');
				} else {
					void goto('/superman/sender-cli/masking');
				}
			}, 1200);
		} catch (err: unknown) {
			errorMessage = err instanceof Error ? err.message : 'Unable to create sender ID(s).';
		} finally {
			submitting = false;
		}
	}

	let filteredCustomers = $derived(
		customers.filter((c) => {
			if (!customerSearch.trim()) return true;
			const q = customerSearch.toLowerCase();
			return c.name.toLowerCase().includes(q) || (c.email && c.email.toLowerCase().includes(q));
		})
	);

	onMount(() => {
		void loadFormOptions();
	});
</script>

<svelte:head>
	<title>Add Sender — SMS Whiteinfo Administrator</title>
</svelte:head>

<div class="max-w-4xl space-y-6">
	<!-- Page Header -->
	<div class="flex items-center justify-between">
		<div>
			<div class="flex items-center gap-2">
				<a href="/superman/sender-cli/masking" class="text-xs text-gray-400 hover:text-gray-600 dark:hover:text-gray-200">&larr; Back to Senders</a>
			</div>
			<h1 class="mt-1 text-2xl font-bold tracking-tight text-gray-950 dark:text-white">Add Sender ID / CLI</h1>
			<p class="text-xs text-gray-500 dark:text-gray-400">
				Register new alphanumeric brand headers or numeric sender routes, bind gateway dispatch and assign customer access.
			</p>
		</div>
	</div>

	{#if errorMessage}
		<div class="rounded-lg border border-rose-200 bg-rose-50 p-4 text-xs text-rose-700 dark:border-rose-900 dark:bg-rose-950/30 dark:text-rose-300">
			{errorMessage}
		</div>
	{/if}

	{#if successMessage}
		<div class="rounded-lg border border-emerald-200 bg-emerald-50 p-4 text-xs text-emerald-700 dark:border-emerald-900 dark:bg-emerald-950/30 dark:text-emerald-300">
			{successMessage}
		</div>
	{/if}

	<!-- Form Card -->
	<form onsubmit={(e) => { e.preventDefault(); void handleSubmit(); }} class="rounded-xl border border-gray-200 bg-white p-6 shadow-xs dark:border-gray-800 dark:bg-gray-900 space-y-6">
		<!-- Creation Mode Switcher -->
		<div>
			<label class="block text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-2">
				Creation Mode
			</label>
			<div class="grid grid-cols-2 gap-3 max-w-md">
				<button
					type="button"
					onclick={() => (creationMode = 'single')}
					class="flex items-center justify-center gap-2 rounded-lg py-2.5 px-4 text-xs font-semibold transition {creationMode === 'single' ? 'bg-emerald-600 text-white shadow-xs' : 'border border-gray-200 bg-gray-50 text-gray-700 hover:bg-gray-100 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-300'}"
				>
					<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" d="M15.75 6a3.75 3.75 0 1 1-7.5 0 3.75 3.75 0 0 1 7.5 0ZM4.501 20.118a7.5 7.5 0 0 1 14.998 0A17.933 17.933 0 0 1 12 21.75c-2.676 0-5.216-.584-7.499-1.632Z" />
					</svg>
					Single Sender
				</button>
				<button
					type="button"
					onclick={() => (creationMode = 'bulk')}
					class="flex items-center justify-center gap-2 rounded-lg py-2.5 px-4 text-xs font-semibold transition {creationMode === 'bulk' ? 'bg-emerald-600 text-white shadow-xs' : 'border border-gray-200 bg-gray-50 text-gray-700 hover:bg-gray-100 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-300'}"
				>
					<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5m-13.5-9L12 3m0 0 4.5 4.5M12 3v13.5" />
					</svg>
					Bulk Import (CSV / Sheet)
				</button>
			</div>
		</div>

		<!-- Main Inputs Grid -->
		<div class="grid grid-cols-1 gap-5 md:grid-cols-2">
			<!-- Gateway Selector -->
			<div>
				<label class="block text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">
					{creationMode === 'bulk' ? 'Default Gateway Channel' : 'Gateway Channel'} <span class="text-rose-500">*</span>
				</label>
				<select
					bind:value={selectedGatewayId}
					required
					class="mt-1.5 w-full rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs font-medium text-gray-900 outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
				>
					{#if gateways.length === 0}
						<option value={null}>Loading gateways...</option>
					{:else}
						{#each gateways as gw}
							<option value={gw.id}>{gw.name}</option>
						{/each}
					{/if}
				</select>
			</div>

			<!-- Sender Type -->
			<div>
				<label class="block text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">
					{creationMode === 'bulk' ? 'Default Sender Type' : 'Sender Type'} <span class="text-rose-500">*</span>
				</label>
				<select
					bind:value={selectedType}
					required
					class="mt-1.5 w-full rounded-lg border border-gray-200 bg-white px-3 py-2 text-xs font-medium text-gray-900 outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
				>
					<option value="alphanumeric">Masking (Brand Alphanumeric, 3-11 chars)</option>
					<option value="numeric">Non-Masking (Numeric CLI, Mobile MSISDN)</option>
					<option value="shortcode">Shortcode (4-6 Digits)</option>
				</select>
			</div>

			<!-- Single Mode: Sender ID input -->
			{#if creationMode === 'single'}
				<div class="md:col-span-2">
					<label class="block text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">
						Sender ID / Header <span class="text-rose-500">*</span>
					</label>
					<input
						type="text"
						bind:value={senderId}
						required
						placeholder={selectedType === 'numeric' ? 'e.g. 8809640913002 or 01712345678' : 'e.g. WHITEINFO or BRANDNAME'}
						class="mt-1.5 w-full rounded-lg border border-gray-200 bg-white px-3 py-2 font-mono text-sm text-gray-900 outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
					/>
					<p class="mt-1 text-[11px] text-gray-400">
						{selectedType === 'numeric' ? 'Enter a valid numeric CLI digits.' : 'Must be 3 to 11 alphanumeric characters. Cannot be purely numeric.'}
					</p>
				</div>
			{:else}
				<!-- Bulk Mode: Textarea or CSV upload -->
				<div class="md:col-span-2 space-y-3">
					<div>
						<label class="block text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">
							Upload CSV / Text File
						</label>
						<input
							type="file"
							accept=".csv,.txt"
							onchange={handleFileUpload}
							class="mt-1.5 block w-full text-xs text-gray-500 file:mr-3 file:rounded-lg file:border-0 file:bg-emerald-50 file:px-3 file:py-1.5 file:text-xs file:font-semibold file:text-emerald-700 hover:file:bg-emerald-100 dark:file:bg-emerald-950/40 dark:file:text-emerald-300"
						/>
					</div>

					<div>
						<label class="block text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">
							Or Paste Sender IDs (One per line or comma-separated)
						</label>
						<textarea
							bind:value={bulkSenderText}
							rows="5"
							placeholder="WHITEINFO&#10;INFOZILLION&#10;8809640913002&#10;8809640913003"
							class="mt-1.5 w-full rounded-lg border border-gray-200 bg-white p-3 font-mono text-xs text-gray-900 outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
						></textarea>
					</div>
				</div>
			{/if}

			<!-- Enabled Switch -->
			<div>
				<label class="block text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">
					Operating Status
				</label>
				<div class="mt-2 flex items-center gap-2">
					<input
						type="checkbox"
						id="is_enabled"
						bind:checked={isEnabled}
						class="h-4 w-4 rounded border-gray-300 text-emerald-600 focus:ring-emerald-500"
					/>
					<label for="is_enabled" class="text-xs text-gray-700 dark:text-gray-300 cursor-pointer">
						Immediately enable sender route after creation
					</label>
				</div>
			</div>
		</div>

		<!-- Customer Access Assignment Section -->
		<div class="border-t border-gray-100 pt-5 dark:border-gray-800">
			<div class="flex items-center justify-between mb-2">
				<div>
					<h3 class="text-xs font-bold uppercase tracking-wider text-gray-900 dark:text-white">Assigned Customers (Optional)</h3>
					<p class="text-[11px] text-gray-400">Select customer users who can dispatch SMS using this sender ID.</p>
				</div>
				<span class="text-xs font-semibold text-emerald-600">{selectedCustomerIds.length} customer(s) selected</span>
			</div>

			<div class="mt-2 space-y-2">
				<input
					type="text"
					bind:value={customerSearch}
					placeholder="Search customers to assign..."
					class="w-full rounded-lg border border-gray-200 bg-gray-50/70 px-3 py-1.5 text-xs outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
				/>

				<div class="max-h-56 overflow-y-auto rounded-lg border border-gray-200 divide-y divide-gray-100 dark:border-gray-800 dark:divide-gray-800 p-1">
					{#if filteredCustomers.length === 0}
						<div class="p-3 text-center text-xs text-gray-400">No matching customers found.</div>
					{:else}
						{#each filteredCustomers as customer}
							<label class="flex items-center gap-2.5 p-2 rounded hover:bg-gray-50 dark:hover:bg-gray-800/40 cursor-pointer">
								<input
									type="checkbox"
									checked={selectedCustomerIds.includes(customer.id)}
									onchange={() => toggleCustomer(customer.id)}
									class="h-4 w-4 rounded border-gray-300 text-emerald-600 focus:ring-emerald-500"
								/>
								<div class="flex-1 min-w-0">
									<p class="text-xs font-medium text-gray-900 dark:text-white truncate">{customer.name}</p>
									{#if customer.email}
										<p class="text-[11px] text-gray-400 truncate">{customer.email}</p>
									{/if}
								</div>
							</label>
						{/each}
					{/if}
				</div>
			</div>
		</div>

		<!-- Submit Actions -->
		<div class="flex items-center justify-end gap-3 border-t border-gray-100 pt-5 dark:border-gray-800">
			<a
				href="/superman/sender-cli/masking"
				class="rounded-lg border border-gray-200 bg-white px-4 py-2 text-xs font-semibold text-gray-700 hover:bg-gray-50 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-300 transition"
			>
				Cancel
			</a>
			<button
				type="submit"
				disabled={submitting}
				class="inline-flex items-center gap-1.5 rounded-lg bg-emerald-600 px-5 py-2 text-xs font-semibold text-white shadow-xs hover:bg-emerald-500 disabled:opacity-50 transition"
			>
				{#if submitting}
					<svg class="h-3.5 w-3.5 animate-spin" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0 3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182m0-4.991v4.99" />
					</svg>
				{/if}
				Save & Register Sender ID
			</button>
		</div>
	</form>
</div>

