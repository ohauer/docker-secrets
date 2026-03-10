FROM alpine:3.21

# Install required tools
RUN apk add --no-cache \
    inotify-tools \
    procps \
    curl \
    bash

# Copy the reload script
COPY reload-sidecar.sh /usr/local/bin/reload-sidecar.sh
RUN chmod +x /usr/local/bin/reload-sidecar.sh

# Run as non-root user
USER 65534:65534

ENTRYPOINT ["/usr/local/bin/reload-sidecar.sh"]
