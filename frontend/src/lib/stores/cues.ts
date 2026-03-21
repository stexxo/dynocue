/**
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 *
 * SPDX-License-Identifier: MPL-2.0
 */

import { writable, get } from 'svelte/store';
import * as Commands from '../../../bindings/github.com/stexxo/dynocue/internal/gui/commands';
import { Cue } from '../../../bindings/github.com/stexxo/dynocue/api/cues';
import { Events } from '@wailsio/runtime';

function createCueStore() {
    const lists = new Map<number, {
        subscribe: (run: (value: Cue[]) => void) => () => void;
        set: (value: Cue[]) => void;
        update: (updater: (value: Cue[]) => Cue[]) => void;
    }>();

    function getStore(cueListNumber: number) {
        let store = lists.get(cueListNumber);
        if (!store) {
            store = writable<Cue[]>([]);
            lists.set(cueListNumber, store);
        }
        return store;
    }

    async function refresh(cueListNumber: number) {
        const store = getStore(cueListNumber);
        try {
            const result = await Commands.EnumerateCue({ cueListNumber });
            if (result && result.cues) {
                store.set(result.cues.map((c: any) => c.cue));
            } else {
                store.set([]);
            }
        } catch (err) {
            console.error(`Failed to enumerate cues for cue list ${cueListNumber}:`, err);
        }
    }

    async function updateMetadata(cueListNumber: number, cueNumber: number, key: string, value: string) {
        try {
            await Commands.UpdateCue({ cueListNumber, cueNumber, key, value });
        } catch (err) {
            console.error(`Failed to update cue ${cueNumber}:`, err);
        }
    }

    async function create(cueListNumber: number, cueNumber: number = 0) {
        try {
            await Commands.CreateCue({ cueListNumber, cueNumber });
        } catch (err) {
            console.error('Failed to create cue:', err);
        }
    }

    async function remove(cueListNumber: number, cueNumber: number) {
        try {
            await Commands.DeleteCue({ cueListNumber, cueNumber });
        } catch (err) {
            console.error(`Failed to delete cue ${cueNumber}:`, err);
        }
    }

    async function move(cueListNumber: number, originalCueNumber: number, newCueNumber: number) {
        try {
            await Commands.MoveCue({ cueListNumber, originalCueNumber, newCueNumber });
        } catch (err) {
            console.error(`Failed to move cue ${originalCueNumber} to ${newCueNumber}:`, err);
        }
    }

    // Subscribe to backend events
    Events.On('event.cue.created', (event: any) => {
        const { cueListNumber, cue } = event.data;
        const store = lists.get(cueListNumber);
        if (!store) return;

        store.update(cues => {
            const newCues = [...cues, cue];
            return newCues.sort((a, b) => a.cueNumber - b.cueNumber);
        });
    });

    Events.On('event.cue.updated', (event: any) => {
        const { cueListNumber, cue } = event.data;
        const store = lists.get(cueListNumber);
        if (!store) return;

        store.update(cues => cues.map(c => 
            c.cueNumber === cue.cueNumber ? cue : c
        ));
    });

    Events.On('event.cue.deleted', (event: any) => {
        const { cueListNumber, cueNumber } = event.data;
        const store = lists.get(cueListNumber);
        if (!store) return;

        store.update(cues => cues.filter(cue => cue.cueNumber !== cueNumber));
    });

    return {
        byListNumber: (cueListNumber: number) => {
            const store = getStore(cueListNumber);
            return {
                subscribe: store.subscribe,
                refresh: () => refresh(cueListNumber),
                updateMetadata: (cueNumber: number, key: string, value: string) => updateMetadata(cueListNumber, cueNumber, key, value),
                create: (cueNumber: number = 0) => create(cueListNumber, cueNumber),
                remove: (cueNumber: number) => remove(cueListNumber, cueNumber),
                move: (originalCueNumber: number, newCueNumber: number) => move(cueListNumber, originalCueNumber, newCueNumber)
            };
        }
    };
}

export const cues = createCueStore();
