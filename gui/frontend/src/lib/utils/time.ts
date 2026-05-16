// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

export function parseTimeToMs(time: any): number {
	if (!time) return 0;
	if (typeof time === 'number') {
		if (!Number.isFinite(time)) return 0;
		if (time > 1e15) return time / 1000000;
		return time;
	}
	const d = new Date(time);
	const ms = d.getTime();
	return isNaN(ms) ? 0 : ms;
}

export function formatTime(nanos: number): string {
	const n = Number(nanos);
	if (!Number.isFinite(n)) return '00:00:00.000';
	const ms = Math.floor(n / 1000000);
	const date = new Date(ms);
	if (isNaN(date.getTime())) return '00:00:00.000';
	const h = date.getUTCHours().toString().padStart(2, '0');
	const m = date.getUTCMinutes().toString().padStart(2, '0');
	const s = date.getUTCSeconds().toString().padStart(2, '0');
	const msPart = date.getUTCMilliseconds().toString().padStart(3, '0');
	return `${h}:${m}:${s}.${msPart}`;
}
