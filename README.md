# pepper

[![Travis CI](https://travis-ci.org/jessfraz/pepper.svg?branch=master)](https://travis-ci.org/jessfraz/pepper)

Named after Pepper Potts. Set all your GitHub repos master branches to be
protected.

You can set which orgs to include and use `--dry-run` to see the
changes before they are actually made. Your user is automatically added to the
repositories it will consider.

Also see [jessfraz/audit](https://github.com/jessfraz/audit) for checking what
collaborators, hooks, deploy keys, and protected branched you have added on
all your GitHub repositories.

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
[OK] jessconf/jessconf:master is already protected
[OK] jessfraz/.vim:master is already protected
[OK] jessfraz/anonymail:master is already protected
[OK] jessfraz/apk-file:master is already protected
[UPDATE] jessfraz/certok:master will be changed to protected
...
[OK] jessfraz/weather:master is already protected
[OK] jessfraz/ykpiv:master is already protected
[OK] maintainerati/wontfix-cabal-site:master is already protected
```
