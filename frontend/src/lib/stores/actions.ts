/**
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 *
 * SPDX-License-Identifier: MPL-2.0
 */

import { writable } from 'svelte/store';
import * as Commands from '../../../bindings/gitlab.com/stexxo/dynocue/internal/gui/commands';
import { CueAction as Action } from '../../../bindings/git lab.com/stexxo/dynocue/api/cues/models';
import { Events } from '@wailsio/runtime';

function createActionStore() {
    const { subscribe, set, update } = writable<Action[]>([]);
    let currentCueListNumber: number | null = null;
    let currentCueNumber: number | null = null;

    async function refresh(cueListNumber: number, cueNumber: number) {
        currentCueListNumber = cueListNumber;
        currentCueNumber = cueNumber;
        try {
            const result = await Commands.EnumerateAction({ cueListNumber, cueNumber });
            if (result && result.actions) {
                // Backend returns list of GetActionOutput which has Action: CueAction { ActionNumber, Label }
                set(result.actions.map((a: any) => a.action));
            } else {
                set([]);
            }
        } catch (err) {
            console.error(`Failed to enumerate actions for cue ${cueNumber} in list ${cueListNumber}:`, err);
        }
    }

    async function updateAction(actionNumber: number, key: string, value: string) {
        if (currentCueListNumber === null || currentCueNumber === null) return;
        try {
            await Commands.UpdateAction({ 
                cueListNumber: currentCueListNumber, 
                cueNumber: currentCueNumber, 
                actionNumber, 
                key, 
                value 
            });
        } catch (err) {
            console.error(`Failed to update action ${actionNumber} for cue ${currentCueNumber}:`, err);
        }
    }

    async function create(actionNumber: number = 0) {
        if (currentCueListNumber === null || currentCueNumber === null) return;
        try {
            await Commands.CreateAction({ 
                cueListNumber: currentCueListNumber, 
                cueNumber: currentCueNumber, 
                actionNumber 
            });
        } catch (err) {
            console.error('Failed to create action:', err);
        }
    }

    async function remove(actionNumber: number) {
        if (currentCueListNumber === null || currentCueNumber === null) return;
        try {
            await Commands.DeleteAction({ 
                cueListNumber: currentCueListNumber, 
                cueNumber: currentCueNumber, 
                actionNumber 
            });
        } catch (err) {
            console.error(`Failed to delete action ${actionNumber}:`, err);
        }
    }

    async function move(originalActionNumber: number, newActionNumber: number) {
        if (currentCueListNumber === null || currentCueNumber === null) return;
        try {
            await Commands.MoveAction({ 
                cueListNumber: currentCueListNumber, 
                cueNumber: currentCueNumber, 
                originalActionNumber, 
                newActionNumber 
            });
        } catch (err) {
            console.error(`Failed to move action ${originalActionNumber} to ${newActionNumber}:`, err);
        }
    }

    // Subscribe to backend events
    Events.On('event.action.created', (event: any) => {
        const { cueListNumber, cueNumber, action } = event.data;
        if (cueListNumber !== currentCueListNumber || cueNumber !== currentCueNumber) return;

        update(actions => {
            const newActions = [...actions, action];
            return newActions.sort((a, b) => a.actionNumber - b.actionNumber);
        });
    });

    Events.On('event.action.updated', (event: any) => {
        const { cueListNumber, cueNumber, action } = event.data;
        if (cueListNumber !== currentCueListNumber || cueNumber !== currentCueNumber) return;

        update(actions => actions.map(a => 
            a.actionNumber === action.actionNumber ? action : a
        ));
    });

    Events.On('event.action.deleted', (event: any) => {
        const { cueListNumber, cueNumber, actionNumber } = event.data;
        if (cueListNumber !== currentCueListNumber || cueNumber !== currentCueNumber) return;

        update(actions => actions.filter(a => a.actionNumber !== actionNumber));
    });

    return {
        subscribe,
        refresh,
        updateAction,
        create,
        remove,
        move
    };
}

export const actions = createActionStore();
