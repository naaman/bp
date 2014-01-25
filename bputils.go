package main

import (
	"code.google.com/p/go-netrc/netrc"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

const executableFlag = 0111

func executable(filemode os.FileMode) bool {
	e := filemode & executableFlag
	return e != 0
}

func recursiveCopy(from string, to string) filepath.WalkFunc {
	toPath := to + "/"

	return func(path string, info os.FileInfo, err error) error {
		src, err := os.Open(path)

		if err != nil {
			return err
		}

		defer src.Close()

		dstPath := strings.Replace(path, from, toPath, 1)

		if info.IsDir() {
			os.MkdirAll(dstPath, info.Mode())
		} else {
			dst, err := os.Create(dstPath)
			if err != nil {
				return err
			}
			defer dst.Close()
			_, err = io.Copy(dst, src)
			return err
		}

		return nil
	}
}

func scriptIsValid(script string) bool {
	scriptStat, err := os.Stat(script)
	if err != nil || !executable(scriptStat.Mode()) {
		return false
	}
	return true
}

func netrcApiKey() string {
	if u, err := user.Current(); err == nil {
		netrcPath := u.HomeDir + "/.netrc"
		if _, err := os.Stat(netrcPath); err == nil {
			key, _ := netrc.FindMachine(netrcPath, "api.heroku.com")
			if key.Password != "" {
				return key.Password
			}
		}
	}
	return ""
}
