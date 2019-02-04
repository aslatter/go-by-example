package main

import (
	"fmt"
	"sync"
)

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

func startCrawl(url string, depth int, fetcher Fetcher) chan string {
	j := &jobInfo{
		foundURLs: make(map[string]struct{}),
		results:   make(chan string),
	}
	startWorker(j, url, depth, fetcher)
	return j.results
}

type jobInfo struct {
	foundURLs map[string]struct{}
	jobCount  int
	mux       sync.Mutex
	results   chan string
}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(j *jobInfo, url string, depth int, fetcher Fetcher) {
	defer func() {
		j.mux.Lock()
		defer j.mux.Unlock()
		j.jobCount--
		if j.jobCount == 0 {
			close(j.results)
		}
	}()

	body, urls, err := fetcher.Fetch(url)
	if err != nil {
		j.results <- err.Error()
		return
	}
	j.results <- fmt.Sprintf("found: %s %q", url, body)
	for _, u := range urls {
		startWorker(j, u, depth-1, fetcher)
	}
	return
}

func startWorker(j *jobInfo, url string, depth int, fetcher Fetcher) {
	if depth <= 0 {
		return
	}
	if !checkURL(j, url) {
		return
	}
	j.mux.Lock()
	defer j.mux.Unlock()
	j.jobCount++
	go Crawl(j, url, depth, fetcher)
}

func checkURL(j *jobInfo, url string) bool {
	j.mux.Lock()
	defer j.mux.Unlock()

	_, ok := j.foundURLs[url]
	if ok {
		return false
	}
	j.foundURLs[url] = struct{}{}
	return true
}

func main() {
	for oneResult := range startCrawl("https://golang.org/", 4, fetcher) {
		fmt.Println(oneResult)
	}
}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
	if res, ok := f[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
	"https://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"https://golang.org/pkg/",
			"https://golang.org/cmd/",
		},
	},
	"https://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"https://golang.org/",
			"https://golang.org/cmd/",
			"https://golang.org/pkg/fmt/",
			"https://golang.org/pkg/os/",
		},
	},
	"https://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
	"https://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
}
