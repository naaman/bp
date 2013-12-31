package main

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

type Buildpack struct {
	detect  string
	compile string
	release string
	env     BuildEnv
}

func NewBuildpack(basedir string) (Buildpack, error) {
	absBasedir, absBasedirErr := filepath.Abs(basedir)
	if absBasedirErr != nil {
		return Buildpack{}, absBasedirErr
	}
	absBindir := absBasedir + "/bin"
	bindir, bindirErr := os.Stat(absBindir)
	if bindirErr != nil || !bindir.IsDir() {
		return Buildpack{}, errors.New("Buildpack does not contain a `bin` directory")
	}
	newBuildpack := Buildpack{
		detect:  absBindir + "/detect",
		compile: absBindir + "/compile",
		release: absBindir + "/release",
	}
	if !newBuildpack.checkScripts() {
		return Buildpack{}, errors.New("Buildpack does not have executable detect, compile, and release scripts.")
	}
	return newBuildpack, nil
}

func (b *Buildpack) checkScripts() bool {
	return scriptIsValid(b.detect) &&
		scriptIsValid(b.compile) &&
		scriptIsValid(b.release)
}

func (b *Buildpack) Run(appdir string) error {
	b.env = newBuildEnv(appdir)

	detectCmd := exec.Command(b.detect, b.env.buildDir)
	if err := execCmd(detectCmd, os.Stdout); err != nil {
		return nil
	}

	compileCmd := exec.Command(b.compile, b.env.buildDir, b.env.cacheDir, b.env.envFile)
	if err := execCmd(compileCmd, os.Stdout); err != nil {
		return err
	}

	releaseCmd := exec.Command(b.release, b.env.buildDir)
	if err := execCmd(releaseCmd, os.Stdout); err != nil {
		return err
	}

	return nil
}

func execCmd(cmd *exec.Cmd, pipe io.Writer) error {
	cmdOut, _ := cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		return err
	}
	io.Copy(pipe, cmdOut)
	err := cmd.Wait()
	return err
}

type BuildEnv struct {
	buildDir string
	cacheDir string
	envFile  string
}

func newBuildEnv(appdir string) BuildEnv {
	buildDir, _ := ioutil.TempDir(os.TempDir(), "build")
	filepath.Walk(appdir, recursiveCopy(appdir, buildDir))
	cacheDir, _ := ioutil.TempDir(os.TempDir(), "cache")

	envFile, _ := ioutil.TempFile(os.TempDir(), ".env")
	return BuildEnv{
		buildDir: buildDir,
		cacheDir: cacheDir,
		envFile:  envFile.Name(),
	}
}
