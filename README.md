# gpxtoolkit

本工具可讀取原始GPX檔案，依指定間距計算出里程航點 (例如 0.1K, 0.2K 等等) 後產生新的GPX檔案。

## Motivation

本工具源於桃園市山岳協會及中華民國山岳協會持續推動[登山路標的標準化及建置](https://www.tytaaa.org.tw/news/7)，為簡化及加速路標架設的GPX前置處理作業，因此開始創作此工具。

## How to Use

### Build from Scratch

Build executable `./gpxtoolkit`:

```shell
make
```

Check for all options:

```shell
./gpxtoolkit -h
```

#### Run as a Command Line Tool

```shell
./gpxtoolkit test.gpx
```

#### Run as a HTTP server

```shell
./gpxtoolkit -D
```

### Build as Docker

Build docker image:

```shell
make docker/build
```

Run docker image as a HTTP server:

```shell
make docker/run
```

### Access Our Public Endpoint

You can also check our public endpoint by a web browser: https://gpxtoolkit.outdoorsafetylab.org/
