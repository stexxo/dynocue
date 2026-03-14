<script lang="ts">
    import { onMount } from 'svelte';
    import { cueLists } from '$lib/stores/cuelists';

    let editingId: number | null = null;
    let editingNumber: number | null = null;
    let editValue = '';
    let editNumberValue = '';
    let initialValue = '';
    let selectedNumbers = new Set<number>();
    let allSelected = false;
    let someSelected = false;

    onMount(() => {
        cueLists.refresh();
    });

    // Reactive selection state
    $: {
        allSelected = $cueLists.length > 0 && selectedNumbers.size === $cueLists.length;
        someSelected = selectedNumbers.size > 0 && !allSelected;
    }

    let selectAllCheckbox: HTMLInputElement;
    $: if (selectAllCheckbox) {
        selectAllCheckbox.indeterminate = someSelected;
    }

    function toggleSelectAll() {
        if (allSelected) {
            selectedNumbers.clear();
        } else {
            selectedNumbers = new Set($cueLists.map(l => l.number));
        }
        selectedNumbers = selectedNumbers; // Trigger reactivity
    }

    function toggleSelect(number: number) {
        if (selectedNumbers.has(number)) {
            selectedNumbers.delete(number);
        } else {
            selectedNumbers.add(number);
        }
        selectedNumbers = selectedNumbers; // Trigger reactivity
    }

    async function deleteSelected() {
        const toDelete = Array.from(selectedNumbers);
        for (const num of toDelete) {
            await cueLists.remove(num);
        }
        selectedNumbers.clear();
        selectedNumbers = selectedNumbers;
    }

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

    function startEditingNumber(number: number) {
        editingNumber = number;
        editNumberValue = number.toString();
        initialValue = number.toString();
    }

    async function saveNumberEdit() {
        if (editingNumber !== null) {
            if (editNumberValue !== initialValue) {
                const newNum = parseFloat(editNumberValue);
                if (!isNaN(newNum) && newNum > 0) {
                    await cueLists.move(editingNumber, newNum);
                }
            }
            editingNumber = null;
        }
    }

    function cancelNumberEdit() {
        editingNumber = null;
    }

    function handleKeyDown(e: KeyboardEvent) {
        if (e.key === 'Enter') {
            saveEdit();
        } else if (e.key === 'Escape') {
            cancelEdit();
        }
    }

    function handleNumberKeyDown(e: KeyboardEvent) {
        if (e.key === 'Enter') {
            saveNumberEdit();
        } else if (e.key === 'Escape') {
            cancelNumberEdit();
        }
    }
</script>

<div class="flex flex-col h-[calc(100vh-theme(spacing.12)-theme(spacing.10))]">
    <div class="flex items-center justify-start bg-base-200 p-2 gap-2">
        <button class="btn btn-primary btn-sm" on:click={() => cueLists.create(0)}>
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-4 h-4 mr-2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
            </svg>
            Add Cue List
        </button>

        <button 
            class="btn btn-error btn-sm btn-outline" 
            on:click={deleteSelected}
            disabled={selectedNumbers.size === 0}
        >
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-4 h-4 mr-2">
                <path stroke-linecap="round" stroke-linejoin="round" d="m14.74 9-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 0 0-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 0 0-7.5 0" />
            </svg>
            Delete Selected ({selectedNumbers.size})
        </button>
    </div>

    <div class="overflow-x-auto overflow-y-auto flex-grow">
        <table class="table table-pin-rows">
            <!-- head -->
            <thead>
            <tr>
                <th class="w-10">
                    <label>
                        <input type="checkbox" class="checkbox checkbox-sm" 
                               bind:this={selectAllCheckbox}
                               checked={allSelected} 
                               on:change={toggleSelectAll} />
                    </label>
                </th>
                <th class="w-24">#</th>
                <th class="min-w-50">Label</th>
                <th class="w-10 text-right"></th>
            </tr>
            </thead>
            <tbody>
            {#each $cueLists as list (list.number)}
                <tr class="h-12 {selectedNumbers.has(list.number) ? 'bg-base-300' : ''}">
                    <td>
                        <label>
                            <input type="checkbox" class="checkbox checkbox-sm" 
                                   checked={selectedNumbers.has(list.number)}
                                   on:change={() => toggleSelect(list.number)} />
                        </label>
                    </td>
                    <td on:click={() => startEditingNumber(list.number)} class="p-0 cursor-pointer hover:bg-base-200 transition-colors" title="Click to edit number">
                        <div class="flex items-center h-full px-4">
                            {#if editingNumber === list.number}
                                <input
                                    type="text"
                                    class="input input-bordered input-sm w-full h-8"
                                    bind:value={editNumberValue}
                                    on:keydown={handleNumberKeyDown}
                                    on:blur={saveNumberEdit}
                                    autofocus
                                />
                            {:else}
                                {list.number}
                            {/if}
                        </div>
                    </td>
                    <td on:click={() => startEditing(list.number, list.label)} class="p-0 cursor-pointer hover:bg-base-200 transition-colors" title="Click to edit label">
                        <div class="flex items-center h-full px-4">
                            {#if editingId === list.number}
                                <input
                                    type="text"
                                    class="input input-bordered input-sm w-full h-8"
                                    bind:value={editValue}
                                    on:keydown={handleKeyDown}
                                    on:blur={saveEdit}
                                    autofocus
                                />
                            {:else}
                                {list.label}
                            {/if}
                        </div>
                    </td>
                    <td class="p-0">
                        <div class="flex items-center justify-end h-full px-4">
                            <a href="/show/cues/{list.number}" class="btn btn-ghost btn-xs btn-square">
                                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-4 h-4">
                                    <path stroke-linecap="round" stroke-linejoin="round" d="m8.25 4.5 7.5 7.5-7.5 7.5" />
                                </svg>
                            </a>
                        </div>
                    </td>
                </tr>
            {/each}
            </tbody>
        </table>
    </div>
</div>


