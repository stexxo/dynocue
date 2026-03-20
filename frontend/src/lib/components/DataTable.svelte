<script lang="ts">
    import { SvelteSet } from 'svelte/reactivity';
    import type { Snippet } from 'svelte';

    export interface ToolbarButton<T> {
        label: string;
        onclick: (selectedItems: T[]) => Promise<any> | any;
        class?: string;
        icon?: Snippet;
        disabled?: (selectedItems: T[]) => boolean;
        divider?: boolean;
    }

    export interface ColumnConfig<T> {
        key?: keyof T;
        label: string;
        width?: string;
        minWidth?: string;
        editable?: boolean;
        onSave?: (item: T, newValue: string) => Promise<any> | any;
        snippet?: Snippet<[T]>;
        align?: 'left' | 'right' | 'center';
    }

    interface Props<T extends { number?: number; cueNumber?: number; cueListNumber?: number; actionNumber?: number }> {
        items: T[];
        columns: ColumnConfig<T>[];
        toolbar?: ToolbarButton<T>[];
        isActive?: (item: T) => boolean;
    }

    let {
        items,
        columns,
        toolbar = [],
        isActive = () => false
    }: Props<any> = $props();

    function getItemNumber(item: any): number {
        return item.cueListNumber ?? item.cueNumber ?? item.actionNumber ?? item.number;
    }

    let selectedNumbers = $state(new SvelteSet<number>());
    let editingCell = $state<{ number: number; key: any } | null>(null);
    let editValue = $state('');
    let initialValue = $state('');

    let allSelected = $derived(items.length > 0 && selectedNumbers.size === items.length);
    let someSelected = $derived(selectedNumbers.size > 0 && !allSelected);

    let selectAllCheckbox = $state<HTMLInputElement | null>(null);

    $effect(() => {
        if (selectAllCheckbox) {
            selectAllCheckbox.indeterminate = someSelected;
        }
    });

    function toggleSelectAll() {
        if (allSelected) {
            selectedNumbers.clear();
        } else {
            for (const item of items) {
                selectedNumbers.add(getItemNumber(item));
            }
        }
    }

    function toggleSelect(number: number) {
        if (selectedNumbers.has(number)) {
            selectedNumbers.delete(number);
        } else {
            selectedNumbers.add(number);
        }
    }

    let selectedItems = $derived(items.filter(item => selectedNumbers.has(getItemNumber(item))));

    async function handleToolbarClick(button: ToolbarButton<any>) {
        await button.onclick(selectedItems);
        // If the action was likely a deletion or something that should clear selection, 
        // we should probably clear selection if the items are no longer in 'items'
        // but for now let's just let the caller handle it or we can clear if they want.
        // Usually, if you delete selected, you want the selection cleared.
        // We'll clear it if the number of selected items changes significantly or just always clear?
        // Actually, let's just clear it to be safe if it's a multi-item action.
        if (selectedNumbers.size > 0) {
            selectedNumbers.clear();
        }
    }

    function startEditing(item: any, column: ColumnConfig<any>) {
        if (!column.editable || !column.key) return;
        editingCell = { number: getItemNumber(item), key: column.key };
        editValue = String(item[column.key]);
        initialValue = editValue;
    }

    async function saveEdit(item: any, column: ColumnConfig<any>) {
        if (!editingCell) return;

        if (editValue !== initialValue) {
            if (column.onSave) {
                await column.onSave(item, editValue);
            }
        }
        editingCell = null;
    }

    function cancelEdit() {
        editingCell = null;
    }

    function handleKeyDown(e: KeyboardEvent, item: any, column: ColumnConfig<any>) {
        if (e.key === 'Enter') {
            saveEdit(item, column);
        } else if (e.key === 'Escape') {
            cancelEdit();
        }
    }
</script>

<div class="flex flex-col h-full">
    <div class="flex items-center justify-start bg-base-200 p-2 gap-2">
        {#each toolbar as button}
            {#if button.divider}
                <div class="divider divider-horizontal mx-0"></div>
            {:else}
                <button
                    class="btn btn-sm {button.class || 'btn-ghost'}"
                    onclick={() => handleToolbarClick(button)}
                    disabled={button.disabled?.(selectedItems)}
                >
                    {#if button.icon}
                        {@render button.icon()}
                    {/if}
                    {button.label}
                </button>
            {/if}
        {/each}
    </div>

    <div class="overflow-x-auto overflow-y-auto grow">
        <table class="table table-pin-rows">
            <thead>
                <tr>
                    <th class="w-10">
                        <label>
                            <input type="checkbox" class="checkbox checkbox-sm"
                                   bind:this={selectAllCheckbox}
                                   checked={allSelected}
                                   onchange={toggleSelectAll} />
                        </label>
                    </th>
                    {#each columns as column}
                        <th class="{column.width} {column.minWidth} {column.align === 'right' ? 'text-right' : ''} {column.align === 'center' ? 'text-center' : ''}">{column.label}</th>
                    {/each}
                </tr>
            </thead>
            <tbody>
                {#each items as item (getItemNumber(item))}
                    <tr class="h-12 {selectedNumbers.has(getItemNumber(item)) ? 'bg-base-300' : ''} {isActive(item) ? 'bg-primary text-primary-content' : ''}">
                        <td>
                            <label>
                                <input type="checkbox" class="checkbox checkbox-sm"
                                       checked={selectedNumbers.has(getItemNumber(item))}
                                       onchange={() => toggleSelect(getItemNumber(item))} />
                            </label>
                        </td>
                        {#each columns as column}
                            {@const isEditing = column.key && editingCell?.number === getItemNumber(item) && editingCell?.key === column.key}
                            <td
                                onclick={() => startEditing(item, column)}
                                class="p-0 {column.editable ? 'cursor-pointer hover:bg-base-200 transition-colors' : ''}"
                                title={column.editable ? `Click to edit ${column.label.toLowerCase()}` : ''}
                            >
                                <div class="flex items-center h-full px-4 {column.align === 'right' ? 'justify-end' : ''} {column.align === 'center' ? 'justify-center' : ''}">
                                    {#if isEditing && column.key}
                                        <input
                                            type="text"
                                            class="input input-bordered input-sm w-full h-8"
                                            bind:value={editValue}
                                            onkeydown={(e) => handleKeyDown(e, item, column)}
                                            onblur={() => saveEdit(item, column)}
                                            autofocus
                                        />
                                    {:else if column.snippet}
                                        {@render column.snippet(item)}
                                    {:else if column.key}
                                        {item[column.key]}
                                    {/if}
                                </div>
                            </td>
                        {/each}
                    </tr>
                {/each}
            </tbody>
        </table>
    </div>
</div>
