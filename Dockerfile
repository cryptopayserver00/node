FROM golang:1.27-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" -o /build/app ./main.go

FROM alpine:3.20 AS runner

RUN apk add --no-cache ca-certificates tzdata wget

WORKDIR /app

COPY --from=builder /build/app /app/app

RUN mkdir -p /app/log

RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup && \
    chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget --no-verbose --tries=1 -O /dev/null http://127.0.0.1:8080/api/health || exit 1

ENTRYPOINT ["/app/app"]