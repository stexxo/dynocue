
export function formatSize(bytes: number) {
    const size = Number(bytes);
    if (!Number.isFinite(size)) return '0 B';

    const units = ['B', 'KB', 'MB', 'GB', 'TB'];
    const sign = size < 0 ? '-' : '';
    let value = Math.abs(size);
    let unitIndex = 0;

    while (value >= 1024 && unitIndex < units.length - 1) {
        value /= 1024;
        unitIndex += 1;
    }

    const formatted =
        unitIndex === 0 ? value.toString() : value.toFixed(value >= 10 ? 0 : 1).replace(/\.0$/, '');
    return `${sign}${formatted} ${units[unitIndex]}`;
}