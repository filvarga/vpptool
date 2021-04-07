FROM ubuntu:20.04

LABEL maintainer="filipvarga89@gmail.com"

ENV TZ=Europe/Bratislava
ENV LC_ALL=C.UTF-8
ENV LANG=C.UTF-8

ARG DEBIAN_FRONTEND=noninteractive
ARG UNATTENDED=y

# install-ext-deps requires libssl-dev
# scpay requires additional stuff to work (netbase, ebtables)
# python3 requirement for make
RUN apt-get update && \
apt-get install -y --no-install-recommends \
ca-certificates build-essential openssh-client git gdb sudo less vim \
python3 python3-pip libssl-dev netbase ebtables && \
rm -rf /var/lib/apt/lists/*

RUN echo '%sudo ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers

ENV USER_HOME=/work
ENV VPP="${USER_HOME}/vpp"

ARG IDU=1000
ARG IDG=1000

RUN groupadd --gid $IDG user && useradd --uid $IDU --gid $IDG \
--home-dir $USER_HOME --create-home --groups sudo user

WORKDIR $USER_HOME

RUN mkdir -p /run/vpp && mkdir -p /var/log/vpp

ARG GIT_URL=https://github.com/FDio/vpp.git

RUN git clone $GIT_URL

WORKDIR $VPP

RUN make install-dep

ENV VPP_BIN="${VPP}/build-root/install-vpp_debug-native/vpp/bin"
ENV STARTUP_CONF=/etc/startup.conf

COPY ./files/bash_aliases $USER_HOME/.bash_aliases

ADD ./files/bin /usr/local/bin/
ADD ./files/etc /etc

ENV PATH="${PATH}:${USER_HOME}/.local/bin"

RUN chown -R $IDU:$IDG $VPP

USER $IDU:$IDG

ARG GIT_MAIL="john.doe@example.com"
ARG GIT_NAME="John Doe"

RUN git config --global user.email $GIT_MAIL && \
git config --global user.name  $GIT_NAME

RUN pip3 install --user ipaddress scapy 
