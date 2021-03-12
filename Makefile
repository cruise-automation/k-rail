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

default: build

ensure:
		dep ensure

build:
		protoc -I plugins/proto/ plugins/proto/plugin.proto --go_out=plugins=grpc:plugins/proto
		CGO_ENABLED=0 go build -o k-rail cmd/main.go
		CGO_ENABLED=0 go build -o plugin plugins/examples/example.go

test:
		CGO_ENABLED=1 go test -race -cover $(shell go list ./... | grep -v /vendor/)
