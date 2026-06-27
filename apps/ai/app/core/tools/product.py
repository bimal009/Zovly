import json
import os
import uuid
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
    conversation_id: str = ""  # so product lookups can record the active product
    active_product_id: str = ""  # product currently under discussion, if any


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


PRODUCTS_PAGE_SIZE = 10


@tool
def get_products_by_category(
    runtime: ToolRuntime[BusinessContext],
    category_slug: str,
    page: int = 1,
) -> str:
    """List the products in a category (paginated), each with its real source_id.

    Use this to DISCOVER which products exist and get their source_id when you don't
    already have one (e.g. the customer asks about a category, or you only know a
    product name). Then call get_product_details with the chosen source_id for live
    price, stock, and variants.

    Results are paged (10 per page). Pass page=1 first; if the result says more pages
    are available, call again with page=2, page=3, ... to see the rest.
    """

    business_id = runtime.context.business_id

    if page < 1:
        page = 1
    offset = (page - 1) * PRODUCTS_PAGE_SIZE

    url = f"{API_BASE_URL}/api/v1/internal/products"
    params = {
        "businessID": business_id,
        "categorySlug": category_slug,
        "limit": PRODUCTS_PAGE_SIZE,
        "offset": offset,
    }
    headers = {
        "Authorization": f"Bearer {INTERNAL_TOKEN}",
        "Accept": "application/json",
    }

    try:
        response = requests.get(url, params=params, headers=headers, timeout=30)
        response.raise_for_status()

        body = response.json()
        products = body.get("data") or []
        total = (body.get("meta") or {}).get("total", len(products))

        if not products:
            if page > 1:
                return f"No more products in category '{category_slug}' (page {page} is empty)."
            return f"No products found in category '{category_slug}'."

        lines = [
            f"{p.get('name', 'Unnamed')} (source_id: {p.get('id', '')})"
            for p in products
        ]

        shown = offset + len(products)
        header = f"Products in '{category_slug}' ({shown} of {total}, page {page}):"
        footer = ""
        if shown < total:
            footer = f"\nMore products available — call again with page={page + 1}."

        return header + "\n" + "\n".join(lines) + footer

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
    conversation_id = runtime.context.conversation_id

    # guard against hallucinated ids (e.g. "abc", "1") — never hit the API with one
    try:
        uuid.UUID(str(source_id))
    except (ValueError, AttributeError, TypeError):
        return (
            "Invalid source_id — it must be a real product source_id taken verbatim "
            "from the provided context. Do not guess or make one up. If you don't have "
            "one, ask the customer which product they mean."
        )

    url = f"{API_BASE_URL}/api/v1/internal/products/{source_id}"
    params = {"businessID": business_id}
    if conversation_id:
        params["conversationID"] = conversation_id  # records this as the active product
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