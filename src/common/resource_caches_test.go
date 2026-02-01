package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Tests that LRUCache can lookup and push out old items correctly. I realize this is not a great test, but it is
// good enough for showing that the get/update/add behaviour works as expected
func TestResourceCache(t *testing.T) {
	const size int = 20
	cache := CreateCache[int, int](size)

	for i := size; i > 0; i-- {
		cache.Add(i, i)
	}

	// TEST CASE: Values are added up to capacity
	assert.Equal(t, size, cache.length)

	// TEST CASE: Values added to an empty cache are added to head
	assert.Equal(t, 1, cache.head.value)
	assert.Equal(t, 20, cache.tail.value)

	// TEST CASE: Last value gets brought to front when accessed, head and tail are updated correctly
	assert.NotNil(t, cache.Get(20))
	assert.Equal(t, 20, cache.head.value)
	assert.Nil(t, cache.head.prev)
	assert.Nil(t, cache.tail.next)

	// TEST CASE: Middle item gets popped and added to head when accessed
	assert.NotNil(t, cache.Get(10))
	assert.Equal(t, 10, cache.head.value)
	assert.Nil(t, cache.head.prev)
	assert.Nil(t, cache.tail.next)

	// TEST CASE: Update swaps the value of the expected node
	cache.Update(10, 99)
	assert.Equal(t, 99, cache.head.value)

	// TEST CASE: Adding to a full cache pops off the last node
	cache.Add(50, 50)
	assert.Equal(t, size, cache.length)
	assert.NotEqual(t, 19, cache.tail.value)

	// TEST CASE: Invalidating a node stitches neighbouring nodes together and decrements length (10 should be between 50 and 20 here)
	cache.Invalidate(10)
	assert.Nil(t, cache.Get(10))
	assert.Equal(t, 20, cache.head.next.value)
	assert.Equal(t, 50, cache.head.value)
	assert.Equal(t, size-1, cache.length)
}
