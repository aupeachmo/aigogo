FROM ubuntu:24.04

RUN apt-get update && apt-get install -y --no-install-recommends \
    python3 \
    python3-pip \
    python3-poetry \
    nodejs \
    npm \
    git \
    ca-certificates \
    golang-go \
    make \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /workspace

# Copy the pre-built Linux binary into the image
COPY bin/aigogo-linux-* /usr/local/bin/aigogo
RUN chmod +x /usr/local/bin/aigogo

CMD ["bash"]
