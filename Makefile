export GOOS = linux
export GOARCH = amd64

build:
	cd cmd && go build -o capper

container:
	docker build -t capper .