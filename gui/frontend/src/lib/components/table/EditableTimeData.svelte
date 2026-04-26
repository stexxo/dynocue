<!--
  This Source Code Form is subject to the terms of the Mozilla Public
  License, v. 2.0. If a copy of the MPL was not distributed with this
  file, You can obtain one at https://mozilla.org/MPL/2.0/.
-->

<script lang="ts">
	import EditableTimeInput from '../inputs/EditableTimeInput.svelte';
	import { formatTime } from '$lib/utils/time';

	interface EditableTimeDataProps {
		value: number; // nanoseconds
		tdClass?: string;
		onSaveEdit: (value: number) => void;
	}

	let editing = $state(false);
	const props: EditableTimeDataProps = $props();

	function onSave(value: number) {
		props.onSaveEdit(value);
		editing = false;
	}

	function onCancel() {
		editing = false;
	}
</script>

<td
	class="relative {props.tdClass}"
	onclick={() => {
		editing = true;
	}}
>
	{#if editing}
		<div class="absolute inset-0 z-20 flex items-center bg-base-100">
			<EditableTimeInput
				value={props.value}
				{onSave}
				{onCancel}
				inputClass="input-sm"
				autofocus={true}
			/>
		</div>
	{/if}
	<div class="min-h-10 cursor-pointer p-2 hover:border {editing ? 'invisible' : ''}">
		{formatTime(props.value)}
	</div>
</td>
