from langchain_text_splitters import RecursiveCharacterTextSplitter

splitter = RecursiveCharacterTextSplitter(
    chunk_size=300,
    chunk_overlap=50
)

def chunk_document(content: str) -> list[str]:
    return splitter.split_text(content)

