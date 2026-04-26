// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cues

import "time"

type Playback struct {
}

func (p *Playback) StartMainLoop() {
	go func() {
		ticker := time.NewTicker(16 * time.Millisecond) // ~60 FPS
		for {
			select {
			case <-ticker.C:
				p.loop()
			}
		}
	}()
}

func (p *Playback) loop() {

}
