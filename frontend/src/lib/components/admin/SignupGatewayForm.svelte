<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { supermanApi } from '$lib/superman-api.js';

	let { id = 0 }: { id?: number } = $props();
	let editing = $derived(id > 0);

	const DEFAULT_TEMPLATE = 'Your SMS WhiteInfo verification code is: {otp}. Valid for {minutes} minutes.';
	const blank = () => ({
		name: '',
		priority: 1,
		endpoint_url: '',
		http_method: 'POST',
		timeout_seconds: 15,
		access_token: '',
		sender_id: '',
		phone_format: 'local_bd',
		otp_template: DEFAULT_TEMPLATE,
		is_active: true
	});

	let form = $state(blank());
	let hasToken = $state(false);
	let showToken = $state(false);
	let saving = $state(false);
	let loading = $state(false);
	let errorMsg = $state('');

	onMount(async () => {
		if (id > 0) {
			loading = true;
			try {
				const res: any = await supermanApi.get(`/api/v1/superman/signup-gateways/${id}`);
				const d = res.data;
				form = {
					name: d.name,
					priority: d.priority,
					endpoint_url: d.endpoint_url,
					http_method: d.http_method,
					timeout_seconds: d.timeout_seconds,
					access_token: '',
					sender_id: d.sender_id,
					phone_format: d.phone_format,
					otp_template: d.otp_template,
					is_active: d.is_active
				};
				hasToken = Boolean(d.has_token);
			} catch (e: unknown) {
				errorMsg = e instanceof Error ? e.message : 'Unable to load gateway.';
			} finally {
				loading = false;
			}
		}
	});

	async function save(another = false) {
		errorMsg = '';
		saving = true;
		try {
			const payload = { ...form, priority: Number(form.priority), timeout_seconds: Number(form.timeout_seconds) };
			if (editing) {
				await supermanApi.put(`/api/v1/superman/signup-gateways/${id}`, payload);
				await goto('/superman/signup-gateways');
			} else {
				await supermanApi.post('/api/v1/superman/signup-gateways', payload);
				if (another) {
					form = blank();
					window.scrollTo({ top: 0, behavior: 'smooth' });
				} else {
					await goto('/superman/signup-gateways');
				}
			}
		} catch (e: unknown) {
			errorMsg = e instanceof Error ? e.message : 'Unable to save gateway.';
		} finally {
			saving = false;
		}
	}

	const input = 'w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm text-gray-900 focus:border-emerald-500 focus:outline-none focus:ring-1 focus:ring-emerald-500 dark:border-gray-700 dark:bg-gray-900 dark:text-white';
	const label = 'mb-1 block text-sm font-semibold text-gray-800 dark:text-gray-200';
</script>

