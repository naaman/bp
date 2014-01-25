package main

import (
	"flag"
	"fmt"
	"github.com/naaman/slug"
	"os"
)

var apiKey = flag.String("key", netrcApiKey(), "API key")
var appName = flag.String("app", "", "Heroku App name")
var srcDir = flag.String("src", "", "Source directory")
var bpDir = flag.String("buildpack", "", "Buildpack directory")

func init() {
	flag.Parse()
	for _, f := range []string{*apiKey, *appName, *srcDir, *bpDir} {
		if f == "" {
			flag.Usage()
			os.Exit(1)
		}
	}
}

func main() {
	bp, err := NewBuildpack(*bpDir)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := bp.Run(*srcDir); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Print("Creating slug...")
	herokuSlug := slug.NewSlug(*apiKey, *appName, bp.env.buildDir)
	fmt.Println("done")

	fmt.Print("Creating slug archive...")
	slugFile := herokuSlug.Archive()
	fmt.Printf("done (%s)\n", slugFile.Name())

	fmt.Print("Uploading slug...")
	herokuSlug.Push()
	fmt.Println("done")

	fmt.Print("Releasing slug...")
	release := herokuSlug.Release()
	fmt.Printf("done (v%d)\n", release.Version)
}
