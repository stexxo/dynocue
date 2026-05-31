<!--
  This Source Code Form is subject to the terms of the Mozilla Public
  License, v. 2.0. If a copy of the MPL was not distributed with this
  file, You can obtain one at https://mozilla.org/MPL/2.0/.
-->

<script lang="ts">
	import AudioFileTable from '$lib/components/audio/AudioFileTable.svelte';
	import { audioStore } from '$lib/stores/audioStore.svelte';

	let dragCounter = $state(0);
	let isDragging = $derived(dragCounter > 0);

	function onDragEnter(e: DragEvent) {
		e.preventDefault();
		dragCounter++;
	}

	function onDragOver(e: DragEvent) {
		e.preventDefault();
		if (e.dataTransfer) {
			e.dataTransfer.dropEffect = 'copy';
		}
	}

	function onDragLeave() {
		dragCounter--;
	}

	async function onDrop(e: DragEvent) {
		e.preventDefault();
		dragCounter = 0;
	}
</script>

<div
	class="drop-zone relative h-full p-4"
	data-file-drop-target
	ondragenter={onDragEnter}
	ondragover={onDragOver}
	ondragleave={onDragLeave}
	ondrop={onDrop}
	role="region"
	id="audio-file-drop"
	aria-label="Audio file drop zone"
>
	{#if isDragging}
		<div
			class="pointer-events-none absolute inset-0 z-50 m-4 flex items-center justify-center rounded-xl border-4 border-dashed border-primary bg-base-300/50"
		>
			<div class="text-2xl font-bold text-primary">Drop audio files here</div>
		</div>
	{/if}
	<AudioFileTable />
</div>
