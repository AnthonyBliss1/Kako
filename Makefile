APP_EXECUTABLE=kako


build:
	GOARCH=arm64 GOOS=darwin go build -o ./builds/darwin/${APP_EXECUTABLE} .
	GOARCH=amd64 GOOS=linux go build -o ./builds/linux/${APP_EXECUTABLE} .
	GOARCH=amd64 GOOS=windows go build -o ./builds/windows/${APP_EXECUTABLE}.exe .

clean:
	go clean
	rm -rf ./builds
