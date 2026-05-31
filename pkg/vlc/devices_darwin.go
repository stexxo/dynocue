//go:build darwin

package vlc

import (
	"errors"
	"slices"
	"strings"

	vlc "github.com/adrg/libvlc-go/v3"
)

// GetDevices uses the macOS CoreAudio (AUHAL) layer to get the list of available audio devices
func GetDevices() ([]AudioDevices, error) {
	// 1. Check if the CoreAudio (auhal) output is available
	outputs, err := vlc.AudioOutputList()
	if err != nil {
		return nil, err
	}

	// 'auhal' is the standard for macOS
	targetOutput := "auhal"

	if !slices.ContainsFunc(outputs, func(output *vlc.AudioOutput) bool {
		return output != nil && output.Name == targetOutput
	}) {
		return nil, errors.New("auhal audio output not found")
	}

	// 2. Get the devices for the auhal output
	devices, err := vlc.ListAudioOutputDevices(targetOutput)
	if err != nil {
		return nil, err
	}

	var out []AudioDevices
	for _, d := range devices {
		// Skip empty entries
		if d.Name == "" {
			continue
		}

		// macOS device names are usually numeric IDs or system strings (e.g., "43", "57", or "default")
		// The Description field holds the human-readable name (e.g., "External Headphones")
		friendlyName := d.Description
		description := d.Description

		// Parse description strings if they contain comma-separated metadata
		if descriptionSplit := strings.Split(d.Description, ","); len(descriptionSplit) > 1 {
			friendlyName = strings.TrimSpace(descriptionSplit[0])
			description = strings.TrimSpace(strings.Join(descriptionSplit[1:], ","))
		}

		out = append(out, AudioDevices{
			Name:         d.Name,       // Usually an integer string (e.g., "53") needed by VLC to switch devices
			FriendlyName: friendlyName, // Human-readable name (e.g., "MacBook Pro Speakers")
			Description:  description,  // Extra details if available
		})
	}

	return out, nil
}
