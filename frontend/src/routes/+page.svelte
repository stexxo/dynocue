<script lang="ts">
	import {
		NewShow,
		OpenShow
	} from '../../bindings/gitlab.com/stexxo/dynocue/dynod/internal/subsystems/gui/api/lifecycleservice';

	let loading = $state(false);

	async function handleNewShow() {
		loading = true;
		await NewShow();
		loading = false;
	}

	async function handleLoadShow() {
		loading = true;
		await OpenShow();
		loading = false;
	}
</script>

<div class="hero min-h-screen">
	<div class="hero-content text-center">
		<div class="max-w-md">
			<h1 class="mb-10 text-6xl font-black">DynoCue</h1>
			{#if loading}
				<div class="flex flex-col items-center gap-4">
					<span class="loading loading-lg loading-spinner text-primary"></span>
				</div>
			{:else}
				<div class="flex justify-center gap-4">
					<button class="btn px-8 btn-lg btn-primary" onclick={handleNewShow}>New Show</button>
					<button class="btn px-8 btn-lg btn-primary" onclick={handleLoadShow}>Open Show</button>
				</div>
			{/if}
		</div>
	</div>
</div>
