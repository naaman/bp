package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

type Buildpack struct {
	detect  string
	compile string
	release string
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

func (b *Buildpack) Run(appdir string) (int, error) {
	buildEnv := newBuildEnv(appdir)

	detectCmd := exec.Command(b.detect, buildEnv.buildDir)
	detectOut, detectErr := detectCmd.CombinedOutput()
	if detectErr != nil {
		return -1, detectErr
	}
	compileCmd := exec.Command(b.compile, buildEnv.buildDir, buildEnv.cacheDir, buildEnv.envFile)
	compileOut, compileErr := compileCmd.CombinedOutput()
	if compileErr != nil {
		return -1, compileErr
	}
	releaseCmd := exec.Command(b.release, buildEnv.buildDir)
	releaseOut, releaseErr := releaseCmd.CombinedOutput()
	if releaseErr != nil {
		return -1, releaseErr
	}
	fmt.Println(string(detectOut))
	fmt.Println(string(compileOut))
	fmt.Println(string(releaseOut))

	return 0, nil
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

func main() {
	baseDir := os.Args[1]
	appDir := os.Args[2]
	baseBuildpack, err := NewBuildpack(baseDir)
	if err != nil {
		fmt.Println(err)
	} else {
		exit, runErr := baseBuildpack.Run(appDir)
		if runErr != nil {
			fmt.Println(runErr)
		} else {
			fmt.Println(exit)
		}
	}
}
