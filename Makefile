BUILD_HOME=/go/src/github.com/TruthHun/DocHub
IMAGE_NAME=truthhun/dochub
VERSION=1.0
.PHONY: build

build: 
	docker run -it -v `pwd`:$(BUILD_HOME) -w $(BUILD_HOME) golang:1.8.3 go get && go build -ldflags '-w -s'  -o .cmd/dochub 

image: build
    @docker build -t $(IMAGE_NAME):$(VERSION) .