# Sentinel errors
How would a worker notice if the data is bad? And more importantly why its bad? Is even the data bad?..
That's why the store package contains so-called sentinel errors:
``` Go
var (
	ErrNotFound = errors.New("no such record exists")
	ErrAlreadyExists = errors.New("record already exists")
	ErrConnectionFailed = errors.New("connection to the database failed")
)
```
The errors just bubble up nicely and the upper layers are not actually concerned with an exact type of errors, so even if the storage changes later (e.g. unique_violation in psql is `23505`, but mysql has a `1062` code for the same thing). Kinda decouples the store and the upper layers. An example use of it is when it 'bubbles up' to the worker, it instantly marks the db as unhealthy using the HealthChecker

# DBHealthChecker
Is a simple pinger with a configurable interval and a configurable timeout in a single ping. It runs all the time, but is activated early by a worker. Initially I had an idea of it running only when the db is found to be down by a worker, but it was too complex to manage it, thus `MarkUnhealthy()` is a sort of legacy thing, that finds its use when the healthchecker itself has a big interval (e.g. 15s) to prevent partitions from Kafka to be dispatched unneccesarily.
The health package can be found [here](../internal/pkg/health/dbhealth.go)!

# Monitoring
The service utilises Prometheus and Grafana, the metrics are pretty simple because:
* I actually (almost) never did any monitoring/metrics scraping before.
* Time.

So the metrics are:
* **HTTP total requsts** and **HTTP request duration** because might be useful for maybe data engeneers (possibly?(maybe..?))
* **Consumer processed count** is the amount of messages (of 3 types: valid, invalid, errors) processed total
* **Cache hits and cache misses** to maybe be able to adjust cache max entries or max size of a single entry according to the needs of usage. Same for the **cache response time**.
* And **Is db up?**, **Database response time** and **Database transient errors** to maybe ease the load if the db starts to hiccup (or dies).

## An example of metrics displayed with grafana
Collected using a [locust](https://github.com/locustio/locust) python script with ~2500 concurrent users (my pc startet throttling at about 1000, which can clearly be seen with the db 'Get' requests taking a lot more time than expected!)
### Database:

![db_metric](db_metric.png)

### Cache:
![cache_metric](cache_metric.png)

### Other Documentation:
* [Database Schema](database.md)
* [Consumer Decision Tree](consumer.md)
* [Cache Implementation](cache.md)
* [JSON Validation](validation.md)

### Back to [Main README](../README.md)