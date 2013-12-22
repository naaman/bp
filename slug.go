package main

import (
  "fmt"
  "flag"
  "net/http"
  "io/ioutil"
  "strings"
  "encoding/json"
)

var apiKey = flag.String("key", "", "API key")
var appName = flag.String("app", "", "Heroku App name")
var srcDir = flag.String("src", "", "Source directory")

type Slug struct {
  appDir string
  putUrl string
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
  entriesJson, _ := json.Marshal(parseProcfile())
  procJson := fmt.Sprintf(`{"process_types":%s}`, string(entriesJson))
  fmt.Println(procJson)
  return herokuPost("slugs", procJson)
}

func createRelease() *http.Request {
  return herokuPost("releases", `{}`)
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
  client := http.DefaultClient
  res, _ := client.Do(createSlug())
  bod, _ := ioutil.ReadAll(res.Body)
  defer res.Body.Close()
  fmt.Println(string(bod))
}
