FROM golang:1.21

WORKDIR /onekonsole/sys-service-order

COPY go.mod go.sum ./

RUN go mod download && go mod verify

RUN go clean -modcache

COPY . . 

RUN go build -o /onekonsole/sys-service-order/build/app

EXPOSE 8020

ENTRYPOINT [ "/onekonsole/sys-service-order/build/app"]