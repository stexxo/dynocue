<!--
  This Source Code Form is subject to the terms of the Mozilla Public
  License, v. 2.0. If a copy of the MPL was not distributed with this
  file, You can obtain one at https://mozilla.org/MPL/2.0/.
-->

<script lang="ts">
	interface EditableTextInputProps {
		value: any;
		type?: string;
		onSave: (value: any) => void;
		onCancel?: () => void;
		label?: string;
		inputClass?: string;
		autofocus?: boolean;
		textarea?: boolean;
	}

	const props: EditableTextInputProps = $props();

	let editValue = $state(props.value);
	let hasChanged = $derived(editValue !== props.value);

	$effect(() => {
		editValue = props.value;
	});

	function handleSave() {
		props.onSave(editValue);
	}

	function handleCancel() {
		if (props.onCancel) {
			props.onCancel();
		} else {
			editValue = props.value;
		}
	}

	function focus(node: HTMLElement) {
		if (props.autofocus) {
			node.focus();
		}
	}
</script>

<div class="form-control relative w-full">
	{#if props.label}
		<label class="label pb-1">
			<span class="label-text {props.textarea ? 'font-semibold' : ''}">{props.label}</span>
		</label>
	{/if}

	<div class="flex flex-row gap-2">
		<div class="flex-1">
			{#if props.textarea}
				<textarea
					class="textarea-bordered textarea w-full {props.inputClass ?? ''}"
					bind:value={editValue}
					onblur={handleSave}
					onkeydown={(e) => {
						if (e.key === 'Escape') handleCancel();
					}}
					use:focus
				></textarea>
			{:else}
				<input
					type={props.type ?? 'text'}
					class="input-bordered input w-full {props.inputClass ?? ''}"
					bind:value={editValue}
					onblur={handleSave}
					onkeydown={(e) => {
						if (e.key === 'Enter') handleSave();
						if (e.key === 'Escape') handleCancel();
					}}
					use:focus
				/>
			{/if}
		</div>
		<div class="flex flex-row items-start gap-1 {props.label ? 'pt-1' : ''}">
			{#if hasChanged}
				<button class="btn btn-square btn-sm btn-success" onclick={handleSave} title="Save">
					<svg
						xmlns="http://www.w3.org/2000/svg"
						class="h-4 w-4"
						fill="none"
						viewBox="0 0 24 24"
						stroke="currentColor"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M5 13l4 4L19 7"
						/>
					</svg>
				</button>
				<button class="btn btn-square btn-ghost btn-sm" onclick={handleCancel} title="Cancel">
					<svg
						xmlns="http://www.w3.org/2000/svg"
						class="h-4 w-4"
						fill="none"
						viewBox="0 0 24 24"
						stroke="currentColor"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M6 18L18 6M6 6l12 12"
						/>
					</svg>
				</button>
			{/if}
		</div>
	</div>
</div>
