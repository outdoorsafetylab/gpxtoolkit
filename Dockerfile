FROM golang:1.18-alpine as go-builder

RUN apk update \
    && apk upgrade \
    && apk add --no-cache \
    && rm -rf /var/cache/apk/* /tmp/*

RUN mkdir -p /src/
COPY . /src/
WORKDIR /src/

ARG GIT_HASH
ARG GIT_TAG
RUN go build -ldflags="-X gpxtoolkit/version.GitHash=${GIT_HASH} -X gpxtoolkit/version.GitTag=${GIT_TAG}" -o gpxtoolkit .

FROM node:alpine as npm-builder

RUN mkdir -p /src/
COPY ./webroot /src/
WORKDIR /src/

RUN npm install
RUN npm run build

FROM alpine

RUN apk update \
    && apk upgrade \
    && apk add --no-cache \
    && rm -rf /var/cache/apk/* /tmp/*

COPY --from=go-builder /src/gpxtoolkit /usr/sbin/
RUN mkdir -p /var/www/html/
COPY --from=npm-builder /src/dist/ /var/www/html/

ENV ELEVATION_URL=
ENV ELEVATION_TOKEN=

ENTRYPOINT [ "/usr/sbin/gpxtoolkit", "-w", "/var/www/html" ]
