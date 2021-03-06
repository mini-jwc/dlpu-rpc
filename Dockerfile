FROM golang as builder

ENV GO111MODULE=on
#ENV GOPROXY=https://goproxy.cn,direct
# CGO_ENABLED alpine禁用cgo

WORKDIR /app
ADD go.mod .
ADD go.sum .
RUN go mod download

COPY . .
RUN go build  -o dlpu-rpc ./

RUN mkdir publish && cp dlpu-rpc publish

FROM alpine
RUN apk add gcompat
WORKDIR /app
COPY --from=builder /app/publish .
COPY --from=builder /app/start.sh .

EXPOSE 8972
ENTRYPOINT ["./start.sh"]
