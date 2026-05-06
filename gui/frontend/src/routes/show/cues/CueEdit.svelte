<!--
  This Source Code Form is subject to the terms of the Mozilla Public
  License, v. 2.0. If a copy of the MPL was not distributed with this
  file, You can obtain one at https://mozilla.org/MPL/2.0/.
-->

<script lang="ts">
	import { cuesStore } from '$lib/stores/cuesStore.svelte';
	import { cuelistsStore } from '$lib/stores/cuelistsStore.svelte';
	import { actionTemplatesStore } from '$lib/stores/actiontemplatesStore.svelte';
	import { actionsStore } from '$lib/stores/actionsStore.svelte';
	import { clickOutside } from '$lib/utils/clickOutside';
	import EditableTimeInput from '$lib/components/inputs/EditableTimeInput.svelte';
	import EditableTextInput from '$lib/components/inputs/EditableTextInput.svelte';
	import ActionDetail from './ActionDetail.svelte';

	interface CueEditProps {
		cueListId?: string;
		cueId?: string;
	}

	let { cueListId = $bindable(), cueId = $bindable() }: CueEditProps = $props();

	let dialog: HTMLDialogElement;

	let cue = $derived.by(() => {
		if (!cueListId || !cueId) return null;
		const cues = cuesStore.cues.get(cueListId);
		return cues?.find((c) => c.cueId === cueId) ?? null;
	});

	let cuelist = $derived.by(() => {
		if (!cueListId) return null;
		return cuelistsStore.cueList(cueListId);
	});

	let actions = $derived.by(() => {
		if (!cueId) return [];
		return actionsStore.actions.get(cueId) ?? [];
	});

	let selectedTemplateId = $state('');
	let searchTerm = $state('');
	let dropdownOpen = $state(false);

	let filteredTemplates = $derived.by(() => {
		const term = searchTerm.toLowerCase();
		return actionTemplatesStore.templates.filter(
			(t) =>
				t.templateName.toLowerCase().includes(term) || t.subsystemName.toLowerCase().includes(term)
		);
	});

	let groupedTemplates = $derived.by(() => {
		const groups = new Map<string, typeof actionTemplatesStore.templates>();
		for (const template of filteredTemplates) {
			const group = groups.get(template.subsystemName) ?? [];
			group.push(template);
			groups.set(template.subsystemName, group);
		}
		return Array.from(groups.entries()).sort(([a], [b]) => a.localeCompare(b));
	});

	let selectedTemplate = $derived.by(() => {
		return actionTemplatesStore.templates.find((t) => t.templateId === selectedTemplateId) ?? null;
	});

	async function createAction() {
		if (!cueListId || !cueId || !selectedTemplateId) return;
		await actionsStore.create(cueListId, cueId, selectedTemplateId);
	}

	export function show(listId: string, id: string) {
		cueListId = listId;
		cueId = id;
		selectedTemplateId = '';
		searchTerm = '';
		dropdownOpen = false;
		dialog.showModal();
	}
</script>

