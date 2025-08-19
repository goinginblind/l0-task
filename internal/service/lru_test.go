package service

import (
	"fmt"
	"testing"

	"github.com/goinginblind/l0-task/internal/domain"
	"github.com/goinginblind/l0-task/internal/pkg/sizeof"
	"github.com/stretchr/testify/assert"
)

func TestNewLRUCache(t *testing.T) {
	cache := NewLRUCache(10, 1024)

	assert.NotNil(t, cache)
	assert.Equal(t, 10, cache.entryCountCap)
	assert.Equal(t, 1024, cache.entrySizeCap)
	assert.NotNil(t, cache.items)
	assert.NotNil(t, cache.evictList)
	assert.Equal(t, 0, cache.currEntryCount)
}

func TestLRUCache_InsertAndGet(t *testing.T) {
	cache := NewLRUCache(2, 1024) // Max 2 entries

	order1 := &domain.Order{OrderUID: "uid1"}
	order2 := &domain.Order{OrderUID: "uid2"}
	order3 := &domain.Order{OrderUID: "uid3"}

	// Insert order1
	cache.Insert(order1)
	val, ok := cache.Get("uid1")
	assert.True(t, ok)
	assert.Equal(t, order1, val)
	assert.Equal(t, 1, cache.currEntryCount)

	// Insert order2
	cache.Insert(order2)
	val, ok = cache.Get("uid2")
	assert.True(t, ok)
	assert.Equal(t, order2, val)
	assert.Equal(t, 2, cache.currEntryCount)

	// Insert order3 (should evict order1)
	cache.Insert(order3)
	val, ok = cache.Get("uid3")
	assert.True(t, ok)
	assert.Equal(t, order3, val)
	assert.Equal(t, 2, cache.currEntryCount)

	// order1 should be evicted
	_, ok = cache.Get("uid1")
	assert.False(t, ok)

	// order2 should still be there
	val, ok = cache.Get("uid2")
	assert.True(t, ok)
	assert.Equal(t, order2, val)

	// Access order2 again to make it most recently used
	cache.Get("uid2")

	// Insert order1 again (should evict order3)
	cache.Insert(order1)
	_, ok = cache.Get("uid3")
	assert.False(t, ok)
	val, ok = cache.Get("uid1")
	assert.True(t, ok)
	assert.Equal(t, order1, val)
	val, ok = cache.Get("uid2")
	assert.True(t, ok)
	assert.Equal(t, order2, val)
}

func TestLRUCache_SizeCap(t *testing.T) {
	cache := NewLRUCache(10, 470) // Max 470 bytes

	// Create an order that is larger than the size cap
	largeOrder := &domain.Order{OrderUID: "large", TrackNumber: "thisisalongtracknumberthatwillmaketheordersizelargerthan10bytes"}
	fmt.Printf("Size of largeOrder: %d\n", sizeof.SizeOf(largeOrder))

	cache.Insert(largeOrder)

	_, ok := cache.Get("large")
	assert.False(t, ok, "Large order should not be inserted due to size cap")
	assert.Equal(t, 0, cache.currEntryCount)

	// Create an order that fits within the size cap
	smallOrder := &domain.Order{OrderUID: "small", TrackNumber: "short"}
	fmt.Printf("Size of smallOrder: %d\n", sizeof.SizeOf(smallOrder))
	cache.Insert(smallOrder)
	val, ok := cache.Get("small")
	assert.True(t, ok)
	assert.Equal(t, smallOrder, val)
	assert.Equal(t, 1, cache.currEntryCount)
}

func TestLRUCache_UpdateExisting(t *testing.T) {
	cache := NewLRUCache(2, 1024)

	order1 := &domain.Order{OrderUID: "uid1", TrackNumber: "track1"}
	cache.Insert(order1)

	updatedOrder1 := &domain.Order{OrderUID: "uid1", TrackNumber: "track1_updated"}
	cache.Insert(updatedOrder1)

	val, ok := cache.Get("uid1")
	assert.True(t, ok)
	assert.Equal(t, updatedOrder1, val)
	assert.Equal(t, 1, cache.currEntryCount)
}

func TestLRUCache_GetNonExistent(t *testing.T) {
	cache := NewLRUCache(2, 1024)

	_, ok := cache.Get("nonexistent")
	assert.False(t, ok)
}
