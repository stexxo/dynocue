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
