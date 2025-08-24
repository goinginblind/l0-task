# Sentinel errors
``` Go
var (
	ErrNotFound = errors.New("no record found")
	ErrAlreadyExists = errors.New("record already exists")
	ErrConnectionFailed = errors.New("failed connection to DB")
)
```
These errors "bubble up", and the upper layers are not actually concerned with the exact type of errors. Thus, even if the data store changes later (e.g., `unique_violation` in PostgreSQL has code `23505`, and MySQL uses code `1062` for the same thing), it will not affect them. An example of use is when an error "bubbles up" to the worker, it immediately marks the database as unhealthy using the HealthChecker.

# DBHealthChecker
This is a simple "pinger" with a configurable interval and a configurable timeout for a single ping. It runs constantly, but its atomic.Bool variable can be changed by the worker earlier than the ping interval. Initially, I had the idea for it to run only when the worker detects that the database is unavailable, but managing this proved too complex. Therefore, `MarkUnhealthy()` is a kind of legacy that finds its use when the healthchecker itself has a large interval (e.g., 15 seconds) or simply to prevent unnecessary sending of partitions from Kafka, for example, to redistribute the load faster.
The `health` package can be found [here](../internal/pkg/health/dbhealth.go)!

# Monitoring
The service uses Prometheus, Grafana, and Node-exporter:
- **Host system metrics:** CPU load, RAM, Load Average (1m, 5m, 15m), Network I/O
- **API metrics:** number of requests and their latencies
- **Consumer metrics:** processing frequency, processing latency, consumer lag, and DLQ sending frequency
- **Cache:** hit/miss rate and cache response speed for Get/Insert requests
- **DB:** connection availability, processing time for Get/Insert requests, and frequency of transient errors (connection violations not longer than 5 ms).

## Example of metrics displayed in Grafana
Collected using a [locust](https://github.com/locustio/locust) python script with ~70 requests per second to `/orders/{id}` almost from the very start of the broker, producer, db, etc.
### System metrics:
![host_metric](../host_metric.png)

### API metrics:
![api_metric](../api_metric.png)

### Consumer metrics:
![cons_metric](../cons_metric.png)

### Cache metrics:
![cache_metric](../cache_metric.png)

### DB metrics:
![db_metric](../db_metric.png)

### Other Documentation:
* [Database Schema](database.md)
* [Consumer Principles](consumer.md)
* [Cache Implementation](cache.md)
* [JSON Validation](validation.md)

### Back to [Main README](../../README.md)