GO := CGO_ENABLED=0 go
APPNAME = nsexport
VERSION :=  $(shell cat version.txt)
IMAGE_TAG = ${VERSION}

.PHONY: build
build:
	$(GO) build -o $(APPNAME) -ldflags "-X main.version=${VERSION}" .

.PHONY: build-arm
build-arm:
	GOARCH=arm $(GO) build -o $(APPNAME) -ldflags "-X main.version=${VERSION}" .

.PHONY: build-win
build-win:
	GOOS=windows $(GO) build -o $(APPNAME).exe -ldflags "-X main.version=${VERSION}" .