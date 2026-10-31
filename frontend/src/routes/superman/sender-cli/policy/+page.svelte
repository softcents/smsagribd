<script lang="ts">
	let testSenderType = $state<'masking' | 'numeric'>('masking');
	let testSenderId = $state('');
	let testMessageBody = $state('');

	interface ValidationResult {
		hasInput: boolean;
		status: 'idle' | 'compliant' | 'warning' | 'rejected';
		errors: string[];
		warnings: string[];
		passes: string[];
		charCount: number;
		partCount: number;
		isUnicode: boolean;
	}

	let validation = $derived.by<ValidationResult>(() => {
		const sender = testSenderId.trim();
		const body = testMessageBody;
		const type = testSenderType;

		if (!sender && !body) {
			return {
				hasInput: false,
				status: 'idle',
				errors: [],
				warnings: [],
				passes: [],
				charCount: 0,
				partCount: 1,
				isUnicode: false
			};
		}

		const errors: string[] = [];
		const warnings: string[] = [];
		const passes: string[] = [];

		// 1. Sender validation
		if (sender) {
			if (type === 'masking') {
				const len = sender.length;
				if (len < 3 || len > 11) {
					errors.push(`Masking Sender ID length must be between 3 and 11 characters (current: ${len}).`);
				} else {
					passes.push(`Length compliant (${len} characters, allowed: 3-11).`);
				}

				if (!/^[a-zA-Z0-9 ]+$/.test(sender)) {
					errors.push('Masking Sender ID must contain only alphanumeric characters (A-Z, a-z, 0-9) and spaces.');
				} else {
					passes.push('Character set compliant (Alphanumeric).');
				}

				if (/^\d+$/.test(sender)) {
					errors.push('A Masking Sender ID cannot be purely numeric digits. Use Non-Masking for numeric CLIs.');
				}

				const restricted = ['bkash', 'nagad', 'rocket', 'upay', 'btrc', 'police', 'rab', 'army', 'govt', 'bangladesh', 'corona', 'emergency', 'otp', 'admin'];
				for (const kw of restricted) {
					if (sender.toLowerCase().includes(kw)) {
						warnings.push(`Sender ID contains restricted term '${kw}'. Verification documentation / NOC required.`);
					}
				}
			} else {
				const clean = sender.replace(/[^0-9]/g, '');
				if (clean.length === 11 && clean.startsWith('01')) {
					passes.push(`Valid Bangladesh MSISDN format (${clean}).`);
				} else if (clean.length >= 4 && clean.length <= 6) {
					passes.push(`Valid Shortcode format (${clean}).`);
				} else {
					errors.push('Non-masking CLI must be an 11-digit mobile number starting with 01 (or 4-6 digit approved shortcode).');
				}
			}
		}

		// 2. Message Body Analysis
		let isUnicode = false;
		let bodyLen = 0;
		let parts = 1;

		if (body) {
			isUnicode = /[^\u0000-\u007F]/.test(body);
			bodyLen = Array.from(body).length;

			if (isUnicode) {
				const partLimit = 70;
				const concatLimit = 67;
				parts = bodyLen <= partLimit ? 1 : Math.ceil(bodyLen / concatLimit);
				passes.push(`Unicode detected (Bengali/UTF-16). Max ${partLimit} chars for 1 part, ${concatLimit} chars/part for multipart.`);
			} else {
				const partLimit = 160;
				const concatLimit = 153;
				parts = bodyLen <= partLimit ? 1 : Math.ceil(bodyLen / concatLimit);
				passes.push(`Standard GSM-7 encoding (ASCII). Max ${partLimit} chars for 1 part, ${concatLimit} chars/part for multipart.`);
			}

			if (parts > 5) {
				warnings.push(`Message requires ${parts} parts. High billing cost per recipient.`);
			}
		}

		let status: 'idle' | 'compliant' | 'warning' | 'rejected' = 'compliant';
		if (errors.length > 0) {
			status = 'rejected';
		} else if (warnings.length > 0) {
			status = 'warning';
		}

		return {
			hasInput: true,
			status,
			errors,
			warnings,
			passes,
			charCount: bodyLen,
			partCount: parts,
			isUnicode
		};
	});
</script>

<svelte:head>
	<title>Sender Policy — SMS Whiteinfo Administrator</title>
