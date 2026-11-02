<script lang="ts">
	import { page } from '$app/stores';
	import OperatorForm from '$lib/components/admin/OperatorForm.svelte';
</script>

<svelte:head><title>Edit Operator — Superman Administrator</title></svelte:head>
{#key $page.params.id}
	<OperatorForm id={Number($page.params.id)} />
{/key}
