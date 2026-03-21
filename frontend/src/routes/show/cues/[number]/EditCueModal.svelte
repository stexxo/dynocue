<script lang="ts">
    import Modal from '$lib/components/Modal.svelte';
    import { cues } from '$lib/stores/cues';
    import type { Cue } from '../../../../../bindings/github.com/stexxo/dynocue/api/cues';
    import { actions } from '$lib/stores/actions';
    import DataTable, { type ColumnConfig, type ToolbarButton } from '$lib/components/DataTable.svelte';
    import type { CueAction } from '../../../../../bindings/github.com/stexxo/dynocue/api/cues/models';

    interface Props {
        cueListNumber: number;
        cue: Cue | null;
        onClose: () => void;
    }

    let { cueListNumber, cue, onClose }: Props = $props();

    const actionStore = $derived(cue ? actions.forCue(cueListNumber, cue.cueNumber) : null);

    $effect(() => {
        if (actionStore) {
            actionStore.refresh();
        }
    });

    const columns: ColumnConfig<CueAction>[] = [
        {
            key: 'actionNumber',
            label: '#',
            width: 'w-25',
            editable: true,
            onSave: (item: CueAction, newValue: string) => {
                const newNum = parseFloat(newValue);
                if (!isNaN(newNum) && newNum > 0 && actionStore) {
                    return actionStore.move(item.actionNumber, newNum);
                }
            }
        },
        {
            key: 'label',
            label: 'Label',
            width: 'w-30',
            editable: true,
            onSave: (item: CueAction, newValue: string) => actionStore?.updateAction(item.actionNumber, 'label', newValue)
        },
        {
            key: 'sourceType',
            label: 'Source Type',
            minWidth: 'w-50',
            editable: true,
            onSave: (item: CueAction, newValue: string) => actionStore?.updateAction(item.actionNumber, 'sourceType', newValue)
        },
        {
            key: 'action',
            label: 'Action',
            width: 'w-32',
            editable: true,
            onSave: (item: CueAction, newValue: string) => actionStore?.updateAction(item.actionNumber, 'action', newValue)
        },
        {
            key: 'target',
            label: 'Target',
            width: 'w-20',
            editable: true,
            onSave: (item: CueAction, newValue: string) => {
                const newTarget = parseFloat(newValue);
                if (!isNaN(newTarget) && actionStore) {
                    return actionStore.updateAction(item.actionNumber, 'target', newValue);
                }
            }
        }
    ];

    const toolbar: ToolbarButton<CueAction>[] = [
        {
            label: 'Add Action',
            class: 'btn-primary',
            icon: addIcon,
            onclick: () => actionStore?.create(0)
        },
        {
            label: 'Delete Selected',
            class: 'btn-error btn-outline',
            icon: deleteIcon,
            disabled: (selected) => selected.length === 0,
            onclick: (selected) => {
                if (!actionStore) return;
                return Promise.all(selected.map(item => actionStore.remove(item.actionNumber)));
            }
        }
    ];

    function updateDescription(e: Event) {
        if (!cue) return;
        const value = (e.target as HTMLTextAreaElement).value;
        cues.byListNumber(cueListNumber).updateMetadata(cue.cueNumber, 'description', value);
    }
</script>

{#snippet addIcon()}
    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-4 h-4 mr-2">
        <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
    </svg>
{/snippet}

{#snippet deleteIcon()}
    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-4 h-4 mr-2">
        <path stroke-linecap="round" stroke-linejoin="round" d="m14.74 9-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 0 0-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 0 0-7.5 0" />
    </svg>
{/snippet}

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
                    class="textarea textarea-bordered h-24" 
                    placeholder="Enter cue description..."
                    value={cue.description}
                    onchange={updateDescription}
                ></textarea>
            </div>

            <div class="divider">Actions</div>

            <div class="h-64 border rounded-lg overflow-hidden">
                {#if actionStore}
                    <DataTable
                        items={$actionStore}
                        {columns}
                        {toolbar}
                        rowKey={(item) => item.actionNumber}
                    />
                {/if}
            </div>
        </div>
    {/if}
</Modal>
