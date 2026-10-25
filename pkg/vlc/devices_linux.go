//go:build linux

package vlc

import (
	"errors"
	"slices"
	"strings"

	vlc "github.com/adrg/libvlc-go/v3"
)

// GetDevices Uses ALSA and dmix to get the list of audio devices available for use
func GetDevices() ([]AudioDevices, error) {
	// Check if ALSA is available
	outputs, err := vlc.AudioOutputList()
	if err != nil {
		return nil, err
	}
	if !slices.ContainsFunc(outputs, func(output *vlc.AudioOutput) bool {
		if output == nil {
			return false
		}
		return output.Name == "alsa"
	}) {
		return nil, errors.New("alsa not found in audio output list")
	}

	// Get Alsa Dmix devices
	var out []AudioDevices
	devices, err := vlc.ListAudioOutputDevices("alsa")
	if err != nil {
		return nil, err
	}
	for _, d := range devices {
		if strings.HasPrefix(d.Name, "dmix:") {
			descriptionSplit := strings.Split(d.Description, ",")

			var friendlyName string
			var description string
			if len(descriptionSplit) > 1 {
				friendlyName = descriptionSplit[0]
				description = strings.Join(descriptionSplit[1:], ",")
			} else {
				friendlyName = d.Name
				description = d.Description
			}

			out = append(out, AudioDevices{Name: d.Name, FriendlyName: friendlyName, Description: description})
		}
	}
	return out, nil
}
