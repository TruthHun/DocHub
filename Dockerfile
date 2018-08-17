FROM ubuntu:16.04

MAINTAINER "TruthHun <TruthHun@QQ.COM>"

# 阿里云源设置
RUN echo "deb http://mirrors.aliyun.com/ubuntu/ xenial main restricted universe multiverse "\
"\ndeb http://mirrors.aliyun.com/ubuntu/ xenial-security main restricted universe multiverse "\
"\ndeb http://mirrors.aliyun.com/ubuntu/ xenial-updates main restricted universe multiverse "\
"\ndeb http://mirrors.aliyun.com/ubuntu/ xenial-proposed main restricted universe multiverse "\
"\ndeb http://mirrors.aliyun.com/ubuntu/ xenial-backports main restricted universe multiverse "\
"\ndeb-src http://mirrors.aliyun.com/ubuntu/ xenial main restricted universe multiverse "\
"\ndeb-src http://mirrors.aliyun.com/ubuntu/ xenial-security main restricted universe multiverse "\
"\ndeb-src http://mirrors.aliyun.com/ubuntu/ xenial-updates main restricted universe multiverse "\
"\ndeb-src http://mirrors.aliyun.com/ubuntu/ xenial-proposed main restricted universe multiverse "\
"\ndeb-src http://mirrors.aliyun.com/ubuntu/ xenial-backports main restricted universe multiverse" > /etc/apt/sources.list

# 安装字符编码支持
RUN apt update -y && apt install -y locales && rm -rf /var/lib/apt/lists/* \
    && localedef -i en_US -c -f UTF-8 -A /usr/share/locale/locale.alias en_US.UTF-8
ENV LANG en_US.utf8

# 时区设置（由 https://github.com/bay1 反馈时区问题并提供的设置）
ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

WORKDIR /www/dochub

# ======== 安装相关依赖 ========
# 中文字体支持，主要是避免文档转换，缺少中文字体而导致乱码的情况
# libreoffice - 将office文档转PDF
# imagemagick - 将svg键jpg，主要用于封面转化
# pdf2svg - 将PDF转成svg
# node 环境，主要是为了安装 svgo 插件，用于将 svg 文件中的一些多余字符去除，以便减小 svg 文件的体积大小
# python、poppler-utils 等，主要是为了安装 calibre 需要
# calibre -  用于将mobi、epub、txt 等电子书转PDF
RUN apt update -y && apt install -y fonts-wqy-zenhei fonts-wqy-microhei \
    wget \
    libreoffice \
    imagemagick \
    pdf2svg \
    nodejs-legacy \
    npm \
    poppler-utils \
    xdg-utils \
    xz-utils \
    python \
    && npm install -g svgo \
    && wget -nv -O- https://download.calibre-ebook.com/linux-installer.sh | sh /dev/stdin \
    && mkdir -p /www/dochub && chmod 0777 -R /www/dochub/

RUN  wget https://github.com/TruthHun/DocHub/releases/download/v2.0/DocHub.V2.0_linux_amd64.zip \
    && apt install unzip -y \
    && unzip DocHub.V2.0_linux_amd64.zip -d /www/dochub/ \
    && rm -rf /www/dochub/__MACOSX \
    && mv /www/dochub/zoneinfo.zip /usr/local/go/lib/time/zoneinfo.zip

CMD [ "./DocHub" ]