// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

import {
	EnumerateActionTemplates,
	GetActionTemplate
} from '../../../bindings/github.com/stexxo/dynocue/gui/services/actiontemplatesservice';
import { ActionTemplate } from '../../../bindings/github.com/stexxo/dynocue/components/cues/types';
import { Events } from '@wailsio/runtime';

/**
 * Store for managing action templates.
 */
class ActionTemplatesStore {
	#templates = $state<ActionTemplate[]>([]);

	constructor() {
		$effect.root(() => {
			this.load();
		});
		Events.On('event.cueing.actions.templates.registered', () => {
			this.load();
		});
		Events.On('event.system.persistence.loaded', () => {
			this.load();
		});
	}

	get templates() {
		return this.#templates;
	}

	async getTemplate(id: string): Promise<ActionTemplate | undefined> {
		const template = this.#templates.find((t) => t.id === id);
		if (template) {
			return template;
		}

		const [remoteTemplate, ok] = await GetActionTemplate(id);
		if (ok && remoteTemplate) {
			return remoteTemplate;
		}

		return undefined;
	}

	async load() {
		const [templates, ok] = await EnumerateActionTemplates();
		if (ok) {
			this.#templates = templates;
		}
	}
}

export const actionTemplatesStore = new ActionTemplatesStore();
