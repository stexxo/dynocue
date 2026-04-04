<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { pageTitle } from '$lib/stores/header';
	let { children } = $props();
	let loading = $state(false);
	import {CloseShow, OpenShow} from "../../../bindings/github.com/stexxo/dynocue/internal/gui/commands";
	import {Dialogs, Window} from "@wailsio/runtime";

	const pages = [
		{page: '/show/cues', label: 'Cue Lists'},
		{page: '/show/audio', label: 'Audio'},
		{page: '/show/video', label: 'Video'},
		{page: '/show/lighting', label: 'Lighting'},
		{page: "/show/settings", label: "Settings"}
	]

	$effect(() => {
		const currentPage = pages.find((p) => p.page === page.url.pathname);
		if (currentPage) {
			pageTitle.set(currentPage.label);
		} else if (page.url.pathname === "/show") {
			pageTitle.set("DynoCue");
		}
	});


	async function NewShowDialog() {
		const selection = await Dialogs.SaveFile({
			Title: "New Show",
			CanCreateDirectories: true,
		})
		if (selection === "") {
			return
		}

		const [filename, success] = await OpenShow(selection)
		if (success) {
			await Window.SetTitle(filename)
			await goto("/show")
		}
	}


	async function OpenShowDialog() {
		const selection = await Dialogs.OpenFile({
			Title: "Open Show",
			CanChooseDirectories: true,
			CanChooseFiles: false,
		})
		if (selection === "") {
			return
		}
		const [filename, success] = await OpenShow(selection)
		if (success) {
			await Window.SetTitle(filename)
			await goto("/show")
		}
	}
</script>

<div class="navbar min-h-0 bg-base-100 py-0.5 shadow-sm">
	<div class="navbar-start">
		<div class="dropdown">
			<div tabindex="0" role="button" class="btn btn-circle btn-ghost">
				<svg
					xmlns="http://www.w3.org/2000/svg"
					class="h-5 w-5"
					fill="none"
					viewBox="0 0 24 24"
					stroke="currentColor"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M4 6h16M4 12h16M4 18h7"
					/>
				</svg>
			</div>
			<ul
				tabindex="-1"
				class="dropdown-content menu z-[100] mt-3 w-52 menu-sm rounded-box bg-base-100 p-2 shadow"
			>
				<li>
					<button
						class="btn btn-ghost btn-sm justify-start w-full font-normal"
						onclick={async () => {
							loading = true;
							await NewShowDialog();
							loading = false;
						}}>New Show</button
					>
				</li>
				<li>
					<button
						class="btn btn-ghost btn-sm justify-start w-full font-normal"
						onclick={async () => {
							loading = true;
							await OpenShowDialog();
							loading = false;
						}}>Open Show</button
					>
				</li>
				<li>
					<button
						class="btn btn-ghost btn-sm justify-start w-full font-normal"
						onclick={async () => {
							loading = true;
							await CloseShow()
							await Window.SetTitle("DynoCue")
							await goto('/');
							loading = false;
						}}>Close Show</button
					>
				</li>
			</ul>
		</div>
	</div>
	<div class="navbar-center">
		<h3>{$pageTitle}</h3>
	</div>
	<div class="navbar-end"></div>
</div>

{@render children()}

{#snippet dashSnippet(p)}
	<button
			onclick={async () => {
			await goto(p.page);
		}}
			class="w-full hover:bg-base-300"
			class:dock-active={page.url.pathname === p.page}
	>
		<span>{p.label}</span>
	</button>
{/snippet}
<div class="dock dock-xs font-sans text-xs font-semibold">
	{#each pages as p}
		{@render dashSnippet(p)}
	{/each}
</div>
