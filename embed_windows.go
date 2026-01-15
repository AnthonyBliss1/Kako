//go:build windows && amd64

package main

import _ "embed"

//go:embed ffmpeg/windows-amd64/ffmpeg.exe
var ffmpegBin []byte

const ffmpegName = "ffmpeg.exe"
