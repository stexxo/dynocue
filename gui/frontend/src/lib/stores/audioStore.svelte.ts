// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

import {
	EnumerateAudioFiles,
	CreateAudioFile,
	ReplaceAudioFile,
	DeleteAudioFile,
	GetAudioFile,
	AddAudioFile,
	ReplaceAudioFileWithDialog,
	UpdateAudioFileAttributes
} from '../../../bindings/github.com/stexxo/dynocue/gui/services/audioservice';
import { AudioFile } from '../../../bindings/github.com/stexxo/dynocue/components/audio/types';
import { Events } from '@wailsio/runtime';

/**
 * Store for managing audio files.
 */
class AudioStore {
	#files = $state<AudioFile[]>([]);

	constructor() {
		this.load();
		Events.On('event.audio.files.created', () => {
			this.load();
		});
		Events.On('event.audio.files.updated', () => {
			this.load();
		});
		Events.On('event.audio.files.deleted', () => {
			this.load();
		});
		Events.On('event.system.persistence.loaded', () => {
			this.load();
		});
	}

	get files() {
		return this.#files;
	}

	audioFile(id: string): AudioFile | undefined {
		return this.#files.find((file) => file.fileId === id);
	}

	async load() {
		const [files, ok] = await EnumerateAudioFiles();
		if (ok) {
			this.#files = files;
		}
	}

	async create(fileLocation: string) {
		const [fileId, ok] = await CreateAudioFile(fileLocation);
		if (!ok) {
			console.error('Failed to create audio file', fileLocation);
		}
		return fileId;
	}

	async replace(fileId: string, fileLocation: string) {
		const ok = await ReplaceAudioFile(fileId, fileLocation);
		if (!ok) {
			console.error('Failed to replace audio file', fileId, fileLocation);
		}
		return ok;
	}

	async deleteFile(fileId: string) {
		const ok = await DeleteAudioFile(fileId);
		if (!ok) {
			console.error('Failed to delete audio file', fileId);
		}
		return ok;
	}

	async getFile(fileId: string) {
		const [file, ok] = await GetAudioFile(fileId);
		if (!ok) {
			console.error('Failed to get audio file', fileId);
		}
		return file;
	}

	async updateFileAttributes(fileId: string, field: 'number' | 'label', value: any) {
		const ok = await UpdateAudioFileAttributes(fileId, field, value);
		if (!ok) {
			console.error('Failed to update audio file attributes', fileId, field);
		}
		return ok;
	}

	async addWithDialog() {
		const ok = await AddAudioFile();
		if (!ok) {
			console.error('Failed to add audio file with dialog');
		}
		return ok;
	}

	async replaceWithDialog(fileId: string) {
		const ok = await ReplaceAudioFileWithDialog(fileId);
		if (!ok) {
			console.error('Failed to replace audio file with dialog', fileId);
		}
		return ok;
	}
}

export const audioStore = new AudioStore();
