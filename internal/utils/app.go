package utils

import (
	"os"

	shell "github.com/ipfs/go-ipfs-api"
)

type App struct {
	IPFS_URL      string
	ipfsShell     *shell.Shell
	IpfsRepoPath  string
	ThisGitRepo   string
	ThisGitRemote string
	ref2hash      map[string]string
}

func GetApp() *App {
	app := &App{
		IPFS_URL:    "localhost:5001",
		ThisGitRepo: os.Getenv("GIT_DIR"),
		ref2hash:    make(map[string]string),
	}
	app.ipfsShell = shell.NewShell(app.IPFS_URL)
	return app
}
