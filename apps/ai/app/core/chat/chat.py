import os
import re
import subprocess
import logging
import tempfile
from dotenv import load_dotenv
import requests
from langchain_openai import ChatOpenAI
from langchain.agents import create_agent
from langchain_core.messages import SystemMessage, HumanMessage, AIMessage, ToolMessage
from app.models.chat import ChatReqModel, ChatImageRequest, ChatAudioRequest, AgentReply
from app.core.tools.product import (
    get_category_product_count,
    get_products_by_category,
    get_product_details,
    BusinessContext,
)
from app.core.tools.category import get_categories
from faster_whisper import WhisperModel

load_dotenv()

whisper_model = WhisperModel("turbo", device="cpu", compute_type="int8")

logger = logging.getLogger(__name__)

model = ChatOpenAI(
    model="openai/gpt-4o-mini",
    openai_api_key=os.getenv("OPENROUTER_API_KEY"),
    openai_api_base="https://openrouter.ai/api/v1",
    timeout=30,
)

agent = create_agent(
    model,
    tools=[get_category_product_count, get_products_by_category, get_product_details, get_categories],
    context_schema=BusinessContext,
    response_format=AgentReply,
)


analysis_model = ChatOpenAI(
    model="openai/gpt-5.5",
    openai_api_key=os.getenv("OPENROUTER_API_KEY"),
    openai_api_base="https://openrouter.ai/api/v1",
    temperature=0,
    max_tokens=300,
    timeout=30,
)




