# L0 Task Project

This project is a Go-based application designed for handling orders, featuring a producer, a consumer, an API service, and a PostgreSQL database. It includes an LRU cache for efficient data retrieval and Prometheus/Grafana for monitoring.

## Features

*   **Order Generation:** A producer service to generate and publish order data
*   **Order Consumption:** A consumer service to process order data concurrently
*   **API Service:** A web API to interact with the service's data
*   **PostgreSQL Database:** Persistent storage for order information
*   **LRU Cache:** In-memory caching for frequently accessed orders
*   **Monitoring:** Integration with Prometheus and Grafana for metrics and dashboards


## Getting Started

### Prerequisites

*   Go (version 1.24.5 or higher)
*   Docker and Docker Compose

### Installation

1.  **Clone the repository:**

    ```bash
    git clone https://github.com/your-repo/l0-task.git
    cd l0-task
    ```

2.  **Set up environment variables:**

    Copy the example environment file and modify it as needed. It is used by the `docker compose`:

    ```bash
    cp .env.example .env
    ```

    Copy the config file for the consumer service:

    ```bash
    cp config.yaml.example config.yaml
    ```

3.  **Build and run with Docker Compose:**
    You can either pull prebuilt images:  
    
    ```bash
    docker pull goinginblind/l0-task-app
    docker pull goinginblind/l0-task-producer
    ```
    or build them locally:
    
    ```bash
    docker compose up --build
    ```
    If images are pulled, you can also use:

    ```bash
    make run-all
    ```
    It runs the whole service with default settings:
    - Python script generates 3600 orders (10% with invalid data)
    - Producer runs at 1 RPS (--rps N can override)
    - PostgreSQL setup with migrations ([Goose migration tool](https://github.com/pressly/goose) required)
    - Kafka with Zookeeper
    - Consumer with default settings
    - Prometheus and Grafana

    Afterwards, clean everything up:

    ```bash
    make clean
    ```
    This deletes generated payloads, containers, and orphans.
    

## Project Structure

```
.env.example        # an example of .env, used by the docker compose
.gitignore          
config.yaml.example # an example of app (consumer-service) config
docker-compose.yml  
Dockerfile          # Dockerfile to build app and producer containers
gen_orders.py       # A Python script used to generate JSON payloads
go.mod              
go.sum
Makefile            # Commands to speed up builds and teardowns
cmd/
├── producer/       # Order producer service
└── service/        # Consumer service
configs/
├── grafana/        # Grafana dashboards and datasources
└── prometheus/     # Prometheus configuration
internal/
├── api/            # API handlers, middleware, and UI templates
├── app/            # Main application logic orchestrator
├── config/         # Configuration loading
├── consumer/       # Order consumer logic
├── domain/         # Core domain models
├── pkg/            # Reusable packages (health, logger, metrics, sizeof)
├── service/        # Business logic layer and LRU cache implementation as its wrapper
└── store/          # Database interaction (with PostgreSQL)
sql/
├── 001_initial_schema.sql                          # Main schema
└── 002_add_timestamps_and_latest_orders_query.sql  # Add timestamps used by cache
```

## API Endpoints

The API service typically runs on port `8080` (configurable via `.env` if running via a container or through regular enviroment variables).

*   `GET /`: Home page (UI).
*   `GET /order/{order_uid}`: Retrieve order details by UID.
*   `GET /metrics`: Endpoint scraped by prometheus

## Technologies Used

*   Go
*   PostgreSQL
*   Kafka 
*   Docker
*   Docker Compose
*   Prometheus
*   Grafana
*   HTML/CSS/JavaScript (UI)
