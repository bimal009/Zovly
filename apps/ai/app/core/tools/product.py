import json
import os
from dataclasses import dataclass

import requests
from dotenv import load_dotenv
from langchain.tools import tool, ToolRuntime

load_dotenv()

API_BASE_URL = os.getenv("API_BASE_URL")
INTERNAL_TOKEN = os.getenv("INTERNAL_TOKEN")


@dataclass
class BusinessContext:
    """Runtime context injected per request — never supplied by the model."""

    business_id: str


@tool
def get_category_product_count(
    runtime: ToolRuntime[BusinessContext],
    category_slug: str = "",
) -> str:
    """Get the number of products for the business, optionally filtered by category slug."""

    business_id = runtime.context.business_id

    url = f"{API_BASE_URL}/api/v1/internal/products/count"

    params = {
        "businessID": business_id,
    }

    if category_slug:
        params["categorySlug"] = category_slug

    headers = {
        "Authorization": f"Bearer {INTERNAL_TOKEN}",
        "Accept": "application/json",
    }

    try:
        response = requests.get(
            url,
            params=params,
            headers=headers,
            timeout=30,
        )

        response.raise_for_status()

        data = response.json()

        count = data.get("data", {}).get("count")

        if count is None:
            return f"Unexpected response: {data}"

        return f"Product count: {count}"

    except requests.HTTPError:
        return f"HTTP {response.status_code}: {response.text}"

    except requests.RequestException as e:
        return f"Request failed: {e}"


@tool
def get_product_details(
    runtime: ToolRuntime[BusinessContext],
    source_id: str,
) -> str:
    """Get full, up-to-date details (price, stock, description, variants) for a single product.

    Use this when the customer asks about a specific product's price, availability, or options.
    Pass the source_id shown next to that product in the retrieved context (e.g. "source_id: abc").
    Prices are in minor units (cents) — divide by 100 for the display amount.
    """

    business_id = runtime.context.business_id

    url = f"{API_BASE_URL}/api/v1/internal/products/{source_id}"
    params = {"businessID": business_id}
    headers = {
        "Authorization": f"Bearer {INTERNAL_TOKEN}",
        "Accept": "application/json",
    }

    try:
        response = requests.get(url, params=params, headers=headers, timeout=30)

        if response.status_code == 404:
            return "Product not found."

        response.raise_for_status()

        product = response.json().get("data")
        if not product:
            return "Product not found."

        return json.dumps(product, ensure_ascii=False)

    except requests.HTTPError:
        return f"HTTP {response.status_code}: {response.text}"
    except requests.RequestException as e:
        return f"Request failed: {e}"