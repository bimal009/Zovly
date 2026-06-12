from langchain_huggingface import HuggingFaceEmbeddings
import numpy as np

embeddings = HuggingFaceEmbeddings(
    model_name="intfloat/multilingual-e5-large",
    encode_kwargs={"normalize_embeddings": True}
)

tests = [
    "नमस्ते मेरो order कहाँ छ?",
    "Where is my order?",
    "mero order kahaa cha bro?",
]

def cosine_similarity(a, b):
    return np.dot(a, b) / (np.linalg.norm(a) * np.linalg.norm(b))

vectors = [embeddings.embed_query(t) for t in tests]

print("Similarity scores:\n")
for i in range(len(tests)):
    for j in range(i + 1, len(tests)):
        score = cosine_similarity(vectors[i], vectors[j])
        print(f"{tests[i][:30]}")
        print(f"{tests[j][:30]}")
        print(f"Score: {score:.4f}\n")