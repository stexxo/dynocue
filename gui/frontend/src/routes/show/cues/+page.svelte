<script>
	import Tabs from '$lib/components/tabs/Tabs.svelte';
	import { TabManager } from '$lib/components/tabs/tabTypes.svelte';
    import CueLists from "./CueLists.svelte";

	let cuesTabState = new TabManager(
		[
			{ id: 'cueLists', label: 'Cue Lists',  content: CueLists},
			{ id: 'cueList1', label: 'Cue List1', closable: true }
		],
		'cueLists'
	);
</script>

<div class="h-full">
    <Tabs tabState={cuesTabState}>


    </Tabs>
</div>
