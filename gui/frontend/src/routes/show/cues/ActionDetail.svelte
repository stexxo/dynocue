<!--
  This Source Code Form is subject to the terms of the Mozilla Public
  License, v. 2.0. If a copy of the MPL was not distributed with this
  file, You can obtain one at https://mozilla.org/MPL/2.0/.
-->

<script lang="ts">
	import { actionsStore } from '$lib/stores/actionsStore.svelte';
	import { actionTemplatesStore } from '$lib/stores/actiontemplatesStore.svelte';
	import { cueExecutionStore } from '$lib/stores/cueExecutionStore.svelte';
	import EditableTextInput from '$lib/components/inputs/EditableTextInput.svelte';
	import EditableTimeInput from '$lib/components/inputs/EditableTimeInput.svelte';
	import EditableTableData from '$lib/components/table/EditableTableData.svelte';
	import EditableTimeData from '$lib/components/table/EditableTimeData.svelte';
	import { formatTime, parseTimeToMs } from '$lib/utils/time';

	interface ActionDetailProps {
		cueListId: string | undefined;
		cueId: string | undefined;
		actionId: string;
	}

	let { cueListId, cueId, actionId }: ActionDetailProps = $props();

	let action = $derived.by(() => {
		if (!cueId) return null;
		const actions = actionsStore.actions.get(cueId);
		return actions?.find((a) => a.actionId === actionId) ?? null;
	});

	let template = $derived.by(() => {
		if (!action) return null;
		return actionTemplatesStore.templates.find((t) => t.templateId === action.templateId) ?? null;
	});
	let isExpanded = $state(false);

	let now = $state(Date.now());

	$effect(() => {
		const interval = setInterval(() => {
			now = Date.now();
		}, 100);
		return () => clearInterval(interval);
	});

	function getElapsed(execution: any) {
		if (!execution?.actionStarted) return 0;
		const start = parseTimeToMs(execution.actionStarted);
		if (start <= 0) return 0;
		return Math.max(0, now - start) * 1000000;
	}
</script>

{#if action && cueListId && cueId}
	{@const execution = cueExecutionStore.getActionExecution(action.actionId)}
	<tr class={execution ? 'bg-emerald-900 hover:bg-emerald-950' : 'hover:bg-base-200'}>
		<td>
			<button
				class="btn h-full rounded-none px-1 btn-ghost btn-xs"
				onclick={(e) => {
					e.stopPropagation();
					isExpanded = !isExpanded;
				}}
			>
				{#if isExpanded}
					<svg
						xmlns="http://www.w3.org/2000/svg"
						viewBox="0 0 20 20"
						fill="currentColor"
						class="h-4 w-4"
					>
						<path
							fill-rule="evenodd"
							d="M5.22 14.78a.75.75 0 001.06 0L10 11.06l3.72 3.72a.75.75 0 101.06-1.06l-4.25-4.25a.75.75 0 00-1.06 0l-4.25 4.25a.75.75 0 000 1.06z"
							clip-rule="evenodd"
						/>
					</svg>
				{:else}
					<svg
						xmlns="http://www.w3.org/2000/svg"
						viewBox="0 0 20 20"
						fill="currentColor"
						class="h-4 w-4"
					>
						<path
							fill-rule="evenodd"
							d="M5.22 5.22a.75.75 0 011.06 0L10 8.94l3.72-3.72a.75.75 0 111.06 1.06l-4.25 4.25a.75.75 0 01-1.06 0L5.22 6.28a.75.75 0 010-1.06z"
							clip-rule="evenodd"
						/>
					</svg>
				{/if}
			</button>
		</td>
		<EditableTableData
			value={action.number}
			inputType="number"
			onSaveEdit={(v) => actionsStore.update(action.actionId, 'number', parseInt(v))}
			tdClass="w-16 border-none"
		/>
		<EditableTableData
			value={action.label}
			inputType="text"
			onSaveEdit={(v) => actionsStore.update(action.actionId, 'label', v)}
			tdClass="w-64 border-none"
		/>
		<td>
			<span class="text-sm opacity-70">
				{template?.templateName ?? 'Unknown Template'}
			</span>
		</td>
		<EditableTimeData
			value={action.delay}
			onSaveEdit={(v) => actionsStore.update(action.actionId, 'delay', v)}
			tdClass="border-none"
			timerActive={execution?.delayActive}
			timerStart={execution?.delayStarted}
		/>
		<td class="w-32">
			<span class="text-sm opacity-70">
				{execution ? formatTime(getElapsed(execution)) : ''}
			</span>
		</td>
		<td>
			<div class="flex gap-2">
				<button
					class="btn text-error btn-ghost btn-xs"
					onclick={() => actionsStore.deleteAction(action.actionId)}
				>
					Delete
				</button>
			</div>
		</td>
	</tr>

	{#if isExpanded}
		<tr>
			<td colspan="7" class="bg-base-200 p-4">
				<div class="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
					{#each action.fields as field}
						{#if field.dataType === 'string'}
							<EditableTextInput
								label={field.fieldLabel || field.fieldName}
								value={field.value}
								onSave={(v) => actionsStore.updateField(action.actionId, field.fieldName, v)}
							/>
						{:else if field.dataType === 'float' || field.dataType === 'int'}
							<EditableTextInput
								label={field.fieldLabel || field.fieldName}
								value={field.value}
								inputType="number"
								onSave={(v) => actionsStore.updateField(action.actionId, field.fieldName, v)}
							/>
						{:else if field.dataType === 'bool'}
							<div class="form-control flex flex-col justify-center">
								<label class="label flex cursor-pointer items-center justify-between">
									<span class="label-text">{field.fieldLabel || field.fieldName}</span>
									<input
										type="checkbox"
										class="checkbox"
										checked={field.value}
										onchange={(e) =>
											actionsStore.updateField(
												action.actionId,
												field.fieldName,
												e.currentTarget.checked
											)}
									/>
								</label>
							</div>
						{:else if field.dataType === 'time'}
							<EditableTimeInput
								label={field.fieldLabel || field.fieldName}
								value={field.value}
								onSave={(v) => actionsStore.updateField(action.actionId, field.fieldName, v)}
							/>
						{/if}
					{/each}
				</div>
			</td>
		</tr>
	{/if}
{/if}
