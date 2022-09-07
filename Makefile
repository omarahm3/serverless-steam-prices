.PHONY: build clean deploy gomodgen dev

build-frontend:
	cd ./frontend && yarn build

build-functions:
	export GO111MODULE=on
	cd ./functions/get/ && env CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -o ../../bin/get get.go && cd ..

build: build-frontend build-functions

clean:
	rm -rf ./bin ./vendor go.sum

deploy: clean build
	sls deploy --verbose

dev: build
	clear && sls invoke local -f get --verbose
