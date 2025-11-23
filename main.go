package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	InputFile  string
	OutputFile string

	MaxVideoLen float64
	MaxFileSize int64
	Preset      string
	CRF         int
	// Mute Flags
	MuteStart string
	MuteEnd   string
	StartTime string
	EndTime   string

	FfmpegBin  string
	FfprobeBin string
	Verbose    bool
	ExtractMP3 bool
}

type Segment struct {
	Start float64
	End   float64
}

func main() {
	inputPtr := flag.String("i", "", "Input video file (required)")
	outputPtr := flag.String("o", "", "Output file (default: auto-generated)")

	startPtr := flag.String("start", "", "Start time (e.g., '10', '00:01:30')")
	endPtr := flag.String("end", "", "End time (e.g., '20', '00:02:00')")

	// Mute Flags
	muteStartPtr := flag.String("mute-start", "", "Start time to mute (e.g., '00:06:00')")
	muteEndPtr := flag.String("mute-end", "", "End time to mute (e.g., '00:06:30')")

	presetPtr := flag.String("preset", "medium", "Encoding preset")
	crfPtr := flag.Int("crf", 23, "CRF Quality")
	verbosePtr := flag.Bool("v", false, "Verbose output")
	mp3Ptr := flag.Bool("mp3", false, "Extract MP3 audio")
	urlPtr := flag.String("url", "", "YouTube Video URL")

	flag.Parse()

	// Check if any flags were provided (excluding default values where possible to detect)
	// A simple way is to check if input is empty, as it's required for non-interactive mode.
	if *inputPtr == "" && *urlPtr == "" {
		// Try interactive mode
		fmt.Println("No input file provided via flags. Entering Interactive Mode...")
		interactiveConfig := interactiveMode()

		// Merge interactive config into the main logic
		// We'll just overwrite the pointers or variables used later
		if interactiveConfig.InputFile != "" {
			*inputPtr = interactiveConfig.InputFile
		}
		// If interactive mode returned a URL (we'll handle this by checking if InputFile is a URL or adding a field)
		// Actually, let's just use InputFile for both and detect if it's a URL.

		*outputPtr = interactiveConfig.OutputFile // might be empty, auto-gen logic handles it
		*startPtr = interactiveConfig.StartTime
		*endPtr = interactiveConfig.EndTime

		*muteStartPtr = interactiveConfig.MuteStart
		*muteEndPtr = interactiveConfig.MuteEnd
		*mp3Ptr = interactiveConfig.ExtractMP3
		// We keep defaults for others or could ask for them too, but let's stick to the requested ones
	}

	if *inputPtr == "" && *urlPtr == "" {
		fmt.Println("Error: Input file or YouTube URL required.")
		os.Exit(1)
	}

	// Handle YouTube Download
	if *urlPtr != "" {
		fmt.Println("YouTube URL provided. Downloading...")
		downloadedFile, err := downloadYoutubeVideo(*urlPtr)
		if err != nil {
			fmt.Printf("Error downloading YouTube video: %v\n", err)
			os.Exit(1)
		}
		*inputPtr = downloadedFile
	} else if strings.HasPrefix(*inputPtr, "http://") || strings.HasPrefix(*inputPtr, "https://") || strings.HasPrefix(*inputPtr, "www.") {
		// Detect URL from interactive input
		fmt.Println("YouTube URL detected. Downloading...")
		downloadedFile, err := downloadYoutubeVideo(*inputPtr)
		if err != nil {
			fmt.Printf("Error downloading YouTube video: %v\n", err)
			os.Exit(1)
		}
		*inputPtr = downloadedFile
	}

	// Validate Input File
	info, err := os.Stat(*inputPtr)
	if os.IsNotExist(err) {
		fmt.Printf("Error: Input file '%s' does not exist.\n", *inputPtr)
		os.Exit(1)
	}
	if err != nil {
		fmt.Printf("Error: Cannot access input file: %v\n", err)
		os.Exit(1)
	}
	if info.IsDir() {
		fmt.Printf("Error: Input '%s' is a directory. Please specify a video file.\n", *inputPtr)
		os.Exit(1)
	}

	outputFile := *outputPtr
	if outputFile == "" {
		ext := filepath.Ext(*inputPtr)
		base := strings.TrimSuffix(*inputPtr, ext)
		suffix := "_cleaned"

		if *muteStartPtr != "" {
			suffix += "_muted"
		}
		outputFile = base + suffix + ext
	}

	cfg := Config{
		InputFile:  *inputPtr,
		OutputFile: outputFile,
		StartTime:  *startPtr,
		EndTime:    *endPtr,
		MuteStart:  *muteStartPtr,
		MuteEnd:    *muteEndPtr,
		Preset:     *presetPtr,
		CRF:        *crfPtr,
		Verbose:    *verbosePtr,
		ExtractMP3: *mp3Ptr,
	}

	cfg.FfmpegBin = resolveBinary("ffmpeg")
	cfg.FfprobeBin = resolveBinary("ffprobe")

	if cfg.FfmpegBin == "" || cfg.FfprobeBin == "" {
		fmt.Println("Error: ffmpeg or ffprobe not found in 'bin' folder or system PATH.")
		fmt.Println("Please run the setup script to download them.")
		os.Exit(1)
	}

	_ = os.MkdirAll(filepath.Dir(cfg.OutputFile), 0755)

	start := time.Now()

	fmt.Println("Mode: Processing (Cut/Mute)...")
	if cfg.ExtractMP3 {
		extractAudio(cfg)
	} else {
		simpleCut(cfg)
	}
	printStats(cfg, time.Since(start))
}

