import {
	EnumerateCueLists,
	CreateCueList,
	UpdateCueListMetadataField,
	DeleteCueList, RenumberCueList
} from "../../../bindings/github.com/stexxo/dynocue/gui/services/cuelistsservice";
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
		Events.On("event.system.persistence.loaded", () => {
			this.load();
		})
	}

	get cuelists() {
		return this.#cuelists;
	}

	cueList(id:string): CueListMetadata | undefined {
		return this.#cuelists.find(list => list.id === id);
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

	async setMetadataField(id: string, field:string, value:any) {
		const ok = await UpdateCueListMetadataField(id, field, value);
		if (!ok) {
			console.error("Failed to set cue list metadata field");
		}
	}

	async deleteCueList(id: string) {
		const ok = await DeleteCueList(id)
		if (!ok) {
			console.error("Failed to delete cue list", id);
		}
	}

	async renumberCuelist(id: string, newNum: number	) {
		const ok = await RenumberCueList(id, newNum);
		if (!ok) {
			console.error("Failed to renumber cue list", id, newNum);
		}
	}
}

export const cuelistsStore = new CuelistsStore();