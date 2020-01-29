# Copyright 2019 Cruise LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#    https://www.apache.org/licenses/LICENSE-2.0
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
.PHONY: all build test image clean

LDFLAGS = -extldflags=-static -s -w
BUILD_FLAGS = -mod=readonly -ldflags '$(LDFLAGS)' -trimpath
BUILD_VERSION ?= manual
IMAGE_NAME = "cruise/k-rail:${BUILD_VERSION}"

all: dist

dist: image

clean:
	rm -f evicter k-rail
	go mod verify

build:
		GO111MODULE=on CGO_ENABLED=0 go build ${BUILD_FLAGS} -o k-rail cmd/k-rail/main.go
		GO111MODULE=on CGO_ENABLED=0 go build ${BUILD_FLAGS} -o evicter cmd/evicter/*.go

test:
		GO111MODULE=on CGO_ENABLED=1 go test -mod=readonly -race  ./...

image: build
	docker build --pull -t $(IMAGE_NAME) .