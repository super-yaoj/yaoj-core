#!/bin/env bash
go generate
go build -o judgeserver -ldflags="-extldflags=-static" ./cmd/judgeserver 
docker build --file build/Dockerfile -t yaoj-judgeserver .
rm judgeserver