<div class="space-y-6">
	<div>
		<nav class="mb-2 text-xs text-gray-500">
			<a href="/superman/signup-gateways" class="hover:text-emerald-600">Signup Register Gateways</a>
			<span class="mx-1">›</span>
			<span class="text-gray-800 dark:text-gray-200">{editing ? 'Edit' : 'Create'}</span>
		</nav>
		<h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">
			{editing ? 'Edit Signup Register Gateway' : 'Create Signup Register Gateway'}
		</h1>
	</div>

	{#if errorMsg}
		<div class="max-w-3xl rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">{errorMsg}</div>
	{/if}

	<div class="max-w-3xl rounded-xl border border-gray-200 bg-white shadow-xs dark:border-gray-800 dark:bg-gray-900">
		<div class="border-b border-gray-100 p-6 dark:border-gray-800">
			<h2 class="text-base font-semibold text-gray-950 dark:text-white">HTTP API Configuration</h2>
			<p class="mt-1 text-sm text-gray-500">
				Same payload as POST /api/v1/messages: bearer token, sender_id, numbers, and body. Used for register-page phone OTPs with priority failover.
			</p>
		</div>

		{#if loading}
			<div class="p-6 text-sm text-gray-500">Loading…</div>
		{:else}
			<div class="grid grid-cols-1 gap-5 p-6 sm:grid-cols-2">
				<div>
					<label class={label} for="sg-name">Gateway Name <span class="text-red-500">*</span></label>
					<input id="sg-name" class={input} placeholder="e.g., Production WhiteInfo API" bind:value={form.name} />
				</div>
				<div>
					<label class={label} for="sg-priority">Priority <span class="text-red-500">*</span></label>
					<input id="sg-priority" type="number" min="1" class={input} bind:value={form.priority} />
					<p class="mt-1 text-xs text-gray-500">Lower number = higher priority (1 is tried first, then 2, 3…).</p>
				</div>
				<div class="sm:col-span-2">
					<label class={label} for="sg-url">Endpoint URL <span class="text-red-500">*</span></label>
					<input id="sg-url" class={input} placeholder="https://sms.whiteinfobd.com/api/v1/messages" bind:value={form.endpoint_url} />
				</div>
				<div>
					<label class={label} for="sg-method">HTTP Method <span class="text-red-500">*</span></label>
					<select id="sg-method" class={input} bind:value={form.http_method}>
						<option value="POST">POST (JSON body)</option>
						<option value="GET">GET (query string)</option>
					</select>
				</div>
				<div>
					<label class={label} for="sg-timeout">Timeout (seconds) <span class="text-red-500">*</span></label>
					<input id="sg-timeout" type="number" min="1" max="120" class={input} bind:value={form.timeout_seconds} />
				</div>
				<div class="sm:col-span-2">
					<label class={label} for="sg-token">Bearer Access Token {#if !editing}<span class="text-red-500">*</span>{/if}</label>
					<div class="flex gap-2">
						<input
							id="sg-token"
							type={showToken ? 'text' : 'password'}
							autocomplete="new-password"
							class={input}
							placeholder={editing && hasToken ? 'Leave blank to keep the current token' : ''}
							bind:value={form.access_token}
						/>
						<button type="button" onclick={() => (showToken = !showToken)} class="rounded-lg border border-gray-300 px-3 text-xs font-semibold text-gray-700 hover:bg-gray-50 dark:border-gray-700 dark:text-gray-300">
							{showToken ? 'Hide' : 'Show'}
						</button>
					</div>
					<p class="mt-1 text-xs text-gray-500">Create this from a customer account: Developer → Access Tokens. Stored encrypted. Shown only when you type a new value.</p>
				</div>
				<div>
					<label class={label} for="sg-sender">Sender ID (label) <span class="text-red-500">*</span></label>
					<input id="sg-sender" class={input} bind:value={form.sender_id} autocomplete="off" />
					<p class="mt-1 text-xs text-gray-500">The sender label from GET /api/v1/senders, not the numeric database id.</p>
				</div>
				<div>
					<label class={label} for="sg-format">Phone Format <span class="text-red-500">*</span></label>
					<select id="sg-format" class={input} bind:value={form.phone_format}>
						<option value="local_bd">Local BD (017XXXXXXXX)</option>
						<option value="intl_880">International (88017XXXXXXXX)</option>
						<option value="intl_plus">International with + (+88017XXXXXXXX)</option>
					</select>
				</div>
				<div class="sm:col-span-2">
					<label class={label} for="sg-template">OTP Message Template <span class="text-red-500">*</span></label>
					<textarea id="sg-template" rows="3" class={input} bind:value={form.otp_template}></textarea>
					<p class="mt-1 text-xs text-gray-500">Use {'{otp}'} for the code and {'{minutes}'} for expiry.</p>
				</div>
				<div class="sm:col-span-2">
					<label class="inline-flex cursor-pointer items-center gap-3">
						<input type="checkbox" class="peer sr-only" bind:checked={form.is_active} />
						<span class="relative h-6 w-11 rounded-full bg-gray-300 transition peer-checked:bg-emerald-600 after:absolute after:left-0.5 after:top-0.5 after:h-5 after:w-5 after:rounded-full after:bg-white after:transition peer-checked:after:translate-x-5"></span>
						<span class="text-sm font-semibold text-gray-800 dark:text-gray-200">Active Gateway</span>
					</label>
					<p class="mt-1 text-xs text-gray-500">Only active gateways are used during register OTP failover.</p>
				</div>
			</div>
		{/if}
	</div>

	<div class="flex flex-wrap gap-3">
		<button type="button" disabled={saving || loading} onclick={() => save(false)} class="rounded-lg bg-emerald-600 px-4 py-2 text-sm font-bold text-white hover:bg-emerald-700 disabled:opacity-50">
			{saving ? 'Saving…' : editing ? 'Save changes' : 'Create'}
		</button>
		{#if !editing}
			<button type="button" disabled={saving} onclick={() => save(true)} class="rounded-lg border border-gray-300 px-4 py-2 text-sm font-semibold text-gray-800 hover:bg-gray-50 disabled:opacity-50 dark:border-gray-700 dark:text-gray-200">
				Create &amp; create another
			</button>
		{/if}
		<a href="/superman/signup-gateways" class="rounded-lg border border-gray-300 px-4 py-2 text-sm font-semibold text-gray-800 hover:bg-gray-50 dark:border-gray-700 dark:text-gray-200">Cancel</a>
	</div>
</div>
