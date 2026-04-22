import {TabManager} from "$lib/components/tabs/tabTypes.svelte";
import CueLists from "./CueLists.svelte";

export const cueListTabState = new TabManager(
    [
        { id: 'cueLists', label: 'Cue Lists', content: CueLists },
    ],
    "cueLists"
);