## Copyright 2014 Ooyala, Inc. All rights reserved.
##
## This file is licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
## except in compliance with the License. You may obtain a copy of the License at
## http://www.apache.org/licenses/LICENSE-2.0
##
## Unless required by applicable law or agreed to in writing, software distributed under the License is
## distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
## See the License for the specific language governing permissions and limitations under the License.

PROJECT_ROOT := $(shell pwd)
ifeq ($(shell pwd | xargs dirname | xargs basename),lib)
	VENDOR_PATH := $(shell pwd | xargs dirname | xargs dirname)/vendor
	ATLANTIS_PATH := $(shell pwd | xargs dirname | xargs dirname)/lib/atlantis
else
	VENDOR_PATH := $(PROJECT_ROOT)/vendor
	ATLANTIS_PATH := $(PROJECT_ROOT)/lib/atlantis
endif

ifndef BASENAME
	BASENAME := "precise64"
endif

ifndef VERSION
	VERSION := "0.1.0"
endif

GOPATH := $(PROJECT_ROOT):$(VENDOR_PATH):$(ATLANTIS_PATH)
export GOPATH

all: build

build:
	@go build -o atlantis-builder builder.go
	@go build -o atlantis-builderd builderd.go

deb: clean build
	@cp -a deb pkg
	@mkdir -p pkg/opt/atlantis/bin
	@mkdir -p pkg/opt/atlantis/builder

	@cp atlantis-mkbase pkg/opt/atlantis/bin/
	@cp atlantis-builder pkg/opt/atlantis/bin/

	@cp -a layers pkg/opt/atlantis/builder/
	@echo $(BASENAME) > pkg/opt/atlantis/builder/layers/basename.txt
	@echo $(VERSION) > pkg/opt/atlantis/builder/layers/version.txt

	@sed -ri "s/__VERSION__/$(VERSION)/" pkg/DEBIAN/control 
	@dpkg -b pkg .

fmt:
	@find . -path ./vendor -prune -o -name \*.go -exec go fmt {} \;

clean:
	@rm -rf atlantis-builder pkg *.deb
