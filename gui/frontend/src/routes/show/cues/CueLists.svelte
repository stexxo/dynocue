<!--
  This Source Code Form is subject to the terms of the Mozilla Public
  License, v. 2.0. If a copy of the MPL was not distributed with this
  file, You can obtain one at https://mozilla.org/MPL/2.0/.
-->

<script lang="ts">
	import CueListTable from '$lib/components/cueing/cuelists/CueListTable.svelte';
	import { cueListTabState } from './cuesListsState.svelte';
	import CueListDetail from './CueListDetail.svelte';
	import type { TabContentProps } from '$lib/components/tabs/tabTypes.svelte';
	import { cuelistsStore } from '$lib/stores/cuelistsStore.svelte';

	const props: TabContentProps = $props();
	$effect(() => {
		props.tabState.setLabel('Cue Lists');
	});
</script>

<div class="flex h-full flex-col p-4">
	<CueListTable
		AllowCreation={true}
		OnOpenCueList={(id) => {
			const label = () => {
				const cl = cuelistsStore.cueList(id);
				return `Cue List ${cl?.number} ${cl?.label !== '' && cl?.label != null ? `- ${cl?.label}` : ''}`;
			};
			cueListTabState.addTab({
				id: id,
				content: CueListDetail,
				props: { id: id },
				closable: true,
				label: label
			});
		}}
	/>
</div>
