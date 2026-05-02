// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

import {
	EnumerateCueLists,
	CreateCueList,
	UpdateCueListAttributesField,
	DeleteCueList,
	RenumberCueList
} from '../../../bindings/github.com/stexxo/dynocue/gui/services/cuelistsservice';
import { CueListAttributes } from '../../../bindings/github.com/stexxo/dynocue/components/cues/types';
import { Events } from '@wailsio/runtime';

/**
 * Store for managing cue lists.
 */
class CuelistsStore {
	#cuelists = $state<CueListAttributes[]>([]);

	constructor() {
		this.load();
		Events.On('event.cueing.cuelists.created', () => {
			this.load();
		});
		Events.On('event.cueing.cuelists.attributes.updated', () => {
			this.load();
		});
		Events.On('event.cueing.cuelists.deleted', () => {
			this.load();
		});
		Events.On('event.cueing.cuelists.renumber', () => {
			this.load();
		});
		Events.On('event.system.persistence.loaded', () => {
			this.load();
		});
	}

	get cuelists() {
		return this.#cuelists;
	}

	cueList(id: string): CueListAttributes | undefined {
		return this.#cuelists.find((list) => list.id === id);
	}

	async load() {
		const [lists, ok] = await EnumerateCueLists();
		if (ok) {
			this.#cuelists = lists;
		}
	}

	async create(num: number) {
		const ok = await CreateCueList(num, 'SEQUENTIAL');
		if (!ok) {
			console.error('Failed to create cue list', num);
		}
	}

	async setAttributesField(id: string, field: string, value: any) {
		const ok = await UpdateCueListAttributesField(id, field, value);
		if (!ok) {
			console.error('Failed to set cue list attributes field');
		}
	}

	async deleteCueList(id: string) {
		const ok = await DeleteCueList(id);
		if (!ok) {
			console.error('Failed to delete cue list', id);
		}
	}

	async renumberCuelist(id: string, newNum: number) {
		const ok = await RenumberCueList(id, newNum);
		if (!ok) {
			console.error('Failed to renumber cue list', id, newNum);
		}
	}
}

export const cuelistsStore = new CuelistsStore();
