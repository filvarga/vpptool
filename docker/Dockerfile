FROM ubuntu:18.04

# general
RUN set -eux; \
    apt-get update; \
    apt-get install -y --no-install-recommends \
    linux-tools-`uname -r` \
    git gdb sudo vim less \
    inetutils-traceroute \
    linux-tools-common \
    ca-certificates \
    build-essential \
    iputils-ping \
    net-tools \
    iproute2 \
    valgrind \
    tcpdump \
    netcat; \
    rm -rf /var/lib/apt/lists/*; \
    mv /usr/sbin/tcpdump /usr/bin/tcpdump

ARG IDU=1000
ARG IDG=1000

RUN set -eux; \
    groupadd -g $IDG user; \
    useradd -u $IDU -g $IDG user -m; \
    usermod -aG sudo user; \
    echo 'user:user' | chpasswd

WORKDIR /opt

RUN set -eux; \
    git clone https://github.com/FDio/vpp.git

WORKDIR /opt/vpp

ARG BIN=build-root/install-vpp_debug-native/vpp/bin

RUN set -eux; \
    ln -s $(pwd)/$BIN/vppctl /usr/bin/vppctl; \
    ln -s $(pwd)/$BIN/vpp /usr/bin/vpp

RUN set -eux; \
    mkdir -p /run/vpp; \
    mkdir -p /var/log/vpp

COPY ./files/startup.sh /usr/bin/startup.sh
COPY ./files/setup.sh /usr/bin/setup.sh
COPY ./files/kill.sh /usr/bin/kill.sh

COPY ./files/startup.conf /etc/startup.conf
COPY ./files/vpp.conf /etc/vpp.conf

RUN set -eux; \
  git config --global user.email "you@example.com"; \
  git config --global user.name "Your Name"

ENV STARTUP_CONF=/etc/startup.conf
ENV LC_ALL=C.UTF-8
ENV LANG=C.UTF-8

CMD ["setup.sh"]
