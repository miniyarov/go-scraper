package scraper_test

import (
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	scraper "github.com/miniyarov/go-scraper/src"
)

func TestNew(t *testing.T) {
	s, err := scraper.NewWithDefaults("\t")

	assert.EqualError(t, err, "parse \"\\t\": net/url: invalid control character in URL")

	s, err = scraper.NewWithDefaults("http://testurl")

	assert.NoError(t, err)
	assert.IsType(t, &scraper.Scraper{}, s)
}

func TestScraper(t *testing.T) {
	testcases := [...]struct {
		name      string
		handlerFn func(http.ResponseWriter, *http.Request)
		scraperFn func(*testing.T, string, scraper.Config)
	}{
		{
			name: "Scraper Run",
			handlerFn: func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(200)
				rw.Write([]byte(`<a href="test1">test1</a>`))
			},
			scraperFn: func(t *testing.T, u string, c scraper.Config) {
				s, err := scraper.New(u, c)

				assert.NoError(t, err)

				s.OnUrlScraped(func(doc *html.Node) {
					var f func(*html.Node)
					f = func(n *html.Node) {
						if n.Type == html.ElementNode && n.Data == "a" {
							for _, a := range n.Attr {
								if a.Key == "href" {
									assert.Equal(t, "test1", a.Val)
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
					assert.Equal(t, u+"/", r.URL.String())
				})

				s.Run()
			},
		},
		{
			name: "Scraper Run Non-OK Status Code",
			handlerFn: func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(404)
			},
			scraperFn: func(t *testing.T, u string, c scraper.Config) {
				s, err := scraper.New(u, c)

				assert.NoError(t, err)

				s.Run()
			},
		},
		{
			name: "Scraper Run Invalid Parsed Url",
			handlerFn: func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(200)
			},
			scraperFn: func(t *testing.T, u string, c scraper.Config) {
				s, err := scraper.New(u, c)

				assert.NoError(t, err)

				s.OnUrlScraped(func(doc *html.Node) {
					s.EnqueueUrl("\t")
				})

				s.Run()
			},
		},
		{
			name: "Scraper Run Non-Host Matched Domain",
			handlerFn: func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(200)
				rw.Write([]byte(`<a href="http://anotherdomain.com">anotherdomain</a>`))
			},
			scraperFn: func(t *testing.T, u string, c scraper.Config) {
				s, err := scraper.New(u, c)

				assert.NoError(t, err)

				s.OnUrlScraped(func(doc *html.Node) {
					var f func(*html.Node)
					f = func(n *html.Node) {
						if n.Type == html.ElementNode && n.Data == "a" {
							for _, a := range n.Attr {
								if a.Key == "href" {
									assert.Equal(t, "http://anotherdomain.com", a.Val)
								}
							}
						}

						for c := n.FirstChild; c != nil; c = c.NextSibling {
							f(c)
						}
					}

					f(doc)
				})

				s.Run()
			},
		},
		{
			name: "Scraper Run Timeout",
			handlerFn: func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(200)
			},
			scraperFn: func(t *testing.T, u string, c scraper.Config) {
				c.Timeout = time.Nanosecond
				s, err := scraper.New(u, c)

				assert.NoError(t, err)

				s.Run()
			},
		},
		{
			name: "Scraper Run Do Not Scrape Scraped Domain",
			handlerFn: func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(200)
			},
			scraperFn: func(t *testing.T, u string, c scraper.Config) {
				s, err := scraper.New(u, c)

				assert.NoError(t, err)

				s.OnUrlScraped(func(doc *html.Node) {
					s.EnqueueUrl(u)
				})

				s.Run()
			},
		},
		{
			name: "Scraper Run Max Request Count",
			handlerFn: func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(200)
			},
			scraperFn: func(t *testing.T, u string, c scraper.Config) {
				c.MaxRequestCount = 1
				s, err := scraper.New(u, c)

				s.OnUrlScraped(func(doc *html.Node) {
					for i := 0; i < 100; i++ {
						s.EnqueueUrl("1")
					}
				})

				assert.NoError(t, err)

				s.Run()
			},
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(testcase.handlerFn))
			defer server.Close()

			c := scraper.Config{
				MaxRequestCount: 5,
				Concurrency:     5,
				Timeout:         5 * time.Second,
				Debug:           false,
				Client:          server.Client(),
			}

			testcase.scraperFn(t, server.URL, c)
		})
	}
}
