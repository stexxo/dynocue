<!--
  This Source Code Form is subject to the terms of the Mozilla Public
  License, v. 2.0. If a copy of the MPL was not distributed with this
  file, You can obtain one at https://mozilla.org/MPL/2.0/.
-->

<script>
	import {NewShow} from "../../bindings/github.com/stexxo/dynocue/gui/selector.js"
	import {withLoading, createLoadingState} from "$lib/loading.svelte"

	const loadingState = createLoadingState();
	const handleNewShow = withLoading(loadingState, NewShow);
</script>

<div class="hero min-h-screen bg-base-200">
	{#if loadingState.isLoading}
		<div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
			<span class="loading loading-spinner loading-lg text-primary"></span>
		</div>
	{/if}
	<div class="hero-content text-center">
		<div class="w-md">
			<h1 class="text-5xl font-bold">DynoCue</h1>
			<div class="flex flex-col gap-4 mt-8">
				<button 
					class="btn btn-primary rounded-md" 
					onclick={handleNewShow}
					disabled={loadingState.isLoading}
				>
					New Local Show
				</button>
				<button class="btn btn-secondary rounded-md" disabled={loadingState.isLoading}>Open Existing Local Show</button>
				<button class="btn btn-accent rounded-md" disabled={loadingState.isLoading}>Connect to Remote Show</button>
				<button class="btn btn-outline rounded-md" disabled={loadingState.isLoading}>Settings</button>
			</div>
		</div>
	</div>
</div>