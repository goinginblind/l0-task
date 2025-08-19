# What cache actually is?
```Go
type LRUCache struct {
	mu             sync.RWMutex
	entryCountCap  int // cap max amount of entries 
	entrySizeCap   int // cap of single entry size
	currEntryCount int
	items          map[string]*list.Element
	evictList      *list.List
}
```
Essentially composed of:
- `RWmutex` to allow concurrency safety
- `entryCountCap` is a cap on the amount of total cache entries
- `entrySizeCap` is a cap on the size of one entry made to prevent 'poison pills', that are essentially valid data of sizes bigger than they should be (e.g. a 100000 item order). So this option exists mostly to prevent such things
- `items` holds pointer to the linked lists elements to allow for O(1) access to the middle of it
- `evictList` is a doubly linked list which allows an O(1) inserts and deletes at head and tail

Originally I couldn't decide on how to cap the cache: either by raw memory or by entries, so I decided 'why not try both?' – it allows to cap by entries while also implcitly capping the cache by size too (cache upper memory limit is `~entry count * single entry mem limit`). 
But there are tradeoffs in all of the cases:
* Raw memory cap makes cache prone to storing less entries than desired
* Entry count cap does not allow the exact cache size to be known
* A Hybrid checks entries sizes and can be a bit approximate still

But it is __safer__ and that's why I chose it. Its simple and its safe (and also made me write a separate [DeepSize-calculator function](../internal/pkg/sizeof/calculator.go), so it is what it is!).


### Other Documentation:
* [Database Schema](database.md)
* [Consumer Decision Tree](consumer.md)
* [JSON Validation](validation.md)
* [Errors, Metrics, and DB Health Checks](misc.md)

### Back to [Main README](../README.md)
