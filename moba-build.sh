#!/bin/bash

VERSION=$(git describe --tags --first-parent)
BUILDTIME=$(date -u)
ldflags="-X 'github.com/Dreamacro/clash/constant.Version=$VERSION'  \
         -X 'github.com/Dreamacro/clash/constant.BuildTime=$BUILDTIME' \
         -w -s -buildid="

mkdir -pv bin
GOARCH=amd64 GOOS=windows GOAMD64=v3 GOBUILD=CGO_ENABLED=0 \
    go build -v -tags with_gvisor  \
             -trimpath \
             -ldflags "$ldflags" \
             -o bin/clash-my.exe

