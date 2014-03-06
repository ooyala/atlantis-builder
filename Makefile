PROJECT_ROOT := $(shell pwd)
ifeq ($(shell pwd | xargs dirname | xargs basename),lib)
	VENDOR_PATH := $(shell pwd | xargs dirname | xargs dirname)/vendor
else
	VENDOR_PATH := $(PROJECT_ROOT)/vendor
endif

GOPATH := $(PROJECT_ROOT):$(VENDOR_PATH)
export GOPATH

all:
	@go build

fmt:
	@find . -name \*.go -exec go fmt {} \;

clean:
	@rm -f atlantis-builder
