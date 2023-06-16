# git-remote-ipfs

## About

A 'native' git protocol helper to push and pull git repos from [IPFS](https://ipfs.io).

## Disclamier

Right now this project is being rewritten.  
Use at your own risk.

Original project [by cryptix](https://github.com/cryptix/git-remote-ipfs) is no longer maintained and still somehow(?) works(-ish).

## Installation

`go install github.com/mainiak/git-remote-ipfs/cmd/git-remote-ipfs@latest`

See [![GoDoc](https://godoc.org/github.com/mainiak/git-remote-ipfs?status.svg)](https://godoc.org/github.com/mainiak/git-remote-ipfs) for usage.

## Usage

```
## Yay
git clone ipfs://ipfs/<CID>

## Eh?
git fetch

## Meh?
git push
```

## Uninstall

```
$ rm -v $(which git-remote-ipfs)
```
