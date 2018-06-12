FROM goodrainapps/alpine:3.4

ADD .cmd/dochub /run/dochub
ADD static /run/static
COPY conf/app.conf.example /run/app.conf
COPY conf/database.conf.example /run/database.conf
COPY conf/email.conf.example /run/email.conf
COPY conf/oss.conf.example /run/oss.conf

WORKDIR /run
EXPOSE 8090
CMD ["/run/dochub"]