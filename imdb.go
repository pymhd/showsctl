package main

import (
	"bytes"
	"fmt"

	"io/ioutil"
	"net/http"
)

const (
	imdbUrl               = "https://www.imdb.com/title/tt"
)

var (
	BeginBlockPattern     = []byte("<h2>Storyline</h2>")
        BeginStoryLinePattern = []byte("<span>")
        EndStoryLinePattern   = []byte("</span>")

)

func getUrlBody(url string) []byte {
	resp, err := http.Get(url)
	must(err)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	must(err)
	return body
}

func getStoryLine(body []byte) []byte {
	start := bytes.Index(body, BeginBlockPattern)
	if start == -1 {
		return []byte("Description Not Found")
	}
	body = body[start:]
	
	s1 := bytes.Index(body, BeginStoryLinePattern)
	s2 := bytes.Index(body, EndStoryLinePattern)
	
	if s1 ==-1 || s2 == -1 {
		return []byte("Description Not Found")
	}
	return body[s1 + len(BeginStoryLinePattern):s2]
}

func getImdbDesc(id int) string {
	url := fmt.Sprintf("%s%d/", imdbUrl, id) 
	
	b := getUrlBody(url)
	story := getStoryLine(b)
	
	return string(story)
}
