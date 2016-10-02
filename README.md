# pepper

[![Travis CI](https://travis-ci.org/jessfraz/pepper.svg?branch=master)](https://travis-ci.org/jessfraz/pepper)

Named after Pepper Potts. Set all your GitHub repos master branches to be
protected.

You can set which orgs to run the script for and use `--dry-run` to see the
changes before they are actually made. Your user is automatically added to the
repositories it will consider.

## Usage

```console
$ pepper -h
pepper - v0.1.0
  -d    run in debug mode
  -dry-run
        do not change branch settings just print the changes that would occur
  -orgs value
        organizations to include
  -token string
        GitHub API token
  -v    print version and exit (shorthand)
  -version
        print version and exit
```

```console
$ pepper --dry-run --token 12345 --orgs jessconf --orgs maintainerati
Branch master for repository jessconf/jessconf is already protected
Branch master for repository jessfraz/.vim will be changed to protected
Branch master for repository jessfraz/anonymail will be changed to protected
Branch master for repository jessfraz/apk-file will be changed to protected
Branch master for repository jessfraz/apparmor-docs will be changed to protected
Branch master for repository jessfraz/audit will be changed to protected
...
Branch master for repository jessfraz/weather will be changed to protected
Branch master for repository jessfraz/ykpiv will be changed to protected
Branch master for repository maintainerati/wontfix-cabal-site is already protected
```
