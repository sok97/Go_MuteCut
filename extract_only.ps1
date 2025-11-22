$ErrorActionPreference = "Stop"
$binDir = Join-Path $PSScriptRoot "bin"
$zipPath = Join-Path $binDir "ffmpeg.zip"

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
    Write-Host "✅ FFmpeg setup complete."
} else {
    Write-Error "Could not find extracted folder."
}
