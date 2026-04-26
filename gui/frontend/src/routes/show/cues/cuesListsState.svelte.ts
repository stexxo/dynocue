import { TabManager } from '$lib/components/tabs/tabTypes.svelte';
import CueLists from './CueLists.svelte';

export const cueListTabState = new TabManager(
	[{ id: 'cueLists', content: CueLists, label: 'Cue Lists' }],
	'cueLists'
);
