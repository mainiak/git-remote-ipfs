package utils

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func push(app *App, src, dst string) error {
	var force = strings.HasPrefix(src, "+")
	if force {
		src = src[1:]
	}
	var present []string
	for _, h := range app.ref2hash {
		present = append(present, h)
	}
	// also: track previously pushed branches in 2nd map and extend present with it
	need2push, err := gitListObjects(app, src, present)
	if err != nil {
		return errors.Wrapf(err, "push: git list objects failed %q %v", src, present)
	}
	n := len(need2push)
	type pair struct {
		Sha1  string
		MHash string
		Err   error
	}
	added := make(chan pair)
	objHash2multi := make(map[string]string, n)
	for _, sha1 := range need2push {
		go func(sha1 string) {
			r, err := gitFlattenObject(app, sha1)
			if err != nil {
				added <- pair{Err: errors.Wrapf(err, "gitFlattenObject failed")}
				return
			}
			mhash, err := app.ipfsShell.Add(r)
			if err != nil {
				added <- pair{Err: errors.Wrapf(err, "shell.Add(%s) failed", sha1)}
				return
			}
			added <- pair{Sha1: sha1, MHash: mhash}
		}(sha1)
	}
	for n > 0 {
		select {
		// add timeout?
		case p := <-added:
			if p.Err != nil {
				return p.Err
			}
			log.Debug("sha1", p.Sha1, "mhash", p.MHash, "msg", "added")
			objHash2multi[p.Sha1] = p.MHash
			n--
		}
	}
	root, err := app.ipfsShell.ResolvePath(app.IpfsRepoPath)
	if err != nil {
		return errors.Wrapf(err, "resolvePath(%s) failed", app.IpfsRepoPath)
	}
	for sha1, mhash := range objHash2multi {
		newRoot, err := app.ipfsShell.PatchLink(root, filepath.Join("objects", sha1[:2], sha1[2:]), mhash, true)
		if err != nil {
			return errors.Wrapf(err, "patchLink failed")
		}
		root = newRoot
		log.Debug("newRoot", newRoot, "sha1", sha1, "msg", "updated object")
	}
	srcSha1, err := gitRefHash(app, src)
	if err != nil {
		return errors.Wrapf(err, "gitRefHash(%s) failed", src)
	}
	h, ok := app.ref2hash[dst]
	if !ok {
		return errors.Errorf("writeRef: app.ref2hash entry missing: %s %+v", dst, app.ref2hash)
	}
	isFF := gitIsAncestor(app, h, srcSha1)
	if isFF != nil && !force {
		// TODO: print "non-fast-forward" to git
		return errors.Errorf("non-fast-forward")
	}
	mhash, err := app.ipfsShell.Add(bytes.NewBufferString(fmt.Sprintf("%s\n", srcSha1)))
	if err != nil {
		return errors.Wrapf(err, "shell.Add(%s) failed", srcSha1)
	}
	root, err = app.ipfsShell.PatchLink(root, dst, mhash, true)
	if err != nil {
		// TODO:print "fetch first" to git
		err = errors.Wrapf(err, "patchLink(%s) failed", app.IpfsRepoPath)
		log.Error("err", err, "msg", "shell.PatchLink failed")
		return errors.Errorf("fetch first")
	}
	log.Debug("newRoot", root, "dst", dst, "hash", srcSha1, "msg", "updated ref")
	// invalidate info/refs and HEAD(?)
	// TODO: unclean: need to put other revs, too make a soft git update-server-info maybe
	noInfoRefsHash, err := app.ipfsShell.Patch(root, "rm-link", "info/refs")
	if err == nil {
		log.Debug("newRoot", noInfoRefsHash, "msg", "rm-link'ed info/refs")
		root = noInfoRefsHash
	} else {
		// todo shell.IsNotExists() ?
		log.Error("err", err, "msg", "shell.Patch rm-link info/refs failed - might be okay... TODO")
	}
	newRemoteURL := fmt.Sprintf("ipfs:///ipfs/%s", root)
	updateRepoCMD := exec.Command("git", "remote", "set-url", app.ThisGitRemote, newRemoteURL)
	out, err := updateRepoCMD.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "updating remote url failed\nOut:%s", string(out))
	}
	log.Debug("msg", "remote updated", "address", newRemoteURL)
	return nil
}
