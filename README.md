# Gohazel

A versions update server writen in Golang. Supports updating an [Electron](https://www.electronjs.org/docs/tutorial/updates) applicaiton.

- Response Github public download url for public repo.
- Response your private server download url for private repo. Cache release information and assets by Github api.

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

## Assets filename

Supporting patterns: `*.exe`,`*.dmg`, `*.rpm`, `*.deb`, `*.AppImage`, `*mac*.zip`, `*darwin*.zip`

References release of atom: https://github.com/atom/atom/releases

## Run

`config.yml`

```yml
bind: ":8080"
debug: true
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
