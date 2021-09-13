FROM alpine
MAINTAINER Philipp C. Heckel <philipp.heckel@gmail.com>

COPY submonkey /usr/bin
RUN apk add ffmpeg curl python3 \
    && ln -s /usr/bin/python3 /usr/bin/python
RUN curl -L https://yt-dl.org/downloads/latest/youtube-dl -o /usr/bin/youtube-dl \
    && chmod a+rx /usr/bin/youtube-dl
RUN mkdir -p /.cache/submonkey \
    && chmod 777 /.cache/submonkey

WORKDIR /submonkey
ENTRYPOINT ["submonkey"]
