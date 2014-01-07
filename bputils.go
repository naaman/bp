package main

import (
	"archive/tar"
	"code.google.com/p/go-netrc/netrc"
	"compress/gzip"
	"io"
	"io/ioutil"
	"log"
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

func handleError(_e error) {
	if _e != nil {
		log.Fatal(_e)
		panic(_e)
	}
}

func targzWalk(dirPath string, tw *tar.Writer) {
	var walkfunc filepath.WalkFunc

	walkfunc = func(path string, fi os.FileInfo, err error) error {
		h, err := tar.FileInfoHeader(fi, "")
		handleError(err)

		h.Name = "./app/" + path

		if fi.Mode()&os.ModeSymlink != 0 {
			linkPath, err := os.Readlink(path)
			handleError(err)
			h.Linkname = linkPath
		}

		err = tw.WriteHeader(h)
		handleError(err)

		if fi.Mode()&os.ModeDir == 0 && fi.Mode()&os.ModeSymlink == 0 {
			fr, err := os.Open(path)
			handleError(err)
			defer fr.Close()

			_, err = io.Copy(tw, fr)
			handleError(err)
		}
		return nil
	}

	filepath.Walk(dirPath, walkfunc)
}

func tarGz(inPath string) *os.File {
	wd, _ := os.Getwd()
	os.Chdir(inPath)
	// file write
	tarFile, _ := ioutil.TempFile("", "slug")
	tarFileName := tarFile.Name() + ".tgz"
	os.Rename(tarFile.Name(), tarFileName)
	fw, err := os.Create(tarFileName)
	handleError(err)

	// gzip write
	gw := gzip.NewWriter(fw)
	defer gw.Close()

	// tar write
	tw := tar.NewWriter(gw)
	defer tw.Close()

	targzWalk(".", tw)

	os.Chdir(wd)
	return fw
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
