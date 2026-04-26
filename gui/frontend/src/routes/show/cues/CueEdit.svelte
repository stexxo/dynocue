<script lang="ts">
	import { cuesStore } from '$lib/stores/cuesStore.svelte';
	import { cuelistsStore } from '$lib/stores/cuelistsStore.svelte';
	import EditableTimeInput from '$lib/components/inputs/EditableTimeInput.svelte';
	import EditableTextInput from '$lib/components/inputs/EditableTextInput.svelte';

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

	export function show(listId: string, id: string) {
		cueListId = listId;
		cueId = id;
		dialog.showModal();
	}
</script>

<dialog bind:this={dialog} id="cue_edit_modal" class="modal">
	{#if cue}
		<div class="modal-box w-2/3 max-w-7xl">
			<div class="flex flex-col gap-5">
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
							if (cue) cuesStore.updateCueMetadata(cue.cueListId, cue.cueId, 'label', v);
						}}
					/>

					<div class="grid grid-cols-2 gap-4">
						<EditableTimeInput
							label="Delay"
							value={cue.delay}
							onSave={(v) => cuesStore.updateCueMetadata(cue.cueListId, cue.cueId, 'delay', v)}
						/>
						<EditableTimeInput
							label="Follow"
							value={cue.follow}
							onSave={(v) => cuesStore.updateCueMetadata(cue.cueListId, cue.cueId, 'follow', v)}
						/>
					</div>

					<div class="md:col-span-2">
						<EditableTextInput
							textarea={true}
							label="Description"
							inputClass="h-24"
							value={cue.description}
							onSave={(v) => {
								if (cue) cuesStore.updateCueMetadata(cue.cueListId, cue.cueId, 'description', v);
							}}
						/>
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
