<script lang="ts">
    import { cuesStore } from "$lib/stores/cuesStore.svelte";
    import { cuelistsStore } from "$lib/stores/cuelistsStore.svelte";
    import EditableTimeInput from "$lib/components/inputs/EditableTimeInput.svelte";
    import EditableTextInput from "$lib/components/inputs/EditableTextInput.svelte";

    interface CueEditProps {
        cueListId?: string;
        cueId?: string;
    }

    let { cueListId = $bindable(), cueId = $bindable() }: CueEditProps = $props();

    let dialog: HTMLDialogElement;

    let cue = $derived.by(() => {
        if (!cueListId || !cueId) return null;
        const cues = cuesStore.cues.get(cueListId);
        return cues?.find(c => c.cueId === cueId) ?? null;
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
        <div class="modal-box w-11/12 max-w-5xl">
            <form method="dialog">
                <button class="btn btn-sm btn-circle btn-ghost absolute right-2 top-2">✕</button>
            </form>
            <div class="flex justify-between items-center mb-6">
                <h3 class="text-2xl font-bold">Edit Cue</h3>
                <div class="flex gap-4 text-lg mr-5">
                    <span class="badge badge-primary p-4">List: {cuelist?.number ?? 'N/A'}</span>
                    <span class="badge badge-secondary p-4">Cue: {cue.number}</span>
                </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
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
                        onSave={(v) => cuesStore.updateCueMetadata(cue.cueListId, cue.cueId, "delay", v)}
                    />
                    <EditableTimeInput
                        label="Follow"
                        value={cue.follow}
                        onSave={(v) => cuesStore.updateCueMetadata(cue.cueListId, cue.cueId, "follow", v)}
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
