package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
  "bytes"
  "strconv"
	"time"
  "os"
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
	fmt.Println(string(procTableJson))
	return herokuPost("slugs", string(procTableJson))
}

func createRelease(s *Slug) *http.Request {
  slugJson := fmt.Sprintf(`{"slug":"%s"}`, s.Id)
  fmt.Println(slugJson)
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

	tarGz(tarFile.Name(), strings.TrimRight(bp.env.buildDir, "/"))
  completeTarFile, _ := os.Open(tarFile.Name())
  defer tarFile.Close()
  tarFileStat, _ := completeTarFile.Stat()
  tarFileSize := strconv.FormatInt(tarFileStat.Size(), 10)

	client := http.DefaultClient
	res, _ := client.Do(createSlug())
	bod, _ := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	slugJson := &Slug{}
	json.Unmarshal(bod, &slugJson)
  fmt.Printf("%+v\n", slugJson)
	putUrl := slugJson.Blob["put"]
  
  f, _ := ioutil.ReadAll(completeTarFile)
  fmt.Println(f)
  req, _ := http.NewRequest("PUT", putUrl, bytes.NewReader(f))
  req.Header.Add("Content-Length", tarFileSize)
  fmt.Printf("%+v\n", req)
  res, err = client.Do(req)
  if err != nil {
    panic(err)
  }
  bod, _ = ioutil.ReadAll(res.Body)
  defer res.Body.Close()
  fmt.Printf("%+v\n", res)
  fmt.Println(string(bod))

  res, _ = client.Do(createRelease(slugJson))
  bod, _ = ioutil.ReadAll(res.Body)
  defer res.Body.Close()
  fmt.Println(string(bod))

}
