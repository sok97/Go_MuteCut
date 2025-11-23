package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func extractAudio(cfg Config) {
	// Determine output filename if not set
	outputFile := cfg.OutputFile
	if outputFile == "" {
		ext := filepath.Ext(cfg.InputFile)
		base := strings.TrimSuffix(cfg.InputFile, ext)
		outputFile = base + ".mp3"
	} else {
		// Ensure output ends with .mp3
		if !strings.HasSuffix(strings.ToLower(outputFile), ".mp3") {
			outputFile += ".mp3"
		}
	}
	cfg.OutputFile = outputFile

	fmt.Printf("Extracting MP3 to: %s\n", cfg.OutputFile)

	// ffmpeg -i input.mp4 -vn -acodec libmp3lame -q:a 2 output.mp3
	args := []string{
		"-i", cfg.InputFile,
		"-vn", // No video
		"-acodec", "libmp3lame",
		"-q:a", "2", // High quality variable bitrate
		"-y", // Overwrite
		cfg.OutputFile,
	}

	runFFmpeg(cfg, args)
}
