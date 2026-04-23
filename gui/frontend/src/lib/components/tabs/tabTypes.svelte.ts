/**
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

export interface Tab {
	id: string;
	content?: any;
	props?: any;
	closable?: boolean;
	label?: string | (() => string);
}

export interface TabContentProps {
	tabState: TabState;
	[key: string]: any;
}

export class TabState{
	#tab: Tab;
	constructor(tab: Tab) {
		this.#tab = tab;
	}
	get label() {
		if (typeof this.#tab.label === 'function') {
			return this.#tab.label();
		}
		return this.#tab.label ?? this.#tab.id;
	}
	setLabel(label: string | (() => string)){
		this.#tab.label = label;
	}
}

export class TabManager {
	items = $state<Tab[]>([]);
	activeId = $state<string>('');

	constructor(initialTabs: Tab[], defaultActiveId: string) {
		this.items = initialTabs;
		this.activeId = defaultActiveId;
	}

	addTab(tab: Tab) {
		this.items = [...this.items, tab];
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
	tabManager: TabManager;
}
