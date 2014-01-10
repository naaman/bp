package main

import (
	"encoding/json"
	"fmt"
	"github.com/naaman/pf"
	"net/http"
	"os"
	"strings"
	"time"
	"io/ioutil"
)

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
	slugDir      string
	httpClient   *http.Client
	release      *Release
	tarFile      *os.File
}

type Release struct {
	Version int `json:"version"`
}

func NewSlug(slugDir string) *Slug {
	slugJson := &Slug{}
	slugJson.slugDir = slugDir

	client := &http.DefaultClient
	res, _ := client.Do(slugJson.createSlug())
	bod, _ := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	json.Unmarshal(bod, &slugJson)
	slugJson.httpClient = *client
	slugJson.slugDir = slugDir
	return slugJson
}

func (s *Slug) Archive() {
	s.tarFile = tarGz(strings.TrimRight(s.slugDir, "/"))
}

func (s *Slug) Push() {
	_, err := s.httpClient.Do(s.putSlug())
	defer s.tarFile.Close()
	if err != nil {
		panic(err)
	}
}

func (s *Slug) Release() {
	res, _ := s.httpClient.Do(s.createRelease())
	bod, _ := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	releaseJson := &Release{}
	json.Unmarshal(bod, &releaseJson)
	s.release = releaseJson
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

func (s *Slug) createSlug() *http.Request {
	procTable := new(ProcessTable)
	procTable.ProcessTypes = make(map[string]string)
	procfile := s.parseProcfile()
	for _, e := range procfile.Entries {
		procTable.ProcessTypes[e.Type] = e.Command
	}
	procTableJson, _ := json.Marshal(procTable)
	return herokuPost("slugs", string(procTableJson))
}

func (s *Slug) createRelease() *http.Request {
	slugJson := fmt.Sprintf(`{"slug":"%s"}`, s.Id)
	return herokuPost("releases", slugJson)
}

func (s *Slug) putSlug() *http.Request {
	tarFileStat, err := s.tarFile.Stat()
	tarFile, _ := os.Open(s.tarFile.Name())
	if err != nil {
		panic(err)
	}
	req, _ := http.NewRequest("PUT", s.Blob["put"], tarFile)
	req.ContentLength = tarFileStat.Size()
	return req
}

func (s *Slug) parseProcfile() *pf.Procfile {
	procfileFile, _ := os.Open(s.slugDir + "/Procfile")
	procfile, _ := pf.ParseProcfile(procfileFile)
	return procfile
}
