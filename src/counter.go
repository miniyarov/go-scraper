package scraper

import "sync/atomic"

type Counter struct {
	count uint32
}

func (c *Counter) Increment() {
	atomic.AddUint32(&c.count, 1)
}

func (c *Counter) Read() uint32 {
	return atomic.LoadUint32(&c.count)
}
