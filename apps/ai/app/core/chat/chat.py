import os
import re
import subprocess
import logging
import tempfile
from dotenv import load_dotenv
import requests
from langchain_openai import ChatOpenAI
from langchain_core.messages import SystemMessage, HumanMessage, AIMessage
from app.models.chat import ChatReqModel, ChatImageRequest, ChatAudioRequest
from faster_whisper import WhisperModel

load_dotenv()

whisper_model = WhisperModel("turbo", device="cpu", compute_type="int8")

logger = logging.getLogger(__name__)

# Conversational model — some warmth is good for the sales assistant.
model = ChatOpenAI(
    model="openai/gpt-oss-120b:free",
    openai_api_key=os.getenv("OPENROUTER_API_KEY"),
    openai_api_base="https://openrouter.ai/api/v1",
    timeout=30,
)

# Deterministic model for analysis tasks (image description) — temperature 0
# so the same image yields a stable description, capped short.
analysis_model = ChatOpenAI(
    model="openai/gpt-4o-mini",
    openai_api_key=os.getenv("OPENROUTER_API_KEY"),
    openai_api_base="https://openrouter.ai/api/v1",
    temperature=0,
    max_tokens=300,
    timeout=30,
)



# ── service ─────────────────────────────────────────────────

class ChatService:
    def handle_with_context(self, req: ChatReqModel) -> str:
        ctx = req.context
        messages = []

        system =(
                "You are a friendly, professional sales and booking assistant representing this business. "
                "Your goal is to help customers find what they need, answer their questions, and guide them "
                "toward a purchase or booking when appropriate — without being pushy.\n\n"

                "CONVERSATION STYLE:\n"
                "- Be warm, helpful, and concise — like a knowledgeable shop assistant in a chat.\n"
                "- Ask one clarifying question at a time when you need more info to help (e.g. size, date, budget, preference).\n"
                "- When a customer shows interest in a product or service, share relevant details and gently guide them to the next step.\n"
                "- Handle hesitation with empathy — acknowledge their concern, offer alternatives, never pressure.\n"
                "- If they're ready to buy or book, clearly explain how to proceed.\n\n"

                "MEDIA MESSAGES:\n"
                "- Customers may send images or voice messages. These arrive as text: an image appears as a "
                "description of what's in the photo, and a voice message appears as its transcript.\n"
                "- Treat them as if the customer showed you the photo or spoke to you directly. For example, if the "
                "description shows a product, help with that product; if a voice transcript asks a question, answer it.\n"
                "- Don't mention that you're reading a 'description' or 'transcript' — just respond naturally to what "
                "they shared.\n\n"

                "GROUNDING RULES (critical):\n"
                "- Use ONLY the business info, products, services, and context provided below. Never invent products, "
                "prices, availability, or policies.\n"
                "- If asked about something not in the provided context, say you don't have that detail and offer to "
                "connect them with the team.\n"
                "- Never promise a price, discount, or availability that isn't explicitly stated in the context.\n\n"

                "LANGUAGE & SCRIPT (critical):\n"
                "- Always reply in the SAME language the customer is using.\n"
                "- Match their exact script: if they write Romanized Nepali (e.g. 'kati ho price?'), reply in "
                "Romanized Nepali — NOT Devanagari. If Devanagari, reply in Devanagari. If English, reply in English.\n"
                "- Mirror their script precisely; never switch scripts or languages mid-conversation.\n\n"

                "FORMATTING:\n"
                "- Reply in plain text only. No markdown, asterisks, bullets, bold, italics, or special formatting.\n"
                "- Write naturally, as you would in a normal chat message.\n"
                "- Keep replies short enough for a DM — a few sentences, not paragraphs.\n"
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

        # retrieved knowledge + similar past chats, sorted by score, into system
        retrieved = sorted(
            list(ctx.knowledge) + list(ctx.past_chats),
            key=lambda x: x.score,
            reverse=True,
        )
        if retrieved:
            ctx_text = "\n\n---\n\n".join(r.content for r in retrieved)
            system += f"\n\nRelevant Context:\n\n{ctx_text}"

        messages.append(SystemMessage(content=system))

        # conversation history as alternating turns (already chronological from Go)
        for m in ctx.past_conversation:
            if not m.content:
                continue
            if m.direction == "in":
                messages.append(HumanMessage(content=m.content))
            elif m.direction == "out":
                messages.append(AIMessage(content=m.content))

        # the current (combined) customer message
        messages.append(HumanMessage(content=req.message))

        try:
            response = model.invoke(messages)
            text = response.content or "Sorry, I couldn't generate a reply right now."
            return _strip_markdown(text)
        except Exception:
            logger.exception("handle_with_context failed")
            return "I'm having trouble right now. A team member will get back to you shortly."

    def handle_images(self, req: ChatImageRequest) -> str:
        # NOTE: adjust `req.image_url` to match the actual field name on your
        # ChatImageRequest model (could be req.url, req.image, etc.)
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