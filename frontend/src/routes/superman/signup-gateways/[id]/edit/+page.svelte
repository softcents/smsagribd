<script lang="ts">
	import { page } from '$app/stores';
	import SignupGatewayForm from '$lib/components/admin/SignupGatewayForm.svelte';
</script>

<svelte:head><title>Edit Signup Register Gateway — Superman Administrator</title></svelte:head>
{#key $page.params.id}
	<SignupGatewayForm id={Number($page.params.id)} />
{/key}
