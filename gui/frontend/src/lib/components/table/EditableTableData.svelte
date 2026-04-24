<script lang="ts">
    import EditableTextInput from "../inputs/EditableTextInput.svelte";

    interface EditableTableDataProps {
        value: any;
        inputType: string;
        tdClass?: string;
        onSaveEdit: (value: any) => void;
    }

    let editing = $state(false)

    const props: EditableTableDataProps = $props()

    function onSave(value: any) {
        props.onSaveEdit(value)
        editing = false
    }

    function onCancel() {
        editing = false
    }
</script>

<td class="relative {props.tdClass}" onclick={() => {editing=true}}>
    {#if editing}
        <EditableTextInput
            value={props.value}
            type={props.inputType}
            onSave={onSave}
            onCancel={onCancel}
            inputClass="input-sm"
            autofocus={true}
        />
    {:else}
        <div class="hover:border min-h-10 cursor-pointer p-2">
            {props.value}
        </div>
    {/if}
</td>