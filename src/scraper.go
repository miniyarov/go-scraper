package scraper

import (
	"context"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

type Scraper struct {
	baseUrl          *url.URL
	queue            chan string
	scraped          map[string]bool
	sema             chan struct{}
	count            *Counter
	wg               *sync.WaitGroup
	requestListeners []func(*http.Request)
	scrapeListeners  []func(*html.Node)
	*Config
}

func New(baseUrl string, config Config) (*Scraper, error) {
	base, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}

	q := make(chan string, config.Concurrency+1)
	sema := make(chan struct{}, 1)

	return &Scraper{
		baseUrl:          base,
		queue:            q,
		scraped:          map[string]bool{},
		sema:             sema,
		count:            &Counter{},
		wg:               &sync.WaitGroup{},
		requestListeners: nil,
		scrapeListeners:  nil,
		Config:           &config,
	}, nil
}

func NewWithDefaults(baseUrl string) (*Scraper, error) {
	c := ConfigDefaults
	return New(baseUrl, c)
}

func (s *Scraper) EnqueueUrl(href string) {
	u, err := url.Parse(href)
	if err != nil {
		s.print(err.Error())
		return
	}

	if len(u.Host) == 0 {
		u.Scheme = s.baseUrl.Scheme
		u.Host = s.baseUrl.Host
	}

	if s.baseUrl.Host != u.Host {
		return
	}

	go func(u *url.URL) {
		s.queue <- clearSlashes(s.baseUrl.String()) + "/" + url.PathEscape(clearSlashes(u.RequestURI()))
	}(u)
}

func (s *Scraper) Run() {
	ctx, cancelFn := context.WithTimeout(context.Background(), s.Timeout)
	defer cancelFn()

	s.RunWithContext(ctx)
}

func (s *Scraper) worker(ctx context.Context, cancel <-chan struct{}) {
	defer s.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-cancel:
			return
		case u := <-s.queue:
			s.sema <- struct{}{}
			if _, ok := s.scraped[u]; !ok && s.count.Read() < s.MaxRequestCount {
				s.scraped[u] = true
				<-s.sema
				s.count.Increment()
				s.doRequest(ctx, u)
			} else {
				<-s.sema
			}
		}
	}
}

func (s *Scraper) RunWithContext(ctx context.Context) {
	s.queue <- clearSlashes(s.baseUrl.String()) + "/"

worker:
	cancel := make(chan struct{})

	var i uint32
	for i = 0; i < s.Concurrency-1; i++ {
		s.wg.Add(1)

		go s.worker(ctx, cancel)
	}

	close(cancel)
	s.wg.Wait()

	if len(s.queue) > 0 {
		goto worker
	}
}

func (s *Scraper) OnUrlScraped(listener func(*html.Node)) {
	s.scrapeListeners = append(s.scrapeListeners, listener)
}

func (s *Scraper) OnRequest(listener func(*http.Request)) {
	s.requestListeners = append(s.requestListeners, listener)
}

func (s *Scraper) doRequest(ctx context.Context, url string) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return
	}

	s.doOnRequest(req)
	resp, err := s.Client.Do(req)
	if err != nil {
		s.print(fmt.Sprintf("unable to visit url %s, reason: %s\n", url, err.Error()))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		s.doPostRequest(resp.Body)
	}
}

func (s *Scraper) doOnRequest(req *http.Request) {
	for _, listener := range s.requestListeners {
		listener(req)
	}
}

func (s *Scraper) doPostRequest(body io.Reader) {
	document, err := html.Parse(body)
	if err != nil {
		s.print(err.Error())
		return
	}

	for _, listener := range s.scrapeListeners {
		listener(document)
	}
}

func (s *Scraper) print(str string) {
	if s.Debug {
		fmt.Println("[debug] " + str)
	}
}

func clearSlashes(path string) string {
	parts := strings.Split(path, "#")
	path = strings.TrimRight(parts[0], "/")
	return strings.TrimLeft(path, "/")
}
