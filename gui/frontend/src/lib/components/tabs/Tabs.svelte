<script lang="ts">
	import {type TabProps, TabState} from './tabTypes.svelte';
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
		class="-mt-(--tab-border) flex-1 min-h-0 rounded-b-box border border-base-300 bg-base-100 p-6"
	>
		{#if activeTab?.content}
			{@const Content = activeTab.content}
			<Content {...activeTab?.props} tabState={new TabState(activeTab)} />
		{/if}
	</div>
</div>
