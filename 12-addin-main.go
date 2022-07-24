package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/go-github/v32/github"
	"github.com/oam-dev/kubevela/pkg/utils"
	"net/http"
	"net/http/httptest"
	"net/url"
	"proj_test/addon"
)

const (
	// baseURLPath is a non-empty FakeClient.BaseURL path to use during tests,
	// to ensure relative URLs are used for all endpoints. See issue #752.
	baseURLPath = "/repos/fourierr/catalog"
)

func setup() (client *github.Client, mux *http.ServeMux, teardown func()) {
	// mux is the HTTP request multiplexer used with the test server.
	mux = http.NewServeMux()

	apiHandler := http.NewServeMux()
	apiHandler.Handle(baseURLPath+"/", http.StripPrefix(baseURLPath, mux))

	// server is a test HTTP server used to provide mock API responses.
	server := httptest.NewServer(apiHandler)
	server.URL = "https://api.github.com"

	// client is the GitHub client being tested and is
	// configured to use test server.
	client = github.NewClient(nil)
	URL, _ := url.Parse(server.URL + baseURLPath + "/")
	client.BaseURL = URL
	client.UploadURL = URL

	return client, mux, server.Close
}

func main() {
	itemStr := `
  {
    "name": "velaux",
    "path": "addons/velaux",
    "sha": "1000bee83f0fba95fb9855138a610d1ad0ef6d71",
    "size": 0,
    "url": "https://api.github.com/repos/fourierr/catalog/contents/addons/velaux?ref=master",
    "html_url": "https://github.com/fourierr/catalog/tree/master/addons/velaux",
    "git_url": "https://api.github.com/repos/fourierr/catalog/git/trees/1000bee83f0fba95fb9855138a610d1ad0ef6d71",
    "download_url": null,
    "type": "dir",
    "_links": {
      "self": "https://api.github.com/repos/fourierr/catalog/contents/addons/velaux?ref=master",
      "git": "https://api.github.com/repos/fourierr/catalog/git/trees/1000bee83f0fba95fb9855138a610d1ad0ef6d71",
      "html": "https://github.com/fourierr/catalog/tree/master/addons/velaux"
    }
  }
`
	item := &github.RepositoryContent{}
	json.Unmarshal([]byte(itemStr), item)
	gith := &addon.GitHelper{
		Client: &github.Client{},
		Meta: &utils.Content{GithubContent: utils.GithubContent{
			Owner: "o",
			Repo:  "r",
		}},
	}
	var r addon.AsyncReader = &addon.GitReader{gith}

	relativePath := r.RelativePath(item)
	fmt.Println(relativePath)
}
