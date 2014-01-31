#!/bin/bash
## Copyright 2014 Ooyala, Inc. All rights reserved.
##
## This file is licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
## except in compliance with the License. You may obtain a copy of the License at
## http://www.apache.org/licenses/LICENSE-2.0
##
## Unless required by applicable law or agreed to in writing, software distributed under the License is
## distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
## See the License for the specific language governing permissions and limitations under the License.


CONFIG_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

REGISTRY_HOST="<insert registry host here>"
BASE_IMAGE="ubuntu:precise"
MAINTAINER="yourteam@yourcompany.com"
MASTER_SSH_PUBLIC_KEY="<path to id_rsa.pub>"
BASE_PROVISION_EXTRA=""
