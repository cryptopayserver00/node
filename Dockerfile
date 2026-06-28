FROM golang:1.25.7-alpine AS builder

WORKDIR /app

# 下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制源码构建
COPY . .
RUN go build -o main .

# 运行阶段
FROM alpine:latest AS runner

WORKDIR /app

# 从构建阶段复制二进制
COPY --from=builder /app/main .

# 复制配置文件
COPY config.yaml .

EXPOSE 8080
CMD ["./main"]