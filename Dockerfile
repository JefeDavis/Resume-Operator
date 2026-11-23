# Build the manager binary on a Chainguard Go image.
# The -dev variant includes a shell/apk so Docker's shell-form RUN works; switch
# to root so the non-root default user can write to WORKDIR and the build cache.
FROM cgr.dev/chainguard/go:latest-dev AS builder
USER root
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY apis/ apis/
COPY controllers/ controllers/
COPY internal/ internal/

# Build a fully static binary (CGO disabled) so it runs on scratch.
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a -o manager main.go

# Final image: scratch. The static binary needs nothing else; CA certificates are
# copied in as insurance for any outbound TLS (in-cluster Kubernetes API calls use
# the mounted cluster CA and do not require these).
FROM scratch
WORKDIR /
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /workspace/manager /manager
USER 65532:65532

ENTRYPOINT ["/manager"]
