export GOOS = linux
export GOARCH = amd64

build:
	cd cmd && go build -o capper && mv capper ../capper

container:
	podman build -t capper .