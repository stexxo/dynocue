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
    number: number;
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

    async function updateMetadata(number: number, key: string, value: string) {
        try {
            await Commands.UpdateCueListMetadata({ number: number, key: key, value: value });
            // The store will be updated by the event handler
        } catch (err) {
            console.error(`Failed to update cue list ${number} metadata:`, err);
        }
    }

    async function create(number: number = 0) {
        try {
            await Commands.CreateCueList({ number: number });
            // The store will be updated by the event handler
        } catch (err) {
            console.error('Failed to create cue list:', err);
        }
    }

    async function remove(number: number) {
        try {
            await Commands.DeleteCueList({ number: number });
            // The store will be updated by the event handler
        } catch (err) {
            console.error(`Failed to delete cue list ${number}:`, err);
        }
    }

    async function move(originalNumber: number, newNumber: number) {
        try {
            await Commands.MoveCueList({ originalNumber: originalNumber, newNumber: newNumber });
            // The store will be updated by the event handlers (delete old, create new)
        } catch (err) {
            console.error(`Failed to move cue list ${originalNumber} to ${newNumber}:`, err);
        }
    }

    // Subscribe to backend events
    Events.On('event.cuelist.created', (event: any) => {
        const { number, label, listType } = event.data;
        const result: CueList = {
            number: number,
            label: label,
            listType: listType
        };
        update(lists => {
            const newList = [...lists, result];
            return newList.sort((a, b) => a.number - b.number);
        });
    });

    Events.On('event.cuelist.updated', (event: any) => {
        const { number, label, listType } = event.data;
        update(lists => lists.map(list => 
            list.number === number ? { ...list, label: label, listType: listType } : list
        ));
    });

    Events.On('event.cuelist.deleted', (event: any) => {
        const number = event.data.number;
        update(lists => lists.filter(list => list.number !== number));
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
