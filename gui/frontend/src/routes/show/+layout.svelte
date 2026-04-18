<!--
  This Source Code Form is subject to the terms of the Mozilla Public
  License, v. 2.0. If a copy of the MPL was not distributed with this
  file, You can obtain one at https://mozilla.org/MPL/2.0/.
-->

<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import {
		NewShow,
		SaveShow,
		SaveShowAs,
		OpenShow,
		CloseShow
	} from '../../../bindings/github.com/stexxo/dynocue/gui/selector.js';
	import { withLoading, createLoadingState } from '$lib/loading.svelte';

	let { children } = $props();

	const loadingState = createLoadingState();
	const handleNewShow = withLoading(loadingState, NewShow);
	const handleSave = withLoading(loadingState, SaveShow);
	const handleSaveAs = withLoading(loadingState, SaveShowAs);
	const handleOpen = withLoading(loadingState, OpenShow);
	const handleClose = withLoading(loadingState, CloseShow);

</script>

<div class="navbar h-12 min-h-12 border-b border-base-300 bg-base-100 px-4">
	<div class="flex-none">
		<ul class="menu menu-horizontal p-0">
			<li>
				<details>
					<summary class="px-3 py-1">File</summary>
					<ul class="z-50 w-56 rounded-md border border-base-300 bg-base-100 p-2 shadow-lg">
						<li><button onclick={handleNewShow}>New</button></li>
						<li><button onclick={handleSave}>Save</button></li>
						<li><button onclick={handleSaveAs}>Save As</button></li>
						<li><button onclick={handleOpen}>Open</button></li>
						<li><button onclick={handleClose}>Close</button></li>
					</ul>
				</details>
			</li>
		</ul>
	</div>
</div>

{#if loadingState.isLoading}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
		<span class="loading loading-lg loading-spinner text-primary"></span>
	</div>
{/if}

<main class="pb-24">
	{@render children()}
</main>

<div class="dock-bottom dock">
	<button
		onclick={() => goto('/show/dashboard')}
		class="dock-item"
		class:dock-active={page.url.pathname === '/show/dashboard'}
	>
		<span class="dock-label">Dashboard</span>
	</button>
	<button
		onclick={() => goto('/show/cues')}
		class="dock-item"
		class:dock-active={page.url.pathname === '/show/cues'}
	>
		<span class="dock-label">Cues</span>
	</button>
	<button
		onclick={() => goto('/show/audio')}
		class="dock-item"
		class:dock-active={page.url.pathname === '/show/audio'}
	>
		<span class="dock-label">Audio</span>
	</button>
	<button
		onclick={() => goto('/show/video')}
		class="dock-item"
		class:dock-active={page.url.pathname === '/show/video'}
	>
		<span class="dock-label">Video</span>
	</button>
	<button
		onclick={() => goto('/show/lighting')}
		class="dock-item"
		class:dock-active={page.url.pathname === '/show/lighting'}
	>
		<span class="dock-label">Lighting</span>
	</button>
	<button
		onclick={() => goto('/show/settings')}
		class="dock-item"
		class:dock-active={page.url.pathname === '/show/settings'}
	>
		<span class="dock-label">Settings</span>
	</button>
</div>
