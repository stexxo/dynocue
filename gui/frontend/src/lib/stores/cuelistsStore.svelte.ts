import {
	EnumerateCueLists,
	CreateCueList,
	SetCueListLabel,
	DeleteCueList, RenumberCueList
} from "../../../bindings/github.com/stexxo/dynocue/gui/cuelistsservice";
import { CueListMetadata } from "../../../bindings/github.com/stexxo/dynocue/components/cues/types";
import { Events } from "@wailsio/runtime";

/**
 * Store for managing cue lists.
 */
class CuelistsStore {
	#cuelists = $state<CueListMetadata[]>([]);

	constructor() {
		this.load();
		Events.On("event.cueing.cuelists.created", () => {
			this.load();
		});
		Events.On("event.cueing.cuelists.metadata.updated", () => {
			this.load();
		})
		Events.On("event.cueing.cuelists.deleted", () => {
			this.load();
		})
		Events.On("event.cueing.cuelists.renumber", () => {
			this.load();
		})
	}

	get cuelists() {
		return this.#cuelists;
	}

	cueList(number:number): CueListMetadata | undefined {
		return this.#cuelists.find(list => list.number === number);
	}

	async load() {
		const [lists, ok] = await EnumerateCueLists();
		if (ok) {
			this.#cuelists = lists;
		}
	}

	async create(num: number) {
		const ok = await CreateCueList(num, "SEQUENTIAL");
		if (!ok) {
			console.error("Failed to create cue list", num);
		}
	}

	async setLabel(num: number, label: string) {
		const ok = await SetCueListLabel(num, label);
		if (!ok) {
			console.error("Failed to set cue list label", num, label);
		}
	}

	async deleteCueList(num: number) {
		const ok = await DeleteCueList(num)
		if (!ok) {
			console.error("Failed to delete cue list", num);
		}
	}

	async renumberCuelist(originalNum: number, newNum: number	) {
		const ok = await RenumberCueList(originalNum, newNum);
		if (!ok) {
			console.error("Failed to renumber cue list", originalNum, newNum);
		}
	}
}

export const cuelistsStore = new CuelistsStore();