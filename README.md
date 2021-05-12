## GO Scraper

A simple http scraper client that crawl given starting url and visits each subsequent same-domain URL found.

### Usage

Scraper can be initialized with either default configuration or with custom configuration.

```go
import scraper "github.com/miniyarov/go-scraper/src"

s, err := scraper.NewWithDefaults(url)
// or
c := scraper.Config{
    MaxRequestCount: 5, // Limit the number of requests to the domain
    Concurrency:     5, // The number of concurrent connections to crawl
    Timeout:         5 * time.Second, // Graceful timeout
    Client:          http.DefaultClient, // A custom http client can be supplied for loose coupling
}
s, err := scraper.New(url, c)

```

### Testing

The project include docker compose yaml that triggers `unit tests` and `integration test`

```bash
docker-compose up
```

Docker composer will actually execute tasks defined inside `Makefile`
```bash
make test
make coverage
SCRAPE_URL=https://en.wikipedia.org make integration
```

### Author

[Ulugbek Miniyarov](https://www.linkedin.com/in/miniyarov/)