#!/bin/bash

# Create bin directory
mkdir -p bin

# Check if ffmpeg is installed system-wide
if command -v ffmpeg &> /dev/null; then
    echo "ffmpeg is already installed on your system."
    echo "You can use the tool directly."
    exit 0
fi

echo "ffmpeg not found. Downloading static build..."

# Detect OS/Arch
OS=$(uname -s)
ARCH=$(uname -m)

if [ "$OS" == "Linux" ] && [ "$ARCH" == "x86_64" ]; then
    URL="https://johnvansickle.com/ffmpeg/releases/ffmpeg-release-amd64-static.tar.xz"
    FILE="ffmpeg-release-amd64-static.tar.xz"
    
    echo "Downloading from $URL..."
    if command -v curl &> /dev/null; then
        curl -L -o $FILE $URL
    elif command -v wget &> /dev/null; then
        wget -O $FILE $URL
    else
        echo "Error: curl or wget not found. Please install one of them."
        exit 1
    fi
    
    echo "Extracting..."
    tar -xf $FILE
    
    # Move binaries to bin/
    # The tarball extracts to a folder like ffmpeg-X.X-amd64-static/
    DIR=$(tar -tf $FILE | head -1 | cut -f1 -d"/")
    cp $DIR/ffmpeg bin/
    cp $DIR/ffprobe bin/
    
    # Cleanup
    rm $FILE
    rm -rf $DIR
    
    chmod +x bin/ffmpeg
    chmod +x bin/ffprobe
    
    echo "Done! ffmpeg and ffprobe installed to bin/"
else
    echo "Automatic download only supported for Linux x86_64."
    echo "Please install ffmpeg manually using your package manager (e.g., sudo apt install ffmpeg)."
fi
