import os
import re
from dotenv import load_dotenv
from langchain_openai import ChatOpenAI
from langchain_core.messages import SystemMessage, HumanMessage, AIMessage
from app.models.chat import ChatReqModel

load_dotenv()

model = ChatOpenAI(
    model="openai/gpt-4o-mini",
    openai_api_key=os.getenv("OPENROUTER_API_KEY"),
    openai_api_base="https://openrouter.ai/api/v1",
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
            return "I'm having trouble right now. A team member will get back to you shortly."


def _strip_markdown(text: str) -> str:
    text = re.sub(r'\*+', '', text)        # remove * and **
    text = re.sub(r'_+', '', text)         # remove _ and __
    text = re.sub(r'#+\s*', '', text)      # remove # headings
    text = re.sub(r'`+', '', text)         # remove backticks
    text = re.sub(r'\[([^\]]+)\]\([^)]+\)', r'\1', text)  # [text](url) → text
    text = re.sub(r'^\s*[-•]\s+', '', text, flags=re.MULTILINE)  # bullet points
    return text.strip()


def get_chat_service() -> ChatService:
    return ChatService()