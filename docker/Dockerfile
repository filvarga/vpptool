#
FROM ubuntu:20.04 as base
LABEL maintainer="filipvarga89@gmail.com"

ENV TZ=Europe/Bratislava
ENV LC_ALL=C.UTF-8
ENV LANG=C.UTF-8

ARG DEBIAN_FRONTEND=noninteractive
ARG UNATTENDED=y

# libssl-dev required by install-ext-deps
# netbase & ebtables required by scapy
# vpp make build requires python3
RUN apt-get update && \
apt-get install -y --no-install-recommends \
ca-certificates build-essential openssh-client \
git gdb sudo less vim wget \
python3 python3-pip libssl-dev netbase ebtables && \
rm -rf /var/lib/apt/lists/*

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

ENV VPP_BIN="${VPP}/build-root/install-vpp_debug-native/vpp/bin"
ENV STARTUP_CONF=/etc/startup.conf

COPY ./files/bash_aliases $USER_HOME/.bash_aliases

ADD ./files/bin /usr/local/bin/
ADD ./files/etc /etc

ENV PATH="${PATH}:${USER_HOME}/.local/bin"

ARG USER_PSWD=

RUN [ ! -z "$USER_PSWD" ] && (echo 'user:$USER_PSWD' | chpasswd) || \
(echo '%sudo ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers)

RUN chown -R $IDU:$IDG $VPP

USER $IDU:$IDG

ARG GIT_MAIL="john.doe@example.com"
ARG GIT_NAME="John Doe"

RUN git config --global user.email $GIT_MAIL && \
git config --global user.name  $GIT_NAME

RUN pip3 install --user ipaddress scapy 

#
FROM vpptool-images:setup as tool
LABEL maintainer="filipvarga89@gmail.com"

ARG DEBIAN_FRONTEND=noninteractive

WORKDIR /tmp

USER root:root

# install code-server 3.9.2
ARG CS_VERSION=

RUN wget -O code-server_${CS_VERSION}_amd64.deb https://github.com/cdr/code-server/releases/download/v${CS_VERSION}/code-server_${CS_VERSION}_amd64.deb && \
dpkg -i code-server_${CS_VERSION}_amd64.deb && \
rm code-server_${CS_VERSION}_amd64.deb

# install golang 1.16.3
ARG GO_VERSION=

RUN wget -O go${GO_VERSION}.linux-amd64.tar.gz \
https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz && \
tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz && \
rm go${GO_VERSION}.linux-amd64.tar.gz

ENV PATH="${PATH}:/usr/local/go/bin"

# install tools
RUN apt-get update && \
apt-get install -y --no-install-recommends tmux && \
rm -rf /var/lib/apt/lists/*

USER user:user

WORKDIR $USER_HOME

# copy user configs
COPY ./files/settings.json .local/share/code-server/User/settings.json
COPY ./files/tmux.conf .tmux.conf

EXPOSE 9090/tcp

ARG CODE_PSWD=
ENV HASHED_PASSWORD=$CODE_PSWD

ENTRYPOINT ["code-server", "--bind-addr", "0.0.0.0:9090"]
CMD ["--disable-update-check"]
