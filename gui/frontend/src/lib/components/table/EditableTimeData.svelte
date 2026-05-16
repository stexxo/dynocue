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
		timerActive?: boolean;
		timerStart?: number | string | Date;
	}

	let editing = $state(false);
	const props: EditableTimeDataProps = $props();

	let now = $state(Date.now());
	const startMs = $derived.by(() => {
		if (!props.timerStart) return 0;
		if (typeof props.timerStart === 'number') {
			if (!Number.isFinite(props.timerStart)) return 0;
			if (props.timerStart > 1e15) return props.timerStart / 1000000;
			return props.timerStart;
		}
		const d = new Date(props.timerStart);
		const ms = d.getTime();
		return isNaN(ms) ? 0 : ms;
	});

	$effect(() => {
		if (props.timerActive && startMs > 0) {
			const interval = setInterval(() => {
				now = Date.now();
			}, 33);
			return () => clearInterval(interval);
		}
	});

	const remaining = $derived.by(() => {
		const val = Number(props.value);
		if (!Number.isFinite(val)) return 0;
		if (!props.timerActive || startMs <= 0) return val;
		const elapsedMs = Math.max(0, now - startMs);
		return Math.max(0, val - elapsedMs * 1000000);
	});

	const progress = $derived.by(() => {
		const val = Number(props.value);
		if (!Number.isFinite(val) || val <= 0 || !props.timerActive || startMs <= 0) return 0;
		const elapsedMs = Math.max(0, now - startMs);
		const durationMs = val / 1000000;
		return Math.min(100, (elapsedMs / durationMs) * 100);
	});

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
		{formatTime(remaining)}
	</div>
	{#if props.timerActive && startMs > 0 && !editing}
		<div class="absolute bottom-0 left-0 h-1 bg-primary" style:width="{progress}%"></div>
	{/if}
</td>
