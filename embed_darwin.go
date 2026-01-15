//go:build darwin && arm64

package main

import _ "embed"

//go:embed ffmpeg/darwin-arm64/ffmpeg
var ffmpegBin []byte

const ffmpegName = "ffmpeg"
