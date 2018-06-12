FROM goodrainapps/alpine:3.4

ENV LD_LIBRARY_PATH $LD_LIBRARY_PATH:/opt/calibre/lib
ENV PATH $PATH:/opt/calibre/bin:/run
ENV CALIBRE_INSTALLER_SOURCE_CODE_URL https://raw.githubusercontent.com/kovidgoyal/calibre/master/setup/linux-installer.py
RUN apk update && \
    apk add --no-cache --upgrade \
    imagemagick \
    poppler-utils \
    poppler-glib \
    msttcorefonts-installer \
    bash \
    ca-certificates \
    gcc \
    mesa-gl \
    python \
    qt5-qtbase-x11 \
    wget \
    xdg-utils \
    xz && \
    wget -O- ${CALIBRE_INSTALLER_SOURCE_CODE_URL} | python -c "import sys; main=lambda:sys.stderr.write('Download failed\n'); exec(sys.stdin.read()); main(install_dir='/opt', isolated=True)" && \
    update-ms-fonts && \
    rm -rf /tmp/calibre-installer-cache
ADD .output /run
ADD static /run/static
COPY conf/app.conf.example /run/app.conf
COPY conf/database.conf.example /run/database.conf
COPY conf/email.conf.example /run/email.conf
COPY conf/oss.conf.example /run/oss.conf

WORKDIR /run
EXPOSE 8090
CMD ["/run/dochub"]