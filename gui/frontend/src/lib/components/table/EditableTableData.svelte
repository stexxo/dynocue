<script lang="ts">
    let editing = $state(false)

    const props: EditableTableDataProps = $props()

    let editValue = $derived(props.value)

    function onSave() {
        props.onSaveEdit(editValue)
        editing = false
    }

    function onCancel() {
        editing = false
        editValue = props.value
    }

    function focus(node: HTMLInputElement) {
        node.focus();
    }
</script>

<td class="relative {props.tdClass}" onclick={() => {editing=true}}>
    {#if editing}
        <div class="flex flex-row w-full gap-2">
            <input
                    type={props.inputType}
                    class="input input-bordered input-sm w-full"
                    bind:value={editValue}
                    onblur={onSave}
                    onkeydown={(e) => {
									if (e.key === 'Enter') onSave();
									if (e.key === 'Escape') onCancel();
								}}
                    use:focus
            />
            <div class="flex flex-row gap-1">
                <button class="btn btn-sm btn-secondary w-7" onclick={onSave}>✓</button>
                <button class="btn btn-sm btn-accent w-7" onclick={onCancel}>x</button>
            </div>
        </div>
    {:else}
        <div class="hover:border min-h-10 cursor-pointer p-2">
            {props.value}
        </div>
    {/if}
</td>