# ruby 1.9.3 Builder Template
#
# VERSION 0.1.0

FROM __REGISTRY_HOST__/builder/__IMAGE__-ruby1.9.3-0.1.0
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
# move the check_mk_checks directory from the app to root
RUN mkdir -p /tmp/build/app/check_mk_checks
RUN mv /tmp/build/app/check_mk_checks/ /
# move log directories to /var/log/
RUN mv /tmp/build/log/* /var/log/
# copy runit stuff
RUN for dir in `find /tmp/build -name "runit*"`; do runitname=$(basename $dir); appdir=$(echo $runitname | sed 's/runit/app/'); mv $dir /etc/sv/$appdir; ln -s /etc/sv/$appdir /etc/service/$appdir; mkdir -p /var/log/$appdir; chown -R user1:user1 /var/log/$appdir; done
# cleanup
RUN rm -rf /tmp/build
ADD etc /tmp/etc
RUN mv /tmp/etc/sv/rsyslog/run /etc/sv/rsyslog/
RUN mv /tmp/etc/rsyslog.d/* /etc/rsyslog.d/
# set default run command
CMD runsvdir /etc/service
