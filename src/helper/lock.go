package helper

import "sync"

type HighGranularityMutex[K Compareable] struct {
	mu     map[K]*sync.Mutex
	s_lock sync.Mutex
}

func (c *HighGranularityMutex[K]) Lock(id K) {
	c.s_lock.Lock()
	// check if s_mu[id] exists
	l, ok := c.mu[id]
	if !ok || l == nil {
		c.mu[id] = &sync.Mutex{}
	}
	l = c.mu[id]
	c.s_lock.Unlock()
	l.Lock()
}

func (c *HighGranularityMutex[K]) Unlock(id K) {
	c.s_lock.Lock()
	// check if s_mu[id] exists
	l, ok := c.mu[id]
	if !ok || l == nil {
		c.s_lock.Unlock()
		return
	}
	l.Unlock()
	delete(c.mu, id)
	c.s_lock.Unlock()
}