func getInputArgs(cfg Config) []string {
	args := []string{}
	if cfg.StartTime != "" {
		args = append(args, "-ss", cfg.StartTime)
	}
	if cfg.EndTime != "" {
		args = append(args, "-to", cfg.EndTime)
	}
	args = append(args, "-i", cfg.InputFile)
	return args
}

func simpleCut(cfg Config) {
	inputArgs := getInputArgs(cfg)

	// Build Filter Chain
	var filters []string
	if cfg.MuteStart != "" && cfg.MuteEnd != "" {

		startSec := parseTimeToSeconds(cfg.MuteStart)
		endSec := parseTimeToSeconds(cfg.MuteEnd)

		filters = append(filters, fmt.Sprintf("volume=0:enable='between(t,%.3f,%.3f)'", startSec, endSec))
	}

	args := append(inputArgs,
		"-c:v", "libx264", "-preset", cfg.Preset, "-crf", strconv.Itoa(cfg.CRF),
		"-c:a", "aac", "-b:a", "192k",
	)

	if len(filters) > 0 {
		args = append(args, "-af", strings.Join(filters, ","))
	}

	args = append(args, "-y", cfg.OutputFile)
	runFFmpeg(cfg, args)
}

// Helper to parse "HH:MM:SS" or "SS" to float seconds
func parseTimeToSeconds(ts string) float64 {
	// Try simple float first
	if val, err := strconv.ParseFloat(ts, 64); err == nil {
		return val
	}

	// Try HH:MM:SS or MM:SS
	parts := strings.Split(ts, ":")
	var seconds float64
	multiplier := 1.0

	for i := len(parts) - 1; i >= 0; i-- {
		val, _ := strconv.ParseFloat(parts[i], 64)
		seconds += val * multiplier
		multiplier *= 60
	}
	return seconds
}

func runFFmpeg(cfg Config, args []string) {
	cmd := exec.Command(cfg.FfmpegBin, args...)
	if cfg.Verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		fmt.Printf(" FFmpeg Error: %v\n", err)
		os.Exit(1)
	}
}

func resolveBinary(name string) string {
	exePath, err := os.Executable()
	if err == nil {
		binPath := filepath.Join(filepath.Dir(exePath), "bin", name)
		if _, err := os.Stat(binPath); err == nil {
			return binPath
		}
		if _, err := os.Stat(binPath + ".exe"); err == nil {
			return binPath + ".exe"
		}
	}

	cwd, err := os.Getwd()
	if err == nil {
		binPath := filepath.Join(cwd, "bin", name)
		if _, err := os.Stat(binPath); err == nil {
			return binPath
		}
		if _, err := os.Stat(binPath + ".exe"); err == nil {
			return binPath + ".exe"
		}
	}

	path, _ := exec.LookPath(name)
	return path
}

func printStats(cfg Config, elapsed time.Duration) {
	fmt.Println("\n Done!")
	fmt.Printf("Output: %s\n", cfg.OutputFile)
}

func interactiveMode() Config {
	scanner := bufio.NewScanner(os.Stdin)
	cfg := Config{}

	// 1. Input File
	fmt.Print("Enter input video file path or YouTube URL: ")
	if scanner.Scan() {
		cfg.InputFile = strings.TrimSpace(scanner.Text())
	}

	// 2. Mode Selection
	fmt.Println("Select Mode:")
	fmt.Println("1. Cut (Trim video)")
	fmt.Println("2. Mute (Mute a section)")
	fmt.Println("3. Extract MP3")
	fmt.Print("Enter choice (1, 2, or 3): ")
	var mode string
	if scanner.Scan() {
		mode = strings.TrimSpace(scanner.Text())
	}

	if mode == "1" {
		fmt.Println("--- Cut Mode ---")
		fmt.Print("Enter Start Time (e.g., 00:05 or 5): ")
		if scanner.Scan() {
			cfg.StartTime = strings.TrimSpace(scanner.Text())
		}
		fmt.Print("Enter End Time (e.g., 00:10 or 10): ")
		if scanner.Scan() {
			cfg.EndTime = strings.TrimSpace(scanner.Text())
		}
	} else if mode == "2" {
		fmt.Println("--- Mute Mode ---")
		fmt.Print("Enter Start Time to Mute (e.g., 00:05 or 5): ")
		if scanner.Scan() {
			cfg.MuteStart = strings.TrimSpace(scanner.Text())
		}
		fmt.Print("Enter End Time to Mute (e.g., 00:10 or 10): ")
		if scanner.Scan() {
			cfg.MuteEnd = strings.TrimSpace(scanner.Text())
		}
	} else if mode == "3" {
		fmt.Println("--- MP3 Extraction Mode ---")
		cfg.ExtractMP3 = true
	} else {
		fmt.Println("Invalid mode selected. Exiting.")
	}

	return cfg
}
