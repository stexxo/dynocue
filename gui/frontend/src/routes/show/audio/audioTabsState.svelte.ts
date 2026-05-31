// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

import { TabManager } from '$lib/components/tabs/tabTypes.svelte';
import AudioFiles from './AudioFiles.svelte';
import AudioDevices from './AudioDevices.svelte';
import AudioSources from './AudioSources.svelte';

export const audioTabState = new TabManager(
	[
		{ id: 'audioFiles', content: AudioFiles, label: 'Audio Files' },
		{ id: 'audioSources', content: AudioSources, label: 'Audio Sources' },
		{ id: 'audioDevices', content: AudioDevices, label: 'Audio Devices' }
	],
	'audioFiles'
);
