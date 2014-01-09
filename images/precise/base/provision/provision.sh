#!/bin/bash

set -ex

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# make sure everything is owned by root
chown -R root:root ${SCRIPT_DIR}

cd ${SCRIPT_DIR}

# Add apt sources
echo "deb http://archive.ubuntu.com/ubuntu precise main universe" > /etc/apt/sources.list
echo "deb http://archive.ubuntu.com/ubuntu/ precise-updates main restricted" > /etc/apt/sources.list.d/precise-update.list
apt-get update

# Install runit (it will fail, that is why the || true is there)
apt-get install -y runit || true
# Fix runit
rm /var/lib/dpkg/info/runit.post*
apt-get -f install

# Install rebuild_authorized_keys
mv sbin/* /usr/local/sbin/
rm -rf sbin

mv etc/sv/* /etc/sv/
mv etc/atlantis /etc

# Add SSH under runit
apt-get install -y openssh-server
mkdir -p /root/.ssh
chmod 700 /root/.ssh
mv ssh/* /root/.ssh/
rm -rf ssh
chmod 600 /root/.ssh/authorized_keys
ln -s /etc/sv/sshd /etc/service/

# Add rsyslog under runit
apt-get install -y rsyslog
mv etc/rsyslog.conf /etc
mkdir /etc/sv/rsyslog
ln -s /etc/sv/rsyslog /etc/service

# Add convenience packages
apt-get -y install curl man-db telnet wget screen tmux tree less strace traceroute ngrep tcpdump

rm -rf etc

# Add app user
useradd -d /home/user1 -m -s /bin/bash user1

# run provision.extra.sh if it exists
if [ -x "./provision.extra.sh" ]; then
  ./provision.extra.sh
fi

echo "provisioning base done."
