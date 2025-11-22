# Go Video Chopper

A lightweight, high-performance video cutter and muter written in Go. It uses FFmpeg to perform precise cuts or mute sections of a video.

## Features

*   **Interactive Mode**: Easy-to-use prompts for cutting and muting without complex flags.
*   **Precise Cutting**: Cut video segments with `-start` and `-end` flags.
*   **Local FFmpeg**: Uses a local `bin/` folder for FFmpeg, keeping your system clean.
*   **Fast Processing**: Optimized filter graphs for single-pass processing.

## Installation

1.  Clone the repository.
2.  Ensure you have Go installed.
3.  Run the setup script (if provided) or manually place `ffmpeg.exe` and `ffprobe.exe` in the `bin/` folder.

## Usage

Run the tool using `go run main.go`.

### Interactive Mode
Run without arguments to enter interactive mode:
```bash
go run main.go
```
Follow the on-screen prompts to select input file, mode (Cut/Mute), and time ranges.


### Cut Only
Trim a video from 00:01:30 to 00:02:00:
```bash
go run main.go -i input.mp4 -start 00:01:30 -end 00:02:00
```


### Mute Range
Mute audio from 00:06:00 to 00:06:30:
```bash
go run main.go -i input.mp4 -mute-start 00:06:00 -mute-end 00:06:30
```

### Options

| Flag | Description | Default |
| :--- | :--- | :--- |
| `-i` | Input video file (Required) | |
| `-o` | Output video file | `*_cleaned.mp4` |

| `-start` | Start time (e.g., `10`, `00:01:30`) | |
| `-end` | End time (e.g., `20`, `00:02:00`) | |

| `-mute-start`| Start time to mute | |
| `-mute-end`| End time to mute | |
| `-crf` | Quality (lower is better) | `23` |
| `-preset` | Encoding speed | `medium` |

## Limitations

*   **Re-encoding**: The tool always re-encodes the video (using H.264/AAC). It does not perform "lossless" stream copying, so quality generation loss is possible, and it is slower than a simple cut.
*   **Tracks**: Only processes the primary video and audio track. Subtitles, chapters, and additional audio tracks (e.g., commentary) will be lost.
*   **Codecs**: Hardcoded to use `libx264` and `aac`.
*   **Platform**: The setup script and binary logic are currently optimized for Windows.

## Project Structure

```
├── bin/            # Local FFmpeg binaries (ignored by git)
├── main.go         # Source code
├── .gitignore      # Git configuration
└── README.md       # Documentation
```
