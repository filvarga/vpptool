FROM vpptool-images:setup

LABEL maintainer="filipvarga89@gmail.com"

ARG DEBIAN_FRONTEND=noninteractive

RUN git clone https://github.com/VundleVim/Vundle.vim.git \
$USER_HOME/.vim/bundle/Vundle.vim

# install ohmyzsh
RUN sh -c "$(wget -O- https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)"

# copy user configs
COPY ./conf/user/tmux.conf $USER_HOME/.tmux.conf
COPY ./conf/user/vimrc $USER_HOME/.vimrc

# TODO: run as user
RUN su -c "vim +PluginInstall +qall" user

RUN apt-get update && \
apt-get install -y --no-install-recommends wget zsh && \
rm -rf /var/lib/apt/lists/*

WORKDIR /tmp

# install golang
ARG GO_VERSION=1.16.3

RUN wget -O go${GO_VERSION}.linux-amd64.tar.gz \
https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz && \
rm go${GO_VERSION}.linux-amd64.tar.gz

ENV PATH="${PATH}:/usr/local/go/bin"

# install code-server
ARG CS_VERSION=3.9.2

RUN wget -O code-server.deb https://github.com/cdr/code-server/releases/download/v${CS_VERSION}/code-server_${CS_VERSION}_amd64.deb

RUN dpkg -i code-server.deb && rm code-server.deb

RUN mkdir -p $USER_HOME/.local/share/code-server/User/

COPY ./conf/code/settings.json $USER_HOME/.local/share/code-server/User/settings.json

ARG SUDO_PASS=
ARG CODE_PASS=toor

RUN [[ ! -z "$SUDO_PASS" ]] && (echo 'user:user' | chpasswd) || \
echo '%sudo ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers

#RUN chown -R $IDU:$IDG $USER_HOME
#USER $IDU:$IDG

EXPOSE 9090/tcp

# TODO: convert to sha25 or what it wants, as an environment variable
ENV PASSWORD=${CODE_PASS}
ENTRYPOINT ["code-server", "--bind-addr", "0.0.0.0:9090"]
CMD ["--disable-update-check"]
