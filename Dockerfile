FROM golang:1.24-alpine AS go-builder

RUN apk update \
    && apk upgrade \
    && apk add --no-cache \
    && rm -rf /var/cache/apk/* /tmp/*

RUN mkdir -p /src/
COPY . /src/
WORKDIR /src/

RUN go mod tidy
RUN go test ./...

ARG GIT_HASH
ARG GIT_TAG
RUN echo "GIT_HASH=${GIT_HASH}" > .env && \
    echo "GIT_TAG=${GIT_TAG}" >> .env
RUN go build -o gpxtoolkit .

FROM node:alpine AS npm-builder

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
COPY --from=go-builder /src/.env /usr/sbin/.env
WORKDIR /usr/sbin/
RUN mkdir -p /var/www/html/
COPY --from=npm-builder /src/dist/ /var/www/html/

ENV ELEVATION_URL=
ENV ELEVATION_TOKEN=

ENTRYPOINT [ "./gpxtoolkit", "serve", "-w", "/var/www/html" ]
