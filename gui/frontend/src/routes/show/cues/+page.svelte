<!--
  This Source Code Form is subject to the terms of the Mozilla Public
  License, v. 2.0. If a copy of the MPL was not distributed with this
  file, You can obtain one at https://mozilla.org/MPL/2.0/.
-->

<script>
	import Tabs from '$lib/components/tabs/Tabs.svelte';
	import { cueListTabState } from './cuesListsState.svelte';
</script>

<div class="h-full">
	<Tabs tabManager={cueListTabState}></Tabs>
</div>
