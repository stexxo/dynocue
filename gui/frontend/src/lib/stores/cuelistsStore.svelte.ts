import { EnumerateCueLists, CreateCueList, SetCueListLabel } from "../../../bindings/github.com/stexxo/dynocue/gui/cuelistsservice";
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
		// We don't need to manually refresh here as we're listening to the event
	}

	async setLabel(num: number, label: string) {
		const ok = await SetCueListLabel(num, label);
		if (!ok) {
			console.error("Failed to set cue list label", num, label);
		}
	}
}

export const cuelistsStore = new CuelistsStore();