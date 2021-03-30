FROM ubuntu:20.04

LABEL maintainer="filipvarga89@gmail.com"

ENV TZ=Europe/Bratislava
ENV LC_ALL=C.UTF-8
ENV LANG=C.UTF-8

ARG DEBIAN_FRONTEND=noninteractive

# docker build requirements
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    git wget sudo zsh ca-certificates openssh-client python3 python3-pip && \
    rm -rf /var/lib/apt/lists/*

ARG HOME=/opt/user

ARG IDU=1000
ARG IDG=1000

RUN groupadd --gid $IDG user; \
    useradd --uid $IDU --gid $IDG \
            --home-dir $HOME \
            --shell /usr/bin/zsh \
            --create-home \
            --groups sudo user

RUN mkdir $HOME/git

ARG GIT_MAIL="john.doe@example.com"
ARG GIT_NAME="John Doe"

RUN git config --global user.email $GIT_MAIL && \
    git config --global user.name  $GIT_NAME

# install user tools
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    gdb vim less tmux build-essential netbase ebtables && \
    rm -rf /var/lib/apt/lists/*

RUN git clone https://github.com/VundleVim/Vundle.vim.git \
  /opt/user/.vim/bundle/Vundle.vim

# install ohmyzsh
RUN sh -c "$(wget -O- https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)"

# copy user configs
COPY ./conf/user/tmux.conf $HOME/.tmux.conf
COPY ./conf/user/vimrc $HOME/.vimrc

RUN su -c "vim +PluginInstall +qall" user

# install golang
ARG GO_VERSION=1.16.3

RUN wget -O go${GO_VERSION}.linux-amd64.tar.gz \
  https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz && \
  rm go${GO_VERSION}.linux-amd64.tar.gz

ENV PATH="${PATH}:/usr/local/go/bin"

# install code-server
ARG CS_VERSION=3.9.2

WORKDIR /tmp

RUN wget -O code-server.deb https://github.com/cdr/code-server/releases/download/v${CS_VERSION}/code-server_${CS_VERSION}_amd64.deb

RUN dpkg -i code-server.deb && rm code-server.deb

RUN mkdir -p $HOME/.local/share/code-server/User/

COPY ./conf/code/settings.json $HOME/.local/share/code-server/User/settings.json

WORKDIR ${HOME}/git

ENV STARTUP_CONF=/etc/startup.conf
ENV VPP="${HOME}/git/vpp"
ENV PATH="${VPP}/build-root/install-vpp_debug-native/vpp/bin:${PATH}"

ADD ./conf/etc /etc/

RUN mkdir -p /run/vpp && \
    mkdir -p /var/log/vpp

ADD ./scripts ${HOME}/bin

ENV PATH="${HOME}/bin:${PATH}"

ARG SUDO_PASS=
ARG CODE_PASS=toor

RUN [[ ! -z "$SUDO_PASS" ]] && (echo 'user:user' | chpasswd) || \
    echo '%sudo ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers

RUN chown -R $IDU:$IDG $HOME

USER $IDU:$IDG

EXPOSE 9090/tcp

# TODO: convert to sha25 or what it wants, as an environment variable
ENV PASSWORD=${CODE_PASS}
ENTRYPOINT ["code-server", "--bind-addr", "0.0.0.0:9090"]
CMD ["--disable-update-check"]
