// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package probe

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Info represents the clean metadata results extracted from the local file.
type Info struct {
	Format    string        `json:"format"`
	Duration  time.Duration `json:"duration"`
	SizeBytes int64         `json:"size_bytes"`
}

type ffprobeLocalSchema struct {
	Format struct {
		FormatName string `json:"format_name"`
		Duration   string `json:"duration"`
		Size       string `json:"size"`
	} `json:"format"`
	Streams []struct {
		CodecName string `json:"codec_name"`
		Duration  string `json:"duration"`
	} `json:"streams"`
}

// ProbeFile extracts format, size, and duration natively from a local file path.
// This works instantly across Windows, macOS, and Linux for all major audio extensions.
func ProbeFile(filePath string) (*Info, error) {
	// Verify the file exists locally before executing a system subprocess
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("local file error: %w", err)
	}

	// We pass the raw file path string directly to ffprobe.
	// We no longer need custom stream settings, probesizes, or stdin overrides.
	args := []string{
		"-v", "error",
		"-show_entries", "stream=codec_name,duration:format=format_name,duration,size",
		"-of", "json",
		filePath,
	}

	cmd := exec.Command("ffprobe", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffprobe failed on file %s: %s: %w", filePath, stderr.String(), err)
	}

	var raw ffprobeLocalSchema
	if err := json.Unmarshal(stdout.Bytes(), &raw); err != nil {
		return nil, fmt.Errorf("failed to decode ffprobe file json: %w", err)
	}

	// 1. Resolve Audio Format
	formatName := "UNKNOWN"
	if len(raw.Streams) > 0 && raw.Streams[0].CodecName != "" {
		formatName = strings.ToUpper(raw.Streams[0].CodecName)
	} else if raw.Format.FormatName != "" {
		formatName = strings.ToUpper(strings.Split(raw.Format.FormatName, ",")[0])
	}

	// 2. Resolve Track Duration
	var durationStr string
	if raw.Format.Duration != "" && raw.Format.Duration != "N/A" {
		durationStr = raw.Format.Duration
	} else if len(raw.Streams) > 0 && raw.Streams[0].Duration != "" && raw.Streams[0].Duration != "N/A" {
		durationStr = raw.Streams[0].Duration
	}

	var duration time.Duration
	if durationStr != "" {
		seconds, err := strconv.ParseFloat(durationStr, 64)
		if err == nil {
			duration = time.Duration(seconds * float64(time.Second))
		}
	}

	// 3. Fallback size tracking via OS filesystem if metadata is blank
	sizeBytes := fileInfo.Size()
	if raw.Format.Size != "" {
		if parsedSize, err := strconv.ParseInt(raw.Format.Size, 10, 64); err == nil {
			sizeBytes = parsedSize
		}
	}

	return &Info{
		Format:    formatName,
		Duration:  duration,
		SizeBytes: sizeBytes,
	}, nil
}
