# What the cosumer does?

```mermaid
---
config:
  layout: elk
  theme: redux-dark
---
flowchart TD
 subgraph Consumer["Consumer"]
        B{"DB is up?"}
        A["poll message from Kafka"]
        C["pause partitions, keep polling"]
        D["send to worker"]
  end
 subgraph subGraph1["One of the N workers"]
        E{"valid JSON?"}
        F["discard"]
        G["process order"]
        I["log"]
        Z["send to dead-line queue"]
        H["commit"]
        K["retry up to N times"]
        M["mark DB unhealthy, NO commit, kafka resends later"]
  end
    A --> B
    B -- No --> C
    B -- Yes --> D
    D --> E
    E -- No --> F
    E -- Yes --> G
    I --> Z
    Z --> H
    G -- invalid/duplicate order or unknown error --> I
    G -- Success --> H
    G -- no connection to db --> K
    K -- recovered --> H
    K -- still fails --> M

```
Notes:
1. **How many workers?**
The amount of workers is configurable via enviroment or `config.yaml`
2. **What happens to invalid, duplicate, or unknown errors?**
These orders are logged by their `order_uid` and then sent to a Dead-Letter Queue (DLQ) for later inspection and potential reprocessing. The default DLQ topic is `orders-dlq`. This prevents "poison pill" messages from blocking the consumer while ensuring no data is lost.
3. **What exactly happens when the database is down?**
The worker first retries with exponential backoff (too, configurable) in case the connection was lost because of the transient errors (e.g. 1 ms hiccup), then if it fails too, the worker skips message without commits, marks db as unhealthy and the service's health checker kicks in. It pings the database once each N seconds waiting for it to be up.
4. **Why keep polling if the database connection is down?**
It keeps the heartbeat of the consumer up on the Kafka-side, so the consumer is not kicked out. If the database connection lost times out, then the whole program will gracefully exit.
5. **What happens if the DB is up again?**
The consumer proceeds as normally after it pings again and gets a response from the db


### Other Documentation:
* [Database Schema](database.md)
* [Cache Implementation](cache.md)
* [JSON Validation](validation.md)
* [Errors, Metrics, and DB Health Checks](misc.md)

### Back to [Main README](../README.md)