<!--
  This Source Code Form is subject to the terms of the Mozilla Public
  License, v. 2.0. If a copy of the MPL was not distributed with this
  file, You can obtain one at https://mozilla.org/MPL/2.0/.
-->

<script lang="ts">
	import { page } from '$app/state';
	import { NewShow, SaveShow, SaveShowAs } from '../../../bindings/github.com/stexxo/dynocue/gui/selector.js';
	import { withLoading, createLoadingState } from '$lib/loading.svelte';

	let { children } = $props();

	const loadingState = createLoadingState();
	const handleNewShow = withLoading(loadingState, NewShow);
	const handleSave = withLoading(loadingState, SaveShow)
	const handleSaveAs = withLoading(loadingState, SaveShowAs)

	function handleOpen() {
		console.log('Open');
	}
</script>

<div class="navbar bg-base-100 border-b border-base-300 px-4 min-h-12 h-12">
	<div class="flex-none">
		<ul class="menu menu-horizontal p-0">
			<li>
				<details>
					<summary class="py-1 px-3">File</summary>
					<ul class="bg-base-100 rounded-md border border-base-300 z-50 p-2 shadow-lg w-56">
						<li><button onclick={handleNewShow}>New</button></li>
						<li><button onclick={handleSave}>Save</button></li>
						<li><button onclick={handleSaveAs}>Save As</button></li>
						<li><button onclick={handleOpen}>Open</button></li>
					</ul>
				</details>
			</li>
		</ul>
	</div>
</div>

{#if loadingState.isLoading}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
		<span class="loading loading-spinner loading-lg text-primary"></span>
	</div>
{/if}

<main class="pb-24">
	{@render children()}
</main>

<div class="dock dock-bottom">
	<a href="/show/dashboard" class="dock-item" class:dock-active={page.url.pathname === '/show/dashboard'}>
		<span class="dock-label">Dashboard</span>
	</a>
	<a href="/show/cues" class="dock-item" class:dock-active={page.url.pathname === '/show/cues'}>
		<span class="dock-label">Cues</span>
	</a>
	<a href="/show/audio" class="dock-item" class:dock-active={page.url.pathname === '/show/audio'}>
		<span class="dock-label">Audio</span>
	</a>
	<a href="/show/video" class="dock-item" class:dock-active={page.url.pathname === '/show/video'}>
		<span class="dock-label">Video</span>
	</a>
	<a href="/show/lighting" class="dock-item" class:dock-active={page.url.pathname === '/show/lighting'}>
		<span class="dock-label">Lighting</span>
	</a>
	<a href="/show/settings" class="dock-item" class:dock-active={page.url.pathname === '/show/settings'}>
		<span class="dock-label">Settings</span>
	</a>
</div>
