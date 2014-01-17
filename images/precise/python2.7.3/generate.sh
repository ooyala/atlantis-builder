#!/bin/bash

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

sudo docker tag $ID $REGISTRY_HOST/builder/precise64-python2.7.3-$VERSION

if [[ "$1" == "-p" ]]; then
  echo "Pushing..."
  sudo docker push $REGISTRY_HOST/builder/precise64-python2.7.3-$VERSION
  echo "Done."
else
  echo "Done. To push the new image:"
  echo "sudo docker push $REGISTRY_HOST/builder/precise64-python2.7.3-$VERSION"
fi
