<!--
  This Source Code Form is subject to the terms of the Mozilla Public
  License, v. 2.0. If a copy of the MPL was not distributed with this
  file, You can obtain one at https://mozilla.org/MPL/2.0/.
-->

<script lang="ts">
	import { audioStore } from '../../stores/audioStore.svelte';
	import ConfirmationModal from '$lib/components/modals/ConfirmationModal.svelte';
	import { clickOutside } from '$lib/utils/clickOutside';

	let files = $derived(audioStore.files);

	let fileToDelete = $state<{ id: string; key: string } | null>(null);
	let deleteModal: ReturnType<typeof ConfirmationModal>;

	function confirmDelete(id: string, key: string) {
		fileToDelete = { id, key };
		deleteModal?.show();
	}

	function handleDelete() {
		if (fileToDelete) {
			audioStore.deleteFile(fileToDelete.id);
			fileToDelete = null;
		}
	}
</script>

<div class="flex h-full w-full flex-col items-center overflow-hidden">
	<div class="flex h-full w-full max-w-7xl flex-col">
		<div class="mb-5 flex w-full flex-none flex-row justify-end">
			<button
				class="btn btn-primary"
				onclick={() => {
					audioStore.addWithDialog();
				}}>Add Audio File</button
			>
		</div>
		<div class="flex-1 overflow-auto">
			<table class="table-pin-rows table">
				<thead class="sticky top-0 z-10 bg-base-100">
					<tr class="bg-base-100">
						<th class="w-1/2">Key</th>
						<th class="w-1/4">ID</th>
						<th class="w-1/4"></th>
					</tr>
				</thead>
				<tbody class="">
					{#each files as file}
						<tr class="hover:bg-base-200">
							<td>{file.key}</td>
							<td class="font-mono text-xs">{file.fileId}</td>
							<td class="flex flex-row justify-end gap-2">
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
												class="btn btn-outline btn-primary mb-2"
												onclick={() => {
													audioStore.replaceWithDialog(file.fileId);
												}}>Replace</button
											>
										</li>
										<li>
											<button
												class="btn btn-outline btn-accent"
												onclick={() => {
													confirmDelete(file.fileId, file.key);
												}}>Delete</button
											>
										</li>
									</ul>
								</details>
							</td>
						</tr>
					{:else}
						<tr>
							<td colspan="3" class="text-center italic text-gray-500"> No audio files found. </td>
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
	message="Are you sure you want to delete audio file '{fileToDelete?.key}'? This action cannot be undone."
	confirmText="Delete"
	onConfirm={handleDelete}
/>
