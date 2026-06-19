<!--
  This Source Code Form is subject to the terms of the Mozilla Public
  License, v. 2.0. If a copy of the MPL was not distributed with this
  file, You can obtain one at https://mozilla.org/MPL/2.0/.
-->

<script lang="ts">
        import { GetAudioOutputDevices } from '../../../../bindings/github.com/stexxo/dynocue/gui/services/audioservice';
	import type { AudioDevice } from '../../../../bindings/github.com/stexxo/dynocue/components/audio/types';

	let devices = $state<AudioDevice[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);

	async function loadDevices() {
		loading = true;
		error = null;
		try {
			const [result, ok] = await GetAudioOutputDevices();
			if (!ok) {
				error = 'Failed to load audio devices.';
				devices = [];
				return;
			}
			devices = (result ?? []).filter((d) => d && (d.name || d.friendlyName || d.description));
		} catch (e) {
			error = e instanceof Error ? e.message : String(e);
			devices = [];
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		loadDevices();
	});
</script>

<div class="flex h-full w-full flex-col items-center overflow-hidden p-4">
	<div class="flex h-full w-full max-w-7xl flex-col">
		<div class="mb-5 flex w-full flex-none flex-row items-center justify-between">
			<h1 class="text-2xl font-bold">Audio Devices</h1>
			<button class="btn btn-primary" onclick={loadDevices} disabled={loading}>
				{loading ? 'Refreshing…' : 'Refresh'}
			</button>
		</div>

		{#if error}
			<div class="alert alert-error mb-4">
				<span>{error}</span>
			</div>
		{/if}

		<div class="flex-1 overflow-auto">
			<table class="table-pin-rows table">
				<thead class="sticky top-0 z-10 bg-base-100">
					<tr class="bg-base-100">
						<th class="w-1/4">Name</th>
						<th class="w-1/4">Friendly Name</th>
						<th>Description</th>
					</tr>
				</thead>
				<tbody>
					{#if loading && devices.length === 0}
						<tr>
							<td colspan="3" class="text-center italic text-gray-500">Loading…</td>
						</tr>
					{:else}
						{#each devices as device}
							<tr class="hover:bg-base-200">
								<td class="truncate font-mono">{device.name}</td>
								<td class="truncate">{device.friendlyName}</td>
								<td class="truncate">{device.description}</td>
							</tr>
						{:else}
							<tr>
								<td colspan="3" class="text-center italic text-gray-500">
									No audio devices found.
								</td>
							</tr>
						{/each}
					{/if}
				</tbody>
			</table>
		</div>
	</div>
</div>
