FROM golang:1.25-bookworm@sha256:3b4a11519ad929d1e1d261a12cff056f0c85b735253d7d861346b9c6f8b36437

ENV DEBIAN_FRONTEND=noninteractive
WORKDIR /opt/chr-pi

# Select the published Linux archive by platform; never install a floating Node.
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates curl git make ripgrep fd-find xz-utils \
    && case "$(dpkg --print-architecture)" in \
        amd64) node_arch=x64; node_sha=df450af89261115ef9f9e3830c3eeb2cc9213b63c720b1af623cb5dcbe2e02de ;; \
        arm64) node_arch=arm64; node_sha=a44aeb94849a299b22df10b9e622ec2f605c2183501bc40590705131de7c740f ;; \
        *) echo 'Pi runtime supports Linux amd64 and arm64 only' >&2; exit 1 ;; \
    esac \
    && curl --fail --location --silent --show-error \
        "https://nodejs.org/dist/v22.23.3/node-v22.23.3-linux-${node_arch}.tar.xz" -o /tmp/node.tar.xz \
    && echo "${node_sha}  /tmp/node.tar.xz" | sha256sum --check --strict \
    && tar -xJf /tmp/node.tar.xz -C /usr/local --strip-components=1 \
    && ln -s /usr/bin/fdfind /usr/local/bin/fd \
    && rm -f /tmp/node.tar.xz \
    && rm -rf /var/lib/apt/lists/*

COPY package.json package-lock.json ./
RUN npm ci --ignore-scripts --no-audit --no-fund \
    && ln -s /opt/chr-pi/node_modules/.bin/pi /usr/local/bin/pi \
    && npm cache clean --force \
    && test "$(pi --version)" = 1.3.0

LABEL org.chr.harness="pi" \
      org.chr.pi.version="1.3.0" \
      org.chr.pi.protocol="pi-rpc/1.3.0" \
      org.chr.node.version="22.23.3"

WORKDIR /workspace
CMD ["bash"]
