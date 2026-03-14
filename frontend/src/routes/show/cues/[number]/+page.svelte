<script lang="ts">
    import { page } from '$app/state';
    import { pageTitle } from '$lib/stores/header';

    $effect(() => {
        pageTitle.set(`Cue List: ${page.params.number}`);
    });
</script>

<div class="p-4">
    <h1 class="text-2xl font-bold">Cue List {page.params.number}</h1>
    <p class="mt-4">This page is under construction.</p>
    <a href="/show/cues" class="btn btn-primary mt-4">Back to Cue Lists</a>
</div>
