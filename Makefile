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
	ATLANTIS_PATH := $(shell pwd | xargs dirname | xargs dirname)/vendor/src/atlantis
else
	VENDOR_PATH := $(PROJECT_ROOT)/vendor
	ATLANTIS_PATH := $(PROJECT_ROOT)/vendor/src/atlantis
endif

ifndef BASENAME
	BASENAME := "precise64"
endif

ifndef VERSION
	VERSION := "0.1.0"
endif

GOPATH := $(PROJECT_ROOT):$(VENDOR_PATH):$(ATLANTIS_PATH)
export GOPATH

GOM := $(VENDOR_PATH)/bin/gom
GOM_VENDOR_NAME := vendor
export GOM_VENDOR_NAME


all: build

$(VENDOR_PATH):
	@echo "Installing Dependencies..."
	@mkdir -p $(VENDOR_PATH) || exit 2
	@GOPATH=$(VENDOR_PATH) go get github.com/ghao-ooyala/gom
	$(GOM) install
	@echo "Done."

build-builder: $(VENDOR_PATH)
	@go build -o atlantis-builder builder.go

build-builderd: $(VENDOR_PATH)
	@go build -o atlantis-builderd builderd.go

build: build-builder build-builderd

DEB_STAGING := $(PROJECT_ROOT)/staging
BUILDER_DIR := $(DEB_STAGING)/opt/atlantis/builder
PKG_BIN_DIR := $(BUILDER_DIR)/bin

deb-builder: clean-builder build-builder
	@mkdir -p $(DEB_STAGING)/DEBIAN
	@mkdir -p $(PKG_BIN_DIR)

	@cp -a $(PROJECT_ROOT)/deb/DEBIAN/control $(DEB_STAGING)/DEBIAN/control
	@cp atlantis-builder $(PKG_BIN_DIR)

	@sed -ri "s/__VERSION__/$(VERSION)/" $(DEB_STAGING)/DEBIAN/control
	@sed -ri "s/__PACKAGE__/atlantis-builder/" $(DEB_STAGING)/DEBIAN/control
	@dpkg -b $(DEB_STAGING) .

deb-builderd: clean-builderd build-builderd
	@mkdir -p $(DEB_STAGING)/DEBIAN
	@cp -a $(PROJECT_ROOT)/deb/* $(DEB_STAGING)
	@mkdir -p $(PKG_BIN_DIR) $(BUILDER_DIR)

	@rm -f $(PKG_BIN_DIR)/atlantis-builder
	@cp atlantis-mkbase $(PKG_BIN_DIR)
	@cp atlantis-builderd $(PKG_BIN_DIR)

	@cp -a layers $(BUILDER_DIR)
	@echo $(BASENAME) > $(BUILDER_DIR)/layers/basename.txt
	@echo $(VERSION) > $(BUILDER_DIR)/layers/version.txt

	@sed -ri "s/__VERSION__/$(VERSION)/" $(DEB_STAGING)/DEBIAN/control
	@sed -ri "s/__PACKAGE__/atlantis-builderd/" $(DEB_STAGING)/DEBIAN/control
	@dpkg -b $(DEB_STAGING) .

deb: clean deb-builder deb-builderd

fmt:
	@find . -path ./vendor -prune -o -name \*.go -exec go fmt {} \;

clean: clean-builder clean-builderd clean-dependencies

.PHONY: clean-builder
clean-builder:
	@rm -rf atlantis-builder $(DEB_STAGING) pkg atlantis-builder_*.deb

.PHONY: clean-builderd
clean-builderd:
	@rm -rf atlantis-builderd $(DEB_STAGING) pkg atlantis-builderd_*.deb

clean-dependencies:
	@rm -rf $(VENDOR_PATH)
