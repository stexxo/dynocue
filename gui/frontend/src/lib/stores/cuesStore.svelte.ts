import {
	EnumerateCues,
	CreateCue,
	UpdateCueMetadata,
	DeleteCue,
	RenumberCue
} from '../../../bindings/github.com/stexxo/dynocue/gui/services/cuesservice';
import { CueMetadata } from '../../../bindings/github.com/stexxo/dynocue/components/cues/types';
import { Events } from '@wailsio/runtime';
import { cuelistsStore } from './cuelistsStore.svelte';

/**
 * Store for managing cues within cue lists.
 */
class CuesStore {
	#cues = $state<Map<string, CueMetadata[]>>(new Map());

	constructor() {
		$effect.root(() => {
			$effect(() => {
				cuelistsStore.cuelists.forEach((list) => {
					if (!this.#cues.has(list.id)) {
						this.load(list.id);
					}
				});
			});
		});

		Events.On('event.cueing.cuelists.deleted', (ev: any) => {
			// Wails emits just the ID as per cueing.go: cl.OnCueListDeleted(func(s string, id *string) { c.app.Event.Emit(s, *id) })
			const cueListId = ev.data as string;
			if (this.#cues.delete(cueListId)) {
				this.#cues = new Map(this.#cues);
			}
		});

		Events.On('event.cueing.cue.created', (ev: any) => {
			const event = ev.data as CueMetadata;
			this.load(event.cueListId);
		});
		Events.On('event.cueing.cue.metadata.updated', (ev: any) => {
			const event = ev.data as CueMetadata;
			this.load(event.cueListId);
		});
		Events.On('event.cueing.cue.deleted', (ev: any) => {
			const event = ev.data as { cueListId: string };
			this.load(event.cueListId);
		});
		Events.On('event.cueing.cue.renumber', (ev: any) => {
			const event = ev.data as { cueListId: string };
			this.load(event.cueListId);
		});
		Events.On('event.system.persistence.loaded', () => {
			this.#cues = new Map();
		});
	}

	get cues(): Map<string, CueMetadata[]> {
		return this.#cues;
	}

	async load(cueListId: string) {
		const [cues, ok] = await EnumerateCues(cueListId);
		if (ok) {
			this.#cues.set(cueListId, cues);
			// Re-assign to trigger Svelte reactivity for the Map
			this.#cues = new Map(this.#cues);
		}
	}

	async create(cueListId: string, num: number) {
		const ok = await CreateCue(cueListId, num);
		if (!ok) {
			console.error('Failed to create cue', cueListId, num);
		}
	}

	async updateCueMetadata(cueListId: string, cueId: string, field: string, value: any) {
		const ok = await UpdateCueMetadata(cueListId, cueId, field, value);
		if (!ok) {
			console.error('Failed to set cue metadata');
		}
	}

	async deleteCue(cueListId: string, cueId: string) {
		const ok = await DeleteCue(cueListId, cueId);
		if (!ok) {
			console.error('Failed to delete cue', cueListId, cueId);
		}
	}

	async renumberCue(cueListId: string, cueId: string, newNum: number) {
		const ok = await RenumberCue(cueListId, cueId, newNum);
		if (!ok) {
			console.error('Failed to renumber cue', cueListId, cueId, newNum);
		}
	}
}

export const cuesStore = new CuesStore();
