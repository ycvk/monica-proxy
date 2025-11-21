FROM golang:alpine AS deps
WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

FROM golang:alpine AS builder
ARG TARGETOS=linux
ARG TARGETARCH=amd64
WORKDIR /app
RUN apk add --no-cache make
COPY --from=deps /app/go.mod /app/go.sum ./
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    make build GOOS=${TARGETOS} GOARCH=${TARGETARCH}

FROM gcr.io/distroless/static:nonroot AS final
WORKDIR /data
COPY --from=builder /app/build/monica /data/monica

EXPOSE 8080
USER nonroot:nonroot
CMD ["./monica"]