package vlc

import (
	"sync"

	vlc "github.com/adrg/libvlc-go/v3"
)

var once sync.Once

func Initialize() error {
	var e error
	once.Do(func() {
		if err := vlc.Init(); err != nil {
			e = err
		}
	})
	return e
}
