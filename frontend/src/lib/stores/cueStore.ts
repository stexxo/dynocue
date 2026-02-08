import { writable } from 'svelte/store';
import { Events } from '@wailsio/runtime';
import { CueService } from "../../../bindings/gitlab.com/stexxo/dynocue/dynod/internal/subsystems/gui/api/index";
import type { CueListNumber } from "../../../bindings/gitlab.com/stexxo/dynocue/dynod/internal/cues";

function createCueStore() {
    // Fixed syntax: use CueListNumber[] or Array<CueListNumber>
    const { subscribe, set } = writable<CueListNumber[]>([]);

    async function refresh() {
        try {
            // Wails v3 Go calls usually return [data, error]
            const [data, ok] = await CueService.GetCueLists();

            if (ok) {
                set(data);
                console.log("Fetched CueLists:", data);
            } else {
                console.error("Failed to fetch CueLists");
            }
        } catch (e) {
            console.error("RPC Call Error:", e);
        }
    }

    // Initialize the Wails Event Listener
    // Note: Ensure the string matches your Go Events.Emit exactly
    const unsubscribe = Events.On('show.cues.lists.updated', () => {
        console.log("CueLists updated!")
        refresh();
    });

    // Initial load
    refresh();

    return {
        subscribe,
        refresh,
        // Optional: exposure to kill listener if store is destroyed
        destroy: () => unsubscribe()
    };
}

export const cueStore = createCueStore();