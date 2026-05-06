<!--
  This Source Code Form is subject to the terms of the Mozilla Public
  License, v. 2.0. If a copy of the MPL was not distributed with this
  file, You can obtain one at https://mozilla.org/MPL/2.0/.
-->

<script lang="ts">
	import { cuelistsStore } from '../../../stores/cuelistsStore.svelte';
	import EditableTableData from '$lib/components/table/EditableTableData.svelte';
	import ConfirmationModal from '$lib/components/modals/ConfirmationModal.svelte';
	import { clickOutside } from '$lib/utils/clickOutside';

	interface CueListTableProps {
		AllowCreation?: boolean;
		OnOpenCueList: (id: string) => void;
	}

	let cuelists = $derived(cuelistsStore.cuelists);
	const props: CueListTableProps = $props();

	let listToDelete = $state<{ id: string; number: number } | null>(null);
	let deleteModal: ReturnType<typeof ConfirmationModal>;

	function confirmDelete(id: string, number: number) {
		listToDelete = { id, number };
		deleteModal?.show();
	}

	function handleDelete() {
		if (listToDelete) {
			cuelistsStore.deleteCueList(listToDelete.id);
			listToDelete = null;
		}
	}
</script>

<div class="flex h-full w-full flex-col items-center overflow-hidden">
	<div class="flex h-full w-full max-w-7xl flex-col">
		<div class="mb-5 flex w-full flex-none flex-row justify-end">
			{#if props.AllowCreation}
				<button
					class="btn btn-primary"
					onclick={() => {
						cuelistsStore.create(0);
					}}>Create Cue List</button
				>
			{/if}
		</div>
		<div class="flex-1 overflow-auto">
			<table class="table-pin-rows table">
				<thead class="sticky top-0 z-10 bg-base-100">
					<tr class="bg-base-100">
						<th class="w-40">#</th>
						<th class="max-w-200 min-w-50">Label</th>
						<th class="max-w-100 min-w-50">Type</th>
						<th class="max-w-100 min-w-50"></th>
					</tr>
				</thead>
				<tbody class="">
					{#each cuelists as list}
						<tr class="hover:bg-base-200">
							<EditableTableData
								inputType="number"
								value={list.number}
								onSaveEdit={(v) => {
									cuelistsStore.setAttributesField(list.cueListId, 'number', v);
								}}
								tdClass="w-40"
							/>
							<EditableTableData
								inputType="text"
								value={list.label}
								onSaveEdit={(v) => {
									cuelistsStore.setAttributesField(list.cueListId, 'label', v);
								}}
								tdClass="max-w-200"
							/>
							<td>{list.cueListType}</td>
							<td class="flex flex-row justify-end gap-2">
								<button
									class="btn btn-outline btn-secondary"
									onclick={() => {
										props.OnOpenCueList(list.cueListId);
									}}>Open</button
								>

								<details
									class="dropdown dropdown-end"
									use:clickOutside={(node) => {
										if (node.hasAttribute('open')) {
											node.removeAttribute('open');
										}
									}}
								>
									<summary class="btn btn-ghost btn-secondary">
										<svg
											xmlns="http://www.w3.org/2000/svg"
											fill="none"
											viewBox="0 0 24 24"
											stroke-width="1.5"
											stroke="currentColor"
											class="size-6"
										>
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												d="M12 6.75a.75.75 0 1 1 0-1.5.75.75 0 0 1 0 1.5ZM12 12.75a.75.75 0 1 1 0-1.5.75.75 0 0 1 0 1.5ZM12 18.75a.75.75 0 1 1 0-1.5.75.75 0 0 1 0 1.5Z"
											/>
										</svg>
									</summary>
									<ul
										class="dropdown-content menu z-[1] mt-2 w-32 rounded-box bg-base-200 p-2 shadow"
									>
										<li>
											<button
												class="btn btn-outline btn-accent"
												onclick={() => {
													confirmDelete(list.cueListId, list.number);
												}}>Delete</button
											>
										</li>
									</ul>
								</details>
							</td>
						</tr>
					{:else}
						<tr>
							<td colspan="3" class="text-center italic text-gray-500"> No cue lists found. </td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</div>
</div>

<ConfirmationModal
	bind:this={deleteModal}
	title="Confirm Delete"
	message="Are you sure you want to delete cue list {listToDelete?.number}? This action cannot be undone."
	confirmText="Delete"
	onConfirm={handleDelete}
/>
