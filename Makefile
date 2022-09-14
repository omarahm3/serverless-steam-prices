.PHONY: build clean deploy gomodgen dev

build-frontend:
	cd ./frontend && yarn && yarn build && cd ..

build-functions:
	export GO111MODULE=on
	cd ./functions/get/ && env CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -o ../../bin/get get.go && cd ../..

build-nt-functions:
	export GO111MODULE=on
	cd ./nt-functions/get/ && env CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -o ../../bin/get get.go && cd ../..

build: build-frontend build-functions

build-nt: build-frontend build-nt-functions

clean:
	rm -rf ./bin ./vendor go.sum

deploy: clean build-functions
	sls deploy --verbose

dev: build
	clear && sls invoke local -f get --verbose
