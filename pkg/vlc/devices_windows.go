//go:build windows

package vlc

import (
	"errors"
	"slices"
	"strings"

	vlc "github.com/adrg/libvlc-go/v3"
)

// GetDevices uses the Windows MMDevice API to get the list of available audio devices
func GetDevices() ([]AudioDevices, error) {
	// 1. Check if the Windows Multimedia Device (mmdevice) output is available
	outputs, err := vlc.AudioOutputList()
	if err != nil {
		return nil, err
	}

	// 'mmdevice' is the standard for modern Windows (WASAPI)
	// Alternatively, you could check for "directsound" if supporting older environments
	targetOutput := "mmdevice"

	if !slices.ContainsFunc(outputs, func(output *vlc.AudioOutput) bool {
		return output != nil && output.Name == targetOutput
	}) {
		return nil, errors.New("mmdevice audio output not found")
	}

	// 2. Get the devices for the mmdevice output
	devices, err := vlc.ListAudioOutputDevices(targetOutput)
	if err != nil {
		return nil, err
	}

	var out []AudioDevices
	for _, d := range devices {
		// Windows device names are usually complex GUIDs (e.g., "{0.0.0.00000000}.{...}")
		// The 'Default' device usually has a Name of "default".
		// We filter out empty names, but you generally want all valid hardware endpoints.
		if d.Name == "" {
			continue
		}

		// Windows descriptions are often already user-friendly (e.g., "Speakers (Realtek High Definition Audio)")
		// We can attempt to clean it up or split it if necessary, but usually, Description works as the FriendlyName.
		friendlyName := d.Description
		description := d.Description

		// If VLC provides a structured description (e.g., "Device Name, Driver"), you can split it:
		if descriptionSplit := strings.Split(d.Description, ","); len(descriptionSplit) > 1 {
			friendlyName = strings.TrimSpace(descriptionSplit[0])
			description = strings.TrimSpace(strings.Join(descriptionSplit[1:], ","))
		}

		out = append(out, AudioDevices{
			Name:         d.Name,       // This will be the GUID string needed by VLC to set the device
			FriendlyName: friendlyName, // Human-readable name
			Description:  description,  // Extra details if available
		})
	}

	return out, nil
}
