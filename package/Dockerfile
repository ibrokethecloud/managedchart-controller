# syntax=docker/dockerfile:1.25.0

FROM registry.suse.com/bci/bci-base:16.1
#awk is needed by sriov-manage utility from nvidia to perform gpu management

ARG TARGETPLATFORM

RUN if [ "$TARGETPLATFORM" != "linux/amd64" ] && [ "$TARGETPLATFORM" != "linux/arm64" ]; then \
    echo "Error: Unsupported TARGETPLATFORM: $TARGETPLATFORM" && \
    exit 1; \
    fi

ENV ARCH=${TARGETPLATFORM#linux/}

COPY bin/managedchart-controller-${ARCH} /bin/managedchart-controller
ENTRYPOINT ["managedchart-controller"]
