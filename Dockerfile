FROM golang:1.15.3-buster

WORKDIR /src/heartfort

COPY . .

RUN go build
