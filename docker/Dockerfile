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
    openssh-client \
    iputils-ping \
    net-tools \
    iproute2 \
    valgrind \
    tcpdump netcat; \
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

COPY ./scripts/startup.sh /scripts/startup
COPY ./scripts/setup.sh   /scripts/setup
COPY ./scripts/kill.sh    /scripts/kill

COPY ./conf/startup.conf /etc/startup.conf
COPY ./conf/vpp.conf     /etc/vpp.conf

RUN set -eux; \
  git config --global user.email "john.doe@example.com"; \
  git config --global user.name "John Doe"

ENV STARTUP_CONF=/etc/startup.conf
ENV LC_ALL=C.UTF-8
ENV LANG=C.UTF-8

CMD ["/scripts/setup"]
