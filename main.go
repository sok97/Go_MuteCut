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

	flag.Parse()

	// Check if any flags were provided (excluding default values where possible to detect)
	// A simple way is to check if input is empty, as it's required for non-interactive mode.
	if *inputPtr == "" {
		// Try interactive mode
		fmt.Println("No input file provided via flags. Entering Interactive Mode...")
		interactiveConfig := interactiveMode()

		// Merge interactive config into the main logic
		// We'll just overwrite the pointers or variables used later
		*inputPtr = interactiveConfig.InputFile
		*outputPtr = interactiveConfig.OutputFile // might be empty, auto-gen logic handles it
		*startPtr = interactiveConfig.StartTime
		*endPtr = interactiveConfig.EndTime

		*muteStartPtr = interactiveConfig.MuteStart
		*muteEndPtr = interactiveConfig.MuteEnd
		// We keep defaults for others or could ask for them too, but let's stick to the requested ones
	}

	if *inputPtr == "" {
		fmt.Println("Error: Input file required.")
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
	simpleCut(cfg)
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
	fmt.Print("Enter input video file path: ")
	if scanner.Scan() {
		cfg.InputFile = strings.TrimSpace(scanner.Text())
	}

	// 2. Mode Selection
	fmt.Println("Select Mode:")
	fmt.Println("1. Cut (Trim video)")
	fmt.Println("2. Mute (Mute a section)")
	fmt.Print("Enter choice (1 or 2): ")
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
	} else {
		fmt.Println("Invalid mode selected. Exiting.")
	}

	return cfg
}
