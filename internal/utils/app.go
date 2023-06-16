package utils

import (
	"os"

	shell "github.com/ipfs/go-ipfs-api"
)

type App struct {
	IPFS_URL      string
	ipfsShell     *shell.Shell
	ipfsRepoPath  string
	thisGitRepo   string
	thisGitRemote string
	ref2hash      map[string]string
}

func GetApp() *App {
	app := &App{
		IPFS_URL:    "localhost:5001",
		thisGitRepo: os.Getenv("GIT_DIR"),
		ref2hash:    make(map[string]string),
	}
	app.ipfsShell = shell.NewShell(app.IPFS_URL)
	return app
}
