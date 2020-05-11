FROM ubuntu:18.04

CMD ["/bin/bash"]

# INFO: missing dependencie in make install-dep !!!
# required by install-ext-deps
RUN set -eux; \
    apt-get update; \
    apt-get install -y --no-install-recommends \
    git gdb sudo vim less \
    ca-certificates \
    build-essential \
    openssh-client \
    libssl-dev; \
    rm -rf /var/lib/apt/lists/*

RUN set -eux; \
  git config --global user.email "john.doe@example.com"; \
  git config --global user.name "John Doe"

ARG IDU=1000
ARG IDG=1000

RUN set -eux; \
    groupadd -g $IDG user; \
    useradd -u $IDU -g $IDG user -m; \
    usermod -aG sudo user; \
    echo 'user:user' | chpasswd

WORKDIR /opt

ARG GIT=https://github.com/FDio/vpp.git

RUN set -eux; \
    git clone $GIT

WORKDIR /opt/vpp

ARG BIN=build-root/install-vpp_debug-native/vpp/bin

RUN set -eux; \
    ln -s $(pwd)/$BIN/vppctl /usr/bin/vppctl; \
    ln -s $(pwd)/$BIN/vpp /usr/bin/vpp

RUN set -eux; \
    mkdir -p /run/vpp; \
    mkdir -p /var/log/vpp

RUN set -eux; \
    UNATTENDED=y make install-dep

COPY ./conf/startup.conf /etc/startup.conf
COPY ./conf/vpp.conf     /etc/vpp.conf

ENV STARTUP_CONF=/etc/startup.conf
ENV LC_ALL=C.UTF-8
ENV LANG=C.UTF-8

COPY ./scripts/kill  /scripts/kill
COPY ./scripts/start /scripts/start
COPY ./scripts/build /scripts/build
COPY ./scripts/cache /scripts/cache
