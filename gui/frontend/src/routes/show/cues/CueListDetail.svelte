<script lang="ts">
	import type {TabContentProps} from "$lib/components/tabs/tabTypes.svelte";
    import CueTable from "$lib/components/cueing/cues/CueTable.svelte";
    const props:TabContentProps = $props()
</script>

<div class="p-4">
    <CueTable CueListId={props.id} />
</div>
