BUILD_HOME=/go/src/github.com/TruthHun/DocHub
IMAGE_NAME=truthhun/dochub
VERSION=1.0
.PHONY: build

build: 
	@docker build -t pdf2svg:builder ./build/Dockerfile
	@docker run -it -v `pwd`/.output:/pdf2svg/.output --rm pdf2svg:builder
	@docker run -it -v `pwd`:$(BUILD_HOME) -w $(BUILD_HOME) golang:1.8.3 go get && go build -ldflags '-w -s'  -o .output/dochub 
image: build
	@docker build -t $(IMAGE_NAME):$(VERSION) .
	@rm -rf .output