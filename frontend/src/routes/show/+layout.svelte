<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	let { children } = $props();
	let loading = $state(false);

	const pages = [
		{page: '/show/cues', label: 'Cues'},
		{page: '/show/audio', label: 'Audio'},
		{page: '/show/video', label: 'Video'},
		{page: '/show/lighting', label: 'Lighting'},
		{page: "/show/settings", label: "Settings"}
	]
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
				class="dropdown-content menu z-1 mt-3 w-52 menu-sm rounded-box bg-base-100 p-2 shadow"
			>
				<li>
					<a
						onclick={async () => {
							loading = true;
							await goto('/show/cues');
							loading = false;
						}}>New Show</a
					>
				</li>
				<li>
					<a
						onclick={async () => {
							loading = true;
							await goto('/show/cues');
							loading = false;
						}}>Open Show</a
					>
				</li>
				<li>
					<a
						onclick={async () => {
							loading = true;
							loading = false;
						}}>Close Show</a
					>
				</li>
			</ul>
		</div>
	</div>
	<div class="navbar-center">
		<h3>DynoCue</h3>
	</div>
	<div class="navbar-end"></div>
</div>

{@render children()}

