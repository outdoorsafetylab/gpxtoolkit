FROM golang:1.16-alpine as builder

RUN apk update \
    && apk upgrade \
    && apk add --no-cache \
    && rm -rf /var/cache/apk/* /tmp/*

RUN mkdir -p /src/
COPY . /src/
WORKDIR /src/

RUN go build -o gpxtoolkit .

FROM alpine

RUN apk update \
    && apk upgrade \
    && apk add --no-cache \
    && rm -rf /var/cache/apk/* /tmp/*

COPY --from=builder /src/gpxtoolkit /usr/sbin/
RUN mkdir -p /var/www/html/
COPY --from=builder /src/webroot/ /var/www/html/

ENTRYPOINT [ "/usr/sbin/gpxtoolkit", "-D", "-w", "/var/www/html" ]
