<script lang="ts">
	import type {TabContentProps} from "$lib/components/tabs/tabTypes.svelte";
    const props:TabContentProps = $props()
</script>

<div class="p-4">
    Detail for Cue List {props.id}
</div>
