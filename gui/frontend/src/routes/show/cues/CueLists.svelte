<script lang="ts">
	import CueListTable from "$lib/components/cuelists/CueListTable.svelte";
	import {cueListTabState} from "./cuesListsState.svelte";
	import {cuelistsStore} from "$lib/stores/cuelistsStore.svelte";
	import CueListDetail from "./CueListDetail.svelte";


</script>

<div class="p-4 h-full flex flex-col">
	<CueListTable AllowCreation={true} OnOpenCueList={(num) => {
		const cueList = cuelistsStore.cueList(num)
		if (cueList){
			cueListTabState.addTab({ id: `cueList-${num}`, label: `Cue List ${num} ${cueList.label !== "" ? `- ${cueList.label}` : ""}`, content: CueListDetail, closable: true })
		}
	}
		}/>
</div>
