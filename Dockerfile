# syntax=docker/dockerfile:1
# Web app to schedule embeded posts to discord via webhook

FROM golang:1.22

RUN mkdir -p /goLive

WORKDIR /goLive

COPY . ./

RUN go mod download

RUN go build -o /goLive

ENV GOLIVE_PORT=:8080

CMD ["./goLiveNotif"]