<dialog bind:this={dialog} id="cue_edit_modal" class="modal">
	{#if cue}
		<div class="modal-box flex h-[90vh] w-2/3 max-w-7xl flex-col">
			<div class="flex min-h-0 grow flex-col gap-5">
				<form method="dialog">
					<button class="btn absolute top-2 right-2 btn-circle btn-ghost btn-sm">✕</button>
				</form>
				<div class="flex items-center justify-between">
					<h3 class="text-2xl font-bold">Edit Cue</h3>
					<div class="mr-5 flex gap-4 text-lg">
						<span class="badge p-4 badge-primary">List: {cuelist?.number ?? 'N/A'}</span>
						<span class="badge p-4 badge-secondary">Cue: {cue.number}</span>
					</div>
				</div>
				<div class="grid grid-cols-1 gap-6 md:grid-cols-2">
					<EditableTextInput
						label="Label"
						value={cue.label}
						onSave={(v) => {
							if (cue) cuesStore.updateCueAttributes(cue.cueListId, cue.cueId, 'label', v);
						}}
					/>

					<div class="grid grid-cols-2 gap-4">
						<EditableTimeInput
							label="Delay"
							value={cue.delay}
							onSave={(v) => cuesStore.updateCueAttributes(cue.cueListId, cue.cueId, 'delay', v)}
						/>
						<EditableTimeInput
							label="Follow"
							value={cue.follow}
							onSave={(v) => cuesStore.updateCueAttributes(cue.cueListId, cue.cueId, 'follow', v)}
						/>
					</div>

					<div class="md:col-span-2">
						<EditableTextInput
							textarea={true}
							label="Description"
							inputClass="h-24"
							value={cue.description}
							onSave={(v) => {
								if (cue) cuesStore.updateCueAttributes(cue.cueListId, cue.cueId, 'description', v);
							}}
						/>
					</div>
				</div>

				<div class="divider">Actions</div>

				<div class="flex min-h-0 grow flex-col gap-4">
					<div class="flex items-start gap-4">
						<div class="form-control grow">
							<div
								class="dropdown w-full"
								class:dropdown-open={dropdownOpen}
								use:clickOutside={() => (dropdownOpen = false)}
							>
								<div
									tabindex="0"
									role="button"
									class="select-bordered select flex w-full items-center justify-between"
									onclick={() => (dropdownOpen = !dropdownOpen)}
								>
									<span class="truncate">
										{selectedTemplate?.templateName ?? 'Select a template'}
									</span>
								</div>
								<!-- svelte-ignore a11y_no_noninteractive_tabindex -->
								<div
									tabindex="0"
									class="dropdown-content menu z-[100] w-full gap-2 rounded-box bg-base-200 p-2 shadow"
								>
									<input
										type="text"
										placeholder="Search templates..."
										class="input-bordered input input-sm w-full"
										bind:value={searchTerm}
										onclick={(e) => e.stopPropagation()}
									/>
									<div class="max-h-60 overflow-y-auto">
										{#each groupedTemplates as [subsystem, templates]}
											<div class="flex items-center gap-2 menu-title">
												<span>{subsystem}</span>
											</div>
											<ul>
												{#each templates as template}
													<li>
														<button
															class="flex flex-col items-start"
															class:active={selectedTemplateId === template.templateId}
															onclick={() => {
																selectedTemplateId = template.templateId;
																dropdownOpen = false;
																searchTerm = '';
															}}
														>
															<span class="font-bold">{template.templateName}</span>
														</button>
													</li>
												{/each}
											</ul>
										{:else}
											<div class="p-2 text-center italic text-gray-500">No templates found</div>
										{/each}
									</div>
								</div>
							</div>
						</div>
						<button class="btn btn-primary" disabled={!selectedTemplateId} onclick={createAction}>
							Add Action
						</button>
					</div>

					<div class="grow overflow-auto">
						<table class="table w-full">
							<thead class="sticky top-0 z-30 bg-base-100">
								<tr>
									<th class="w-16"></th>
									<th class="w-64">Label</th>
									<th class="w-40">Template</th>
									<th class="w-32">Delay</th>
									<th class="w-16"></th>
								</tr>
							</thead>
							<tbody>
								{#each actions as action (action.id)}
									<ActionDetail {cueListId} {cueId} actionId={action.id} />
								{/each}
							</tbody>
						</table>
					</div>
				</div>
			</div>
		</div>
		<form method="dialog" class="modal-backdrop">
			<button>close</button>
		</form>
	{:else}
		<div class="modal-box">
			<h3 class="text-lg font-bold text-error">Cue Not Found</h3>
			<p class="py-4">The requested cue could not be found.</p>
			<div class="modal-action">
				<form method="dialog">
					<button class="btn">Close</button>
				</form>
			</div>
		</div>
	{/if}
</dialog>
