<script lang="ts">
	import { type TabProps } from '$lib/components/tabTypes.svelte';
	let props: TabProps = $props();

	let activeTab = $derived(props.tabState.getActive());
</script>

<div role="tablist" class="tabs-lift tabs">
	{#each props.tabState.items as tab}
		<div
			role="tab"
			tabindex="0"
			class="tab gap-2 {tab.id === props.tabState.activeId ? 'tab-active' : ''}"
			onclick={(e)=>{
				  if (e.target !== e.currentTarget) return;
				props.tabState.select(tab.id)}}
			onkeydown={(e) => (e.key === 'Enter' || e.key === ' ') && props.tabState.select(tab.id)}
		>
			{tab.label}

			{#if tab.closable}
				<button
					type="button"
					class="btn relative z-10 btn-circle btn-xs"
					onclick={(e) => {
						e.stopPropagation();
						props.tabState.closeTab(tab.id);
					}}
				>
					✕
				</button>
			{/if}
		</div>
		<!--   <button class="tab {tab.id === props.tabState.activeId ? 'tab-active' : ''}" onclick="{() => props.tabState.select(tab.id)}">{tab.label}</button> -->
	{/each}
</div>

<div class="p-6 bg-base-100 border border-base-300 rounded-b-box h-full">
	{#if activeTab?.content}
		{@const Content = activeTab.content}
		<Content {...activeTab?.props} />
	{/if}
</div>
