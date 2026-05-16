<!--
  This Source Code Form is subject to the terms of the Mozilla Public
  License, v. 2.0. If a copy of the MPL was not distributed with this
  file, You can obtain one at https://mozilla.org/MPL/2.0/.
-->

<script lang="ts">
	import { cuesStore } from '../../../stores/cuesStore.svelte';
	import { cueExecutionStore } from '$lib/stores/cueExecutionStore.svelte';
	import EditableTableData from '$lib/components/table/EditableTableData.svelte';
	import EditableTimeData from '$lib/components/table/EditableTimeData.svelte';
	import ConfirmationModal from '$lib/components/modals/ConfirmationModal.svelte';
	import { clickOutside } from '$lib/utils/clickOutside';
	import { GoToCue } from '../../../../../bindings/github.com/stexxo/dynocue/gui/services/executionservice';

	interface CueTableProps {
		CueListId: string;
		onEdit: (cueListId: string, cueId: string) => void;
	}

	const props: CueTableProps = $props();
	let cues = $derived(cuesStore.cues.get(props.CueListId));

	let cueToDelete = $state<{ cueListId: string; cueId: string; number: number } | null>(null);
	let deleteModal: ReturnType<typeof ConfirmationModal>;

	function confirmDelete(cueListId: string, cueId: string, number: number) {
		cueToDelete = { cueListId, cueId, number };
		deleteModal?.show();
	}

	function handleDelete() {
		if (cueToDelete) {
			cuesStore.deleteCue(cueToDelete.cueListId, cueToDelete.cueId);
			cueToDelete = null;
		}
	}
</script>

<div class="flex h-full w-full flex-col items-center overflow-hidden">
	<div class="max-w-10xl flex h-full w-full flex-col">
		<div class="mb-5 flex w-full flex-none flex-row justify-end">
			<button
				class="btn btn-primary"
				onclick={() => {
					cuesStore.create(props.CueListId, 0);
				}}>Create Cue</button
			>
		</div>
		<div class="flex-1 overflow-auto">
			<table class="table-pin-rows table">
				<thead class="sticky top-0 z-10 bg-base-100">
					<tr class="bg-base-100">
						<th class="w-25 text-center"></th>
						<th class="w-100 text-center">#</th>
						<th class="max-w-200 min-w-100">Label</th>
						<th class="w-100">Delay</th>
						<th class="w-100">Follow</th>
						<th class="w-100"></th>
					</tr>
				</thead>

				<tbody class="">
					{#each cues ?? [] as list}
						{@const execution = cueExecutionStore.getExecution(list.cueId)}
						<tr
							class="hover:cursor-pointer {execution?.selected && execution?.active
								? 'bg-indigo-900 hover:bg-indigo-950'
								: execution?.selected
									? 'bg-cyan-900 hover:bg-cyan-950'
									: execution?.active
										? 'bg-emerald-900 hover:bg-emerald-950'
										: 'hover:bg-base-200'}"
						>
							<td>
								{#if execution?.active}
									<span class="loading loading-xs loading-spinner">playing</span>
								{:else}
									<svg
										xmlns="http://www.w3.org/2000/svg"
										viewBox="0 0 24 24"
										fill="currentColor"
										class="size-4 opacity-20"
									>
										<title>Stopped</title>
										<rect x="6" y="6" width="12" height="12" rx="2" />
									</svg>
								{/if}
							</td>
							<EditableTableData
								inputType="number"
								value={list.number}
								onSaveEdit={(v) => {
									cuesStore.updateCueAttributes(list.cueListId, list.cueId, 'number', v);
								}}
								tdClass="w-40 text-center"
							/>
							<EditableTableData
								inputType="text"
								value={list.label}
								onSaveEdit={(v) => {
									cuesStore.updateCueAttributes(list.cueListId, list.cueId, 'label', v);
								}}
								tdClass="min-w-100 max-w-200"
							/>
							<EditableTimeData
								tdClass="w-100"
								value={list.delay}
								onSaveEdit={(v) => {
									cuesStore.updateCueAttributes(list.cueListId, list.cueId, 'delay', v);
								}}
							/>
							<EditableTimeData
								tdClass="w-100"
								value={list.follow}
								onSaveEdit={(v) => {
									cuesStore.updateCueAttributes(list.cueListId, list.cueId, 'follow', v);
								}}
							/>
							<td class="flex flex-row justify-end gap-1">
								<button
									class="btn btn-soft btn-secondary"
									onclick={() => {
										props.onEdit(list.cueListId, list.cueId);
									}}
								>
									Edit
								</button>
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
										class="dropdown-content menu z-[1] w-32 gap-2 rounded-box bg-base-200 p-2 shadow"
									>
										<li>
											<button
												class="btn btn-outline btn-primary"
												onclick={() => {
													GoToCue(list.cueId);
												}}>Go To</button
											>
										</li>
										<li>
											<button
												class="btn btn-outline btn-error"
												onclick={() => {
													confirmDelete(list.cueListId, list.cueId, list.number);
												}}>Delete</button
											>
										</li>
									</ul>
								</details>
							</td>
						</tr>
					{:else}
						<tr>
							<td colspan="3" class="text-center italic text-gray-500"> No cues found. </td>
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
	message="Are you sure you want to delete cue {cueToDelete?.number}? This action cannot be undone."
	confirmText="Delete"
	onConfirm={handleDelete}
/>
