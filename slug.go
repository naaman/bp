package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"os"
	"time"
)

var apiKey = flag.String("key", "", "API key")
var appName = flag.String("app", "", "Heroku App name")
var srcDir = flag.String("src", "", "Source directory")
var bpDir = flag.String("buildpack", "", "Buildpack directory")

type ProcessTable struct {
	ProcessTypes map[string]string `json:"process_types"`
}

type Slug struct {
	Blob         map[string]string `json:"blob"`
	Commit       *string           `json:"commit"`
	CreatedAt    time.Time         `json:"created_at"`
	Id           string            `json:"id"`
	ProcessTypes map[string]string `json:"process_types"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

type Release struct {
	Version int `json:"version"`
}

func herokuReq(method string, resource string, body string) *http.Request {
	reqUrl := fmt.Sprintf("https://api.heroku.com/apps/%s/%s", *appName, resource)
	req, _ := http.NewRequest(method, reqUrl, strings.NewReader(body))
	req.Header.Add("Accept", "application/vnd.heroku+json; version=3")
	req.SetBasicAuth("", *apiKey)
	return req
}

func herokuPost(resource string, body string) *http.Request {
	req := herokuReq("POST", resource, body)
	req.Header.Add("Content-Type", "application/json")
	return req
}

func createSlug() *http.Request {
	procTable := &ProcessTable{ProcessTypes: parseProcfile()}
	procTableJson, _ := json.Marshal(procTable)
	return herokuPost("slugs", string(procTableJson))
}

func createRelease(s *Slug) *http.Request {
	slugJson := fmt.Sprintf(`{"slug":"%s"}`, s.Id)
	return herokuPost("releases", slugJson)
}

func parseProcfile() map[string]string {
	procfile, _ := ioutil.ReadFile(*srcDir + "/Procfile")
	lines := strings.Split(string(procfile), "\n")
	entries := make(map[string]string)
	for _, l := range lines {
		entry := strings.SplitN(l, ":", 2)
		if len(entry) == 2 {
			entries[entry[0]] = entry[1]
		}
	}
	return entries
}

func main() {
	flag.Parse()
	bp, err := NewBuildpack(*bpDir)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	exit, err := bp.Run(*srcDir)

	if err != nil {
		fmt.Println(err)
		os.Exit(exit)
	}

	tarFile, _ := ioutil.TempFile(os.TempDir(), "slug")
	fmt.Println(tarFile.Name())

	fmt.Print("Creating slug archive...")
	tarGz(tarFile.Name(), strings.TrimRight(bp.env.buildDir, "/"))
	fmt.Println("done")
	defer tarFile.Close()

	tarFileStat, _ := tarFile.Stat()

	client := &http.DefaultClient
	fmt.Print("Creating slug...")
	res, _ := client.Do(createSlug())
	bod, _ := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	slugJson := &Slug{}
	json.Unmarshal(bod, &slugJson)
	fmt.Println("done")
	putUrl := slugJson.Blob["put"]

	req, _ := http.NewRequest("PUT", putUrl, tarFile)
	req.ContentLength = tarFileStat.Size()

	fmt.Print("Uploading slug...")
	res, err = client.Do(req)
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