</svelte:head>

<div class="space-y-6">
	<!-- Header Banner -->
	<div class="rounded-2xl bg-gradient-to-r from-emerald-600 via-teal-600 to-cyan-600 p-6 text-white shadow-lg">
		<div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
			<div>
				<div class="inline-flex items-center gap-1.5 rounded-full bg-white/20 px-3 py-1 text-xs font-semibold text-white backdrop-blur-md">
					<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75m-3-7.036A11.959 11.959 0 0 1 3.598 6 11.99 11.99 0 0 0 3 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285Z" />
					</svg>
					Regulatory Compliance Framework
				</div>
				<h1 class="mt-2 text-2xl font-bold tracking-tight text-white">Sender & CLI Policy Standards</h1>
				<p class="mt-1 text-xs text-emerald-100 max-w-2xl">
					Official technical specifications, BTRC regulatory rules, character encodings, and compliance constraints governing Masking and Non-Masking Sender IDs.
				</p>
			</div>
			<div class="flex items-center gap-3">
				<span class="inline-flex items-center gap-1.5 rounded-xl bg-white/10 px-4 py-2 text-xs font-semibold text-white backdrop-blur-md ring-1 ring-white/20">
					<span class="h-2 w-2 rounded-full bg-emerald-300 animate-pulse"></span>
					Policy Engine Active
				</span>
			</div>
		</div>
	</div>

	<!-- Interactive Policy Compliance Validator Tool -->
	<div class="overflow-hidden rounded-2xl bg-white p-6 shadow-xs ring-1 ring-gray-950/5 dark:bg-gray-900 dark:ring-white/10 space-y-6">
		<div class="flex items-center gap-3 border-b border-gray-100 pb-4 dark:border-gray-800">
			<div class="rounded-lg bg-emerald-500/10 p-2 text-emerald-600 dark:text-emerald-400">
				<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" d="M9.813 15.904 9 18.75l-.813-2.846a4.5 4.5 0 0 0-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 0 0 3.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 0 0 3.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 0 0-3.09 3.09ZM18.259 8.715 18 9.75l-.259-1.035a3.375 3.375 0 0 0-2.455-2.456L14.25 6l1.036-.259a3.375 3.375 0 0 0 2.455-2.456L18 2.25l.259 1.035a3.375 3.375 0 0 0 2.456 2.456L21.75 6l-1.035.259a3.375 3.375 0 0 0-2.456 2.456ZM16.894 20.567 16.5 21.75l-.394-1.183a2.25 2.25 0 0 0-1.423-1.423L13.5 18.75l1.183-.394a2.25 2.25 0 0 0 1.423-1.423l.394-1.183.394 1.183a2.25 2.25 0 0 0 1.423 1.423l1.183.394-1.183.394a2.25 2.25 0 0 0-1.423 1.423Z" />
				</svg>
			</div>
			<div>
				<h3 class="text-sm font-bold text-gray-950 dark:text-white">Interactive Sender Policy & Encoding Validator</h3>
				<p class="text-xs text-gray-500 dark:text-gray-400">Test any proposed Sender ID or message payload in real time against telecom regulatory compliance filters.</p>
			</div>
		</div>

		<div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
			<!-- Input Controls -->
			<div class="space-y-4">
				<div>
					<label class="block text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">Sender Category</label>
					<div class="mt-1.5 grid grid-cols-2 gap-2">
						<button
							type="button"
							onclick={() => (testSenderType = 'masking')}
							class="flex items-center justify-center gap-2 rounded-lg py-2 px-3 text-xs font-semibold transition {testSenderType === 'masking' ? 'bg-emerald-600 text-white shadow-xs' : 'bg-gray-100 text-gray-700 hover:bg-gray-200 dark:bg-gray-800 dark:text-gray-300'}"
						>
							<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" d="M15 9h3.75M15 12h3.75M15 15h3.75M4.5 19.5h15a2.25 2.25 0 0 0 2.25-2.25V6.75A2.25 2.25 0 0 0 19.5 4.5h-15a2.25 2.25 0 0 0-2.25 2.25v10.5A2.25 2.25 0 0 0 4.5 19.5Z" />
							</svg>
							Masking (Brand Alphanumeric)
						</button>
						<button
							type="button"
							onclick={() => (testSenderType = 'numeric')}
							class="flex items-center justify-center gap-2 rounded-lg py-2 px-3 text-xs font-semibold transition {testSenderType === 'numeric' ? 'bg-blue-600 text-white shadow-xs' : 'bg-gray-100 text-gray-700 hover:bg-gray-200 dark:bg-gray-800 dark:text-gray-300'}"
						>
							<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" d="M2.25 6.75c0 8.284 6.716 15 15 15h2.25a2.25 2.25 0 0 0 2.25-2.25v-1.372c0-.516-.351-.966-.852-1.091l-4.423-1.106c-.44-.11-.902.055-1.173.417l-.97 1.293c-.282.376-.769.542-1.21.38a12.035 12.035 0 0 1-7.143-7.143c-.162-.441.004-.928.38-1.21l1.293-.97c.363-.271.527-.734.417-1.173L6.963 3.102a1.125 1.125 0 0 0-1.091-.852H4.5A2.25 2.25 0 0 0 2.25 4.5v2.25Z" />
							</svg>
							Non-Masking (Numeric CLI)
						</button>
					</div>
				</div>

				<div>
					<label class="block text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">Test Sender ID / CLI</label>
					<input
						type="text"
						bind:value={testSenderId}
						placeholder={testSenderType === 'masking' ? 'e.g. WHITEINFO or MYBRAND' : 'e.g. 01712345678 or 10666'}
						class="mt-1 block w-full rounded-lg border border-gray-200 bg-white py-2 px-3 font-mono text-xs shadow-xs focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
					/>
				</div>

				<div>
					<div class="flex items-center justify-between">
						<label class="block text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">Sample Message Content</label>
						<span class="text-[11px] font-mono text-gray-400">
							{validation.charCount} chars | {validation.partCount} part(s)
						</span>
					</div>
					<textarea
						bind:value={testMessageBody}
						rows="3"
						placeholder="Type message text here to test GSM-7 vs Unicode encoding and part calculation..."
						class="mt-1 block w-full rounded-lg border border-gray-200 bg-white p-2.5 text-xs shadow-xs focus:border-emerald-500 dark:border-gray-700 dark:bg-gray-800 dark:text-white"
					></textarea>
				</div>
			</div>

			<!-- Validation Results Display -->
			<div class="rounded-xl border border-gray-200 bg-gray-50/70 p-5 dark:border-gray-800 dark:bg-gray-800/40 flex flex-col justify-between">
				<div>
					<div class="flex items-center justify-between">
						<span class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">Compliance Verdict</span>
						{#if !validation.hasInput}
							<span class="inline-flex items-center rounded-md bg-gray-200 px-2.5 py-0.5 text-xs font-medium text-gray-700 dark:bg-gray-700 dark:text-gray-300">
								Awaiting Input
							</span>
						{:else if validation.status === 'compliant'}
							<span class="inline-flex items-center gap-1.5 rounded-md bg-emerald-100 px-2.5 py-0.5 text-xs font-bold text-emerald-700 dark:bg-emerald-950/60 dark:text-emerald-300">
								<svg class="h-3.5 w-3.5 text-emerald-600" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
									<path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
								</svg>
								Fully Compliant
							</span>
						{:else if validation.status === 'warning'}
							<span class="inline-flex items-center gap-1.5 rounded-md bg-amber-100 px-2.5 py-0.5 text-xs font-bold text-amber-700 dark:bg-amber-950/60 dark:text-amber-300">
								<svg class="h-3.5 w-3.5 text-amber-600" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
									<path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" />
								</svg>
								Requires Attention
							</span>
						{:else}
							<span class="inline-flex items-center gap-1.5 rounded-md bg-rose-100 px-2.5 py-0.5 text-xs font-bold text-rose-700 dark:bg-rose-950/60 dark:text-rose-300">
								<svg class="h-3.5 w-3.5 text-rose-600" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
									<path stroke-linecap="round" stroke-linejoin="round" d="m9.75 9.75 4.5 4.5m0-4.5-4.5 4.5M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
								</svg>
								Non-Compliant
							</span>
						{/if}
					</div>

					<div class="mt-4 space-y-2">
						{#if !validation.hasInput}
							<p class="text-xs text-gray-500">Enter a sender ID and optional message text on the left to verify BTRC rules.</p>
						{/if}

						{#each validation.errors as err}
							<div class="flex items-start gap-2 rounded-lg bg-rose-50 p-2.5 text-xs text-rose-700 dark:bg-rose-950/30 dark:text-rose-300">
								<svg class="h-4 w-4 shrink-0 text-rose-500 mt-0.5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
									<path stroke-linecap="round" stroke-linejoin="round" d="m9.75 9.75 4.5 4.5m0-4.5-4.5 4.5M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
								</svg>
								<span>{err}</span>
							</div>
						{/each}

						{#each validation.warnings as warn}
							<div class="flex items-start gap-2 rounded-lg bg-amber-50 p-2.5 text-xs text-amber-700 dark:bg-amber-950/30 dark:text-amber-300">
								<svg class="h-4 w-4 shrink-0 text-amber-500 mt-0.5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
									<path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9.303 3.376c-.866 1.5-.217 3.374-1.948 3.374H4.645c-1.73 0-2.813-1.874-1.948-3.374L10.051 3.378c.866-1.5 3.032-1.5 3.898 0l7.354 12.748ZM12 15.75h.007v.008H12v-.008Z" />
								</svg>
								<span>{warn}</span>
							</div>
						{/each}

						{#each validation.passes as pass}
							<div class="flex items-start gap-2 rounded-lg bg-emerald-50 p-2.5 text-xs text-emerald-700 dark:bg-emerald-950/30 dark:text-emerald-300">
								<svg class="h-4 w-4 shrink-0 text-emerald-500 mt-0.5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
									<path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
								</svg>
								<span>{pass}</span>
							</div>
						{/each}
					</div>
				</div>

				<div class="mt-4 pt-3 border-t border-gray-200 text-xs text-gray-500 dark:border-gray-700 flex items-center justify-between">
					<span>Encoding: <strong class="text-gray-900 dark:text-white">{validation.isUnicode ? 'UTF-16 (Unicode)' : 'GSM-7 (Standard)'}</strong></span>
					<span>Billing Parts: <strong class="text-gray-900 dark:text-white">{validation.partCount}</strong></span>
				</div>
			</div>
		</div>
	</div>

	<!-- Official Regulatory Documentation Grid -->
	<div class="grid grid-cols-1 gap-5 md:grid-cols-2">
		<!-- Card 1: BTRC Masking Regulations -->
		<div class="rounded-xl border border-gray-200 bg-white p-5 shadow-xs dark:border-gray-800 dark:bg-gray-900 space-y-3">
			<div class="flex items-center gap-2 text-emerald-600 dark:text-emerald-400">
				<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" d="M12 21a9.004 9.004 0 0 0 8.716-6.747M12 21a9.004 9.004 0 0 1-8.716-6.747M12 21c2.485 0 4.5-4.03 4.5-9S14.485 3 12 3m0 18c-2.485 0-4.5-4.03-4.5-9S9.515 3 12 3m0 0a8.997 8.997 0 0 1 7.843 4.582M12 3a8.997 8.997 0 0 0-7.843 4.582m15.686 0A11.953 11.953 0 0 1 12 10.5c-2.998 0-5.74-1.1-7.843-2.918m15.686 0A8.959 8.959 0 0 1 21 12c0 .778-.099 1.533-.284 2.253m0 0A17.919 17.919 0 0 1 12 16.5c-3.162 0-6.133-.815-8.716-2.247m0 0A9.015 9.015 0 0 1 3 12c0-1.605.42-3.113 1.157-4.418" />
				</svg>
				<h4 class="text-sm font-bold text-gray-900 dark:text-white">BTRC Masking Regulations</h4>
			</div>
			<ul class="text-xs text-gray-600 dark:text-gray-300 space-y-2 list-disc list-inside">
				<li>Length must strictly stay between <strong>3 and 11 characters</strong>.</li>
				<li>Only alphanumeric English characters <code class="font-mono bg-gray-100 dark:bg-gray-800 px-1 py-0.5 rounded">[A-Za-z0-9]</code> and single spaces allowed.</li>
				<li>No purely numeric sender IDs allowed under masking categories.</li>
				<li>Trade License, TIN Certificate, and Operator NOC required for brand approval.</li>
				<li>Financial and government-mimicking sender IDs strictly embargoed without official direct BTRC consent.</li>
			</ul>
		</div>

		<!-- Card 2: Non-Masking CLI Standards -->
		<div class="rounded-xl border border-gray-200 bg-white p-5 shadow-xs dark:border-gray-800 dark:bg-gray-900 space-y-3">
			<div class="flex items-center gap-2 text-blue-600 dark:text-blue-400">
				<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" d="M10.5 1.5H8.25A2.25 2.25 0 0 0 6 3.75v16.5a2.25 2.25 0 0 0 2.25 2.25h7.5A2.25 2.25 0 0 0 18 20.25V3.75a2.25 2.25 0 0 0-2.25-2.25H13.5m-3 0V3h3V1.5m-3 0h3m-3 18.75h3" />
				</svg>
				<h4 class="text-sm font-bold text-gray-900 dark:text-white">Non-Masking CLI Standards</h4>
			</div>
			<ul class="text-xs text-gray-600 dark:text-gray-300 space-y-2 list-disc list-inside">
				<li>Must be an authenticated 11-digit mobile number format starting with <code class="font-mono bg-gray-100 dark:bg-gray-800 px-1 py-0.5 rounded">013-019</code>.</li>
				<li>Approved shortcodes (4 to 6 digits) are routed under special telecom gateway profiles.</li>
				<li>CLI spoofing or foreign numbering plan injection triggers automated gateway lockout.</li>
				<li>Subject to National Do-Not-Disturb (DND) scrubbing rules.</li>
			</ul>
		</div>

		<!-- Card 3: Character Encoding & Concatenation -->
		<div class="rounded-xl border border-gray-200 bg-white p-5 shadow-xs dark:border-gray-800 dark:bg-gray-900 space-y-3">
			<div class="flex items-center gap-2 text-purple-600 dark:text-purple-400">
				<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" d="M7.5 21 3 16.5m0 0L7.5 12M3 16.5h13.5m0-13.5L21 7.5m0 0L16.5 12M21 7.5H7.5" />
				</svg>
				<h4 class="text-sm font-bold text-gray-900 dark:text-white">Character Encoding & Billing Parts</h4>
			</div>
			<ul class="text-xs text-gray-600 dark:text-gray-300 space-y-2 list-disc list-inside">
				<li><strong>GSM 03.38 (ASCII):</strong> 1st part up to 160 chars; multipart chunks at 153 chars each.</li>
				<li><strong>UCS-2 (Unicode / Bengali):</strong> 1st part up to 70 chars; multipart chunks at 67 chars each.</li>
				<li>Special GSM escape characters (<code class="font-mono bg-gray-100 dark:bg-gray-800 px-1 py-0.5 rounded">[]&#123;&#125;|^~\€</code>) consume 2 septets each.</li>
				<li>Concatenated SMS injects an invisible 6-byte UDH (User Data Header) per segment.</li>
			</ul>
		</div>

		<!-- Card 4: Blacklisted Keywords & Telecom Content Safeguards -->
		<div class="rounded-xl border border-gray-200 bg-white p-5 shadow-xs dark:border-gray-800 dark:bg-gray-900 space-y-3">
			<div class="flex items-center gap-2 text-rose-600 dark:text-rose-400">
				<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m0-10.036A11.959 11.959 0 0 1 3.598 6 11.99 11.99 0 0 0 3 9.75c0 5.592 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.57-.598-3.75h-.152c-3.196 0-6.1-1.249-8.25-3.286Zm0 13.036h.007v.008H12v-.008Z" />
				</svg>
				<h4 class="text-sm font-bold text-gray-900 dark:text-white">Blacklisted Keywords & Spoof Safeguards</h4>
			</div>
			<ul class="text-xs text-gray-600 dark:text-gray-300 space-y-2 list-disc list-inside">
				<li>Keywords containing MFS brands (bKash, Nagad, Rocket, Upay) are blocked from self-registration.</li>
				<li>Government agency names (Police, RAB, Army, BTRC, NID) require verified signed authorization.</li>
				<li>Spam and lottery terms trigger immediate audit review and temporary route suspension.</li>
				<li>All sender assignment events are tracked in the permanent tamper-evident Sender Audit log.</li>
			</ul>
		</div>
	</div>
</div>

