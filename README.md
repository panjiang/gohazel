# Gohazel

[![Build][build-status-image]][build-status-url]

A versions update server writen in Golang. Supports updating an [Electron](https://www.electronjs.org/docs/tutorial/updates) application.

- **Private repo** - Response your proxy server download url. Cache release information and assets by Github api.
- **Public repo** - Response Github public download url directly. Alse support proxy.

[build-status-url]: https://travis-ci.org/panjiang/gohazel
[build-status-image]: https://travis-ci.org/panjiang/gohazel.svg?branch=master

## Contents

- [Difference from Hazel](#difference-from-hazel)
- [URL Pathes](#url-pathes)
- [Assets Filename](#assets-filename)
- [Config File](#config-file)
- [Run with Container](#run-with-container)
  - [Docker](#docker)
  - [Docker Compose](#docker-compose)

## Difference from Hazel

The project is inspired by [Hazel](https://github.com/vercel/hazel). Hazel is very complicated to deploy, because it is coded in NodeJS.

Gohazel not only translated hazel to **Golang**, but also made some ajustments and optimizations.

- Cache assets into your server disk for private repo.
- Separate user requests and cache logic, for fast response.
- Cache latest release data, in case there is no any information for serving while fetching failed from Github at startup.

## URL Pathes

### `/`

Overview repo and cached release information.

### `/download`

Responses download url (`"Location"`) for detected platform which parsed from user agent.

```console
$ curl http://localhost:8400/download
```

- Github directly

```json
{"Location":"https://github.com/atom/atom/releases/download/v1.52.0/AtomSetup.exe"}
```

- Server proxy

```json
{"Location":"http://localhost:8400/assets/atom/atom/v1.52.0/AtomSetup.exe"}
```

### `/download/:platform`

Responses download url for specified platform in uri.

```console
$ curl http://localhost:8400/download/darwin
```

- Github directly

```json
{"Location":"https://github.com/atom/atom/releases/download/v1.52.0/atom-mac.zip"}
```

- Server proxy

```json
{"Location":"http://localhost:8400/assets/atom/atom/v1.52.0/atom-mac.zip"}
```

### `/update/:platform/:version`

Check update info

```
$ curl http://localhost:8400/update/win/v0.0.1
{"name":"v1.52.0","notes":"## Notable Changes...","pub_data":"2020-10-13T14:11:00Z","url":"http://localhost:8400/download/exe?update=true"}
```

### `/update/win32/:version/RELEASES`

For Squirrel Windows

## Assets Filename

Supporting patterns: `*.exe`,`*.dmg`, `*.rpm`, `*.deb`, `*.AppImage`, `*mac*.zip`, `*darwin*.zip`

References release of atom: https://github.com/atom/atom/releases

## Command Flags

```text
Usage: gohazel [options]
Server Options:
    -addr             Server listen address.
    -base_url         The server base URL.
    -debug            Open log debug level.
    -cache_dir        Cache files store in.
    -proxy_download   Proxy assets download with the server.
    -github_owner     Gihtub owner name.
    -github_repo      Github repository name.
    -github_token     Github api token for private repo.
    -config           Or specify a YAML configuration file.
```

## Or Config File

`config.yml`

```yml
bind: ":8400"
debug: false
debugGin: false
baseURL: http://localhost:8400
cacheDir: /assets
proxyDownload: false
github:
  owner: atom
  repo: atom
  token:
  pre: false
```

## Run with Container

Docker Repository: [panjiang/gohazel](https://hub.docker.com/repository/docker/panjiang/gohazel)

- Store cache files in `/data/gohazel/assets`

### Docker

```console
docker run -d --rm --name gohazel \
        -v /data/gohazel/config.yml:/config.yml\
		-v /data/gohazel/assets:/assets \
		-p 8400:8400 \
		panjiang/gohazel
```

### Docker Compose

`docker-compose.yml`

```yml
version: "3.7"
services:
  gohazel:
    container_name: gohazel
    image: panjiang/gohazel:latest
    ports:
      - "8400:8400"
    volumes:
      - /data/gohazel/assets:/assets
    command:
      - /gohazel
      - -addr=:8400
      - -base_url=http://localhost:8400
      - -cache_dir=/assets
      - -proxy_download=false
      - -github_owner=atom
      - -github_repo=atom
      - -github_token=
```

```console
$ docker-compose up -d
```
