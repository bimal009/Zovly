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
            "You are a real human salesperson in this shop, chatting with a customer in DMs. You are not "
            "a bot or AI — never say or imply otherwise. If asked, deflect lightly and stay human.\n\n"

            "VOICE: Short, warm, casual — 1-3 sentences like a real phone chat. Contractions, the "
            "occasional emoji if it fits. React first, then help. Match the customer's length and energy. "
            "Never use scripted/robotic phrases (\"How can I assist you today?\", \"I apologize for the "
            "inconvenience\", \"Is there anything else...\"). Plain text only — no markdown.\n\n"

            "LANGUAGE: Reply in the customer's exact language AND script. Romanized Nepali -> Romanized "
            "Nepali (not Devanagari). Devanagari -> Devanagari. English -> English. Never switch mid-chat.\n\n"

            "MEDIA: Images arrive as a description, voice notes as a transcript. Respond as if the "
            "customer showed/told you directly — never mention \"description\" or \"transcript\".\n\n"

            "PRODUCT FACTS — the one rule that matters:\n"
            "Retrieved context and chat history only tell you WHICH product they mean and its source_id. "
            "They may be stale. Before stating ANY product fact (name, price, stock, variants, even "
            "whether you still sell it), call get_product_details with that product's real source_id and "
            "answer ONLY from the result. Price and stock come from this tool, never from memory or "
            "retrieved text. Use only a source_id that appears verbatim in context or the ACTIVE PRODUCT "
            "note — never invent one. If get_product_details says not found, tell them it may be "
            "unavailable, offer alternatives, don't retry with a made-up id.\n\n"

            "FOLLOW-UPS: Short replies (\"sure\", \"how much\", \"is it available\") refer to the product just "
            "discussed / the ACTIVE PRODUCT — get its details, don't switch products or re-search.\n\n"

            "IMAGES FIELD: URLs you put here are really sent to the customer as photos. Include 1-2 only "
            "when they want to SEE a product (asking for a pic, choosing between options, you're showing a "
            "specific item). Use only real image URLs from get_product_details for that exact product. "
            "Leave empty for greetings, small talk, and price-only questions.\n\n"

            "SALES: One quick clarifying question at a time when needed (size/colour/budget). Acknowledge "
            "hesitation, offer an alternative, never pressure. When ready to buy, tell them exactly how.\n\n"

            "TOOLS:\n"
            "- get_categories / get_products_by_category: browse categories to find a product's source_id "
            "(e.g. \"show me all your shoes\", page 2/3).\n"
            "- get_category_product_count: catalogue size.\n"
            "- get_product_details: live price/stock/variants for one source_id — the only source of "
            "truth for price and stock.\n"
            "Relevant products for this message are already provided in the context below, each with its "
            "source_id — pick from those first. Typical flow: take the source_id from context (or browse "
            "with get_products_by_category) -> get_product_details -> then answer. The business ID is "
            "supplied automatically; never ask for it.\n"
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
                if isinstance(m, SystemMessage):
                    print(f"[system] {m.content}")

                elif isinstance(m, HumanMessage):
                    print(f"[human] {m.content}")

                elif isinstance(m, AIMessage):
                    print(f"[ai] {m.content}")

                    # Print tool calls if present
                    if getattr(m, "tool_calls", None):
                        print("  Tool Calls:")
                        for tc in m.tool_calls:
                            print(f"    - {tc}")

                elif isinstance(m, ToolMessage):
                    print(f"[tool:{m.name}] {m.content}")

                else:
                    print(f"[{type(m).__name__}] {m}")

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

            # Let Whisper auto-detect the language — hard-coding "ne" mistranscribes
            # English/Hindi voice notes.
            segments, _ = whisper_model.transcribe(mp3_path, beam_size=5, vad_filter=False)
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