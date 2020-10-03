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

- `/` - Overview repo and cached release information.
- `/download` - Responses download url (`"Location"`) for detected platform which parsed from user agent.

```console
$ curl http://{your-update-server-domain}/download
{"Location":"http://{your-update-server-domain}/assets/Crownote.Setup.1.0.0.exe"}
```

- `/download/:platform` Responses download url for specified platform in uri.

```console
$ curl http://{your-update-server-domain}/download/mac
{"Location":"http://{your-update-server-domain}/assets/Crownote-1.0.0.dmg"}
```

- `/update/:platform/:version` Check update info

```
$ curl http://172.20.117.157:8080/update/win/v0.0.1
{"name":"v1.0.0","notes":"1. Basic notebook functions.\r\n2. Synchronize with remote.","pub_data":"2020-10-03T03:25:37Z","url":"http://{your-update-server-domain}/download/exe?update=true"}
```

- `/update/win32/:version/RELEASES` For Squirrel Windows
