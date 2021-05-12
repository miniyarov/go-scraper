// +build integration

package go_scraper

import (
	scraper "github.com/miniyarov/go-scraper/src"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html"
	"net/http"
	urlpkg "net/url"
	"os"
	"testing"
)

func TestScraper(t *testing.T) {
	url, ok := os.LookupEnv("SCRAPE_URL")
	if !ok {
		panic("Environment variable SCRAPE_URL must be set")
	}

	s, err := scraper.NewWithDefaults(url)

	assert.NoError(t, err)

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

	h, err := urlpkg.Parse(url)

	assert.NoError(t, err)

	s.OnRequest(func(r *http.Request) {
		assert.Equal(t, h.Host, r.URL.Host)
	})

	s.Run()
}
