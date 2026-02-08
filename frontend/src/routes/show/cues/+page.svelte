<script>
    import { CueService } from "../../../../bindings/gitlab.com/stexxo/dynocue/dynod/internal/subsystems/gui/api/index.js"
    import { cueStore } from '$lib/stores/cueStore';

</script>

{#snippet cueListTableRow(cueList)}
    <tr class="hover:bg-base-300 hover:cursor-pointer">
        <th>{cueList.Number}</th>
        <td>{cueList.Label}</td>
        <td class="flex flex-row gap-2">
            <button class="btn btn-primary btn-ghost">Edit</button>
            <button class="btn btn-error btn-ghost">Delete</button>
        </td>
    </tr>
{/snippet}

<div class="overflow-x-auto">
    <table class="table">
        <!-- head -->
        <thead>
        <tr>
            <th>#</th>
            <th>Label</th>
            <th class="w-1/6"></th>
        </tr>
        </thead>
        <tbody>
            {#each $cueStore as cueList}
                {@render cueListTableRow(cueList)}
            {/each}
        </tbody>
    </table>
</div>

<button on:click={()=>{CueService.CreateCueList()}} class="btn btn-xl btn-circle  bg-slate-500 hover:brightness-80  fixed bottom-15 right-5 shadow-md shadow-black">
        <svg
                aria-label="New"
                xmlns="http://www.w3.org/2000/svg"
                fill="none"
                viewBox="0 0 24 24"
                stroke-width="2"
                stroke="currentColor"
                class="size-6"
        >
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
        </svg>
</button>

