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
        I["log, discard"]
        Z["Planned dead-line queue"]
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
    I --> Z & H
    Z --> H
    F --> H
    G -- invalid/duplicate order or unknown error --> I
    G -- Success --> H
    G -- no connection to db --> K
    K -- recovered --> H
    K -- still fails --> M
    style Z stroke-dasharray: 5 5
    linkStyle 6 stroke-dasharray: 5 5,fill:none
    linkStyle 8 stroke-dasharray: 5 5,fill:none

```
Notes:
1. **How many workers?**
The amount of workers is configurable via enviroment or `config.yaml`
2. **Why are invalids, duplicates and unknown errors discarded?**
These orders are logged by their order_uid, and I intended to send them to a dead-line queue, but had no time (as of now) to implement it.
3. **What exactly happens when the database is down?**
The worker first retries with exponential backoff (too, configurable) in case the connection was lost because of the transient errors (e.g. 1 ms hiccup), then if it fails too, the worker skips message without commits, marks db as unhealthy and the service's health checker kicks in. It pings the database once each N seconds waiting for it to be up.
4. **Why keep polling if the database connection is down?**
It keeps the heartbeat of the consumer up on the Kafka-side, so the consumer is not kicked out. If the database connection lost times out, then the whole program will gracefully exit.
5. **What happens if the DB is up again?**
The consumer proceeds as normally, it again
