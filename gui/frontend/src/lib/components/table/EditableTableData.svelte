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
		prefix?: import('svelte').Snippet;
	}

	let editing = $state(false);

	const props: EditableTableDataProps = $props();

	function onSave(value: any) {
		const parsedValue = props.inputType === 'number' ? Number(value) : value;
		props.onSaveEdit(parsedValue);
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
	<div class="flex items-center">
		{#if props.prefix}
			{@render props.prefix()}
		{/if}
		<div class="relative flex-1">
			{#if editing}
				<div class="absolute inset-0 z-20 flex items-center bg-base-100">
					<EditableTextInput
						value={props.value}
						inputType={props.inputType}
						{onSave}
						{onCancel}
						inputClass="input-sm"
						autofocus={true}
					/>
				</div>
			{/if}
			<div class="min-h-10 cursor-pointer p-2 hover:border {editing ? 'invisible' : ''}">
				{props.value}
			</div>
		</div>
	</div>
</td>
