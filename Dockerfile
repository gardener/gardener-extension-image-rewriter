# SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

############# builder
FROM golang:1.24.3 AS builder

WORKDIR /go/src/github.com/gardener/gardener-extension-image-rewriter

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG EFFECTIVE_VERSION

RUN make install EFFECTIVE_VERSION=$EFFECTIVE_VERSION

############# base
FROM gcr.io/distroless/static-debian12:nonroot AS base
WORKDIR /

############# gardener-extension-image-rewriter
FROM base AS image-rewriter

COPY --from=builder /go/bin/gardener-extension-image-rewriter /gardener-extension-image-rewriter
ENTRYPOINT ["/gardener-extension-image-rewriter"]
