<script lang="ts">
    import { type Snippet } from 'svelte';

    interface Props {
        open: boolean;
        onClose: () => void;
        title?: string;
        children?: Snippet;
        actions?: Snippet;
    }

    let { open, onClose, title, children, actions }: Props = $props();

    function handleKeydown(e: KeyboardEvent) {
        if (e.key === 'Escape' && open) {
            onClose();
        }
    }
</script>

{#if open}
    <div class="modal modal-open">
        <div class="modal-box w-[90vw] max-w-none h-[90vh] max-h-none flex flex-col relative">
            <button 
                class="btn btn-lg btn-circle btn-ghost absolute right-2 top-2 text-2xl" 
                onclick={onClose}
                aria-label="Close modal"
            >
                ✕
            </button>
            
            {#if title}
                <h3 class="font-bold text-lg">{title}</h3>
            {/if}
            
            <div class="grow py-4 overflow-auto">
                {@render children?.()}
            </div>
            
            {#if actions}
                <div class="modal-action">
                    {@render actions()}
                </div>
            {/if}
        </div>
        
        <div
            class="modal-backdrop"
            onclick={onClose}
            onkeydown={handleKeydown}
            role="button"
            tabindex="0"
            aria-label="Close modal"
        >
            <button class="cursor-default" tabindex="-1">close</button>
        </div>
    </div>
{/if}
