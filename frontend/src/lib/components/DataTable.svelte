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
        class?: string;
        editable?: boolean;
        onSave?: (item: T, newValue: string) => Promise<any> | any;
        snippet?: Snippet<[T]>;
    }

    interface Props<T> {
        items: T[];
        columns: ColumnConfig<T>[];
        toolbar?: ToolbarButton<T>[];
        activeKey: string | number;
        rowKey: (item: T) => string | number;
    }

    let {
        items,
        columns,
        toolbar = [],
        activeKey,
        rowKey
    }: Props<T> = $props();

    let selectedKeys = $state(new SvelteSet<string | number>());
    let editingCell = $state<{ key: string | number; columnKey: any } | null>(null);
    let editValue = $state('');
    let initialValue = $state('');

    let allSelected = $derived(items.length > 0 && selectedKeys.size === items.length);
    let someSelected = $derived(selectedKeys.size > 0 && !allSelected);

    let selectAllCheckbox = $state<HTMLInputElement | null>(null);

    $effect(() => {
        if (selectAllCheckbox) {
            selectAllCheckbox.indeterminate = someSelected;
        }
    });

    function toggleSelectAll() {
        if (allSelected) {
            selectedKeys.clear();
        } else {
            for (const item of items) {
                selectedKeys.add(rowKey(item));
            }
        }
    }

    function toggleSelect(key: string | number) {
        if (selectedKeys.has(key)) {
            selectedKeys.delete(key);
        } else {
            selectedKeys.add(key);
        }
    }

    let selectedItems = $derived(items.filter(item => selectedKeys.has(rowKey(item))));

    async function handleToolbarClick(button: ToolbarButton<any>) {
        await button.onclick(selectedItems);
        if (selectedKeys.size > 0) {
            selectedKeys.clear();
        }
    }

    function startEditing(item: any, column: ColumnConfig<any>) {
        if (!column.editable || !column.key) return;
        editingCell = { key: rowKey(item), columnKey: column.key };
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
                        <th class="{column.class}">{column.label}</th>
                    {/each}
                </tr>
            </thead>
            <tbody>
                {#each items as item (rowKey(item))}
                    <tr class="h-12 {selectedKeys.has(rowKey(item)) ? 'bg-base-300' : ''} {activeKey === rowKey(item) ? 'bg-primary text-primary-content' : ''}">
                        <td>
                            <label>
                                <input type="checkbox" class="checkbox checkbox-sm"
                                       checked={selectedKeys.has(rowKey(item))}
                                       onchange={() => toggleSelect(rowKey(item))} />
                            </label>
                        </td>
                        {#each columns as column}
                            {@const isEditing = column.key && editingCell?.key === rowKey(item) && editingCell?.columnKey === column.key}
                            <td
                                onclick={() => startEditing(item, column)}
                                class="p-0 {column.editable ? 'cursor-pointer hover:bg-base-200 transition-colors' : ''}"
                                title={column.editable ? `Click to edit ${column.label.toLowerCase()}` : ''}
                            >
                                <div class="flex items-center h-full px-4">
                                    {#if isEditing && column.key}
                                        <input
                                            type="text"
                                            class="input input-bordered input-sm w-full h-8"
                                            bind:value={editValue}
                                            onkeydown={(e) => handleKeyDown(e, item, column)}
                                            onblur={() => saveEdit(item, column)}
                                            use={(node) => node.focus()}
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
