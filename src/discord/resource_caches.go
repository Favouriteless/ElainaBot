package discord

import "sync"

var RoleCache = CreateCache[Role](10)
var MessageCache = CreateCache[Message](50)
var ChannelCache = CreateCache[Channel](20)

// ResourceCache stores the last N objects passed to it in a fixed-size slice. Lookups in both direction run in O(1)
type ResourceCache[V any] struct {
	values   []V
	id2index map[Snowflake]uint16
	index2id map[uint16]Snowflake
	head     uint16
	mutex    sync.RWMutex
}

// CreateCache initializes a ResourceCache with the given size and returns a pointer to it
func CreateCache[V any](size uint16) *ResourceCache[V] {
	return &ResourceCache[V]{
		values:   make([]V, size),
		id2index: make(map[Snowflake]uint16, size),
		index2id: make(map[uint16]Snowflake, size),
	}
}

func (c *ResourceCache[V]) Get(key Snowflake) *V {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if v, exists := c.id2index[key]; exists {
		val := c.values[v] // Return a pointer to this copy so we don't let the original value get modified
		return &val
	}
	return nil
}

func (c *ResourceCache[V]) Add(key Snowflake, value V) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if id, exists := c.index2id[c.head]; exists {
		delete(c.id2index, id) // Clear id -> index entry if there's already an entry in the cache
	}

	c.index2id[c.head] = key
	c.id2index[key] = c.head
	c.values[c.head] = value
	c.head++
	if c.head >= uint16(len(c.values)) {
		c.head = 0
	}
}

// Delete removes a given resource from the cache, it not immediately clear space in the cache. The value will only be
// overwritten the next time the head passes by its index
func (c *ResourceCache[V]) Invalidate(key Snowflake) {
	c.mutex.Lock()
	delete(c.id2index, key)
	c.mutex.Unlock()
}

func (c *ResourceCache[V]) Update(key Snowflake, value V) {
	c.mutex.Lock()
	if _, exists := c.id2index[key]; exists {
		c.values[c.head] = value
	}
	c.mutex.Unlock()
}
