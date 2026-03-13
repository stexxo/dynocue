<script lang="ts">
	import { Dialogs } from "@wailsio/runtime";
	import {Window} from "@wailsio/runtime";
	import {OpenShow} from "../../bindings/gitlab.com/stexxo/dynocue/internal/gui/commands";
	import { goto } from '$app/navigation';

	let menuPage = "main"

	async function NewShowDialog() {
		menuPage = "loading"
		const selection = await Dialogs.SaveFile({
			Title: "New Show",
			CanCreateDirectories: true,
		})
		const [filename, success] = await OpenShow(selection)
		if (success) {
			await Window.SetTitle(filename)
			await goto("/show")
		} else {
			menuPage = "local"
		}
	}


	async function OpenShowDialog() {
		menuPage = "loading"
		const selection = await Dialogs.OpenFile({
			Title: "Open Show",
			CanChooseDirectories: true,
			CanChooseFiles: false,
		})
		const [filename, success] = await OpenShow(selection)
		if (success) {
			await Window.SetTitle(filename)
			await goto("/show")
		} else {
			menuPage = "local"
		}
	}

</script>

<div class="hero min-h-screen">
	<div class="hero-content text-center">
		<div class="max-w-md">
			<h1 class="mb-10 text-6xl font-black">DynoCue</h1>
			{#if menuPage === "main"}
				<div class="flex justify-center gap-4">
					<button class="btn px-8 btn-lg btn-primary" onclick={()=>{menuPage = "local"}}>Local</button>
					<button class="btn px-8 btn-lg btn-accent">Remote</button>
				</div>
			{:else if menuPage === "local"}
				<div class="flex flex-col gap-4">
					<div class="flex justify-center gap-4">
						<button class="btn px-8 btn-lg btn-primary" onclick={NewShowDialog}>New Show</button>
						<button class="btn px-8 btn-lg btn-accent" onclick="{OpenShowDialog}">Open Show</button>
					</div>
					<div class="flex justify-center gap-4">
						<button class="btn btn-neutral btn-sm" onclick={()=>{menuPage = "main"}}>Back</button>
					</div>
				</div>
			{:else if menuPage === "loading"}
				<span class="loading loading-spinner loading-xs"></span>
			{/if}
		</div>
	</div>
</div>
