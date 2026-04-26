<!--
  This Source Code Form is subject to the terms of the Mozilla Public
  License, v. 2.0. If a copy of the MPL was not distributed with this
  file, You can obtain one at https://mozilla.org/MPL/2.0/.
-->


<script lang="ts">
	import EditableTextInput from '../inputs/EditableTextInput.svelte';

	interface EditableTableDataProps {
		value: any;
		inputType: string;
		tdClass?: string;
		onSaveEdit: (value: any) => void;
	}

	let editing = $state(false);

	const props: EditableTableDataProps = $props();

	function onSave(value: any) {
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
		<EditableTextInput
			value={props.value}
			type={props.inputType}
			{onSave}
			{onCancel}
			inputClass="input-sm"
			autofocus={true}
		/>
	{:else}
		<div class="min-h-10 cursor-pointer p-2 hover:border">
			{props.value}
		</div>
	{/if}
</td>
