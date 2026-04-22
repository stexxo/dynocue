<script lang="ts">
	import { cuelistsStore } from "../../stores/cuelistsStore.svelte";
	import "./CuelistsTableTypes.svelte.ts";
	import EditableTableData from "$lib/components/table/EditableTableData.svelte";

	let cuelists = $derived(cuelistsStore.cuelists);
	const props : CueListTableProps = $props()


</script>

<div class="flex flex-row justify-center w-full">
	<div class="w-full max-w-7xl">
		<div class="mb-5 w-full flex flex-row justify-end">
			{#if props.AllowCreation}
				<button class="btn btn-primary" onclick={() => {cuelistsStore.create(0)}}>Create Cue List</button>
			{/if}
		</div>
		<div class="max-h-full overflow-auto">
			<table class="table table-pin-rows">
				<thead class="sticky">
					<tr>
						<th class="w-40">#</th>
						<th class="min-w-50 max-w-200">Label</th>
						<th class="min-w-50 max-w-100">Type</th>
						<th class="min-w-50 max-w-100"></th>
					</tr>
				</thead>
				<tbody class="overflow-y-auto">
					{#each cuelists as list}
						<tr>
							<td>{list.number}</td>
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
