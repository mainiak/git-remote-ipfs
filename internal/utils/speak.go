package utils

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
)

// speakGit acts like a git-remote-helper
// see this for more: https://www.kernel.org/pub/software/scm/git/docs/gitremote-helpers.html
func SpeakGit(r io.Reader, w io.Writer) error {
	//debugLog := logging.Logger("git")
	//r = debug.NewReadLogrus(debugLog, r)
	//w = debug.NewWriteLogrus(debugLog, w)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		text := scanner.Text()
		switch {

		case text == "capabilities":
			fmt.Fprintln(w, "fetch")
			fmt.Fprintln(w, "push")
			fmt.Fprintln(w, "")

		case strings.HasPrefix(text, "list"):
			var (
				forPush = strings.Contains(text, "for-push")
				err     error
				head    string
			)
			if err = listInfoRefs(forPush); err == nil { // try .git/info/refs first
				if head, err = listHeadRef(); err != nil {
					return err
				}
			} else { // alternativly iterate over the refs directory like git-remote-dropbox
				if forPush {
					log.Log("msg", "for-push: should be able to push to non existant.. TODO #2")
				}
				log.Log("err", err, "msg", "didn't find info/refs in repo, falling back...")
				if err = listIterateRefs(forPush); err != nil {
					return err
				}
			}
			if len(ref2hash) == 0 {
				return errors.New("did not find _any_ refs...")
			}
			// output
			for ref, hash := range ref2hash {
				if head == "" && strings.HasSuffix(ref, "master") {
					// guessing head if it isnt set
					head = hash
				}
				fmt.Fprintf(w, "%s %s\n", hash, ref)
			}
			fmt.Fprintf(w, "%s HEAD\n", head)
			fmt.Fprintln(w)

		case strings.HasPrefix(text, "fetch "):
			for scanner.Scan() {
				fetchSplit := strings.Split(text, " ")
				if len(fetchSplit) < 2 {
					return errors.Errorf("malformed 'fetch' command. %q", text)
				}
				err := fetchObject(fetchSplit[1])
				if err == nil {
					fmt.Fprintln(w)
					continue
				}
				// TODO isNotExist(err) would be nice here
				//log.Log("sha1", fetchSplit[1], "name", fetchSplit[2], "err", err, "msg", "fetchLooseObject failed, trying packed...")

				err = fetchPackedObject(fetchSplit[1])
				if err != nil {
					return errors.Wrap(err, "fetchPackedObject() failed")
				}
				text = scanner.Text()
				if text == "" {
					break
				}
			}
			fmt.Fprintln(w, "")

		case strings.HasPrefix(text, "push"):
			for scanner.Scan() {
				pushSplit := strings.Split(text, " ")
				if len(pushSplit) < 2 {
					return errors.Errorf("malformed 'push' command. %q", text)
				}
				srcDstSplit := strings.Split(pushSplit[1], ":")
				if len(srcDstSplit) < 2 {
					return errors.Errorf("malformed 'push' command. %q", text)
				}
				src, dst := srcDstSplit[0], srcDstSplit[1]
				f := []interface{}{
					"src", src,
					"dst", dst,
				}
				log.Log(append(f, "msg", "got push"))
				if src == "" {
					fmt.Fprintf(w, "error %s %s\n", dst, "delete remote dst: not supported yet - please open an issue on github")
				} else {
					if err := push(src, dst); err != nil {
						fmt.Fprintf(w, "error %s %s\n", dst, err)
						return err
					}
					fmt.Fprintln(w, "ok", dst)
				}
				text = scanner.Text()
				if text == "" {
					break
				}
			}
			fmt.Fprintln(w, "")

		case text == "":
			break

		default:
			return errors.Errorf("Error: default git speak: %q", text)
		}
	}
	if err := scanner.Err(); err != nil {
		return errors.Wrap(err, "scanner.Err()")
	}
	return nil
}
