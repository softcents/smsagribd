<script lang="ts">
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { supermanApi } from '$lib/superman-api.js';
	import ExportDropdown from '$lib/components/common/ExportDropdown.svelte';
	import type { ExportColumn } from '$lib/utils/export.js';

	const customerId = page.params.id;

	interface Customer {
		id: number;
		name: string;
		email: string;
		phone: string;
		business_name: string | null;
		balance: number;
		kyc_status: string;
		status: string;
		nid_number?: string | null;
		tariff?: string | null;
		created_at: string;
		last_login_at?: string | null;
	}

	interface CustomerDetail {
		customer: Customer;
		company_name: string | null;
		nid_number: string | null;
		front_nid: string | null;
		back_nid: string | null;
		live_photo: string | null;
		business_address: string | null;
		trade_license_number: string | null;
		trade_license_file: string | null;
		kyc_submitted_at: string | null;
		kyc_verified_at: string | null;
		api_access?: boolean;
		sms_auto_approved?: boolean;
		tariff_plan_id?: number | null;
		senders?: Array<{ id: number; sender_id: string; type: string; gateway_name: string }>;
		ip_whitelists?: Array<{ id: number; ip_address: string; status: string }>;
		recent_transactions?: Array<{ id: number; amount: number; direction: string; balance_before: number; balance_after: number; description: string | null; created_at: string }>;
	}

	let detail = $state<CustomerDetail | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);

	type CustomerTx = NonNullable<CustomerDetail['recent_transactions']>[number];
	const exportTransactionColumns: ExportColumn<CustomerTx>[] = [
		{ header: 'ID', getValue: (t) => t.id },
		{ header: 'Direction', getValue: (t) => t.direction },
		{ header: 'Amount (BDT)', getValue: (t) => `BDT ${t.amount.toFixed(2)}` },
		{ header: 'Balance Before (BDT)', getValue: (t) => `BDT ${t.balance_before.toFixed(2)}` },
		{ header: 'Balance After (BDT)', getValue: (t) => `BDT ${t.balance_after.toFixed(2)}` },
		{ header: 'Description', getValue: (t) => t.description || '—' },
		{ header: 'Date', getValue: (t) => t.created_at }
	];

	// Modals
	let depositModalOpen = $state(false);
	let depositAmount = $state(500);
	let depositDescription = $state('');
	let depositSubmitting = $state(false);

	let editModalOpen = $state(false);
	let editData = $state({
		name: '',
		email: '',
		phone: '',
		business_name: '',
		status: 'approved',
		api_access: false,
		sms_auto_approved: false
	});
	let editSubmitting = $state(false);

	let zoomedImage = $state<{ src: string; title: string } | null>(null);

	async function loadCustomer() {
		loading = true;
		error = null;
		try {
			const res = await supermanApi.get<{ data: CustomerDetail }>(`/api/v1/superman/customers/${customerId}`);
			detail = res.data;
			editData = {
				name: detail.customer.name,
				email: detail.customer.email,
				phone: detail.customer.phone,
				business_name: detail.customer.business_name ?? '',
				status: detail.customer.status,
				api_access: Boolean(detail.api_access),
				sms_auto_approved: Boolean(detail.sms_auto_approved)
			};
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Unable to load customer details.';
		} finally {
			loading = false;
		}
	}

	async function submitDeposit() {
		if (depositAmount === 0) return;
		depositSubmitting = true;
		try {
			const res = await supermanApi.post<{ message: string; data: { balance: number } }>(
				`/api/v1/superman/customers/${customerId}/deposit`,
				{ amount: depositAmount, description: depositDescription }
			);
			if (detail) {
				detail.customer.balance = res.data.balance;
			}
			depositModalOpen = false;
			window.alert(res.message);
			void loadCustomer();
		} catch (err: unknown) {
			window.alert(err instanceof Error ? err.message : 'Unable to adjust balance.');
		} finally {
			depositSubmitting = false;
		}
	}

	async function submitEdit() {
		editSubmitting = true;
		try {
			await supermanApi.put(`/api/v1/superman/customers/${customerId}`, editData);
			editModalOpen = false;
			window.alert('Customer updated successfully.');
			void loadCustomer();
		} catch (err: unknown) {
			window.alert(err instanceof Error ? err.message : 'Unable to update customer.');
		} finally {
			editSubmitting = false;
		}
	}

	async function approveKYC() {
		if (!confirm('Approve KYC verification for this customer?')) return;
		try {
			await supermanApi.post(`/api/v1/superman/customers/${customerId}/approve-kyc`);
			window.alert('KYC verified successfully.');
			void loadCustomer();
		} catch (err: unknown) {
			window.alert(err instanceof Error ? err.message : 'Unable to approve KYC.');
		}
	}

	async function rejectKYC() {
		const reason = prompt('Please provide reason for rejecting KYC:');
		if (!reason) return;
		try {
			await supermanApi.post(`/api/v1/superman/customers/${customerId}/reject-kyc`, {
				reason,
				admin_remarks: 'Rejected from customer detail page'
			});
			window.alert('KYC rejected.');
			void loadCustomer();
		} catch (err: unknown) {
			window.alert(err instanceof Error ? err.message : 'Unable to reject KYC.');
		}
	}

	async function requestResubmit() {
		const reason = prompt('Please provide reason for KYC resubmission:');
		if (!reason) return;
		try {
			await supermanApi.post(`/api/v1/superman/customers/${customerId}/resubmit-kyc`, {
				reason,
				fields: 'front_nid, back_nid, live_photo'
			});
			window.alert('Resubmission requested from customer.');
			void loadCustomer();
		} catch (err: unknown) {
			window.alert(err instanceof Error ? err.message : 'Unable to request resubmission.');
		}
	}

	onMount(() => {
		void loadCustomer();
	});
