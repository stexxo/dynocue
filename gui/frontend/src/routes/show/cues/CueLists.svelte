<script lang="ts">
	import CueListTable from "$lib/components/cuelists/CueListTable.svelte";
	import {cueListTabState} from "./cuesListsState.svelte";
	import CueListDetail from "./CueListDetail.svelte";
	import type {TabContentProps} from "$lib/components/tabs/tabTypes.svelte";
	import {cuelistsStore} from "$lib/stores/cuelistsStore.svelte";

	const props:TabContentProps = $props()
	$effect(() => {
		props.tabState.setLabel("Cue Lists");
	});
</script>

<div class="p-4 h-full flex flex-col">
	<CueListTable AllowCreation={true} OnOpenCueList={(id) => {
			const label = () => {
				const cl = cuelistsStore.cueList(id);
				return `Cue List ${cl?.number} ${cl?.label !== "" && cl?.label != null ? `- ${cl?.label}` : ""}`;
			};
			cueListTabState.addTab({ id: id, content: CueListDetail, props: {id: id}, closable: true, label: label })
	}
		}/>
</div>
