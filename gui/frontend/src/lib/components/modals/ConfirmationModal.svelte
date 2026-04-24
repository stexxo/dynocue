<script lang="ts">
    interface ConfirmationModalProps {
        title: string;
        message: string;
        confirmText?: string;
        cancelText?: string;
        onConfirm: () => void;
        onCancel?: () => void;
    }

    const {
        title,
        message,
        confirmText = "Confirm",
        cancelText = "Cancel",
        onConfirm,
        onCancel
    }: ConfirmationModalProps = $props();

    let dialog: HTMLDialogElement | undefined = $state();

    export function show() {
        dialog?.showModal();
    }

    export function close() {
        dialog?.close();
    }

    function handleConfirm() {
        onConfirm();
        close();
    }

    function handleCancel() {
        if (onCancel) onCancel();
        close();
    }
</script>

<dialog bind:this={dialog} class="modal">
    <div class="modal-box">
        <h3 class="font-bold text-lg">{title}</h3>
        <p class="py-4">{message}</p>
        <div class="modal-action">
            <button class="btn" onclick={handleCancel}>{cancelText}</button>
            <button class="btn btn-error" onclick={handleConfirm}>{confirmText}</button>
        </div>
    </div>
    <form method="dialog" class="modal-backdrop">
        <button onclick={handleCancel}>close</button>
    </form>
</dialog>
