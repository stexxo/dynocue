// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

import { TabManager } from '$lib/components/tabs/tabTypes.svelte';
import CueLists from './CueLists.svelte';

export const cueListTabState = new TabManager(
	[{ id: 'cueLists', content: CueLists, label: 'Cue Lists' }],
	'cueLists'
);
