import json
from locust import HttpUser, task, between

# Load order IDs
with open("mock.json") as f:
    orders = json.load(f)
order_ids = [order["order_uid"] for order in orders]

class OrdersUser(HttpUser):
    wait_time = between(1, 3)  # wait 1–3 seconds between tasks

    @task
    def get_order(self):
        import random
        order_id = random.choice(order_ids)
        self.client.get(f"/orders/{order_id}")
