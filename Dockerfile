FROM golang:1.16-alpine as builder

RUN apk update \
    && apk upgrade \
    && apk add --no-cache \
    && rm -rf /var/cache/apk/* /tmp/*

RUN mkdir -p /src/
COPY . /src/
WORKDIR /src/

ARG GIT_HASH
ARG GIT_TAG
RUN go build -ldflags="-X main.GitHash=${GIT_HASH} -X main.GitTag=${GIT_TAG}" -o gpxtoolkit .

FROM alpine

RUN apk update \
    && apk upgrade \
    && apk add --no-cache \
    && rm -rf /var/cache/apk/* /tmp/*

COPY --from=builder /src/gpxtoolkit /usr/sbin/
RUN mkdir -p /var/www/html/
COPY --from=builder /src/webroot/ /var/www/html/

ENV ELEVATION_URL=
ENV ELEVATION_TOKEN=

ENTRYPOINT [ "/usr/sbin/gpxtoolkit", "-D", "-w", "/var/www/html" ]
