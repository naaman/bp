package main

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"launchpad.net/goyaml"
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

	wd, _ := os.Getwd()
	os.Chdir(b.env.buildDir)

	detectCmd := exec.Command(b.detect, b.env.buildDir)
	if err := execCmd(detectCmd, os.Stdout, os.Stderr); err != nil {
		return nil
	}

	compileCmd := exec.Command(b.compile, b.env.buildDir, b.env.cacheDir, b.env.envFile)
	if err := execCmd(compileCmd, os.Stdout, os.Stderr); err != nil {
		return err
	}

	releaseCmd := exec.Command(b.release, b.env.buildDir)
	releaseOut := new(bytes.Buffer)
	if err := execCmd(releaseCmd, releaseOut, os.Stderr); err != nil {
		return err
	}
	writeProcfile(b.env, releaseOut.Bytes())

	os.Chdir(wd)

	return nil
}

func execCmd(cmd *exec.Cmd, stdout io.Writer, stderr io.Writer) error {
	cmdOut, _ := cmd.StdoutPipe()
	cmdErr, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
		return err
	}
	io.Copy(stdout, cmdOut)
	io.Copy(stderr, cmdErr)
	err := cmd.Wait()
	return err
}

func writeProcfile(env BuildEnv, releaseOut []byte) {
	procfile, err := os.Create(env.buildDir + "/Procfile")

	if os.IsExist(err) {
		return
	}

	releaseYaml := make(map[string]map[string]string)
	goyaml.Unmarshal(releaseOut, &releaseYaml)
	if processes, exists := releaseYaml["default_process_types"]; exists {
		for k, v := range processes {
			procfile.WriteString(k + ": " + v)
		}
	}
	procfile.Close()
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
