#!/bin/bash

CONFIG_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

REGISTRY_HOST="<insert registry host here>"
BASE_IMAGE="ubuntu:precise"
MAINTAINER="yourteam@yourcompany.com"
MASTER_SSH_PUBLIC_KEY="<path to id_rsa.pub>"
BASE_PROVISION_EXTRA=""
