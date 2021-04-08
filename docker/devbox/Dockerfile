FROM vpptool-images:setup

LABEL maintainer="filipvarga89@gmail.com"

ARG DEBIAN_FRONTEND=noninteractive

WORKDIR /tmp

USER root:root

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

# =$(echo -n $INPUT | sha256sum | cut -c -64 -z)
ARG CODE_PASS=ce5ca673d13b36118d54a7cf13aeb0ca012383bf771e713421b4d1fd841f539a
ENV HASHED_PASSWORD=CODE_PASS

ENTRYPOINT ["code-server", "--bind-addr", "0.0.0.0:9090"]
CMD ["--disable-update-check"]
