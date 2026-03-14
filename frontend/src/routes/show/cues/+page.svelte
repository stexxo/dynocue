<script lang="ts">
    import { onMount } from 'svelte';
    import { cueLists } from '$lib/stores/cuelists';

    let editingId: number | null = null;
    let editValue = '';
    let initialValue = '';

    onMount(() => {
        cueLists.refresh();
    });

    function startEditing(number: number, currentLabel: string) {
        editingId = number;
        editValue = currentLabel;
        initialValue = currentLabel;
    }

    async function saveEdit() {
        if (editingId !== null) {
            if (editValue !== initialValue) {
                await cueLists.updateMetadata(editingId, "label", editValue);
            }
            editingId = null;
        }
    }

    function cancelEdit() {
        editingId = null;
    }

    function handleKeyDown(e: KeyboardEvent) {
        if (e.key === 'Enter') {
            saveEdit();
        } else if (e.key === 'Escape') {
            cancelEdit();
        }
    }
</script>

<div class="flex flex-col h-[calc(100vh-theme(spacing.12)-theme(spacing.10))]">
    <div class="flex items-center justify-start bg-base-200 p-2">
        <button class="btn btn-primary btn-sm" on:click={() => cueLists.create(0)}>
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-4 h-4 mr-2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
            </svg>
            Add Cue List
        </button>
    </div>

    <div class="overflow-x-auto overflow-y-auto flex-grow">
        <table class="table table-pin-rows">
            <!-- head -->
            <thead>
            <tr>
                <th>#</th>
                <th>Label</th>
                <th class="w-1/6"></th>
            </tr>
            </thead>
            <tbody>
            {#each $cueLists as list (list.number)}
                <tr>
                    <td>{list.number}</td>
                    <td on:dblclick={() => startEditing(list.number, list.label)}>
                        {#if editingId === list.number}
                            <input
                                type="text"
                                class="input input-bordered input-sm w-full"
                                bind:value={editValue}
                                on:keydown={handleKeyDown}
                                on:blur={saveEdit}
                                autofocus
                            />
                        {:else}
                            {list.label}
                        {/if}
                    </td>
                    <td class="flex justify-end gap-2">
                        <button class="btn btn-sm btn-error btn-outline" on:click={() => cueLists.remove(list.number)}>Delete</button>
                    </td>
                </tr>
            {/each}
            </tbody>
        </table>
    </div>
</div>


