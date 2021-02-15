package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"sync"
)

var wg sync.WaitGroup

type RepositoryNames struct {
	Repositories []string
}

type ChangeableAttributesStruct struct {
	DeleteEnabled bool
	WriteEnabled  bool
	ReadEnabled   bool
	ListEnabled   bool
}
type Manifest struct {
	Digest               string
	ImageSize            int64
	CreatedTime          string
	LastUpdateTime       string
	Architecture         string
	Os                   string
	MediaType            string
	ConfigMediaType      string
	Tags                 []string
	ChangeableAttributes ChangeableAttributesStruct
}

type RepositoryManifests struct {
	Registry  string
	ImageName string
	Manifests []Manifest
}

type Tag struct {
	Name                 string
	Digest               string
	CreatedTime          string
	LastUpdateTime       string
	Signed               bool
	ChangeableAttributes ChangeableAttributesStruct
}

type RepositoryTags struct {
	Registry  string
	ImageName string
	Tags      []Tag
}

// Implementing the sort functions
type manifestByDate []Manifest

func (s manifestByDate) Len() int {
	return len(s)
}

func (s manifestByDate) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s manifestByDate) Less(i, j int) bool {
	return s[i].CreatedTime < s[j].CreatedTime
}

// Implementing the sort functions
type tagByDate []Tag

func (s tagByDate) Len() int {
	return len(s)
}

func (s tagByDate) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s tagByDate) Less(i, j int) bool {
	return s[i].CreatedTime < s[j].CreatedTime
}

// Dirty util function for checking errors
func checkError(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

// wrapper for executing HTTP requests with basic auth
// returns the status code and the response body as a byte array
func executeRequest(method string, url string, user string, pass string, body io.Reader) (int, []byte) {
	req, err := http.NewRequest(method, url, body)
	checkError(err)
	req.SetBasicAuth(user, pass)

	client := http.Client{}
	response, err := client.Do(req)
	checkError(err)
	defer response.Body.Close()

	resBody, err := ioutil.ReadAll(response.Body)

	return response.StatusCode, resBody
}

func getRepositoryNames(config *Config) (RepositoryNames, error) {
	regUrl := fmt.Sprintf("%s/acr/v1/_catalog", config.registryName)

	if config.verboseLogging {
		log.Printf("Obtaining Repository names from %s\n", regUrl)
	}

	status, repos := executeRequest("GET", regUrl, config.clientId, config.clientSecret, nil)

	if status == 200 {
		var out RepositoryNames
		json.Unmarshal(repos, &out)
		return out, nil
	} else {
		return RepositoryNames{}, fmt.Errorf("HTTP request for Repository Names returned %d", status)
	}
}

func getRepositoryManifests(config *Config, repo string) (RepositoryManifests, error) {
	tagUrl := fmt.Sprintf("%s/acr/v1/%s/_manifests", config.registryName, repo)

	if config.verboseLogging {
		log.Printf("Getting Manifest list for %s\n", repo)
	}

	status, repos := executeRequest("GET", tagUrl, config.clientId, config.clientSecret, nil)

	if status == 200 {
		var out RepositoryManifests
		json.Unmarshal(repos, &out)
		return out, nil
	} else {
		return RepositoryManifests{}, fmt.Errorf("HTTP request for %s Repository Manifests returned %d", repo, status)
	}
}

func deleteManifests(config *Config, repo string, manifests []Manifest, limit int) {
	for index := 0; index < len(manifests)-limit; index++ {
		if config.verboseLogging {
			log.Printf("Deleting Manifest %s...\n", manifests[index].Digest)
		}

		if !config.dryRun {
			tagUrl := fmt.Sprintf("%s/v2/%s/manifests/%s", config.registryName, repo, manifests[index].Digest)
			status, _ := executeRequest("DELETE", tagUrl, config.clientId, config.clientSecret, nil)
			if status != 202 {
				log.Fatalf("HTTP request for Deleting Manifest %s returned %d", manifests[index].Digest, status)
			}
		}
	}
}

func getRepositoryTags(config *Config, repo string) (RepositoryTags, error) {
	tagUrl := fmt.Sprintf("%s/acr/v1/%s/_tags", config.registryName, repo)

	if config.verboseLogging {
		log.Printf("Getting Tags list for %s\n", repo)
	}

	status, repos := executeRequest("GET", tagUrl, config.clientId, config.clientSecret, nil)

	if status == 200 {
		var out RepositoryTags
		json.Unmarshal(repos, &out)
		return out, nil
	} else {
		return RepositoryTags{}, fmt.Errorf("HTTP request for %s Repository Tags returned %d", repo, status)
	}
}

func deleteTags(config *Config, repo string, tags []Tag, limit int) {
	for index := 0; index < len(tags)-limit; index++ {
		if config.verboseLogging {
			log.Printf("Deleting Tag %s...\n", tags[index].Name)
		}

		if !config.dryRun {
			tagUrl := fmt.Sprintf("%s/acr/v1/%s/_tags/%s", config.registryName, repo, tags[index].Name)
			status, _ := executeRequest("DELETE", tagUrl, config.clientId, config.clientSecret, nil)
			if status != 202 {
				log.Fatalf("HTTP request for Deleting Tag %s returned %d", tags[index].Name, status)
			}
		}
	}
}

func processRepo(config *Config, repo string) {
	tags, err := getRepositoryTags(config, repo)
	checkError(err)
	sort.Sort(tagByDate(tags.Tags))

	if config.verboseLogging {
		log.Println("---")
		log.Printf("Repository: %s\n", tags.ImageName)
		for _, v := range tags.Tags {
			log.Printf("Tag %v \t %v\n", v.Name, v.CreatedTime)
		}
	}

	deleteTags(config, tags.ImageName, tags.Tags, config.imageLimit)

	manifests, err := getRepositoryManifests(config, repo)
	checkError(err)
	sort.Sort(manifestByDate(manifests.Manifests))

	if config.verboseLogging {
		log.Println("---")
		log.Printf("Repository: %s\n", manifests.ImageName)
		for _, v := range manifests.Manifests {
			log.Printf("Manifest %v \t %v \t %v\n", v.Digest, v.Tags, v.CreatedTime)
		}
	}

	deleteManifests(config, manifests.ImageName, manifests.Manifests, config.imageLimit)
	wg.Done()
}

func main() {
	c := Config{}
	c.Setup()
	flag.Parse()

	if !c.IsValid() {
		fmt.Println("Invalid input parameter or environment variable usage.\nPlease use the (-h)elp menu.")
		os.Exit(1)
	}

	if c.verboseLogging {
		log.Printf("Config being used: %v\n", c)
	}

	repos, err := getRepositoryNames(&c)
	checkError(err)

	for _, v := range repos.Repositories {
		wg.Add(1)
		go processRepo(&c, v)
	}

	wg.Wait()
	log.Println("Complete")
}
