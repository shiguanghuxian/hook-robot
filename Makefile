BINARY=hook-robot

default:
	@echo 'Usage of make: [ build | linux | windows | run | clean ]'

build: 
	go build -o ./bin/${BINARY} ./

linux: 
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/${BINARY} ./

docker: 
	

run: build
	cd bin && ./${BINARY}

clean: 
	rm -f ./${BINARY}*

.PHONY: default build linux run docker clean