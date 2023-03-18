FROM golang:alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOPROXY=https://goproxy.cn,https://goproxy.io,direct

WORKDIR /build

COPY . .
RUN go mod download

RUN go build -o /app/wallcrow .

FROM alpine

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

RUN apk update --no-cache && apk add --no-cache ca-certificates tzdata
ENV TZ Asia/Shanghai

WORKDIR /app
COPY --from=builder /app/wallcrow /app/wallcrow

VOLUME /dir

EXPOSE 8088

ENTRYPOINT ["./wallcrow"]
