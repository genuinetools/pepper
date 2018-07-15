<p align="left"><img src="logo/horizontal.png" alt="pepper" height="160px"></p>

[![Travis CI](https://img.shields.io/travis/genuinetools/pepper.svg?style=for-the-badge)](https://travis-ci.org/genuinetools/pepper)
[![GoDoc](https://img.shields.io/badge/godoc-reference-5272B4.svg?style=for-the-badge)](https://godoc.org/github.com/genuinetools/pepper)
[![Github All Releases](https://img.shields.io/github/downloads/genuinetools/pepper/total.svg?style=for-the-badge)](https://github.com/genuinetools/pepper/releases)

Named after Pepper Potts. Set all your GitHub repos master branches to be
protected.

You can set which orgs to include and use `--dry-run` to see the
changes before they are actually made. Your user is automatically added to the
repositories it will consider.

Also see [genuinetools/audit](https://github.com/genuinetools/audit) for checking what
collaborators, hooks, deploy keys, and protected branched you have added on
all your GitHub repositories.

## Installation

#### Binaries

For installation instructions from binaries please visit the [Releases Page](https://github.com/genuinetools/pepper/releases).

- **darwin** [386](https://github.com/genuinetools/pepper/releases/download/v0.5.2/pepper-darwin-386) / [amd64](https://github.com/genuinetools/pepper/releases/download/v0.5.2/pepper-darwin-amd64)
- **freebsd** [386](https://github.com/genuinetools/pepper/releases/download/v0.5.2/pepper-freebsd-386) / [amd64](https://github.com/genuinetools/pepper/releases/download/v0.5.2/pepper-freebsd-amd64)
- **linux** [386](https://github.com/genuinetools/pepper/releases/download/v0.5.2/pepper-linux-386) / [amd64](https://github.com/genuinetools/pepper/releases/download/v0.5.2/pepper-linux-amd64) / [arm](https://github.com/genuinetools/pepper/releases/download/v0.5.2/pepper-linux-arm) / [arm64](https://github.com/genuinetools/pepper/releases/download/v0.5.2/pepper-linux-arm64)
- **solaris** [amd64](https://github.com/genuinetools/pepper/releases/download/v0.5.2/pepper-solaris-amd64)
- **windows** [386](https://github.com/genuinetools/pepper/releases/download/v0.5.2/pepper-windows-386) / [amd64](https://github.com/genuinetools/pepper/releases/download/v0.5.2/pepper-windows-amd64)

#### Via Go

```bash
$ go get github.com/genuinetools/pepper
```

## Usage

```console
$ pepper -h
 _ __   ___ _ __  _ __   ___ _ __
| '_ \ / _ \ '_ \| '_ \ / _ \ '__|
| |_) |  __/ |_) | |_) |  __/ |
| .__/ \___| .__/| .__/ \___|_|
|_|        |_|   |_|

 Set all your GitHub repos master branches to be protected.
 Version: v0.5.2
 Build: 8b7274f

  -d        run in debug mode
  -dry-run  do not change branch settings just print the changes that would occur
  -nouser   do not include your user
  -orgs     organizations to include
  -token    GitHub API token (or env var GITHUB_TOKEN)
  -url      GitHub Enterprise URL
  -v        print version and exit (shorthand)
  -version  print version and exit
```

```console
$ pepper --dry-run --token 12345 --orgs jessconf --orgs maintainerati
[OK] jessconf/jessconf:master is already protected
[OK] genuinetools/.vim:master is already protected
[OK] genuinetools/anonymail:master is already protected
[OK] genuinetools/apk-file:master is already protected
[UPDATE] genuinetools/certok:master will be changed to protected
...
[OK] genuinetools/weather:master is already protected
[OK] genuinetools/ykpiv:master is already protected
[OK] maintainerati/wontfix-cabal-site:master is already protected
```
