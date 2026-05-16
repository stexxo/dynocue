<!--
  This Source Code Form is subject to the terms of the Mozilla Public
  License, v. 2.0. If a copy of the MPL was not distributed with this
  file, You can obtain one at https://mozilla.org/MPL/2.0/.
-->

<script lang="ts">
	import type { TabContentProps } from '$lib/components/tabs/tabTypes.svelte';
	import CueTable from '$lib/components/cueing/cues/CueTable.svelte';
	import CueEdit from './CueEdit.svelte';
	import { GoToNextCue } from '../../../../bindings/github.com/stexxo/dynocue/gui/services/executionservice';
	const props: TabContentProps = $props();

	let cueEdit: ReturnType<typeof CueEdit>;
</script>

<div class="flex h-full flex-col p-4">
	<div class="mb-4">
		<button class="btn w-full font-bold btn-lg btn-primary" onclick={() => GoToNextCue(props.id)}>
			Go
		</button>
	</div>
	<CueTable onEdit={(listId, id) => cueEdit.show(listId, id)} CueListId={props.id} />
</div>

<CueEdit bind:this={cueEdit} />
