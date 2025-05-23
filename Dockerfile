# 使用官方的 Golang 镜像作为基础镜像
FROM golang:1.24-alpine AS builder

ARG APP_NAME
ENV APP_NAME=$APP_NAME

LABEL maintainer="weirwei <nightnessss@163.com>"
# 设置编译目录
WORKDIR $GOPATH/$APP_NAME
COPY go.mod $GOPATH/${APP_NAME}/
COPY go.sum $GOPATH/${APP_NAME}/

# 下载依赖
ENV GOPROXY='https://goproxy.cn,https://goproxy.io,direct'
RUN go mod download

# 编译
COPY . $GOPATH/${APP_NAME}/

# 使用阿里云镜像，加快 apk 包的下载速度
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# --- 添加这一行：安装构建所需的工具，包括 gcc ---
# --no-cache 可以避免在镜像中保留 apk 缓存，减小镜像大小
RUN apk --no-cache add build-base

ENV CGO_ENABLED=1

RUN go build -o /usr/local/bin/${APP_NAME} main.go

FROM alpine:3.20

ARG APP_NAME
ENV APP_NAME=$APP_NAME

# 设置运行目录
WORKDIR /usr/local/bin/

# 拷贝编译好的程序
COPY --from=builder /usr/local/bin/${APP_NAME} /usr/local/bin/
# 拷贝配置文件
COPY ./conf/yaml /usr/local/bin/conf/yaml

# 运行程序
CMD ["sh", "-c", "/usr/local/bin/$APP_NAME | tee /usr/local/bin/app.log"]