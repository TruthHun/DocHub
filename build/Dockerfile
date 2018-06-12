FROM goodrainapps/alpine:3.4
LABEL version="0.2.3-alpine3.6"

RUN apk add --update-cache make gcc libc-dev cairo-dev poppler-dev git

ENV PDF2SVG_VERSION v0.2.3

RUN git clone https://github.com/dawbarton/pdf2svg pdf2svg

WORKDIR /pdf2svg

RUN git checkout $PDF2SVG_VERSION && ./configure && make
CMD [ "cp","/pdf2svg/pdf2svg","/pdf2svg/.output/pdf2svg"]