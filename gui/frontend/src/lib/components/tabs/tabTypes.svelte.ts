/**
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

export interface Tab {
	id: string;
	label: string;
	content?: any; // Can be text, or a reference to a Component
	props?: any;
	closable?: boolean;
}

export class TabManager {
	items = $state<Tab[]>([]);
	activeId = $state<string>('');

	constructor(initialTabs: Tab[], defaultActiveId: string) {
		this.items = initialTabs;
		this.activeId = defaultActiveId;
	}

	addTab(tab: Tab) {
		this.items.push(tab);
		this.activeId = tab.id; // Switch to new tab
	}

	closeTab(id: string) {
		if (this.activeId === id && this.items.length > 1) {
			const index = this.items.findIndex((t) => t.id === id);
			const nextTab = this.items[index + 1] || this.items[index - 1];
			this.activeId = nextTab.id;
		} else if (this.activeId === id && this.items.length == 1) {
			this.activeId = '';
		}

		this.items = this.items.filter((t) => t.id !== id);
	}

	getActive(): Tab | undefined {
		return this.items.find((t) => t.id === this.activeId);
	}

	select(id: string) {
		this.activeId = id;
	}
}

export interface TabProps {
	tabState: TabManager;
}
