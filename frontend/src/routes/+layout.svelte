<script lang="ts">
	import '../app.css';
	import Toast from '$lib/components/Toast.svelte';
	import ImpersonationBanner from '$lib/components/ImpersonationBanner.svelte';

	let { children } = $props();
</script>

{@render children()}
<ImpersonationBanner />
<Toast />
