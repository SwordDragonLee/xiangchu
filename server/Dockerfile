# 多阶段构建：golang 编译 → alpine 运行
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o /out/server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone
COPY --from=build /out/server /usr/local/bin/server
COPY server/configs/config.yaml /app/configs/config.yaml
WORKDIR /app
EXPOSE 8080
# 运行时通过 XC_ 前缀环境变量覆盖配置（见 docker-compose.yml）
ENTRYPOINT ["server", "-c", "configs/config.yaml"]
