import os

import requests
from dotenv import load_dotenv
from langchain.tools import tool, ToolRuntime

from app.core.tools.product import BusinessContext

load_dotenv()

API_BASE_URL = os.getenv("API_BASE_URL")
INTERNAL_TOKEN = os.getenv("INTERNAL_TOKEN")


@tool
def get_categories(runtime: ToolRuntime[BusinessContext]) -> str:
    """List all product categories for the business (name and slug for each)."""

    business_id = runtime.context.business_id

    url = f"{API_BASE_URL}/api/v1/internal/categories"

    params = {"businessID": business_id}

    headers = {
        "Authorization": f"Bearer {INTERNAL_TOKEN}",
        "Accept": "application/json",
    }

    try:
        response = requests.get(url, params=params, headers=headers, timeout=30)
        response.raise_for_status()

        data = response.json()
        categories = data.get("data") or []

        if not categories:
            return "No categories found."

        lines = []
        for cat in categories:
            name = cat.get("name", "")
            slug = cat.get("slug", "")
            lines.append(f"{name} (slug: {slug})")

        return "Categories:\n" + "\n".join(lines)

    except requests.HTTPError:
        return f"HTTP {response.status_code}: {response.text}"

    except requests.RequestException as e:
        return f"Request failed: {e}"
