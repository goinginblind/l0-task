import json
import random
import string
from datetime import datetime, timedelta
import uuid
import argparse

# --------------------------------------------------------------------------------------------------------------------------------------
# Helpers
# --------------------------------------------------------------------------------------------------------------------------------------
NAMES = ["Walter White", "Jessie Pinkman", "Gus Fring", "Saul Goodman", "Skylar White", "Tuco Salamanca", "Ignacio Varga"]
CITIES = ["New York", "Los Angeles", "London", "Berlin", "Tokyo", "Albuqerque", "Moscow"]
REGIONS = ["California", "Berlin", "Tokyo Prefecture", "Moscow Region", "New Mexico"]
EMAIL_DOMAINS = ["gmail.com", "yahoo.com", "example.com", "mail.ru"]
CURRENCIES = ["USD", "EUR", "GBP", "JPY", "RUB", "CNY"]
DELIVERY_SERVICES = ["meest", "fedex", "dhl", "ups", "Russian Post"]
PROVIDERS = ["wbpay", "paypal", "stripe", "banktransfer", "MIR"]
BRANDS = ["Vivienne Sabo", "Maybelline", "L'Oreal", "MAC", "NYX", "Converse", "Vans", "Nike"]
ITEM_NAMES = ["T-shirt", "Shoes", "Jacket", "Bag", "Mascaras"]


def random_string(length=10):
    return ''.join(random.choices(string.ascii_lowercase + string.digits, k=length))

def random_phone():
    # It's supposed to be e164 valid
    country_code = str(random.choice([1, 44, 49, 81, 7]))  # sample valid codes
    national_number_length = random.randint(7, 11)  # ensures total <= 15 digits
    national_number = ''.join(random.choices(string.digits, k=national_number_length))
    return f"+{country_code}{national_number}"

def random_date(start_year=2020, end_year=2025):
    start = datetime(start_year, 1, 1)
    end = datetime(end_year, 1, 1)
    return start + timedelta(seconds=random.randint(0, int((end - start).total_seconds())))

def generate_item(track_number):
    return {
        "chrt_id": random.randint(1000000, 9999999),
        "track_number": track_number,
        "price": random.randint(50, 1000),
        "rid": uuid.uuid4().hex[:16],
        "name": random.choice(ITEM_NAMES),
        "sale": random.randint(0, 70),
        "size": random.choice(["S", "M", "L", "XL", "0"]),
        "total_price": random.randint(50, 2000),
        "nm_id": random.randint(100000, 999999),
        "brand": random.choice(BRANDS),
        "status": random.randint(100, 300)
    }

def generate_order(min_items=1, max_items=3):
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
        "items": [generate_item(track_number) for _ in range(random.randint(min_items, max_items))],
        "locale": random.choice(["en", "ru", "de", "fr"]),
        "internal_signature": "",
        "customer_id": random_string(8),
        "delivery_service": random.choice(DELIVERY_SERVICES),
        "shardkey": str(random.randint(1, 10)),
        "sm_id": random.randint(1, 200),
        "date_created": random_date().strftime("%Y-%m-%dT%H:%M:%SZ"),
        "oof_shard": str(random.randint(1, 5))
    }

# --------------------------------------------------------------------------------------------------------------------------------------
# Invalid Data Injection
# --------------------------------------------------------------------------------------------------------------------------------------
def inject_invalid_data(order):
    choice = random.choice([
        "phone", "email", "amount", "locale", "items",
        "shardkey", "sm_id", "goods_total", "currency"
    ])

    if choice == "phone":
        order["delivery"]["phone"] = "12345abc"  # invalid e164
    elif choice == "email":
        order["delivery"]["email"] = "not-an-email"
    elif choice == "amount":
        order["payment"]["amount"] = -100  # fails gte=0
    elif choice == "locale":
        order["locale"] = "xx_123"  # invalid bcp47
    elif choice == "items":
        order["items"] = []  # fails min=1
    elif choice == "shardkey":
        order["shardkey"] = "abc"  # not numeric
    elif choice == "sm_id":
        order["sm_id"] = 0  # fails gt=0
    elif choice == "goods_total":
        order["payment"]["goods_total"] = 0  # fails gt=0
    elif choice == "currency":
        order["payment"]["currency"] = "NOT"  # invalid iso4217

    return order

# --------------------------------------------------------------------------------------------------------------------------------------
# Script w/ flags
# --------------------------------------------------------------------------------------------------------------------------------------
# 100% valid orders:
# python gen_orders.py -n 50 --min-items 1 --max-items 10 --invalid-rate 0.0 -o mock.json

# 80%  valid orders:                                                     ↓ this flag == 20% invalid
# python gen_orders.py -n 50 --min-items 1 --max-items 10 --invalid-rate 0.2 -o mock.json
if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Generate mock orders JSON (valid + invalid).")
    parser.add_argument("-n", "--num-orders", type=int, default=100, help="Number of orders to generate")
    parser.add_argument("--min-items", type=int, default=1, help="Minimum items per order")
    parser.add_argument("--max-items", type=int, default=5, help="Maximum items per order")
    parser.add_argument("--invalid-rate", type=float, default=0.0, help="Fraction of orders to make invalid (0.0–1.0)")
    parser.add_argument("-o", "--output", default="mock_orders.json", help="Output file name")
    args = parser.parse_args()

    orders = []
    for _ in range(args.num_orders):
        order = generate_order(args.min_items, args.max_items)
        if random.random() < args.invalid_rate:
            order = inject_invalid_data(order)
        orders.append(order)

    with open(args.output, "w") as f:
        json.dump(orders, f, indent=3)

    print(f"Generated {args.num_orders} orders ({args.invalid_rate*100:.0f}% invalid) into {args.output}")