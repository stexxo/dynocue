<script lang="ts">
	import { cuelistsStore } from "../../stores/cuelistsStore.svelte";
	import type { CueListMetadata } from "../../../../bindings/github.com/stexxo/dynocue/components/cues/types";
	import "./CuelistsTableTypes.svelte.ts";

	let cuelists = $derived(cuelistsStore.cuelists);
	const props : CueListTableProps = $props()

	let editingId = $state<number | null>(null);
	let editValue = $state("");

	function startEdit(list: CueListMetadata) {
		editingId = list.number;
		editValue = list.label;
	}

	function saveEdit() {
		if (editingId !== null) {
			cuelistsStore.setLabel(editingId, editValue);
		}
		editingId = null;
	}

	function cancelEdit() {
		editingId = null;
	}

	function focus(node: HTMLInputElement) {
		node.focus();
	}
</script>

<div class="overflow-x-auto">
	<div class="mb-5 w-full flex flex-row justify-end">
		{#if props.AllowCreation}
			<button class="btn btn-primary" onclick={() => {cuelistsStore.create(0)}}>Create Cue List</button>
		{/if}
	</div>
	<table class="table w-full">
		<thead>
			<tr>
				<th>Number</th>
				<th>Label</th>
				<th>Type</th>
				<th></th>
			</tr>
		</thead>
		<tbody>
			{#each cuelists as list}
				<tr>
					<td>{list.number}</td>
					<td ondblclick={() => startEdit(list)}>
						{#if editingId === list.number}
							<input
								type="text"
								class="input input-bordered input-sm w-full"
								bind:value={editValue}
								onblur={saveEdit}
								onkeydown={(e) => {
									if (e.key === 'Enter') saveEdit();
									if (e.key === 'Escape') cancelEdit();
								}}
								use:focus
							/>
						{:else}
							{list.label}
						{/if}
					</td>
					<td>{list.cueListType}</td>
					<td><button class="btn btn-outline btn-secondary" onclick={()=>{props.OnOpenCueList(list.number)}}>Open</button></td>
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
