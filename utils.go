package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	_ "image/png"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

func EnsureFFmpeg() (string, error) {
	if len(ffmpegBin) == 0 || ffmpegName == "" {
		return "", fmt.Errorf("ffmpeg not available for this platform: %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	cacheRoot, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	appDir := filepath.Join(cacheRoot, "kako", "bin")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		return "", err
	}

	sum := fmt.Sprintf("%x", sha256.Sum256(ffmpegBin))[:16]
	outPath := filepath.Join(appDir, fmt.Sprintf("%s-%s", sum, ffmpegName))

	if st, err := os.Stat(outPath); err == nil && st.Size() == int64(len(ffmpegBin)) {
		return outPath, nil
	}

	tmp := outPath + ".tmp"
	if err := os.WriteFile(tmp, ffmpegBin, 0o755); err != nil {
		return "", err
	}

	if err := os.Rename(tmp, outPath); err != nil {
		_ = os.Remove(tmp)
		if st, statErr := os.Stat(outPath); statErr == nil && st.Size() == int64(len(ffmpegBin)) {
			return outPath, nil
		}
		return "", err
	}

	if runtime.GOOS != "windows" {
		_ = os.Chmod(outPath, 0o755)
	}
	return outPath, nil
}

func ExtractFrames(ffmpegPath string, frameRate float64, outputDir string, inputMp4 string) error {
	ffmpegPath = strings.TrimSpace(ffmpegPath)
	outputDir = strings.TrimSpace(outputDir)
	inputMp4 = strings.TrimSpace(inputMp4)

	if ffmpegPath == "" {
		return fmt.Errorf("ffmpegPath is required")
	}
	if inputMp4 == "" {
		return fmt.Errorf("inputMp4 is required")
	}
	if frameRate <= 0 {
		return fmt.Errorf("frameRate must be > 0 (got %v)", frameRate)
	}
	if outputDir == "" {
		return fmt.Errorf("outputDir is required")
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create outputDir: %w", err)
	}

	outPattern := filepath.Join(outputDir, "frame_%06d.png")

	fpsArg := "fps=" + strconv.FormatFloat(frameRate, 'f', -1, 64)

	args := []string{
		"-hide_banner",
		"-loglevel", "error",
		"-i", inputMp4,
		"-vf", fpsArg,
		"-vsync", "vfr",
		"-start_number", "1",
		outPattern,
	}

	cmd := exec.Command(ffmpegPath, args...)

	var stderr bytes.Buffer
	cmd.Stdout = nil
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return fmt.Errorf("ffmpeg failed: %s", msg)
	}

	return nil
}

func CropImages(dimensions CropDimension, inputPath string, cropTest bool) error {
	images, err := os.ReadDir(inputPath)
	if err != nil {
		fmt.Printf("Failed to read dir: %q\n", err)
		return err
	}

	for _, img := range images {
		if img.Type().IsDir() {
			continue
		}

		imagePath := path.Join(inputPath, img.Name())

		file, err := os.Open(imagePath)
		if err != nil {
			fmt.Printf("Failed to open file: %q\n", err)
			return err
		}
		defer file.Close()

		src, _, err := image.Decode(file)
		if err != nil {
			fmt.Printf("Failed to decode file: %q\n", err)
			return err
		}

		cropArea, err := cropRectFor(src, dimensions)
		if err != nil {
			fmt.Println(err)
			return err
		}

		cropped := image.NewRGBA(image.Rect(0, 0, cropArea.Dx(), cropArea.Dy()))
		draw.Draw(cropped, cropped.Bounds(), src, cropArea.Min, draw.Src)

		if cropTest {
			ogExt := filepath.Ext(img.Name())
			ogNameNoExt := img.Name()[:len(img.Name())-len(ogExt)]

			testFileName := fmt.Sprintf("crop_test_%s.jpg", ogNameNoExt)

			outputFP := path.Join(inputPath, "Crop", "Test", testFileName)

			if err := os.MkdirAll(filepath.Dir(outputFP), 0o755); err != nil {
				fmt.Printf("Failed to create output dir: %q\n", err)
				return err
			}
			outFile, err := os.Create(outputFP)
			if err != nil {
				fmt.Printf("Failed to create file: %q\n", err)
				return err
			}
			defer outFile.Close()

			if err := jpeg.Encode(outFile, cropped, nil); err != nil {
				fmt.Printf("Failed to encode file: %q\n", err)
				return err
			}
			return nil
		}

		ogExt := filepath.Ext(img.Name())
		ogNameNoExt := img.Name()[:len(img.Name())-len(ogExt)]

		fileName := fmt.Sprintf("cropped_%s.jpg", ogNameNoExt)
		outputFP := path.Join(inputPath, "Crop", fileName)

		if err := os.MkdirAll(filepath.Dir(outputFP), 0o755); err != nil {
			fmt.Printf("Failed to create output dir: %q\n", err)
			return err
		}
		outFile, err := os.Create(outputFP)
		if err != nil {
			fmt.Printf("Failed to create file: %q\n", err)
			return err
		}
		defer outFile.Close()

		if err := jpeg.Encode(outFile, cropped, nil); err != nil {
			fmt.Printf("Failed to encode file: %q\n", err)
			return err
		}

	}
	return nil
}

func cropRectFor(src image.Image, d CropDimension) (image.Rectangle, error) {
	b := src.Bounds()

	minX := b.Min.X + d.Left
	minY := b.Min.Y + d.Top
	maxX := b.Max.X - d.Right
	maxY := b.Max.Y - d.Bottom

	r := image.Rect(minX, minY, maxX, maxY).Intersect(b)
	if r.Empty() || r.Dx() <= 0 || r.Dy() <= 0 {
		return image.Rectangle{}, fmt.Errorf("invalid crop: src=%v margins=%+v result=%v", b, d, r)
	}
	return r, nil
}
