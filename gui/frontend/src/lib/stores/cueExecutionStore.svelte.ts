// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

import {
	EnumerateCueExecutions,
	GetCueExecution,
	GetSelectedCue
} from '../../../bindings/github.com/stexxo/dynocue/gui/services/executionservice';
import { CueExecution } from '../../../bindings/github.com/stexxo/dynocue/components/cues/types';
import { Events } from '@wailsio/runtime';
import { cuelistsStore } from './cuelistsStore.svelte';

/**
 * Store for managing cue execution states.
 */
class CueExecutionStore {
	#executions = $state<Map<string, CueExecution>>(new Map());
	#selectedCueIds = $state<Map<string, string>>(new Map());
	#loadedCueListIds = $state<Set<string>>(new Set());

	constructor() {
		// Automatically load executions for all cue lists
		$effect.root(() => {
			$effect(() => {
				cuelistsStore.cuelists.forEach((list) => {
					if (!this.#loadedCueListIds.has(list.cueListId)) {
						this.loadForCueList(list.cueListId);
					}
				});
			});
		});

		// Listen for execution events
		Events.On('event.cueing.execution.started', (ev: any) => {
			this.loadForCueList(ev.data.cueListId);
		});
		Events.On('event.cueing.execution.finished', (ev: any) => {
			this.loadForCueList(ev.data.cueListId);
		});
		Events.On('event.cueing.execution.unselected', (ev: any) => {
			this.loadForCueList(ev.data.cueListId);
		});
		Events.On('event.cueing.execution.deleted', (ev: any) => {
			this.loadForCueList(ev.data.cueListId);
		});

		// Cleanup on cue list deletion
		Events.On('event.cueing.cuelists.deleted', (ev: any) => {
			const cueListId = ev.data as string;
			this.#loadedCueListIds.delete(cueListId);
			this.#loadedCueListIds = new Set(this.#loadedCueListIds);
			this.#selectedCueIds.delete(cueListId);
			this.#selectedCueIds = new Map(this.#selectedCueIds);

			const toDelete: string[] = [];
			this.#executions.forEach((ex, cueId) => {
				if (ex.cueListId === cueListId) {
					toDelete.push(cueId);
				}
			});
			toDelete.forEach((id) => this.#executions.delete(id));
			this.#executions = new Map(this.#executions);
		});

		// Reset on persistence load
		Events.On('event.system.persistence.loaded', () => {
			this.#executions = new Map();
			this.#selectedCueIds = new Map();
			this.#loadedCueListIds = new Set();
		});
	}

	/**
	 * Returns the execution state for a specific cue.
	 * @param cueId The ID of the cue.
	 */
	getExecution(cueId: string): CueExecution | undefined {
		return this.#executions.get(cueId);
	}

	/**
	 * Returns the currently selected execution for a cue list.
	 * @param cueListId The ID of the cue list.
	 */
	getSelectedExecution(cueListId: string): CueExecution | undefined {
		const cueId = this.#selectedCueIds.get(cueListId);
		return cueId ? this.#executions.get(cueId) : undefined;
	}

	/**
	 * Returns all executions for a specific cue list.
	 * @param cueListId The ID of the cue list.
	 */
	getExecutionsForList(cueListId: string): CueExecution[] {
		return Array.from(this.#executions.values()).filter((ex) => ex.cueListId === cueListId);
	}

	/**
	 * Loads or reloads all executions for a specific cue list.
	 * @param cueListId The ID of the cue list.
	 */
	async loadForCueList(cueListId: string) {
		const [executions, ok] = await EnumerateCueExecutions(cueListId);
		if (ok) {
			this.#loadedCueListIds.add(cueListId);
			this.#loadedCueListIds = new Set(this.#loadedCueListIds);

			// Remove old executions for this list
			const toDelete: string[] = [];
			this.#executions.forEach((ex, cueId) => {
				if (ex.cueListId === cueListId) {
					toDelete.push(cueId);
				}
			});
			toDelete.forEach((id) => this.#executions.delete(id));

			// Add new ones
			executions.forEach((ex: CueExecution) => {
				this.#executions.set(ex.cueId, ex);
				if (ex.selected) {
					this.#selectedCueIds.set(cueListId, ex.cueId);
				}
			});

			// If none are selected, ensure we clear the selected entry
			if (!executions.some((ex: CueExecution) => ex.selected)) {
				this.#selectedCueIds.delete(cueListId);
			}

			// Trigger Svelte reactivity
			this.#executions = new Map(this.#executions);
			this.#selectedCueIds = new Map(this.#selectedCueIds);
		}
	}

	/**
	 * Refreshes a single cue execution.
	 * @param cueId The ID of the cue.
	 */
	async refreshExecution(cueId: string) {
		const [ex, ok] = await GetCueExecution(cueId);
		if (ok && ex) {
			this.#executions.set(cueId, ex);
			if (ex.selected) {
				this.#selectedCueIds.set(ex.cueListId, cueId);
			}
			this.#executions = new Map(this.#executions);
			this.#selectedCueIds = new Map(this.#selectedCueIds);
		}
	}

	/**
	 * Refreshes the selected cue for a cue list.
	 * @param cueListId The ID of the cue list.
	 */
	async refreshSelected(cueListId: string) {
		const [ex, ok] = await GetSelectedCue(cueListId);
		if (ok) {
			if (ex) {
				this.#executions.set(ex.cueId, ex);
				this.#selectedCueIds.set(cueListId, ex.cueId);
			} else {
				this.#selectedCueIds.delete(cueListId);
			}
			this.#executions = new Map(this.#executions);
			this.#selectedCueIds = new Map(this.#selectedCueIds);
		}
	}
}

export const cueExecutionStore = new CueExecutionStore();
