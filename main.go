package main

import (
	"flag"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
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

	client := &http.DefaultClient
	fmt.Print("Creating slug...")
	res, _ := client.Do(createSlug())
	bod, _ := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	slugJson := &Slug{}
	json.Unmarshal(bod, &slugJson)
	fmt.Println("done")

	fmt.Print("Creating slug archive...")
	tarFile := tarGz(strings.TrimRight(bp.env.buildDir, "/"))
	fmt.Printf("done (%s)\n", tarFile.Name())

	fmt.Print("Uploading slug...")
	res, err = client.Do(putSlug(slugJson, tarFile))
	defer tarFile.Close()
	fmt.Println("done")
	if err != nil {
		panic(err)
	}

	fmt.Print("Releasing slug...")
	res, _ = client.Do(createRelease(slugJson))
	bod, _ = ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	releaseJson := &Release{}
	json.Unmarshal(bod, &releaseJson)
	fmt.Printf("done (v%d)\n", releaseJson.Version)
}
