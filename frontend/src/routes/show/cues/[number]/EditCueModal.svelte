<script lang="ts">
    import Modal from '$lib/components/Modal.svelte';
    import { cues } from '$lib/stores/cues';
    import type { Cue } from '../../../../bindings/gitlab.com/stexxo/dynocue/api/cues/models';

    interface Props {
        cue: Cue | null;
        onClose: () => void;
    }

    let { cue, onClose }: Props = $props();

    function updateDescription(e: Event) {
        if (!cue) return;
        const value = (e.target as HTMLTextAreaElement).value;
        cues.updateMetadata(cue.cueNumber, 'description', value);
    }
</script>

<Modal 
    open={!!cue} 
    onClose={onClose} 
    title={cue ? `Cue: ${cue.cueNumber} - ${cue.label}` : ''}
>
    {#if cue}
        <div class="flex flex-col gap-4">
            <div class="form-control w-full">
                <label class="label" for="description">
                    <span class="label-text">Description</span>
                </label>
                <textarea 
                    id="description"
                    class="textarea textarea-bordered h-32" 
                    placeholder="Enter cue description..."
                    value={cue.description}
                    onchange={updateDescription}
                ></textarea>
            </div>
        </div>
    {/if}
</Modal>
