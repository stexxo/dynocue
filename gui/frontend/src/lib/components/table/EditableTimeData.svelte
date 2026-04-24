<script lang="ts">
    import EditableTimeInput from "../inputs/EditableTimeInput.svelte";
    import { formatTime } from "$lib/utils/time";

    interface EditableTimeDataProps {
        value: number; // nanoseconds
        tdClass?: string;
        onSaveEdit: (value: number) => void;
    }

    let editing = $state(false)
    const props: EditableTimeDataProps = $props()

    function onSave(value: number) {
        props.onSaveEdit(value)
        editing = false
    }

    function onCancel() {
        editing = false
    }
</script>

<td class="relative {props.tdClass}" onclick={() => {editing=true}}>
    {#if editing}
        <EditableTimeInput
            value={props.value}
            onSave={onSave}
            onCancel={onCancel}
            inputClass="input-sm"
            autofocus={true}
        />
    {:else}
        <div class="hover:border min-h-10 cursor-pointer p-2">
            {formatTime(props.value)}
        </div>
    {/if}
</td>
