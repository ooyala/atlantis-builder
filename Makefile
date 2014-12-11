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

build-builder:
	@go build -o atlantis-builder builder.go

build-builderd:
	@go build -o atlantis-builderd builderd.go

build: build-builder build-builderd

DEB_STAGING := $(PROJECT_ROOT)/staging
PKG_INSTALL_DIR := $(DEB_STAGING)/opt/atlantis
PKG_BIN_DIR := $(PKG_INSTALL_DIR)/opt/atlantis/bin

deb-builder: clean-builder build-builder
	@cp -a $(PROJECT_ROOT)/deb $(DEB_STAGING)
	@mkdir -p $(PKG_BIN_DIR)
	@mkdir -p $(PKG_INSTALL_DIR)/builder

	@cp atlantis-mkbase $(PKG_BIN_DIR)
	@cp atlantis-builder $(PKG_BIN_DIR)

	@cp -a layers $(PKG_INSTALL_DIR)/builder/
	@echo $(BASENAME) > $(PKG_INSTALL_DIR)/builder/layers/basename.txt
	@echo $(VERSION) > $(PKG_INSTALL_DIR)/builder/layers/version.txt

	@sed -ri "s/__VERSION__/$(VERSION)/" $(DEB_STAGING)/DEBIAN/control
	@sed -ri "s/__PACKAGE__/atlantis-builder/" $(DEB_STAGING)/DEBIAN/control
	@dpkg -b $(DEB_STAGING) .

deb-builderd: clean-builderd build-builderd
	@cp -a $(PROJECT_ROOT)/deb $(DEB_STAGING)
	@rm $(DEB_STAGING)/DEBIAN/postinst $(DEB_STAGING)/DEBIAN/postrm
	@rm -rf $(DEB_STAGING)/usr
	@mkdir -p $(PKG_BIN_DIR)

	@cp atlantis-builderd $(PKG_BIN_DIR)

	@sed -ri "s/__VERSION__/$(VERSION)/" $(DEB_STAGING)/DEBIAN/control
	@sed -ri "s/__PACKAGE__/atlantis-builderd/" $(DEB_STAGING)/DEBIAN/control
	@dpkg -b $(DEB_STAGING) .

deb: deb-builder deb-builderd

fmt:
	@find . -path ./vendor -prune -o -name \*.go -exec go fmt {} \;

clean:
cleanall: clean-builder clean-builderd

.PHONY: clean-builder
clean-builder:
	@rm -rf atlantis-builder $(DEB_STAGING) pkg atlantis-builder_*.deb

.PHONY: clean-builderd
clean-builderd:
	@rm -rf atlantis-builderd $(DEB_STAGING) pkg atlantis-builderd_*.deb
