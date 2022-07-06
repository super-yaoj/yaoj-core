#!/bin/env bash
go generate
CC="clang" go build -o judgeserver -ldflags="-extldflags=-static" ./cmd/judgeserver 
docker build --file build/Dockerfile -t ghcr.io/super-yaoj/judgeserver .
rm judgeserver