<script lang="ts">
	import { cuelistsStore } from "../../../stores/cuelistsStore.svelte";
	import EditableTableData from "$lib/components/table/EditableTableData.svelte";

	interface CueListTableProps {
		AllowCreation?: boolean;
		OnOpenCueList: (id: string) => void;
	}

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
					<tr class="hover:bg-base-200">
						<EditableTableData inputType="number" value={list.number} onSaveEdit={(v)=>{cuelistsStore.renumberCuelist(list.id, v)}} tdClass="w-40"/>
						<EditableTableData inputType="text" value={list.label} onSaveEdit={(v)=>{cuelistsStore.setMetadataField(list.id, "label", v)}} tdClass="max-w-200"/>
						<td>{list.cueListType}</td>
						<td class="flex flex-row justify-end gap-2">
							<button class="btn btn-outline btn-secondary" onclick={()=>{props.OnOpenCueList(list.id)}}>Open</button>

							<details class="dropdown dropdown-end">
								<summary class="btn btn-ghost btn-secondary">
									<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-6">
										<path stroke-linecap="round" stroke-linejoin="round" d="M12 6.75a.75.75 0 1 1 0-1.5.75.75 0 0 1 0 1.5ZM12 12.75a.75.75 0 1 1 0-1.5.75.75 0 0 1 0 1.5ZM12 18.75a.75.75 0 1 1 0-1.5.75.75 0 0 1 0 1.5Z" />
									</svg>
								</summary>
								<ul class="menu dropdown-content bg-base-200 rounded-box z-[1] w-32 p-2 shadow mt-2">
									<li><button  class="btn btn-outline btn-accent" onclick={()=>{cuelistsStore.deleteCueList(list.id)}}>Delete</button></li>
								</ul>
							</details>
						</td>
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
