<script>
	import Tabs from '$lib/components/tabs/Tabs.svelte';
	import { cueListTabState } from './cuesListsState.svelte';
</script>

<div class="h-full">
	<Tabs tabManager={cueListTabState}></Tabs>
</div>
