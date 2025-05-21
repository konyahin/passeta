package main

import (
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"konyahin.xyz/passeta/view"
)

var passwords []string

func main() {
	passDir := os.Getenv("PASSWORD_STORE_DIR")
	if passDir == "" {
		log.Fatal("Can't find pass directory! $PASSWORD_STORE_DIR is empty.")
	}

	filepath.WalkDir(passDir, func(path string, d fs.DirEntry, err error) error {
		switch {
		case path == passDir:
			return nil
		case d.IsDir() && strings.Contains(path, "/."):
			return fs.SkipDir
		case d.IsDir() || strings.Contains(path, "/."):
			return nil
		}

		name, _ := strings.CutPrefix(path, passDir)
		name, _ = strings.CutSuffix(name, ".gpg")
		passwords = append(passwords, name)
		return nil
	})

	view.SetItems(passwords)

	view.SetOnSearchCallback(func(search string) {
		searchs := strings.Split(search, " ")
		items := make([]string, 0, len(passwords))
	pass:
		for _, password := range passwords {
			for _, searchPart := range searchs {
				if !strings.Contains(password, searchPart) {
					continue pass
				}
			}
			items = append(items, password)
		}
		view.SetItems(items)
	})

	view.SetOnDoneCallback(func(name string, new bool) {
		if new {
			passwordChannel := make(chan string)
			view.RequestPassword(passwordChannel)
			go func() {
				password := <-passwordChannel

				cmd := exec.Command("pass", "insert", "-e", name)
				cmd.Stdin = strings.NewReader(password)
				output, err := cmd.CombinedOutput()
				if err != nil {
					view.SetStatusErrorString("error: " + string(output))
				} else {
					view.SetStatusString(name + " created")
					passwords = append(passwords, name)
					view.SetItems(passwords)
				}
				view.Redraw()
			}()
			return
		}

		cmd := exec.Command("pass", "-c", name)
		output, err := cmd.CombinedOutput()
		if err != nil {
			view.SetStatusErrorString("error: " + string(output))
		} else {
			view.SetStatusString(name + " in your clipboard")
		}
	})

	view.SetOnDeleteCallback(func(i int) {
		name := passwords[i]
		cmd := exec.Command("pass", "delete", "-rf", name)
		output, err := cmd.CombinedOutput()
		if err != nil {
			view.SetStatusErrorString("error: " + string(output))
		} else {
			view.SetStatusString(name + " was deleted")
		}

		passwords = slices.Delete(passwords, i, i+1)
		view.SetItems(passwords)
	})

	view.Run()
}
