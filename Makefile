GO := CGO_ENABLED=0 go
APPNAME = nsexport
VERSION :=  $(shell cat version.txt)
IMAGE_NAME = blutz1982/$(APPNAME)
IMAGE_TAG = ${VERSION}

.PHONY: build
build:
	$(GO) build -o build/$(APPNAME) -ldflags "-X main.version=${VERSION}" .

.PHONY: build-arm
build-arm:
	GOARCH=arm $(GO) build -o build/$(APPNAME)-arm -ldflags "-X main.version=${VERSION}" .

.PHONY: build-win
build-win:
	GOOS=windows $(GO) build -o build/$(APPNAME).exe -ldflags "-X main.version=${VERSION}" .

.PHONY: all
all: build build-win build-arm

.PHONY: build-image
build-image:
	docker build -t ${IMAGE_NAME}:${IMAGE_TAG} .
	docker tag ${IMAGE_NAME}:${IMAGE_TAG} ${IMAGE_NAME}:latest

.PHONY: docker
docker: build-image push-image

.PHONY: push-image
push-image:
	docker push ${IMAGE_NAME}:${IMAGE_TAG}
	docker push ${IMAGE_NAME}:latest