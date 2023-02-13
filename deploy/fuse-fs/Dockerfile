FROM cgr.dev/chainguard/go:1.19 as builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/fuse-fs ./cmd/fuse-fs
COPY internal/ internal
RUN CGO_ENABLED=0 go build -ldflags '-s -w' -o fuse-fs ./cmd/fuse-fs

FROM cgr.dev/chainguard/alpine-base
RUN apk add fuse
COPY --from=builder /src/fuse-fs /fuse-fs
ENTRYPOINT ["/fuse-fs"]

