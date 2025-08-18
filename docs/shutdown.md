## Graceful Shutdown

The service is designed to shut down gracefully to prevent data loss and ensure clean resource management. If it receives a `SIGINT` or a `SIGTERM` signal it does the next things:

1.  **Shutdown Initiated:** Upon receiving a signal, the application begins the shutdown process.
2.  **Consumer Stop:** The Kafka consumer is signaled to stop polling for new messages and finish processing any in-flight messages.
3.  **HTTP Server Shutdown:** The HTTP server stops accepting new connections and waits for existing requests to complete, up to a configured timeout (`shutdown_timeout` in `config.yaml`).
4.  **Resource Cleanup:** Finally, resources like the database connection pool are closed.

This ensures that the application shuts down cleanly, without interrupting ongoing work or leaving connections open.