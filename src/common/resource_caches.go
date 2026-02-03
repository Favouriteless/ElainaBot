package common

import (
	"sync"
)

var RoleCache = CreateCache[Snowflake, Role](10)
var MessageCache = CreateCache[Snowflake, Message](50)
var ChannelCache = CreateCache[Snowflake, Channel](20)
var GuildCache = CreateCache[Snowflake, Guild](3)
var GuildMemberCache = CreateCache[Snowflake, GuildMember](10)

// LRUCache stores the last N objects passed to it in a fixed-size slice. Both read and write run in O(1),
// implemented as a doubly linked list with a hash table for fast lookups.
type LRUCache[K comparable, V any] struct {
	head     *cacheNode[K, V]
	tail     *cacheNode[K, V]
	index    map[K]*cacheNode[K, V]
	length   int
	capacity int
	mutex    sync.RWMutex
}

type cacheNode[K comparable, V any] struct {
	key   K
	value V
	next  *cacheNode[K, V]
	prev  *cacheNode[K, V]
}

// pop this node out of the list, stitching the two adjacent nodes together
func (n *cacheNode[K, V]) pop() {
	if n.prev != nil {
		n.prev.next = n.next
	}
	if n.next != nil {
		n.next.prev = n.prev
	}
	n.prev = nil
	n.next = nil
}

// CreateCache initializes a LRUCache with the given capacity and returns a pointer to it
func CreateCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	return &LRUCache[K, V]{
		index:    make(map[K]*cacheNode[K, V]),
		capacity: capacity,
	}
}

func (c *LRUCache[K, V]) Get(key K) *V {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if node, exists := c.index[key]; exists {
		v := node.value

		if node != c.head { // If the element isn't already first, pop it out and move it to first place
			if node == c.tail {
				c.tail = node.prev
			}
			node.pop()

			node.next = c.head
			if c.head != nil {
				c.head.prev = node
			}
			c.head = node
		}

		return &v // Return pointer to a copy-- the cache entry value should not be modified
	}
	return nil
}

func (c *LRUCache[K, V]) Add(key K, value V) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	node := &cacheNode[K, V]{key: key, value: value}
	c.index[key] = node
	c.length++

	if c.length == 1 { // First element added is the tail
		c.tail = node
	}

	node.next = c.head // Set head to new value and update the backwards pointer
	if c.head != nil {
		c.head.prev = node
	}
	c.head = node

	if c.length > c.capacity { // Last node is discarded when over capacity
		c.length--
		delete(c.index, c.tail.key)
		c.tail = c.tail.prev
		c.tail.next = nil
	}
}

// Invalidate discards the node belonging to key and patches the two adjacent nodes together
func (c *LRUCache[K, V]) Invalidate(key K) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.length == 0 {
		return
	}

	if node, exists := c.index[key]; exists {
		delete(c.index, node.key)
		node.pop()
		c.length--

		if node == c.head { // Element can be both the head and tail simultaneously if length was 1
			c.head = node.next
		}
		if node == c.tail {
			c.tail = node.prev
		}
	}
}

// Update replaces the value mapped to key if it exists, otherwise nothing will happen.
func (c *LRUCache[K, V]) Update(key K, value V) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if node, exists := c.index[key]; exists {
		node.value = value
	}
}
