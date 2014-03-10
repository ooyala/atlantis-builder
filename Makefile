PROJECT_ROOT := $(shell pwd)
ifeq ($(shell pwd | xargs dirname | xargs basename),lib)
	VENDOR_PATH := $(shell pwd | xargs dirname | xargs dirname)/vendor
else
	VENDOR_PATH := $(PROJECT_ROOT)/vendor
endif

ifndef VERSION
	VERSION := "0.2.0"
endif

GOPATH := $(PROJECT_ROOT):$(VENDOR_PATH)
export GOPATH

all: build

build:
	@go build

fmt:
	@find . -name \*.go -exec go fmt {} \;

deb: clean build
	@cp -a deb pkg
	@cp atlantis-builder pkg/usr/bin/
	@sed -ri "s/__VERSION__/$(VERSION)/" pkg/DEBIAN/control 
	@dpkg --build pkg .

clean:
	@rm -f atlantis-builder
	@rm -rf pkg
