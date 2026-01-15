//go:build linux && amd64

package main

import _ "embed"

//go:embed ffmpeg/linux-amd64/ffmpeg
var ffmpegBin []byte

const ffmpegName = "ffmpeg"
