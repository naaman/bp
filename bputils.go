package main

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"log"
	"os"
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

// Tar a directory from:
// http://stackoverflow.com/questions/13611100/how-to-write-a-directory-not-just-the-files-in-it-to-a-tar-gz-file-in-golang

func handleError(_e error) {
	if _e != nil {
		log.Fatal(_e)
	}
}

func targzWrite(path string, tw *tar.Writer, fi os.FileInfo) {
  h, err := tar.FileInfoHeader(fi, path)
  handleError(err)

	err = tw.WriteHeader(h)
	handleError(err)

  if fi.Mode()&os.ModeSymlink == 0 {
    fr, err := os.Open(path)
    handleError(err)
    defer fr.Close()

    _, err = io.Copy(tw, fr)
    handleError(err)
  }
}

func targzWalk(dirPath string, tw *tar.Writer) {
	filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
      targzWrite(path, tw, info)
		}

		return nil
	})
}

func tarGz(outFilePath string, inPath string) {
  wd, _ := os.Getwd()
  os.Chdir(inPath)
	// file write
	fw, err := os.Create(outFilePath)
	handleError(err)
	defer fw.Close()

	// gzip write
	gw := gzip.NewWriter(fw)
	defer gw.Close()

	// tar write
	tw := tar.NewWriter(gw)
	defer tw.Close()

	targzWalk(".", tw)

  os.Chdir(wd)
}
