#!/bin/bash
d=$(pwd)

env PKG_CONFIG=arm-linux-gnueabi-pkg-config PKG_CONFIG_PATH=$d CC=arm-linux-gnueabi-gcc CGO_ENABLED=1 GOOS=linux GOARCH=arm GOARM=6 go build -v -ldflags="-extld=$CC -rpath=$d/lib/arm-linux-gnueabi" .
env PKG_CONFIG_PATH=$d arm-linux-gnueabi-pkg-config --cflags libusb-1.0
#-ldflags  "-linkmode external -extldflags -static"
#-I/usr/include/libusb-1.0
