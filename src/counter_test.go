package scraper_test

import (
	scraper "github.com/miniyarov/go-scraper/src"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCounter(t *testing.T) {
	c := scraper.Counter{}

	assert.Equal(t, uint32(0), c.Read())

	c.Increment()

	assert.Equal(t, uint32(1), c.Read())
}
