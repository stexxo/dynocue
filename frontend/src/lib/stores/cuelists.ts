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

export interface CueList {
    cueListNumber: number;
    label: string;
    listType: string;
}

function createCueListStore() {
    const { subscribe, set, update } = writable<CueList[]>([]);

    async function refresh() {
        try {
            const result = await Commands.EnumerateCueList({});
            if (result && result.cueLists) {
                set(result.cueLists);
            }
        } catch (err) {
            console.error('Failed to enumerate cue lists:', err);
        }
    }

    async function updateMetadata(cueListNumber: number, key: string, value: string) {
        try {
            await Commands.UpdateCueListMetadata({ cueListNumber: cueListNumber, key: key, value: value });
            // The store will be updated by the event handler
        } catch (err) {
            console.error(`Failed to update cue list ${cueListNumber} metadata:`, err);
        }
    }

    async function create(cueListNumber: number = 0) {
        try {
            await Commands.CreateCueList({ cueListNumber: cueListNumber });
            // The store will be updated by the event handler
        } catch (err) {
            console.error('Failed to create cue list:', err);
        }
    }

    async function remove(cueListNumber: number) {
        try {
            await Commands.DeleteCueList({ cueListNumber: cueListNumber });
            // The store will be updated by the event handler
        } catch (err) {
            console.error(`Failed to delete cue list ${cueListNumber}:`, err);
        }
    }

    async function move(originalCueListNumber: number, newCueListNumber: number) {
        try {
            await Commands.MoveCueList({ originalCueListNumber: originalCueListNumber, newCueListNumber: newCueListNumber });
            // The store will be updated by the event handlers (delete old, create new)
        } catch (err) {
            console.error(`Failed to move cue list ${originalCueListNumber} to ${newCueListNumber}:`, err);
        }
    }

    // Subscribe to backend events
    Events.On('event.cuelist.created', (event: any) => {
        const { cueListNumber, label, listType } = event.data;
        const result: CueList = {
            cueListNumber: cueListNumber,
            label: label,
            listType: listType
        };
        update(lists => {
            const newList = [...lists, result];
            return newList.sort((a, b) => a.cueListNumber - b.cueListNumber);
        });
    });

    Events.On('event.cuelist.updated', (event: any) => {
        const { cueListNumber, label, listType } = event.data;
        update(lists => lists.map(list => 
            list.cueListNumber === cueListNumber ? { ...list, label: label, listType: listType } : list
        ));
    });

    Events.On('event.cuelist.deleted', (event: any) => {
        const cueListNumber = event.data.cueListNumber;
        update(lists => lists.filter(list => list.cueListNumber !== cueListNumber));
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

export const cueLists = createCueListStore();
