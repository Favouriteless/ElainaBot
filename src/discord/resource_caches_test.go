package discord

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests that ResourceCache can lookup and push out old items correctly. I realize this is not a great test, but it is
// good enough for showing that the get/update/add behaviour works as expected
func TestResourceCache(t *testing.T) {
	const size byte = 20
	cache := CreateCache[byte](size)

	for i := byte(0); i < size; i++ {
		cache.Add(Snowflake(i), i)
	}

	// TEST CASE: Access count tie-- the first tied value (19, 19) should be replaced by (42, 42)
	cache.Add(Snowflake(42), 42)
	require.Nil(t, cache.Get(Snowflake(19)))
	assert.Equal(t, byte(42), *cache.Get(42))

	for i := byte(0); i < size-1; i++ { // We access every item EXCEPT (19, 19)
		cache.Get(Snowflake(i))
	}

	// TEST CASE: The least accessed item (0, 0) at index 19 should be replaced by (244, 244)
	cache.Add(Snowflake(244), 244)
	require.Nil(t, cache.Get(Snowflake(0)))
	assert.Equal(t, byte(244), cache.values[19])

	// TEST CASE: Update should update the existing value (13, 13) to (13, 17)
	cache.Update(Snowflake(13), 17)
	assert.Equal(t, byte(17), *cache.Get(Snowflake(13)))
}
