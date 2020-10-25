# Gohazel

[![Build][build-status-image]][build-status-url]

A versions update server writen in Golang. Supports updating an [Electron](https://www.electronjs.org/docs/tutorial/updates) application.

- **Private repo** - Response your proxy server download url. Cache release information and assets by Github api.
- **Public repo** - Response Github public download url directly. Alse support proxy.

[build-status-url]: https://travis-ci.org/panjiang/gohazel
[build-status-image]: https://travis-ci.org/panjiang/gohazel.svg?branch=master
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
$ curl http://localhost:8080/download
```

- Github directly

```json
{"Location":"https://github.com/atom/atom/releases/download/v1.52.0/AtomSetup.exe"}
```

- Server proxy

```json
{"Location":"http://localhost:8080/assets/atom/atom/v1.52.0/AtomSetup.exe"}
```

### `/download/:platform`

Responses download url for specified platform in uri.

```console
$ curl http://localhost:8080/download/darwin
```

- Github directly

```json
{"Location":"https://github.com/atom/atom/releases/download/v1.52.0/atom-mac.zip"}
```

- Server proxy

```json
{"Location":"http://localhost:8080/assets/atom/atom/v1.52.0/atom-mac.zip"}
```

### `/update/:platform/:version`

Check update info

```
$ curl http://localhost:8080/update/win/v0.0.1
{"name":"v1.52.0","notes":"## Notable Changes...","pub_data":"2020-10-13T14:11:00Z","url":"http://localhost:8080/download/exe?update=true"}
```

### `/update/win32/:version/RELEASES`

For Squirrel Windows

## Assets Filename

Supporting patterns: `*.exe`,`*.dmg`, `*.rpm`, `*.deb`, `*.AppImage`, `*mac*.zip`, `*darwin*.zip`

References release of atom: https://github.com/atom/atom/releases

## Config File

`config.yml`

```yml
bind: ":8080"
debug: false
debugGin: false
baseURL: http://localhost:8080
cacheDir: /assets
proxyDownload: false
github:
  owner: atom
  repo: atom
  token:
  pre: false
```

- `baseURL` - Public base URL of the server.
- `cacheDIr` - The directory for store cache release info and asset files.
- `proxyDownload` - Whether to let the server proxy assets download.
- `github.owner` - Account username.
- `github.repo` - Repository name.
- `github.token` - The Github API Token for fetching release info and download assets from private repo.
- `github.pre` - Whether to fetch the pre-release versions.

## Run with Container

Docker Repository: [panjiang/gohazel](https://hub.docker.com/repository/docker/panjiang/gohazel)

- Write your config file `/data/gohazel/config.yml`
- Store cache files in `/data/gohazel/assets`

### Docker

```console
docker run -d --name gohazel \
		-v /data/gohazel/config.yml:/app/config.yml \
		-v /data/gohazel/assets:/assets \
		-p 8080:8080 \
		panjiang/gohazel:latest
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
      - "8080:8080"
    volumes:
      - /data/gohazel/config.yml:/app/config.yml
      - /data/gohazel/assets:/assets
```

```console
$ docker-compose up -d
```
