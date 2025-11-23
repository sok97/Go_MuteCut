package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kkdai/youtube/v2"
)

func downloadYoutubeVideo(url string) (string, error) {
	fmt.Println("Initializing YouTube client...")
	client := youtube.Client{}

	fmt.Printf("Fetching video info for: %s\n", url)
	video, err := client.GetVideo(url)
	if err != nil {
		return "", fmt.Errorf("failed to get video info: %w", err)
	}

	fmt.Printf("Found video: %s\n", video.Title)

	// Find the best format that has both audio and video
	// The library's formats are sorted by quality usually, but we need to check for audio
	var format *youtube.Format
	formats := video.Formats.WithAudioChannels() // Filter formats with audio

	if len(formats) > 0 {
		// Pick the first one (usually best quality muxed)
		// Or we could sort by quality if needed, but default order is often decent for muxed
		format = &formats[0]
	} else {
		return "", fmt.Errorf("no suitable video format with audio found")
	}

	fmt.Printf("Downloading format: %s (Quality: %s)\n", format.MimeType, format.QualityLabel)

	stream, _, err := client.GetStream(video, format)
	if err != nil {
		return "", fmt.Errorf("failed to get stream: %w", err)
	}
	defer stream.Close()

	// Sanitize filename
	cleanTitle := sanitizeFilename(video.Title)
	outputFile := cleanTitle + ".mp4"
	// Ensure unique filename
	outputFile = ensureUniqueFilename(outputFile)

	fmt.Printf("Downloading to: %s\n", outputFile)
	file, err := os.Create(outputFile)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, stream)
	if err != nil {
		return "", fmt.Errorf("failed to download video: %w", err)
	}

	return outputFile, nil
}

func sanitizeFilename(name string) string {
	// Remove invalid characters
	re := regexp.MustCompile(`[<>:"/\\|?*]`)
	return re.ReplaceAllString(name, "_")
}

func ensureUniqueFilename(path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}

	ext := filepath.Ext(path)
	base := strings.TrimSuffix(path, ext)

	for i := 1; ; i++ {
		newPath := fmt.Sprintf("%s_%d%s", base, i, ext)
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
	}
}
