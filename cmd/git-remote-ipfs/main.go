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

	log "github.com/sirupsen/logrus"

	"github.com/mainiak/git-remote-ipfs/internal/path"
	"github.com/mainiak/git-remote-ipfs/internal/utils"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Allow user to toggle level of output
	debug_str, debug_exists := os.LookupEnv("DEBUG_LEVEL")
	if debug_exists && debug_str == "DEBUG" {
		log.SetLevel(log.DebugLevel)
	} else {
		// Only log the warning severity or above.
		log.SetLevel(log.WarnLevel)
	}
}

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
	errc chan<- error
)

func main() {
	var err error
	//log = logging.Logger("git-remote-ipfs")

	app := utils.GetApp()

	// env var and arguments
	if app.thisGitRepo == "" {
		log.Fatal("could not get GIT_DIR env var")
	}

	if app.thisGitRepo == ".git" {
		cwd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		app.thisGitRepo = filepath.Join(cwd, ".git")
	}

	var u string // repo url
	v := len(os.Args[1:])
	switch v {
	case 2:
		app.thisGitRemote = os.Args[1]
		u = os.Args[2]
	default:
		log.Fatal(fmt.Sprintf("usage: unknown # of args: %d\n%v", v, os.Args[1:]))
	}

	// parse passed URL
	for _, pref := range []string{"ipfs://ipfs/", "ipfs:///ipfs/"} {
		if strings.HasPrefix(u, pref) {
			u = "/ipfs/" + u[len(pref):]
		}
	}

	var p path.Path
	p, err = path.ParsePath(u)
	if err != nil {
		panic(err)
	}

	app.ipfsRepoPath = p.String()

	// interrupt / error handling
	go func() {
		if err := utils.Interrupt(); err != nil {
			panic(err)
		}
	}()

	err = app.SpeakGit(os.Stdin, os.Stdout)
	if err != nil {
		panic(err)
	}
}
