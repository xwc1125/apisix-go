# 运行golang程序
FROM alpine:latest
MAINTAINER xwc1125 <xwc1125@qq.com>

USER root

##定义时区参数
ENV TZ=Asia/Shanghai
# 设置时区
RUN rm -rf /etc/localtime &&\
    ln -sv /usr/share/zoneinfo/$TZ /etc/localtime &&\
    echo '$TZ' >/etc/timezone

ENV APPDIR=/data
ENV TEMPDIR=/temp
RUN mkdir -p $APPDIR $TEMPDIR

COPY ./ $TEMPDIR

RUN cp $TEMPDIR/app /usr/bin

RUN rm -rf $TEMPDIR

# 添加账户
RUN mkdir -p /data /home/app \
    && adduser -D app \
    && chown -R 755 /data /home/app
# 运行非特权的app账户
USER app

WORKDIR $APPDIR
VOLUME $APPDIR/

EXPOSE 8080

ENTRYPOINT ["app","version"]