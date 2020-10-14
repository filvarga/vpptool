FROM ubuntu:18.04

# general
RUN set -eux; \
    apt-get update; \
    apt-get install -y --no-install-recommends ca-certificates \
    sudo gcc g++ python git zlib1g-dev pciutils kmod strace wget; \
    rm -rf /var/lib/apt/lists/*

WORKDIR /opt

RUN set -eux; \
    git clone https://github.com/cisco-system-traffic-generator/trex-core.git

WORKDIR /opt/trex-core

ADD patches /patches

RUN if [ $(ls -1 *.patch 2>/dev/null | wc -l) != 0 ]; then \
      find /patches -type f -name '*\.patch' | xargs -n1 git apply; \
    fi

WORKDIR /opt/trex-core/linux_dpdk

RUN set -eux; \
    ./b configure; \
    ./b build

ENV LC_ALL=C.UTF-8
ENV LANG=C.UTF-8

# extra tools
RUN set -eux; \
    apt-get update; \
    apt-get install -y --no-install-recommends \
    inetutils-traceroute less iputils-ping \
    net-tools iproute2 tcpdump netcat vim; \
    rm -rf /var/lib/apt/lists/*; \
    mv /usr/sbin/tcpdump /usr/bin/tcpdump

WORKDIR /opt/trex-core/scripts

CMD ["bash"]
