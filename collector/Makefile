
PROJECT_NAME := collector
BUILD_NUMBER := latest
DOCKER_REGISTRY?=aporeto
DOCKER_IMAGE_NAME?=$(PROJECT_NAME)
DOCKER_IMAGE_TAG?=$(BUILD_NUMBER)

build:
	glide install
	env GOOS=linux GOARCH=386 go build

install: build
	sudo cp collector docker/collector

package:
	cp collector docker/collector

clean:
	rm -rf vendor
	rm -rf docker/collector
