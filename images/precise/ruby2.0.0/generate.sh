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


set -e

VERSION="0.1.1"

if [[ "$1" == "" ]]; then
  echo "Please specify a configuration script"
  exit 1
fi
source $1
if [[ "$MAINTAINER" == "" ]]; then
  echo "Please specify a MAINTAINER in your configuration script"
  exit 1
fi
if [[ "$REGISTRY_HOST" == "" ]]; then
  echo "Please specify a REGISTRY_HOST in your configuration script"
  exit 1
fi

sed -e s#__REGISTRY_HOST__#$REGISTRY_HOST#g -e s#__MAINTAINER__#$MAINTAINER#g Dockerfile.template >Dockerfile

echo "Building..."
(
  set -o pipefail
  sudo docker build . 2>&1 | tee docker_build.log
  if tail -2 docker_build.log | grep -q Error; then
    exit 1
  fi
)

line=`tail -1 docker_build.log`
if [[ ! $line =~ [^[:space:]] ]] ; then
	line=`tail -2 docker_build.log | head -1`
fi
ID=`echo $line | awk '{print $3;}'`

sudo docker tag $ID $REGISTRY_HOST/builder/precise64-ruby2.0.0-$VERSION

if [[ "$1" == "-p" ]]; then
  echo "Pushing..."
  sudo docker push $REGISTRY_HOST/builder/precise64-ruby2.0.0-$VERSION
  echo "Done."
else
  echo "Done. To push the new image:"
  echo "sudo docker push $REGISTRY_HOST/builder/precise64-ruby2.0.0-$VERSION"
fi
