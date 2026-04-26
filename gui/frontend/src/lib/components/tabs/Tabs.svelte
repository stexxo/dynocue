<!--
  This Source Code Form is subject to the terms of the Mozilla Public
  License, v. 2.0. If a copy of the MPL was not distributed with this
  file, You can obtain one at https://mozilla.org/MPL/2.0/.
-->

<script lang="ts">
	import { type TabProps, TabState } from './tabTypes.svelte';
	let props: TabProps = $props();

	let activeTab = $derived(props.tabManager.getActive());
</script>

<div class="flex h-full flex-col">
	<div role="tablist" class="tabs-lifted tabs flex-none">
		{#each props.tabManager.items as tab, i}
			<div
				role="tab"
				tabindex={i}
				class="tab gap-2 {tab.id === props.tabManager.activeId ? 'tab-active' : ''}"
				onclick={() => props.tabManager.select(tab.id)}
			>
				{new TabState(tab).label}

				{#if tab.closable}
					<button
						type="button"
						class="btn z-10 btn-circle btn-xs"
						onclick={(e) => {
							e.stopPropagation();
							props.tabManager.closeTab(tab.id);
						}}
					>
						✕
					</button>
				{/if}
			</div>
		{/each}
	</div>

	<div
		class="-mt-(--tab-border) min-h-0 flex-1 rounded-b-box border border-base-300 bg-base-100 p-6"
	>
		{#if activeTab?.content}
			{@const Content = activeTab.content}
			<Content {...activeTab?.props} tabState={new TabState(activeTab)} />
		{/if}
	</div>
</div>
