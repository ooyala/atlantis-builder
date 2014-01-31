## Copyright 2014 Ooyala, Inc. All rights reserved.
##
## This file is licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
## except in compliance with the License. You may obtain a copy of the License at
## http://www.apache.org/licenses/LICENSE-2.0
##
## Unless required by applicable law or agreed to in writing, software distributed under the License is
## distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
## See the License for the specific language governing permissions and limitations under the License.

# ruby 2.0.0 Builder Template
#
# VERSION 0.1.1

FROM __REGISTRY_HOST__/builder/__IMAGE__-ruby2.0.0-0.1.1
MAINTAINER __MAINTAINER__

# add app source
RUN mkdir /tmp/build
ADD build /tmp/build/
RUN chown -R root:root /tmp/build
# run the setup command (if it exists) to update the box
RUN cd /tmp/build/app && __SETUP_COMMAND__
# copy build info
RUN mkdir -p /etc/atlantis && cp -a /tmp/build/info /etc/atlantis/build && chown -R user1:user1 /etc/atlantis
# package app
RUN bundle install --gemfile /tmp/build/app/Gemfile
# copy packaged app to /opt
RUN mkdir -p /opt
RUN mv /tmp/build/app /opt/app
RUN chown -R user1:user1 /opt/app
# copy runit stuff
RUN for dir in `find /tmp/build -name "runit*"`; do runitname=$(basename $dir); appdir=$(echo $runitname | sed 's/runit/app/'); mv $dir /etc/sv/$appdir; ln -s /etc/sv/$appdir /etc/service/$appdir; done
# cleanup
RUN rm -rf /tmp/build
ADD etc /tmp/etc
RUN mv /tmp/etc/sv/rsyslog/run /etc/sv/rsyslog/
RUN mv /tmp/etc/rsyslog.d/* /etc/rsyslog.d/
# set default run command
CMD runsvdir /etc/service
