/**
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 *
 * SPDX-License-Identifier: MPL-2.0
 */

import { writable } from 'svelte/store';
import * as Commands from '../../../bindings/gitlab.com/stexxo/dynocue/internal/gui/commands';
import { Cue } from '../../../bindings/gitlab.com/stexxo/dynocue/api/cues';
import { Events } from '@wailsio/runtime';

function createCueStore() {
    const { subscribe, set, update } = writable<Cue[]>([]);
    let currentCueListNumber: number | null = null;

    async function refresh(cueListNumber: number) {
        currentCueListNumber = cueListNumber;
        try {
            const result = await Commands.EnumerateCue({ cueListNumber });
            if (result && result.cues) {
                set(result.cues.map((c: any) => c.cue));
            } else {
                set([]);
            }
        } catch (err) {
            console.error(`Failed to enumerate cues for cue list ${cueListNumber}:`, err);
        }
    }

    async function updateMetadata(cueNumber: number, key: string, value: string) {
        if (currentCueListNumber === null) return;
        try {
            await Commands.UpdateCue({ cueListNumber: currentCueListNumber, cueNumber, key, value });
        } catch (err) {
            console.error(`Failed to update cue ${cueNumber}:`, err);
        }
    }

    async function create(cueNumber: number = 0) {
        if (currentCueListNumber === null) return;
        try {
            await Commands.CreateCue({ cueListNumber: currentCueListNumber, cueNumber });
        } catch (err) {
            console.error('Failed to create cue:', err);
        }
    }

    async function remove(cueNumber: number) {
        if (currentCueListNumber === null) return;
        try {
            await Commands.DeleteCue({ cueListNumber: currentCueListNumber, cueNumber });
        } catch (err) {
            console.error(`Failed to delete cue ${cueNumber}:`, err);
        }
    }

    async function move(originalCueNumber: number, newCueNumber: number) {
        if (currentCueListNumber === null) return;
        try {
            await Commands.MoveCue({ cueListNumber: currentCueListNumber, originalCueNumber, newCueNumber });
        } catch (err) {
            console.error(`Failed to move cue ${originalCueNumber} to ${newCueNumber}:`, err);
        }
    }

    // Subscribe to backend events
    Events.On('event.cue.created', (event: any) => {
        const { cueListNumber, cue } = event.data;
        if (cueListNumber !== currentCueListNumber) return;

        update(cues => {
            const newCues = [...cues, cue];
            return newCues.sort((a, b) => a.cueNumber - b.cueNumber);
        });
    });

    Events.On('event.cue.updated', (event: any) => {
        const { cueListNumber, cue } = event.data;
        if (cueListNumber !== currentCueListNumber) return;

        update(cues => cues.map(c => 
            c.cueNumber === cue.cueNumber ? cue : c
        ));
    });

    Events.On('event.cue.deleted', (event: any) => {
        const { cueListNumber, cueNumber } = event.data;
        if (cueListNumber !== currentCueListNumber) return;

        update(cues => cues.filter(cue => cue.cueNumber !== cueNumber));
    });

    return {
        subscribe,
        refresh,
        updateMetadata,
        create,
        remove,
        move
    };
}

export const cues = createCueStore();
