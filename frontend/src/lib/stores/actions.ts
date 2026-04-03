/**
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 *
 * SPDX-License-Identifier: MPL-2.0
 */

import { writable } from 'svelte/store';
import * as Commands from '../../../bindings/github.com/stexxo/dynocue/internal/gui/commands';
import { CueAction as Action } from '../../../bindings/github.com/stexxo/dynocue/api/cues/models';
import { Events } from '@wailsio/runtime';

function createActionStore() {
    const cueActions = new Map<string, {
        subscribe: (run: (value: Action[]) => void) => () => void;
        set: (value: Action[]) => void;
        update: (updater: (value: Action[]) => Action[]) => void;
    }>();

    function getKey(cueListNumber: number, cueNumber: number) {
        return `${cueListNumber}:${cueNumber}`;
    }

    function getStore(cueListNumber: number, cueNumber: number) {
        const key = getKey(cueListNumber, cueNumber);
        let store = cueActions.get(key);
        if (!store) {
            store = writable<Action[]>([]);
            cueActions.set(key, store);
        }
        return store;
    }

    async function refresh(cueListNumber: number, cueNumber: number) {
        const store = getStore(cueListNumber, cueNumber);
        try {
            const result = await Commands.EnumerateAction({ cueListNumber, cueNumber });
            if (result && result.actions) {
                store.set(result.actions.map((a: any) => a.action));
            } else {
                store.set([]);
            }
        } catch (err) {
            console.error(`Failed to enumerate actions for cue ${cueNumber} in list ${cueListNumber}:`, err);
        }
    }

    async function updateAction(cueListNumber: number, cueNumber: number, actionNumber: number, key: string, value: string) {
        try {
            await Commands.UpdateAction({ 
                cueListNumber, 
                cueNumber, 
                actionNumber, 
                key, 
                value 
            });
        } catch (err) {
            console.error(`Failed to update action ${actionNumber} for cue ${cueNumber}:`, err);
        }
    }

    async function create(cueListNumber: number, cueNumber: number, actionNumber: number = 0) {
        try {
            await Commands.CreateAction({ 
                cueListNumber, 
                cueNumber, 
                actionNumber 
            });
        } catch (err) {
            console.error('Failed to create action:', err);
        }
    }

    async function remove(cueListNumber: number, cueNumber: number, actionNumber: number) {
        try {
            await Commands.DeleteAction({ 
                cueListNumber, 
                cueNumber, 
                actionNumber 
            });
        } catch (err) {
            console.error(`Failed to delete action ${actionNumber}:`, err);
        }
    }

    async function move(cueListNumber: number, cueNumber: number, originalActionNumber: number, newActionNumber: number) {
        try {
            await Commands.MoveAction({ 
                cueListNumber, 
                cueNumber, 
                originalActionNumber, 
                newActionNumber 
            });
        } catch (err) {
            console.error(`Failed to move action ${originalActionNumber} to ${newActionNumber}:`, err);
        }
    }

    // Subscribe to backend events
    Events.On('event.action.created', (event: any) => {
        const { action } = event.data;
        const store = cueActions.get(getKey(action.cueListNumber, action.cueNumber));
        if (!store) return;

        store.update(actions => {
            const newActions = [...actions, action];
            return newActions.sort((a, b) => a.actionNumber - b.actionNumber);
        });
    });

    Events.On('event.action.updated', (event: any) => {
        const { action } = event.data;
        const store = cueActions.get(getKey(action.cueListNumber, action.cueNumber));
        if (!store) return;

        store.update(actions => actions.map(a => 
            a.actionNumber === action.actionNumber ? action : a
        ));
    });

    Events.On('event.action.deleted', (event: any) => {
        const { action } = event.data;
        const store = cueActions.get(getKey(action.cueListNumber, action.cueNumber));
        if (!store) return;

        store.update(actions => actions.filter(a => a.actionNumber !== action.actionNumber));
    });

    return {
        byCue: (cueListNumber: number, cueNumber: number) => {
            const store = getStore(cueListNumber, cueNumber);
            return {
                subscribe: store.subscribe,
                refresh: () => refresh(cueListNumber, cueNumber),
                updateAction: (actionNumber: number, key: string, value: string) => updateAction(cueListNumber, cueNumber, actionNumber, key, value),
                create: (actionNumber: number = 0) => create(cueListNumber, cueNumber, actionNumber),
                remove: (actionNumber: number) => remove(cueListNumber, cueNumber, actionNumber),
                move: (originalActionNumber: number, newActionNumber: number) => move(cueListNumber, cueNumber, originalActionNumber, newActionNumber)
            };
        }
    };
}

export const actions = createActionStore();
