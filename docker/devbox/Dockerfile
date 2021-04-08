FROM vpptool-images:setup

LABEL maintainer="filipvarga89@gmail.com"

ARG DEBIAN_FRONTEND=noninteractive

USER root:root

WORKDIR /tmp

RUN apt-get update && \
apt-get install -y --no-install-recommends wget zsh && \
rm -rf /var/lib/apt/lists/*

# install code-server
ARG CS_VERSION=3.9.2

RUN wget -O code-server_${CS_VERSION}_amd64.deb https://github.com/cdr/code-server/releases/download/v${CS_VERSION}/code-server_${CS_VERSION}_amd64.deb && \
dpkg -i code-server_${CS_VERSION}_amd64.deb && \
rm code-server_${CS_VERSION}_amd64.deb

# install golang
ARG GO_VERSION=1.16.3

RUN wget -O go${GO_VERSION}.linux-amd64.tar.gz \
https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz && \
tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz && \
rm go${GO_VERSION}.linux-amd64.tar.gz

ENV PATH="${PATH}:/usr/local/go/bin"

USER user:user

WORKDIR $USER_HOME

# copy user configs
COPY ./files/tmux.conf $USER_HOME/.tmux.conf
COPY ./files/vimrc $USER_HOME/.vimrc

RUN git clone https://github.com/VundleVim/Vundle.vim.git \
$USER_HOME/.vim/bundle/Vundle.vim

RUN mkdir -p $USER_HOME/.local/share/code-server/User/
COPY ./files/settings.json $USER_HOME/.local/share/code-server/User/settings.json

# install ohmyzsh
RUN sh -c "$(wget -O- https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)"

# TODO: fix the password for sudo
#ARG SUDO_PASS=
#RUN [[ ! -z "$SUDO_PASS" ]] && (echo 'user:user' | chpasswd) || \
#echo '%sudo ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers

# TODO: run as user
#RUN vim +PluginInstall +qall

EXPOSE 9090/tcp

ARG CODE_PASS=toor
# TODO: convert to sha25 or what it wants, as an environment variable
ENV PASSWORD=${CODE_PASS}
ENTRYPOINT ["code-server", "--bind-addr", "0.0.0.0:9090"]
CMD ["--disable-update-check"]
