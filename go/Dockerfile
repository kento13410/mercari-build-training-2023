# build環境
FROM golang:1.20.2-alpine as builder

WORKDIR /go

COPY go/go.mod ./
COPY go/go.sum ./
COPY go/app ./

ENV GOPATH=
RUN go mod download
RUN apk add --no-cache git

COPY go .
WORKDIR /go/app
RUN go build -o main .


# 実行環境
FROM alpine

RUN addgroup -S mercari && adduser -S trainee -G mercari
# RUN chown -R trainee:mercari /path/to/db

USER trainee

WORKDIR /go

COPY --from=builder go/app ./app
COPY --from=builder go/images ./

CMD ["./app/main"]
