<script lang="ts">
	import { type TabProps } from './tabTypes.svelte';
	let props: TabProps = $props();

	let activeTab = $derived(props.tabState.getActive());
</script>

<div class="flex h-full flex-col">
	<div role="tablist" class="tabs-lifted tabs flex-none">
		{#each props.tabState.items as tab}
			<button
				type="button"
				role="tab"
				class="tab gap-2 {tab.id === props.tabState.activeId ? 'tab-active' : ''}"
				onclick={() => props.tabState.select(tab.id)}
			>
				{tab.label}

				{#if tab.closable}
					<button
						type="button"
						class="btn z-10 btn-circle btn-xs"
						onclick={(e) => {
							e.stopPropagation();
							props.tabState.closeTab(tab.id);
						}}
					>
						✕
					</button>
				{/if}
			</button>
		{/each}
	</div>

	<div
		class="-mt-[var(--tab-border)] flex-1 overflow-auto rounded-b-box border border-base-300 bg-base-100 p-6"
	>
		{#if activeTab?.content}
			{@const Content = activeTab.content}
			<Content {...activeTab?.props} />
		{/if}
	</div>
</div>
