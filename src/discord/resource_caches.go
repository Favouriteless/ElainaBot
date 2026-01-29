package discord

import (
	"sync"
)

var RoleCache = CreateCache[Role](10)
var MessageCache = CreateCache[Message](50)
var ChannelCache = CreateCache[Channel](20)
var GuildCache = CreateCache[Guild](3)
var GuildMemberCache = CreateCache[GuildMember](10)

// ResourceCache stores the last N objects passed to it in a fixed-size slice. Lookups in both direction run in O(1)
type ResourceCache[V any] struct {
	values   []V
	records  []readRecord       // Map of id -> access count. Incremented each total an element is read, decremented after N reads.
	id2index map[Snowflake]byte // Map of obj id -> values index
	index2id []Snowflake        // Map of values index -> obj id
	buffer   []int16            // Circular buffer storing the history of indexes read from
	head     byte
	size     byte
	mutex    sync.RWMutex
}

type readRecord struct {
	count byte  // Number of times the record was read in the last N reads
	total int64 // Time of the last read in nanoseconds
}

// CreateCache initializes a ResourceCache with the given size and returns a pointer to it
func CreateCache[V any](size byte) *ResourceCache[V] {
	out := ResourceCache[V]{
		values:   make([]V, size),
		records:  make([]readRecord, size),
		id2index: make(map[Snowflake]byte, size),
		index2id: make([]Snowflake, size),
		buffer:   make([]int16, size),
		size:     size,
	}
	for i := 0; byte(i) < size; i++ {
		out.buffer[i] = -1 // Buffer gets initialized with -1, meaning no value
	}
	return &out
}

func (c *ResourceCache[V]) Get(key Snowflake) *V {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if i, exists := c.id2index[key]; exists {
		c.records[i].count++ // Update read record of value
		c.records[i].total++

		c.buffer[c.head] = int16(i) // Set the most recent read index to i
		c.head = (c.head + 1) % c.size

		if last := c.buffer[c.head]; last != -1 {
			c.records[last].count-- // Decrement the least recent read
		}

		s := c.values[i]
		return &s // Return pointer to a copy-- the original should not be modified
	}
	return nil
}

func (c *ResourceCache[V]) Add(key Snowflake, value V) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	var i byte
	read := c.records[0]
	for j, v := range c.records { // Find unassigned or least accessed/the oldest record
		if v.total == 0 || v.count < read.count || (v.count == 0 && v.total < read.total) {
			i = byte(j)
			read = v
		}
	}

	delete(c.id2index, c.index2id[i]) // Delete id -> index mapping of the value being replaced

	for j, v := range c.buffer { // Delete increment history
		if v == int16(i) {
			c.buffer[j] = -1
		}
	}

	c.values[i] = value
	c.records[i] = readRecord{count: 1, total: 1}
	c.index2id[i] = key
	c.id2index[key] = i
}

// Invalidate marks a given resource as invalidated and deletes it's id -> index mapping, allowing it to be overwritten.
func (c *ResourceCache[V]) Invalidate(key Snowflake) {
	c.mutex.Lock()
	if i, exists := c.id2index[key]; exists {
		delete(c.id2index, key)
		c.index2id[i] = 0
	}
	c.mutex.Unlock()
}

// Update replaces the value mapped to key if it exists, otherwise nothing will happen.
func (c *ResourceCache[V]) Update(key Snowflake, value V) {
	c.mutex.Lock()
	if i, exists := c.id2index[key]; exists {
		c.values[i] = value
	}
	c.mutex.Unlock()
}
