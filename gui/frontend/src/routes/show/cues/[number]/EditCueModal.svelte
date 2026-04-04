<script lang="ts">
    import Modal from '$lib/components/Modal.svelte';
    import type { Cue } from '../../../../../bindings/github.com/stexxo/dynocue/api/cues';
    import { actions } from '$lib/stores/actions';
    import DataTable, { type ColumnConfig, type ToolbarButton } from '$lib/components/DataTable.svelte';
    import type { CueAction } from '../../../../../bindings/github.com/stexxo/dynocue/api/cues';

    interface Props {
        cueListNumber: number;
        cue: Cue | null;
        onClose: () => void;
    }

    let { cueListNumber, cue, onClose }: Props = $props();

    const actionStore = $derived(actions.byCue(cueListNumber, cue?.cueNumber ?? 0));

    const columns: ColumnConfig<CueAction>[] = [
        { key: 'actionNumber', label: '#', class: 'w-16' },
        { key: 'label', label: 'Label', editable: true, onSave: (item, val) => actionStore?.updateAction(item.actionNumber, 'label', val) },
        { key: 'sourceType', label: 'Type', editable: true, onSave: (item, val) => actionStore?.updateAction(item.actionNumber, 'sourceType', val) },
        { key: 'action', label: 'Action', editable: true, onSave: (item, val) => actionStore?.updateAction(item.actionNumber, 'action', val) },
        { key: 'target', label: 'Target', editable: true, onSave: (item, val) => actionStore?.updateAction(item.actionNumber, 'target', val) },
    ];

    const toolbar: ToolbarButton<CueAction>[] = [
        {
            label: 'Add Action',
            onclick: () => actionStore?.create(),
            class: 'btn-primary'
        },
        {
            label: 'Delete',
            onclick: (selected) => {
                selected.forEach(item => actionStore?.remove(item.actionNumber));
            },
            disabled: (selected) => selected.length === 0,
            class: 'btn-error'
        }
    ];

    $effect(() => {
        if (cue) {
            actionStore.refresh();
        }
    });

</script>

<Modal
    open={!!cue} 
    onClose={onClose} 
    title={cue ? `Cue: ${cue.cueNumber} - ${cue.label}` : ''}
>
    {#if cue}
        <DataTable
                items={$actionStore}
                columns={columns}
                toolbar={toolbar}
                rowKey={(item) => item.actionNumber}
        />
    {/if}
</Modal>
