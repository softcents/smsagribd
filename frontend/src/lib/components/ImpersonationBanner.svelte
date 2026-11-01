<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';

	const MARKER = 'sms_impersonation';

	let info = $state<{ portal: string; name?: string; email?: string; returnTo?: string } | null>(null);
	let leaving = $state(false);
	let visible = $derived(info !== null && !$page.url.pathname.startsWith('/superman'));

	onMount(() => {
		try {
			const raw = sessionStorage.getItem(MARKER);
			info = raw ? JSON.parse(raw) : null;
		} catch {
			info = null;
		}
	});

	async function leave() {
		if (!info || leaving) return;
		leaving = true;
		const returnTo = info.returnTo || '/superman';
		try {
			if (info.portal === 'sp') {
				const { spApi } = await import('$lib/sp-api.js');
				try {
					await spApi.post('/api/v1/sp/auth/logout');
				} catch {
					// best effort
				}
				spApi.clearToken();
			} else if (info.portal === 'customer') {
				const token = sessionStorage.getItem('sms_customer_session_token');
				if (token) {
					try {
						await fetch('/api/v1/customer/auth/logout', {
							method: 'POST',
							headers: { Accept: 'application/json', Authorization: `Bearer ${token}` }
						});
					} catch {
						// best effort
					}
				}
				sessionStorage.removeItem('sms_customer_session_token');
				localStorage.removeItem('sms_customer_access_token');
			}
		} finally {
			sessionStorage.removeItem(MARKER);
			window.location.assign(returnTo);
		}
	}
</script>

{#if visible && info}
	<div
		class="fixed bottom-4 left-1/2 z-[9999] flex max-w-[calc(100vw-2rem)] -translate-x-1/2 items-center gap-3 rounded-full border border-amber-300 bg-amber-50 px-4 py-2 text-sm text-amber-900 shadow-lg"
	>
		<span class="truncate">
			Viewing as <strong>{info.name ?? 'user'}</strong>
			<span class="hidden sm:inline">(Admin session)</span>
		</span>
		<button
			type="button"
			onclick={leave}
			disabled={leaving}
			class="shrink-0 rounded-full bg-amber-600 px-3 py-1 text-xs font-semibold text-white hover:bg-amber-700 disabled:opacity-60"
		>
			{leaving ? 'Leaving…' : 'Leave & return to Admin'}
		</button>
	</div>
{/if}
