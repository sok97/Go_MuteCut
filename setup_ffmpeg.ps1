$ErrorActionPreference = "Stop"

$binDir = Join-Path $PSScriptRoot "bin"
if (-not (Test-Path $binDir)) {
    New-Item -ItemType Directory -Path $binDir | Out-Null
}

$ffmpegUrl = "https://www.gyan.dev/ffmpeg/builds/ffmpeg-release-essentials.zip"
$zipPath = Join-Path $binDir "ffmpeg.zip"

Write-Host "Downloading FFmpeg..."
Invoke-WebRequest -Uri $ffmpegUrl -OutFile $zipPath

Write-Host "Extracting..."
Expand-Archive -Path $zipPath -DestinationPath $binDir -Force

# Move binaries
$extractedFolder = Get-ChildItem -Path $binDir -Directory | Where-Object { $_.Name -like "ffmpeg-*-essentials_build" } | Select-Object -First 1

if ($extractedFolder) {
    $binSubDir = Join-Path $extractedFolder.FullName "bin"
    Move-Item -Path "$binSubDir\ffmpeg.exe" -Destination $binDir -Force
    Move-Item -Path "$binSubDir\ffprobe.exe" -Destination $binDir -Force
    
    # Cleanup
    Remove-Item -Path $extractedFolder.FullName -Recurse -Force
    Remove-Item -Path $zipPath -Force
    Write-Host "FFmpeg installed successfully."
} else {
    Write-Error "Could not find extracted folder."
}
