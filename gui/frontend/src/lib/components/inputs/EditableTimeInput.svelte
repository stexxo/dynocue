<script lang="ts">
    import { formatTime } from "$lib/utils/time";

    interface EditableTimeInputProps {
        value: number; // nanoseconds
        onSave: (value: number) => void;
        label?: string;
        inputClass?: string;
        onCancel?: () => void;
        autofocus?: boolean;
    }

    const props: EditableTimeInputProps = $props();

    let editValue = $state("");
    let hasChanged = $derived(editValue !== formatTime(props.value));

    export { formatTime };

    $effect(() => {
        editValue = formatTime(props.value);
    })

    function handleSave() {
        const parts = editValue.split(':');
        let h = 0, m = 0, s = 0, ms = 0;

        if (parts.length === 3) {
            h = parseInt(parts[0], 10) || 0;
            m = parseInt(parts[1], 10) || 0;
            const secParts = parts[2].split('.');
            s = parseInt(secParts[0], 10) || 0;
            if (secParts.length > 1) {
                ms = parseInt(secParts[1].padEnd(3, '0').substring(0, 3), 10) || 0;
            }
        } else if (parts.length === 2) {
            m = parseInt(parts[0], 10) || 0;
            const secParts = parts[1].split('.');
            s = parseInt(secParts[0], 10) || 0;
            if (secParts.length > 1) {
                ms = parseInt(secParts[1].padEnd(3, '0').substring(0, 3), 10) || 0;
            }
        } else if (parts.length === 1 && parts[0] !== "") {
            const secParts = parts[0].split('.');
            s = parseInt(secParts[0], 10) || 0;
            if (secParts.length > 1) {
                ms = parseInt(secParts[1].padEnd(3, '0').substring(0, 3), 10) || 0;
            }
        } else if (parts[0] === "") {
            props.onSave(0);
            return;
        }

        const totalMs = (((h * 3600) + (m * 60) + s) * 1000) + ms;
        const valueToSave = totalMs * 1000000;
        props.onSave(valueToSave);
    }

    function handleCancel() {
        if (props.onCancel) {
            props.onCancel();
        } else {
            editValue = formatTime(props.value);
        }
    }

    function onInput(e: Event) {
        const input = e.target as HTMLInputElement;
        let val = input.value.replace(/[^\d:.]/g, '');
        const dotIndex = val.indexOf('.');
        if (dotIndex !== -1) {
            val = val.substring(0, dotIndex + 1) + val.substring(dotIndex + 1).replace(/\./g, '');
        }
        editValue = val;
    }

    function focus(node: HTMLInputElement) {
        if (props.autofocus) {
            node.focus();
        }
    }
</script>

<div class="form-control w-full relative">
    {#if props.label}
        <label class="label pb-1">
            <span class="label-text">{props.label}</span>
        </label>
    {/if}
    <div class="flex flex-row gap-2">
        <div class="flex-1">
            <input
                type="text"
                class="input input-bordered w-full {props.inputClass ?? ''}"
                bind:value={editValue}
                oninput={onInput}
                onblur={handleSave}
                onkeydown={(e) => {
                    if (e.key === 'Enter') handleSave();
                    if (e.key === 'Escape') handleCancel();
                }}
                use:focus
            />
        </div>
        <div class="flex flex-row gap-1 items-start {props.label ? 'pt-1' : ''}">
            {#if hasChanged}
                <button class="btn btn-square btn-sm btn-success" onclick={handleSave} title="Save">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                    </svg>
                </button>
                <button class="btn btn-square btn-sm btn-ghost" onclick={handleCancel} title="Cancel">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                </button>
            {/if}
        </div>
    </div>
</div>
