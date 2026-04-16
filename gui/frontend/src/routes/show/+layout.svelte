<!--
  This Source Code Form is subject to the terms of the Mozilla Public
  License, v. 2.0. If a copy of the MPL was not distributed with this
  file, You can obtain one at https://mozilla.org/MPL/2.0/.
-->

<script lang="ts">
	import { page } from '$app/state';
	let { children } = $props();
</script>

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
