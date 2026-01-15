package main

import (
	"bufio"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

var ffmpegPath string

type Scanner struct {
	*bufio.Scanner
}

type CropDimension struct {
	Top    int
	Bottom int
	Right  int
	Left   int
}

func (d *CropDimension) String() string {
	b, _ := json.MarshalIndent(d, "", "  ")
	return string(b)
}

type Mp4File struct {
	Path      string
	ParentDir string
	FileName  string
	Extension string
	Contents  []byte
}

func (f *Mp4File) String() string {
	b, _ := json.MarshalIndent(Mp4File{Path: f.Path, ParentDir: f.ParentDir, FileName: f.FileName, Extension: f.Extension}, "", "  ")
	return string(b)
}

// Simple Scans
// ~~~~~~~~~~~~~~~

func (s *Scanner) scanYes() (ok bool) {
	if s.Scan() {
		input := strings.ToLower(s.Text())

		if input == "y" || input == "yes" {
			return true
		}
	}

	return false
}

func (s *Scanner) scanNum(constraint int) (input int, ok bool) {
	if s.Scan() {
		input, err := strconv.Atoi(s.Text())
		if err != nil {
			return 0, false
		}

		if constraint == 0 {
			if input != 0 {
				return input, true
			} else {
				fmt.Print("Oopsie... Please enter a valid number greater than 0\n")
			}
		} else {
			if input < constraint {
				if input != 0 {
					return input, true
				} else {
					fmt.Print("Oopsie... Please enter a valid number greater than 0\n")
				}
			} else {
				fmt.Printf("Sorry... that number exceeds the maximum allowed limit of %d\n", constraint)
				return input, false

			}
		}
	}

	return 0, false
}

func (s *Scanner) scanDir() (dir string, entries int, ok bool) {
	if s.Scan() {
		input := s.Text()

		entries, err := os.ReadDir(input)
		if err != nil {
			return "", 0, false
		}

		return input, len(entries), true
	}

	return "", 0, false
}

func (s *Scanner) scanFile() (file *Mp4File, ok bool) {
	if s.Scan() {
		input := s.Text()

		contents, err := os.ReadFile(input)
		if err != nil {
			return nil, false
		}

		parentDir := filepath.Dir(input)
		fileName := path.Base(input)
		ext := path.Ext(fileName)
		return &Mp4File{ParentDir: parentDir, Path: input, FileName: fileName, Extension: ext, Contents: contents}, true
	}

	return nil, false
}

func (s *Scanner) scanDimensions() (CropDimension, bool) {
	if s.Scan() {
		input := strings.TrimSpace(s.Text())

		d := strings.Split(input, ",")

		if len(d) > 4 || len(d) < 4 {
			return CropDimension{}, false
		}

		var dimensions CropDimension

		for i, letter := range d {
			number, err := strconv.Atoi(strings.TrimSpace(letter))
			if err != nil {
				return CropDimension{}, false
			}

			switch i {
			case 0:
				dimensions.Top = number

			case 1:
				dimensions.Bottom = number

			case 2:
				dimensions.Right = number

			case 3:
				dimensions.Left = number
			}
		}

		return dimensions, true
	}
	return CropDimension{}, false
}

// Scans with prompt for specific items
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (s *Scanner) dirPrompt() (dir string) {
	for {
		fmt.Print("\nWhat file directory are we dealing with today?  ")
		dir, entries, ok := s.scanDir()
		if ok {
			fmt.Printf("> File directory found! Contains %d files\n\n", entries)

			fmt.Print("Is this correct? (y/Yes)  ")
			if ok := s.scanYes(); ok {
				return dir
			}
		} else {
			fmt.Print("Oopsie... I can't find that folder...\n")
		}
	}
}

func (s *Scanner) mp4Prompt() (file *Mp4File) {
	for {
		fmt.Print("\nWhat MP4 file are we working with today?  ")
		file, ok := s.scanFile()
		if ok && file.Extension == ".mp4" {
			fmt.Printf("> %s found!\n\n", file.FileName)

			fmt.Print("Ready to proceed? (y/Yes)  ")
			if ok := s.scanYes(); ok {
				return file
			}
		} else {
			fmt.Print("Oopsie... I can't find that file...\n")
		}
	}
}

func (s *Scanner) frPrompt() (frameRate int) {
	for {
		fmt.Print("\nPlease enter your desired frame rate at which to extract images from the MP4 file:  ")
		frameRate, ok := s.scanNum(200)
		if ok {
			fmt.Printf("> I will extract %d image(s) every second of video time\n\n", frameRate)

			fmt.Print("Ready to proceed? (y/Yes)  ")
			if ok := s.scanYes(); ok {
				return frameRate
			}
		}
	}
}

func (s *Scanner) cropPrompt() CropDimension {
	for {
		fmt.Print("\nPlease select an option based on your cropping preference:")
		fmt.Print("\n[1] Top\n[2] Bottom\n[3] Right\n[4] Left\n[5] All Sides\n> ")

		pref, ok := s.scanNum(0)
		if ok {
			switch pref {

			case 1:
				fmt.Print("\nHow many pixels do you want REMOVED from the Top?  ")
				tSize, ok := s.scanNum(0)
				if ok {
					fmt.Printf("> I will remove %d pixels from the top of each image\n\n", tSize)

					fmt.Print("Ready to proceed? (y/Yes)  ")
					if ok := s.scanYes(); ok {
						return CropDimension{Top: tSize}
					}
				}

			case 2:
				fmt.Print("\nHow many pixels do you want REMOVED from the Bottom?  ")
				bSize, ok := s.scanNum(0)
				if ok {
					fmt.Printf("> I will remove %d pixels from the bottom of each image\n\n", bSize)

					fmt.Print("Ready to proceed? (y/Yes)  ")
					if ok := s.scanYes(); ok {
						return CropDimension{Bottom: bSize}
					}
				}

			case 3:
				fmt.Print("\nHow many pixels do you want REMOVED from the Right?  ")
				rSize, ok := s.scanNum(0)
				if ok {
					fmt.Printf("> I will remove %d pixels from the right side of each image\n\n", rSize)

					fmt.Print("Ready to proceed? (y/Yes)  ")
					if ok := s.scanYes(); ok {
						return CropDimension{Right: rSize}
					}
				}

			case 4:
				fmt.Print("\nHow many pixels do you want REMOVED from the Left?  ")
				lSize, ok := s.scanNum(0)
				if ok {
					fmt.Printf("> I will remove %d pixels from the left side of each image\n\n", lSize)

					fmt.Print("Ready to proceed? (y/Yes)  ")
					if ok := s.scanYes(); ok {
						return CropDimension{Left: lSize}
					}
				}

			case 5:
				fmt.Print("\nPlease enter the cropped dimensions as: Top, Bottom, Right, Left\nThis is the number of pixels you want REMOVED from each side\n(Enter '0' to keep the side unchanged)\n> ")
				dimensions, ok := s.scanDimensions()
				if ok {
					fmt.Printf("\n> I will remove pixels from each image accordingly:\n> Top: %d\n> Bottom: %d\n> Right: %d\n> Left: %d\n\n",
						dimensions.Top, dimensions.Bottom, dimensions.Right, dimensions.Left)

					fmt.Print("Ready to proceed? (y/Yes)  ")
					if ok := s.scanYes(); ok {
						return dimensions
					}
				}

			default:
				fmt.Print("\nSorry... thats not a valid option...\n")
			}
		}
	}
}

func (s *Scanner) runCrop(d CropDimension, inputDir string) (ok bool) {
	dimensions := d

	for {
		fmt.Print("\n> Let's run a quick test to confirm...\n")

		// fmt.Println(dimensions)

		if err := CropImages(dimensions, inputDir, true); err != nil {
			fmt.Print("\nOopsie... someting went wrong with the crop...\n")
			return false
		}

		fmt.Print("\nCheck the image in the /Crop/Test folder\n\n")

		fmt.Print("Is this good to go? (y/Yes)  ")
		if ok := s.scanYes(); ok {
			fmt.Print("\nHold tight... Cropping images...\n")

			if err := CropImages(dimensions, inputDir, false); err != nil {
				fmt.Print("\nOopsie... someting went wrong with the crop...\n")
				return false
			}

			return true
		}
		dimensions = s.cropPrompt()
	}
}

func init() {
	var err error

	ffmpegPath, err = EnsureFFmpeg()
	if err != nil {
		log.Fatal("Could not find FFMpeg Path")
	}
}

func main() {
	scanner := Scanner{bufio.NewScanner(os.Stdin)}

	fmt.Print(`
 __  __     ______     __  __     ______    
/\ \/ /    /\  __ \   /\ \/ /    /\  __ \   
\ \  _"-.  \ \  __ \  \ \  _"-.  \ \ \/\ \  
 \ \_\ \_\  \ \_\ \_\  \ \_\ \_\  \ \_____\ 
  \/_/\/_/   \/_/\/_/   \/_/\/_/   \/_____/ 
                                            
`)

	fmt.Print("\nWelcome to Kako!\n")

	file := scanner.mp4Prompt()

	// Debug
	// fmt.Println(file)

	frameRate := scanner.frPrompt()

	fmt.Print("\nHold tight... Extracting images...\n")

	outputDir := path.Join(file.ParentDir, "Kako")

	if err := ExtractFrames(ffmpegPath, float64(frameRate), outputDir, file.Path); err != nil {
		log.Fatalf("Failed to extract frames from MP4: %q", err)
	}

	images, err := os.ReadDir(outputDir)
	if err != nil {
		fmt.Println("Ooof... I can't find the extracted images... I must... have... failed...?")
	}

	fmt.Printf("> Success! %d images have been extracted!\n\n", len(images))

	fmt.Print("Do you also need these images cropped? (y/Yes)  ")

	if ok := scanner.scanYes(); !ok {
		fmt.Print("\nOkie dokie! See you later!\n\n")
	}

	dimensions := scanner.cropPrompt()

	if ok := scanner.runCrop(dimensions, outputDir); !ok {
		return
	}

	cropImages, err := os.ReadDir(outputDir + "/Crop")
	if err != nil {
		fmt.Println("Ooof... I can't find the extracted images... I must... have... failed...?")
	}

	var cropCount int
	for _, img := range cropImages {
		if img.Type().IsDir() {
			continue
		}
		cropCount++
	}

	fmt.Printf("> Success! %d images have been cropped!\n\n", cropCount)

	fmt.Println("Thank you for using Kako, see you next time!")
}
