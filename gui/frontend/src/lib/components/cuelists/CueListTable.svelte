<script lang="ts">
	import { cuelistsStore } from "../../stores/cuelistsStore.svelte";
	import "./CuelistsTableTypes.svelte.ts";
	import EditableTableData from "$lib/components/table/EditableTableData.svelte";

	let cuelists = $derived(cuelistsStore.cuelists);
	const props : CueListTableProps = $props()


</script>

<div class="flex flex-col items-center w-full h-full overflow-hidden">
	<div class="w-full max-w-7xl h-full flex flex-col">
		<div class="mb-5 w-full flex-none flex flex-row justify-end">
			{#if props.AllowCreation}
				<button class="btn btn-primary" onclick={() => {cuelistsStore.create(0)}}>Create Cue List</button>
			{/if}
		</div>
		<div class="flex-1 overflow-auto">
			<table class="table table-pin-rows">
				<thead class="sticky top-0 z-10 bg-base-100">
					<tr class="bg-base-100">
						<th class="w-40">#</th>
						<th class="min-w-50 max-w-200">Label</th>
						<th class="min-w-50 max-w-100">Type</th>
						<th class="min-w-50 max-w-100"></th>
					</tr>
				</thead>
				<tbody class="">
					{#each cuelists as list}
						<tr>
							<EditableTableData inputType="number" value={list.number} onSaveEdit={(v)=>{cuelistsStore.renumberCuelist(list.number, v)}} tdClass="w-40"/>
							<EditableTableData inputType="text" value={list.label} onSaveEdit={(v)=>{cuelistsStore.setLabel(list.number, v)}} tdClass="max-w-200"/>
							<td>{list.cueListType}</td>
							<td class="flex flex-row justify-center"><button class="btn btn-outline btn-secondary" onclick={()=>{props.OnOpenCueList(list.number)}}>Open</button></td>
						</tr>
					{:else}
						<tr>
							<td colspan="3" class="text-center italic text-gray-500">
								No cue lists found.
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>

	</div>
</div>
