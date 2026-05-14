/**
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

/**
 * Represents the loading state of an operation.
 */
export class LoadingState {
	isLoading = $state(false);
}

/**
 * Creates a new, local loading state.
 * @returns A reactive loading state instance.
 */
export function createLoadingState(): LoadingState {
	return new LoadingState();
}

/**
 * Wraps an async function with loading state management.
 * @param state The loading state to manage.
 * @param fn The async function to execute.
 * @returns A wrapped function that manages the loading state.
 */
export function withLoading<T extends (...args: any[]) => Promise<any>>(
	state: LoadingState,
	fn: T
): (...args: Parameters<T>) => Promise<ReturnType<T>> {
	return async (...args: Parameters<T>) => {
		if (state.isLoading) return;

		state.isLoading = true;
		try {
			return await fn(...args);
		} finally {
			state.isLoading = false;
		}
	};
}
