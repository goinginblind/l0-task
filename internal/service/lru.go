package service

import (
	"container/list"
	"sync"

	"github.com/goinginblind/l0-task/internal/domain"
	"github.com/goinginblind/l0-task/internal/pkg/sizeof"
)

// cacheEntry is a container that stores the order and its key, its used by cache
type cacheEntry struct {
	key   string
	value *domain.Order
}

type LRUCache struct {
	mu             sync.RWMutex
	entryCountCap  int // cap max amount of entries in bytes
	entrySizeCap   int // cap of single entry size
	currEntryCount int
	items          map[string]*list.Element
	evictList      *list.List
}

func NewLRUCache(entryCountCap, entrySizeCap int) *LRUCache {
	return &LRUCache{
		entryCountCap: entryCountCap,
		entrySizeCap:  entrySizeCap,
		items:         make(map[string]*list.Element),
		evictList:     list.New(),
	}
}

func (c *LRUCache) Get(key string) (*domain.Order, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.items[key]
	if ok {
		c.evictList.MoveToFront(elem)
		return elem.Value.(*cacheEntry).value, true
	}

	return nil, false
}

func (c *LRUCache) Insert(value *domain.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := value.OrderUID
	elem, ok := c.items[key]
	if !ok {
		if sizeof.SizeOf(value) > c.entrySizeCap {
			return
		}
		entry := &cacheEntry{key, value}
		elem := c.evictList.PushFront(entry)
		c.items[key] = elem
		c.currEntryCount++
	} else {
		// TODO: order irl won't be immutable, but for now we will imagine that they are...
		// so no accounting for size change, just move it as LRU
		c.evictList.MoveToFront(elem)
		elem.Value.(*cacheEntry).value = value
	}

	// Evict oldest items if cache amount cap hit
	for c.currEntryCount > c.entryCountCap {
		c.removeOldest()
	}
}

// removeOldest is a helper function which deletes the LRU entry
// from the cache, pops the linked list from the back
func (c *LRUCache) removeOldest() {
	elem := c.evictList.Back()
	if elem != nil {
		entry := c.evictList.Remove(elem).(*cacheEntry)
		delete(c.items, entry.key)
		c.currEntryCount--
	}
}
