// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

import {
	EnumerateActions,
	CreateAction,
	UpdateAction,
	UpdateActionField,
	DeleteAction
} from '../../../bindings/github.com/stexxo/dynocue/gui/services/actionsservice';
import { Action } from '../../../bindings/github.com/stexxo/dynocue/components/cues/types/models';
import { Events } from '@wailsio/runtime';
import { cuesStore } from './cuesStore.svelte';

/**
 * Store for managing actions within cues.
 */
class ActionsStore {
	#actions = $state<Map<string, Action[]>>(new Map());

	constructor() {
		$effect.root(() => {
			$effect(() => {
				cuesStore.cues.forEach((cues) => {
					cues.forEach((cue) => {
						if (!this.#actions.has(cue.cueId)) {
							this.load(cue.cueListId, cue.cueId);
						}
					});
				});
			});
		});

		Events.On('event.cueing.cue.deleted', (ev: any) => {
			const event = ev.data as { cueId: string };
			if (this.#actions.delete(event.cueId)) {
				this.#actions = new Map(this.#actions);
			}
		});

		Events.On('event.cueing.actions.created', (ev: any) => {
			const event = ev.data as { cueListId: string; cueId: string };
			this.load(event.cueListId, event.cueId);
		});

		Events.On('event.cueing.actions.updated', (ev: any) => {
			const event = ev.data as { cueListId: string; cueId: string };
			this.load(event.cueListId, event.cueId);
		});

		Events.On('event.cueing.actions.deleted', (ev: any) => {
			const event = ev.data as { cueListId: string; cueId: string };
			this.load(event.cueListId, event.cueId);
		});

		Events.On('event.system.persistence.loaded', () => {
			this.#actions = new Map();
		});
	}

	get actions(): Map<string, Action[]> {
		return this.#actions;
	}

	async load(cueListId: string, cueId: string) {
		const [actions, ok] = await EnumerateActions(cueListId, cueId);
		if (ok) {
			this.#actions.set(cueId, actions);
			// Re-assign to trigger Svelte reactivity for the Map
			this.#actions = new Map(this.#actions);
		}
	}

	async create(cueListId: string, cueId: string, templateId: string) {
		const [action, ok] = await CreateAction(cueListId, cueId, templateId);
		if (!ok) {
			console.error('Failed to create action', cueListId, cueId, templateId);
		}
		return action;
	}

	async update(cueListId: string, cueId: string, actionId: string, field: string, value: any) {
		const [action, ok] = await UpdateAction(cueListId, cueId, actionId, field, value);
		if (!ok) {
			console.error('Failed to update action', actionId, field, value);
		}
		return action;
	}

	async updateField(
		cueListId: string,
		cueId: string,
		actionId: string,
		fieldName: string,
		value: any
	) {
		const [action, ok] = await UpdateActionField(cueListId, cueId, actionId, fieldName, value);
		if (!ok) {
			console.error('Failed to update action field', actionId, fieldName, value);
		}
		return action;
	}

	async deleteAction(cueListId: string, cueId: string, actionId: string) {
		const ok = await DeleteAction(cueListId, cueId, actionId);
		if (!ok) {
			console.error('Failed to delete action', actionId);
		}
		return ok;
	}
}

export const actionsStore = new ActionsStore();
