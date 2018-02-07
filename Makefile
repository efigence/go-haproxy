# generate version number
version=$(shell git describe --tags --long --always|sed 's/^v//')

all: vendor | glide.lock
	@if [ "z$(TRAVIS)" != "z" ] ; then \
	    killall -9 haproxy || true;\
	fi
	-@go fmt
	go test
#go build -ldflags "-X main.version=$(version)" $(binfile).go



static: glide.lock vendor
	go build -ldflags "-X main.version=$(version) -extldflags \"-static\"" -o $(binfile).static $(binfile).go

arm:
	GOARCH=arm go build  -ldflags "-X main.version=$(version) -extldflags \"-static\"" -o $(binfile).arm $(binfile).go
	GOARCH=arm64 go build  -ldflags "-X main.version=$(version) -extldflags \"-static\"" -o $(binfile).arm64 $(binfile).go
clean:
	rm -rf vendor
	rm -rf _vendor
vendor: glide.lock
	@if [ "z$(TRAVIS)" != "z" ] && ! glide -version ; then \
		echo "Detected travis env, installing glide" ;\
		go get github.com/Masterminds/glide;\
	fi
	glide install && touch vendor
glide.lock: glide.yaml
	glide update && touch glide.lock
glide.yaml:

version:
	@echo $(version)