</script>

<svelte:head>
	<title>{detail ? `${detail.customer.name} — Customer Profile` : 'Customer Profile'} — SMS Whiteinfo</title>
</svelte:head>

<div class="space-y-6 text-gray-900 dark:text-white antialiased font-sans">
	<!-- Breadcrumb -->
	<nav class="flex items-center gap-1.5 text-xs text-gray-500 dark:text-gray-400">
		<a href="/superman/users/customers" class="hover:text-emerald-600 transition-colors">Customers</a>
		<span>›</span>
		<span class="font-medium text-gray-900 dark:text-white">{detail?.customer.name ?? `Customer #${customerId}`}</span>
	</nav>

	{#if loading}
		<div class="rounded-2xl bg-white p-16 text-center dark:bg-gray-900 ring-1 ring-gray-950/5 dark:ring-white/10">
			<div class="inline-block h-8 w-8 animate-spin rounded-full border-2 border-emerald-600 border-t-transparent"></div>
			<p class="mt-3 text-xs text-gray-500">Loading customer account, KYC documents, and permissions...</p>
		</div>
	{:else if error || !detail}
		<div class="rounded-2xl bg-white p-8 text-center dark:bg-gray-900 ring-1 ring-gray-950/5">
			<p class="text-rose-600 text-sm">{error || 'Customer not found.'}</p>
			<a href="/superman/users/customers" class="mt-3 inline-block text-xs font-semibold text-emerald-600 hover:underline">
				← Back to Customers list
			</a>
		</div>
	{:else}
		<!-- Customer Profile Hero Header -->
		<div class="rounded-2xl bg-white p-6 shadow-xs ring-1 ring-gray-950/5 dark:bg-gray-900 dark:ring-white/10">
			<div class="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-4">
				<div class="flex items-start gap-4">
					<div class="flex h-14 w-14 shrink-0 items-center justify-center rounded-2xl bg-emerald-100 text-emerald-800 dark:bg-emerald-950/60 dark:text-emerald-300 text-xl font-bold">
						{detail.customer.name.slice(0, 1).toUpperCase()}
					</div>
					<div class="space-y-1">
						<div class="flex flex-wrap items-center gap-2">
							<h1 class="text-xl font-bold tracking-tight text-gray-950 dark:text-white">{detail.customer.name}</h1>
							<span class="rounded-md bg-gray-100 px-2 py-0.5 text-[11px] font-mono font-bold text-gray-600 dark:bg-gray-800 dark:text-gray-300">
								ID #{detail.customer.id}
							</span>
							{#if detail.customer.status === 'approved'}
								<span class="inline-flex items-center gap-1 rounded-full bg-emerald-50 px-2.5 py-0.5 text-[11px] font-semibold text-emerald-700 border border-emerald-200 dark:bg-emerald-950/40 dark:border-emerald-800 dark:text-emerald-400">
									<span class="h-1.5 w-1.5 rounded-full bg-emerald-500"></span>
									Account Approved
								</span>
							{:else}
								<span class="inline-flex items-center gap-1 rounded-full bg-amber-50 px-2.5 py-0.5 text-[11px] font-semibold text-amber-700 border border-amber-200 dark:bg-amber-950/40 dark:border-amber-800 dark:text-amber-400">
									<span class="h-1.5 w-1.5 rounded-full bg-amber-500"></span>
									{detail.customer.status}
								</span>
							{/if}
							{#if detail.customer.kyc_status === 'verified'}
								<span class="inline-flex items-center gap-1 rounded-full bg-emerald-50 px-2.5 py-0.5 text-[11px] font-semibold text-emerald-700 border border-emerald-200 dark:bg-emerald-950/40 dark:border-emerald-800 dark:text-emerald-400">
									<span class="h-1.5 w-1.5 rounded-full bg-emerald-500"></span>
									KYC Verified
								</span>
							{:else}
								<span class="inline-flex items-center gap-1 rounded-full bg-amber-50 px-2.5 py-0.5 text-[11px] font-semibold text-amber-700 border border-amber-200 dark:bg-amber-950/40 dark:border-amber-800 dark:text-amber-400">
									<span class="h-1.5 w-1.5 rounded-full bg-amber-500"></span>
									KYC: {detail.customer.kyc_status}
								</span>
							{/if}
						</div>
						<div class="flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-gray-500 dark:text-gray-400">
							<span>✉ {detail.customer.email}</span>
							<span>📞 {detail.customer.phone || 'No phone'}</span>
							{#if detail.customer.business_name}
								<span>🏢 {detail.customer.business_name}</span>
							{/if}
						</div>
					</div>
				</div>

				<!-- Header Actions -->
				<div class="flex flex-wrap items-center gap-2">
					<ExportDropdown
						title={`${detail.customer.name} - Transactions Ledger`}
						columns={exportTransactionColumns}
						getData={() => detail?.recent_transactions ?? []}
						disabled={!detail?.recent_transactions || detail.recent_transactions.length === 0}
					/>
					<button
						type="button"
						onclick={() => (depositModalOpen = true)}
						class="inline-flex items-center gap-1.5 rounded-lg bg-emerald-600 px-3.5 py-2 text-xs font-bold text-white shadow-xs hover:bg-emerald-700 transition"
					>
						+ Make Deposit
					</button>
					<button
						type="button"
						onclick={() => (editModalOpen = true)}
						class="inline-flex items-center gap-1.5 rounded-lg border border-gray-300 bg-white px-3.5 py-2 text-xs font-semibold text-gray-700 shadow-2xs hover:bg-gray-50 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-200 transition"
					>
						✏ Edit Profile
					</button>
					{#if detail.customer.kyc_status !== 'verified'}
						<button
							type="button"
							onclick={approveKYC}
							class="inline-flex items-center gap-1.5 rounded-lg bg-emerald-700 px-3 py-2 text-xs font-bold text-white hover:bg-emerald-800 transition"
						>
							Approve KYC
						</button>
						<button
							type="button"
							onclick={requestResubmit}
							class="inline-flex items-center gap-1.5 rounded-lg bg-orange-600 px-3 py-2 text-xs font-bold text-white hover:bg-orange-700 transition"
						>
							Request Resubmit
						</button>
						<button
							type="button"
							onclick={rejectKYC}
							class="inline-flex items-center gap-1.5 rounded-lg bg-rose-600 px-3 py-2 text-xs font-bold text-white hover:bg-rose-700 transition"
						>
							Reject KYC
						</button>
					{/if}
				</div>
			</div>
		</div>

		<!-- Section 1: Customer Details & Financials -->
		<div class="grid grid-cols-1 md:grid-cols-3 gap-6">
			<!-- Wallet Card -->
			<div class="rounded-xl bg-white p-5 shadow-xs ring-1 ring-gray-950/5 dark:bg-gray-900 dark:ring-white/10 flex flex-col justify-between">
				<div>
					<span class="text-xs font-semibold text-gray-500 uppercase tracking-wider">Current Wallet Balance</span>
					<div class="mt-2 text-3xl font-extrabold text-emerald-600 dark:text-emerald-400 font-mono tracking-tight">
						৳{detail.customer.balance.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
					</div>
					<p class="mt-1 text-xs text-gray-400">Available funds for SMS sending and service charges</p>
				</div>
				<button
					type="button"
					onclick={() => (depositModalOpen = true)}
					class="mt-4 w-full rounded-lg bg-emerald-50 py-2 text-xs font-bold text-emerald-700 hover:bg-emerald-100 dark:bg-emerald-950/40 dark:text-emerald-300 text-center transition"
				>
					+ Adjust Balance
				</button>
			</div>

			<!-- Account Details -->
			<div class="md:col-span-2 rounded-xl bg-white p-5 shadow-xs ring-1 ring-gray-950/5 dark:bg-gray-900 dark:ring-white/10 space-y-4">
				<h3 class="text-xs font-bold uppercase tracking-wider text-gray-700 dark:text-gray-300">Account Credentials &amp; Details</h3>
				<div class="grid grid-cols-2 sm:grid-cols-3 gap-4 text-xs">
					<div>
						<span class="text-gray-400 font-medium">Full Name</span>
						<p class="font-semibold text-gray-900 dark:text-white mt-0.5">{detail.customer.name}</p>
					</div>
					<div>
						<span class="text-gray-400 font-medium">Email Address</span>
						<p class="font-semibold text-gray-900 dark:text-white mt-0.5">{detail.customer.email}</p>
					</div>
					<div>
						<span class="text-gray-400 font-medium">Phone Number</span>
						<p class="font-mono font-semibold text-gray-900 dark:text-white mt-0.5">{detail.customer.phone || '—'}</p>
					</div>
					<div>
						<span class="text-gray-400 font-medium">Registered Date</span>
						<p class="font-semibold text-gray-700 dark:text-gray-300 mt-0.5">{new Date(detail.customer.created_at).toLocaleDateString()}</p>
					</div>
					<div>
						<span class="text-gray-400 font-medium">Last Login</span>
						<p class="font-semibold text-gray-700 dark:text-gray-300 mt-0.5">{detail.customer.last_login_at || 'Never'}</p>
					</div>
					<div>
						<span class="text-gray-400 font-medium">Business / Company</span>
						<p class="font-semibold text-gray-700 dark:text-gray-300 mt-0.5">{detail.customer.business_name || detail.company_name || '—'}</p>
					</div>
				</div>
			</div>
		</div>

		<!-- Section 2: Identity & KYC Verification Documents (Front NID, Back NID, Selfie) -->
		<div class="rounded-xl bg-white p-6 shadow-xs ring-1 ring-gray-950/5 dark:bg-gray-900 dark:ring-white/10 space-y-5">
			<div class="flex items-center justify-between pb-3 border-b border-gray-100 dark:border-gray-800">
				<div>
					<h3 class="text-sm font-bold text-gray-900 dark:text-white">Identity Verification &amp; Documents</h3>
					<p class="text-xs text-gray-500 mt-0.5">National ID verification and live photo selfie</p>
				</div>
				<div class="flex items-center gap-3 text-xs">
					<span class="text-gray-500">NID: <strong class="font-mono font-bold text-gray-900 dark:text-white">{detail.nid_number || 'Not Submitted'}</strong></span>
				</div>
			</div>

			<div class="grid grid-cols-1 sm:grid-cols-3 gap-6">
				<!-- Front Side of NID -->
				<div class="space-y-2">
					<div class="flex items-center justify-between">
						<span class="text-xs font-semibold text-gray-700 dark:text-gray-300">Front Side of NID</span>
						{#if detail.front_nid}
							<button
								type="button"
								onclick={() => (zoomedImage = { src: detail!.front_nid!, title: 'Front Side of NID' })}
								class="text-[11px] font-semibold text-emerald-600 hover:underline"
							>
								Zoom 🔍
							</button>
						{/if}
					</div>
					<div class="h-52 rounded-xl border border-gray-200 bg-gray-50 dark:border-gray-700 dark:bg-gray-800 flex items-center justify-center p-2 overflow-hidden shadow-2xs">
						{#if detail.front_nid}
							<img src={detail.front_nid.startsWith('http') ? detail.front_nid : `/storage/${detail.front_nid}`} alt="Front NID" class="max-h-full max-w-full object-contain rounded-lg cursor-pointer" onclick={() => (zoomedImage = { src: detail!.front_nid!, title: 'Front Side of NID' })} />
						{:else}
							<span class="text-xs text-gray-400 italic">No document uploaded</span>
						{/if}
					</div>
				</div>

				<!-- Back Side of NID -->
				<div class="space-y-2">
					<div class="flex items-center justify-between">
						<span class="text-xs font-semibold text-gray-700 dark:text-gray-300">Back Side of NID</span>
						{#if detail.back_nid}
							<button
								type="button"
								onclick={() => (zoomedImage = { src: detail!.back_nid!, title: 'Back Side of NID' })}
								class="text-[11px] font-semibold text-emerald-600 hover:underline"
							>
								Zoom 🔍
							</button>
						{/if}
					</div>
					<div class="h-52 rounded-xl border border-gray-200 bg-gray-50 dark:border-gray-700 dark:bg-gray-800 flex items-center justify-center p-2 overflow-hidden shadow-2xs">
						{#if detail.back_nid}
							<img src={detail.back_nid.startsWith('http') ? detail.back_nid : `/storage/${detail.back_nid}`} alt="Back NID" class="max-h-full max-w-full object-contain rounded-lg cursor-pointer" onclick={() => (zoomedImage = { src: detail!.back_nid!, title: 'Back Side of NID' })} />
						{:else}
							<span class="text-xs text-gray-400 italic">No document uploaded</span>
						{/if}
					</div>
				</div>

				<!-- Live Selfie -->
				<div class="space-y-2">
					<div class="flex items-center justify-between">
						<span class="text-xs font-semibold text-gray-700 dark:text-gray-300">Live Camera Selfie</span>
						{#if detail.live_photo}
							<button
								type="button"
								onclick={() => (zoomedImage = { src: detail!.live_photo!, title: 'Live Camera Selfie' })}
								class="text-[11px] font-semibold text-emerald-600 hover:underline"
							>
								Zoom 🔍
							</button>
						{/if}
					</div>
					<div class="h-52 rounded-xl border border-gray-200 bg-gray-50 dark:border-gray-700 dark:bg-gray-800 flex items-center justify-center p-2 overflow-hidden shadow-2xs">
						{#if detail.live_photo}
							<img src={detail.live_photo.startsWith('http') ? detail.live_photo : `/storage/${detail.live_photo}`} alt="Selfie" class="max-h-full max-w-full object-contain rounded-lg cursor-pointer" onclick={() => (zoomedImage = { src: detail!.live_photo!, title: 'Live Camera Selfie' })} />
						{:else}
							<span class="text-xs text-gray-400 italic">No live photo captured</span>
						{/if}
					</div>
				</div>
			</div>
		</div>

		<!-- Section 3: SMS Permissions & Senders & IP Whitelist -->
		<div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
			<!-- Permissions & Tariff -->
			<div class="rounded-xl bg-white p-5 shadow-xs ring-1 ring-gray-950/5 dark:bg-gray-900 dark:ring-white/10 space-y-4">
				<h3 class="text-xs font-bold uppercase tracking-wider text-gray-700 dark:text-gray-300">SMS Permissions &amp; Tariff Plan</h3>
				<div class="space-y-3">
					<div class="flex items-center justify-between p-3 rounded-lg bg-gray-50 dark:bg-gray-800/40">
						<div>
							<span class="text-xs font-semibold text-gray-900 dark:text-white">API Access</span>
							<p class="text-[11px] text-gray-500">Allow sending messages via Developer REST API</p>
						</div>
						<span class="rounded px-2 py-0.5 text-xs font-bold {detail.api_access ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300' : 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400'}">
							{detail.api_access ? 'Enabled' : 'Disabled'}
						</span>
					</div>

					<div class="flex items-center justify-between p-3 rounded-lg bg-gray-50 dark:bg-gray-800/40">
						<div>
							<span class="text-xs font-semibold text-gray-900 dark:text-white">Bulk SMS Auto-Approve</span>
							<p class="text-[11px] text-gray-500">Dispatches campaigns directly without manual Superman approval</p>
						</div>
						<span class="rounded px-2 py-0.5 text-xs font-bold {detail.sms_auto_approved ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300' : 'bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-300'}">
							{detail.sms_auto_approved ? 'Auto-Approved' : 'Manual Review'}
						</span>
					</div>

					<div class="p-3 rounded-lg bg-gray-50 dark:bg-gray-800/40">
						<span class="text-xs font-semibold text-gray-900 dark:text-white">Active Tariff Plan</span>
						<p class="text-xs font-bold text-sky-700 dark:text-sky-300 mt-1">{detail.customer.tariff || 'Default Operator Plan'}</p>
					</div>
				</div>
			</div>

			<!-- Assigned Senders & IP Whitelist -->
			<div class="rounded-xl bg-white p-5 shadow-xs ring-1 ring-gray-950/5 dark:bg-gray-900 dark:ring-white/10 space-y-4">
				<h3 class="text-xs font-bold uppercase tracking-wider text-gray-700 dark:text-gray-300">Assigned Senders &amp; IP Whitelist</h3>
				<div>
					<span class="text-xs font-medium text-gray-500">Assigned Sender IDs</span>
					<div class="mt-2 flex flex-wrap gap-1.5">
						{#if detail.senders && detail.senders.length > 0}
							{#each detail.senders as s}
								<span class="inline-flex items-center gap-1 rounded bg-sky-50 px-2 py-1 text-xs font-semibold text-sky-700 border border-sky-200 dark:bg-sky-950/40 dark:border-sky-800 dark:text-sky-300">
									{s.sender_id} <span class="text-[10px] text-gray-400">({s.gateway_name})</span>
								</span>
							{/each}
						{:else}
							<span class="text-xs text-gray-400 italic">No custom senders assigned.</span>
						{/if}
					</div>
				</div>

				<div class="pt-3 border-t border-gray-100 dark:border-gray-800">
					<span class="text-xs font-medium text-gray-500">Approved API IP Whitelist</span>
					<div class="mt-2 flex flex-wrap gap-1.5">
						{#if detail.ip_whitelists && detail.ip_whitelists.length > 0}
							{#each detail.ip_whitelists as ip}
								<span class="inline-flex items-center gap-1 rounded bg-indigo-50 px-2 py-1 font-mono text-xs font-semibold text-indigo-700 border border-indigo-200 dark:bg-indigo-950/40 dark:border-indigo-800 dark:text-indigo-300">
									{ip.ip_address}
								</span>
							{/each}
						{:else}
							<span class="text-xs text-amber-600 dark:text-amber-400 font-medium">No dedicated IPs whitelisted. Permissive policy applies.</span>
						{/if}
					</div>
				</div>
			</div>
		</div>

		<!-- Section 4: Recent Wallet Transactions -->
		<div class="rounded-xl bg-white shadow-xs ring-1 ring-gray-950/5 dark:bg-gray-900 dark:ring-white/10 overflow-hidden">
			<div class="px-5 py-3.5 border-b border-gray-100 dark:border-gray-800 flex items-center justify-between">
				<h3 class="text-xs font-bold uppercase tracking-wider text-gray-700 dark:text-gray-300">Recent Transactions &amp; Deposits</h3>
				<ExportDropdown
					title={`${detail.customer.name} - Recent Transactions`}
					columns={exportTransactionColumns}
					getData={() => detail?.recent_transactions ?? []}
					disabled={!detail?.recent_transactions || detail.recent_transactions.length === 0}
				/>
			</div>
			<div class="overflow-x-auto">
				<table class="w-full text-xs text-left">
					<thead class="bg-gray-50 dark:bg-gray-800/50 text-[10px] uppercase font-bold text-gray-400">
						<tr>
							<th class="px-4 py-2.5">ID</th>
							<th class="px-4 py-2.5">Amount</th>
							<th class="px-4 py-2.5">Direction</th>
							<th class="px-4 py-2.5">Balance After</th>
							<th class="px-4 py-2.5">Description</th>
							<th class="px-4 py-2.5 text-right">Date</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
						{#if !detail.recent_transactions || detail.recent_transactions.length === 0}
							<tr>
								<td colspan="6" class="px-4 py-6 text-center text-xs text-gray-400">No transaction records found.</td>
							</tr>
						{:else}
							{#each detail.recent_transactions as tx}
								<tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30">
									<td class="px-4 py-2 font-mono text-gray-400">#{tx.id}</td>
									<td class="px-4 py-2 font-mono font-bold {tx.direction === 'credit' ? 'text-emerald-600' : 'text-rose-600'}">
										{tx.direction === 'credit' ? '+' : '-'}৳{tx.amount.toFixed(2)}
									</td>
									<td class="px-4 py-2 uppercase text-[10px] font-semibold">{tx.direction}</td>
									<td class="px-4 py-2 font-mono">৳{tx.balance_after.toFixed(2)}</td>
									<td class="px-4 py-2 text-gray-600 dark:text-gray-400 truncate max-w-xs">{tx.description || '—'}</td>
									<td class="px-4 py-2 text-right font-mono text-gray-400">{new Date(tx.created_at).toLocaleString()}</td>
								</tr>
							{/each}
						{/if}
					</tbody>
				</table>
			</div>
		</div>
	{/if}
</div>

<!-- ============================================================
     ZOOM IMAGE MODAL
     ============================================================ -->
{#if zoomedImage}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-gray-950/80 p-4 backdrop-blur-sm cursor-zoom-out"
		onclick={() => (zoomedImage = null)}
		role="button"
		tabindex="-1"
		onkeydown={(e) => e.key === 'Escape' && (zoomedImage = null)}
	>
		<div class="relative max-h-[90vh] max-w-4xl rounded-2xl overflow-hidden bg-white dark:bg-gray-900 p-2 shadow-2xl" onclick={(e) => e.stopPropagation()} role="dialog">
			<div class="flex items-center justify-between px-4 py-2 border-b border-gray-100 dark:border-gray-800">
				<h4 class="text-xs font-bold text-gray-900 dark:text-white">{zoomedImage.title}</h4>
				<button type="button" onclick={() => (zoomedImage = null)} class="text-xs text-gray-400 hover:text-gray-700">✕ Close</button>
			</div>
			<div class="p-3 flex items-center justify-center">
				<img src={zoomedImage.src.startsWith('http') ? zoomedImage.src : `/storage/${zoomedImage.src}`} alt={zoomedImage.title} class="max-h-[75vh] max-w-full object-contain rounded" />
			</div>
		</div>
	</div>
{/if}

<!-- ============================================================
     DEPOSIT MODAL
     ============================================================ -->
{#if depositModalOpen && detail}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-gray-950/60 p-4 backdrop-blur-xs">
		<div class="w-full max-w-md rounded-2xl bg-white p-6 shadow-2xl ring-1 ring-gray-900/10 dark:bg-gray-900 dark:ring-white/10">
			<h3 class="text-base font-bold text-gray-900 dark:text-white">Adjust Customer Balance</h3>
			<p class="text-xs text-gray-500 mt-0.5">{detail.customer.name} (Current: ৳{detail.customer.balance.toFixed(2)})</p>

			<div class="mt-4 space-y-3">
				<div>
					<label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">
						Amount (BDT) <span class="text-gray-400 font-normal">— enter negative value to deduct</span>
					</label>
					<input
						type="number"
						step="0.01"
						bind:value={depositAmount}
						class="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm text-gray-900 shadow-2xs outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
					/>
				</div>

				<div>
					<label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Remarks / Note</label>
					<input
						type="text"
						bind:value={depositDescription}
						placeholder="e.g. Bank deposit reference #12345"
						class="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-xs text-gray-900 shadow-2xs outline-none focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
					/>
				</div>
			</div>

			<div class="mt-6 flex items-center justify-end gap-2">
				<button
					type="button"
					onclick={() => (depositModalOpen = false)}
					class="rounded-lg border border-gray-300 px-3.5 py-2 text-xs font-semibold text-gray-700 hover:bg-gray-50 dark:border-gray-700 dark:text-gray-300"
				>
					Cancel
				</button>
				<button
					type="button"
					disabled={depositSubmitting}
					onclick={submitDeposit}
					class="rounded-lg bg-emerald-600 px-4 py-2 text-xs font-bold text-white shadow-xs hover:bg-emerald-700 disabled:opacity-50"
				>
					{depositSubmitting ? 'Adjusting...' : 'Confirm Balance Adjustment'}
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- ============================================================
     EDIT MODAL
     ============================================================ -->
{#if editModalOpen && detail}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-gray-950/60 p-4 backdrop-blur-xs">
		<div class="w-full max-w-lg rounded-2xl bg-white p-6 shadow-2xl ring-1 ring-gray-900/10 dark:bg-gray-900 dark:ring-white/10">
			<h3 class="text-base font-bold text-gray-900 dark:text-white">Edit Customer Profile</h3>

			<div class="mt-4 space-y-3">
				<div>
					<label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Full Name</label>
					<input
						type="text"
						bind:value={editData.name}
						class="w-full rounded-lg border border-gray-300 px-3 py-1.5 text-xs dark:border-gray-700 dark:bg-gray-800 dark:text-white"
					/>
				</div>

				<div class="grid grid-cols-2 gap-3">
					<div>
						<label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Email</label>
						<input
							type="email"
							bind:value={editData.email}
							class="w-full rounded-lg border border-gray-300 px-3 py-1.5 text-xs dark:border-gray-700 dark:bg-gray-800 dark:text-white"
						/>
					</div>
					<div>
						<label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Phone</label>
						<input
							type="text"
							bind:value={editData.phone}
							class="w-full rounded-lg border border-gray-300 px-3 py-1.5 text-xs dark:border-gray-700 dark:bg-gray-800 dark:text-white"
						/>
					</div>
				</div>

				<div class="grid grid-cols-2 gap-3">
					<div>
						<label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Business Name</label>
						<input
							type="text"
							bind:value={editData.business_name}
							class="w-full rounded-lg border border-gray-300 px-3 py-1.5 text-xs dark:border-gray-700 dark:bg-gray-800 dark:text-white"
						/>
					</div>
					<div>
						<label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Account Status</label>
						<select
							bind:value={editData.status}
							class="w-full rounded-lg border border-gray-300 px-3 py-1.5 text-xs dark:border-gray-700 dark:bg-gray-800 dark:text-white"
						>
							<option value="approved">Approved</option>
							<option value="pending">Pending</option>
							<option value="rejected">Rejected</option>
							<option value="banned">Banned</option>
						</select>
					</div>
				</div>

				<div class="pt-2 border-t border-gray-100 dark:border-gray-800 space-y-2">
					<label class="flex items-center gap-2 cursor-pointer">
						<input type="checkbox" bind:checked={editData.api_access} class="rounded text-emerald-600" />
						<span class="text-xs font-semibold text-gray-800 dark:text-gray-200">API Access Enabled</span>
					</label>
					<label class="flex items-center gap-2 cursor-pointer">
						<input type="checkbox" bind:checked={editData.sms_auto_approved} class="rounded text-emerald-600" />
						<span class="text-xs font-semibold text-gray-800 dark:text-gray-200">Bulk SMS Auto-Approve</span>
					</label>
				</div>
			</div>

			<div class="mt-6 flex items-center justify-end gap-2">
				<button
					type="button"
					onclick={() => (editModalOpen = false)}
					class="rounded-lg border border-gray-300 px-3.5 py-2 text-xs font-semibold text-gray-700 hover:bg-gray-50 dark:border-gray-700 dark:text-gray-300"
				>
					Cancel
				</button>
				<button
					type="button"
					disabled={editSubmitting}
					onclick={submitEdit}
					class="rounded-lg bg-emerald-600 px-4 py-2 text-xs font-bold text-white shadow-xs hover:bg-emerald-700 disabled:opacity-50"
				>
					{editSubmitting ? 'Saving...' : 'Save Profile'}
				</button>
			</div>
		</div>
	</div>
{/if}
