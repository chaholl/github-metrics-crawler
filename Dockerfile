# Build the manager binary
FROM golang:1.16 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY crawlers/ crawlers/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o metrics-crawler main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/metrics-crawler .
USER nonroot:nonroot

ENV GITHUB_ORG=""
ENV GITHUB_TOKEN=""
ENV POLLING_INTERVAL_MINUTES="30"

EXPOSE 8080/tcp

ENTRYPOINT ["/metrics-crawler"]