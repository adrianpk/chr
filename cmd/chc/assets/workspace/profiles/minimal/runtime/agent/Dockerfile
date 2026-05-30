FROM golang:1.25-bookworm

ENV DEBIAN_FRONTEND=noninteractive
WORKDIR /workspace

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        bash \
        ca-certificates \
        curl \
        git \
        make \
        nodejs \
        npm \
    && npm install -g @openai/codex \
    && rm -rf /var/lib/apt/lists/*

CMD ["bash"]
