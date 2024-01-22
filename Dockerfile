FROM golang:1.21 AS build

WORKDIR /onekonsole

COPY go.mod . 
COPY go.sum . 
COPY app.go . 
COPY helpers.go . 
COPY main.go . 

RUN go mod download && go mod verify

RUN go clean --modcache

RUN CGO_ENABLED=0 go build -o /onekonsole/app

FROM gcr.io/distroless/static-debian11

EXPOSE 8020

COPY --from=build /onekonsole/app /

CMD ["/app"]