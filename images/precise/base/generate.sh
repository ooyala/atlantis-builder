#!/bin/bash

set -e

# This is the version with which to tag the image, rev each new build
VERSION="0.1.1"

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

if [[ "$1" == "" ]]; then
  echo "Please specify a configuration script"
  exit 1
fi
source $1
if [[ "$BASE_IMAGE" == "" ]]; then
  echo "Please specify a BASE_IMAGE in your configuration script"
  exit 1
fi
if [[ "$MAINTAINER" == "" ]]; then
  echo "Please specify a MAINTAINER in your configuration script"
  exit 1
fi
if [[ "$REGISTRY_HOST" == "" ]]; then
  echo "Please specify a REGISTRY_HOST in your configuration script"
  exit 1
fi

sed -e s#__BASE_IMAGE__#$BASE_IMAGE#g -e s#__MAINTAINER__#$MAINTAINER#g Dockerfile.template >Dockerfile

if [[ "$BASE_PROVISION_EXTRA" != "" ]]; then
  cp $BASE_PROVISION_EXTRA $SCRIPT_DIR/provision/provision.extra.sh
  chmod +x $SCRIPT_DIR/provision/provision.extra.sh
fi

if [[ "$MASTER_SSH_PUBLIC_KEY" != "" ]]; then
  mkdir -p $SCRIPT_DIR/provision/ssh/authorized_keys.d
  cp $MASTER_SSH_PUBLIC_KEY $SCRIPT_DIR/provision/ssh/authorized_keys.d/master.pub
fi

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

sudo docker tag $ID $REGISTRY_HOST/base/precise64-$VERSION

if [[ "$2" == "-p" ]]; then
  echo "Pushing..."
  sudo docker push $REGISTRY_HOST/base/precise64-$VERSION
  echo "Done."
else
  echo "Done. To push the new image:"
  echo "sudo docker push $REGISTRY_HOST/base/precise64-$VERSION"
fi
