<script lang="ts">
    import { onMount } from 'svelte';
    import { page } from '$app/state';
    import { pageTitle } from '$lib/stores/header';
    import DataTable, { type ToolbarButton, type ColumnConfig } from '$lib/components/DataTable.svelte';
    import { cues, type Cue } from '$lib/stores/cues';
    import { goto } from '$app/navigation';

    const cueListNumber = parseFloat(page.params.number || '0');

    onMount(() => {
        cues.refresh(cueListNumber);
    });

    $effect(() => {
        pageTitle.set(`Cue List: ${page.params.number}`);
    });

    const columns: ColumnConfig<Cue>[] = [
        { 
            key: 'cueNumber', 
            label: '#', 
            width: 'w-24', 
            editable: true,
            onSave: (item: Cue, newValue: string) => {
                const newNum = parseFloat(newValue);
                if (!isNaN(newNum) && newNum > 0) {
                    return cues.move(item.cueNumber, newNum);
                }
            }
        },
        { 
            key: 'label', 
            label: 'Label', 
            minWidth: 'min-w-50', 
            editable: true,
            onSave: (item: Cue, newValue: string) => cues.updateMetadata(item.cueNumber, 'label', newValue)
        },
        {
            label: '',
            width: 'w-10',
            align: 'right',
            snippet: rowEnd
        }
    ];

    const toolbar: ToolbarButton<Cue>[] = [
        {
            label: 'Back',
            icon: backIcon,
            onclick: () => goto('/show/cues')
        },
        {
            label: '',
            onclick: () => {},
            divider: true 
        },
        {
            label: 'Add Cue',
            class: 'btn-primary',
            icon: addIcon,
            onclick: () => cues.create(0)
        },
        {
            label: 'Delete Selected',
            class: 'btn-error btn-outline',
            icon: deleteIcon,
            disabled: (selected) => selected.length === 0,
            onclick: (selected) => Promise.all(selected.map(item => cues.remove(item.cueNumber)))
        }
    ];
</script>

{#snippet backIcon()}
    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-4 h-4 mr-2">
        <path stroke-linecap="round" stroke-linejoin="round" d="M10.5 19.5 3 12m0 0 7.5-7.5M3 12h18" />
    </svg>
{/snippet}

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

{#snippet rowEnd(cue)}
    <button class="btn btn-ghost btn-xs btn-square">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-4 h-4">
            <path stroke-linecap="round" stroke-linejoin="round" d="M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L6.832 19.82a4.5 4.5 0 01-1.897 1.13l-2.685.8.8-2.685a4.5 4.5 0 011.13-1.897L16.863 4.487zm0 0l1.514-1.515" />
        </svg>
    </button>
{/snippet}

<div class="flex flex-col h-[calc(100vh-theme(spacing.12)-theme(spacing.10))]">
    <DataTable
        items={$cues}
        {columns}
        {toolbar}
    />
</div>
