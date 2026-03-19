/**
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 *
 * SPDX-License-Identifier: MPL-2.0
 */

import { writable } from 'svelte/store';
import * as Commands from '../../../bindings/gitlab.com/stexxo/dynocue/internal/gui/commands';
import { Events } from '@wailsio/runtime';

export interface Cue {
    number: number;
    label: string;
}

function createCueStore() {
    const { subscribe, set, update } = writable<Cue[]>([]);
    let currentCueListNumber: number | null = null;

    async function refresh(cueListNumber: number) {
        currentCueListNumber = cueListNumber;
        try {
            const result = await Commands.EnumerateCue({ cueListNumber });
            if (result && result.cues) {
                set(result.cues);
            } else {
                set([]);
            }
        } catch (err) {
            console.error(`Failed to enumerate cues for cue list ${cueListNumber}:`, err);
        }
    }

    async function updateMetadata(number: number, key: string, value: string) {
        if (currentCueListNumber === null) return;
        try {
            await Commands.UpdateCueMetadata({ cueListNumber: currentCueListNumber, number, key, value });
        } catch (err) {
            console.error(`Failed to update cue ${number} metadata:`, err);
        }
    }

    async function create(number: number = 0) {
        if (currentCueListNumber === null) return;
        try {
            await Commands.CreateCue({ cueListNumber: currentCueListNumber, number });
        } catch (err) {
            console.error('Failed to create cue:', err);
        }
    }

    async function remove(number: number) {
        if (currentCueListNumber === null) return;
        try {
            await Commands.DeleteCue({ cueListNumber: currentCueListNumber, number });
        } catch (err) {
            console.error(`Failed to delete cue ${number}:`, err);
        }
    }

    async function move(originalNumber: number, newNumber: number) {
        if (currentCueListNumber === null) return;
        try {
            await Commands.MoveCue({ cueListNumber: currentCueListNumber, originalNumber, newNumber });
        } catch (err) {
            console.error(`Failed to move cue ${originalNumber} to ${newNumber}:`, err);
        }
    }

    // Subscribe to backend events
    Events.On('event.cue.created', (event: any) => {
        const { cueListNumber, number, label } = event.data;
        if (cueListNumber !== currentCueListNumber) return;

        update(cues => {
            const newCues = [...cues, { number, label }];
            return newCues.sort((a, b) => a.number - b.number);
        });
    });

    Events.On('event.cue.updated', (event: any) => {
        const { cueListNumber, number, label } = event.data;
        if (cueListNumber !== currentCueListNumber) return;

        update(cues => cues.map(cue => 
            cue.number === number ? { ...cue, label } : cue
        ));
    });

    Events.On('event.cue.deleted', (event: any) => {
        const { cueListNumber, number } = event.data;
        if (cueListNumber !== currentCueListNumber) return;

        update(cues => cues.filter(cue => cue.number !== number));
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
