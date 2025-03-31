FROM golang:1.20-alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED 0
ENV GOPROXY https://goproxy.cn,direct
# 设置编码
ENV LANG C.UTF-8

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

RUN apk update --no-cache && apk add --no-cache tzdata

WORKDIR /go/release

ADD go.mod .
ADD go.sum .
RUN go mod download
COPY . .
RUN GOOS=linux go build -ldflags="-s -w" -o  go-main .


FROM alpine

# 安装tzdata包
RUN apk add --no-cache tzdata

# todo
ENV PROFILE_ACTIVES="prod"

#ENV PROFILE_ACTIVES="local"

ENV TZ=Asia/Shanghai


COPY --from=builder /go/release/go-main /apps/prod-mcgg2025wa-server/go-main

#COPY --from=builder /go/release/resources/config/application-$PROFILE_ACTIVES.yml /apps/prod-mcgg2025wa-server/resources/config/application-prod.yml

# todo 临时修改
COPY --from=builder /go/release/resources/config /apps/prod-mcgg2025wa-server/resources/config
COPY --from=builder /go/release/resources/mapper /apps/prod-mcgg2025wa-server/resources/mapper
COPY --from=builder /go/release/resources/image /apps/prod-mcgg2025wa-server/resources/image


# COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

WORKDIR /apps/prod-mcgg2025wa-server
# 暴露端口
EXPOSE 9000

ENTRYPOINT ["/apps/prod-mcgg2025wa-server/go-main","start"]
