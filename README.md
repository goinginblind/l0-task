# Project Title: Wildberries L0 Task (Order Service)
## Пока только план для ридми.мд (к обеду будет) 

## Description
A Go-based microservice for processing and serving order information. It consumes order data from Kafka, stores it in a PostgreSQL database, and provides a web interface to view order details. A separate producer application is included to simulate order generation and publishing to Kafka.

## Features
- **Kafka Consumer:** Reliably consumes order messages from a Kafka topic.
- **Data Persistence:** Stores order data in a PostgreSQL database.
- **HTTP API & UI:** Provides a simple web interface to retrieve and display order details by UID.
- **Order Producer:** A command-line tool to generate and publish mock order data to Kafka.
- **Database Migrations:** Manages database schema changes using `goose`.
- **Containerized Environment:** Uses Docker Compose for easy setup of PostgreSQL and Kafka (Zookeeper + Broker).
- **Structured Logging:** Utilizes Zap for efficient and structured logging.
- **Input Validation:** Validates incoming order data.
- **In-memory Cache:** Implements an in-memory cache for frequently accessed orders to improve performance.

## Architecture
The application follows a layered architecture:
- **Entrypoint (`cmd/service/main.go`):** Initializes and runs the application.
- **Orchestration (`internal/app`):** Wires up and manages dependencies between different components.
- **Delivery Layer:**
    - **`internal/api` (HTTP + UI):** Handles incoming HTTP requests, serves static assets, and renders HTML templates for the web interface.
    - **`internal/consumer` (Kafka):** Consumes messages from Kafka, processes them, and handles database health checks to pause/resume consumption.
- **Business Logic Layer (`internal/service`):** Contains the core business logic for order processing, including an in-memory cache decorator.
- **Data Access Layer (`internal/store`):** Manages interactions with the PostgreSQL database.
- **Shared Dependencies:**
    - **`internal/domain`:** Defines the core data structures (e.g., `Order`).
    - **`internal/pkg`:** Provides common utilities like logging, health checks, and size calculation.
    - **`internal/config`:** Handles application configuration.

