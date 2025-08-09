import json
import random
import string
from datetime import datetime, timedelta
import uuid

NAMES = ["Walter White", "Jessie Pinkman", "Gus Fring", "Saul Goodman", "Skylar White"]
CITIES = ["New York", "Los Angeles", "London", "Berlin", "Tokyo"]
REGIONS = ["California", "Berlin", "Tokyo Prefecture", "Moscow Region"]
EMAIL_DOMAINS = ["gmail.com", "yahoo.com", "example.com", "mail.ru"]
CURRENCIES = ["USD", "EUR", "GBP", "JPY", "RUB", "CNY"]
DELIVERY_SERVICES = ["meest", "fedex", "dhl", "ups", "Russioan Post"]
PROVIDERS = ["wbpay", "paypal", "stripe", "banktransfer", "MIR"]
BRANDS = ["Vivienne Sabo", "Maybelline", "L'Oreal", "MAC", "NYX", "Converse", "Vans", "Nike"]

def random_string(length=10):
    return ''.join(random.choices(string.ascii_lowercase + string.digits, k=length))

def random_phone():
    return f"+{random.randint(1, 999)}{random.randint(100000000, 999999999)}"

def random_date():
    start = datetime(2020, 1, 1)
    end = datetime(2025, 1, 1)
    return start + timedelta(seconds=random.randint(0, int((end - start).total_seconds())))

def generate_order():
    order_uid = uuid.uuid4().hex[:16]
    track_number = "TRK" + random_string(10).upper()
    delivery_name = random.choice(NAMES)
    email = delivery_name.replace(" ", ".").lower() + "@" + random.choice(EMAIL_DOMAINS)

    return {
        "order_uid": order_uid,
        "track_number": track_number,
        "entry": random.choice(["WBIL", "ENTR", "GATE"]),
        "delivery": {
            "name": delivery_name,
            "phone": random_phone(),
            "zip": str(random.randint(100000, 999999)),
            "city": random.choice(CITIES),
            "address": f"{random.randint(1, 100)} {random.choice(['Main St', 'High St', 'Ploshad Mira', 'Broadway'])}",
            "region": random.choice(REGIONS),
            "email": email
        },
        "payment": {
            "transaction": order_uid,
            "request_id": "",
            "currency": random.choice(CURRENCIES),
            "provider": random.choice(PROVIDERS),
            "amount": random.randint(100, 5000),
            "payment_dt": int(datetime.now().timestamp()),
            "bank": random.choice(["alpha", "beta", "gamma"]),
            "delivery_cost": random.randint(100, 2000),
            "goods_total": random.randint(50, 2000),
            "custom_fee": random.randint(0, 100)
        },
        "items": [
            {
                "chrt_id": random.randint(1000000, 9999999),
                "track_number": track_number,
                "price": random.randint(50, 1000),
                "rid": uuid.uuid4().hex[:16],
                "name": random.choice(["T-shirt", "Shoes", "Jacket", "Bag", "Mascaras"]),
                "sale": random.randint(0, 70),
                "size": random.choice(["S", "M", "L", "XL", "0"]),
                "total_price": random.randint(50, 2000),
                "nm_id": random.randint(100000, 999999),
                "brand": random.choice(BRANDS),
                "status": random.randint(100, 300)
            }
        ],
        "locale": random.choice(["en", "ru", "de", "fr"]),
        "internal_signature": "",
        "customer_id": random_string(8),
        "delivery_service": random.choice(DELIVERY_SERVICES),
        "shardkey": str(random.randint(1, 10)),
        "sm_id": random.randint(1, 200),
        "date_created": random_date().strftime("%Y-%m-%dT%H:%M:%SZ"),
        "oof_shard": str(random.randint(1, 5))
    }

if __name__ == "__main__":
    orders = [generate_order() for _ in range(10)]
    with open("valid_mock_orders.json", "w") as f:
        json.dump(orders, f, indent=3)
    print("Generated 10 mock orders into valid_mock_orders.json")
