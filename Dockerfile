FROM registry.suse.com/bci/golang:1.26 AS builder

ARG MK_HOST_ARCH
ENV ARCH=$MK_HOST_ARCH
ENV GOTOOLCHAIN=auto
ARG CONTAINER_WORKDIR=/go/src/github.com/harvester/managedchart-controller


RUN go install golang.org/x/tools/cmd/goimports@v0.43.0
RUN go install github.com/incu6us/goimports-reviser/v3@v3.10.0
RUN go install github.com/mikefarah/yq/v4@v4.52.5

## install golangci
RUN go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2

ENV GO111MODULE=on

ENV KUBEBUILDER_ASSETS="/usr/local/kubebuilder/envtest"
ENV HOME=/go/src/github.com/harvester/managedchart-controller

# ---- base ----
FROM builder AS base
WORKDIR /go/src/github.com/harvester/managedchart-controller

# to exclude some files, add them in .dockerignore
COPY . .

# ---- build ----
FROM base AS build
ARG MK_REPO_ID

RUN --mount=type=cache,target=/go/pkg/mod,id=managedchart-controller-go-mod-${MK_REPO_ID} \
    --mount=type=cache,target=/go/src/github.com/harvester/managedchart-controller/.cache/go-build,id=managedchart-controller-go-build-${MK_REPO_ID} \
    ./scripts/build

FROM scratch AS build-output
COPY --from=build /go/src/github.com/harvester/managedchart-controller/bin/ /bin/


# ---- validate ----
FROM base AS validate
ARG MK_REPO_ID

RUN --mount=type=cache,target=/go/pkg/mod,id=managedchart-controller-go-mod-${MK_REPO_ID} \
    --mount=type=cache,target=/go/src/github.com/harvester/managedchart-controller/.cache/go-build,id=managedchart-controller-go-build-${MK_REPO_ID} \
    ./scripts/validate

# ---- test ----
FROM base AS test

# ---- generate ----
FROM base AS generate
ARG MK_REPO_ID

RUN --mount=type=cache,target=/go/pkg/mod,id=managedchart-controller-go-mod-${MK_REPO_ID} \
    --mount=type=cache,target=/go/src/github.com/harvester/managedchart-controller/.cache/go-build,id=managedchart-controller-go-build-${MK_REPO_ID} \
    ./scripts/generate



FROM scratch AS generated-controller
COPY --from=generate /go/src/github.com/harvester/managedchart-controller/pkg/generated/ /generated/
