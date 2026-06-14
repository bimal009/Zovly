from app.models.faq import FaqRequest

def text_formatter(faq: FaqRequest) -> str:
    return f"Q: {faq.question}\nA: {faq.answer}"