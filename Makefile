.PHONY: build

default:build

build: tidy
	GOOS=linux GOARCH=arm64  go build -o bin/crypto-analytics.arm64 .

tidy:
	go mod tidy

image:
	docker build -t crypto-analytics .