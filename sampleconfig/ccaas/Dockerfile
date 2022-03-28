FROM golang:1.16-alpine3.14 as builder
WORKDIR /opt
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN apk --no-cache add build-base
COPY sampleconfig/ccaas/ccaas_chaincode/ .
RUN GOPROXY="https://goproxy.io" GOARCH=amd64 GOOS=linux go build -o fabric-chaincode-server basic.go

FROM alpine:3.14 AS final
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN apk add tzdata
RUN cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
RUN echo "Asia/Shanghai" >  /etc/timezone
RUN apk del tzdata
WORKDIR /opt
COPY --from=builder /opt/fabric-chaincode-server /opt
EXPOSE 9999
CMD /opt/fabric-chaincode-server