``` Mermaid
---
config:
  layout: elk
---
flowchart TD
 subgraph Entrypoint["Entrypoint"]
        A["cmd/service/main.go"]
  end
 subgraph subGraph1["Orchestration Layer"]
        B["internal/app"]
  end
 subgraph subGraph2[" "]
    direction LR
        Entrypoint
        subGraph1
  end
 subgraph subGraph3["Delivery Layer"]
        C["internal/api (HTTP + UI)"]
        D["internal/consumer (Kafka)"]
  end
 subgraph subGraph4["Business Logic Layer"]
        E_Cache["internal/service (+ a cache decorator)"]
        E["internal/service (core logic)"]
  end
 subgraph subGraph5["Data Access Layer"]
        F["internal/store"]
  end
 subgraph subGraph6["Application Core"]
    direction LR
        subGraph3
        subGraph4
        subGraph5
  end
 subgraph subGraph7["Shared Dependencies"]
    direction LR
        G["internal/domain"]
        H["internal/pkg (logger, health, sizeof)"]
        I["internal/config"]
  end
 subgraph subGraph8["External Services"]
    direction LR
        K[("Kafka Broker")]
        J[("PostgreSQL DB")]
  end
    A -- Initializes --> B
    B -- Wires up dependencies --> C & D & E_Cache & F & H & I
    C -- Uses --> E_Cache
    D -- Uses --> E_Cache
    E_Cache -- Decorates --> E
    E_Cache -- Preloads from --> F
    E -- Uses --> F
    C -- Read from --> I
    D -- Read from --> I
    E_Cache -- Read from --> I
    E -- Read from --> I
    F -- Read from --> I
    H -- Read from --> I
    C -- Use --> H
    D -- Use --> H
    E_Cache -- Use --> H & G
    E -- Use --> H & G
    F -- Use --> H & G
    D -- Consumes from --> K
    F -- Talks to --> J
    style B fill:#D50000,color:#FFFFFF
    style subGraph1 fill:#FFCDD2,stroke:#000000,color:#000000
    style C fill:#FF6D00,stroke:#333,stroke-width:2px,color:#fff
    style D fill:#FFD600,stroke:#333,stroke-width:2px,color:#fff
    style E_Cache fill:#AA00FF,stroke:#333,stroke-width:2px,color:#fff
    style E fill:#2962FF,stroke:#333,stroke-width:2px,color:#fff
    style F fill:#00C853,stroke:#333,stroke-width:2px,color:#fff
    style subGraph3 fill:#FFF9C4,color:#000000
    style subGraph4 fill:#BBDEFB,color:#000000
    style subGraph5 fill:#C8E6C9,color:#000000
    linkStyle 1 stroke:#D50000,fill:none
    linkStyle 2 stroke:#D50000,fill:none
    linkStyle 3 stroke:#D50000,fill:none
    linkStyle 4 stroke:#D50000,fill:none
    linkStyle 5 stroke:#D50000,fill:none
    linkStyle 6 stroke:#D50000,fill:none
    linkStyle 7 stroke:#FF6D00,fill:none
    linkStyle 8 stroke:#FFD600,fill:none
    linkStyle 9 stroke:#AA00FF,fill:none
    linkStyle 10 stroke:#AA00FF,fill:none
    linkStyle 11 stroke:#2962FF,fill:none
    linkStyle 12 stroke:#FF6D00,fill:none
    linkStyle 13 stroke:#FFD600,fill:none
    linkStyle 14 stroke:#AA00FF,fill:none
    linkStyle 15 stroke:#2962FF,fill:none
    linkStyle 16 stroke:#00C853,fill:none
    linkStyle 18 stroke:#FF6D00,fill:none
    linkStyle 19 stroke:#FFD600,fill:none
    linkStyle 20 stroke:#AA00FF,fill:none
    linkStyle 21 stroke:#AA00FF,fill:none
    linkStyle 22 stroke:#2962FF,fill:none
    linkStyle 23 stroke:#2962FF,fill:none
    linkStyle 24 stroke:#00C853,fill:none
    linkStyle 25 stroke:#00C853,fill:none
    linkStyle 26 stroke:#FFD600,fill:none
    linkStyle 27 stroke:#00C853,fill:none
```

## Getting Started

### Prerequisites
- Docker and Docker Compose
- Go (version 1.24.5 or higher)
- Python 3 (for `gen_orders.py`)

### Setup and Installation
1.  **Clone the repository:**
    ```bash
    git clone https://github.com/goinginblind/l0-task.git
    cd l0-task
    ```
2.  **Environment Variables:**
    Copy `.env.example` to `.env` and configure your environment variables (e.g., PostgreSQL credentials, Kafka broker addresses).
    ```bash
    cp .env.example .env
    # Edit .env file with your desired configurations
    ```
3.  **Build and Run Services (using Makefile):**
    The `setup` command will build the Go binaries, generate mock orders, start Docker containers (PostgreSQL, Zookeeper, Kafka), and run database migrations.
    ```bash
    make setup
    ```

### Running the Application

1.  **Start the Producer:**
    This will send generated orders from `orders.json` to Kafka.
    ```bash
    make runp
    ```
2.  **Start the Service:**
    This will start the HTTP server and Kafka consumer.
    ```bash
    make runs
    ```

### Accessing the Application
- **Web UI:** Open your browser to `http://localhost:8080` (or whatever port is configured in `.env`).
- **Order Details:** You can view specific order details by navigating to `http://localhost:8080/orders/<ORDER_UID>`.

## API Endpoints
- `GET /home`: Home page.
- `GET /orders/{order_uid}`: Retrieve details for a specific order.

## Database Schema
The PostgreSQL database includes the following main tables:
- `orders`: Stores primary order information.
- `deliveries`: Stores delivery details, linked to `orders` by `order_id`.
- `payments`: Stores payment details, linked to `orders` by `order_id`.
- `items`: Stores individual item details within an order, linked to `orders` by `order_id`.

## Testing
Run all Go tests:
```bash
make test
```

## Cleanup
To stop and remove Docker containers, and clean up generated files:
```bash
make clean
```
