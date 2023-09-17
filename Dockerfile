FROM golang:1.21


RUN mkdir /build
COPY / /build

WORKDIR /build
