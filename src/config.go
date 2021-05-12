package scraper

import (
	"net/http"
	"time"
)

type Config struct {
	MaxRequestCount uint32
	Concurrency     uint32
	Timeout         time.Duration
	Debug           bool
	Client          *http.Client
}

var ConfigDefaults = Config{
	MaxRequestCount: 10,
	Concurrency:     10,
	Timeout:         5 * time.Second,
	Debug:           false,
	Client:          http.DefaultClient,
}
