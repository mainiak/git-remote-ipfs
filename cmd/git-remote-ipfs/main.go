/*
git-remote-helper implements a git-remote helper that uses the ipfs transport.

# TODO

Currently assumes a IPFS Daemon at localhost:5001

Not completed: new Push (issue #2), IPNS, URLs like fs:/ipfs/.. (issue #3), embedded IPFS node

...

	$ git clone ipfs://ipfs/$hash/repo.git
	$ cd repo && make $stuff
	$ git commit -a -m 'done!'
	$ git push origin
	=> clone-able as ipfs://ipfs/$newHash/repo.git

# Links

https://ipfs.io

https://github.com/whyrusleeping/git-ipfs-rehost

https://git-scm.com/docs/gitremote-helpers

https://git-scm.com/book/en/v2/Git-Internals-Plumbing-and-Porcelain

https://git-scm.com/docs/gitrepository-layout

https://git-scm.com/book/en/v2/Git-Internals-Transfer-Protocols
*/
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cryptix/go/logging"
	"github.com/mainiak/git-remote-ipfs/internal/path"
	"github.com/mainiak/git-remote-ipfs/internal/utils"
)

const usageMsg = `usage git-remote-ipfs <repository> [<URL>]
supports:

* ipfs://ipfs/$hash/path..
* ipfs:///ipfs/$hash/path..

`

func usage() {
	fmt.Fprint(os.Stderr, usageMsg)
	os.Exit(2)
}

var (
	errc  chan<- error
	log   logging.Interface
	check = logging.CheckFatal
)

func logFatal(msg string) {
	log.Log("event", "fatal", "msg", msg)
	os.Exit(1)
}

func main() {
	// logging
	logging.SetupLogging(nil)
	log = logging.Logger("git-remote-ipfs")

	app := utils.GetApp()

	// env var and arguments
	if app.thisGitRepo == "" {
		logFatal("could not get GIT_DIR env var")
	}

	if app.thisGitRepo == ".git" {
		cwd, err := os.Getwd()
		logging.CheckFatal(err)
		app.thisGitRepo = filepath.Join(cwd, ".git")
	}

	var u string // repo url
	v := len(os.Args[1:])
	switch v {
	case 2:
		app.thisGitRemote = os.Args[1]
		u = os.Args[2]
	default:
		logFatal(fmt.Sprintf("usage: unknown # of args: %d\n%v", v, os.Args[1:]))
	}

	// parse passed URL
	for _, pref := range []string{"ipfs://ipfs/", "ipfs:///ipfs/"} {
		if strings.HasPrefix(u, pref) {
			u = "/ipfs/" + u[len(pref):]
		}
	}
	p, err := path.ParsePath(u)
	check(err)

	app.ipfsRepoPath = p.String()

	// interrupt / error handling
	go func() {
		check(utils.Interrupt())
	}()

	check(app.SpeakGit(os.Stdin, os.Stdout))
}
