package main

import (
	"fmt"
	"net/http"

	scraper "github.com/miniyarov/go-scraper/src"
	"golang.org/x/net/html"
)

func main() {
	s, err := scraper.NewWithDefaults("https://tr.wikipedia.com")
	if err != nil {
		panic(err)
	}

	s.OnUrlScraped(func(doc *html.Node) {
		var f func(*html.Node)
		f = func(n *html.Node) {
			if n.Type == html.ElementNode && n.Data == "a" {
				for _, a := range n.Attr {
					if a.Key == "href" {
						s.EnqueueUrl(a.Val)
					}
				}
			}

			for c := n.FirstChild; c != nil; c = c.NextSibling {
				f(c)
			}
		}

		f(doc)
	})

	s.OnRequest(func(r *http.Request) {
		fmt.Println("Scraping URL:", r.URL.String())
	})

	s.Run()
}