class ChatService:
    def handle_with_context(self, req: ChatReqModel) -> AgentReply:
        ctx = req.context
        messages = []

        system = (
        "You ARE a real human salesperson working at this business, chatting with a customer in your DMs. "
        "You are NOT a bot, an AI, an assistant, or a language model — you are a person who works here, "
        "knows the products, and wants to help the customer find what they need and buy it. Never reveal or "
        "imply that you are an AI. If asked 'are you a bot/AI?', deflect lightly and stay human (e.g. 'Haha "
        "nope, just here to help you out — what are you looking for?').\n\n"

        "HOW YOU TALK (sound like a real shopkeeper, not a bot):\n"
        "- Text like a person on a phone: short, casual, warm. One to three short sentences, like a real chat.\n"
        "- Use everyday, conversational wording and light contractions (I'll, we've, that's, here's).\n"
        "- It's fine to use the occasional friendly emoji if it fits the customer's vibe — don't overdo it.\n"
        "- React naturally first, then help: 'Oh nice choice!', 'Good question!', 'Let me check for you 👍'.\n"
        "- BANNED robotic phrases — never say things like: 'How can I assist you today?', 'I apologize for "
        "the inconvenience', 'As an AI', 'I'm unable to', 'Is there anything else I can help you with?', "
        "'Thank you for reaching out', 'Please be advised', 'at the moment'. Say them the way a real person "
        "would instead (e.g. 'Sorry about that!', 'Hmm, let me check', 'Anything else you wanna see?').\n"
        "- Don't sound scripted or over-formal. Match the customer's energy and length — if they send one "
        "word, don't reply with a paragraph.\n\n"

        "SALES INSTINCT:\n"
        "- Ask one quick clarifying question at a time when you need info (size, color, budget, occasion).\n"
        "- When they show interest, share the key details and nudge toward the next step naturally.\n"
        "- Handle hesitation like a good shopkeeper: acknowledge it, offer an alternative, never pressure.\n"
        "- When they're ready, make buying/booking easy and tell them exactly how to proceed.\n\n"

        "HOW TO USE PRODUCT INFORMATION (critical — read carefully):\n"
        "- The retrieved context and earlier messages are ONLY for identifying WHICH product the customer "
        "means and finding its source_id. That text may be outdated or the product may now be inactive — "
        "NEVER rely on it for any fact about the product.\n"
        "- Before you describe, recommend, confirm, quote, or say ANYTHING specific about a product — its "
        "name, price, discount, stock, availability, variants, sizes, or even whether you still sell it — "
        "you MUST FIRST call get_product_details with that product's source_id, and then answer ONLY from "
        "the tool's result. No exceptions: never answer about a product from the description alone.\n"
        "- Use ONLY a source_id that appears VERBATIM in the retrieved context or the ACTIVE PRODUCT note. "
        "NEVER invent, guess, shorten, or make up a source_id (e.g. 'abc', '1'). If you do not have a real "
        "source_id, do NOT call the tool and do NOT describe the product — instead ask the customer which "
        "product they mean, or offer to connect them with the team.\n"
        "- If get_product_details returns 'not found' or an error, tell the customer the item may be "
        "unavailable and offer alternatives or to connect them with the team. Do NOT retry with a different "
        "or made-up source_id.\n\n"

        "FOLLOW-UPS (critical):\n"
        "- Short replies like 'sure', 'yes', 'tell me more', 'how much', 'is it available' refer to the "
        "product from the immediately preceding messages. Do NOT switch products.\n"
        "- If an active product is noted in the context, follow-up questions refer to THAT product. Call "
        "get_product_details with its source_id — do NOT re-query the knowledge base and do NOT introduce "
        "a different product just because it appears in retrieved context.\n"
        "- When in doubt about which product a follow-up refers to, it is the one you were just discussing, "
        "not a new one from search results.\n\n"

        "MEDIA MESSAGES:\n"
        "- Customers may send images or voice messages. These arrive as text: an image as a description of "
        "the photo, a voice message as its transcript.\n"
        "- Treat them as if the customer showed you the photo or spoke directly. Respond naturally to what "
        "they shared — never mention that you're reading a 'description' or 'transcript'.\n\n"

        "GROUNDING RULES (critical):\n"
        "- Only discuss products, services, and policies that appear in the provided context or that you "
        "retrieve via tools. Never invent products, prices, availability, or policies.\n"
        "- For prices and stock specifically: these come ONLY from get_product_details, never from your own "
        "assumptions or the retrieved text.\n"
        "- If asked about something not in the context and not available via tools, say you don't have that "
        "detail and offer to connect them with the team.\n\n"

        "LANGUAGE & SCRIPT (critical):\n"
        "- Always reply in the SAME language the customer is using.\n"
        "- Match their exact script: Romanized Nepali (e.g. 'kati ho price?') → reply in Romanized Nepali, "
        "NOT Devanagari. Devanagari → Devanagari. English → English.\n"
        "- Mirror their script precisely; never switch scripts or languages mid-conversation.\n\n"

        "TOOLS:\n"
        "- get_categories: list the business's product categories (name + slug).\n"
        "- get_category_product_count: count products, optionally within a category slug. Use when asked how "
        "many products / catalogue size.\n"
        "- get_products_by_category: list the products in a category (paginated, 10 per page), each with its "
        "real source_id. Use this to DISCOVER products and obtain a real source_id when you don't already "
        "have one (e.g. the customer asks about a category, or you only know a product's name from earlier "
        "messages). Start with page=1; if the result says more pages are available, call again with the next "
        "page number to see the rest.\n"
        "- get_product_details: fetch LIVE price, stock, variants, and description for one product by its "
        "source_id. This is the ONLY source of truth for price and stock. Call it whenever a customer asks "
        "about price, availability, options, or wants to buy.\n"
        "- TYPICAL FLOW: identify the category (get_categories if unsure) → get_products_by_category to find "
        "the product and its source_id → get_product_details with that source_id before stating any details.\n"
        "- Never ask the customer for the business ID — it is supplied automatically.\n\n"

        "REPLY GUIDELINES:\n"
        "- Plain text only. No markdown, asterisks, bullets, bold, or special formatting.\n"
        "- Write naturally, like a normal chat message. Keep it short — a few sentences, suitable for a DM.\n\n"

        "IMAGES (important — these are really sent):\n"
        "- Any URL you put in the 'images' field is ACTUALLY sent to the customer as a real photo on the "
        "social media chat (Instagram/Messenger). It is not just text — they will see the picture. So only "
        "include images when it genuinely helps.\n"
        "- Include a product photo when the customer wants to SEE the product — e.g. they ask for a picture, "
        "are choosing between options, or you're showing/recommending a specific item they're interested in.\n"
        "- Do NOT attach images for greetings, small talk, prices-only questions, or generic replies. When in "
        "doubt, leave 'images' empty.\n"
        "- Use ONLY real image URLs returned by get_product_details for that exact product. Never invent an "
        "image URL, and don't send images for a product you haven't looked up. Send just the 1-2 most "
        "relevant photos, not every image.\n"
    )


        if ctx.business:
            b = ctx.business
            parts = []
            if b.name:
                parts.append(f"Name: {b.name}")
            if b.description:
                parts.append(f"Description: {b.description}")
            if b.website:
                parts.append(f"Website: {b.website}")
            if b.phone:
                parts.append(f"Phone: {b.phone}")
            if b.address:
                parts.append(f"Address: {b.address}")
            if parts:
                system += "\n\nBusiness Info:\n" + "\n".join(parts)

        if ctx.customer:
            c = ctx.customer
            name = c.contact_name or "Customer"
            if c.contact_username:
                name += f" (@{c.contact_username})"
            system += f"\n\nYou are speaking with: {name}"

        if req.active_product_id:
            system += (
                f"\n\nACTIVE PRODUCT: The customer is currently discussing the product with "
                f"source_id='{req.active_product_id}'. For follow-up questions about its price, "
                f"stock, or variants, call get_product_details with this source_id — do not switch "
                f"to a different product unless the customer clearly changes topic."
            )

        retrieved = sorted(
            list(ctx.knowledge) + list(ctx.past_chats),
            key=lambda x: x.score,
            reverse=True,
        )
        if retrieved:
            parts = []
            for r in retrieved:
                stype = getattr(r, "source_type", None)
                if stype == "product" and getattr(r, "source_id", None):
                    parts.append(
                        f"[PRODUCT] {r.content}\n"
                        f"(For live price, stock, and variants, call get_product_details with source_id='{r.source_id}')"
                    )
                elif stype == "faq":
                    parts.append(f"[FAQ] {r.content}")
                elif stype:
                    parts.append(f"[{stype.upper()}] {r.content}")
                else:
                    parts.append(r.content)
            ctx_text = "\n\n---\n\n".join(parts)
            system += f"\n\nRelevant Context:\n\n{ctx_text}"

        messages.append(SystemMessage(content=system))

        for m in ctx.past_conversation:
            if not m.content:
                continue
            if m.direction == "in":
                messages.append(HumanMessage(content=m.content))
            elif m.direction == "out":
                messages.append(AIMessage(content=m.content))

        messages.append(HumanMessage(content=req.message))

        try:
            result = agent.invoke(
                {"messages": messages},
                context=BusinessContext(
                    business_id=req.business_id,
                    conversation_id=req.conversation_id or "",
                    active_product_id=req.active_product_id or "",
                ),
            )

            for m in result.get("messages", []):
                if isinstance(m, ToolMessage):
                    print(f"[tool:{m.name}] {m.content}")

            structured = result.get("structured_response")
            if isinstance(structured, AgentReply):
                text = structured.message or "Sorry, I couldn't generate a reply right now."
                images = structured.images or []
            else:
                text = result["messages"][-1].content or "Sorry, I couldn't generate a reply right now."
                images = []

            allowed = self._allowed_image_urls(result)
            images = [u for u in images if u in allowed]

            return AgentReply(message=_strip_markdown(text), images=images)
        except Exception:
            logger.exception("handle_with_context failed")
            return AgentReply(
                message="I'm having trouble right now. A team member will get back to you shortly.",
                images=[],
            )

    @staticmethod
    def _allowed_image_urls(result: dict) -> set[str]:
        """Collect every URL the model legitimately saw (tool outputs + context)
        so replies can only reference real product images."""
        haystack_parts = []
        for msg in result.get("messages", []):
            content = getattr(msg, "content", None)
            if isinstance(content, str):
                haystack_parts.append(content)
        haystack = "\n".join(haystack_parts)
        return set(re.findall(r'https?://[^\s"\'<>)\]]+', haystack))

    def handle_images(self, req: ChatImageRequest) -> str:
      
        image_url = req.url
        if not image_url:
            return "No image was provided to analyze."

        system = (
            "You are an image analysis assistant. Your task is to examine the provided image "
            "and return a single-line description containing the most important visual details.\n\n"

            "ANALYZE:\n"
            "- Main subject or object.\n"
            "- Product type or category (if applicable).\n"
            "- Dominant colors.\n"
            "- Style, design, material, and notable features.\n"
            "- Visible text, logos, brands, or labels.\n"
            "- Setting or background.\n"
            "- People, clothing, poses, expressions, or activities if present.\n"
            "- Purpose of the image (advertisement, product showcase, promotional post, event, infographic, meme, etc.).\n"
            "- Main message or intent the image is trying to communicate.\n\n"

            "OUTPUT RULES:\n"
            "- Return exactly one line of plain text.\n"
            "- Do not use markdown, bullet points, JSON, or special formatting.\n"
            "- Be concise but information-dense.\n"
            "- If the image is a social media post or advertisement, include a short description of the marketing message.\n"
            "- If text is visible, summarize its meaning.\n"
            "- Do not speculate beyond what is reasonably visible.\n"
            "- If the image is unclear, state what can be confidently observed.\n\n"

            "EXAMPLE OUTPUT:\n"
            "Blue men's running shoes with white soles displayed on a clean studio background, sporty athletic design, product advertisement highlighting comfort and performance.\n"
        )

        messages = [
            SystemMessage(content=system),
            HumanMessage(content=[
                {"type": "text", "text": "Describe this image in one line, following the rules above."},
                {"type": "image_url", "image_url": {"url": image_url}},
            ]),
        ]

        try:
            response = analysis_model.invoke(messages)
            text = response.content or "Sorry, I couldn't analyze the image right now."
            return _single_line(_strip_markdown(text))
        except Exception:
            logger.exception("handle_images failed for url=%s", image_url)
            return "I'm having trouble analyzing this image right now. A team member will get back to you shortly."





    def handle_audio(self, req: ChatAudioRequest) -> str:
        audio_url = req.url
        if not audio_url:
            return "No audio was provided to analyze."

        mp4_path = None
        mp3_path = None

        try:
            response = requests.get(audio_url, timeout=60, headers={"User-Agent": "zovly-ai/1.0"})
            response.raise_for_status()

            with tempfile.NamedTemporaryFile(suffix=".mp4", delete=False) as tmp:
                tmp.write(response.content)
                mp4_path = tmp.name

            mp3_path = mp4_path.replace(".mp4", ".mp3")
            subprocess.run(
                ["ffmpeg", "-i", mp4_path, "-vn", "-ar", "16000", "-ac", "1", "-q:a", "0", mp3_path, "-y"],
                check=True, capture_output=True
            )

            segments, _ = whisper_model.transcribe(mp3_path, beam_size=5, language="ne", vad_filter=False)
            text = _single_line(_strip_markdown(" ".join(seg.text.strip() for seg in segments)))

            return text or "No speech could be transcribed from the audio."

        except subprocess.CalledProcessError as e:
            logger.error("ffmpeg failed: %s", e.stderr.decode())
            return "I'm having trouble processing this audio file."
        except Exception:
            logger.exception("handle_audio failed for url=%s", audio_url)
            return "I'm having trouble transcribing this audio right now."
        finally:
            for path in filter(None, [mp4_path, mp3_path]):
                try:
                    os.remove(path)
                except OSError:
                    pass

def _strip_markdown(text: str) -> str:
    text = re.sub(r'\*+', '', text)        # remove * and **
    text = re.sub(r'_+', '', text)         # remove _ and __
    text = re.sub(r'#+\s*', '', text)      # remove # headings
    text = re.sub(r'`+', '', text)         # remove backticks
    text = re.sub(r'\[([^\]]+)\]\([^)]+\)', r'\1', text)  # [text](url) → text
    text = re.sub(r'^\s*[-•]\s+', '', text, flags=re.MULTILINE)  # bullet points
    return text.strip()


def _single_line(text: str) -> str:
    # collapse any newlines / repeated whitespace into a single clean line
    return " ".join(text.split())


def get_chat_service() -> ChatService:
    return ChatService()