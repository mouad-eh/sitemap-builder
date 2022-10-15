package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/mouad-eh/html-link-parser/link"
)

const xmlns = "http://www.sitemaps.org/schemas/sitemap/0.9"

type loc struct {
	Value string `xml:"loc"`
}

type urlset struct {
	Urls  []loc  `xml:"url"`
	Xmlns string `xml:"xmlns,attr"`
}

func main() {
	urlFlag := flag.String("url", "https://gophercises.com", "the url that you want to build the sitemap for.")
	maxDepth := flag.Int("depth", 3, "the depth of sitemap you want to build.")
	flag.Parse()

	pages := bfs(*urlFlag, *maxDepth)
	toXml := urlset{
		Xmlns: xmlns,
	}
	for _, page := range pages {
		toXml.Urls = append(toXml.Urls, loc{page})
	}
	fmt.Print(xml.Header)
	enc := xml.NewEncoder(os.Stdout)
	enc.Indent("", "  ")
	err := enc.Encode(toXml)
	if err != nil {
		panic(err)
	}
	fmt.Println()
}

func bfs(urlStr string, depth int) []string {
	seen := make(map[string]struct{})
	// the empty struct is a kind of optimization https://dave.cheney.net/2014/03/25/the-empty-struct
	var q map[string]struct{}
	// you can do type empty struct{}
	nq := map[string]struct{}{
		urlStr: struct{}{},
	}
	for i := 0; i <= depth; i++ {
		q, nq = nq, make(map[string]struct{})
		// and here you do this if you don't have a max depth
		// if len(q) == 0 {break}
		for url, _ := range q {
			if _, ok := seen[url]; ok {
				continue
			}
			seen[url] = struct{}{}
			for _, link := range get(url) {
				// you can do this for optimizing the maxDepth reached
				// if _, ok := seen[link]; ok {
				// 	continue
				// }
				nq[link] = struct{}{}
			}
		}
	}
	ret := make([]string, 0, len(seen))
	for link, _ := range seen {
		ret = append(ret, link)
	}
	return ret
}

func get(urlStr string) []string {
	resp, err := http.Get(urlStr)
	if err != nil {
		panic(err)
		// you can do return []string{}
	}
	defer resp.Body.Close()

	reqUrl := resp.Request.URL // gets you the real url if there is a redirect
	baseUrl := &url.URL{
		Scheme: reqUrl.Scheme,
		Host:   reqUrl.Host,
	}
	base := baseUrl.String()
	return filter(hrefs(resp.Body, base), withPrefix(base))
	// why not passing just withPrefix without calling the function which will be returning a function
}

func filter(links []string, keepfn func(link string) bool) []string {
	var filteredLinks []string
	for _, link := range links {
		if keepfn(link) {
			filteredLinks = append(filteredLinks, link)
		}
	}
	return filteredLinks
}

func hrefs(htmlPage io.Reader, base string) []string {
	links, _ := link.Parse(htmlPage)
	var ret []string
	for _, l := range links {
		switch {
		case strings.HasPrefix(l.Href, "/"):
			ret = append(ret, base+l.Href)
		case strings.HasPrefix(l.Href, "http"):
			ret = append(ret, l.Href)
		}
	}
	return ret
}

func withPrefix(pfx string) func(string) bool {
	return func(link string) bool {
		return strings.HasPrefix(link, pfx)
	}
}
