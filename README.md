# Kako

This is a simple application that extracts image frames from an MP4 video file. The user also has the option to crop the extracted images.

## Usage

- Clone the repository
```bash
git clone https://github.com/AnthonyBliss1/Kako.git
```

- Ensure dependencies are installed
```bash
cd Kako && go mod tidy
```

- Make the `ffmpeg.sh` script executable
``` bash
chmod +x ffmpeg.sh
```

- Run the script
``` bash
./ffmpeg.sh
```

- Create the binaries (if you wish to have an executable)
```bash
make build
```

- Run the executable (based on your OS)
```bash
cd builds/linux && ./kako
```
