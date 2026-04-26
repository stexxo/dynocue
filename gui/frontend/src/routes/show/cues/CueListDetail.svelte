<script lang="ts">
	import type { TabContentProps } from '$lib/components/tabs/tabTypes.svelte';
	import CueTable from '$lib/components/cueing/cues/CueTable.svelte';
	import CueEdit from './CueEdit.svelte';
	const props: TabContentProps = $props();

	let cueEdit: ReturnType<typeof CueEdit>;
</script>

<div class="flex h-full flex-col p-4">
	<CueTable onEdit={(listId, id) => cueEdit.show(listId, id)} CueListId={props.id} />
</div>

<CueEdit bind:this={cueEdit} />
