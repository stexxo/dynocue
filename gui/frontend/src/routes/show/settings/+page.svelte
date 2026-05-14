<!--
  This Source Code Form is subject to the terms of the Mozilla Public
  License, v. 2.0. If a copy of the MPL was not distributed with this
  file, You can obtain one at https://mozilla.org/MPL/2.0/.
-->

<script>
	let showName = $state('New Show');

	function handleSubmit(event) {
		event.preventDefault();
		// TODO: Implement save logic
		console.log('Saving show name:', showName);
	}
</script>

<div class="mx-auto max-w-2xl p-8 pb-24">
	<div class="card border border-base-300 bg-base-100 shadow-xl">
		<div class="card-body gap-6">
			<h2 class="card-title text-2xl font-bold">Show Settings</h2>
			<div class="divider my-0"></div>

			<form onsubmit={handleSubmit} class="space-y-6">
				<div class="form-control w-full">
					<label class="label" for="show-name">
						<span class="label-text font-semibold">Show Name</span>
					</label>
					<input
						id="show-name"
						type="text"
						bind:value={showName}
						placeholder="Enter show name"
						class="input-bordered input w-full rounded-md"
					/>
					<div class="label">
						<span class="label-text-alt text-base-content/60"
							>This name will be displayed in the application header and dashboard.</span
						>
					</div>
				</div>

				<div class="mt-4 card-actions justify-end">
					<button type="submit" class="btn rounded-md px-8 btn-primary"> Save Changes </button>
				</div>
			</form>
		</div>
	</div>
</div>
