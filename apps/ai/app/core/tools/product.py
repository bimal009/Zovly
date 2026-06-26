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