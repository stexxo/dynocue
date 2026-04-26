export function formatTime(nanos: number): string {
	const ms = Math.floor(nanos / 1000000);
	const date = new Date(ms);
	const h = date.getUTCHours().toString().padStart(2, '0');
	const m = date.getUTCMinutes().toString().padStart(2, '0');
	const s = date.getUTCSeconds().toString().padStart(2, '0');
	const msPart = date.getUTCMilliseconds().toString().padStart(3, '0');
	return `${h}:${m}:${s}.${msPart}`;
}
